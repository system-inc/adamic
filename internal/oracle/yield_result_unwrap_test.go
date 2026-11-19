package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Keep this TypeScript generator fixture on the standard runtime-count ledger.
func init() {
	const path = "internal/oracle/testdata/generators/yield_result_unwrap.ts"
	additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, false, nil, false, false) }})
}

// The named and in-place generators yield the same items. Compare the whole
// source with Node on each backend, including WASI when explicitly enabled.
func TestYieldResultUnwrapAgrees(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generators/yield_result_unwrap.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if diff := disagreement(run{stdout: []byte("x\nx\n")}, truth); diff != "" {
		t.Fatal(diff)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"native", "sanitized", "javascript", "wasi"} {
		t.Run(backend, func(t *testing.T) {
			var got run
			switch backend {
			case "native":
				got = released(t, program)
			case "sanitized":
				got, _ = natively(t, program)
			case "javascript":
				got = onJavaScriptBackend(t, program)
			case "wasi":
				if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
					t.Skip("set ADAMIC_ORACLE_WASI=1")
				}
				got = onWASI(t, native.C(program))
			}
			if diff := disagreement(truth, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestYieldResultUnwrapRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/yield_result_unwrap_refused/hazard.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if diff := disagreement(run{stdout: []byte("true\n")}, truth); diff != "" {
		t.Fatal(diff)
	}
	_, err = lowered(t, path)
	var gap *lower.NotYet
	want := path + ":22:16: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet"
	if !errors.As(err, &gap) || err.Error() != want {
		t.Fatalf("want %s, got %v", want, err)
	}
}

// own_symbol declares its own IteratorYieldResult, which gets the ordinary walk rather than the bundled
// lib's unwrap. Its pair reaches a Map past container_fit.go's depth and grows no generic on its path, so
// since #rt5evf5 head 2 the walk follows it to its end, a real Map where a ReadonlyMap is held, and it lowers
// (it stopped at the depth bound before). Mutant: the gate stopping every generic pair past the depth that
// can reach a map stops it again.
func TestYieldResultUnwrapOwnSymbolAgrees(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/yield_result_unwrap_refused/own_symbol.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if diff := disagreement(run{stdout: []byte("1\n")}, truth); diff != "" {
		t.Fatal(diff)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := natively(t, program)
	for backend, got := range map[string]run{"sanitized": sanitized, "native": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(truth, got); diff != "" {
			t.Errorf("%s: %s", backend, diff)
		}
	}
}

// S holds IterableIterator<S>. It lowers today; keep its cycle handling bounded
// and compare its result rather than merely accepting any terminating outcome.
func TestYieldResultUnwrapRecursive(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/yield_result_unwrap_refused/recursive.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if diff := disagreement(run{stdout: []byte("1\n")}, truth); diff != "" {
		t.Fatal(diff)
	}
	started := time.Now()
	program, err := lowered(t, path)
	t.Logf("recursive compile: %s", time.Since(started))
	if err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(truth, onJavaScriptBackend(t, program)); diff != "" {
		t.Fatal(diff)
	}
}
