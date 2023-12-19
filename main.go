package main

import (
	"fmt"
	"time"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
)

func main() {
	fmt.Println("debug 11")

	runtimeService, err := remote.NewRemoteRuntimeService("unix:///run/containerd/containerd.sock", 2*time.Minute)
	if err != nil {
		fmt.Printf("err-1 %s", err)
		return
	}

	stats, err := runtimeService.ListContainerStats(&runtimeapi.ContainerStatsFilter{})
	if err != nil {
		fmt.Printf("err-2 %s", err)
		return
	}

	for _, s := range stats {
		fmt.Printf("info %s", s.String())
	}

	fmt.Println("debug 22")
}
