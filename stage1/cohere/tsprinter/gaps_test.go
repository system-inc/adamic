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

func TestCompilerGaps(t *testing.T) {
	t.Parallel()
	for _, item := range []struct{ name, reason, output string }{
		{"defaultSort", "Adamic 0.1 refuses sort without a comparator", "im\n"},
	} {
		// Not parallel within the parent: a parallel subtest queues behind every waiting test in the unit, and that wait
		// counts as the parent's own time (259 s at share 8, landable-16c, for one Node run, a load and a lower).
		t.Run(item.name, func(t *testing.T) {
			path, err := filepath.Abs("gaps/" + item.name + ".ts")
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.ExitCode != 0 || len(node.Stderr) != 0 || string(node.Stdout) != item.output {
				t.Fatalf("Node: %#v", node)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), item.reason) {
				t.Fatalf("expected recorded gap %q, got %v", item.reason, err)
			}
			t.Logf("Node %q; stage 0 %s", item.output, err)
		})
	}
}

// prefixUpdateValue.ts lowers on compiler/area-stack (views slice 1, Oct 8): a numeric
// PrefixUnaryExpression whose value is read. Held to Node on native, the JavaScript backend and
// the leak check. The port's separate decrement statement still stands; retiring it is cohere's.
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedPrefixUpdateValueGap(t *testing.T) {
	path, err := filepath.Abs("gaps/prefixUpdateValue.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	binary := filepath.Join(sourceSanitized(t, prefixUpdateValueSource), "port")
	native := sanitizedRun(t, binary)
	for _, result := range []oracles.Run{onNode(t, path), native, onJavaScriptBackend(t, program)} {
		if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != "2\n" {
			t.Fatalf("prefix update value: %+v", result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestNumberConstructor(t *testing.T) {
	path, err := filepath.Abs("testdata/numberConstructor.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	binary := filepath.Join(sourceSanitized(t, numberConstructorSource), "port")
	native := sanitizedRun(t, binary)
	for _, result := range []oracles.Run{onNode(t, path), native, onJavaScriptBackend(t, program)} {
		if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != "17\n" {
			t.Fatalf("Number constructor: %+v", result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
