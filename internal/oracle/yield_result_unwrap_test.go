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
	// own_symbol stops at container_fit.go's depth bound, a generic pair past it that can reach a Map, and
	// says so: no value there is going where a map is held (#rt5evf5). hazard stops at the real check.
	const containerFit = "stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet"
	const depthGate = "stage 0 can't lower a generic comparison deeper than 4 positions that can reach a Map yet"
	for _, probe := range []struct{ name, node, where, stop string }{
		{"hazard", "true\n", ":22:16", containerFit},
		{"own_symbol", "1\n", ":16:16", depthGate},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/yield_result_unwrap_refused", probe.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if diff := disagreement(run{stdout: []byte(probe.node)}, truth); diff != "" {
				t.Fatal(diff)
			}
			_, err = lowered(t, path)
			var gap *lower.NotYet
			want := path + probe.where + ": " + probe.stop
			if !errors.As(err, &gap) || err.Error() != want {
				t.Fatalf("want %s, got %v", want, err)
			}
		})
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
