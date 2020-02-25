package creator

import (
	"github.com/RHsyseng/operator-utils/pkg/webconsole/utils"
	"github.com/ghodss/yaml"
	consolev1 "github.com/openshift/api/console/v1"
)

type ConsoleCliDownloadCreatorImpl struct{
}

func (con *ConsoleCliDownloadCreatorImpl) Create(yamlStr string) (bool, error) {
	client, err := utils.GetClient()
	if err != nil {
		return false, err
	}
	obj := &consolev1.ConsoleCLIDownload{}
	err = yaml.Unmarshal([]byte(yamlStr), obj)
	if err != nil {
		return false, err
	}
	_, err = client.ConsoleCliDownloads.Create(obj)
	if err != nil {
		return false, err
	}
	return true, nil
}

func NewConsoleCliDownloadCreatorImpl() Creator {
	return new(ConsoleCliDownloadCreatorImpl)
}
