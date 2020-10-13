package labelPolicy

import corev1 "k8s.io/api/core/v1"

type LabelPolicyType string

// NOTE: add these values to the accepted values in apis/config/v1alpha1/clusterconfig_types.go > LabelPolicy > Policy
const (
	LabelPolicyOneTrue LabelPolicyType = "LabelPolicyOneTrue"
	LabelPolicyAllTrue LabelPolicyType = "LabelPolicyAllTrue"
)

type LabelPolicy interface {
	Process(physicalNodes *corev1.NodeList, key string) (value string)
}

func GetInstance(policyType LabelPolicyType) LabelPolicy {
	switch policyType {
	case LabelPolicyOneTrue:
		return &OneTrue{}
	case LabelPolicyAllTrue:
		return &AllTrue{}
	default:
		return &OneTrue{}
	}
}
