package creator

import (
	"github.com/RHsyseng/operator-utils/pkg/webconsole/utils"
	"github.com/ghodss/yaml"
	consolev1 "github.com/openshift/api/console/v1"
)

type ConsoleLinkCreatorImpl struct {
}

func (con *ConsoleLinkCreatorImpl) Create(yamlStr string) (bool, error) {
	client, err := utils.GetClient()
	if err != nil {
		return false, err
	}
	obj := &consolev1.ConsoleLink{}
	err = yaml.Unmarshal([]byte(yamlStr), obj)
	if err != nil {
		return false, err
	}
	_, err = client.ConsoleLink.Create(obj)
	if err != nil {
		return false, err
	}
	return true, nil
}

func NewConsoleLinkCreatorImpl() Creator {
	return new(ConsoleLinkCreatorImpl)
}
