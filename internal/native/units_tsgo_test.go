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
func tsgoArchive(test *testing.T) string {
	test.Helper()
	return buildcache.GoBuild(test, "tsgo.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"})
}

// TestProduct_TsGoArchive builds the checker archive the split units link, so the tree build builds it ahead.
func TestProduct_TsGoArchive(test *testing.T) {
	test.Parallel()
	tsgoArchive(test)
}

// splitTsGoOptions are the options both the whole baseline and every split build are built with.
var splitTsGoOptions = Options{Sanitize: true, Jobs: 5}

// splitTsGoSource is the C both units build: typeaware's main.ts loaded, lowered and emitted with TSGoC, a product
// built ahead by TestProduct_SplitTSGoSource. Every unit used to load and lower it again, a few seconds each. Built
// in process, so its declared inputs are every file go compiles into the packages that make it (load, lowering and
// emission), the program's own sources with everything they import (typeaware's unary_minus.ts reaches lint's
// finding.ts and its suggestions.a) and this recipe; under the tree builder's trace its key is what it read.
func splitTsGoSource(test *testing.T) string {
	test.Helper()
	compiler, goInputsError := buildcache.GoInputs(
		"main.c",
		"./internal/load ./internal/lower ./internal/native",
		nil,
		nil,
	)
	if goInputsError != nil {
		test.Fatal(goInputsError)
	}
	directory := buildcache.Product(test, buildcache.Inputs{
		Name: "split-tsgo-source",
		Files: append(
			[]string{"stage1/cohere/typeaware", "stage1/cohere/lint/finding.ts", "stage1/cohere/lint/suggestions.a",
				"stage1/typescript", "internal/native/units_tsgo_test.go"},
			compiler.Files...,
		),
		Flags:     compiler.Flags,
		Toolchain: compiler.Toolchain,
	}, func(directory string) error {
		program, loadError := load.Load([]string{"../../stage1/cohere/typeaware/main.ts"})
		if loadError != nil {
			return loadError
		}
		program.EnableTsGo()
		lowered, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		source, tsGoCError := TsGoC(lowered)
		if tsGoCError != nil {
			return tsGoCError
		}
		return os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644)
	})
	source, readError := os.ReadFile(filepath.Join(directory, "main.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

// splitTsGoBaseline is the whole sanitized checker both units compare their split builds with, BuildTSGo of the
// source product against the archive product, a product built ahead by TestProduct_SplitTSGoBaseline. Both inputs
// are products, so they key it by their bytes: the same bytes wherever they were fetched.
func splitTsGoBaseline(test *testing.T, source, archive string) string {
	test.Helper()
	archiveBytes, readError := os.ReadFile(archive)
	if readError != nil {
		test.Fatal(readError)
	}
	optionBytes, encodeError := json.Marshal(splitTsGoOptions)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	directory := buildcache.Product(test, buildcache.Inputs{
		Name:  "split-tsgo-baseline",
		Files: []string{"internal/native", "bridge/tsgo"},
		Flags: []string{
			fmt.Sprintf("source=%x", sha256.Sum256([]byte(source))),
			fmt.Sprintf("archive=%x", sha256.Sum256(archiveBytes)),
			"options=" + string(optionBytes),
		},
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}, func(directory string) error {
		// Baseline compilation is a separate product unit, never a behavior unit's setup.
		if !strings.HasPrefix(test.Name(), "TestProduct_") {
			return fmt.Errorf("split-tsgo-baseline must be built ahead by TestProduct_SplitTSGoBaseline")
		}
		return BuildTSGo(source, filepath.Join(directory, "whole"), archive, splitTsGoOptions)
	})
	return filepath.Join(directory, "whole")
}

// TestProduct_SplitTsGoSource builds the C the split units build, so the tree build builds it ahead.
func TestProduct_SplitTsGoSource(test *testing.T) {
	test.Parallel()
	splitTsGoSource(test)
}

// TestProduct_SplitTSGoBaseline builds the whole checker the split units compare with, so the tree build builds it
// ahead.
func TestProduct_SplitTSGoBaseline(test *testing.T) {
	test.Parallel()
	splitTsGoBaseline(test, splitTsGoSource(test), tsgoArchive(test))
}

// splitTsGoSplit is the split sanitized checker built from the source product against the archive product with the
// units cache bypassed, a product built ahead by TestProduct_SplitTSGoSplit. Its compile is the whole of the agreement
// test's cost: 23 units and about 17 s of clang quiet, 24 s at twice the box's load, one 1 MB unit the long pole
// (#dgcy49s). Bypassed, its build reads no units cache, so nothing outside its inputs reaches it; a build mixing hits
// and misses is TestSplitMixesHitsAndMisses's. Keyed as the baseline is, by its inputs' bytes.
func splitTsGoSplit(test *testing.T, source, archive string) string {
	test.Helper()
	test.Setenv("ADAMIC_GATE_UNCACHED", "1")
	archiveBytes, readError := os.ReadFile(archive)
	if readError != nil {
		test.Fatal(readError)
	}
	optionBytes, encodeError := json.Marshal(splitTsGoOptions)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	directory := buildcache.Product(test, buildcache.Inputs{
		Name:  "split-tsgo-split",
		Files: []string{"internal/native", "bridge/tsgo"},
		Flags: []string{
			fmt.Sprintf("source=%x", sha256.Sum256([]byte(source))),
			fmt.Sprintf("archive=%x", sha256.Sum256(archiveBytes)),
			"options=" + string(optionBytes),
			"uncached=1",
		},
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}, func(directory string) error {
		return BuildSplitTSGo(source, filepath.Join(directory, "split"), archive, splitTsGoOptions)
	})
	return filepath.Join(directory, "split")
}

// TestProduct_SplitTsGoSplit builds the split checker the agreement test runs, so the tree build builds it ahead.
// Not parallel: splitTSGoSplit sets ADAMIC_GATE_UNCACHED for the process.
func TestProduct_SplitTsGoSplit(test *testing.T) {
	archive := tsgoArchive(test)
	splitTsGoSplit(test, splitTsGoSource(test), archive)
}

// The split checker, built ahead, prints what the whole one prints. Each was compiled by a product test, so this test
// is a fetch, two runs and a compare. It used to compile the split checker itself, once per cache mode: a cached build
// on a box with an empty units cache was the bypass's compile plus writes, about 27 s at twice the box's load and 60.73
// s at load 82, which reddened a change that touched nothing here (#dgcy49s). Not parallel: splitTSGoSplit sets
// ADAMIC_GATE_UNCACHED for the process.
func TestSplitTsGoAgrees(test *testing.T) {
	phase := time.Now()
	archive := tsgoArchive(test)
	source := splitTsGoSource(test)
	whole := splitTsGoBaseline(test, source, archive)
	split := splitTsGoSplit(test, source, archive)
	test.Logf("product fetch phase %.3fs", time.Since(phase).Seconds())
	directory := test.TempDir()
	input := filepath.Join(directory, "sample.a")
	configuration := filepath.Join(directory, "tsconfig.json")
	manifest := filepath.Join(directory, "manifest")
	for name, contents := range map[string]string{
		input:         "const value = 1; const negative = -value;\n",
		configuration: `{"compilerOptions":{"strict":true,"target":"esnext"},"files":["sample.a"]}`,
		manifest:      input + "\n",
	} {
		if writeError := os.WriteFile(name, []byte(contents), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	want, combinedOutputError := exec.Command(whole, configuration, manifest).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("whole checker: %v\n%s", combinedOutputError, want)
	}
	got, combinedOutputError := exec.Command(split, configuration, manifest).CombinedOutput()
	if combinedOutputError != nil || !bytes.Equal(got, want) {
		test.Fatalf("split checker: %v\nwhole=%s\nsplit=%s", combinedOutputError, want, got)
	}
	test.Logf("split checker: %d identical bytes: %s", len(got), got)
}
