package openshift

import (
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

//type MasterType string
//
//const (
//	OpenShift  MasterType = "OpenShift"
//	Kubernetes MasterType = "Kubernetes"
//)

var log = logf.Log.WithName("env")

func IsOpenShift() (bool, error) {
	kubeconfig, err := config.GetConfig()
	if err != nil {
		return false, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeconfig)
	if err != nil {
		return false, err
	}
	apiList, err := discoveryClient.ServerGroups()
	if err != nil {
		return false, err
	}
	apiGroups := apiList.Groups
	log.Info("In IsOpenshift", "apiGroups", apiGroups)
	for i := 0; i < len(apiGroups); i++ {
		if apiGroups[i].Name == "route.openshift.io" {
			log.Info("In IsOpenshift => returning true, nil")
			return true, nil
		}
	}
	return false, nil
}