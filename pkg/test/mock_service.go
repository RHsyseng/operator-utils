package test

import (
	"context"

	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	consolev1 "github.com/openshift/api/console/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var log = ctrl.Log.WithName("operatorutils.test")

type MockPlatformService struct {
	Client              clientv1.Client
	scheme              *runtime.Scheme
	CreateFunc          func(ctx context.Context, obj clientv1.Object, opts ...clientv1.CreateOption) error
	DeleteFunc          func(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteOption) error
	GetFunc             func(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error
	ListFunc            func(ctx context.Context, list clientv1.ObjectList, opts ...clientv1.ListOption) error
	UpdateFunc          func(ctx context.Context, obj clientv1.Object, opts ...clientv1.UpdateOption) error
	PatchFunc           func(ctx context.Context, obj clientv1.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error
	DeleteAllOfFunc     func(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteAllOfOption) error
	GetCachedFunc       func(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error
	ImageStreamTagsFunc func(namespace string) imagev1.ImageStreamTagInterface
	GetSchemeFunc       func() *runtime.Scheme
	StatusFunc          func() clientv1.StatusWriter
	SubResourceFunc     func(subResource string) client.SubResourceClient
}

var knownTypes = map[schema.GroupVersion][]runtime.Object{
	corev1.SchemeGroupVersion: {
		&corev1.PersistentVolumeClaim{},
		&corev1.ServiceAccount{},
		&corev1.Secret{},
		&corev1.Service{},
		&corev1.PersistentVolumeClaimList{},
		&corev1.ServiceAccountList{},
		&corev1.ServiceList{}},
	oappsv1.GroupVersion: {
		&oappsv1.DeploymentConfig{},
		&oappsv1.DeploymentConfigList{},
	},
	appsv1.SchemeGroupVersion: {
		&appsv1.StatefulSet{},
		&appsv1.StatefulSetList{},
	},
	routev1.GroupVersion: {
		&routev1.Route{},
		&routev1.RouteList{},
	},
	oimagev1.GroupVersion: {
		&oimagev1.ImageStream{},
		&oimagev1.ImageStreamList{},
	},
	rbacv1.SchemeGroupVersion: {
		&rbacv1.Role{},
		&rbacv1.RoleList{},
		&rbacv1.RoleBinding{},
		&rbacv1.RoleBindingList{},
	},
	buildv1.GroupVersion: {
		&buildv1.BuildConfig{},
		&buildv1.BuildConfigList{},
	},
	consolev1.GroupVersion: {
		&consolev1.ConsoleLink{},
		&consolev1.ConsoleLinkList{},
	},
}

type MockPlatformServiceBuilder struct {
	apiObjects    []runtime.Object
	extraObjects  []runtime.Object
	schemeBuilder *scheme.Builder
}

func NewMockPlatformServiceBuilder(schemeBuilder runtime.SchemeBuilder) *MockPlatformServiceBuilder {
	builder := &scheme.Builder{
		GroupVersion:  corev1.SchemeGroupVersion,
		SchemeBuilder: schemeBuilder,
	}
	return &MockPlatformServiceBuilder{schemeBuilder: builder}
}

func (builder *MockPlatformServiceBuilder) WithScheme(objs ...runtime.Object) {
	builder.apiObjects = objs
}

func (builder *MockPlatformServiceBuilder) WithExtraScheme(objs ...runtime.Object) {
	builder.extraObjects = objs
}

func (builder *MockPlatformServiceBuilder) Build() *MockPlatformService {
	registerObjs := builder.apiObjects
	registerObjs = append(registerObjs, builder.extraObjects...)
	builder.schemeBuilder.Register(registerObjs...)
	scheme, _ := builder.schemeBuilder.Build()
	for gv, types := range knownTypes {
		for _, t := range types {
			scheme.AddKnownTypes(gv, t)
		}
	}
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	log.V(1).Info("Fake client created as %v", client)
	mockImageStreamTag := &MockImageStreamTag{}
	return &MockPlatformService{
		Client: client,
		scheme: scheme,
		CreateFunc: func(ctx context.Context, obj clientv1.Object, opts ...clientv1.CreateOption) error {
			return client.Create(ctx, obj, opts...)
		},
		DeleteFunc: func(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteOption) error {
			return client.Delete(ctx, obj, opts...)
		},
		GetFunc: func(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error {
			return client.Get(ctx, key, obj)
		},
		ListFunc: func(ctx context.Context, list clientv1.ObjectList, opts ...clientv1.ListOption) error {
			return client.List(ctx, list, opts...)
		},
		UpdateFunc: func(ctx context.Context, obj clientv1.Object, opts ...clientv1.UpdateOption) error {
			return client.Update(ctx, obj, opts...)
		},
		PatchFunc: func(ctx context.Context, obj clientv1.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
			return client.Patch(ctx, obj, patch, opts...)
		},
		DeleteAllOfFunc: func(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteAllOfOption) error {
			return client.DeleteAllOf(ctx, obj, opts...)
		},
		GetCachedFunc: func(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error {
			return client.Get(ctx, key, obj)
		},
		ImageStreamTagsFunc: func(namespace string) imagev1.ImageStreamTagInterface {
			return mockImageStreamTag
		},
		GetSchemeFunc: func() *runtime.Scheme {
			return scheme
		},
		StatusFunc: func() clientv1.StatusWriter {
			return client.Status()
		},
		SubResourceFunc: func(subResource string) clientv1.SubResourceClient {
			return client.SubResource(subResource)
		},
	}
}

func (service *MockPlatformService) Create(ctx context.Context, obj clientv1.Object, opts ...clientv1.CreateOption) error {
	return service.CreateFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) Delete(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteOption) error {
	return service.DeleteFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) Get(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error {
	return service.GetFunc(ctx, key, obj)
}

func (service *MockPlatformService) List(ctx context.Context, list clientv1.ObjectList, opts ...clientv1.ListOption) error {
	return service.ListFunc(ctx, list, opts...)
}

func (service *MockPlatformService) Update(ctx context.Context, obj clientv1.Object, opts ...clientv1.UpdateOption) error {
	return service.UpdateFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) Patch(ctx context.Context, obj clientv1.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
	return service.PatchFunc(ctx, obj, patch, opts...)
}

func (service *MockPlatformService) DeleteAllOf(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteAllOfOption) error {
	return service.DeleteAllOfFunc(ctx, obj, opts...)
}

func (service *MockPlatformService) GetCached(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error {
	return service.GetCachedFunc(ctx, key, obj)
}

func (service *MockPlatformService) ImageStreamTags(namespace string) imagev1.ImageStreamTagInterface {
	return service.ImageStreamTagsFunc(namespace)
}

func (service *MockPlatformService) GetScheme() *runtime.Scheme {
	return service.GetSchemeFunc()
}

func (service *MockPlatformService) Status() clientv1.StatusWriter {
	return service.StatusFunc()
}

func (service *MockPlatformService) SubResource(subResource string) clientv1.SubResourceClient {
	return service.SubResourceFunc(subResource)
}

func (service *MockPlatformService) IsMockService() bool {
	return true
}
