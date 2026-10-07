package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var requireFixtures = []string{"internal/oracle/testdata/require_fs.a"}

// Node's standard source runner is ESM, where require is not a global. These
// fixtures contain no imports: execute their stripped source as CommonJS in a
// Node context. No Adamic output participates in this observation.
func requireOnNode(t *testing.T, how inputRun, path string) run {
	t.Helper()
	return executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", "-e", `const fs = require('node:fs');
const { stripTypeScriptTypes } = require('node:module');
const vm = require('node:vm');
const source = stripTypeScriptTypes(fs.readFileSync(process.argv[1], 'utf8'));
vm.runInNewContext(source, { require, console }, { filename: process.argv[1] });`, path)
}

func TestRequireFSAgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, requireFixtures[0]))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	how := inputRun{directory: shared}
	truth := requireOnNode(t, how, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "true true\ntrue\n" {
		t.Fatalf("Node source: %+v", truth)
	}
	backend := inputBackend(t, how, program, shared)
	got, binary := inputNatively(t, how, program, shared)
	for name, observation := range map[string]run{"native": got, "javascript": backend} {
		if difference := disagreement(truth, observation); difference != "" {
			t.Errorf("%s: %s", name, difference)
		}
	}
	if leaked := inputLeaks(t, how, program, binary); leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
	t.Log("Node source: true true / true; both backends, ASan/UBSan and leaks agree")
}

func TestRequirePathHostGapHeldToNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"require_path.a", "require_node_path.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/require_gaps", name))
			if err != nil {
				t.Fatal(err)
			}
			truth := requireOnNode(t, inputRun{directory: sharedDirectory(t)}, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "a/b\na\n" {
				t.Fatalf("Node source: %+v", truth)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want path host NotYet, got %v", err)
			}
			t.Logf("Node source: a/b / a; host gap: %v", err)
		})
	}
}
