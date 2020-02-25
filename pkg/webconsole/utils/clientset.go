package utils

import (
	"fmt"

	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"

	configv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	consoleclientv1 "github.com/openshift/client-go/console/clientset/versioned/typed/console/v1"
)

// ClientSet is a set of Kubernetes clients.
type ClientSet struct {
	// embedded
	Core                   clientcorev1.CoreV1Interface
	ConsoleCliDownloads    consoleclientv1.ConsoleCLIDownloadInterface
	ConsoleExternalLogLink consoleclientv1.ConsoleExternalLogLinkInterface
	ConsoleLink            consoleclientv1.ConsoleLinkInterface
	ConsoleNotification    consoleclientv1.ConsoleNotificationInterface
	ConsoleYAMLSample      consoleclientv1.ConsoleYAMLSampleInterface
	Console                configv1.ConsolesGetter
}

const (
	TargetNamespace	= "openshift-console"
)

// NewClientset creates a set of Kubernetes clients. The default kubeconfig is
// used if not provided.
func NewClientset(kubeconfig *restclient.Config) (*ClientSet, error) {
	var err error
	if kubeconfig == nil {
		kubeconfig, err = GetConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to get kubeconfig: %s", err)
		}
	}
	clientset := &ClientSet{}
	clientset.Core, err = clientcorev1.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	configClient, err := configv1.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset.Console = configClient
	consoleClient, err := consoleclientv1.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset.ConsoleCliDownloads = consoleClient.ConsoleCLIDownloads()
	clientset.ConsoleExternalLogLink = consoleClient.ConsoleExternalLogLinks()
	clientset.ConsoleLink = consoleClient.ConsoleLinks()
	clientset.ConsoleNotification = consoleClient.ConsoleNotifications()
	clientset.ConsoleYAMLSample = consoleClient.ConsoleYAMLSamples()

	return clientset, nil
}

func MustNewClientset(kubeconfig *restclient.Config) (*ClientSet, error) {
	clientset, err := NewClientset(kubeconfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func GetClient() (*ClientSet, error) {
	client, err := MustNewClientset(nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}