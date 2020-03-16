package openshift

import (
	"github.com/ghodss/yaml"
	oappsv1 "github.com/openshift/api/apps/v1"
	v1 "github.com/openshift/api/console/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

func TestGetConsoleYAMLSample(t *testing.T) {
	var inputYaml = `
apiVersion: v1
kind: DeploymentConfig
metadata:
 name: sample-dc
 annotations:
   consoleName: sample-deploymentconfig
   consoleDesc: Sample Deployment Config
   consoleTitle: Sample Deployment Config
spec:
   replicas: 2
`
	original := &oappsv1.DeploymentConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), original))

	yamlSample, err := GetConsoleYAMLSample(original)
	assert.NoError(t, err)

	assert.Equal(t, "sample-deploymentconfig", yamlSample.ObjectMeta.Name)
	assert.Equal(t, "openshift-console", yamlSample.ObjectMeta.Namespace)
	assert.Equal(t, "v1", yamlSample.Spec.TargetResource.APIVersion)
	assert.Equal(t, "DeploymentConfig", yamlSample.Spec.TargetResource.Kind)
	assert.Equal(t, v1.ConsoleYAMLSampleTitle("Sample Deployment Config"), yamlSample.Spec.Title)
	assert.Equal(t, v1.ConsoleYAMLSampleDescription("Sample Deployment Config"), yamlSample.Spec.Description)

	yamlContent := yamlSample.Spec.YAML
	actual := &oappsv1.DeploymentConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(string(yamlContent)), actual))

	original.SetAnnotations(nil)
	assert.EqualValues(t, original, actual, "original yaml should be the same as the actual yaml")
}

func TestGetConsoleYAMLSampleWithNoAnnotations(t *testing.T) {
	var inputYaml = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
`
	original := &appsv1.Deployment{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), original))

	yamlSample, err := GetConsoleYAMLSample(original)
	assert.NoError(t, err)

	assert.Equal(t, "nginx-deployment-yamlsample", yamlSample.ObjectMeta.Name)
	assert.Equal(t, "openshift-console", yamlSample.ObjectMeta.Namespace)
	assert.Equal(t, "apps/v1", yamlSample.Spec.TargetResource.APIVersion)
	assert.Equal(t, "Deployment", yamlSample.Spec.TargetResource.Kind)
	assert.Equal(t, v1.ConsoleYAMLSampleTitle("nginx-deployment-yamlsample"), yamlSample.Spec.Title)
	assert.Equal(t, v1.ConsoleYAMLSampleDescription("nginx-deployment-yamlsample"), yamlSample.Spec.Description)

	yamlContent := yamlSample.Spec.YAML
	actual := &appsv1.Deployment{}
	assert.NoError(t, yaml.Unmarshal([]byte(string(yamlContent)), actual))

	assert.EqualValues(t, original, actual, "original yaml should be the same as the actual yaml")
}
