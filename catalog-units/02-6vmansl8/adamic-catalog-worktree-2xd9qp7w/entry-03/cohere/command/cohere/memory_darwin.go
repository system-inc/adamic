package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
)

// availableMemory reads the kernel's page counts. See darwinAvailableMemory for which pages count.
func availableMemory() (uint64, error) {
	var counts [3]uint64
	for index, name := range []string{"vm.page_free_count", "vm.page_pageable_external_count", "vm.page_purgeable_count"} {
		count, err := sysctlCount(name)
		if err != nil {
			return 0, fmt.Errorf("reading %s: %w", name, err)
		}
		counts[index] = count
	}
	return darwinAvailableMemory(uint64(os.Getpagesize()), counts[0], counts[1], counts[2]), nil
}

// sysctlCount reads an unsigned counter of either width. The counts are not all one size
// (vm.page_purgeable_count is wider than syscall.SysctlUint32 accepts), and syscall.Sysctl drops a
// trailing zero byte, so the value is zero-extended back to eight bytes, which is exact for a
// little-endian integer of four or eight.
func sysctlCount(name string) (uint64, error) {
	value, err := syscall.Sysctl(name)
	if err != nil {
		return 0, err
	}
	if len(value) > 8 {
		return 0, fmt.Errorf("%d bytes, wider than a counter", len(value))
	}
	var padded [8]byte
	copy(padded[:], value)
	return binary.LittleEndian.Uint64(padded[:]), nil
}
