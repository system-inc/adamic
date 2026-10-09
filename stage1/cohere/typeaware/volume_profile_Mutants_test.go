package typeaware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

const testVolumeProfileMutantsShards = 3

// Fixed generated mutation modes from profile_test.go at 7a10c877. Corpus
// growth cannot move a mode between these three slices.
type volumeProfileMutantChange struct{ name, from, to string }

func volumeProfileMutantChanges() []volumeProfileMutantChange {
	return []volumeProfileMutantChange{
		{"binding-slot", "candidates.push(slot);", "candidates.push(0);"},
		{"scope-containment", "if(node.pos > container.pos || node.end < container.end)", "if(node.pos > container.pos)"},
		{"first-binding", "if(candidates[candidates.length - 2] !== at)", "if(true)"},
	}
}

func volumeProfileMutantSlices() [][]volumeProfileMutantChange {
	changes := volumeProfileMutantChanges()
	return [][]volumeProfileMutantChange{changes[0:1], changes[1:2], changes[2:3]}
}

func TestVolumeProfileMutantsUnion(t *testing.T) {
	t.Parallel()
	changes, slices := volumeProfileMutantChanges(), volumeProfileMutantSlices()
	if len(slices) != testVolumeProfileMutantsShards {
		t.Fatalf("enumerated %d slices, want %d shards", len(slices), testVolumeProfileMutantsShards)
	}
	assigned := make(map[string]int, len(changes))
	for _, slice := range slices {
		for _, change := range slice {
			assigned[change.name]++
		}
	}
	if len(assigned) != len(changes) {
		t.Fatalf("shard union covers %d unique mutants, enumerated %d", len(assigned), len(changes))
	}
	for _, change := range changes {
		if assigned[change.name] != 1 {
			t.Fatalf("mutant %s assigned to %d shards, want exactly one", change.name, assigned[change.name])
		}
	}
	t.Logf("union coverage: %d source mutants in %d shards", len(assigned), len(slices))
}

func TestVolumeProfileMutants_000(t *testing.T) {
	t.Parallel()
	runVolumeProfileMutantShard(t, 0)
}

func TestVolumeProfileMutants_001(t *testing.T) {
	t.Parallel()
	runVolumeProfileMutantShard(t, 1)
}

func TestVolumeProfileMutants_002(t *testing.T) {
	t.Parallel()
	runVolumeProfileMutantShard(t, 2)
}

// Each top-level leaf owns its setup, truth, scratch space and source mutant.
// Cached products are immutable and must never be removed by a test.
func runVolumeProfileMutantShard(t *testing.T, shardIndex int) {
	t.Helper()
	started := time.Now()
	changes := volumeProfileMutantSlices()[shardIndex]
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: t.TempDir()}
	// These remain ordinary Go builds until buildcache.GoBuild lands on main.
	stage0 := filepath.Join(h.directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	oracle := volumeOracle(h, "volume-oracle", "oracle_volume.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	sources := append(volumeControls(),
		"function first(){const value=1;}function second(){const value=2;}",
		"const X=1;const C=class X<X>{m(){const X=2;}};",
		"const a=1;const b=2;function f(a:number,b:number){return a+b;}")
	var paths []string
	for i, text := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), text+"\nexport {};\n"))
	}
	paths = append(paths, h.write("native-globals.d.ts", "declare const console: {log():void};\n"), h.write("native-console.ts", "const detached=console.log;\nexport {};\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.must("controls-go", exec.Command(oracle, config, manifest))
	// Hash the actual build executables as well as the imported source tree.
	var products []string
	for _, path := range []string{stage0, archive} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		products = append(products, fmt.Sprintf("%x", sha256.Sum256(data)))
	}
	files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{"stage1/typescript", "stage1/cohere/lint"}
	for _, path := range files {
		relative, err := filepath.Rel(repository, path)
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, relative)
	}
	toolchain := []string{buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}
	t.Logf("%s (setup): %.6fs; slice coverage: %d source mutants, %d controls", t.Name(), time.Since(started).Seconds(), len(changes), len(paths))
	for _, change := range changes {
		// Include mutated source bytes explicitly, independent of scratch paths.
		digest := sha256.New()
		for _, path := range files {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == "shadow.ts" {
				if strings.Count(source, change.from) != 1 {
					t.Fatalf("nonunique mutant %s", change.name)
				}
				source = strings.Replace(source, change.from, change.to, 1)
			}
			fmt.Fprintf(digest, "%s %d\n%s", filepath.Base(path), len(source), source)
		}
		flags := append([]string{"build volume_suite.ts --tsgo checker.a", "sanitize=false", "repository=" + repository, fmt.Sprintf("mutated-source=%x", digest.Sum(nil))}, products...)
		directory := buildcache.Product(t, buildcache.Inputs{
			Name:  "typeaware volume mutant " + change.name,
			Files: inputs, Flags: flags, Toolchain: toolchain,
		}, func(directory string) error {
			builder := &harness{t: t, repository: repository, directory: directory}
			profileSourceMutant(builder, stage0, archive, change.name, change.from, change.to)
			return nil
		})
		leafStarted := time.Now()
		shard := &harness{t: t, repository: repository, directory: t.TempDir()}
		got := shard.must(change.name+"-run", exec.Command(filepath.Join(directory, change.name), config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived byte oracle", change.name)
		}
		t.Logf("%s: exit 0, independent Go byte oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
		t.Logf("leaf after product fetch: %.6fs", time.Since(leafStarted).Seconds())
		t.Logf("shard including product fetch: %.6fs; budget 60s", time.Since(started).Seconds())
	}
}
