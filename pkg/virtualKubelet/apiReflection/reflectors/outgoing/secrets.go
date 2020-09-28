package outgoing

import (
	"context"
	"errors"
	apimgmt "github.com/liqotech/liqo/pkg/virtualKubelet/apiReflection"
	ri "github.com/liqotech/liqo/pkg/virtualKubelet/apiReflection/reflectors/reflectorsInterfaces"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"strings"
)

type SecretsReflector struct {
	ri.APIReflector
}

func (r *SecretsReflector) SetSpecializedPreProcessingHandlers() {
	r.SetPreProcessingHandlers(ri.PreProcessingHandlers{
		AddFunc:    r.PreAdd,
		UpdateFunc: r.PreUpdate,
		DeleteFunc: r.PreDelete})
}

func (r *SecretsReflector) HandleEvent(e interface{}) {
	var err error

	event := e.(watch.Event)
	secret, ok := event.Object.(*corev1.Secret)
	if !ok {
		klog.Error("REFLECTION: cannot cast object to Secret")
		return
	}
	klog.V(3).Infof("REFLECTION: received %v for Secret %v/%v", event.Type, secret.Namespace, secret.Name)

	switch event.Type {
	case watch.Added:
		_, err := r.GetForeignClient().CoreV1().Secrets(secret.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if kerrors.IsAlreadyExists(err) {
			klog.V(3).Infof("REFLECTION: The remote Secret %v/%v has not been created: %v", secret.Namespace, secret.Name, err)
		}
		if err != nil && !kerrors.IsAlreadyExists(err) {
			klog.Errorf("REFLECTION: Error while updating the remote Secret %v/%v - ERR: %v", secret.Namespace, secret.Name, err)
		} else {
			klog.V(3).Infof("REFLECTION: remote Secret %v/%v correctly created", secret.Namespace, secret.Name)
		}

	case watch.Modified:
		if _, err = r.GetForeignClient().CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("REFLECTION: Error while updating the remote Secret %v/%v - ERR: %v", secret.Namespace, secret.Name, err)
		} else {
			klog.V(3).Infof("REFLECTION: remote Secret %v/%v correctly updated", secret.Namespace, secret.Name)
		}

	case watch.Deleted:
		if err := r.GetForeignClient().CoreV1().Secrets(secret.Namespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{}); err != nil {
			klog.Errorf("REFLECTION: Error while deleting the remote Secret %v/%v - ERR: %v", secret.Namespace, secret.Name, err)
		} else {
			klog.V(3).Infof("REFLECTION: remote Secret %v/%v correctly deleted", secret.Namespace, secret.Name)
		}
	}
}

func (r *SecretsReflector) KeyerFromObj(obj interface{}, remoteNamespace string) string {
	cm, ok := obj.(*corev1.Secret)
	if !ok {
		return ""
	}
	return strings.Join([]string{remoteNamespace, cm.Name}, "/")
}

func (r *SecretsReflector) CleanupNamespace(localNamespace string) {
	foreignNamespace, err := r.NattingTable().NatNamespace(localNamespace, false)
	if err != nil {
		klog.Error(err)
		return
	}

	objects := r.ForeignInformer(foreignNamespace).GetStore().List()
	for _, obj := range objects {
		secret := obj.(*corev1.Secret)
		if err := r.GetForeignClient().CoreV1().Secrets(foreignNamespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{}); err != nil {
			klog.Errorf("error while deleting Secret %v/%v - ERR: %v", secret.Name, secret.Namespace, err)
		}
	}
}

func (r *SecretsReflector) PreAdd(obj interface{}) interface{} {
	secretLocal := obj.(*corev1.Secret)
	klog.V(3).Infof("PreAdd routine started for Secret %v/%v", secretLocal.Namespace, secretLocal.Name)

	nattedNs, err := r.NattingTable().NatNamespace(secretLocal.Namespace, false)
	if err != nil {
		klog.Error(err)
		return nil
	}

	secretRemote := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretLocal.Name,
			Namespace: nattedNs,
			Labels:    make(map[string]string),
		},
		Data:       secretLocal.Data,
		StringData: secretLocal.StringData,
		Type:       secretLocal.Type,
	}
	for k, v := range secretLocal.Labels {
		secretRemote.Labels[k] = v
	}
	secretRemote.Labels[apimgmt.LiqoLabelKey] = apimgmt.LiqoLabelValue

	klog.V(3).Infof("PreAdd routine completed for configmap %v/%v", secretLocal.Namespace, secretLocal.Name)
	return secretRemote
}

func (r *SecretsReflector) PreUpdate(newObj interface{}, _ interface{}) interface{} {
	newSecret := newObj.(*corev1.Secret).DeepCopy()

	nattedNs, err := r.NattingTable().NatNamespace(newSecret.Namespace, false)
	if err != nil {
		klog.Error(err)
		return nil
	}


	name := r.KeyerFromObj(newObj, nattedNs)
	oldRemoteObj, exists, err := r.ForeignInformer(nattedNs).GetStore().GetByKey(name)
	if err != nil {
		klog.Error(err)
		return nil
	}
	if !exists {
		err = r.ForeignInformer(nattedNs).GetStore().Resync()
		if err != nil {
			klog.Error(err)
			return nil
		}
	}
	oldRemoteSvc := oldRemoteObj.(*corev1.Secret)

	newSecret.SetNamespace(nattedNs)
	newSecret.SetResourceVersion(oldRemoteSvc.ResourceVersion)
	newSecret.SetUID(oldRemoteSvc.UID)
	return newSecret
}

func (r *SecretsReflector) PreDelete(obj interface{}) interface{} {
	serviceLocal := obj.(*corev1.Secret)
	klog.V(3).Infof("PreDelete routine started for configmap %v/%v", serviceLocal.Namespace, serviceLocal.Name)

	nattedNs, err := r.NattingTable().NatNamespace(serviceLocal.Namespace, false)
	if err != nil {
		klog.Error(err)
		return nil
	}
	serviceLocal.Namespace = nattedNs

	klog.V(3).Infof("PreDelete routine completed for configmap %v/%v", serviceLocal.Namespace, serviceLocal.Name)
	return serviceLocal
}

func addSecretsIndexers() cache.Indexers {
	i := cache.Indexers{}
	i["secrets"] = func(obj interface{}) ([]string, error) {
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			return []string{}, errors.New("cannot convert obj to secret")
		}
		return []string{
			strings.Join([]string{secret.Namespace, secret.Name}, "/"),
		}, nil
	}
	return i
}
