package factory

import (
	"github.com/RHsyseng/operator-utils/pkg/webconsole/creator"
)

const (
	ConsoleYAMLSample      = "ConsoleYAMLSample"
	ConsoleCLIDownload     = "ConsoleCLIDownload"
	ConsoleNotification    = "ConsoleNotification"
	ConsoleExternalLogLink = "ConsoleExternalLogLink"
	ConsoleLink            = "ConsoleLink"
)

var NullCreatorImpl = creator.NewNullCreatorImpl()

func GetCreator(kind string) creator.Creator {
	switch kind {
	case ConsoleYAMLSample:
		return creator.NewConsoleYamlSampleCreatorImpl()
	case ConsoleExternalLogLink:
		return creator.NewConsoleExternalLogLinkCreatorImpl()
	case ConsoleCLIDownload:
		return creator.NewConsoleCliDownloadCreatorImpl()
	case ConsoleLink:
		return creator.NewConsoleLinkCreatorImpl()
	case ConsoleNotification:
		return creator.NewConsoleNotificationCreatorImpl()
	}
	return NullCreatorImpl
}
