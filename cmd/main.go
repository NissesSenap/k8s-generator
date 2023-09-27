// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/NissesSenap/k8s-generator/pkg/dispatcher"

func main() {
	_ = dispatcher.NewCommand().Execute()
}
