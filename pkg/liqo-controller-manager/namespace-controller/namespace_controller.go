/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package namespace_controller

import (
	"context"
	mapsv1alpha1 "github.com/liqotech/liqo/apis/virtualKubelet/v1alpha1"
	const_ctrl "github.com/liqotech/liqo/pkg/liqo-controller-manager"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/slice"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	mappingLabel          = "mapping.liqo.io"
	offloadingLabel       = "offloading.liqo.io"
	offloadingPrefixLabel = "offloading.liqo.io/"
	namespaceFinalizer    = "namespace-controller.liqo.io/finalizer"
)

func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	namespace := &corev1.Namespace{}
	if err := r.Get(context.TODO(), req.NamespacedName, namespace); err != nil {
		klog.Error(err, " --> Unable to get namespace")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	namespaceMaps := &mapsv1alpha1.NamespaceMapList{}
	if err := r.List(context.TODO(), namespaceMaps); err != nil {
		klog.Error(err, " --> Unable to List NamespaceMaps")
		return ctrl.Result{}, err
	}

	if len(namespaceMaps.Items) == 0 {
		klog.Info(" No namespaceMaps at the moment")
		return ctrl.Result{}, nil
	}

	removeMappings := make(map[string]mapsv1alpha1.NamespaceMap)
	for _, namespaceMap := range namespaceMaps.Items {
		removeMappings[namespaceMap.GetLabels()[const_ctrl.VirtualNodeClusterId]] = namespaceMap
	}

	if namespace.GetDeletionTimestamp().IsZero() {
		if !slice.ContainsString(namespace.GetFinalizers(), namespaceFinalizer, nil) {
			namespace.SetFinalizers(append(namespace.GetFinalizers(), namespaceFinalizer))
			if err := r.Patch(context.TODO(), namespace, client.Merge); err != nil {
				klog.Error(err, " --> Unable to add finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if slice.ContainsString(namespace.GetFinalizers(), namespaceFinalizer, nil) {

			if err := r.removeRemoteNamespaces(namespace.GetName(), removeMappings); err != nil {
				return ctrl.Result{}, err
			}

			klog.Info(" Someone try to delete namespace, ok delete!!")

			namespace.SetFinalizers(slice.RemoveString(namespace.GetFinalizers(), namespaceFinalizer, nil))
			if err := r.Update(context.TODO(), namespace); err != nil {
				klog.Error(err, " --> Unable to remove finalizer")
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// 1. If mapping.liqo.io label is not present there are no remote namespaces associated with this namespace, removeMappings is full
	if remoteNamespaceName, ok := namespace.GetLabels()[mappingLabel]; ok {

		// 2.a If offloading.liqo.io is present there are remote namespaces on all virtual nodes
		if _, ok = namespace.GetLabels()[offloadingLabel]; ok {
			klog.Infof(" Offload namespace '%s' on all remote clusters", namespace.GetName())
			if err := r.createRemoteNamespaces(namespace, remoteNamespaceName, removeMappings); err != nil {
				return ctrl.Result{}, err
			}

			for k := range removeMappings {
				delete(removeMappings, k)
			}

		} else {

			// 2.b Iterate on all virtual nodes' labels, if the namespace has all the requested labels, is necessary to
			// offload it onto remote cluster associated with the virtual node
			nodes := &corev1.NodeList{}
			if err := r.List(context.TODO(), nodes, client.MatchingLabels{const_ctrl.TypeLabel: const_ctrl.TypeNode}); err != nil {
				klog.Error(err, " --> Unable to List all virtual nodes")
				return ctrl.Result{}, err
			}

			if len(nodes.Items) == 0 {
				klog.Info(" No VirtualNode at the moment")
				return ctrl.Result{}, nil
			}

			for _, node := range nodes.Items {
				if checkOffloadingLabels(namespace, &node) {
					if err := r.createRemoteNamespace(namespace, remoteNamespaceName, removeMappings[node.Annotations[const_ctrl.VirtualNodeClusterId]]); err != nil {
						return ctrl.Result{}, err
					}
					delete(removeMappings, node.Annotations[const_ctrl.VirtualNodeClusterId])
					klog.Infof(" Offload namespace '%s' on remote cluster: %s", namespace.GetName(), node.Annotations[const_ctrl.VirtualNodeClusterId])
				}

			}
		}

	}

	if len(removeMappings) > 0 {
		klog.Info(" Delete all unnecessary mapping in NamespaceMaps")
		if err := r.removeRemoteNamespaces(namespace.GetName(), removeMappings); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		WithEventFilter(manageLabelPredicate()).
		Complete(r)
}
