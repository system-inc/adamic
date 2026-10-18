package testguard

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MemoryLimit is the bytes the unit may hold: the smallest cgroup v2 memory.max from this process's own cgroup up to
// the root, read the way Share reads the unit's CPUs from GOMAXPROCS rather than the host's. A runner starts a unit into
// a cgroup of its own (memory.max 16 GiB beside 8 CPUs) without a cgroup namespace, so /sys/fs/cgroup/memory.max is the
// root's, which states none; the unit's limit is under the path /proc/self/cgroup names. A test that admits children by
// weight sizes its budget from this, not from a constant: json's assumed 8 GiB in a 16 GiB unit and ran one sanitized
// child at a time, so eight parallel shards queued on one slot until one passed its 90 s deadline (#q2qjes8). ok is
// false where no limit is stated (macOS, cgroup v1, every memory.max on the path "max"), and the caller keeps its
// default.
func MemoryLimit() (bytes int64, ok bool) {
	membership, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return 0, false
	}
	return memoryLimitOf("/sys/fs/cgroup", string(membership))
}

// memoryLimitOf reads the cgroup v2 hierarchy mounted at root for the cgroup a /proc/<pid>/cgroup file names.
func memoryLimitOf(root, membership string) (int64, bool) {
	var group string
	found := false
	for line := range strings.Lines(membership) {
		if path, unified := strings.CutPrefix(strings.TrimSpace(line), "0::"); unified {
			group, found = path, true
		}
	}
	if !found {
		return 0, false
	}
	var smallest int64
	for directory := filepath.Join(root, group); ; directory = filepath.Dir(directory) {
		if bytes, ok := memoryMaxOf(filepath.Join(directory, "memory.max")); ok && (smallest == 0 || bytes < smallest) {
			smallest = bytes
		}
		if directory == root || !strings.HasPrefix(directory, root) {
			break
		}
	}
	return smallest, smallest > 0
}

func memoryMaxOf(path string) (int64, bool) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	bytes, err := strconv.ParseInt(strings.TrimSpace(string(contents)), 10, 64)
	if err != nil || bytes <= 0 {
		return 0, false
	}
	return bytes, true
}
