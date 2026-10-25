package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// splitTSGoOptions are the options both the whole baseline and every split build are built with.
var splitTSGoOptions = Options{Sanitize: true, Jobs: 5}

// splitTSGoSource is the C both units build: typeaware's main.ts loaded, lowered and emitted with TSGoC, a product
// built ahead by TestProduct_SplitTSGoSource. Every unit used to load and lower it again, a few seconds each. Built
// in process, so its declared inputs are every file go compiles into the packages that make it (load, lowering and
// emission), the program's own sources with everything they import (typeaware's unary_minus.ts reaches lint's
// finding.ts and its suggestions.a) and this recipe; under the tree builder's trace its key is what it read.
func splitTSGoSource(t *testing.T) string {
	t.Helper()
	compiler, err := buildcache.GoInputs("main.c", "./internal/load ./internal/lower ./internal/native", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.Product(t, buildcache.Inputs{
		Name: "split-tsgo-source",
		Files: append([]string{"stage1/cohere/typeaware", "stage1/cohere/lint/finding.ts", "stage1/cohere/lint/suggestions.a",
			"stage1/typescript", "internal/native/units_tsgo_test.go"}, compiler.Files...),
		Flags:     compiler.Flags,
		Toolchain: compiler.Toolchain,
	}, func(directory string) error {
		program, err := load.Load([]string{"../../stage1/cohere/typeaware/main.ts"})
		if err != nil {
			return err
		}
		program.EnableTSGo()
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		source, err := TSGoC(lowered)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644)
	})
	source, err := os.ReadFile(filepath.Join(directory, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

// splitTSGoBaseline is the whole sanitized checker both units compare their split builds with, BuildTSGo of the
// source product against the archive product, a product built ahead by TestProduct_SplitTSGoBaseline. Both inputs
// are products, so they key it by their bytes: the same bytes wherever they were fetched.
func splitTSGoBaseline(t *testing.T, source, archive string) string {
	t.Helper()
	archiveBytes, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	optionBytes, err := json.Marshal(splitTSGoOptions)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.Product(t, buildcache.Inputs{
		Name:      "split-tsgo-baseline",
		Files:     []string{"internal/native", "bridge/tsgo"},
		Flags:     []string{fmt.Sprintf("source=%x", sha256.Sum256([]byte(source))), fmt.Sprintf("archive=%x", sha256.Sum256(archiveBytes)), "options=" + string(optionBytes)},
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}, func(directory string) error {
		// Baseline compilation is a separate product unit, never a behavior unit's setup.
		if !strings.HasPrefix(t.Name(), "TestProduct_") {
			return fmt.Errorf("split-tsgo-baseline must be built ahead by TestProduct_SplitTSGoBaseline")
		}
		return BuildTSGo(source, filepath.Join(directory, "whole"), archive, splitTSGoOptions)
	})
	return filepath.Join(directory, "whole")
}

// TestProduct_SplitTSGoSource builds the C the split units build, so the tree build builds it ahead.
func TestProduct_SplitTSGoSource(t *testing.T) {
	t.Parallel()
	splitTSGoSource(t)
}

// TestProduct_SplitTSGoBaseline builds the whole checker the split units compare with, so the tree build builds it
// ahead.
func TestProduct_SplitTSGoBaseline(t *testing.T) {
	t.Parallel()
	splitTSGoBaseline(t, splitTSGoSource(t), tsgoArchive(t))
}

// Not parallel: this test explicitly compares the object cache with its bypass on the same input. Each unit's own
// work is its cache mode's split build and the comparison; the source, the archive and the whole baseline are
// products built ahead.
func runSplitTSGoAgreesUnit(t *testing.T, unit int) {
	t.Helper()
	phase := time.Now()
	archive := tsgoArchive(t)
	source := splitTSGoSource(t)
	whole := splitTSGoBaseline(t, source, archive)
	t.Logf("product fetch phase %.3fs", time.Since(phase).Seconds())
	phase = time.Now()
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
	want, err := exec.Command(whole, config, manifest).CombinedOutput()
	if err != nil {
		t.Fatalf("whole checker: %v\n%s", err, want)
	}
	t.Logf("whole checker run phase %.3fs", time.Since(phase).Seconds())
	shard := testShard{unit, 2}
	for index, uncached := range splitCacheModes {
		if !shard.owns(index) {
			continue
		}
		t.Run("cache_"+uncached, func(t *testing.T) {
			t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
			split := filepath.Join(directory, "split")
			if err := BuildSplitTSGo(source, split, archive, splitTSGoOptions); err != nil {
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
