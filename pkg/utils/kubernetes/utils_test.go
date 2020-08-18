package kubernetes

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/fake"
	k8sTesting "k8s.io/client-go/testing"
	"testing"
)

func TestCustomResourceDefinitionExists(t *testing.T) {
	client := &fake.FakeDiscovery{
		Fake:               &k8sTesting.Fake{},
		FakedServerVersion: nil,
	}
	client.Resources = []*metav1.APIResourceList{
		{
			TypeMeta:     metav1.TypeMeta{},
			GroupVersion: "console.openshift.io/v1",
			APIResources: []metav1.APIResource{{Kind: "ConsoleYAMLSample"}},
		},
	}
	gvk := schema.GroupVersionKind{Group: "console.openshift.io", Version: "v1", Kind: "ConsoleYAMLSample"}
	err := customResourceDefinitionExists(gvk, client)
	assert.Nil(t, err, "Failed to find ", gvk)

	gvk = schema.GroupVersionKind{Group: "console.openshift.io", Version: "v2", Kind: "ConsoleYAMLSample"}
	err = customResourceDefinitionExists(gvk, client)
	assert.NotNil(t, err, "Did not expect to find ", gvk)
}
