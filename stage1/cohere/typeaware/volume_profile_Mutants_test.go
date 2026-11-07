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
		{
			"scope-containment",
			"if(node.pos > container.pos || node.end < container.end)",
			"if(node.pos > container.pos)",
		},
		{"first-binding", "if(candidates[candidates.length - 2] !== at)", "if(true)"},
	}
}

func volumeProfileMutantSlices() [][]volumeProfileMutantChange {
	changes := volumeProfileMutantChanges()
	return [][]volumeProfileMutantChange{changes[0:1], changes[1:2], changes[2:3]}
}

func TestVolumeProfileMutantsUnion(test *testing.T) {
	test.Parallel()
	changes, slices := volumeProfileMutantChanges(), volumeProfileMutantSlices()
	if len(slices) != testVolumeProfileMutantsShards {
		test.Fatalf("enumerated %d slices, want %d shards", len(slices), testVolumeProfileMutantsShards)
	}
	assigned := make(map[string]int, len(changes))
	for _, slice := range slices {
		for _, change := range slice {
			assigned[change.name]++
		}
	}
	if len(assigned) != len(changes) {
		test.Fatalf("shard union covers %d unique mutants, enumerated %d", len(assigned), len(changes))
	}
	for _, change := range changes {
		if assigned[change.name] != 1 {
			test.Fatalf("mutant %s assigned to %d shards, want exactly one", change.name, assigned[change.name])
		}
	}
	test.Logf("union coverage: %d source mutants in %d shards", len(assigned), len(slices))
}

func TestVolumeProfileMutants_000(test *testing.T) {
	test.Parallel()
	runVolumeProfileMutantShard(test, 0)
}

func TestVolumeProfileMutants_001(test *testing.T) {
	test.Parallel()
	runVolumeProfileMutantShard(test, 1)
}

func TestVolumeProfileMutants_002(test *testing.T) {
	test.Parallel()
	runVolumeProfileMutantShard(test, 2)
}

// Each top-level leaf owns its setup, truth, scratch space and source mutant.
// Cached products are immutable and must never be removed by a test.
func runVolumeProfileMutantShard(test *testing.T, shardIndex int) {
	test.Helper()
	started := time.Now()
	changes := volumeProfileMutantSlices()[shardIndex]
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	h := &harness{t: test, repository: repository, directory: test.TempDir()}
	oracle := volumeProfileMutantOracle(h)
	configuration := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	sources := append(volumeControls(),
		"function first(){const value=1;}function second(){const value=2;}",
		"const X=1;const C=class X<X>{m(){const X=2;}};",
		"const a=1;const b=2;function f(a:number,b:number){return a+b;}")
	var paths []string
	for index, text := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", index), text+"\nexport {};\n"))
	}
	paths = append(
		paths,
		h.write("native-globals.d.ts", "declare const console: {log():void};\n"),
		h.write("native-console.ts", "const detached=console.log;\nexport {};\n"),
	)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.must("controls-go", exec.Command(oracle, configuration, manifest))
	test.Logf(
		"%s (setup): %.6fs; slice coverage: %d source mutants, %d controls",
		test.Name(),
		time.Since(started).Seconds(),
		len(changes),
		len(paths),
	)
	for _, change := range changes {
		binary := volumeProfileMutantProduct(h, change)
		leafStarted := time.Now()
		shard := &harness{t: test, repository: repository, directory: test.TempDir()}
		got := shard.must(change.name+"-run", exec.Command(binary, configuration, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			test.Fatalf("%s mutant survived byte oracle", change.name)
		}
		test.Logf(
			"%s: exit 0, independent Go byte oracle catches byte %d; %s",
			change.name,
			firstDifference(got.stdout, truth.stdout),
			summary(got.stdout),
		)
		test.Logf("leaf after product fetch: %.6fs", time.Since(leafStarted).Seconds())
		test.Logf("shard including product fetch: %.6fs; budget 60s", time.Since(started).Seconds())
	}
}

func volumeProfileMutantProduct(h *harness, change volumeProfileMutantChange) string {
	test, repository := h.t, h.repository
	stage0, archive := h.stage0(), h.archive(false)
	// Hash the actual build executables as well as the imported source tree.
	var products []string
	for _, path := range []string{stage0, archive} {
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		products = append(products, fmt.Sprintf("%x", sha256.Sum256(data)))
	}
	files, globError := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*.ts"))
	if globError != nil {
		test.Fatal(globError)
	}
	inputs := []string{"stage1/typescript", "stage1/cohere/lint"}
	for _, path := range files {
		relative, relativePathError := filepath.Rel(repository, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		inputs = append(inputs, relative)
	}
	toolchain := []string{buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}
	// Include mutated source bytes explicitly, independent of scratch paths.
	digest := sha256.New()
	for _, path := range files {
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(data)
		if filepath.Base(path) == "shadow.ts" {
			if strings.Count(source, change.from) != 1 {
				test.Fatalf("nonunique mutant %s", change.name)
			}
			source = strings.Replace(source, change.from, change.to, 1)
		}
		fmt.Fprintf(digest, "%s %d\n%s", filepath.Base(path), len(source), source)
	}
	flags := append([]string{
		"build volume_suite.ts --tsgo checker.a",
		"sanitize=false",
		"repository=" + repository,
		fmt.Sprintf("mutated-source=%x", digest.Sum(nil)),
	}, products...)
	directory := buildcache.Product(test, buildcache.Inputs{
		Name:  "typeaware volume mutant " + change.name,
		Files: inputs, Flags: flags, Toolchain: toolchain,
	}, func(directory string) error {
		// The mutant's build is the product's fill, shared setup the unit's own budget bounds, like every other
		// product build here, never typeAwareChildLimit's 90 s for case work: on a loaded builder that clock, not the
		// code, decided the verdict (Eve's tree ffd70d0d killed all three builds at 90.1 s, #h4hrxpe).
		builder := &harness{t: test, repository: repository, directory: directory, sharedSetup: true}
		volumeProfileSourceMutant(builder, stage0, archive, change.name, change.from, change.to)
		return nil
	})
	return filepath.Join(directory, change.name)
}

func volumeProfileMutantOracle(harness *harness) string {
	harness.t.Helper()
	return volumeOracle(harness, "volume-oracle", "oracle_volume.go")
}

func volumeProfileSourceMutant(harness *harness, stage0, archive, name, from, to string) string {
	harness.t.Helper()
	directory := filepath.Join(harness.directory, name+"-source")
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		harness.t.Fatal(createDirectoryError)
	}
	sourceDirectory := filepath.Join(harness.repository, "stage1/cohere/typeaware")
	files, globError := filepath.Glob(filepath.Join(sourceDirectory, "*.ts"))
	if globError != nil {
		harness.t.Fatal(globError)
	}
	for _, path := range files {
		data, readError := os.ReadFile(path)
		if readError != nil {
			harness.t.Fatal(readError)
		}
		source := string(data)
		if filepath.Base(path) == "shadow.ts" {
			if strings.Count(source, from) != 1 {
				harness.t.Fatalf("nonunique mutant %s", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(
			source,
			"../../typescript/",
			filepath.Join(harness.repository, "stage1/typescript")+"/",
		)
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(harness.repository, "stage1/cohere/lint")+"/")
		if writeError := os.WriteFile(
			filepath.Join(directory, filepath.Base(path)),
			[]byte(source),
			0600,
		); writeError != nil {
			harness.t.Fatal(writeError)
		}
	}
	binary := filepath.Join(harness.directory, name)
	harness.must(
		name,
		exec.Command(stage0, "build", filepath.Join(directory, "volume_suite.ts"), "-o", binary, "--tsgo", archive),
	)
	// The source's imports name this checkout for the build above; the product names it <repository> (#tqrqx60).
	if relativeFilesError := buildcache.RelativeFiles(directory); relativeFilesError != nil {
		harness.t.Fatal(relativeFilesError)
	}
	return binary
}
