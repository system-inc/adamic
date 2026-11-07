package json

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestDocumentedStageZeroGaps(test *testing.T) {
	test.Parallel()
	for _, gap := range []struct {
		file, message, output string
		args                  []string
	}{
		{"repeatInTry.ts", "a try around repeat", "xxx\n", []string{"a", "b", "c"}},
	} {
		test.Run(gap.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join("gaps", gap.file))
			if err != nil {
				test.Fatal(err)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				test.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), loaded)
			if err == nil || !strings.Contains(err.Error(), gap.message) {
				test.Fatalf("gap changed: %v; update GAPS.md", err)
			}
			result := onNode(test, path, gap.args...)
			if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != gap.output {
				test.Fatalf("Node: %+v", result)
			}
			test.Logf("%s; Node %q", err, gap.output)
		})
	}
}

// emptyFallback.ts lowers on compiler/area-stack (views slice 1): the untyped [] fallback is no
// longer an array of never. It is held to Node on native ASan/UBSan, the JavaScript backend and
// the leak check. The port's checked panic fallback still stands; retiring it is cohere's change.
func TestClosedEmptyFallbackGap(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join("gaps", "emptyFallback.ts"))
	if err != nil {
		test.Fatal(err)
	}
	program := lowered(test, path)
	binary := filepath.Join(gapPortProduct(test, "emptyFallback.ts", native.Options{Sanitize: true}), "port")
	nativeRun := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(test, path)}, {"JavaScript backend", onJavaScriptBackend(test, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "0\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

func TestClosedMultiplePushGap(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join("gaps", "multiplePush.ts"))
	if err != nil {
		test.Fatal(err)
	}
	program := lowered(test, path)
	binary := filepath.Join(gapPortProduct(test, "multiplePush.ts", native.Options{Sanitize: true}), "port")
	nativeRun := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(test, path)}, {"JavaScript backend", onJavaScriptBackend(test, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "a,b\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}
