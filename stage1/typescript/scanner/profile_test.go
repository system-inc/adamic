package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Opt-in artifacts outlive test.TempDir so callgrind and the timing runner use the same snapshot.
// Not parallel: ADAMIC_SCANNER_PROFILE_DIR artifact paths.
func TestProfileArtifacts(test *testing.T) {
	directory := os.Getenv("ADAMIC_SCANNER_PROFILE_DIR")
	if directory == "" {
		test.Skip("set ADAMIC_SCANNER_PROFILE_DIR to a scratch directory")
	}
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	for _, name := range portFiles {
		data, readError := os.ReadFile(name)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	files := sourceFiles(test, compilerSource(test))
	var manifest strings.Builder
	for _, path := range files {
		manifest.WriteString("scan\t" + path + "\n")
	}
	if writeError := os.WriteFile(
		filepath.Join(directory, "compiler.txt"),
		[]byte(manifest.String()),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	oracle, readError := os.ReadFile(goOracle(test))
	if readError != nil {
		test.Fatal(readError)
	}
	if writeError := os.WriteFile(filepath.Join(directory, "oracle"), oracle, 0755); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := native.C(lowered)
	if writeError := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if buildError := native.Build(source, filepath.Join(directory, "scanner"), native.Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	if buildError := native.Build(
		source,
		filepath.Join(directory, "counted"),
		native.Options{Count: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
	paths := corpusfiles.Repository(test, repository, []string{"internal/native/runtime"}, []string{"*.c", "*.h"})
	units := []string{filepath.Join(directory, "main.c")}
	for _, sourcePath := range paths {
		name := filepath.Base(sourcePath)
		data, readError := os.ReadFile(sourcePath)
		if readError != nil {
			test.Fatal(readError)
		}
		path := filepath.Join(directory, name)
		if writeError := os.WriteFile(path, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		if strings.HasSuffix(name, ".c") {
			units = append(units, path)
		}
	}
	// The runtime's units are compiled with the program's features as flags, as main.c's #defines say.
	flags := append(native.SourceFlags(source, native.Options{}), "-g", "-o", filepath.Join(directory, "profiled"))
	flags = append(flags, units...)
	flags = append(flags, "-lm")
	execute(test, "", "clang", flags...)
	test.Logf("release, counted and -O2 -g profiling builds saved in %s", directory)
}

// The measured release builds, not just sanitized builds, must retain the entire answer protocol.
func TestProfileSnapshotsAgree(test *testing.T) {
	test.Parallel()
	asked := os.Getenv("ADAMIC_SCANNER_PROFILE_SNAPSHOTS")
	if asked == "" {
		test.Skip("set ADAMIC_SCANNER_PROFILE_SNAPSHOTS to the artifact directories")
	}
	corpus := askedCorpus(test)
	want := execute(test, "", goOracle(test), "--manifest", corpus.all)
	for _, directory := range filepath.SplitList(asked) {
		for _, side := range []struct {
			name string
			run  execution
		}{
			{"release", execute(test, "", filepath.Join(directory, "scanner"), "--manifest", corpus.all)},
			{"profiled", execute(test, "", filepath.Join(directory, "profiled"), "--manifest", corpus.all)},
			{"Node", node(test, directory, corpus.all, false)},
		} {
			if diff := difference(side.run.output, want.output); diff != "" {
				test.Fatalf("%s %s: %s", directory, side.name, diff)
			}
		}
		test.Logf(
			"%s: release, -O2 -g profiled and Node: %d identical answer bytes",
			filepath.Base(directory),
			len(want.output),
		)
	}
}
