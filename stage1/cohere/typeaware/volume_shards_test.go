package typeaware

import (
	"bytes"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type volumeShard struct {
	ids []string
	run func(*harness)
}

func volumeUnion(expected []string, shards []volumeShard) error {
	want := make(map[string]bool, len(expected))
	for _, id := range expected {
		if want[id] {
			return fmt.Errorf("repeated unsplit case %q", id)
		}
		want[id] = true
	}
	seen := make(map[string]bool, len(expected))
	for _, shard := range shards {
		for _, id := range shard.ids {
			if !want[id] {
				return fmt.Errorf("unexpected sharded case %q", id)
			}
			if seen[id] {
				return fmt.Errorf("repeated sharded case %q", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(want) {
		return fmt.Errorf("union has %d cases, unsplit has %d", len(seen), len(want))
	}
	return nil
}

func volumeSelected(value string, index int) (bool, error) {
	if value == "" {
		return true, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return false, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n (zero based), got %q", value)
	}
	i, e1 := strconv.Atoi(parts[0])
	n, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || n < 1 || i < 0 || i >= n {
		return false, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return index%n == i, nil
}

func volumeRunShards(t *testing.T, repository string, expected []string, shards []volumeShard) {
	t.Helper()
	if err := volumeUnion(expected, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("volume-union cases=%d shards=%d: exact IDs, no missing or repeated case", len(expected), len(shards))
	for i, shard := range shards {
		selected, err := volumeSelected(os.Getenv("ADAMIC_TEST_SHARD"), i)
		if err != nil {
			t.Fatal(err)
		}
		if !selected {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			h := &harness{t: t, repository: repository, directory: t.TempDir()}
			started := time.Now()
			shard.run(h)
			t.Logf("volume-shard cases=%d logic=%.6fs ids=%s", len(shard.ids), time.Since(started).Seconds(), strings.Join(shard.ids, ","))
		})
	}
}

func volumeIDs(prefix string, paths []string) []string {
	ids := make([]string, len(paths))
	for i, path := range paths {
		ids[i] = prefix + ":" + filepath.Clean(path)
	}
	return ids
}

func volumeManifest(t *testing.T, config, manifest string) []string {
	t.Helper()
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, path := range strings.Split(string(data), "\n") {
		if path == "" {
			continue
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(config), path)
		}
		paths = append(paths, filepath.Clean(path))
	}
	sort.Strings(paths)
	return paths
}

// Files above 128 KiB get a unit of their own; other files use four-file units.
// Each unit compares the independent Go oracle to BOTH native products, including
// sanitizer/leak stderr checks. Config declarations and imported dependencies
// remain available to each program through the unchanged compiler loader.
func volumeCorpusShards(t *testing.T, prefix string, paths []string, config, oracle, binary, asan string) []volumeShard {
	t.Helper()
	dataInputs := volumeOracleData(t, prefix, config)
	var groups [][]string
	var small []string
	flush := func() {
		if len(small) > 0 {
			groups = append(groups, small)
			small = nil
		}
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() > 128*1024 {
			flush()
			groups = append(groups, []string{path})
		} else {
			small = append(small, path)
			if len(small) == 4 {
				flush()
			}
		}
	}
	flush()
	shards := make([]volumeShard, 0, len(groups))
	for _, group := range groups {
		shards = append(shards, volumeShard{ids: volumeIDs(prefix, group), run: func(h *harness) {
			manifest := h.write("cases.manifest", strings.Join(group, "\n")+"\n")
			volumeCompareProducts(h, dataInputs, prefix, oracle, binary, asan, config, manifest)
		}})
	}
	return shards
}

// Side executions are independent parallel leaves; compare their complete outputs
// after the group joins. One independent Go result is compared to EACH native
// side, and every native side must have empty sanitizer/leak stderr.
func volumeCompareProducts(h *harness, dataInputs buildcache.Inputs, name, oracle, binary, asan, config, manifest string) {
	var results [3]result
	ok := h.t.Run("sides", func(t *testing.T) {
		for i, side := range []struct{ name, binary string }{{"oracle", oracle}, {"native", binary}, {"asan", asan}} {
			t.Run(side.name, func(t *testing.T) {
				t.Parallel()
				child := &harness{t: t, repository: h.repository, directory: t.TempDir()}
				if side.name == "oracle" {
					results[i] = volumeOracleOutput(child, dataInputs, side.binary, config, manifest)
				} else {
					results[i] = child.must(name+"-"+side.name, exec.Command(side.binary, config, manifest))
				}
			})
		}
	})
	if !ok {
		return
	}
	for i, side := range []string{"native", "asan"} {
		got, want := results[i+1], results[0]
		if len(got.stderr) != 0 {
			h.t.Fatalf("%s sanitizer stderr: %s", side, got.stderr)
		}
		if !bytes.Equal(got.stdout, want.stdout) {
			at := firstDifference(got.stdout, want.stdout)
			h.t.Fatalf("%s %s mismatch byte %d: native %q Go %q", name, side, at, got.stdout[max(0, at-50):min(len(got.stdout), at+250)], want.stdout[max(0, at-50):min(len(want.stdout), at+250)])
		}
	}
	h.t.Logf("%s: %d identical finding bytes on native and asan; %s", name, len(results[0].stdout), summary(results[0].stdout))
}

func TestVolumeShardUnionAndSelection(t *testing.T) {
	for _, invalid := range [][]volumeShard{{{ids: []string{"a"}}}, {{ids: []string{"a", "a", "b"}}}, {{ids: []string{"a", "b", "c"}}}} {
		if volumeUnion([]string{"a", "b"}, invalid) == nil {
			t.Fatal("invalid union accepted")
		}
	}
	if err := volumeUnion([]string{"a", "b"}, []volumeShard{{ids: []string{"b"}}, {ids: []string{"a"}}}); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"0/0", "-1/2", "2/2", "1", "x/2"} {
		if _, err := volumeSelected(value, 0); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	for index := 0; index < 19; index++ {
		count := 0
		for i := 0; i < 5; i++ {
			selected, err := volumeSelected(fmt.Sprintf("%d/5", i), index)
			if err != nil {
				t.Fatal(err)
			}
			if selected {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("case %d selected %d times", index, count)
		}
	}
}

// The subprocess runs the same dispatcher and byte comparison as the real test.
// Only case-004 disagrees, so only shard-001 may fail.
func TestVolumeShardPlantedDisagreement(t *testing.T) {
	if os.Getenv("ADAMIC_VOLUME_PLANT_PROBE") == "1" {
		directory := t.TempDir()
		oracle := filepath.Join(directory, "oracle")
		native := filepath.Join(directory, "native")
		for path, script := range map[string]string{oracle: "#!/bin/sh\ncat \"$2\"\n", native: "#!/bin/sh\nsed 's/case-004/disagreement/' \"$2\"\n"} {
			if err := os.WriteFile(path, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
		}
		var paths []string
		for i := 0; i < 9; i++ {
			path := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
			if err := os.WriteFile(path, nil, 0600); err != nil {
				t.Fatal(err)
			}
			paths = append(paths, path)
		}
		shards := volumeCorpusShards(t, "plant", paths, "", oracle, native, native)
		volumeRunShards(t, directory, volumeIDs("plant", paths), shards)
		return
	}
	command := exec.Command(os.Args[0], "-test.run=^TestVolumeShardPlantedDisagreement$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_VOLUME_PLANT_PROBE=1", "ADAMIC_TEST_SHARD=")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("planted disagreement survived")
	}
	var failed []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--- FAIL: TestVolumeShardPlantedDisagreement/shard-") {
			failed = append(failed, strings.Fields(line)[2])
		}
	}
	if len(failed) != 1 || failed[0] != "TestVolumeShardPlantedDisagreement/shard-001" {
		t.Fatalf("wrong failing shards %v:\n%s", failed, output)
	}
	t.Log("planted case-004 caught exactly by shard-001")
}
