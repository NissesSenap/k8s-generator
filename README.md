# k8s-generator

Is a tool to build k8s manifests using Kustomize functions.

I will use kustomize example as a [base](https://github.com/kubernetes-sigs/kustomize/blob/master/functions/examples/fn-framework-application/README.md)

## Generate dockerfile

This is only needed to be run once.

```shell
go run cmd/main.go gen .
```

## Run tests

A. Run `make example` in the root of the example to run the function with the test data in `pkg/exampleapp/v1alpha1/testdata/success/basic`.

B. Run `go run cmd/main.go [FILE]` in the root of the example. Try it with the test input from one of the cases in `pkg/exampleapp/v1alpha1/testdata/success`. For example: `go run cmd/main.go pkg/exampleapp/v1alpha1/testdata/success/basic/config.yaml`.

C. Build the binary with `make build`, then run it with `app-fn [FILE]`. Try it with the test input from one of the cases in `pkg/exampleapp/v1alpha1/testdata/success`. For example: `app-fn pkg/exampleapp/v1alpha1/testdata/success/basic/config.yaml`.

## Run KRM with Kustomize

This assumes that you have built the container image first.

```shell
kustomize build --enable-alpha-plugins
```
