package openshift

import (
	"github.com/RHsyseng/operator-utils/internal/platform"
	"golang.org/x/mod/semver"
	"strings"
)

type OpenShiftVersion struct {
	Version string `json:"ocpVersion"`
}

func (info OpenShiftVersion) MajorVersion() string {
	return semver.Major(info.Version)
}

func (info OpenShiftVersion) MinorVersion() string {
	ver := semver.MajorMinor(info.Version)
	if ver != "" {
		return strings.Split(info.Version, ".")[1]
	}
	return ""
}

func (info OpenShiftVersion) PrereleaseVersion() string {
	return semver.Prerelease(info.Version)
}

func (info OpenShiftVersion) BuildVersion() string {
	return semver.Build(info.Version)
}

func (info OpenShiftVersion) String() string {
	return "OpenShiftVersion [" +
		"Version: " + info.Version + "]"
}

func (v OpenShiftVersion) Compare(o OpenShiftVersion) int {
	return semver.Compare(v.Version, o.Version)
}

// full generated 'version' API fetch result struct @
// gist.github.com/jeremyary/5a66530611572a057df7a98f3d2902d5
type PlatformClusterInfo struct {
	Status struct {
		Desired struct {
			Version string `json:"version"`
		} `json:"desired"`
	} `json:"status"`
}

/*
MapKnownVersion maps from K8S version of PlatformInfo to equivalent OpenShift version

Result: OpenShiftVersion{ Version: 4.1.2 }
*/
func K8SOpenshiftVersionMap(info platform.PlatformInfo) OpenShiftVersion {
	k8sToOcpMap := map[string]string{
		"1.10+": "v3.10",
		"1.10":  "v3.10",
		"1.11+": "v3.11",
		"1.11":  "v3.11",
		"1.13+": "v4.1",
		"1.13":  "v4.1",
		"1.14+": "v4.2",
		"1.14":  "v4.2",
		"1.16+": "v4.3",
		"1.16":  "v4.3",
	}
	return OpenShiftVersion{Version: semver.MajorMinor(k8sToOcpMap[info.K8SVersion])}
}