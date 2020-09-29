package logic

import (
	advtypes "github.com/liqotech/liqo/api/sharing/v1alpha1"
	"github.com/liqotech/liqo/internal/tray-agent/agent/client"
	app "github.com/liqotech/liqo/internal/tray-agent/app-indicator"
)

// callback function for the ACTION "Show Advertisements". It shows the Advertisements CRs currently in the cluster,
// indicating whether they are 'ACCEPTED' or not.
func actionShowAdv() {
	i := app.GetIndicator()
	if ctrl := i.AgentCtrl(); ctrl != nil {
		advController := ctrl.Controller(client.CRAdvertisement)
		advCache := advController.Store
		if ctrl.Connected() && advController.Running() {
			// start indicator ACTION
			act, pres := i.Action(aShowPeers)
			if !pres {
				return
			}
			i.SelectAction(aShowPeers)
			i.SetIcon(app.IconLiqoMain)
			// exec ACTION
			if !ctrl.Mocked() {
				for _, obj := range advCache.List() {
					adv := obj.(*advtypes.Advertisement)
					element := act.UseListChild()
					element.SetTitle(client.DescribeAdvertisement(adv))
				}
			}
		} else {
			i.NotifyNoConnection()
		}
	} else {
		i.NotifyNoConnection()
	}

}
