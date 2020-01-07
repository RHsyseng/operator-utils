package detector

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	discoveryFake "k8s.io/client-go/discovery/fake"
	k8sTesting "k8s.io/client-go/testing"
	"testing"
	"time"
)

func TestDetectorDetects(t *testing.T) {
	crdDiscovered := false
	dc := &discoveryFake.FakeDiscovery{
		Fake:               &k8sTesting.Fake{},
		FakedServerVersion: nil,
	}

	d, err := NewAutoDetect(dc)
	if err != nil {
		t.Fatalf("expected no errors, got: %s", err.Error())
	}

	// run very frequently, for faster tests
	d.Start(1 * time.Nanosecond)
	d.AddCRDTrigger(&appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
	}, func(crd runtime.Object) {
		crdDiscovered = true

	})

	//wait a few intervals
	time.Sleep(2 * time.Millisecond)

	if crdDiscovered {
		t.Fatalf("CRD Discovered too early")
	}

	dc.Resources = []*metav1.APIResourceList{
		&metav1.APIResourceList{
			TypeMeta:     metav1.TypeMeta{},
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{{Kind: "deployment"}},
		},
	}

	time.Sleep(2 * time.Millisecond)
	if !crdDiscovered {
		t.Fatalf("CRD not discovered correctly")
	}
}
