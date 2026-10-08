package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/predicate_callback_contract.a", "internal/oracle/testdata/predicate_helper_return.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

func TestPredicateCheckedNarrowing(t *testing.T) {
	for _, branch := range []string{"true", "false", "callback"} {
		t.Run(branch, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/predicate_hatch_"+branch+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refusal *lower.Refused
			if !errors.As(err, &refusal) {
				t.Fatalf(".a must refuse its unproven predicate: %v", err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			ts := filepath.Join(t.TempDir(), "hatch.ts")
			if err := os.WriteFile(ts, source, 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, ts)
			if err != nil {
				t.Fatal(err)
			}
			oracle := onNode(t, path)
			if oracle.exitCode != 0 {
				t.Fatalf("Node must run on: %d %s", oracle.exitCode, oracle.stderr)
			}
			js := onJavaScriptBackend(t, program)
			native, _ := natively(t, program)
			release := released(t, program)
			for name, observed := range map[string]run{"javascript": js, "sanitized": native, "release": release} {
				if observed.exitCode != 70 || !strings.Contains(string(observed.stderr), "predicate narrowing failed") {
					t.Errorf("%s must fire the narrowing check: exit %d stderr %q", name, observed.exitCode, observed.stderr)
				}
			}
			if d := disagreement(js, native); d != "" {
				t.Error(d)
			}
			if d := disagreement(native, release); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestPredicateOpenContractsStayRefused(t *testing.T) {
	for name, source := range map[string]string{
		"bodyless":         `type A={readonly kind:'a'};type B={readonly kind:'b'};type Guard=(n:A|B)=>n is A;`,
		"open payload":     `type N={readonly kind:number};type A=N & {readonly label:string};function guard(n:N):n is A{return true;}`,
		"extra refinement": `type A={readonly kind:'a'};type B={readonly kind:'b'};type Refined=A & {readonly label:string};function guard(n:A|B):n is Refined{return true;}`,
	} {
		t.Run(name, func(t *testing.T) {
			extension := ".a"
			if name == "bodyless" {
				extension = ".ts"
			}
			path := filepath.Join(t.TempDir(), "contract"+extension)
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var refusal *lower.Refused
			if !errors.As(err, &refusal) {
				t.Fatalf("unproven .a contract must stay refused: %v", err)
			}
		})
	}
}
