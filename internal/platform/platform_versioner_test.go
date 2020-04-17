package platform

import (
	"testing"

	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"
)

type FakeDiscoverer struct {
	info               PlatformInfo
	serverInfo         *version.Info
	groupList          *v1.APIGroupList
	doc                *openapi_v2.Document
	client             rest.Interface
	ServerVersionError error
	ServerGroupsError  error
	OpenAPISchemaError error
}

func (d FakeDiscoverer) ServerVersion() (*version.Info, error) {
	if d.ServerVersionError != nil {
		return nil, d.ServerVersionError
	}
	return d.serverInfo, nil
}

func (d FakeDiscoverer) ServerGroups() (*v1.APIGroupList, error) {
	if d.ServerGroupsError != nil {
		return nil, d.ServerGroupsError
	}
	return d.groupList, nil
}

func (d FakeDiscoverer) OpenAPISchema() (*openapi_v2.Document, error) {
	if d.OpenAPISchemaError != nil {
		return nil, d.OpenAPISchemaError
	}
	return d.doc, nil
}

func (d FakeDiscoverer) RESTClient() rest.Interface {
	return d.client
}

type FakePlatformVersioner struct {
	Info PlatformInfo
	Err  error
}

func (pv FakePlatformVersioner) GetPlatformInfo(d Discoverer, cfg *rest.Config) (PlatformInfo, error) {
	if pv.Err != nil {
		return pv.Info, pv.Err
	}
	return pv.Info, nil
}

func TestK8SBasedPlatformVersioner_GetPlatformInfo(t *testing.T) {

	pv := K8SBasedPlatformVersioner{}
	fakeErr := errors.New("uh oh")

	cases := []struct {
		label        string
		discoverer   Discoverer
		config       *rest.Config
		expectedInfo PlatformInfo
		expectedErr  bool
	}{
		{
			label: "case 1", // trigger error in client.ServerVersion(), only Name present on Info
			discoverer: FakeDiscoverer{
				ServerVersionError: fakeErr,
			},
			config:       &rest.Config{},
			expectedInfo: PlatformInfo{Name: Kubernetes},
			expectedErr:  true,
		},
		{
			label: "case 2", // trigger error in client.ServerGroups(), K8S major/minor now present
			discoverer: FakeDiscoverer{
				ServerGroupsError: fakeErr,
				serverInfo: &version.Info{
					Major: "1",
					Minor: "2",
				},
			},
			config:       &rest.Config{},
			expectedInfo: PlatformInfo{Name: Kubernetes, K8SVersion: "1.2"},
			expectedErr:  true,
		},
		{
			label: "case 3", // trigger no errors, simulate K8S platform (no OCP route present)
			discoverer: FakeDiscoverer{
				serverInfo: &version.Info{
					Major: "1",
					Minor: "2",
				},
				groupList: &v1.APIGroupList{
					TypeMeta: v1.TypeMeta{},
					Groups:   []v1.APIGroup{},
				},
			},
			config:       &rest.Config{},
			expectedInfo: PlatformInfo{Name: Kubernetes, K8SVersion: "1.2"},
			expectedErr:  false,
		},
		{
			label: "case 4", // trigger no errors, simulate OCP route present
			discoverer: FakeDiscoverer{
				serverInfo: &version.Info{
					Major: "1",
					Minor: "2",
				},
				groupList: &v1.APIGroupList{
					TypeMeta: v1.TypeMeta{},
					Groups:   []v1.APIGroup{{Name: "route.openshift.io"}},
				},
			},
			config:       &rest.Config{},
			expectedInfo: PlatformInfo{Name: OpenShift, K8SVersion: "1.2"},
			expectedErr:  false,
		},
	}

	for _, c := range cases {
		info, err := pv.GetPlatformInfo(c.discoverer, c.config)
		assert.Equal(t, c.expectedInfo, info, c.label+": mismatch in returned PlatformInfo")
		if c.expectedErr {
			assert.Error(t, err, c.label+": expected error, but none occurred")
		}
	}
}

func TestClientCallVersionComparsion(t *testing.T) {
	pv := K8SBasedPlatformVersioner{}
	testcases := []struct {
		label        string
		discoverer   Discoverer
		config       []*rest.Config
		expectedInfo int
		expectedErr  bool
	}{
		{
			label: "case 1",
			discoverer: FakeDiscoverer{
				serverInfo: &version.Info{
					Major: "1",
					Minor: "16",
				},
				groupList: &v1.APIGroupList{
					TypeMeta: v1.TypeMeta{},
					Groups:   []v1.APIGroup{{Name: "route.openshift.io"}},
				},
			},
			config:       []*rest.Config{&rest.Config{}},
			expectedInfo: 0,
			expectedErr:  false,
		},
		{
			label: "case 2",
			discoverer: FakeDiscoverer{
				serverInfo: &version.Info{
					Major: "1",
					Minor: "16+",
				},
				groupList: &v1.APIGroupList{
					TypeMeta: v1.TypeMeta{},
					Groups:   []v1.APIGroup{{Name: "route.openshift.io"}},
				},
			},
			config:       []*rest.Config{&rest.Config{}},
			expectedInfo: 0,
			expectedErr:  false,
		},
		{
			label: "case 3",
			discoverer: FakeDiscoverer{
				serverInfo: &version.Info{
					Major: "1",
					Minor: "14+",
				},
				groupList: &v1.APIGroupList{
					TypeMeta: v1.TypeMeta{},
					Groups:   []v1.APIGroup{{Name: "route.openshift.io"}},
				},
			},
			config:       []*rest.Config{&rest.Config{}},
			expectedInfo: -1,
			expectedErr:  false,
		},
		{
			label: "case 4",
			discoverer: FakeDiscoverer{
				serverInfo: &version.Info{
					Major: "1",
					Minor: "14not",
				},
				groupList: &v1.APIGroupList{
					TypeMeta: v1.TypeMeta{},
					Groups:   []v1.APIGroup{{Name: "route.openshift.io"}},
				},
			},
			config:       []*rest.Config{&rest.Config{}},
			expectedInfo: -1,
			expectedErr:  false,
		},
		{
			label: "case 5",
			discoverer: FakeDiscoverer{
				serverInfo: &version.Info{
					Major: "1",
					Minor: "10",
				},
				groupList: &v1.APIGroupList{
					TypeMeta: v1.TypeMeta{},
					Groups:   []v1.APIGroup{{Name: "route.openshift.io"}},
				},
			},
			config:       []*rest.Config{&rest.Config{}},
			expectedInfo: -1,
			expectedErr:  false,
		},
	}

	versionToTest := "v4.3"
	for _, tc := range testcases {
		res, err := pv.CompareOpenShiftVersions(tc.discoverer, versionToTest, tc.config)
		if tc.expectedErr {
			assert.Error(t, err, "expected error-"+tc.label)
		} else {
			assert.NoError(t, err, "unexpected error-"+tc.label)
		}
		assert.Equal(t, tc.expectedInfo, res, "The expected and actual versions should be the same."+tc.label)
	}
}

func TestCompareVersions(t *testing.T) {
	pv := K8SBasedPlatformVersioner{}

	//v4.03 -> ""
	res := pv.CompareVersions("v4.3", "v4.03")
	assert.Equal(t, 1, res)

	res = pv.CompareVersions("v4.3", "v4.3")
	assert.Equal(t, 0, res)

	res = pv.CompareVersions("v4.3", "v4.1")
	assert.Equal(t, 1, res)

	res = pv.CompareVersions("v4.3", "v4.4")
	assert.Equal(t, -1, res)

	// "" vs ""
	res = pv.CompareVersions("", "v4.03")
	assert.Equal(t, 0, res)

	res = pv.CompareVersions("v4.3", "")
	assert.Equal(t, 1, res)
}
