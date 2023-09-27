package v1alpha1

import (
	"embed"
	"encoding/json"
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
	if a.AppImage == "" {
		a.AppImage = fmt.Sprintf("registry.example.com/path/to/%s", a.ObjectMeta.Name)
	}
	for i := range a.Workloads.WebWorkers {
		if a.Workloads.WebWorkers[i].Replicas == nil {
			a.Workloads.WebWorkers[i].Replicas = func() *int { three := 3; return &three }()
		}
		if a.Workloads.WebWorkers[i].Resources == "" {
			a.Workloads.WebWorkers[i].Resources = ResourceBinSizeSmall
		}
	}

	for i := range a.Workloads.JobWorkers {
		if a.Workloads.JobWorkers[i].Replicas == nil {
			a.Workloads.JobWorkers[i].Replicas = func() *int { one := 1; return &one }()
		}
		if a.Workloads.JobWorkers[i].Resources == "" {
			a.Workloads.JobWorkers[i].Resources = ResourceBinSizeSmall
		}
	}
	return nil
}

func (a *ExampleApp) Validate() error {
	seenDomains := make(map[string]bool)
	for i, worker := range a.Workloads.WebWorkers {
		for _, domain := range worker.Domains {
			if seenDomains[domain] {
				return errors.Errorf("duplicate domain %q in worker %d", domain, i)
			}
			seenDomains[domain] = true
		}
	}
	return nil
}

func (a ExampleApp) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	templates := make([]framework.ResourceTemplate, 0)
	for _, worker := range a.Workloads.JobWorkers {
		templates = append(templates, framework.ResourceTemplate{
			Templates:    parser.TemplateFiles("templates/job_worker.template.yaml").FromFS(templateFS),
			TemplateData: a.jobWorkerTemplateData(worker),
		})
	}
	for _, worker := range a.Workloads.WebWorkers {
		templates = append(templates, framework.ResourceTemplate{
			Templates:    parser.TemplateFiles("templates/web_worker.template.yaml").FromFS(templateFS),
			TemplateData: a.webWorkerTemplateData(worker),
		})
	}

	var patches []framework.PatchTemplate
	if a.Datastores.PostgresInstance != "" {
		templates = append(templates, framework.ResourceTemplate{
			TemplateData: map[string]interface{}{"Name": a.Datastores.PostgresInstance},
			Templates: parser.TemplateStrings(`apiVersion: apps.example.com/v1
kind: PostgresSecretRequest
metadata:
  name: {{ .Name }}
`)})
		patches = append(patches, framework.PatchTemplate(&framework.ContainerPatchTemplate{
			Templates:        parser.TemplateFiles("templates/postgres_secret_env_patch.template.yaml").FromFS(templateFS),
			ContainerMatcher: framework.ContainerNameMatcher("app"),
		}))
	}

	return items, nil
}

type resourceBucket struct {
	Requests resourceAllocation `yaml:"requests" json:"requests"`
	Limits   resourceAllocation `yaml:"limits" json:"limits"`
}

type resourceAllocation struct {
	CPU    string `yaml:"cpu" json:"cpu"`
	Memory string `yaml:"memory" json:"memory"`
}

//nolint:gochecknoglobals
var resourceBucketConversion = map[ResourceBinSize]resourceBucket{
	"small": {
		Requests: resourceAllocation{CPU: "100m", Memory: "128Mi"},
		Limits:   resourceAllocation{CPU: "500m", Memory: "512Mi"},
	},
	"medium": {
		Requests: resourceAllocation{CPU: "1", Memory: "1Gi"},
		Limits:   resourceAllocation{CPU: "2", Memory: "2Gi"},
	},
	"large": {
		Requests: resourceAllocation{CPU: "8", Memory: "8Gi"},
		Limits:   resourceAllocation{CPU: "8", Memory: "8Gi"},
	},
}

const anArbitraryMultiplier = 2

func (a ExampleApp) jobWorkerTemplateData(w JobWorker) map[string]interface{} {
	resourcesJson, err := json.Marshal(resourceBucketConversion[w.Resources])
	if err != nil {
		panic("failed to marshal resources for job worker" + err.Error())
	}

	return map[string]interface{}{
		"Name":            w.Name,
		"AppImage":        a.AppImage,
		"QueueList":       strings.Join(w.Queues, ","),
		"ProcessPoolSize": len(w.Queues) * anArbitraryMultiplier,
		"Resources":       string(resourcesJson),
		"Replicas":        w.Replicas,
		"Environment":     a.Env,
	}
}

func (a ExampleApp) webWorkerTemplateData(w WebWorker) map[string]interface{} {
	resourcesJson, err := json.Marshal(resourceBucketConversion[w.Resources])
	if err != nil {
		panic("failed to marshal resources for web worker" + err.Error())
	}

	return map[string]interface{}{
		"Name":        w.Name,
		"AppImage":    a.AppImage,
		"Resources":   string(resourcesJson),
		"Replicas":    w.Replicas,
		"Port":        containerPort,
		"Environment": a.Env,
	}
}
