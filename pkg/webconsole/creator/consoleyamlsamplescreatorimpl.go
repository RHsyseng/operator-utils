package creator

import (
	"errors"
	"github.com/ghodss/yaml"
	"github.com/RHsyseng/operator-utils/pkg/webconsole/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"

	consolev1 "github.com/openshift/api/console/v1"
)

type ConsoleYamlSampleCreatorImpl struct {
}

func (con *ConsoleYamlSampleCreatorImpl) Create(yamlStr string) (bool, error) {
	client, err := utils.GetClient()
	if err != nil {
		return false, err
	}
	obj := &CustomResourceDefinition{}
	err = yaml.Unmarshal([]byte(yamlStr), obj)
	if err != nil {
		return false, err
	}

	snippetStr := obj.Annotations["consolesnippet"]
	var snippet bool = false
	if snippetStr != "" {
		tmp, err := strconv.ParseBool(snippetStr)
		if err != nil {
			return false, errors.New("Unable to parse snippet as boolean.")
		}
		snippet = tmp
	}
	title, _ := obj.Annotations["consoletitle"]
	desc, _ := obj.Annotations["consoledesc"]
	name, _ := obj.Annotations["consolename"]

	yamlSample := &consolev1.ConsoleYAMLSample{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: utils.TargetNamespace,
		},
		Spec: consolev1.ConsoleYAMLSampleSpec{
			TargetResource: v1.TypeMeta{
				APIVersion: obj.TypeMeta.APIVersion,
				Kind:       obj.TypeMeta.Kind,
			},
			Title:       consolev1.ConsoleYAMLSampleTitle(title),
			Description: consolev1.ConsoleYAMLSampleDescription(desc),
			YAML:        consolev1.ConsoleYAMLSampleYAML(yamlStr),
			Snippet:     snippet,
		},
	}
	yamlSample, err = client.ConsoleYAMLSample.Create(yamlSample)
	if err != nil {
		return false, err
	}
	return true, nil
}

func NewConsoleYamlSampleCreatorImpl() Creator {
	return new(ConsoleYamlSampleCreatorImpl)
}
