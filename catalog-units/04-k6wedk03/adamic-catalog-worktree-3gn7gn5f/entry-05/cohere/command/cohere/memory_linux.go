package main

import "os"

// availableMemory reads /proc/meminfo and, inside a container, the cgroup's limit. See
// linuxAvailableMemory.
func availableMemory() (uint64, error) {
	meminfo, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	read := func(path string) []byte {
		contents, _ := os.ReadFile(path)
		return contents
	}
	return linuxAvailableMemory(meminfo, linuxCgroupMemory{
		v2Max:     read("/sys/fs/cgroup/memory.max"),
		v2Current: read("/sys/fs/cgroup/memory.current"),
		v1Limit:   read("/sys/fs/cgroup/memory/memory.limit_in_bytes"),
		v1Usage:   read("/sys/fs/cgroup/memory/memory.usage_in_bytes"),
	})
}
