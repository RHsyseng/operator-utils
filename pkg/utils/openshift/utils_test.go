package openshift

import (
	"github.com/RHsyseng/operator-utils/internal/platform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenShiftVersion_MapKnownVersion(t *testing.T) {

	cases := []struct {
		label              string
		info               platform.PlatformInfo
		expectedOCPVersion string
	}{
		{
			label:              "case 1",
			info:               platform.PlatformInfo{K8SVersion: ""},
			expectedOCPVersion: "",
		},
		{
			label:              "case 2",
			info:               platform.PlatformInfo{K8SVersion: "1.10+"},
			expectedOCPVersion: "v3.10",
		},
		{
			label:              "case 3",
			info:               platform.PlatformInfo{K8SVersion: "1.11+"},
			expectedOCPVersion: "v3.11",
		},
		{
			label:              "case 4",
			info:               platform.PlatformInfo{K8SVersion: "1.13+"},
			expectedOCPVersion: "v4.1",
		},
	}

	for _, v := range cases {
		assert.Equal(t, v.expectedOCPVersion, MapKnownVersion(v.info).Version, v.label+": expected OCP version to match")
	}
}
