package lane

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// The two Python suites phase:stage3 ran beside the upstream suite are Go tests of their own, so Planner places each
// like any other test and a red names it: apply's tests (stage3/test_apply.py, whose ordinary-apply case adapts
// 00-setup alone) and the lane's checker and table tests. The lane's normalize-api.cjs reads stage3/api's install,
// which a tree holds, so neither needs the phase's npm ci. The lane's other suites, which only hand-run prove_*.py
// scripts ever ran, run beside them (#mpesmv7): a suite no gate runs can't fail.

// Explicit names make a newly discovered test red until a group takes it. Keep the
// module literal in each wrapper so the stage3 suite census still sees test_apply.
// The checkout case alone exceeds six seconds; splitting its wrapper cannot divide it.
var applyPythonGroups = map[string][]string{
	"TestApplyCheckoutPython": {
		"test_apply.ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean",
	},
	"TestApplyProductPython": {
		"test_apply.ProductTests.test_build_with_pinned_packages_never_runs_npm",
		"test_apply.ProductTests.test_build_without_api_packages_installs_nothing",
		"test_apply.ProductTests.test_corrupt_product_never_builds",
		"test_apply.ProductTests.test_existing_output_is_refused",
		"test_apply.ProductTests.test_input_key_covers_all_inputs",
		"test_apply.ProductTests.test_manifest_mutant_is_rejected",
		"test_apply.ProductTests.test_missing_product_builds_and_reports",
		"test_apply.ProductTests.test_new_adaptation_directory_changes_hash",
		"test_apply.ProductTests.test_only_manifest_absence_is_a_cache_miss",
		"test_apply.ProductTests.test_packages_come_installed_from_the_lockfile",
		"test_apply.ProductTests.test_packages_digest_pins_every_byte",
		"test_apply.ProductTests.test_payload_manifest_mutant_is_rejected",
		"test_apply.ProductTests.test_payload_mutant_is_rejected",
		"test_apply.ProductTests.test_publish_hook_describes_local_payloads",
		"test_apply.ProductTests.test_restore_fetches_verified_product",
		"test_apply.ProductTests.test_shards_partition_every_case",
		"test_apply.ProductTests.test_symlinked_directory_digests_as_its_link",
		"test_apply.ProductTests.test_typescript_source_is_the_pinned_gate_input",
		"test_apply.ProductTests.test_unconfigured_cold_cache_builds_without_remote_read",
		"test_apply.ProductTests.test_unkeyed_parser_override_is_refused",
		"test_apply.ProductTests.test_write_table_is_explicit",
	},
	"TestApplyProofShardPython": {
		"test_apply.ProofShardTests.test_every_indexed_read_file_has_one_shard",
	},
}

func TestApplyCheckoutPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3", "test_apply")
}

func TestApplyProductPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3", "test_apply")
}

func TestApplyProofShardPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3", "test_apply")
}

// Read the actual wrappers, not just the selector table: deleting a wrapper must
// leave its group and its test names red even while other wrappers keep the module covered.
func TestApplyPythonGroupsCoverEveryTestOnce(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "stage3/lane/python_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	active := map[string]bool{}
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestApply") || function.Body == nil {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) != 3 {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "runPython" {
				return true
			}
			directory, directoryOK := call.Args[1].(*ast.BasicLit)
			module, moduleOK := call.Args[2].(*ast.BasicLit)
			if directoryOK && moduleOK && directory.Value == `"stage3"` && module.Value == `"test_apply"` {
				active[function.Name.Name] = true
			}
			return true
		})
	}
	counts := map[string]int{}
	for group, names := range applyPythonGroups {
		if !active[group] {
			t.Errorf("apply group %s has no wrapper; tests: %s", group, strings.Join(names, ", "))
			continue
		}
		for _, name := range names {
			counts[name]++
		}
	}
	for group := range active {
		if len(applyPythonGroups[group]) == 0 {
			t.Errorf("apply wrapper %s has no test names", group)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := boundedCommand(ctx, filepath.Join(root, "stage3"), "python3", "-E", "-S", "-B", "-c", `
import json, unittest, test_apply
def names(suite):
    for test in suite:
        if isinstance(test, unittest.TestSuite):
            yield from names(test)
        else:
            yield test.id()
print(json.dumps(sorted(names(unittest.defaultTestLoader.loadTestsFromModule(test_apply)))))
`)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("discover apply tests: %v\n%s", err, output)
	}
	var discovered []string
	if err := json.Unmarshal(output, &discovered); err != nil {
		t.Fatalf("decode apply tests: %v\n%s", err, output)
	}
	for _, name := range discovered {
		if counts[name] != 1 {
			t.Errorf("%s covered %d times, want exactly once", name, counts[name])
		}
		delete(counts, name)
	}
	for name := range counts {
		t.Errorf("apply group names undiscovered test %s", name)
	}
	slices.Sort(discovered)
	t.Logf("%d apply tests covered by %d wrappers:\n%s", len(discovered), len(active), strings.Join(discovered, "\n"))
}

// Each lane suite is a test of its own, so a red names the suite and a test's tail is one suite's, never eight's
// under one bound.

func TestLaneCheckPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_check")
}

func TestLaneTablePython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_table")
}

func TestLaneArtifactPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_artifact")
}

func TestLaneCaseObserverPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_case_observer")
}

func TestLaneMeasurePlanPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_measure_plan")
}

func TestLaneProfilePython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_profile")
}

func TestLaneShardsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_shards")
}

func TestLaneWorkerEnvironmentPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_worker_environment")
}

// stage3's other suites run from their own directories, which their imports name (caches, verdict's run and census,
// the scanner driver's runner), so each is a test of its own here (#mpesmv7).

func TestCachesPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3", "test_caches")
}

func TestVerdictPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/verdict", "test_verdict")
}

func TestScannerScratchRunnerPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/drivers/scanner", "test_scratch_runner")
}

// runPython runs modules' unittest suites from directory, with a stage3 cache of the test's own, and fails with
// unittest's output when any of them fails.
func runPython(t *testing.T, directory string, modules ...string) {
	t.Helper()
	// Select by wrapper name while retaining the module literal for the file census.
	if directory == "stage3" && len(modules) == 1 && modules[0] == "test_apply" {
		modules = applyPythonGroups[t.Name()]
		if len(modules) == 0 {
			t.Fatalf("apply wrapper %s has no test names", t.Name())
		}
	}
	root := repositoryRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	// -E and -S keep python off the box's PYTHON* settings and site directories, as the adapted tree's build runs it.
	command := boundedCommand(ctx, filepath.Join(root, directory), "python3", append([]string{"-E", "-S", "-B", "-m", "unittest"}, modules...)...)
	// -B is this python's; the suites' own python children read PYTHONDONTWRITEBYTECODE, so none writes a
	// __pycache__ into the tree (stage3/oracle's, on Cloud, before).
	command.Env = append(command.Env, "STAGE3_CACHE="+t.TempDir(), "PYTHONDONTWRITEBYTECODE=1")
	// Ordinary apply clones the pinned TypeScript from the gate input, which the toolchain's environment doesn't carry.
	if source, ok := os.LookupEnv("ADAMIC_TYPESCRIPT_SOURCE"); ok {
		command.Env = append(command.Env, "ADAMIC_TYPESCRIPT_SOURCE="+source)
	}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%s in %s: %v\n%s", command, directory, err, output)
	}
}
