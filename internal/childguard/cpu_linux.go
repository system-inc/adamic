//go:build linux

package childguard

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Linux exposes these four fields in USER_HZ (100 ticks/second), regardless
// of the kernel scheduler's CONFIG_HZ. cutime/cstime retain reaped descendants'
// CPU; summing them with each group member's own CPU includes delegated work.
func statCPU(data string) (group int, cpu time.Duration, err error) {
	end := strings.LastIndexByte(data, ')')
	if end < 0 {
		return 0, 0, fmt.Errorf("invalid /proc stat command")
	}
	fields := strings.Fields(data[end+1:])
	if len(fields) < 15 {
		return 0, 0, fmt.Errorf("short /proc stat")
	}
	group, err = strconv.Atoi(fields[2])
	if err != nil {
		return 0, 0, err
	}
	for _, index := range []int{11, 12, 13, 14} {
		ticks, parseErr := strconv.ParseInt(fields[index], 10, 64)
		if parseErr != nil || ticks < 0 {
			return 0, 0, fmt.Errorf("invalid /proc CPU field %q", fields[index])
		}
		cpu += time.Duration(ticks) * 10 * time.Millisecond
	}
	return group, cpu, nil
}

func processGroupCPU(group int) (time.Duration, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}
	var total time.Duration
	found := false
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		data, err := os.ReadFile("/proc/" + entry.Name() + "/stat")
		// A process can disappear while /proc is enumerated.
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return 0, err
		}
		pgid, cpu, err := statCPU(string(data))
		if err != nil {
			return 0, err
		}
		if pgid == group {
			found = true
			total += cpu
		}
	}
	if !found {
		return 0, fmt.Errorf("process group %d absent from /proc", group)
	}
	return total, nil
}
