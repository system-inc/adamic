package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// The unused identity casts register recursive array view contracts. The first
// find used to lower its property receiver twice with different contracts.
func TestOptionalReceiverViewContractAgrees(t *testing.T) {
	t.Parallel()
	for _, backend := range []string{"native", "sanitized", "javascript", "wasi"} {
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			if backend == "wasi" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
				t.Skip("set ADAMIC_ORACLE_WASI=1 to run the WASI oracle")
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_receiver_view_contract.a"))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			var got run
			switch backend {
			case "native":
				got = releasedUncached(t, program)
			case "sanitized":
				var binary string
				got, binary = nativelyUncached(t, program)
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			case "javascript":
				got = onJavaScriptBackend(t, program)
			case "wasi":
				got = onWASI(t, native.C(program))
			}
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("%s: %s; Node stdout=%q stderr=%q exit=%d; backend stdout=%q stderr=%q exit=%d", backend, diff, want.stdout, want.stderr, want.exitCode, got.stdout, got.stderr, got.exitCode)
			}
		})
	}
}

func TestOptionalReceiverArgumentOverlapRefuses(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_receiver_once_refused/argument_overlap.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.NotYet
	if !errors.As(err, &refused) || refused.What != "an optional intrinsic without exactly one explicit receiver operand" {
		t.Fatalf("want the argument/receiver overlap refusal, got %v", err)
	}
}
