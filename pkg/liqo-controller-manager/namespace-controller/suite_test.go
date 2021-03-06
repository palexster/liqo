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
	"flag"
	"fmt"
	mapsv1alpha1 "github.com/liqotech/liqo/apis/virtualKubelet/v1alpha1"
	const_ctrl "github.com/liqotech/liqo/pkg/liqo-controller-manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	// +kubebuilder:scaffold:imports
)

const (
	nameVirtualNode1    = "virtual-node-1"
	nameVirtualNode2    = "virtual-node-2"
	nameNamespaceTest   = "namespace-test"
	nameRemoteNamespace = "namespace-test-remoto"
	mapNamespaceName    = "default"

	remoteClusterId1 = "6a0e9f-b52-4ed0"
	remoteClusterId2 = "899890-dsd-323"

	randomLabel              = "random"
	offloadingCluster1Label1 = "offloading.liqo.io/cluster-1"
	offloadingCluster1Label2 = "offloading.liqo.io/AWS"
	offloadingCluster2Label1 = "offloading.liqo.io/cluster-2"
	offloadingCluster2Label2 = "offloading.liqo.io/GKE"
)

var (
	cfg          *rest.Config
	k8sClient    client.Client
	testEnv      *envtest.Environment
	ctx          context.Context
	namespace    *corev1.Namespace
	nms          *mapsv1alpha1.NamespaceMapList
	nm1          *mapsv1alpha1.NamespaceMap
	nm2          *mapsv1alpha1.NamespaceMap
	virtualNode1 *corev1.Node
	virtualNode2 *corev1.Node
	flags        *flag.FlagSet
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "deployments", "liqo", "crds")},
	}

	flags = &flag.FlagSet{}
	klog.InitFlags(flags)
	_ = flags.Set("v", "2")

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = mapsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&NamespaceReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	namespace = &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: nameNamespaceTest,
			Labels: map[string]string{
				randomLabel: "",
			},
		},
	}

	virtualNode1 = &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: nameVirtualNode1,
			Annotations: map[string]string{
				const_ctrl.VirtualNodeClusterId: remoteClusterId1,
			},
			Labels: map[string]string{
				const_ctrl.TypeLabel:     const_ctrl.TypeNode,
				offloadingCluster1Label1: "",
				offloadingCluster1Label2: "",
			},
		},
	}

	virtualNode2 = &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: nameVirtualNode2,
			Annotations: map[string]string{
				const_ctrl.VirtualNodeClusterId: remoteClusterId2,
			},
			Labels: map[string]string{
				const_ctrl.TypeLabel:     const_ctrl.TypeNode,
				offloadingCluster2Label1: "",
				offloadingCluster2Label2: "",
			},
		},
	}

	nm1 = &mapsv1alpha1.NamespaceMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "namespaceresources.liqo.io/v1",
			Kind:       "NamespaceMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", remoteClusterId1),
			Namespace:    mapNamespaceName,
			Labels: map[string]string{
				const_ctrl.VirtualNodeClusterId: remoteClusterId1,
			},
		},
	}

	nm2 = &mapsv1alpha1.NamespaceMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "namespaceresources.liqo.io/v1",
			Kind:       "NamespaceMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", remoteClusterId2),
			Namespace:    mapNamespaceName,
			Labels: map[string]string{
				const_ctrl.VirtualNodeClusterId: remoteClusterId2,
			},
		},
	}

	Expect(k8sClient.Create(context.TODO(), virtualNode1)).Should(Succeed())
	Expect(k8sClient.Create(context.TODO(), virtualNode2)).Should(Succeed())
	Expect(k8sClient.Create(context.TODO(), nm1)).Should(Succeed())
	Expect(k8sClient.Create(context.TODO(), nm2)).Should(Succeed())
	Expect(k8sClient.Create(context.TODO(), namespace)).Should(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
