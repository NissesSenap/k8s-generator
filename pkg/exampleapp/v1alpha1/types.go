// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// +groupName=platform.example.com
// +versionName=v1alpha1
// +kubebuilder:validation:Required

package v1alpha1

import (
	_ "embed"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate ./crd_gen.sh

//go:embed platform.example.com_exampleapps.yaml
var CRDString string

const Group = "platform.example.com"
const Version = "v1alpha1"
const Kind = "ExampleApp"

//nolint:gochecknoglobals
var GroupVersion = strings.Join([]string{Group, Version}, "/")

type ExampleApp struct {
	// Embedding these structs is required to use controller-gen to produce the CRD
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// +kubebuilder:validation:Enum=production;staging;ephemeral
	Env string `json:"env" yaml:"env"`

	App App `json:"app" yaml:"app"`

	// +optional
	Ingress Ingress `json:"ingress" yaml:"ingress"`
	// +optional
	Overrides Overrides `json:"overrides,omitempty" yaml:"overrides,omitempty"`
}

type App struct {
	AppType []string `json:"appType" yaml:"appType"`
	// +optional
	Image string `json:"image" yaml:"image"`
	// +kubebuilder:validation:Enum=scala;python;npm;rust;go
	Language string `json:"language" yaml:"language"`
	// +optional
	Replicas int `json:"replicas" yaml:"replicas"`
	// +optional
	IamPolicy IamPolicy `json:"iamPolicy" yaml:"iamPolicy"`
}

type IamPolicy struct {
	// +optional
	ServiceAccountProject string `json:"serviceAccountProject" yaml:"serviceAccountProject"`
	// +optional
	ServiceAccount string `json:"serviceAccount" yaml:"serviceAccount"`
}

type Ingress struct {
	URL string `json:"url" yaml:"url"`
	// +optional
	Path string `json:"path" yaml:"path"`
	// +optional
	TLSSecret string `json:"tlsSecret" yaml:"tlsSecret"`
	// +optional
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

// +kubebuilder:validation:MinProperties=1
type Overrides struct {
	// +optional
	AdditionalResources []string `json:"additionalResources,omitempty" yaml:"additionalResources,omitempty"`

	// +optional
	ResourcePatches []string `json:"resourcePatches,omitempty" yaml:"resourcePatches,omitempty"`

	// +optional
	ContainerPatches []string `json:"containerPatches,omitempty" yaml:"containerPatches,omitempty"`
}
