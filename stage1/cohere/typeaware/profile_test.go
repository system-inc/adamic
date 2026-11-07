package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func profileSourceMutant(harness *harness, stage0, archive, name, from, to string) string {
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
	binary := harness.build(stage0, name, filepath.Join(directory, "volume_suite.ts"), archive, false)
	// The source's imports name this checkout for the build above; the product names it <repository> (#tqrqx60).
	if relativeFilesError := buildcache.RelativeFiles(directory); relativeFilesError != nil {
		harness.t.Fatal(relativeFilesError)
	}
	return binary
}

func volumeProfileMissingBinding(h *harness, stage0, archive string) string {
	return volumeProfileNative(
		h,
		stage0,
		archive,
		false,
		"typeaware volume missing-binding namedScopes wrong",
		func(local *harness) string {
			return profileSourceMutant(
				local,
				stage0,
				archive,
				"missing-binding",
				"this.namedScopes.set(binding.name, fresh);",
				"this.namedScopes.set('wrong', fresh);",
			)
		},
		"this.namedScopes.set(binding.name, fresh);",
		"this.namedScopes.set('wrong', fresh);",
	)
}

func TestShadowIndexMissingBinding_000(test *testing.T) {
	test.Parallel()
	started := time.Now()
	harness := volumeProfileHarness(test)
	stage0 := volumeProfileStage0(harness)
	archive := harness.archive(false)
	test.Logf("setup: %.3fs", time.Since(started).Seconds())
	mutant := volumeProfileMissingBinding(harness, stage0, archive)
	source := harness.write("input.ts", "const x=1;\nexport {};\n")
	manifest := harness.write("manifest", source+"\n")
	got := harness.run(
		"missing-binding-run",
		exec.Command(
			mutant,
			filepath.Join(harness.repository, "stage1/cohere/typeaware/testdata/tsconfig.json"),
			manifest,
		),
	)
	if code, ok := got.err.(*exec.ExitError); !ok ||
		code.ExitCode() != 70 ||
		!bytes.Contains(got.stderr, []byte("missing binding index")) {
		test.Fatalf("missing binding index was not refused: %v %s", got.err, got.stderr)
	}
	test.Log("omitted binding key: compiled mutant panics 70, missing binding index; cooked=false")
}
