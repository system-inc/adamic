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

func TestDelimiterExpressionMatchesNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("gaps/1_delimiter_expression.ts")
	if err != nil {
		t.Fatal(err)
	}
	answer := onNode(t, path)
	clean(t, "Node gap", answer)
	equal(t, "Node gap", answer.stdout, []byte("a\\*b\n"))
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(delimiterExpressionSanitized(t), "port")
	nativeRun := sanitizedRun(t, binary)
	clean(t, "native delimiter expression", nativeRun)
	equal(t, "native delimiter expression", nativeRun.stdout, answer.stdout)
	backend := onJavaScriptBackend(t, lowered)
	clean(t, "backend delimiter expression", backend)
	equal(t, "backend delimiter expression", backend.stdout, answer.stdout)
	if report := leaks(t, lowered, binary); report != "" {
		t.Fatalf("leaks:\n%s", report)
	}
}

// The gap's sanitized native build is a keyed product, built ahead by this test and nothing else, so
// TestDelimiterExpressionMatchesNode compiles no C (#tze35sy, Loom's build law): built in the test it was about 11 to
// 12 s of markdowninline#1/5's time (Loom's uploaded test logs, Home, Oct 18).
//
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_MarkdownInlineDelimiterExpressionSanitized(t *testing.T) {
	buildcache.RequireArtifacts(t, delimiterExpressionSanitized(t), buildcache.Artifact{Name: "port"})
}

// delimiterExpressionSanitized is gaps/1_delimiter_expression.ts lowered in this process and built by clang under the
// sanitizers, keyed by the compiler's paths, the gap's source, this file, the native options and flags, the environment
// native.Build and clang read, and clang's version.
func delimiterExpressionSanitized(t *testing.T) string {
	t.Helper()
	options := native.Options{Sanitize: true}
	flags := append([]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	inputs := buildcache.Inputs{
		Name:      "markdowninline gap 1_delimiter_expression native sanitized",
		Files:     append(compilerinputs.Lowering(), "stage1/cohere/markdowninline/gaps/1_delimiter_expression.ts", "stage1/cohere/markdowninline/gaps_test.go"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		path, err := filepath.Abs("gaps/1_delimiter_expression.ts")
		if err != nil {
			return err
		}
		program, err := load.Load([]string{path})
		if err != nil {
			return fmt.Errorf("Load: %w", err)
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return fmt.Errorf("Lower: %w", err)
		}
		return native.Build(native.C(lowered), filepath.Join(directory, "port"), options)
	})
}
