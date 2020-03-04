package test

import (
	"fmt"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetRoutes(count int) []routev1.Route {
	var slice []routev1.Route
	for i := 0; i < count; i++ {
		rte := routev1.Route{
			TypeMeta: metav1.TypeMeta{},
			Spec:     routev1.RouteSpec{},
			Status:   routev1.RouteStatus{},
		}
		rte.Name = fmt.Sprintf("%s%d", "rte", (i + 1))
		slice = append(slice, rte)
	}
	return slice
}

func GetServices(count int) []corev1.Service {
	var slice []corev1.Service
	for i := 0; i < count; i++ {
		svc := corev1.Service{
			TypeMeta: metav1.TypeMeta{},
			Spec:     corev1.ServiceSpec{},
			Status:   corev1.ServiceStatus{},
		}
		svc.Name = fmt.Sprintf("%s%d", "svc", (i + 1))
		slice = append(slice, svc)
	}
	return slice
}

func GetDeploymentConfigs(count int) []oappsv1.DeploymentConfig {
	var slice []oappsv1.DeploymentConfig
	for i := 0; i < count; i++ {
		dc := oappsv1.DeploymentConfig{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: oappsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{},
			},
			Status: oappsv1.DeploymentConfigStatus{
				ReadyReplicas: 0,
			},
		}
		dc.Name = fmt.Sprintf("%s%d", "dc", (i + 1))
		slice = append(slice, dc)
	}
	return slice
}

func GetBuildConfigs(count int) []buildv1.BuildConfig {
	var slice []buildv1.BuildConfig
	for i := 0; i < count; i++ {
		bc := buildv1.BuildConfig{}
		bc.Name = fmt.Sprintf("%s%d", "bc", (i + 1))
		slice = append(slice, bc)
	}
	return slice
}

func GetDeployments(count int) []appsv1.Deployment {
	var slice []appsv1.Deployment
	for i := 0; i < count; i++ {
		dc := appsv1.Deployment{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{},
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 0,
			},
		}
		dc.Name = fmt.Sprintf("%s%d", "deployment", i+1)
		slice = append(slice, dc)
	}
	return slice
}

func GetSecrets(count int) []corev1.Secret {
	var slice []corev1.Secret
	for i := 0; i < count; i++ {
		secret := corev1.Secret{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Data:       map[string][]byte{},
			StringData: map[string]string{},
		}
		secret.Name = fmt.Sprintf("%s%d", "secret", (i + 1))
		slice = append(slice, secret)
	}
	return slice
}

func GetEnvVars(count int, ordered bool) []corev1.EnvVar {
	var slice []corev1.EnvVar
	suffix := 0
	for i := 0; i < count; i++ {
		if ordered {
			suffix = i + 1
		} else {
			suffix = count - i
		}
		env := corev1.EnvVar{Name: fmt.Sprintf("VAR%d", suffix), Value: fmt.Sprintf("value_%d", suffix)}
		slice = append(slice, env)
	}

	return slice
}
