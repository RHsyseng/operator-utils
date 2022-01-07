package write

import (
	"context"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestFluentAPI(t *testing.T) {
	scheme := getScheme(t)

	noOwnership := New(fake.NewClientBuilder().WithScheme(scheme).Build())
	assert.Nil(t, noOwnership.ownerRefs, "Do not expect ownerRefs to be set")
	assert.Nil(t, noOwnership.ownerController, "Do not expect ownerController to be set")

	ownerRefs := New(fake.NewClientBuilder().WithScheme(scheme).Build()).WithOwnerController(&corev1.Service{}, scheme)
	assert.Nil(t, ownerRefs.ownerRefs, "Do not expect ownerRefs to be set")
	assert.NotNil(t, ownerRefs.ownerController, "Expect ownerController to be set")

	controler := New(fake.NewClientBuilder().WithScheme(scheme).Build()).WithOwnerReferences(v1.OwnerReference{})
	assert.NotNil(t, controler.ownerRefs, "Expect ownerRefs to be set")
	assert.Nil(t, controler.ownerController, "Do not expect ownerController to be set")
}

func TestCreateService(t *testing.T) {
	scheme := getScheme(t)
	cli := fake.NewClientBuilder().WithScheme(scheme).Build()
	requestedService := corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service1",
			Namespace: "namespace",
		},
		Spec: corev1.ServiceSpec{
			SessionAffinity: corev1.ServiceAffinityClientIP,
		},
	}
	requestedService.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
	added, err := New(cli).AddResources([]client.Object{&requestedService})
	assert.Nil(t, err, "Expect no errors creating a simple object")
	assert.True(t, added, "Object should be added")

	existingService := corev1.Service{}
	err = cli.Get(context.TODO(), types.NamespacedName{Name: "service1", Namespace: "namespace"}, &existingService)
	assert.Nil(t, err, "Expect no errors loading existing object")
	assert.Equal(t, requestedService, existingService)
}

func TestUpdateService(t *testing.T) {
	scheme := getScheme(t)
	clusterIP := "1.2.3.4"
	cli := fake.NewClientBuilder().WithScheme(scheme).Build()
	requestedService := corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service1",
			Namespace: "namespace",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:       clusterIP,
			SessionAffinity: corev1.ServiceAffinityClientIP,
		},
	}
	requestedService.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
	added, err := New(cli).AddResources([]client.Object{&requestedService})
	assert.Nil(t, err, "Expect no errors creating a simple object")
	assert.True(t, added, "Object should be added")

	updatedService := corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service1",
			Namespace: "namespace",
		},
		Spec: corev1.ServiceSpec{
			SessionAffinity: corev1.ServiceAffinityNone,
		},
	}
	updatedService.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
	updated, err := New(cli).UpdateResources([]client.Object{&requestedService}, []client.Object{&updatedService})
	assert.Nil(t, err, "Expect no errors updating object")
	assert.True(t, updated, "Object should be updated")

	existingService := corev1.Service{}
	err = cli.Get(context.TODO(), types.NamespacedName{Name: "service1", Namespace: "namespace"}, &existingService)
	assert.Nil(t, err, "Expect no errors loading existing object")
	//Update call should set the existing ClusterIP on the object before writing it:
	updatedService.Spec.ClusterIP = clusterIP
	assert.Equal(t, updatedService, existingService, "Expected Cluster IP to be set on the updating object")
}

func getScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	err := corev1.SchemeBuilder.AddToScheme(scheme)
	assert.Nil(t, err, "Expect no errors building scheme")
	return scheme
}
