package labelPolicy

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var nodes = &v1.NodeList{
	Items: []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"test1": "true",
					"test2": "true",
					"test3": "true",
					"test4": "false",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"test1": "false",
					"test2": "true",
					"test4": "false",
				},
			},
		},
	},
}

func TestOneTrue(t *testing.T) {
	policy := GetInstance(LabelPolicyOneTrue)
	assert.NotNil(t, policy)

	val := policy.Process(nodes, "test1")
	assert.EqualValues(t, "true", val)

	val = policy.Process(nodes, "test2")
	assert.EqualValues(t, "true", val)

	val = policy.Process(nodes, "test3")
	assert.EqualValues(t, "true", val)

	val = policy.Process(nodes, "test4")
	assert.EqualValues(t, "false", val)

	val = policy.Process(nodes, "test5")
	assert.EqualValues(t, "false", val)
}

func TestAllTrue(t *testing.T) {
	policy := GetInstance(LabelPolicyAllTrue)
	assert.NotNil(t, policy)

	val := policy.Process(nodes, "test1")
	assert.EqualValues(t, "false", val)

	val = policy.Process(nodes, "test2")
	assert.EqualValues(t, "true", val)

	val = policy.Process(nodes, "test3")
	assert.EqualValues(t, "false", val)

	val = policy.Process(nodes, "test4")
	assert.EqualValues(t, "false", val)

	val = policy.Process(nodes, "test5")
	assert.EqualValues(t, "false", val)
}
