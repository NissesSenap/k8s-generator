package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Metadata struct {
	Name string `yaml:"name"`
}

type AppSpec struct {
	Image string `yaml:"image"`
	Port  int32  `yaml:"port"`
	Size  string `yaml:"size,omitempty"`
}

type App struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     AppSpec  `yaml:"spec"`
}

func main() {
	config := &App{}
	fn := framework.TemplateProcessor{
		TemplateData:       config,
		PostProcessFilters: []kio.Filter{kio.FilterFunc(filterAppFromResources)},
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: parser.TemplateStrings(DEPLOYMENT_TEMPLATE, SERVICE_TEMPLATE),
		}},
	}
	cmd := command.Build(fn, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func filterAppFromResources(items []*yaml.RNode) ([]*yaml.RNode, error) {
	var newNodes []*yaml.RNode
	for i := range items {
		meta, err := items[i].GetMeta()
		if err != nil {
			return nil, err
		}
		// remove resources with the kind App from the resource list
		if meta.Kind == "App" && meta.APIVersion == "app.innoq.com/v1" {
			continue
		}
		newNodes = append(newNodes, items[i])
	}
	items = newNodes
	return items, nil
}
