package main

import (
	"syscall"
	"unsafe"
)

// memoryStatus is Windows' MEMORYSTATUSEX, laid out field for field.
type memoryStatus struct {
	length                   uint32
	memoryLoad               uint32
	totalPhysical            uint64
	availablePhysical        uint64
	totalPageFile            uint64
	availablePageFile        uint64
	totalVirtual             uint64
	availableVirtual         uint64
	availableExtendedVirtual uint64
}

var globalMemoryStatus = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")

// availableMemory is the physical memory Windows reports available, which already counts its standby
// cache as reclaimable.
func availableMemory() (uint64, error) {
	status := memoryStatus{length: uint32(unsafe.Sizeof(memoryStatus{}))}
	if succeeded, _, err := globalMemoryStatus.Call(uintptr(unsafe.Pointer(&status))); succeeded == 0 {
		return 0, err
	}
	return status.availablePhysical, nil
}
