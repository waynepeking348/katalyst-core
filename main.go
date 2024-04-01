package main

import (
	"fmt"
	"runtime"
	"time"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
)

func runtimeService() {
	fmt.Println("debug start")

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

	fmt.Println("debug end")
}

func sliceLength() []func() {
	return []func(){
		func() {
			s := make([]int, 5*1024*1024, 5*1024*1024)
			runtime.GC()
			printMemUsage()
			fmt.Printf("1 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append([]int{}, s[4*1024*1024:]...)
			runtime.GC()
			printMemUsage()
			fmt.Printf("1 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
		func() {
			s := make([]int, 5*1024*1024, 5*1024*1024)
			runtime.GC()
			printMemUsage()
			fmt.Printf("2 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append(s[:1])
			runtime.GC()
			printMemUsage()
			fmt.Printf("2 cap %v, len %v, %p\n", cap(a), len(a), a)

			for i := 0; i < 1024*1024; i++ {
				a = append(a, 1)
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("2 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
		func() {
			s := make([]int, 5*1024*1024, 5*1024*1024)
			runtime.GC()
			printMemUsage()
			fmt.Printf("3 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append(s[4*1024*1024:])
			runtime.GC()
			printMemUsage()
			fmt.Printf("3 cap %v, len %v, %p\n", cap(a), len(a), a)

			for i := 0; i < 1024*1024; i++ {
				a = append(a, 1)
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("3 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
		func() {
			s := make([]int, 5*1024*1024, 5*1024*1024)
			runtime.GC()
			printMemUsage()
			fmt.Printf("4 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append(s[:1], s[4*1024*1024:]...)
			runtime.GC()
			printMemUsage()
			fmt.Printf("4 cap %v, len %v, %p\n", cap(a), len(a), a)

			for i := 0; i < 1024*1024; i++ {
				a = append(a, 1)
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("4 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
		func() {
			s := make([]*int, 5*1024*1024, 5*1024*1024)
			for i := 0; i < 5*1024*1024; i++ {
				a := 0
				s[i] = &a
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("5 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append([]*int{}, s[4*1024*1024:]...)
			runtime.GC()
			printMemUsage()
			fmt.Printf("5 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
		func() {
			s := make([]*int, 5*1024*1024, 5*1024*1024)
			for i := 0; i < 5*1024*1024; i++ {
				a := 0
				s[i] = &a
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("6 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append(s[:1])
			runtime.GC()
			printMemUsage()
			fmt.Printf("6 cap %v, len %v, %p\n", cap(a), len(a), a)

			for i := 0; i < 1024*1024; i++ {
				x := 0
				a = append(a, &x)
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("6 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
		func() {
			s := make([]*int, 5*1024*1024, 5*1024*1024)
			for i := 0; i < 5*1024*1024; i++ {
				a := 0
				s[i] = &a
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("7 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append(s[4*1024*1024:])
			runtime.GC()
			printMemUsage()
			fmt.Printf("7 cap %v, len %v, %p\n", cap(a), len(a), a)

			for i := 0; i < 1024*1024; i++ {
				x := 0
				a = append(a, &x)
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("7 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
		func() {
			s := make([]*int, 5*1024*1024, 5*1024*1024)
			for i := 0; i < 5*1024*1024; i++ {
				a := 0
				s[i] = &a
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("8 cap %v, len %v, %p\n", cap(s), len(s), s)

			a := append(s[:1], s[4*1024*1024:]...)
			runtime.GC()
			printMemUsage()
			fmt.Printf("8 cap %v, len %v, %p\n", cap(a), len(a), a)

			for i := 0; i < 1024*1024; i++ {
				x := 0
				a = append(a, &x)
			}
			runtime.GC()
			printMemUsage()
			fmt.Printf("8 cap %v, len %v, %p\n", cap(a), len(a), a)
		},
	}
}

func printMemUsage() {
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func main() {
	for _, f := range sliceLength() {
		f()
		fmt.Println()
		fmt.Println()
	}
}
