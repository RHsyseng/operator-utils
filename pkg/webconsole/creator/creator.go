package creator

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Creator interface {
	Create(yamlStr string) (bool, error)
}

type CustomResourceDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}


