package kubernetes

import (
	"errors"
	"github.com/RHsyseng/operator-utils/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

var defaultFn = func() error { return nil }

func TestFinalizerManager_RegisterFinalizer(t *testing.T) {
	var cases = []struct {
		name       string
		finalizers []string
		functions  []func() error
		expected   []string
		shouldFail bool
	}{
		{
			"Register one finalizer",
			[]string{"finalizer1"},
			[]func() error{defaultFn},
			[]string{"finalizer1"},
			false,
		},
		{
			"Register multiple finalizers",
			[]string{"finalizer1", "finalizer2", "finalizer3"},
			[]func() error{defaultFn, defaultFn, defaultFn},
			[]string{"finalizer1", "finalizer2", "finalizer3"},
			false,
		},
		{
			"Repeated finalizers should be replaced",
			[]string{"finalizer1", "finalizer2", "finalizer1"},
			[]func() error{func() error { return errors.New("Unexpected") }, defaultFn, defaultFn},
			[]string{"finalizer1", "finalizer2"},
			false,
		}, {
			"Empty finalizer should produce an error",
			[]string{""},
			[]func() error{defaultFn},
			[]string{},
			true,
		}, {
			"nil function should produce an error",
			[]string{"finalizer1"},
			[]func() error{nil},
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
		client := test.NewMockClient()
		mgr := NewFinalizerManager(&client)
		hasUpdated := false
		client.WhenUpdate(pod, func() error {
			hasUpdated = true
			return nil
		})
		var err error
		for i, f := range c.finalizers {
			err = mgr.RegisterFinalizer(pod, f, c.functions[i])
		}
		if c.shouldFail {
			assert.NotNil(t, err, c.name)
		} else {
			assert.Nil(t, err, c.name)
			assert.Len(t, pod.GetFinalizers(), len(c.expected), c.name)
			for _, e := range c.expected {
				assert.Contains(t, pod.GetFinalizers(), e, c.name)
				assert.Nil(t, mgr.objects[pod.GetUID()][e](), c.name)
			}
			assert.True(t, hasUpdated, c.name)
		}
	}
}

func TestFinalizerManager_RegisterFinalizer_IsFinalizing(t *testing.T) {
	client := test.NewMockClient()
	mgr := NewFinalizerManager(&client)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			DeletionTimestamp: &metav1.Time{},
		},
	}
	client.WhenUpdate(pod, func() error {
		return errors.New("Should not be updated")
	})
	expected := "finalizer.example.com"

	err := mgr.RegisterFinalizer(pod, expected, func() error { return nil })

	assert.Nil(t, err)
	assert.NotContains(t, pod.GetFinalizers(), expected)
}

func TestFinalizerManager_RegisterFinalizer_UpdateFails(t *testing.T) {
	client := test.NewMockClient()
	mgr := NewFinalizerManager(&client)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example",
			UID:  types.UID("134"),
		},
	}
	client.WhenUpdate(pod, func() error {
		return errors.New("Some error")
	})
	expected := "finalizer.example.com"

	err := mgr.RegisterFinalizer(pod, expected, func() error { return nil })

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
		client := test.NewMockClient()
		mgr := FinalizerManager{Client: &client, objects: map[types.UID]ObjectFinalizer{
			pod.GetUID(): {
				"finalizer1": defaultFn,
				"finalizer2": defaultFn,
				"finalizer3": defaultFn,
			},
		}}
		hasUpdated := false
		client.WhenUpdate(pod, func() error {
			hasUpdated = true
			return nil
		})
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
				assert.Nil(t, mgr.objects[pod.GetUID()][e](), c.name)
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
	client := test.NewMockClient()
	mgr := FinalizerManager{Client: &client, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			"finalizer1": defaultFn,
			"finalizer2": defaultFn,
			"finalizer3": defaultFn,
		},
	}}

	client.WhenUpdate(pod, func() error {
		return errors.New("Some error")
	})

	expected := "finalizer1"
	err := mgr.UnregisterFinalizer(pod, expected)

	assert.NotNil(t, err)
	assert.Len(t, mgr.objects[pod.GetUID()], 3)
}

func TestFinalizerManager_Finalize(t *testing.T) {
	count := 0
	onFinalize := func() error {
		count++
		return nil
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			Finalizers:        []string{"finalizer1", "finalizer2", "finalizer3"},
			DeletionTimestamp: &metav1.Time{},
		},
	}
	client := test.NewMockClient()
	mgr := FinalizerManager{Client: &client, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			"finalizer1": onFinalize,
			"finalizer2": onFinalize,
			"finalizer3": onFinalize,
		},
	}}
	updated := 0
	client.WhenUpdate(pod, func() error {
		updated++
		return nil
	})

	err := mgr.FinalizeOnDelete(pod)

	assert.Nil(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, updated)
	assert.Empty(t, mgr.objects[pod.GetUID()])
}

func TestFinalizerManager_Finalize_NotFinalizing(t *testing.T) {
	count := 0
	onFinalize := func() error {
		count++
		return nil
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "example",
			UID:        types.UID("134"),
			Finalizers: []string{"finalizer1", "finalizer2", "finalizer3"},
		},
	}
	client := test.NewMockClient()
	mgr := FinalizerManager{Client: &client, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			"finalizer1": onFinalize,
			"finalizer2": onFinalize,
			"finalizer3": onFinalize,
		},
	}}
	updated := 0
	client.WhenUpdate(pod, func() error {
		updated++
		return nil
	})

	err := mgr.FinalizeOnDelete(pod)

	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, updated)
	assert.NotEmpty(t, mgr.objects[pod.GetUID()])
}

func TestFinalizerManager_Finalize_UpdatingErrors(t *testing.T) {
	count := 0
	onFinalize := func() error {
		if count == 1 {
			return errors.New("Some error")
		}
		count++
		return nil
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			Finalizers:        []string{"finalizer1", "finalizer2", "finalizer3"},
			DeletionTimestamp: &metav1.Time{},
		},
	}
	client := test.NewMockClient()
	mgr := FinalizerManager{Client: &client, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			"finalizer1": onFinalize,
			"finalizer2": onFinalize,
			"finalizer3": onFinalize,
		},
	}}
	updated := 0
	client.WhenUpdate(pod, func() error {
		updated++
		return nil
	})

	err := mgr.FinalizeOnDelete(pod)

	assert.Error(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, updated)
	assert.Len(t, mgr.objects[pod.GetUID()], 2)
}

func TestFinalizerManager_Finalize_FinalizerErrors(t *testing.T) {
	count := 0
	onFinalize := func() error {
		count++
		return nil
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			UID:               types.UID("134"),
			Finalizers:        []string{"finalizer1", "finalizer2", "finalizer3"},
			DeletionTimestamp: &metav1.Time{},
		},
	}
	client := test.NewMockClient()
	mgr := FinalizerManager{Client: &client, objects: map[types.UID]ObjectFinalizer{
		pod.GetUID(): {
			"finalizer1": onFinalize,
			"finalizer2": onFinalize,
			"finalizer3": onFinalize,
		},
	}}
	updated := 0
	client.WhenUpdate(pod, func() error {
		if updated == 1 {
			return errors.New("Some error")
		}
		updated++
		return nil
	})

	err := mgr.FinalizeOnDelete(pod)

	assert.Error(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, 1, updated)
	assert.Len(t, mgr.objects[pod.GetUID()], 2)
}
