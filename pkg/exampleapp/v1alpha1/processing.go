// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"bytes"
	"embed"
	"fmt"
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

//go:embed templates/*
var templateFS embed.FS

func (a *ExampleApp) Schema() (*spec.Schema, error) {
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk(Group, Version, Kind), CRDString)
	return schema, errors.WrapPrefixf(err, "parsing %s schema", Kind)
}

func (a *ExampleApp) Default() error {
	if a.App.Image == "" {
		a.App.Image = fmt.Sprintf("registry.example.com/path/to/%s", a.ObjectMeta.Name)
	}
	switch a.Env {
	case "production":
		a.App.Replicas = 3
	case "staging":
		a.App.Replicas = 1
	}

	return nil
}

func (a *ExampleApp) Validate() error {
	if !strings.HasSuffix(a.Ingress.Domain, "example.com") && !strings.HasSuffix(a.Ingress.Domain, "example.io") {
		return errors.Errorf("ingress %q must be in example.com or example.io", a.Ingress.Domain)
	}
	return nil
}

func (a ExampleApp) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	templates := make([]framework.ResourceTemplate, 0)
	templates = append(templates, framework.ResourceTemplate{
		Templates:    parser.TemplateFiles("templates/app.template.yaml").FromFS(templateFS),
		TemplateData: a.appTemplateData(a.App),
	})

	var patches []framework.PatchTemplate

	if len(a.Overrides.AdditionalResources) > 0 {
		templates = append(templates, framework.ResourceTemplate{
			Templates:    parser.TemplateFiles(a.Overrides.AdditionalResources...).WithExtensions(".yaml", ".template.yaml"),
			TemplateData: a,
		})
	}

	for i, resource := range a.Overrides.ResourcePatches {
		overridePatches, err := a.resourceSMPsFromOverrides(resource, i, patches)
		if err != nil {
			return nil, err
		}
		patches = append(patches, overridePatches...)
	}

	if len(a.Overrides.ContainerPatches) > 0 {
		patches = append(patches, framework.PatchTemplate(&framework.ContainerPatchTemplate{
			Templates:    parser.TemplateFiles(a.Overrides.ContainerPatches...).WithExtensions(".yaml", ".template.yaml"),
			TemplateData: a,
		}))
	}

	items, err := framework.TemplateProcessor{
		ResourceTemplates: templates,
		PatchTemplates:    patches,
	}.Filter(items)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "processing templates")
	}

	return items, nil
}

// resourceSMPsFromOverrides parses the resource template and returns a patch that
// is targeted to match resources with the same GVKNN the patch itself contains.
// TODO: This is standard SMP semantics, so the framework should make this easier.
func (a ExampleApp) resourceSMPsFromOverrides(resource string, i int, patches []framework.PatchTemplate) ([]framework.PatchTemplate, error) {
	tpl, err := parser.TemplateFiles(resource).WithExtensions(".yaml", ".template.yaml").Parse()
	if err != nil {
		return nil, errors.WrapPrefixf(err, "parsing resource template %d", i)
	}
	for _, template := range tpl {
		var b bytes.Buffer
		if err := template.Execute(&b, a); err != nil {
			return nil, errors.WrapPrefixf(err, "failed to render patch template %v", template.DefinedTemplates())
		}
		var id yaml.ResourceMeta
		err := yaml.Unmarshal(b.Bytes(), &id)
		if err != nil {
			return nil, errors.WrapPrefixf(err, "failed to unmarshal resource identifier from %v", template.DefinedTemplates())
		}
		selector := framework.MatchAll(
			framework.GVKMatcher(strings.Join([]string{id.APIVersion, id.Kind}, "/")), framework.NameMatcher(id.Name),
			framework.NamespaceMatcher(id.Namespace))
		selector.FailOnEmptyMatch = true
		patches = append(patches, framework.PatchTemplate(&framework.ResourcePatchTemplate{
			Templates: parser.TemplateFiles(a.Overrides.ResourcePatches...).WithExtensions(".yaml", ".template.yaml"),
			Selector:  selector,
		}))
	}
	return patches, nil
}

func (a ExampleApp) appTemplateData(w App) map[string]interface{} {
	return map[string]interface{}{
		"AppType":     w.AppType,
		"Environment": a.Env,
		"Image":       w.Image,
		"Name":        a.ObjectMeta.Name,
		"Replicas":    w.Replicas,
	}
}
