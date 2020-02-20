package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

//OnFinalize Callback function to be executed when an object is being deleted
type OnFinalize func() error

type ObjectFinalizer struct {
	object     runtime.Object
	finalizers map[string]OnFinalize
}

type FinalizerManager struct {
	objects map[types.UID]ObjectFinalizer
	Client  client.Client
}

func NewFinalizerManager(client client.Client) FinalizerManager {
	return FinalizerManager{
		objects: map[types.UID]ObjectFinalizer{},
		Client:  client,
	}
}

//RegisterFinalizer registers a finalizer function to be executed when the object is marked to be deleted.
func (mgr *FinalizerManager) RegisterFinalizer(owner metav1.Object, name string, onFinalize OnFinalize) error {
	if mgr.IsFinalizing(owner) {
		return nil
	}
	err := validateFinalizer(name, onFinalize)
	if err != nil {
		return err
	}
	controllerutil.AddFinalizer(owner, name)
	obj, err := toRuntimeObj(owner)
	if err != nil {
		return err
	}
	err = mgr.Client.Update(context.TODO(), obj)
	if err != nil {
		return err
	}
	_, ok := mgr.objects[owner.GetUID()]
	if !ok {
		mgr.objects[owner.GetUID()] = ObjectFinalizer{
			object:     obj,
			finalizers: map[string]OnFinalize{},
		}
	}
	mgr.objects[owner.GetUID()].finalizers[name] = onFinalize
	return nil
}

//UnregisterFinalizer removes a finalizer and updates the owner object
func (mgr *FinalizerManager) UnregisterFinalizer(owner metav1.Object, name string) error {
	err := validateFinalizerName(name)
	if err != nil {
		return err
	}
	obj, err := meta.Accessor(owner)
	controllerutil.RemoveFinalizer(obj, name)
	objFinalizer := mgr.objects[obj.GetUID()]
	err = mgr.Client.Update(context.TODO(), objFinalizer.object)
	if err != nil {
		return err
	}
	delete(mgr.objects[obj.GetUID()].finalizers, name)
	return nil
}

//Finalize triggers all the finalizers registered for the given object
func (mgr *FinalizerManager) Finalize(owner metav1.Object) error {
	if !mgr.IsFinalizing(owner) {
		return nil
	}
	for n, f := range mgr.objects[owner.GetUID()].finalizers {
		err := f()
		if err != nil {
			return err
		}
		err = mgr.UnregisterFinalizer(owner, n)
		if err != nil {
			return err
		}
	}
	return nil
}

//IsFinalizing An object is considered to be finalizing when its deletionTimestamp is not null
func (mgr *FinalizerManager) IsFinalizing(owner metav1.Object) bool {
	return owner.GetDeletionTimestamp() != nil
}

func toRuntimeObj(obj interface{}) (runtime.Object, error) {
	switch t := obj.(type) {
	case runtime.Object:
		return t, nil
	default:
		return nil, fmt.Errorf("object does not implement the runtime.Object interface")
	}
}

func validateFinalizer(name string, onFinalize OnFinalize) error {
	err := validateFinalizerName(name)
	if err != nil {
		return err
	}
	if onFinalize == nil {
		return errors.New("the finalizer OnFinalize function must not be nil")
	}
	return nil
}

func validateFinalizerName(name string) error {
	if len(strings.TrimSpace(name)) == 0 {
		return errors.New("the finalizer name must not be empty")
	}
	return nil
}
