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
	data, err := os.ReadFile("gaps/classInterfaceMethod.ts.txt")
	if err != nil {
		test.Fatal(err)
	}
	directory := test.TempDir()
	port := filepath.Join(directory, "gap.ts")
	if err = os.WriteFile(port, data, 0644); err != nil {
		test.Fatal(err)
	}
	node := onNode(test, port)
	if node.ExitCode != 0 || len(node.Stderr) != 0 || string(node.Stdout) != "17\n" {
		test.Fatalf("Node truth: %+v", node)
	}
	loaded, err := load.Load([]string{port})
	if err != nil {
		test.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	isGap := func(err error) bool {
		var diagnostic *lower.NotYet
		return errors.As(err, &diagnostic) &&
			diagnostic.Where == port+":6:28" &&
			diagnostic.What == "a class method through a view that erases its prototype origin"
	}
	if !isGap(err) {
		test.Fatalf("recorded lowering gap changed: %v", err)
	}
	test.Logf("Node prints 17; lowering refuses before native emission: %s", err)
	// A safe explicit callback is also the mutant proving this gap check can reject a normal run.
	fixed, err := rewritten(string(data), classInterfaceWorkaroundSource.replacements)
	if err != nil {
		test.Fatal(err)
	}
	workaround := filepath.Join(directory, "workaround.ts")
	if err = os.WriteFile(workaround, []byte(fixed), 0644); err != nil {
		test.Fatal(err)
	}
	loaded, err = load.Load([]string{workaround})
	if err != nil {
		test.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil || isGap(err) {
		test.Fatalf("explicit callback mutant lowering: %v", err)
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
	data, err := os.ReadFile("gaps/optionalBooleanFunction.ts.txt")
	if err != nil {
		test.Fatal(err)
	}
	path := filepath.Join(test.TempDir(), "gap.a")
	if err = os.WriteFile(path, data, 0644); err != nil {
		test.Fatal(err)
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
