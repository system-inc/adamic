package tsprinter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// The .txt suffix keeps this separately executed native gap out of the formatter corpus.
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClassInterfaceMethodGap(test *testing.T) {
	data, readError := os.ReadFile("gaps/classInterfaceMethod.ts.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	directory := test.TempDir()
	port := filepath.Join(directory, "gap.ts")
	if writeError := os.WriteFile(port, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	node := onNode(test, port)
	if node.ExitCode != 0 || len(node.Stderr) != 0 || string(node.Stdout) != "17\n" {
		test.Fatalf("Node truth: %+v", node)
	}
	loaded, loadError := load.Load([]string{port})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), loaded)
	isGap := func(lowerError error) bool {
		var diagnostic *lower.NotYet
		return errors.As(lowerError, &diagnostic) &&
			diagnostic.Where == port+":6:28" &&
			diagnostic.What == "a class method through a view that erases its prototype origin"
	}
	if !isGap(lowerError) {
		test.Fatalf("recorded lowering gap changed: %v", lowerError)
	}
	test.Logf("Node prints 17; lowering refuses before native emission: %s", lowerError)
	// A safe explicit callback is also the mutant proving this gap check can reject a normal run.
	fixed, rewrittenError := rewritten(string(data), classInterfaceWorkaroundSource.replacements)
	if rewrittenError != nil {
		test.Fatal(rewrittenError)
	}
	workaround := filepath.Join(directory, "workaround.ts")
	if writeError := os.WriteFile(workaround, []byte(fixed), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError = load.Load([]string{workaround})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil || isGap(lowerError) {
		test.Fatalf("explicit callback mutant lowering: %v", lowerError)
	}
	binary := filepath.Join(sourceSanitized(test, classInterfaceWorkaroundSource), "port")
	safe := sanitizedRun(test, binary)
	for _, result := range []oracles.Run{onNode(test, workaround), safe, onJavaScriptBackend(test, program)} {
		if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != "17\n" {
			test.Fatalf("explicit callback: %+v", result)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Log(
		"explicit callback prints 17 on Node, native and backend; leak-free; gap check rejects this normal-run mutant",
	)
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedOptionalBooleanFunctionGap(test *testing.T) {
	data, readError := os.ReadFile("gaps/optionalBooleanFunction.ts.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	path := filepath.Join(test.TempDir(), "gap.a")
	if writeError := os.WriteFile(path, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program := lowered(test, path)
	binary := filepath.Join(sourceSanitized(test, optionalBooleanFunctionSource), "port")
	nativeRun := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result oracles.Run
	}{
		{"Node", onNode(test, path)},
		{"native ASan/UBSan", nativeRun},
		{"JavaScript backend", onJavaScriptBackend(test, program)},
	} {
		if side.result.ExitCode != 0 || len(side.result.Stderr) != 0 || string(side.result.Stdout) != "absent\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.ExitCode, side.result.Stdout, side.result.Stderr)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}
