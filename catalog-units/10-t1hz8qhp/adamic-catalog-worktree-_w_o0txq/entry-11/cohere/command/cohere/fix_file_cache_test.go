package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/typescript"
	"github.com/system-inc/cohere/internal/types/program"
)

/*
 * The fix path's re-lint shares one file cache across its rules, as the lint walk does (#xn1k1gz).
 *
 * proposalsForText built every rule's context without a FileCache, and rule.Cached reads a nil cache as
 * "recompute every time". A derivation meant to be paid once per file was paid on every ask: no-invalid-this
 * asks for the file's comments at each function, so on a 2.8 MB esbuild bundle in angular/angular (formatted,
 * then re-linted) the whole-file comment scan ran once per function, and `--no-fix` took 26 minutes and 19 GB
 * on one core while its two halves each finished in seconds.
 */

// TestTheFixPathsReLintComputesEachDerivationOncePerText: every rule re-linting a text shares one cache, so a
// derivation asked for at every node by two rules is computed once.
func TestTheFixPathsReLintComputesEachDerivationOncePerText(t *testing.T) {
	t.Parallel()
	computed := 0
	asking := func(name string) rule.Rule {
		return rule.Rule{Name: name, Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{ast.KindIdentifier: func(node *ast.Node) {
				rule.Cached(ctx.FileCache, "fix-file-cache-probe", func() int {
					computed++
					return computed
				})
			}}
		}}
	}
	text := "const first = second;\nconst third = first + second;\n"
	if _, err := proposalsForText("/Probe.ts", text, &program.Graph{}, []rule.Rule{asking("probe-a"), asking("probe-b")}); err != nil {
		t.Fatalf("proposalsForText: %v", err)
	}
	if computed != 1 {
		t.Errorf("the derivation was computed %d times for one text asked at every identifier by two rules, want once", computed)
	}
}

// bundle is an esbuild-shaped file of count small modules, each a commented function using `this`, the shape
// that made no-invalid-this ask for the file's comments once per function.
func bundle(count int) string {
	var text strings.Builder
	text.WriteString("\"use strict\";\n(() => {\n")
	for index := range count {
		fmt.Fprintf(&text, "  // module %d\n  var module%d = function() {\n    /* the module's body */\n    return this.value%d;\n  };\n", index, index, index)
	}
	text.WriteString("})();\n")
	return text.String()
}

// TestTheFixPathsReLintGrowsLinearlyOnABundle: re-linting a bundle four times the size allocates about four
// times as much. Each function rescanning the whole file's comments made it about sixteen.
//
// Not parallel: testing.AllocsPerRun reads the process's allocation count, which a parallel test would add to,
// and it refuses to run inside one.
func TestTheFixPathsReLintGrowsLinearlyOnABundle(t *testing.T) {
	rules := []rule.Rule{typescript.NoInvalidThis}
	allocations := func(count int) float64 {
		text := bundle(count)
		return testing.AllocsPerRun(3, func() {
			if _, err := proposalsForText("/main.js", text, &program.Graph{}, rules); err != nil {
				t.Fatalf("proposalsForText: %v", err)
			}
		})
	}
	small, large := allocations(100), allocations(400)
	if ratio := large / small; ratio > 6 {
		t.Errorf("four times the bundle allocated %.1f times as much (%.0f against %.0f objects), want about four: the re-lint is quadratic in the file",
			ratio, large, small)
	}
}

// lockedBuffer is a buffer a timer's goroutine and the test can both use.
type lockedBuffer struct {
	mutex sync.Mutex
	text  strings.Builder
}

func (b *lockedBuffer) Write(bytes []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.text.Write(bytes)
}

func (b *lockedBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.text.String()
}

// TestAFileStillInTheFixPhaseIsNamedWhileItRuns: a file that outlasts the notice is named, with where to
// report it, while it is still running, and a file that finishes first says nothing (#xn1k1gz). The runaway
// that found this sat 26 minutes on one core with nothing on screen.
func TestAFileStillInTheFixPhaseIsNamedWhileItRuns(t *testing.T) {
	t.Parallel()
	for name, testCase := range map[string]struct {
		after, takes time.Duration
		named        bool
	}{
		"a file that outlasts the notice is named": {after: time.Millisecond, takes: 200 * time.Millisecond, named: true},
		"a file that finishes first says nothing":  {after: time.Minute, takes: 0, named: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			out := &lockedBuffer{}
			seenWhileRunning := false
			result, err := watchedFile(out, testCase.after, "/bundle/main.js", func() (edit.FileResult, error) {
				time.Sleep(testCase.takes)
				seenWhileRunning = strings.Contains(out.String(), "/bundle/main.js")
				return edit.FileResult{FileName: "/bundle/main.js"}, nil
			})
			if err != nil || result.FileName != "/bundle/main.js" {
				t.Fatalf("the file's own result did not come back: %+v, %v", result, err)
			}
			if seenWhileRunning != testCase.named {
				t.Errorf("named while running %t, want %t: %q", seenWhileRunning, testCase.named, out.String())
			}
			if testCase.named && !strings.Contains(out.String(), reportBugsAt) {
				t.Errorf("the notice does not say where to report it: %q", out.String())
			}
		})
	}
}
