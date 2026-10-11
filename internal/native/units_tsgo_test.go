package native

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// tsgoArchive is the checker archive (./bridge/tsgo/archive as a c-archive), a product built ahead by
// TestProduct_TSGoArchive. The units used to read it from ADAMIC_CLANG_TSGO_ARCHIVE, which only Workshop's env.sh
// set, so on a runner both skipped (landable-15: internal/native's skip census).
func tsgoArchive(t *testing.T) string {
	t.Helper()
	return buildcache.GoBuild(t, "tsgo.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"})
}

// TestProduct_TSGoArchive builds the checker archive the split units link, so the tree build builds it ahead.
func TestProduct_TSGoArchive(t *testing.T) {
	t.Parallel()
	tsgoArchive(t)
}

// Not parallel: this test explicitly compares the object cache with its bypass on the same input.
func runSplitTSGoAgreesUnit(t *testing.T, unit int) {
	t.Helper()
	archive := tsgoArchive(t)
	program, err := load.Load([]string{"../../stage1/cohere/typeaware/main.ts"})
	if err != nil {
		t.Fatal(err)
	}
	program.EnableTSGo()
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source, err := TSGoC(lowered)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	input := filepath.Join(directory, "sample.a")
	config := filepath.Join(directory, "tsconfig.json")
	manifest := filepath.Join(directory, "manifest")
	for name, contents := range map[string]string{
		input:    "const value = 1; const negative = -value;\n",
		config:   `{"compilerOptions":{"strict":true,"target":"esnext"},"files":["sample.a"]}`,
		manifest: input + "\n",
	} {
		if err := os.WriteFile(name, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	options := Options{Sanitize: true, Jobs: 5}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cache, err := newTestBuildCache(repository)
	if err != nil {
		t.Fatal(err)
	}
	archiveBytes, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	archiveProduct, err := cache.Tree("checker-archive", [][]byte{archiveBytes}, func(destination string) error {
		return os.WriteFile(filepath.Join(destination, "tsgo.a"), archiveBytes, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	archive = filepath.Join(archiveProduct, "tsgo.a")
	optionBytes, _ := json.Marshal(options)
	built, err := cache.Tree("whole-tsgo-baseline", [][]byte{[]byte(source), archiveBytes, optionBytes}, func(destination string) error {
		return BuildTSGo(source, filepath.Join(destination, "whole"), archive, options)
	})
	if err != nil {
		t.Fatal(err)
	}
	whole := filepath.Join(built, "whole")
	want, err := exec.Command(whole, config, manifest).CombinedOutput()
	if err != nil {
		t.Fatalf("whole checker: %v\n%s", err, want)
	}
	shard := testShard{unit, 2}
	for index, uncached := range splitCacheModes {
		if !shard.owns(index) {
			continue
		}
		t.Run("cache_"+uncached, func(t *testing.T) {
			t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
			split := filepath.Join(directory, "split")
			if err := BuildSplitTSGo(source, split, archive, options); err != nil {
				t.Fatal(err)
			}
			got, err := exec.Command(split, config, manifest).CombinedOutput()
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("checker uncached=%s: %v\nwhole=%s\nsplit=%s", uncached, err, want, got)
			}
			t.Logf("checker uncached=%s: %d identical bytes: %s", uncached, len(got), got)
		})
	}
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestSplitTSGoAgreesUnit00(t *testing.T) {
	runSplitTSGoAgreesUnit(t, 0)
}

// Not parallel: this unit changes process environment or uses the WASI toolchain.
func TestSplitTSGoAgreesUnit01(t *testing.T) {
	runSplitTSGoAgreesUnit(t, 1)
}
