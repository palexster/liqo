package kubevirt

import (
	v1 "k8s.io/api/core/v1"
)

const (
	KubevirtKvm      v1.ResourceName = "devices.kubevirt.io/kvm"
	KubevirtTun      v1.ResourceName = "devices.kubevirt.io/tun"
	KubevirtVhostNet v1.ResourceName = "devices.kubevirt.io/vhost-net"
)
