package kubernetes

import (
	"context"
	"fmt"
	"testing"

	"github.com/RHsyseng/operator-utils/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MockReconciler struct {
	ReconcileFn func(context context.Context, request reconcile.Request) (reconcile.Result, error)
}

func (r *MockReconciler) Reconcile(context context.Context, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

type MockFinalizer struct {
	name         string
	onFinalizeFn func(owner client.Object, service PlatformService) error
}

func (m *MockFinalizer) GetName() string {
	return m.name
}

func (m *MockFinalizer) OnFinalize(owner client.Object, service PlatformService) error {
	if m.onFinalizeFn == nil {
		return nil
	}
	return m.onFinalizeFn(owner, service)
}

func (m *MockFinalizer) setOnFinalizeFn(onFinalizeFn func(owner client.Object, service PlatformService) error) {
	m.onFinalizeFn = onFinalizeFn
}

func TestExtendedReconciler_IsFinalizing(t *testing.T) {
	extReconciler := BuildTestExtendedReconciler()

	assert.False(t, extReconciler.isFinalizing(extReconciler.Resource))

	extReconciler.Resource.SetDeletionTimestamp(&metav1.Time{})
	assert.True(t, extReconciler.isFinalizing(extReconciler.Resource))
}

func TestExtendedReconciler_RegisterFinalizer(t *testing.T) {
	extReconciler := BuildTestExtendedReconciler()

	assert.Len(t, extReconciler.Finalizers, 0)

	err := extReconciler.RegisterFinalizer(&MockFinalizer{})
	assert.Errorf(t, err, "the finalizer name must not be empty")
	assert.Len(t, extReconciler.Finalizers, 0)

	err = extReconciler.RegisterFinalizer(&MockFinalizer{
		name: "finalizer1",
	})
	assert.Nil(t, err)
	assert.Len(t, extReconciler.Finalizers, 1)

	err = extReconciler.RegisterFinalizer(&MockFinalizer{
		name: "finalizer2",
	})
	assert.Nil(t, err)
	assert.Len(t, extReconciler.Finalizers, 2)

	err = extReconciler.RegisterFinalizer(&MockFinalizer{
		name: "finalizer2",
	})
	assert.Nil(t, err)
	assert.Len(t, extReconciler.Finalizers, 2)
}

func TestExtendedReconciler_UnregisterFinalizer(t *testing.T) {
	extReconciler := BuildTestExtendedReconciler()
	extReconciler.Finalizers = map[string]Finalizer{
		"f1": &MockFinalizer{},
		"f2": &MockFinalizer{},
	}

	err := extReconciler.UnregisterFinalizer("")
	assert.Errorf(t, err, "the finalizer name must not be empty")
	assert.Len(t, extReconciler.Finalizers, 2)

	err = extReconciler.UnregisterFinalizer("f1")
	assert.Nil(t, err)
	assert.Len(t, extReconciler.Finalizers, 1)

	err = extReconciler.UnregisterFinalizer("f1")
	assert.Nil(t, err)
	assert.Len(t, extReconciler.Finalizers, 1)

	err = extReconciler.UnregisterFinalizer("f2")
	assert.Nil(t, err)
	assert.Len(t, extReconciler.Finalizers, 0)
}

func TestExtendedReconciler_FinalizeOnDelete(t *testing.T) {
	extReconciler := BuildTestExtendedReconciler()
	extReconciler.Finalizers = map[string]Finalizer{
		"f1": &MockFinalizer{},
		"f2": &MockFinalizer{},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "somepod",
			Namespace: "somenamespace",
		},
	}
	pod.SetFinalizers([]string{"f1", "f2"})

	extReconciler.Service.Create(context.TODO(), pod)

	err := extReconciler.finalizeOnDelete(pod)
	assert.Nil(t, err)
	assert.Len(t, pod.GetFinalizers(), 2)
	assert.Len(t, extReconciler.Finalizers, 2)

	extReconciler.Service.Delete(context.TODO(), pod)
	err = extReconciler.Service.Get(context.TODO(), types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}, pod)
	assert.Nil(t, err)
	err = extReconciler.finalizeOnDelete(pod)
	assert.Nil(t, err)
	assert.Empty(t, pod.GetFinalizers())
	assert.Len(t, extReconciler.Finalizers, 2)
}

func TestExtendedReconciler_FinalizeOnDeleteUnregisteredFinalizer(t *testing.T) {
	extReconciler := BuildTestExtendedReconciler()
	extReconciler.Finalizers = map[string]Finalizer{
		"f1": &MockFinalizer{},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "somepod",
			Namespace: "somenamespace",
		},
	}
	pod.SetFinalizers([]string{"f1", "f2"})
	pod.SetDeletionTimestamp(&metav1.Time{})
	extReconciler.Service.Create(context.TODO(), pod)

	extReconciler.Service.Delete(context.TODO(), pod)
	err := extReconciler.Service.Get(context.TODO(), types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}, pod)
	assert.Nil(t, err)

	err = extReconciler.finalizeOnDelete(pod)
	assert.Errorf(t, err, "finalizer f2 does not have a Finalizer handler registered")

	newPod := &v1.Pod{}
	err = extReconciler.Service.Get(context.TODO(), types.NamespacedName{Name: pod.GetName(), Namespace: pod.GetNamespace()}, newPod)
	assert.Nil(t, err)
	assert.Len(t, newPod.GetFinalizers(), 1)
	assert.Len(t, extReconciler.Finalizers, 1)
}

func TestExtendedReconciler_FinalizeOnDeleteErrorOnFinalize(t *testing.T) {
	extReconciler := BuildTestExtendedReconciler()
	extReconciler.Finalizers = map[string]Finalizer{
		"f1": &MockFinalizer{
			onFinalizeFn: func(owner client.Object, service PlatformService) error {
				return fmt.Errorf("Foo error")
			},
		},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "somepod",
			Namespace: "somenamespace",
		},
	}
	pod.SetFinalizers([]string{"f1"})
	pod.SetDeletionTimestamp(&metav1.Time{})
	extReconciler.Service.Create(context.TODO(), pod)

	err := extReconciler.finalizeOnDelete(pod)
	assert.Errorf(t, err, "Foo error")

	newPod := &v1.Pod{}
	err = extReconciler.Service.Get(context.TODO(), types.NamespacedName{Name: pod.GetName(), Namespace: pod.GetNamespace()}, newPod)
	assert.Nil(t, err)
	assert.Len(t, newPod.GetFinalizers(), 1)
	assert.Len(t, extReconciler.Finalizers, 1)
}

func TestExtendedReconciler_FinalizeOnDeleteErrorOnUpdate(t *testing.T) {
	mockService := BuildMockPlatformService()
	mockService.UpdateFunc = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
		return fmt.Errorf("Foo error")
	}
	extReconciler := BuildTestExtendedReconciler()
	extReconciler.Service = mockService
	extReconciler.Finalizers = map[string]Finalizer{
		"f1": &MockFinalizer{},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "somepod",
			Namespace: "somenamespace",
		},
	}
	pod.SetFinalizers([]string{"f1"})
	pod.SetDeletionTimestamp(&metav1.Time{})
	extReconciler.Service.Create(context.TODO(), pod)

	err := extReconciler.finalizeOnDelete(pod)
	assert.Errorf(t, err, "Foo error")

	newPod := &v1.Pod{}
	err = mockService.Get(context.TODO(), types.NamespacedName{Name: pod.GetName(), Namespace: pod.GetNamespace()}, newPod)
	assert.Nil(t, err)
	assert.Len(t, newPod.GetFinalizers(), 1)
	assert.Len(t, extReconciler.Finalizers, 1)
}

func TestExtendedReconciler_Reconcile(t *testing.T) {
	extReconciler := BuildTestExtendedReconciler()
	var f1Invoked, f2Invoked bool
	f1 := MockFinalizer{name: "f1", onFinalizeFn: func(owner client.Object, service PlatformService) error {
		f1Invoked = true
		return nil
	}}
	f2 := MockFinalizer{name: "f2", onFinalizeFn: func(owner client.Object, service PlatformService) error {
		f2Invoked = true
		return nil
	}}
	extReconciler.Finalizers = map[string]Finalizer{
		"f1": &f1,
		"f2": &f2,
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "somepod",
			Namespace: "somenamespace",
		},
	}
	pod.SetFinalizers([]string{"f1", "f2"})

	extReconciler.Service.Create(context.TODO(), pod)

	request := reconcile.Request{}
	request.Namespace = pod.GetNamespace()
	request.Name = pod.GetName()
	result, err := extReconciler.Reconcile(request)
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{}, result)
	assert.Len(t, pod.GetFinalizers(), 2)
	assert.Len(t, extReconciler.Finalizers, 2)

	extReconciler.Service.Delete(context.TODO(), pod)
	err = extReconciler.Service.Get(context.TODO(), types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}, pod)
	assert.Nil(t, err)

	result, err = extReconciler.Reconcile(request)
	assert.Nil(t, err)
	assert.Equal(t, reconcile.Result{}, result)

	newPod := &v1.Pod{}
	err = extReconciler.Service.Get(context.TODO(), request.NamespacedName, newPod)
	assert.NotNil(t, err)
	assert.Len(t, newPod.GetFinalizers(), 0)
	assert.Len(t, extReconciler.Finalizers, 2)
	assert.True(t, f1Invoked)
	assert.True(t, f2Invoked)
}

func BuildTestExtendedReconciler() ExtendedReconciler {
	service := BuildMockPlatformService()
	reconciler := &MockReconciler{}
	customResource := &v1.Pod{}
	return NewExtendedReconciler(service, reconciler, customResource)
}

func BuildMockPlatformService() *test.MockPlatformService {
	return test.NewMockPlatformServiceBuilder(v1.SchemeBuilder).Build()
}
