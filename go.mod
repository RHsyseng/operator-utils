module github.com/RHsyseng/operator-utils

go 1.13

require (
	github.com/coreos/prometheus-operator v0.41.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/spec v0.19.9
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-openapi/validate v0.19.10
	github.com/googleapis/gnostic v0.4.0
	github.com/openshift/api v0.0.0-20200521101457-60c476765272
	github.com/openshift/client-go v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.13.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1

	// OpenShift release-4.5
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200526144822-34f54f12813a
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200521150516-05eb9880269c

	// Pinned to kubernetes-1.18.2
	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.2
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
