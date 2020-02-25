package webconsole

import (
	"errors"
	creator "github.com/RHsyseng/operator-utils/pkg/webconsole/creator"
	"github.com/RHsyseng/operator-utils/pkg/webconsole/factory"
	"github.com/ghodss/yaml"
)

func ApplyMultipleWebConsoleYamls(yamlString []string) (map[int]string, error) {
	if yamlString == nil || len(yamlString) == 0 {
		return nil, errors.New("Empty yaml list")
	}
	resMap := make(map[int]string)
	for idx, yamlStr := range yamlString {
		err := ApplyWebConsoleYaml(yamlStr)
		if err != nil {
			resMap[idx] = err.Error()
		} else {
			resMap[idx] = "Applied"
		}
	}
	return resMap, nil
}

func ApplyWebConsoleYaml(yamlStr string) error {
	obj := &creator.CustomResourceDefinition{}
	err := yaml.Unmarshal([]byte(yamlStr), obj)
	if err != nil {
		return err
	}
	//check for any non ConsoleYAMLsamples
	creator := factory.GetCreator(obj.Kind)
	if creator == factory.NullCreatorImpl {
		//check for ConsoleYAMLsamples
		kind := obj.Annotations["consolekind"]
		creator = factory.GetCreator(kind)
	}
	_, err = creator.Create(yamlStr)
	if err != nil {
		return err
	} else if creator == factory.NullCreatorImpl {
		return errors.New("Unrecognized web console yaml")
	}
	return nil
}
