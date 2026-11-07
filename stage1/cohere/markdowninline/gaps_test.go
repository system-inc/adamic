package markdowninline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestDelimiterExpressionMatchesNode(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("gaps/1_delimiter_expression.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	answer := onNode(test, path)
	clean(test, "Node gap", answer)
	equal(test, "Node gap", answer.stdout, []byte("a\\*b\n"))
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := filepath.Join(delimiterExpressionSanitized(test), "port")
	nativeRun := sanitizedRun(test, binary)
	clean(test, "native delimiter expression", nativeRun)
	equal(test, "native delimiter expression", nativeRun.stdout, answer.stdout)
	backend := onJavaScriptBackend(test, lowered)
	clean(test, "backend delimiter expression", backend)
	equal(test, "backend delimiter expression", backend.stdout, answer.stdout)
	if report := leaks(test, lowered, binary); report != "" {
		test.Fatalf("leaks:\n%s", report)
	}
}

// The gap's sanitized native build is a keyed product, built ahead by this test and nothing else, so
// TestDelimiterExpressionMatchesNode compiles no C (#tze35sy, Loom's build law): built in the test it was about 11 to
// 12 s of markdowninline#1/5's time (Loom's uploaded test logs, Home, Oct 18).
//
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_MarkdownInlineDelimiterExpressionSanitized(test *testing.T) {
	buildcache.RequireArtifacts(test, delimiterExpressionSanitized(test), buildcache.Artifact{Name: "port"})
}

// delimiterExpressionSanitized is gaps/1_delimiter_expression.ts lowered in this process and built by clang under the
// sanitizers, keyed by the compiler's paths, the gap's source, this file, the native options and flags, the environment
// native.Build and clang read, and clang's version.
func delimiterExpressionSanitized(test *testing.T) string {
	test.Helper()
	options := native.Options{Sanitize: true}
	flags := append(
		[]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)},
		native.Flags(options)...,
	)
	for _, name := range []string{
		"ADAMIC_NATIVE_SPLIT",
		"CPATH",
		"C_INCLUDE_PATH",
		"LIBRARY_PATH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
	} {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	inputs := buildcache.Inputs{
		Name: "markdowninline gap 1_delimiter_expression native sanitized",
		Files: append(
			compilerinputs.Lowering(),
			"stage1/cohere/markdowninline/gaps/1_delimiter_expression.ts",
			"stage1/cohere/markdowninline/gaps_test.go",
		),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		path, absolutePathError := filepath.Abs("gaps/1_delimiter_expression.ts")
		if absolutePathError != nil {
			return absolutePathError
		}
		program, loadError := load.Load([]string{path})
		if loadError != nil {
			return fmt.Errorf("Load: %w", loadError)
		}
		lowered, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return fmt.Errorf("Lower: %w", lowerError)
		}
		return native.Build(native.C(lowered), filepath.Join(directory, "port"), options)
	})
}
