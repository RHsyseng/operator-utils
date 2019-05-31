package openshift

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("env")

// IsOpenShift checks for the OpenShift API
func IsOpenShift() (bool, error) {

	log.Info("Detect if openshift is running")

	value, ok := os.LookupEnv("OPERATOR_OPENSHIFT")
	if ok {
		log.Info("Set by env-var 'OPERATOR_OPENSHIFT': " + value)
		return strings.ToLower(value) == "true", nil
	} else {
		log.Info("env-var lookup failed" + value)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Error getting config: %v")
		return false, err
	}

	groupName := "route.openshift.io"
	gv := schema.GroupVersion{Group: groupName, Version: "v1"}
	cfg.APIPath = "/apis"

	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)

	cfg.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: codecs}

	if cfg.UserAgent == "" {
		cfg.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	cfg.GroupVersion = &gv

	client, err := rest.RESTClientFor(cfg)

	if err != nil {
		log.Error(err, "Error getting client: %v")
		return false, err
	}
	getOpenShiftEnvVersion(cfg.Host)

	_, err = client.Get().DoRaw()

	return err == nil, nil
}
func getOpenShiftEnvVersion(url string) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url+"/version/openshift?timeout=32s", nil)
	if err != nil {
		log.Error(err, "Error to make http call to get openshift version : %v")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err, "Error to make http call to get openshift version : %v")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "Error in reading resp.Body: %v")
		}
		var data map[string]interface{}
		json.Unmarshal(bodyBytes, &data)
		log.Info(fmt.Sprintf("OpenShift Version is %s.%s", data["major"], data["minor"]))

	}

}