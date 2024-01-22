package kubernetes

import (
	"context"

	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = ctrl.Log.WithName("operatorutils.kubernetes")

type KubernetesPlatformService struct {
	client      clientv1.Client
	cache       cachev1.Cache
	imageClient *imagev1.ImageV1Client
	scheme      *runtime.Scheme
}

func GetInstance(mgr manager.Manager) KubernetesPlatformService {
	imageClient, err := imagev1.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "Error getting image client")
		return KubernetesPlatformService{}
	}

	return KubernetesPlatformService{
		client:      mgr.GetClient(),
		cache:       mgr.GetCache(),
		imageClient: imageClient,
		scheme:      mgr.GetScheme(),
	}
}

func (service *KubernetesPlatformService) Create(ctx context.Context, obj clientv1.Object, opts ...clientv1.CreateOption) error {
	return service.client.Create(ctx, obj, opts...)
}

func (service *KubernetesPlatformService) Delete(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteOption) error {
	return service.client.Delete(ctx, obj, opts...)
}

func (service *KubernetesPlatformService) Get(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error {
	return service.client.Get(ctx, key, obj)
}

func (service *KubernetesPlatformService) List(ctx context.Context, list clientv1.ObjectList, opts ...clientv1.ListOption) error {
	return service.client.List(ctx, list, opts...)
}

func (service *KubernetesPlatformService) Update(ctx context.Context, obj clientv1.Object, opts ...clientv1.UpdateOption) error {
	return service.client.Update(ctx, obj, opts...)
}

func (service *KubernetesPlatformService) Patch(ctx context.Context, obj clientv1.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
	return service.client.Patch(ctx, obj, patch, opts...)
}

func (service *KubernetesPlatformService) DeleteAllOf(ctx context.Context, obj clientv1.Object, opts ...clientv1.DeleteAllOfOption) error {
	return service.client.DeleteAllOf(ctx, obj, opts...)
}

func (service *KubernetesPlatformService) GetCached(ctx context.Context, key clientv1.ObjectKey, obj clientv1.Object) error {
	return service.cache.Get(ctx, key, obj)
}

func (service *KubernetesPlatformService) ImageStreamTags(namespace string) imagev1.ImageStreamTagInterface {
	return service.imageClient.ImageStreamTags(namespace)
}

func (service *KubernetesPlatformService) GetScheme() *runtime.Scheme {
	return service.scheme
}

func (service *KubernetesPlatformService) Status() clientv1.StatusWriter {
	return service.client.Status()
}

func (service *KubernetesPlatformService) SubResource(subResource string) client.SubResourceClient {
	return service.client.SubResource(subResource)
}

func (service *KubernetesPlatformService) IsMockService() bool {
	return false
}
