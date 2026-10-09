package main

import (
	"errors"
	"strings"
	"testing"
)

func TestMemoryPolicyTurnsCollectionOffUnderHalfOfWhatIsAvailable(t *testing.T) {
	t.Parallel()
	noEnvironment := func(string) (string, bool) { return "", false }
	policy := decideMemoryPolicy(noEnvironment, func() (uint64, error) { return 108e9, nil })
	if policy.ceiling != 54e9 {
		t.Fatalf("ceiling %d, want half of 108 GB", policy.ceiling)
	}
	if got, want := policy.line(), "memory: collector off, ceiling 54 GB (half of 108 GB available)"; got != want {
		t.Fatalf("line %q, want %q", got, want)
	}
	small := decideMemoryPolicy(noEnvironment, func() (uint64, error) { return 7e9, nil })
	if got, want := small.line(), "memory: collector off, ceiling 3.5 GB (half of 7.0 GB available)"; got != want {
		t.Fatalf("line %q, want %q", got, want)
	}
}

func TestMemoryPolicyStepsAsideForTheEnvironment(t *testing.T) {
	t.Parallel()
	for _, set := range []map[string]string{{"GOGC": "100"}, {"GOMEMLIMIT": "2GiB"}, {"GOGC": "off", "GOMEMLIMIT": "6GiB"}} {
		lookup := func(name string) (string, bool) { value, ok := set[name]; return value, ok }
		asked := false
		policy := decideMemoryPolicy(lookup, func() (uint64, error) { asked = true; return 108e9, nil })
		if policy.ceiling != 0 || asked {
			t.Fatalf("%v: the policy set a ceiling of %d (memory read: %v) where the environment decides", set, policy.ceiling, asked)
		}
		for name, value := range set {
			if !strings.Contains(policy.line(), name+"="+value) {
				t.Fatalf("%v: line %q does not name %s", set, policy.line(), name)
			}
		}
	}
}

func TestMemoryPolicyKeepsTheDefaultWhenMemoryCannotBeRead(t *testing.T) {
	t.Parallel()
	noEnvironment := func(string) (string, bool) { return "", false }
	for _, reading := range []func() (uint64, error){
		func() (uint64, error) { return 0, errors.New("no reader") },
		func() (uint64, error) { return 0, nil },
	} {
		policy := decideMemoryPolicy(noEnvironment, reading)
		if policy.ceiling != 0 || !strings.Contains(policy.line(), "Go's default") {
			t.Fatalf("an unreadable reading turned collection off: %+v, %q", policy, policy.line())
		}
	}
}

func TestDarwinAvailableMemoryCountsFreeFileBackedAndPurgeablePages(t *testing.T) {
	t.Parallel()
	if got := darwinAvailableMemory(16384, 2_898_869, 1_800_168, 57_667); got != 16384*(2_898_869+1_800_168+57_667) {
		t.Fatalf("got %d", got)
	}
}

func TestLinuxAvailableMemoryTakesTheSmallerOfMemAvailableAndTheCgroup(t *testing.T) {
	t.Parallel()
	meminfo := []byte("MemTotal:       16384000 kB\nMemFree:         1000000 kB\nMemAvailable:    8000000 kB\n")
	host := uint64(8_000_000 * 1024)
	cases := []struct {
		name   string
		cgroup linuxCgroupMemory
		want   uint64
	}{
		{"no cgroup", linuxCgroupMemory{}, host},
		{"v2 without a limit", linuxCgroupMemory{v2Max: []byte("max\n"), v2Current: []byte("1000\n")}, host},
		{"v2 limit with usage", linuxCgroupMemory{v2Max: []byte("4294967296\n"), v2Current: []byte("1073741824\n")}, 3 << 30},
		{"v2 limit without usage", linuxCgroupMemory{v2Max: []byte("4294967296\n")}, 4 << 30},
		{"v2 over its limit", linuxCgroupMemory{v2Max: []byte("1000\n"), v2Current: []byte("2000\n")}, 0},
		{"v1 unlimited", linuxCgroupMemory{v1Limit: []byte("9223372036854771712\n"), v1Usage: []byte("1000\n")}, host},
		{"v1 limit with usage", linuxCgroupMemory{v1Limit: []byte("2147483648\n"), v1Usage: []byte("1073741824\n")}, 1 << 30},
		{"a cgroup above the host", linuxCgroupMemory{v2Max: []byte("68719476736\n"), v2Current: []byte("0\n")}, host},
	}
	for _, entry := range cases {
		got, err := linuxAvailableMemory(meminfo, entry.cgroup)
		if err != nil || got != entry.want {
			t.Errorf("%s: got %d, %v; want %d", entry.name, got, err, entry.want)
		}
	}
	if _, err := linuxAvailableMemory([]byte("MemTotal: 1 kB\n"), linuxCgroupMemory{}); err == nil {
		t.Error("a meminfo without MemAvailable read as a number")
	}
}

func TestATestBinaryKeepsTheDefaultCollector(t *testing.T) {
	t.Parallel()
	if !activeMemoryPolicy.testBinary || activeMemoryPolicy.ceiling != 0 {
		t.Fatalf("the test binary's policy is %+v, which would hold every graph it builds", activeMemoryPolicy)
	}
}
