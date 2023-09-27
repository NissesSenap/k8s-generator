# k8s-generator

Is a tool to build k8s manifests using Kustomize functions.

I will use kustomize example as a [base](https://github.com/kubernetes-sigs/kustomize/blob/master/functions/examples/fn-framework-application/README.md)

## Generate dockerfile

```shell
go run *.go gen .
```

## Run KRM

```shell
kustomize build --enable-alpha-plugins

```
