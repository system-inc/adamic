package typeaware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const volumeAgreementCompilerGroups = 16
const volumeAgreementRepositoryGroups = 16

type volumeAgreementCorpusFile struct{ key, path string }

func volumeAgreementCorpusOwner(key string, groups int) int {
	sum := sha256.Sum256([]byte(key))
	return int(sum[0]) % groups
}
func volumeAgreementCorpusFiles(t *testing.T, kind string) (string, []volumeAgreementCorpusFile) {
	t.Helper()
	var paths []string
	var root, config string
	switch kind {
	case "compiler":
		root = os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if root == "" {
			t.Skip("ADAMIC_TYPESCRIPT_SOURCE is not configured")
		}
		var err error
		root, err = filepath.Abs(root)
		if err != nil {
			t.Fatal(err)
		}
		paths = pinnedCompilerFiles(t, root)
		config = filepath.Join(root, "src/compiler/tsconfig.json")
	case "repository":
		manifest := os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST")
		if manifest == "" {
			t.Skip("ADAMIC_VOLUME_REPOSITORY_MANIFEST is not configured")
		}
		root = volumeGuardPreparationHarness(t).repository
		data, err := os.ReadFile(manifest)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if path != "" {
				if !filepath.IsAbs(path) {
					path = filepath.Join(root, path)
				}
				paths = append(paths, filepath.Clean(path))
			}
		}
		config = filepath.Join(root, "tsconfig.json")
	default:
		t.Fatalf("unknown corpus %s", kind)
	}
	seen := map[string]bool{}
	var files []volumeAgreementCorpusFile
	for _, path := range paths {
		key, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		key = filepath.ToSlash(key)
		if seen[key] {
			t.Fatalf("duplicate %s corpus root %s", kind, key)
		}
		seen[key] = true
		files = append(files, volumeAgreementCorpusFile{key, path})
	}
	if len(files) == 0 {
		t.Fatalf("empty %s corpus", kind)
	}
	return config, files
}
func volumeAgreementCorpusUnion(t *testing.T, kind string, groups int) {
	t.Helper()
	_, files := volumeAgreementCorpusFiles(t, kind)
	for _, mode := range []string{"plain", "asan"} {
		seen := map[string]int{}
		for shard := 0; shard < groups; shard++ {
			for _, file := range files {
				if volumeAgreementCorpusOwner(file.key, groups) == shard {
					seen[file.key]++
				}
			}
		}
		if len(seen) != len(files) {
			t.Fatalf("%s %s union=%d want %d", kind, mode, len(seen), len(files))
		}
		for _, file := range files {
			if seen[file.key] != 1 {
				t.Fatalf("%s %s file %s occurs %d times", kind, mode, file.key, seen[file.key])
			}
		}
		t.Logf("%s %s counted union: %d files exactly once in %d groups", kind, mode, len(seen), groups)
	}
}
func runVolumeAgreementCorpusShard(t *testing.T, kind string, groups, unit int) {
	t.Helper()
	// Live enumeration and required product fetches precede this unit's work clock.
	config, files := volumeAgreementCorpusFiles(t, kind)
	group := unit % groups
	name := "volume"
	mode := "plain"
	if unit >= groups {
		name = "volume-asan"
		mode = "asan"
	}
	var paths []string
	for _, file := range files {
		if volumeAgreementCorpusOwner(file.key, groups) == group {
			paths = append(paths, file.path)
		}
	}
	t.Logf("%s %s group %d: %d/%d file roots", kind, mode, group, len(paths), len(files))
	if len(paths) == 0 {
		t.Log("own work=0s")
		return
	}
	oracle := volumeAgreementFetch(t, "oracle")
	binary := volumeAgreementFetch(t, name)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() {
		t.Logf("own work=%.6fs", time.Since(started).Seconds())
		if ctx.Err() != nil {
			t.Error("corpus shard own work exceeded 60s")
		}
	}()
	h := volumeGuardPreparationHarness(t)
	h.ctx = ctx
	manifest := h.write(fmt.Sprintf("%s-%s-%03d.manifest", kind, mode, group), strings.Join(paths, "\n")+"\n")
	volumeGuardCompare(h, kind+"-"+mode, oracle, binary, config, manifest)
}
func TestVolumeAgreementCompilerUnion(t *testing.T) {
	t.Parallel()
	volumeAgreementCorpusUnion(t, "compiler", volumeAgreementCompilerGroups)
}
func TestVolumeAgreementRepositoryUnion(t *testing.T) {
	t.Parallel()
	volumeAgreementCorpusUnion(t, "repository", volumeAgreementRepositoryGroups)
}
func TestVolumeAgreementCompiler_000(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 0)
}
func TestVolumeAgreementCompiler_001(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 1)
}
func TestVolumeAgreementCompiler_002(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 2)
}
func TestVolumeAgreementCompiler_003(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 3)
}
func TestVolumeAgreementCompiler_004(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 4)
}
func TestVolumeAgreementCompiler_005(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 5)
}
func TestVolumeAgreementCompiler_006(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 6)
}
func TestVolumeAgreementCompiler_007(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 7)
}
func TestVolumeAgreementCompiler_008(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 8)
}
func TestVolumeAgreementCompiler_009(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 9)
}
func TestVolumeAgreementCompiler_010(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 10)
}
func TestVolumeAgreementCompiler_011(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 11)
}
func TestVolumeAgreementCompiler_012(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 12)
}
func TestVolumeAgreementCompiler_013(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 13)
}
func TestVolumeAgreementCompiler_014(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 14)
}
func TestVolumeAgreementCompiler_015(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 15)
}
func TestVolumeAgreementCompiler_016(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 16)
}
func TestVolumeAgreementCompiler_017(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 17)
}
func TestVolumeAgreementCompiler_018(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 18)
}
func TestVolumeAgreementCompiler_019(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 19)
}
func TestVolumeAgreementCompiler_020(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 20)
}
func TestVolumeAgreementCompiler_021(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 21)
}
func TestVolumeAgreementCompiler_022(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 22)
}
func TestVolumeAgreementCompiler_023(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 23)
}
func TestVolumeAgreementCompiler_024(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 24)
}
func TestVolumeAgreementCompiler_025(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 25)
}
func TestVolumeAgreementCompiler_026(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 26)
}
func TestVolumeAgreementCompiler_027(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 27)
}
func TestVolumeAgreementCompiler_028(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 28)
}
func TestVolumeAgreementCompiler_029(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 29)
}
func TestVolumeAgreementCompiler_030(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 30)
}
func TestVolumeAgreementCompiler_031(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "compiler", volumeAgreementCompilerGroups, 31)
}
func TestVolumeAgreementRepository_000(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 0)
}
func TestVolumeAgreementRepository_001(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 1)
}
func TestVolumeAgreementRepository_002(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 2)
}
func TestVolumeAgreementRepository_003(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 3)
}
func TestVolumeAgreementRepository_004(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 4)
}
func TestVolumeAgreementRepository_005(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 5)
}
func TestVolumeAgreementRepository_006(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 6)
}
func TestVolumeAgreementRepository_007(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 7)
}
func TestVolumeAgreementRepository_008(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 8)
}
func TestVolumeAgreementRepository_009(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 9)
}
func TestVolumeAgreementRepository_010(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 10)
}
func TestVolumeAgreementRepository_011(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 11)
}
func TestVolumeAgreementRepository_012(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 12)
}
func TestVolumeAgreementRepository_013(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 13)
}
func TestVolumeAgreementRepository_014(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 14)
}
func TestVolumeAgreementRepository_015(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 15)
}
func TestVolumeAgreementRepository_016(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 16)
}
func TestVolumeAgreementRepository_017(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 17)
}
func TestVolumeAgreementRepository_018(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 18)
}
func TestVolumeAgreementRepository_019(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 19)
}
func TestVolumeAgreementRepository_020(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 20)
}
func TestVolumeAgreementRepository_021(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 21)
}
func TestVolumeAgreementRepository_022(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 22)
}
func TestVolumeAgreementRepository_023(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 23)
}
func TestVolumeAgreementRepository_024(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 24)
}
func TestVolumeAgreementRepository_025(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 25)
}
func TestVolumeAgreementRepository_026(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 26)
}
func TestVolumeAgreementRepository_027(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 27)
}
func TestVolumeAgreementRepository_028(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 28)
}
func TestVolumeAgreementRepository_029(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 29)
}
func TestVolumeAgreementRepository_030(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 30)
}
func TestVolumeAgreementRepository_031(t *testing.T) {
	t.Parallel()
	runVolumeAgreementCorpusShard(t, "repository", volumeAgreementRepositoryGroups, 31)
}
