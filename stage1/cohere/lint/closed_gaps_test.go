package lint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// closedComparatorGaps are the gaps compiler/area-next closed, each with what Node prints for it.
var closedComparatorGaps = []struct{ file, answer string }{
	{"3_numeric_or.ts", "2\n"},
	{"2_optional_index.ts", "1\n"},
	{"4_last_index_position.ts", "1\n"},
}

// Gaps compiler/area-next closed: each lowers now, and native (sanitized) and the JavaScript backend
// print what Node prints. GAPS.md keeps the history. Each gap's sanitized build is a keyed product, built ahead by its
// TestProduct_ClosedComparatorGap test and by nothing else, so this test compiles no C (#mq768wp, Loom's build law):
// built here, the three builds were 10.0 to 13.6 s each of a unit's 8-CPU share (Home's unit 1 of 4a769ad6's run, Oct
// 19), and the lowering and both Node runs are about a second.
func TestClosedComparatorGaps(test *testing.T) {
	test.Parallel()
	for _, probe := range closedComparatorGaps {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join("gaps", probe.file))
			if err != nil {
				test.Fatal(err)
			}
			runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			if err != nil {
				test.Fatal(err)
			}
			if answer := execute(
				test,
				"",
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				path,
			); string(answer.output) != probe.answer {
				test.Fatalf("Node prints %q, want %q", answer.output, probe.answer)
			}
			lowered, err := closedComparatorGapLowered(probe.file)
			if err != nil {
				test.Fatalf("closed gap stopped lowering: %v", err)
			}
			binary := filepath.Join(closedComparatorGapNative(test, probe.file), "gap")
			if got := execute(test, "", binary); string(got.output) != probe.answer {
				test.Fatalf("native prints %q, Node %q", got.output, probe.answer)
			}
			emitted := filepath.Join(test.TempDir(), "gap.mjs")
			if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0o644); err != nil {
				test.Fatal(err)
			}
			if got := execute(
				test,
				"",
				"node",
				"--disable-warning=ExperimentalWarning",
				runner,
				emitted,
			); string(got.output) != probe.answer {
				test.Fatalf("JavaScript backend prints %q, Node %q", got.output, probe.answer)
			}
		})
	}
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestProduct_ClosedComparatorGapNumericOr(test *testing.T) {
	buildcache.RequireArtifacts(
		test,
		closedComparatorGapNative(test, "3_numeric_or.ts"),
		buildcache.Artifact{Name: "gap"},
	)
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestProduct_ClosedComparatorGapOptionalIndex(test *testing.T) {
	buildcache.RequireArtifacts(
		test,
		closedComparatorGapNative(test, "2_optional_index.ts"),
		buildcache.Artifact{Name: "gap"},
	)
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestProduct_ClosedComparatorGapLastIndexPosition(test *testing.T) {
	buildcache.RequireArtifacts(
		test,
		closedComparatorGapNative(test, "4_last_index_position.ts"),
		buildcache.Artifact{Name: "gap"},
	)
}

// closedComparatorGapNative is the gap file built natively under the sanitizer: lowered in this process, then compiled
// by clang. The compiler lowering it is linked into this test binary, where no trace sees it read, so its key names the
// compiler's paths (compilerinputs.Lowering) beside the gap and this file, which holds the build.
func closedComparatorGapNative(test *testing.T, file string) string {
	test.Helper()
	options := native.Options{Sanitize: true}
	flags := append(
		[]string{"load.Load(" + file + ")", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)},
		native.Flags(options)...,
	)
	// native.Build splits by ADAMIC_NATIVE_SPLIT when its options don't, so the switch is an input.
	flags = append(flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	inputs := buildcache.Inputs{
		Name: "lint-closed-comparator-gap-" + strings.TrimSuffix(file, ".ts") + "-sanitized",
		Files: append(
			compilerinputs.Lowering(),
			"stage1/cohere/lint/gaps/"+file,
			"stage1/cohere/lint/closed_gaps_test.go",
		),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		lowered, err := closedComparatorGapLowered(file)
		if err != nil {
			return err
		}
		return native.Build(native.C(lowered), filepath.Join(directory, "gap"), options)
	})
}

// closedComparatorGapLowered loads and lowers one gap file.
func closedComparatorGapLowered(file string) (*ir.Program, error) {
	path, err := filepath.Abs(filepath.Join("gaps", file))
	if err != nil {
		return nil, err
	}
	program, err := load.Load([]string{path})
	if err != nil {
		return nil, err
	}
	return lower.Lower(context.Background(), program)
}
