package kubernetes

import (
	"errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func CustomResourceDefinitionExists(gvk schema.GroupVersionKind) (bool, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return false, err
	}
	return customResourceDefinitionExists(gvk, cfg)
}

func customResourceDefinitionExists(gvk schema.GroupVersionKind, cfg *rest.Config) (bool, error) {
	client, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false, err
	}
	api, err := client.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return false, err
	}
	for _, a := range api.APIResources {
		if a.Kind == gvk.Kind {
			return true, nil
		}
	}
	return false, errors.New(gvk.String() + " Kind not found ")
}
