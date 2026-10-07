package main

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

const hostPlatform = "linux"

func loadAverage() string {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(b))
}
func cpuQuota() string {
	b, err := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if err != nil {
		return "unavailable"
	}
	return strings.TrimSpace(string(b))
}
func processorCount() string {
	s, err := output("nproc")
	if err != nil {
		return strconv.Itoa(runtime.NumCPU())
	}
	return s
}
