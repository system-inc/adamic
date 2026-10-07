package main

import (
	"runtime"
	"strconv"
	"strings"
)

const hostPlatform = "darwin"

func loadAverage() string {
	s, err := output("sysctl", "-n", "vm.loadavg")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(strings.Trim(s, "{}"))
}
func cpuQuota() string { return "not-applicable" }
func processorCount() string {
	s, err := output("sysctl", "-n", "hw.logicalcpu")
	if err != nil {
		return strconv.Itoa(runtime.NumCPU())
	}
	return s
}
