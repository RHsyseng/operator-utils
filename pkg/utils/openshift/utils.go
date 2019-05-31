package openshift

import (
	"encoding/json"

	log "k8s.io/klog"
	api "k8s.io/kubernetes/pkg/api/unversioned"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	restclient "k8s.io/kubernetes/pkg/client/restclient"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type MasterType string

const (
	OpenShift  MasterType = "OpenShift"
	Kubernetes MasterType = "Kubernetes"
)

func IsOpenShift() (bool, error) {

	client, _ := NewClient(cmdutil.NewFactory(nil))
	typeOfMaster := TypeOfMaster(client)

	return typeOfMaster == OpenShift, nil
}

func TypeOfMaster(c *clientset.Clientset) MasterType {
	res, err := c.CoreClient.RESTClient().Get().AbsPath("").DoRaw()
	if err != nil {
		log.Fatalf("Could not discover the type of your installation: %v", err)
	}

	var rp api.RootPaths
	err = json.Unmarshal(res, &rp)
	if err != nil {
		log.Fatalf("Could not discover the type of your installation: %v", err)
	}
	for _, p := range rp.Paths {
		if p == "/oapi" {
			return OpenShift
		}
	}
	return Kubernetes
}

func NewClient(f cmdutil.Factory) (*clientset.Clientset, *restclient.Config) {
	var err error

	cfg, err := f.ClientConfig()
	if err != nil {
		log.Error("Could not initialise a client - is your server setting correct?\n\n")
		log.Fatalf("%v", err)
	}

	c, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Could not initialise a client: %v", err)
	}

	return c, cfg
}
