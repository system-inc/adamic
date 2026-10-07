package oracle

import (
	"path/filepath"
	"testing"
)

var requireFixtures = []string{"internal/oracle/testdata/require_fs.a", "internal/oracle/testdata/require_path.a", "internal/oracle/testdata/require_node_path.a"}

// Node's standard source runner is ESM, where require is not a global. These
// fixtures contain no imports: execute their stripped source as CommonJS in a
// Node context. No Adamic output participates in this observation.
func requireOnNode(t *testing.T, how inputRun, path string) run {
	t.Helper()
	return executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", "-e", `const fs = require('node:fs');
const { stripTypeScriptTypes } = require('node:module');
const vm = require('node:vm');
const source = stripTypeScriptTypes(fs.readFileSync(process.argv[1], 'utf8'));
vm.runInNewContext(source, { require, console, Error, performance, process }, { filename: process.argv[1] });`, path)
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
			t.Errorf("%s: %s (Node %q, backend %q)", name, difference, truth.stdout, observation.stdout)
		}
	}
	if leaked := inputLeaks(t, how, program, binary); leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
	t.Log("Node source: true true / true; both backends, ASan/UBSan and leaks agree")
}

func TestRequirePathAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range requireFixtures[1:] {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, name))
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
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "a/b\na\n" {
				t.Fatalf("Node source: %+v", truth)
			}
			backend := inputBackend(t, how, program, shared)
			got, binary := inputNatively(t, how, program, shared)
			for name, observation := range map[string]run{"native": got, "javascript": backend} {
				if difference := disagreement(truth, observation); difference != "" {
					t.Errorf("%s: %s (Node %q, backend %q)", name, difference, truth.stdout, observation.stdout)
				}
			}
			if leaked := inputLeaks(t, how, program, binary); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
		})
	}
}

var requirePerformanceFixtures = []string{"internal/oracle/testdata/require_perf_hooks.a", "internal/oracle/testdata/require_node_perf_hooks.a"}

func TestRequirePerformanceAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range requirePerformanceFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, name))
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
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node source: %+v", truth)
			}
			backend := inputBackend(t, how, program, shared)
			got, binary := inputNatively(t, how, program, shared)
			for name, observation := range map[string]run{"native": got, "javascript": backend} {
				if difference := disagreement(truth, observation); difference != "" {
					t.Errorf("%s: %s (Node %q, backend %q)", name, difference, truth.stdout, observation.stdout)
				}
			}
			if leaked := inputLeaks(t, how, program, binary); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
		})
	}
}
