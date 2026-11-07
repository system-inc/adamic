package tsprinter

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestCompilerGaps(test *testing.T) {
	test.Parallel()
	for _, item := range []struct{ name, reason, output string }{
		{"defaultSort", "Adamic 0.1 refuses sort without a comparator", "im\n"},
	} {
		// Not parallel within the parent: a parallel subtest queues behind every waiting test in the unit, and that
		// wait counts as the parent's own time (259 s at share 8, landable-16c, for one Node run, a load and a lower).
		test.Run(item.name, func(test *testing.T) {
			path, err := filepath.Abs("gaps/" + item.name + ".ts")
			if err != nil {
				test.Fatal(err)
			}
			node := onNode(test, path)
			if node.ExitCode != 0 || len(node.Stderr) != 0 || string(node.Stdout) != item.output {
				test.Fatalf("Node: %#v", node)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				test.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), item.reason) {
				test.Fatalf("expected recorded gap %q, got %v", item.reason, err)
			}
			test.Logf("Node %q; stage 0 %s", item.output, err)
		})
	}
}

// prefixUpdateValue.ts lowers on compiler/area-stack (views slice 1, Oct 8): a numeric
// PrefixUnaryExpression whose value is read. Held to Node on native, the JavaScript backend and
// the leak check. The port's separate decrement statement still stands; retiring it is cohere's.
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedPrefixUpdateValueGap(test *testing.T) {
	path, err := filepath.Abs("gaps/prefixUpdateValue.ts")
	if err != nil {
		test.Fatal(err)
	}
	program := lowered(test, path)
	binary := filepath.Join(sourceSanitized(test, prefixUpdateValueSource), "port")
	native := sanitizedRun(test, binary)
	for _, result := range []oracles.Run{onNode(test, path), native, onJavaScriptBackend(test, program)} {
		if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != "2\n" {
			test.Fatalf("prefix update value: %+v", result)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestNumberConstructor(test *testing.T) {
	path, err := filepath.Abs("testdata/numberConstructor.ts")
	if err != nil {
		test.Fatal(err)
	}
	program := lowered(test, path)
	binary := filepath.Join(sourceSanitized(test, numberConstructorSource), "port")
	native := sanitizedRun(test, binary)
	for _, result := range []oracles.Run{onNode(test, path), native, onJavaScriptBackend(test, program)} {
		if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != "17\n" {
			test.Fatalf("Number constructor: %+v", result)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}
