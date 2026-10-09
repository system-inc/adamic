package estree

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// miscPlan validates the complete union even when only some shards are selected.
func miscPlan(ids []string, count int) ([][]int, error) {
	if count < 1 || len(ids) < count {
		return nil, fmt.Errorf("%d cases cannot enumerate %d shards", len(ids), count)
	}
	shards := make([][]int, count)
	expected := make(map[string]bool, len(ids))
	for i, id := range ids {
		if id == "" || expected[id] {
			return nil, fmt.Errorf("empty or repeated case id %q", id)
		}
		expected[id] = true
		shards[i%count] = append(shards[i%count], i)
	}
	seen := make(map[string]bool, len(ids))
	total := 0
	for _, shard := range shards {
		for _, i := range shard {
			id := ids[i]
			if !expected[id] || seen[id] {
				return nil, fmt.Errorf("unexpected or repeated case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if len(shards) != count || total != len(ids) || len(seen) != len(expected) {
		return nil, fmt.Errorf("shard union differs: shards=%d cases=%d unique=%d", len(shards), total, len(seen))
	}
	for id := range expected {
		if !seen[id] {
			return nil, fmt.Errorf("missing case %q", id)
		}
	}
	return shards, nil
}

func miscIDs(prefix string, count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = fmt.Sprintf("%s-%03d", prefix, i)
	}
	return ids
}

// ADAMIC_TEST_SHARD=i/n selects shard indexes congruent to i modulo n.
// Unset runs every shard. Each leaf retains every implementation for its cases.
func miscRunShards(t *testing.T, count int, ids []string, check func(*testing.T, int)) {
	t.Helper()
	shards, err := miscPlan(ids, count)
	if err != nil {
		t.Fatal(err)
	}
	selected, boxes := 0, 1
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		selected, err = strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		boxes, err = strconv.Atoi(parts[1])
		if err != nil || boxes < 1 || boxes > count || selected < 0 || selected >= boxes {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
	}
	t.Logf("union: %d cases, %d unique ids, %d shards", len(ids), len(ids), len(shards))
	for shard, indexes := range shards {
		if shard%boxes != selected {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", shard), func(t *testing.T) {
			t.Parallel()
			for _, i := range indexes {
				check(t, i)
			}
		})
	}
}

func miscCompare(want, got []byte, mutant bool) error {
	diff := firstDifference(want, got)
	if mutant {
		if diff == "" {
			return fmt.Errorf("mutant survived")
		}
		return nil
	}
	if diff != "" {
		return fmt.Errorf("%s", diff)
	}
	return nil
}

// Re-execute this tiny proof test, driving the same planner, parallel runner and
// checker as the real test. Exactly one planted case must fail exactly its leaf.
func miscPlantedProof(t *testing.T, count int, ids []string, check func(bool) error) {
	t.Helper()
	planted := len(ids) - 1
	marker := "ADAMIC_ESTREE_MISC_PROOF"
	if os.Getenv(marker) == t.Name() {
		miscRunShards(t, count, ids, func(t *testing.T, i int) {
			if err := check(i == planted); err != nil {
				t.Fatal(err)
			}
		})
		return
	}
	command := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v", "-test.timeout=20s")
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(value, marker+"=") {
			command.Env = append(command.Env, value)
		}
	}
	command.Env = append(command.Env, marker+"="+t.Name())
	output, err := command.CombinedOutput()
	leaf := fmt.Sprintf("%s/shard-%03d", t.Name(), planted%count)
	if err == nil || bytes.Count(output, []byte("--- FAIL: "+t.Name()+"/shard-")) != 1 || !bytes.Contains(output, []byte("--- FAIL: "+leaf+" ")) {
		t.Fatalf("planted failure must be caught only by %s: exit=%v\n%s", leaf, err, output)
	}
	t.Logf("planted case %s caught only by %s", ids[planted], leaf)
}

func miscOracle(t *testing.T) string {
	t.Helper()
	start := time.Now()
	path := goOracle(t)
	t.Logf("build Go oracle: %.6fs", time.Since(start).Seconds())
	return path
}

// Products are built once before the parallel leaves. This callback writes only
// into directory and can be passed to internal/buildcache when it is available.
func miscBuild(t *testing.T, path string) (string, string) {
	t.Helper()
	directory := t.TempDir()
	product := func(directory string) error {
		program, err := load.Load([]string{path})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := native.Build(native.C(lowered), filepath.Join(directory, "port"), native.Options{Sanitize: true}); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "port.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	}
	start := time.Now()
	if err := product(directory); err != nil {
		t.Fatal(err)
	}
	t.Logf("build lowered/sanitized/emitted products: %.6fs", time.Since(start).Seconds())
	return filepath.Join(directory, "port"), filepath.Join(directory, "port.mjs")
}

// CPU includes this test process and all completed child processes.
func miscCPU() float64 {
	var self, child syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &self); err != nil {
		panic(err)
	}
	if err := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &child); err != nil {
		panic(err)
	}
	return float64(self.Utime.Sec+self.Stime.Sec+child.Utime.Sec+child.Stime.Sec) + float64(self.Utime.Usec+self.Stime.Usec+child.Utime.Usec+child.Stime.Usec)/1e6
}

func miscStart(t *testing.T) func() {
	start, cpu := time.Now(), miscCPU()
	t.Cleanup(func() { t.Logf("total CPU: %.6fs", miscCPU()-cpu) })
	return func() { t.Logf("setup wall: %.6fs", time.Since(start).Seconds()) }
}

func TestMiscShardUnionRejectsInvalidEnumeration(t *testing.T) {
	for _, item := range []struct {
		ids   []string
		count int
	}{
		{[]string{"a", "a"}, 2}, {[]string{"a", ""}, 2}, {[]string{"a"}, 2}, {[]string{"a"}, 0},
	} {
		if _, err := miscPlan(item.ids, item.count); err == nil {
			t.Fatalf("invalid enumeration accepted: %+v", item)
		}
	}
}
