package openshift

import (
	"github.com/RHsyseng/operator-utils/internal/platform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestK8SVersionHelpers(t *testing.T) {

	ocpTestVersions := []struct {
		version string
		major   string
		minor   string
	}{
		{"1.11+", "1", "11+"},
		{"1.13+", "1", "13+"},
	}

	for _, v := range ocpTestVersions {

		info := platform.PlatformInfo{K8SVersion: v.version}
		assert.Equal(t, v.major, info.K8SMajorVersion(), "K8SMajorVersion mismatch")
		assert.Equal(t, v.minor, info.K8SMinorVersion(), "K8SMinorVersion mismatch")
	}
}

func TestPlatformInfo_String(t *testing.T) {

	info := platform.PlatformInfo{Name: platform.OpenShift, K8SVersion: "456", OS: "foo/bar"}

	assert.Equal(t, "PlatformInfo [Name: OpenShift, K8SVersion: 456, OS: foo/bar]",
		info.String(), "PlatformInfo String() yields malformed result of %s", info.String())
}

func TestVersionHelpers(t *testing.T) {

	ocpTestVersions := []struct {
		version    string
		major      string
		minor      string
		prerelease string
		build      string
	}{
		{"v3.11.0+69", "v3", "11", "", "+69"},
		{"v4.1.0-0+rc.1", "v4", "1", "-0", "+rc.1"},
		{"v1.2.0+3.4.5.6", "v1", "2", "", "+3.4.5.6"},
		{"v1.2", "v1", "2", "", ""},
		{"v1.2.3-prerel", "v1", "2", "-prerel", ""},
	}

	for _, v := range ocpTestVersions {

		info := OpenShiftVersion{Version: v.version}
		assert.Equal(t, v.major, info.MajorVersion(), "OCPMajorVersion mismatch")
		assert.Equal(t, v.minor, info.MinorVersion(), "OCPMinorVersion mismatch")
		assert.Equal(t, v.prerelease, info.PrereleaseVersion(), "OCPPrereleaseVersion mismatch")
		assert.Equal(t, v.build, info.BuildVersion(), "OCPBuildVersion mismatch")
	}
}

func TestOpenShiftVersion_String(t *testing.T) {

	info := OpenShiftVersion{Version: "1.1.1+"}

	assert.Equal(t, "OpenShiftVersion [Version: 1.1.1+]",
		info.String(), "OpenShiftVersion String() yields malformed result of %s", info.String())
}

func TestVersionComparsion(t *testing.T) {
	targetVersions := []OpenShiftVersion{
		OpenShiftVersion{Version: "v3.11"},
		OpenShiftVersion{Version: "v4.1"},
		OpenShiftVersion{Version: "v4.3"},
		OpenShiftVersion{Version: "v4.4"},
		OpenShiftVersion{Version: "v4.3fail"},
	}

	currOCPVersion := OpenShiftVersion{Version: "v4.3"}
	res := currOCPVersion.Compare(targetVersions[0])
	assert.Equal(t, 1, res, "cur. ocp version should be bigger than target.")

	res = currOCPVersion.Compare(targetVersions[1])
	assert.Equal(t, 1, res, "cur. ocp version should be bigger than target.")

	res = currOCPVersion.Compare(targetVersions[2])
	assert.Equal(t, 0, res, "cur. ocp version should be the same as target.")

	res = currOCPVersion.Compare(targetVersions[3])
	assert.Equal(t, -1, res, "cur. ocp version should be smaller than target.")

	res = currOCPVersion.Compare(targetVersions[4])
	assert.Equal(t, 1, res, "cur. ocp version should be greater than target.")
}
