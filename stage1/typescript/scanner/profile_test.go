package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Opt-in artifacts outlive t.TempDir so callgrind and the timing runner use the same snapshot.
func TestProfileArtifacts(t *testing.T) {
	directory := os.Getenv("ADAMIC_SCANNER_PROFILE_DIR")
	if directory == "" {
		// census: measurement Opt-in timing or profile artifact comparison; does not replace correctness verification.
		t.Skip("set ADAMIC_SCANNER_PROFILE_DIR to a scratch directory")
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range portFiles {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	files := sourceFiles(t, compilerSource(t))
	var manifest strings.Builder
	for _, path := range files {
		manifest.WriteString("scan\t" + path + "\n")
	}
	if err := os.WriteFile(filepath.Join(directory, "compiler.txt"), []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	oracle, err := os.ReadFile(goOracle(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "oracle"), oracle, 0755); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(lowered)
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := native.Build(source, filepath.Join(directory, "scanner"), native.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := native.Build(source, filepath.Join(directory, "counted"), native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	runtime := filepath.Join(repository, "internal/native/runtime")
	entries, err := os.ReadDir(runtime)
	if err != nil {
		t.Fatal(err)
	}
	units := []string{filepath.Join(directory, "main.c")}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(runtime, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(directory, entry.Name())
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(entry.Name(), ".c") {
			units = append(units, path)
		}
	}
	flags := append(native.Flags(native.Options{}), "-g", "-o", filepath.Join(directory, "profiled"))
	flags = append(flags, units...)
	flags = append(flags, "-lm")
	execute(t, "", "clang", flags...)
	t.Logf("release, counted and -O2 -g profiling builds saved in %s", directory)
}

// The measured release builds, not just sanitized builds, must retain the entire answer protocol.
func TestProfileSnapshotsAgree(t *testing.T) {
	asked := os.Getenv("ADAMIC_SCANNER_PROFILE_SNAPSHOTS")
	if asked == "" {
		// census: measurement Opt-in timing or profile artifact comparison; does not replace correctness verification.
		t.Skip("set ADAMIC_SCANNER_PROFILE_SNAPSHOTS to the artifact directories")
	}
	corpus := askedCorpus(t)
	want := execute(t, "", goOracle(t), "--manifest", corpus.all)
	for _, directory := range filepath.SplitList(asked) {
		for _, side := range []struct {
			name string
			run  execution
		}{
			{"release", execute(t, "", filepath.Join(directory, "scanner"), "--manifest", corpus.all)},
			{"profiled", execute(t, "", filepath.Join(directory, "profiled"), "--manifest", corpus.all)},
			{"Node", node(t, directory, corpus.all, false)},
		} {
			if diff := difference(side.run.output, want.output); diff != "" {
				t.Fatalf("%s %s: %s", directory, side.name, diff)
			}
		}
		t.Logf("%s: release, -O2 -g profiled and Node: %d identical answer bytes", filepath.Base(directory), len(want.output))
	}
}
