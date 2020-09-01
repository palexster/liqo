package client

import (
	"fmt"
	advertisementApi "github.com/liqoTech/liqo/api/sharing/v1alpha1"
	advertisement "github.com/liqoTech/liqo/internal/advertisement-operator"
	"strings"
)

//createAdvertisementController creates a new CRDController for the Liqo Advertisement CRD.
func createAdvertisementController(kubeconfig string) (*CRDController, error) {
	controller := &CRDController{
		addFunc:    advertisementAddFunc,
		updateFunc: advertisementUpdateFunc,
		deleteFunc: advertisementDeleteFunc,
	}
	//init client
	newClient, err := advertisementApi.CreateAdvertisementClient(kubeconfig, nil)
	if err != nil {
		return nil, err
	}
	/*The CRDClient generated by CreateAdvertisementClient contains a cache which is already running, but
	with no handler functions. In order to allow further uses of it and to create a consistent starting status
	for each client, this cache is immediately stopped.*/
	close(newClient.Stop)
	controller.CRDClient = newClient
	controller.resource = string(CRAdvertisement)
	return controller, nil
}

//advertisementAddFunc is the ADD event handler for the Advertisement CRDController.
func advertisementAddFunc(obj interface{}) {
	newAdv := obj.(*advertisementApi.Advertisement)
	if newAdv.Status.AdvertisementStatus == advertisement.AdvertisementAccepted {
		agentCtrl.NotifyChannel(ChanAdvAccepted) <- newAdv.Name
	} else {
		agentCtrl.NotifyChannel(ChanAdvNew) <- newAdv.Name
	}
}

//advertisementUpdateFunc is the UPDATE event handler for the Advertisement CRDController.
func advertisementUpdateFunc(oldObj interface{}, newObj interface{}) {
	oldAdv := oldObj.(*advertisementApi.Advertisement)
	newAdv := newObj.(*advertisementApi.Advertisement)
	if oldAdv.Status.AdvertisementStatus != advertisement.AdvertisementAccepted && newAdv.Status.AdvertisementStatus == advertisement.AdvertisementAccepted {
		agentCtrl.NotifyChannel(ChanAdvAccepted) <- newAdv.Name
	} else if oldAdv.Status.AdvertisementStatus == advertisement.AdvertisementAccepted && newAdv.Status.AdvertisementStatus != advertisement.AdvertisementAccepted {
		agentCtrl.NotifyChannel(ChanAdvRevoked) <- newAdv.Name
	}
}

//advertisementDeleteFunc is the DELETE event handler for the Advertisement CRDController.
func advertisementDeleteFunc(obj interface{}) {
	adv := obj.(*advertisementApi.Advertisement)
	agentCtrl.NotifyChannel(ChanAdvDeleted) <- adv.Name
}

//DescribeAdvertisement provides a textual representation of an Advertisement CR
//that can be displayed in a MenuNode.
func DescribeAdvertisement(adv *advertisementApi.Advertisement) string {
	str := strings.Builder{}
	prices := adv.Spec.Prices
	str.WriteString(fmt.Sprintf("• ClusterID: %v\n", adv.Spec.ClusterId))
	str.WriteString(fmt.Sprintf("\t• STATUS: %v\n", adv.Status.AdvertisementStatus))
	str.WriteString("\t• Available Resources:\n")
	str.WriteString(fmt.Sprintf("\t\t- shared cpu = %v ", adv.Spec.ResourceQuota.Hard.Cpu()))
	if CpuPrice, cFound := prices["cpu"]; cFound {
		str.WriteString(fmt.Sprintf("[price %v]", CpuPrice.String()))
	}
	str.WriteString("\n")
	str.WriteString(fmt.Sprintf("\t\t-shared memory = %v ", adv.Spec.ResourceQuota.Hard.Memory()))
	if MemPrice, mFound := prices["memory"]; mFound {
		str.WriteString(fmt.Sprintf("[price %v]", MemPrice.String()))
	}
	str.WriteString("\n")
	str.WriteString(fmt.Sprintf("\t\t-available pods = %v ", adv.Spec.ResourceQuota.Hard.Pods()))
	return str.String()
}
