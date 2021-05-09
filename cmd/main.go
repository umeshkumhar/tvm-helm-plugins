package main

import (
	"github.com/trilioData/tvm-helm-plugins/cmd/root"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	if err := root.Execute(); err != nil {
		panic(err)
	}
}
