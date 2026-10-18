package testguard

import (
	"os"
	"path/filepath"
	"testing"
)

// A runner's unit: memory.max 16 GiB on the unit's own cgroup, "max" above it, none at the root. The limit is the
// smallest stated on the path from the process's cgroup to the root, wherever the cgroup sits.
func TestMemoryLimitReadsTheUnitsOwnCgroup(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(group, contents string) {
		t.Helper()
		directory := filepath.Join(root, group)
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "memory.max"), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("system.slice/loom-serve.service", "max\n")
	write("system.slice/loom-serve.service/unit-a", "17179869184\n")
	write("system.slice/loom-serve.service/unit-b", "max\n")
	write("capped", "8589934592\n")
	write("capped/inner", "17179869184\n")
	for _, check := range []struct {
		name, membership string
		want             int64
		ok               bool
	}{
		{"unit", "0::/system.slice/loom-serve.service/unit-a\n", 17179869184, true},
		{"the smaller limit above", "0::/capped/inner\n", 8589934592, true},
		{"no limit on the path", "0::/system.slice/loom-serve.service/unit-b\n", 0, false},
		{"root", "0::/\n", 0, false},
		{"cgroup v1 only", "4:memory:/unit-a\n", 0, false},
		{"absent cgroup", "0::/nowhere\n", 0, false},
	} {
		if got, ok := memoryLimitOf(root, check.membership); got != check.want || ok != check.ok {
			t.Errorf("%s: memoryLimitOf %d %t, want %d %t", check.name, got, ok, check.want, check.ok)
		}
	}
}
