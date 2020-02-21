package kubernetes

import (
	"context"
	"errors"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

//OnFinalize Callback function to be executed when an object is being deleted
type OnFinalize func() error

type ObjectFinalizer map[string]OnFinalize

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
func (mgr *FinalizerManager) RegisterFinalizer(owner resource.KubernetesResource, name string, onFinalize OnFinalize) error {
	if mgr.IsFinalizing(owner) {
		return nil
	}
	err := validateFinalizer(name, onFinalize)
	if err != nil {
		return err
	}
	controllerutil.AddFinalizer(owner, name)
	if err != nil {
		return err
	}
	err = mgr.Client.Update(context.TODO(), owner)
	if err != nil {
		return err
	}
	_, ok := mgr.objects[owner.GetUID()]
	if !ok {
		mgr.objects[owner.GetUID()] = map[string]OnFinalize{}
	}
	mgr.objects[owner.GetUID()][name] = onFinalize
	return nil
}

//UnregisterFinalizer removes a finalizer and updates the owner object
func (mgr *FinalizerManager) UnregisterFinalizer(owner resource.KubernetesResource, name string) error {
	err := validateFinalizerName(name)
	if err != nil {
		return err
	}
	obj, err := meta.Accessor(owner)
	controllerutil.RemoveFinalizer(obj, name)
	err = mgr.Client.Update(context.TODO(), owner)
	if err != nil {
		return err
	}
	delete(mgr.objects[obj.GetUID()], name)
	return nil
}

//FinalizeOnDelete triggers all the finalizers registered for the given object in case the owner is being deleted
func (mgr *FinalizerManager) FinalizeOnDelete(owner resource.KubernetesResource) error {
	if !mgr.IsFinalizing(owner) {
		return nil
	}
	for n, f := range mgr.objects[owner.GetUID()] {
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
