package kubernetes

import (
	"context"
	"errors"
	"github.com/RHsyseng/operator-utils/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type MockFinalizer struct {
	name         string
	onFinalizeFn func() error
}

func (m *MockFinalizer) getName() string {
	return m.name
}

func (m *MockFinalizer) onFinalize() error {
	if m.onFinalizeFn == nil {
		return nil
	}
	return m.onFinalizeFn()
}

func (m *MockFinalizer) setOnFinalizeFn(onFinalizeFn func() error) {
	m.onFinalizeFn = onFinalizeFn
}

func NewMockFinalizer(name string) *MockFinalizer {
	return &MockFinalizer{name, func() error { return nil }}
}

func NewMockFinalizerWithError(name string) *MockFinalizer {
	return &MockFinalizer{name, func() error { return errors.New("Mock error") }}
}

func TestFinalizerManager_RegisterFinalizer(t *testing.T) {
	var cases = []struct {
		name       string
		finalizers []Finalizer
		expected   []string
		shouldFail bool
	}{
		{
			"Register one finalizer",
			[]Finalizer{NewMockFinalizer("finalizer1")},
			[]string{"finalizer1"},
			false,
		},
		{
			"Register multiple finalizers",
			[]Finalizer{NewMockFinalizer("finalizer1"), NewMockFinalizer("finalizer2"), NewMockFinalizer("finalizer3")},
			[]string{"finalizer1", "finalizer2", "finalizer3"},
			false,
		},
		{
			"Repeated finalizers should be replaced",
			[]Finalizer{NewMockFinalizerWithError("finalizer1"), NewMockFinalizer("finalizer2"), NewMockFinalizer("finalizer1")},
			[]string{"finalizer1", "finalizer2"},
			false,
		}, {
			"Empty finalizer should produce an error",
			[]Finalizer{NewMockFinalizer("")},
			[]string{},
			true,
		},
	}

	for _, c := range cases {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "example",
				UID:  types.UID("134"),
			},
		}
		mockPlatformSvc := BuildMockPlatformService()
		mockPlatformSvc.Create(context.TODO(), pod)
		mgr := NewFinalizerManager(mockPlatformSvc)
		hasUpdated := false
		mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
			hasUpdated = true
			return mockPlatformSvc.Client.Update(ctx, obj, opts...)
		}
		var err error
		for _, f := range c.finalizers {
			err = mgr.RegisterFinalizer(pod, f)
		}
		if c.shouldFail {
			assert.NotNil(t, err, c.name)
		} else {
			assert.Nil(t, err, c.name)
			assert.Len(t, pod.GetFinalizers(), len(c.expected), c.name)
			for _, e := range c.expected {
				assert.Contains(t, pod.GetFinalizers(), e, c.name)
				assert.Nil(t, mgr.objects[pod.GetUID()][e].onFinalize(), c.name)
			}
			assert.True(t, hasUpdated, c.name)
		}
	}
}

func TestFinalizerManager_RegisterFinalizer_IsFinalizing(t *testing.T) {
	mockPlatformSvc := BuildMockPlatformService()
	mgr := NewFinalizerManager(mockPlatformSvc)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			DeletionTimestamp: &metav1.Time{},
		},
	}
	mockPlatformSvc.Create(context.TODO(), pod)
	mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
		return errors.New("Should not be updated")
	}
	expected := NewMockFinalizer("finalizer.example.com")

	err := mgr.RegisterFinalizer(pod, expected)

	assert.Nil(t, err)
	assert.NotContains(t, pod.GetFinalizers(), expected)
}

func TestFinalizerManager_RegisterFinalizer_UpdateFails(t *testing.T) {
	mockPlatformSvc := BuildMockPlatformService()
	mgr := NewFinalizerManager(mockPlatformSvc)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example",
			UID:  types.UID("134"),
		},
	}
	mockPlatformSvc.Create(context.TODO(), pod)
	mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
		return errors.New("Some error")
	}
	expected := NewMockFinalizer("finalizer.example.com")

	err := mgr.RegisterFinalizer(pod, expected)

	assert.NotNil(t, err)
	assert.NotContains(t, mgr.objects[pod.GetUID()], expected)
}

func TestFinalizerManager_UnregisterFinalizer(t *testing.T) {
	var cases = []struct {
		name       string
		finalizers []string
		expected   []string
		shouldFail bool
	}{
		{
			"Unregister one finalizer",
			[]string{"finalizer1"},
			[]string{"finalizer2", "finalizer3"},
			false,
		},
		{
			"Unregister multiple finalizers",
			[]string{"finalizer3", "finalizer2"},
			[]string{"finalizer1"},
			false,
		},
		{
			"Unregister all finalizers",
			[]string{"finalizer1", "finalizer2", "finalizer3"},
			[]string{},
			false,
		},
		{
			"Missing finalizers should be ignored",
			[]string{"finalizer1", "finalizer4"},
			[]string{"finalizer2", "finalizer3"},
			false,
		}, {
			"Empty finalizer should produce an error",
			[]string{""},
			[]string{},
			true,
		},
	}

	for _, c := range cases {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "example",
				UID:        types.UID("134"),
				Finalizers: []string{"finalizer1", "finalizer2", "finalizer3"},
			},
		}
		mockPlatformSvc := BuildMockPlatformService()
		mockPlatformSvc.Create(context.TODO(), pod)
		mgr := FinalizerManager{PlatformService: mockPlatformSvc, objects: map[types.UID]ObjectFinalizer{
			pod.GetUID(): {
				"finalizer1": NewMockFinalizer("finalizer1"),
				"finalizer2": NewMockFinalizer("finalizer2"),
				"finalizer3": NewMockFinalizer("finalizer3"),
			},
		}}
		hasUpdated := false
		mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
			hasUpdated = true
			return mockPlatformSvc.Client.Update(ctx, obj, opts...)
		}
		var err error
		for _, f := range c.finalizers {
			err = mgr.UnregisterFinalizer(pod, f)
		}
		if c.shouldFail {
			assert.NotNil(t, err, c.name)
		} else {
			assert.Nil(t, err, c.name)
			assert.Len(t, pod.GetFinalizers(), len(c.expected), c.name)
			for _, e := range c.expected {
				assert.Contains(t, pod.GetFinalizers(), e, c.name)
				assert.Nil(t, mgr.objects[pod.GetUID()][e].onFinalize(), c.name)
			}
			assert.True(t, hasUpdated, c.name)
		}
	}
}

func TestFinalizerManager_UnregisterFinalizer_UpdateFails(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "example",
			UID:        types.UID("134"),
			Finalizers: []string{"finalizer1", "finalizer2", "finalizer3"},
		},
	}
	mockPlatformSvc := BuildMockPlatformService()
	mockPlatformSvc.Create(context.TODO(), pod)
	mgr := FinalizerManager{PlatformService: mockPlatformSvc, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			"finalizer1": NewMockFinalizer("finalizer1"),
			"finalizer2": NewMockFinalizer("finalizer2"),
			"finalizer3": NewMockFinalizer("finalizer3"),
		},
	}}

	mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
		return errors.New("Some error")
	}

	expected := "finalizer1"
	err := mgr.UnregisterFinalizer(pod, expected)

	assert.NotNil(t, err)
	assert.Len(t, mgr.objects[pod.GetUID()], 3)
}

func TestFinalizerManager_FinalizeOnDelete(t *testing.T) {
	count := 0
	finalizers := []*MockFinalizer{
		NewMockFinalizer("finalizer1"),
		NewMockFinalizer("finalizer2"),
		NewMockFinalizer("finalizer3"),
	}
	onFinalize := func() error {
		count++
		return nil
	}
	for _, f := range finalizers {
		f.setOnFinalizeFn(onFinalize)
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			Finalizers:        []string{"finalizer1", "finalizer2", "finalizer3"},
			DeletionTimestamp: &metav1.Time{},
		},
	}
	mockPlatformSvc := BuildMockPlatformService()
	mockPlatformSvc.Create(context.TODO(), pod)
	mgr := FinalizerManager{PlatformService: mockPlatformSvc, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			finalizers[0].getName(): finalizers[0],
			finalizers[1].getName(): finalizers[1],
			finalizers[2].getName(): finalizers[2],
		},
	}}
	updated := 0
	mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
		updated++
		return mockPlatformSvc.Client.Update(ctx, obj, opts...)
	}

	err := mgr.FinalizeOnDelete(pod)

	assert.Nil(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, updated)
	assert.Empty(t, mgr.objects[pod.GetUID()])
}

func TestFinalizerManager_FinalizeOnDelete_NotFinalizing(t *testing.T) {
	count := 0
	finalizers := []*MockFinalizer{
		NewMockFinalizer("finalizer1"),
		NewMockFinalizer("finalizer2"),
		NewMockFinalizer("finalizer3"),
	}
	onFinalize := func() error {
		count++
		return nil
	}
	for _, f := range finalizers {
		f.onFinalizeFn = onFinalize
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "example",
			UID:        types.UID("134"),
			Finalizers: []string{"finalizer1", "finalizer2", "finalizer3"},
		},
	}
	mockPlatformSvc := BuildMockPlatformService()
	mockPlatformSvc.Create(context.TODO(), pod)
	updated := 0
	mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
		updated++
		return mockPlatformSvc.Client.Update(ctx, obj, opts...)
	}
	mgr := FinalizerManager{PlatformService: mockPlatformSvc, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			finalizers[0].getName(): finalizers[0],
			finalizers[1].getName(): finalizers[1],
			finalizers[2].getName(): finalizers[2],
		},
	}}

	err := mgr.FinalizeOnDelete(pod)

	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, updated)
	assert.NotEmpty(t, mgr.objects[pod.GetUID()])
}

func TestFinalizerManager_FinalizeOnDelete_UpdatingErrors(t *testing.T) {
	count := 0
	finalizers := []*MockFinalizer{
		NewMockFinalizer("finalizer1"),
		NewMockFinalizer("finalizer2"),
		NewMockFinalizer("finalizer3"),
	}
	onFinalize := func() error {
		if count == 1 {
			return errors.New("Some error")
		}
		count++
		return nil
	}
	for _, f := range finalizers {
		f.onFinalizeFn = onFinalize
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			Finalizers:        []string{"finalizer1", "finalizer2", "finalizer3"},
			DeletionTimestamp: &metav1.Time{},
		},
	}
	mockPlatformSvc := BuildMockPlatformService()
	mockPlatformSvc.Create(context.TODO(), pod)
	mgr := FinalizerManager{PlatformService: mockPlatformSvc, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			finalizers[0].getName(): finalizers[0],
			finalizers[1].getName(): finalizers[1],
			finalizers[2].getName(): finalizers[2],
		},
	}}
	updated := 0
	mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
		updated++
		return mockPlatformSvc.Client.Update(ctx, obj, opts...)
	}

	err := mgr.FinalizeOnDelete(pod)

	assert.Error(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, updated)
	assert.Len(t, mgr.objects[pod.GetUID()], 2)
}

func TestFinalizerManager_FinalizeOnDelete_FinalizerErrors(t *testing.T) {
	count := 0
	finalizers := []*MockFinalizer{
		NewMockFinalizer("finalizer1"),
		NewMockFinalizer("finalizer2"),
		NewMockFinalizer("finalizer3"),
	}
	onFinalize := func() error {
		count++
		return nil
	}
	for _, f := range finalizers {
		f.onFinalizeFn = onFinalize
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			Finalizers:        []string{"finalizer1", "finalizer2", "finalizer3"},
			DeletionTimestamp: &metav1.Time{},
		},
	}
	mockPlatformSvc := BuildMockPlatformService()
	mockPlatformSvc.Create(context.TODO(), pod)
	mgr := FinalizerManager{PlatformService: mockPlatformSvc, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			finalizers[0].getName(): finalizers[0],
			finalizers[1].getName(): finalizers[1],
			finalizers[2].getName(): finalizers[2],
		},
	}}
	updated := 0
	mockPlatformSvc.UpdateFunc = func(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
		if updated == 1 {
			return errors.New("Some error")
		}
		updated++
		return mockPlatformSvc.Client.Update(ctx, obj, opts...)
	}

	err := mgr.FinalizeOnDelete(pod)

	assert.Error(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, 1, updated)
	assert.Len(t, mgr.objects[pod.GetUID()], 2)
}

func BuildMockPlatformService() *test.MockPlatformService {
	return test.NewMockPlatformServiceBuilder(v1.SchemeBuilder).Build()
}
