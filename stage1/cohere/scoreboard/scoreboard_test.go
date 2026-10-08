package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMutantIsCaught(t *testing.T) {
	t.Parallel()
	base := Execution{Output: "case 0\np.ts:1:1\n  no-debugger  Unexpected 'debugger' statement.\n\nrange 0 9 unexpected fix\t\t\t0 9\nfixed\t\n"}
	mutant := base
	mutant.Output = strings.Replace(base.Output, "Unexpected 'debugger' statement.", "planted divergence", 1)
	if classify(base, mutant) != "diverge" {
		t.Fatal("wrong-message mutant survived")
	}
	cells := []Cell{{File: "p.ts", Rule: "no-debugger", Family: "core", Status: classify(base, mutant), Findings: 1}}
	r := Report{Cells: cells}
	summarize(&r)
	if r.PerRule["no-debugger"].Diverge != 1 || r.PerFamily["core"].Diverge != 1 || r.PerFile["p.ts"].Diverge != 1 {
		t.Fatal("mutant lost in an aggregation")
	}
}
func TestFailuresCannotAgree(t *testing.T) {
	t.Parallel()
	for _, e := range []Execution{{Error: "exit 70"}, {Output: "skipped typed no program\n"}, {Output: "refused format/typescript function-types\n"}} {
		if classify(e, e) != "blocked" {
			t.Fatalf("false agreement: %+v", e)
		}
	}
}
func TestMinimize(t *testing.T) {
	t.Parallel()
	got := minimize("const prefix=0;debugger; const suffix=1;", func(s string) bool { return strings.Contains(s, "debugger;") })
	if got != "debugger;" {
		t.Fatalf("got %q", got)
	}
}
func TestManifestRejectsSubstitution(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "p.ts")
	if err := os.WriteFile(p, []byte("debugger;"), 0600); err != nil {
		t.Fatal(err)
	}
	if validate(Manifest{Dependencies: DependencyInputs{Go: "uninstalled", Node: "uninstalled"}, Name: "quiet", Files: []string{p}}, true, dir) == nil {
		t.Fatal("accepted one file as quiet hundred")
	}
	if validate(Manifest{Dependencies: DependencyInputs{Go: "uninstalled", Node: "uninstalled"}, Name: "fixtures", Files: []string{p, p}}, false, dir) == nil {
		t.Fatal("accepted duplicate")
	}
	if err := validate(Manifest{Dependencies: DependencyInputs{Go: "uninstalled", Node: "uninstalled"}, Name: "fixtures", Files: []string{p}}, false, dir); err != nil {
		t.Fatal(err)
	}
}
func TestProcessTimeoutIsBlocked(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := execute(ctx, "", "go", "version")
	if result.Error == "" || classify(result, result) != "blocked" {
		t.Fatal("canceled process became agreement")
	}
}
func TestLiveOracleAndNode(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	d := Driver{Root: root, Scratch: scratch, Limit: 30 * time.Second}
	if err = d.build(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(scratch, "probe.ts")
	if err = os.WriteFile(path, []byte("debugger;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a, b := d.pair(path, "no-debugger", "", "")
	if classify(a, b) != "agree" || !strings.Contains(a.Output, "\nrange ") {
		t.Fatalf("Go/Node live check: %+v\n%+v", a, b)
	}
	for _, fixture := range []struct{ name, path, marker string }{
		{"comment-loss", "comment-loss.ts.txt", "//"},
		{"BOM preservation", "typescript-compiler.ts.txt", "\ufeff"},
	} {
		a, b := d.format(filepath.Join(root, "stage1/cohere/scoreboard/testdata", fixture.path))
		if classify(a, b) != "diverge" || !strings.Contains(a.Output, fixture.marker) || strings.Contains(b.Output, fixture.marker) {
			t.Fatalf("named fixture %s lost its failure: Go=%+v Node=%+v", fixture.name, a, b)
		}
	}
	// Formatting reductions must not switch a valid witness to a parser error.
	invalid := filepath.Join(scratch, "invalid.ts")
	if err = os.WriteFile(invalid, []byte("="), 0600); err != nil {
		t.Fatal(err)
	}
	ctxValid, cancelValid := context.WithTimeout(context.Background(), 30*time.Second)
	validity := execute(ctxValid, root, d.Catalog, "--valid", invalid)
	cancelValid()
	if validity.Error == "" {
		t.Fatal("syntax-invalid formatter reduction accepted")
	}
	ctxValid, cancelValid = context.WithTimeout(context.Background(), 30*time.Second)
	validity = execute(ctxValid, root, d.Catalog, "--valid", path)
	cancelValid()
	if validity.Error != "" {
		t.Fatalf("valid fixture refused: %+v", validity)
	}
	// Plant a clean-running source Node adapter that changes only the oracle-visible message.
	wrapper := filepath.Join(scratch, "mutant.mjs")
	source := `import { spawnSync } from 'node:child_process';
const result = spawnSync(process.execPath, process.argv.slice(2), { encoding: 'utf8' });
process.stdout.write(result.stdout.replace("A debugger statement stops execution", 'planted divergence'));
process.stderr.write(result.stderr); process.exit(result.status);
`
	if err = os.WriteFile(wrapper, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mutant := execute(ctx, root, "node", wrapper, "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(root, "stage1/cohere/lint/main.ts"), "--manifest", filepath.Join(scratch, "row.txt"))
	t.Logf("live witness debugger;: Go and Node agree; clean-running Node message mutant status=%s, findings=%d", classify(a, mutant), strings.Count(a.Output, "\nrange "))
	if classify(a, mutant) != "diverge" {
		t.Fatalf("clean-running Node message mutant survived: %+v", mutant)
	}
}

func TestPartitionPreservesBytes(t *testing.T) {
	t.Parallel()
	text := "case 0\nskipped typed/rule no program\np.ts:1:1\n  no-debugger  message\n\nrange 0 9 id fix\t\t\t0 9\np.ts:2:1\n  eqeqeq  equality\n\nrange 10 12 eq suggestion\t===\tstrict\t10 12\nfixed\tchanged\n"
	first := ruleAnswer(Execution{Output: text}, "no-debugger")
	if first.Output != "p.ts:1:1\n  no-debugger  message\n\nrange 0 9 id fix\t\t\t0 9\n" {
		t.Fatalf("wrong partition: %q", first.Output)
	}
	if classify(ruleAnswer(Execution{Output: text}, "typed/rule"), ruleAnswer(Execution{Output: text}, "typed/rule")) != "blocked" {
		t.Fatal("typed skip lost")
	}
	mutant := strings.Replace(text, "===", "==", 1)
	if classify(ruleAnswer(Execution{Output: text}, "eqeqeq"), ruleAnswer(Execution{Output: mutant}, "eqeqeq")) != "diverge" {
		t.Fatal("suggestion-edit mutant lost")
	}
	if classify(first, ruleAnswer(Execution{Output: mutant}, "no-debugger")) != "agree" {
		t.Fatal("other rule polluted")
	}
}

func TestDependencySnapshotBoundary(t *testing.T) {
	t.Parallel()
	for _, inputs := range []DependencyInputs{{}, {Go: "uninstalled"}, {Go: "installed", Node: "uninstalled"}, {Go: "uninstalled", Node: "installed"}, {Go: "unknown", Node: "unknown"}} {
		if inputs.validate() == nil {
			t.Fatalf("accepted incomparable snapshots: %+v", inputs)
		}
	}
	for _, state := range []string{"installed", "uninstalled"} {
		if err := (DependencyInputs{Go: state, Node: state}).validate(); err != nil {
			t.Fatal(err)
		}
	}
}
