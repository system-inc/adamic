package oracle

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// tsc writes CompilerOptions through its index signature (commandLineParser's options[opt.name] = value), so any value
// of the slot's union can sit under a declared key. records_union_options.ts reads declared options as Node does; the
// lying programs put a number and a null under outFile?: string, which Node prints and Adamic stops at, exit 70.
var recordUnionLying = []struct{ name, node string }{{"lying", "3\n"}, {"lying_null", "null\n"}}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/records_union_options.ts", true, false})
	for _, lying := range recordUnionLying {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_union_" + lying.name + ".ts", true, true})
	}
}

const recordDeclaredMemberPanic = "adamic: panic: dictionary member does not have its declared type\n"

// recordDeclaredReadStop says how a lying program failed to stop at the declared read's checked unbox on each backend
// that didn't, sanitized, release or JavaScript, or "" when every one of them stopped there.
func recordDeclaredReadStop(t *testing.T, program *ir.Program) string {
	t.Helper()
	sanitized, _ := natively(t, program)
	want := run{exitCode: 70, stderr: []byte(recordDeclaredMemberPanic)}
	failures := []string{}
	for _, backend := range []struct {
		name string
		got  run
	}{{"sanitized", sanitized}, {"release", released(t, program)}, {"JavaScript", onJavaScriptBackend(t, program)}} {
		if difference := disagreement(want, backend.got); difference != "" {
			failures = append(failures, fmt.Sprintf("%s: %s: exit %d, stdout %q, stderr %q", backend.name, difference, backend.got.exitCode, backend.got.stdout, backend.got.stderr))
		}
	}
	return strings.Join(failures, "; ")
}

func TestRecordUnionDeclaredReadStops(t *testing.T) {
	t.Parallel()
	for _, lying := range recordUnionLying {
		t.Run(lying.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_union_"+lying.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != lying.node || len(truth.stderr) != 0 {
				t.Fatalf("Node no longer reads the lying value: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if failure := recordDeclaredReadStop(t, program); failure != "" {
				t.Fatal(failure)
			}
		})
	}
}

// Mutant: the declared read with its checked unbox dropped, so the slot is read at the declared type unchecked. The
// stop's pin must fail on it.
func TestRecordUnionDeclaredReadMutant(t *testing.T) {
	t.Parallel()
	for _, lying := range recordUnionLying {
		t.Run(lying.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_union_"+lying.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for i, f := range program.Functions {
				if f.Name != "record_declared_member" {
					continue
				}
				body := []ir.Statement{}
				for _, s := range f.Body {
					if check, ok := s.(ir.If); ok && len(check.Then) == 1 {
						if _, ok := check.Then[0].(ir.Panic); ok {
							changed = true
							continue
						}
					}
					body = append(body, s)
				}
				program.Functions[i].Body = body
			}
			if !changed {
				t.Fatal("mutant removed no checked unbox")
			}
			failure := recordDeclaredReadStop(t, program)
			if failure == "" {
				t.Fatal("the unchecked declared read still stopped on every backend")
			}
			t.Logf("unchecked declared read caught by the stop's pin: %s", failure)
		})
	}
}
