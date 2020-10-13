package labelPolicy

import corev1 "k8s.io/api/core/v1"

type OneTrue struct{}

func (ot *OneTrue) Process(physicalNodes *corev1.NodeList, key string) (value string) {
	for _, node := range physicalNodes.Items {
		if v, ok := node.Labels[key]; ok && v == "true" {
			return "true"
		}
	}
	return "false"
}
