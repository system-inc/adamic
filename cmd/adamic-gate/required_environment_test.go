package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/boundedrun"
)

func TestRequiredEnvironmentConstAudit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(name, source string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module gateprobe\n\ngo 1.27\n")
	write("constants.go", "package gateprobe\nconst prefix = \"ADAMIC_GATE_\"\nconst gateName = prefix + \"COHERE\"\n")
	for _, row := range []struct {
		source string
		valid  bool
	}{
		{`if os.Getenv(gateName) != "1" { t.Skip("required") }`, true},
		{`const local = "ignored"; _ = local; if os.Getenv(gateName) != "1" { t.Skipf("missing %s", gateName) }`, true},
		{`const gateName = "ADAMIC_GATE_COHERE"; if os.Getenv(gateName) != "1" { t.Skip("shadow") }`, false},
		{`gateName := "ADAMIC_GATE_COHERE"; if os.Getenv(gateName) != "1" { t.Skip("dynamic") }`, false},
		{`if os.Getenv(computeName()) != "1" { t.Skip("dynamic") }`, false},
	} {
		write("probe_test.go", "package gateprobe\nimport (\"os\";\"testing\")\nfunc TestProbe(t *testing.T) {"+row.source+"}\n")
		gates, err := environmentGates(root, []string{"ADAMIC_GATE_COHERE"})
		if row.valid {
			if err != nil || strings.Join(gates["gateprobe::TestProbe"], ",") != "ADAMIC_GATE_COHERE" {
				t.Fatalf("const gate: %v %v", gates, err)
			}
		} else if err == nil {
			t.Fatal("dynamic gate accepted", row.source)
		}
	}
}

// Not parallel: real shard/merge entry points use the working directory and environment.
func TestMergeRefusesConstRequiredGateSkip(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("GOWORK", "off")
	t.Setenv("ADAMIC_TOOLS", "")
	t.Setenv("ADAMIC_MARKDOWNWIDTH_DEPS", "")
	t.Setenv("WASI_SYSROOT", "")
	t.Setenv("ADAMIC_GATE_COHERE", "1")
	temporary := os.Getenv("TMPDIR")
	t.Cleanup(func() { os.Setenv("TMPDIR", temporary) })
	for _, name := range []string{"cmd/adamic-gate", "internal", "constgateprobe"} {
		if err := os.MkdirAll(name, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for name, source := range map[string]string{
		"go.mod":                       "module github.com/system-inc/adamic\n\ngo 1.27\n",
		"constgateprobe/gate.go":       "package constgateprobe\nconst gateVariable = \"ADAMIC_GATE_COHERE\"\n",
		"constgateprobe/probe_test.go": "package constgateprobe\nimport (\"os\";\"testing\")\nfunc TestRequired(t *testing.T) { if os.Getenv(gateVariable) != \"1\" { t.Skip(\"required input absent\") } }\n",
		timingPath:                     "{}\n",
	} {
		if err := os.WriteFile(name, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	declareFixtureCensus(t, root)
	gitFixture(t, root, "init", "-q")
	gitFixture(t, root, "config", "user.name", "Gate fixture")
	gitFixture(t, root, "config", "user.email", "gate@example.invalid")
	gitFixture(t, root, "add", ".")
	gitFixture(t, root, "commit", "-qm", "Const gate fixture")
	p, err := makePlan(2)
	if err != nil {
		t.Fatal(err)
	}
	if p.Environment == nil || p.Environment.Shard != 1 || p.WASI != nil || len(p.Units) != 1 || p.Units[0].Shard != 1 || len(p.Units[0].RequiredEnvironment) != 1 || !strings.Contains(p.Environment.Command, "ADAMIC_GATE_COHERE=1") {
		t.Fatal("required shard not planned", p)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatal(err)
	}
	scratch, err := os.MkdirTemp(cache, "gate-required-environment-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(scratch) })
	evidence := t.TempDir()
	dirs := []string{filepath.Join(evidence, "shard-0"), filepath.Join(evidence, "shard-1")}
	for index, dir := range dirs {
		if err := shard(index, 2, dir, scratch, false); err != nil {
			t.Fatal(err)
		}
	}
	if err := merge(dirs, filepath.Join(evidence, "green")); err != nil {
		t.Fatal("required input control", err)
	}
	os.Setenv("ADAMIC_GATE_COHERE", "")
	if err := shard(1, 2, filepath.Join(evidence, "refused"), scratch, false); err == nil || !strings.Contains(err.Error(), "ADAMIC_GATE_COHERE=1 required") {
		t.Fatal("missing input preflight accepted", err)
	}
	// Produce an actual skipped run with the input absent. Rebuild its complete
	// evidence so merge's required-skip verdict is the only reason it goes red.
	pkgDir := filepath.Join(dirs[1], "packages", shortHash("github.com/system-inc/adamic/constgateprobe"))
	var checkpoint packageEvidence
	if err := loadJSON(filepath.Join(pkgDir, "complete.json"), &checkpoint); err != nil {
		t.Fatal(err)
	}
	cmd, release := boundedrun.Command(boundedrun.Build, "go", checkpoint.Invocations[0].Args...)
	cmd.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1")
	err = runJSONCommand(cmd, filepath.Join(pkgDir, "test.jsonl"))
	release()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint.LogDigest, err = fileDigest(filepath.Join(pkgDir, "test.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	checkpoint.StderrDigest, err = fileDigest(filepath.Join(pkgDir, "test.stderr"))
	if err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(filepath.Join(pkgDir, "complete.json"), checkpoint); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"test.jsonl", "test.stderr"} {
		b, err := os.ReadFile(filepath.Join(pkgDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dirs[1], name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var s summary
	if err := loadJSON(filepath.Join(dirs[1], "summary.json"), &s); err != nil {
		t.Fatal(err)
	}
	s.Results, s.CacheLines, _, err = readLog(filepath.Join(dirs[1], "test.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	s.Pass, s.Fail, s.Skip = totals(canonical(s.Results))
	if s.Skip != 1 {
		t.Fatal("fixture did not skip", s.Results)
	}
	if err := saveJSON(filepath.Join(dirs[1], "summary.json"), s); err != nil {
		t.Fatal(err)
	}
	os.Setenv("ADAMIC_GATE_COHERE", "1")
	red := filepath.Join(evidence, "red")
	if err := merge(dirs, red); err == nil {
		t.Fatal("required-input skip merged green")
	}
	var m merged
	if err := loadJSON(filepath.Join(red, "merged.json"), &m); err != nil {
		t.Fatal(err)
	}
	if m.Green || len(m.Errors) != 1 || !strings.Contains(m.Errors[0], "required-input skip") || !strings.Contains(m.Errors[0], "TestRequired") {
		t.Fatal("merge did not isolate and name required skip", m.Errors)
	}
	t.Log(m.Errors[0])
}

func TestCurrentCohereGateIsRequired(t *testing.T) {
	t.Parallel()
	variables, err := requiredGateVariables("../..")
	if err != nil {
		t.Fatal(err)
	}
	gates, err := environmentGates("../..", variables)
	if err != nil {
		t.Fatal(err)
	}
	key := "github.com/system-inc/adamic/cloud::TestRepositoryPassesCohereBaseline"
	if strings.Join(gates[key], ",") != "ADAMIC_GATE_COHERE" {
		t.Fatal("cohere baseline gate not required", gates[key])
	}
	for _, variables := range gates {
		for _, name := range variables {
			if name == "ADAMIC_WASI_RUNTIME" {
				t.Fatal("optional runtime override mistaken for a skip gate")
			}
		}
	}
}
