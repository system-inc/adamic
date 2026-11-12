package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

const inspectorDoomedFixture = "internal/oracle/testdata/node_inspector_doomed.a"
const inspectorDoomedConstructorTrap = `node:inspector.Session was reached through inspector, whose require("inspector") throws on every host Adamic targets`
const inspectorDoomedPostTrap = `node:inspector.Session.post was reached through session, whose require("inspector") throws on every host Adamic targets`

func TestInspectorDoomedSourceWitnessMatchesNode(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/node_inspector_doomed.node-witness.json"))
	if err != nil {
		t.Fatal(err)
	}
	var witness struct {
		Fixture string `json:"fixture"`
		Stdout  string `json:"stdout"`
		Stderr  string `json:"stderr"`
		Exit    int    `json:"exit_code"`
	}
	if err := json.Unmarshal(data, &witness); err != nil {
		t.Fatal(err)
	}
	if witness.Fixture != inspectorDoomedFixture {
		t.Fatal("witness fixture mismatch")
	}
	got := onSliceNode(t, filepath.Join(repository, inspectorDoomedFixture))
	want := run{stdout: []byte(witness.Stdout), stderr: []byte(witness.Stderr), exitCode: witness.Exit}
	if why := disagreement(want, got); why != "" {
		t.Fatalf("committed Node witness: %s: %#v", why, got)
	}
}
func TestInspectorDoomedRequiresThrowBeforeTrap(t *testing.T) {
	t.Parallel()
	p, err := lowered(t, filepath.Join(repository, inspectorDoomedFixture))
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("ERR_INSPECTOR_NOT_AVAILABLE\n")}
	sanitized, _ := nativelyUncached(t, p)
	for name, got := range map[string]run{"native": sanitized, "release": releasedUncached(t, p), "javascript": onSliceJavaScript(t, p)} {
		if why := disagreement(want, got); why != "" {
			t.Fatalf("inspector require must throw before a doomed use: %s: %s; exit=%d stdout=%q stderr=%q", name, why, got.exitCode, got.stdout, got.stderr)
		}
	}
}
func TestInspectorDoomedBrokenAxiomStopsLoudly(t *testing.T) {
	t.Parallel()
	p, err := lowered(t, filepath.Join(repository, inspectorDoomedFixture))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(p.Strings, inspectorDoomedConstructorTrap) {
		t.Fatal("constructor trap text changed")
	}
	if os.Getenv("ADAMIC_INSPECTOR_AXIOM_BROKEN") != "1" {
		return
	}
	want := run{stderr: []byte("adamic: panic: " + inspectorDoomedConstructorTrap + "\n"), exitCode: 70}
	actual, _ := nativelyUncached(t, p)
	for name, got := range map[string]run{"native": actual, "release": releasedUncached(t, p), "javascript": onSliceJavaScript(t, p), "wasi": onWASI(t, native.C(p))} {
		if why := disagreement(want, got); why != "" {
			t.Fatalf("broken axiom %s: %s: %#v", name, why, got)
		}
	}
	t.Log("broken inspector axiom: trap exit=70 stdout=empty")
}
func TestInspectorDoomedReadBehavior(t *testing.T) {
	t.Parallel()
	dropped := os.Getenv("ADAMIC_INSPECTOR_READ_DROPPED") == "1"
	for _, name := range []string{"d1", "d1-control", "hoisted", "hoisted-uncaught"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/lower/testdata/inspector_doomed", name+".ts")
			p, err := lowered(t, path)
			if name == "d1-control" || (!dropped && name == "d1") {
				truth := onSliceNode(t, path)
				if why := disagreement(run{stdout: []byte("require threw\ncaught ReferenceError\n")}, truth); why != "" {
					t.Fatalf("escaped source Node: %s: %#v", why, truth)
				}
				var stop *lower.NotYet
				if !errors.As(err, &stop) || stop.What != "reading session" {
					t.Fatalf("d1 must preserve its ordinary named read refusal: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte("early\n")}
			if dropped {
				want = run{stderr: []byte("adamic: panic: " + inspectorDoomedPostTrap + "\n"), exitCode: 70}
				if name == "d1" {
					want.stdout = []byte("require threw\n")
				}
			}
			if !dropped && name == "hoisted-uncaught" {
				want = run{stderr: []byte("adamic: panic: ReferenceError: Cannot access 'session' before initialization\n"), exitCode: 70}
			}
			if !dropped {
				truth := onSliceNode(t, path)
				if why := disagreement(want, truth); why != "" {
					t.Fatalf("Node %s: %s: %#v", name, why, truth)
				}
			}
			actual, _ := nativelyUncached(t, p)
			outcomes := map[string]run{"native": actual, "release": releasedUncached(t, p), "javascript": onSliceJavaScript(t, p)}
			if os.Getenv("ADAMIC_TEST_WASI") == "1" && os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				outcomes["wasi"] = onWASI(t, native.C(p))
			}
			for backend, got := range outcomes {
				if why := disagreement(want, got); why != "" {
					t.Fatalf("%s %s: %s: %#v", name, backend, why, got)
				}
			}
			if dropped {
				t.Log(name + " trap exit=70")
			}
		})
	}
}
func inspectorDoomedOracleMutant(t *testing.T, mode string) string {
	t.Helper()
	var changes []buildcache.Mutation
	switch mode {
	case "Axiom":
		changes = []buildcache.Mutation{{File: "internal/lower/library_node_commonjs.go", Before: `effects.Body = append(effects.Body, ir.Evaluate{Value: ir.NodeFSFile{Operation: "host_inspector_unavailable", Of: ir.Number}})`, After: "// mutant: the no-inspector require returns its placeholder"}}
	case "ReadDropOracle":
		changes = []buildcache.Mutation{{File: "internal/lower/inspector_doomed.go", Before: "read, err := l.expression(root)", After: "_, err := ir.Expression(nil), error(nil)"}, {File: "internal/lower/inspector_doomed.go", Before: "ir.Evaluate{Value: read}, ", After: ""}}
	default:
		t.Fatal(mode)
	}
	return buildcache.GoTestMutated(t, "", "inspector-doomed-"+mode+".test", "./internal/oracle", nil, changes)
}
func TestProduct_InspectorDoomedAxiomMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedOracleMutant(t, "Axiom")
}
func TestProduct_InspectorDoomedReadDropOracleMutant(t *testing.T) {
	t.Parallel()
	inspectorDoomedOracleMutant(t, "ReadDropOracle")
}
func inspectorDoomedOracleChild(t *testing.T, binary, pin string, environment []string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-test.run", "^"+pin+"$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), environment...)
	stdout, stderr, err := childguard.Capture(ctx, cmd)
	if len(stderr) != 0 {
		t.Logf("child stderr: %s", stderr)
	}
	return string(stdout), err
}
func TestInspectorDoomedAxiomMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := inspectorDoomedOracleMutant(t, "Axiom")
	out, err := inspectorDoomedOracleChild(t, binary, "TestInspectorDoomedRequiresThrowBeforeTrap", nil)
	var exited *exec.ExitError
	const line = "inspector require must throw before a doomed use"
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(out, line) {
		t.Fatalf("broken axiom not caught: %v %s", err, out)
	}
	t.Log("TestInspectorDoomedRequiresThrowBeforeTrap goes red: " + line)
	out, err = inspectorDoomedOracleChild(t, binary, "TestInspectorDoomedBrokenAxiomStopsLoudly", []string{"ADAMIC_INSPECTOR_AXIOM_BROKEN=1", "ADAMIC_TEST_WASI=1", "ADAMIC_ORACLE_WASI=1"})
	if err != nil || !strings.Contains(out, "trap exit=70 stdout=empty") {
		t.Fatalf("broken axiom must reach exact trap: %v %s", err, out)
	}
	t.Log(out)
}
func TestInspectorDoomedReadDropOracleMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := inspectorDoomedOracleMutant(t, "ReadDropOracle")
	out, err := inspectorDoomedOracleChild(t, binary, "TestInspectorDoomedReadBehavior", []string{"ADAMIC_INSPECTOR_READ_DROPPED=1", "ADAMIC_TEST_WASI=1", "ADAMIC_ORACLE_WASI=1"})
	if err != nil || !strings.Contains(out, "d1 trap exit=70") || !strings.Contains(out, "hoisted trap exit=70") {
		t.Fatalf("dropping read must reach both exact traps: %v %s", err, out)
	}
	t.Log(out)
}
