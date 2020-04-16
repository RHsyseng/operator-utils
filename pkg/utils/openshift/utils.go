package openshift

import (
	"github.com/RHsyseng/operator-utils/internal/platform"
	"k8s.io/client-go/rest"
)

/*
GetPlatformInfo examines the Kubernetes-based environment and determines the running platform, version, & OS.
Accepts <nil> or instantiated 'cfg' rest config parameter.

Result: PlatformInfo{ Name: OpenShift, K8SVersion: 1.13+, OS: linux/amd64 }
*/
func GetPlatformInfo(cfg *rest.Config) (platform.PlatformInfo, error) {
	return platform.K8SBasedPlatformVersioner{}.GetPlatformInfo(nil, cfg)
}

/*
IsOpenShift is a helper method to simplify boolean OCP checks against GetPlatformInfo results
Accepts <nil> or instantiated 'cfg' rest config parameter.
*/
func IsOpenShift(cfg *rest.Config) (bool, error) {
	info, err := GetPlatformInfo(cfg)
	if err != nil {
		return false, err
	}
	return info.IsOpenShift(), nil
}

/*
Deprecated:
LookupOpenShiftVersion fetches OpenShift version info from API endpoints
*** NOTE: OCP 4.1+ requires elevated user permissions, see PlatformVersioner for details
Accepts <nil> or instantiated 'cfg' rest config parameter.

Result: OpenShiftVersion{ Version: 4.1.2 }
*/
func LookupOpenShiftVersion(cfg *rest.Config) (platform.OpenShiftVersion, error) {
	return platform.K8SBasedPlatformVersioner{}.LookupOpenShiftVersion(nil, cfg)
}

/*
Compare the runtime OpenShift with the version passed in.
version: Semantic format
cfg : OpenShift platform config, use runtime config if nil is passed in.
return:
	-1 : if ver1 <  OpenShiftVersion
	 0 : if ver1 == OpenShiftVersion
     1 : if ver1 > OpenShiftVersion
The int value returned should be discarded if err is not nil
*/
func CompareOpenShiftVersion(version string) (int, error) {
	return platform.K8SBasedPlatformVersioner{}.CompareOpenShiftVersion(version)
}

/*
* return MajorMinor format, e.g. v4.4
*/
func GetOpenShiftVersion() (string, error) {
	return platform.K8SBasedPlatformVersioner{}.GetOpenShiftVersion()
}

/*
MapKnownVersion maps from K8S version of PlatformInfo to equivalent OpenShift version

Result: OpenShiftVersion{ Version: v4.1 }
*/
func MapKnownVersion(info platform.PlatformInfo) platform.OpenShiftVersion {
	return platform.MapKnownVersion(info)
}
