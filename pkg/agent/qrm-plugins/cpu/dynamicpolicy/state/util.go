/*
Copyright 2022 The Katalyst Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package state

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	advisorapi "github.com/kubewharf/katalyst-core/pkg/agent/qrm-plugins/cpu/dynamicpolicy/cpuadvisor"
	"github.com/kubewharf/katalyst-core/pkg/util/machine"
)

// notice that pool-name may not have direct mapping relations with qos-level, for instance
// - both isolated_shared_cores and dedicated_cores fall into PoolNameDedicated
const (
	PoolNameShare     = "share"
	PoolNameReclaim   = "reclaim"
	PoolNameDedicated = "dedicated"
	PoolNameReserve   = "reserve"

	// PoolNameFallback is not a real pool, and is a union of
	// all none-reclaimed pools to put pod should have been isolated
	PoolNameFallback = "fallback"
)

var (
	// StaticPools are generated by cpu plugin statically,
	// and they will be ignored when reading cpu advisor list and watch response.
	StaticPools = sets.NewString(
		PoolNameReserve,
	)

	// ResidentPools are guaranteed existing in state,
	// and they are usually used to ensure stability.
	ResidentPools = sets.NewString(
		PoolNameReclaim,
	).Union(StaticPools)
)

var GetContainerRequestedCores func(allocationInfo *AllocationInfo) int

// GetIsolatedQuantityMapFromPodEntries returns a map to indicates isolation info,
// and the map is formatted as pod -> container -> isolated-quantity
func GetIsolatedQuantityMapFromPodEntries(podEntries PodEntries, ignoreAllocationInfos []*AllocationInfo) map[string]map[string]int {
	ret := make(map[string]map[string]int)
	for podUID, entries := range podEntries {
		if entries.IsPoolEntry() {
			continue
		}

	containerLoop:
		for containerName, allocationInfo := range entries {
			// only filter dedicated_cores without numa_binding
			if allocationInfo == nil || CheckNumaBinding(allocationInfo) || !CheckDedicated(allocationInfo) {
				continue
			}

			for _, ignoreAllocationInfo := range ignoreAllocationInfos {
				if allocationInfo.PodUid == ignoreAllocationInfo.PodUid && allocationInfo.ContainerName == ignoreAllocationInfo.ContainerName {
					continue containerLoop
				}
			}

			// if there is no more cores to allocate, we will put dedicated_cores without numa_binding
			// to pool rather than isolation. calling this function means we will start to adjust allocation,
			// and we will try to isolate those containers, so we will treat them as containers to be isolated.
			var quantity int
			if allocationInfo.OwnerPoolName != PoolNameDedicated {
				quantity = GetContainerRequestedCores(allocationInfo)
			} else {
				quantity = allocationInfo.AllocationResult.Size()
			}
			if quantity == 0 {
				klog.Warningf("[GetIsolatedQuantityMapFromPodEntries] isolated pod: %s/%s container: %s get zero quantity",
					allocationInfo.PodNamespace, allocationInfo.PodName, allocationInfo.ContainerName)
				continue
			}

			if ret[podUID] == nil {
				ret[podUID] = make(map[string]int)
			}
			ret[podUID][containerName] = quantity
		}
	}
	return ret
}

// GetSharedQuantityMapFromPodEntries returns a map to indicates quantity info for each shared pool,
// and the map is formatted as pool -> quantity
func GetSharedQuantityMapFromPodEntries(podEntries PodEntries, ignoreAllocationInfos []*AllocationInfo) map[string]int {
	ret := make(map[string]int)
	for _, entries := range podEntries {
		if entries.IsPoolEntry() {
			continue
		}

	containerLoop:
		for _, allocationInfo := range entries {
			// only count shared_cores not isolated.
			// if there is no more cores to allocate, we will put dedicated_cores without numa_binding to pool rather than isolation.
			// calling this function means we will start to adjust allocation, and we will try to isolate those containers,
			// so we will treat them as containers to be isolated.
			if allocationInfo == nil || !CheckShared(allocationInfo) {
				continue
			}

			for _, ignoreAllocationInfo := range ignoreAllocationInfos {
				if allocationInfo.PodUid == ignoreAllocationInfo.PodUid && allocationInfo.ContainerName == ignoreAllocationInfo.ContainerName {
					continue containerLoop
				}
			}

			if poolName := allocationInfo.GetOwnerPoolName(); poolName != advisorapi.EmptyOwnerPoolName {
				ret[poolName] += GetContainerRequestedCores(allocationInfo)
			}
		}
	}
	return ret
}

// GenerateMachineStateFromPodEntries returns NUMANodeMap for given resource based on
// machine info and reserved resources along with existed pod entries
func GenerateMachineStateFromPodEntries(topology *machine.CPUTopology, podEntries PodEntries) (NUMANodeMap, error) {
	if topology == nil {
		return nil, fmt.Errorf("GenerateMachineStateFromPodEntries got nil topology")
	}

	machineState := make(NUMANodeMap)
	for _, numaNode := range topology.CPUDetails.NUMANodes().ToSliceInt64() {
		numaNodeState := &NUMANodeState{}
		numaNodeAllCPUs := topology.CPUDetails.CPUsInNUMANodes(int(numaNode)).Clone()
		allocatedCPUsInNumaNode := machine.NewCPUSet()

		for podUID, containerEntries := range podEntries {
			for containerName, allocationInfo := range containerEntries {
				if containerName != "" && allocationInfo != nil {

					// the container hasn't cpuset assignment in the current NUMA node
					if allocationInfo.OriginalTopologyAwareAssignments[int(numaNode)].Size() == 0 &&
						allocationInfo.TopologyAwareAssignments[int(numaNode)].Size() == 0 {
						continue
					}

					// only modify allocated and default properties in NUMA node state for dedicated_cores with NUMA binding
					if CheckNumaBinding(allocationInfo) {
						allocatedCPUsInNumaNode = allocatedCPUsInNumaNode.Union(allocationInfo.OriginalTopologyAwareAssignments[int(numaNode)])
					}

					topologyAwareAssignments, _ := machine.GetNumaAwareAssignments(topology, allocationInfo.AllocationResult.Intersection(numaNodeAllCPUs))
					originalTopologyAwareAssignments, _ := machine.GetNumaAwareAssignments(topology, allocationInfo.OriginalAllocationResult.Intersection(numaNodeAllCPUs))

					numaNodeAllocationInfo := allocationInfo.Clone()
					numaNodeAllocationInfo.AllocationResult = allocationInfo.AllocationResult.Intersection(numaNodeAllCPUs)
					numaNodeAllocationInfo.OriginalAllocationResult = allocationInfo.OriginalAllocationResult.Intersection(numaNodeAllCPUs)
					numaNodeAllocationInfo.TopologyAwareAssignments = topologyAwareAssignments
					numaNodeAllocationInfo.OriginalTopologyAwareAssignments = originalTopologyAwareAssignments

					numaNodeState.SetAllocationInfo(podUID, containerName, numaNodeAllocationInfo)
				}
			}
		}

		numaNodeState.AllocatedCPUSet = allocatedCPUsInNumaNode.Clone()
		numaNodeState.DefaultCPUSet = numaNodeAllCPUs.Difference(numaNodeState.AllocatedCPUSet)
		machineState[int(numaNode)] = numaNodeState
	}
	return machineState, nil
}
