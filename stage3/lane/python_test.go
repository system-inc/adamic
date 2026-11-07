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

func TestApplyCheckoutPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3", "test_apply")
}

func TestApplyProductPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3", "test_apply")
}

func TestApplyProofShardPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3", "test_apply")
}

// Read the actual wrappers, not just the selector table: deleting a wrapper must
// leave its group and its test names red even while other wrappers keep the module covered.
func TestApplyPythonGroupsCoverEveryTestOnce(test *testing.T) {
	test.Parallel()
	root := repositoryRoot(test)
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "stage3/lane/python_test.go"), nil, 0)
	if err != nil {
		test.Fatal(err)
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
			test.Errorf("apply group %s has no wrapper; tests: %s", group, strings.Join(names, ", "))
			continue
		}
		for _, name := range names {
			counts[name]++
		}
	}
	for group := range active {
		if len(applyPythonGroups[group]) == 0 {
			test.Errorf("apply wrapper %s has no test names", group)
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
		test.Fatalf("discover apply tests: %v\n%s", err, output)
	}
	var discovered []string
	if err := json.Unmarshal(output, &discovered); err != nil {
		test.Fatalf("decode apply tests: %v\n%s", err, output)
	}
	for _, name := range discovered {
		if counts[name] != 1 {
			test.Errorf("%s covered %d times, want exactly once", name, counts[name])
		}
		delete(counts, name)
	}
	for name := range counts {
		test.Errorf("apply group names undiscovered test %s", name)
	}
	slices.Sort(discovered)
	test.Logf(
		"%d apply tests covered by %d wrappers:\n%s",
		len(discovered),
		len(active),
		strings.Join(discovered, "\n"),
	)
}

// Each lane suite is a test of its own, so a red names the suite and a test's tail is one suite's, never eight's
// under one bound.

func TestLaneCheckPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_check")
}

// Keep each group below 25 seconds on one processor, including cases that invoke
// the generator more than once. Every case retains its own proof tree and process.
var generateApiPythonGroups = map[string][]string{
	"TestLaneGenerateApiPropertyGuardsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_line_applied_and_guard_accepts",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_applied_and_guard_accepts",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_undo_is_rejected",
	},
	"TestLaneGenerateApiMemberGuardsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_member_applied_and_guard_accepts",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_member_undo_is_rejected",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_member_before_requires_full_text",
	},
	"TestLaneGenerateApiInsertGuardsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_insert_applied_and_guard_accepts",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_insert_undo_is_rejected",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_function_return_applied_and_guard_accepts",
	},
	"TestLaneGenerateApiOwnerRejectionsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_ambiguous_path_without_line_is_rejected",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_ambiguous_owner",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_missing_owner",
	},
	"TestLaneGenerateApiLineRejectionsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_wrong_line_names_path_and_line",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_non_integer_line_is_rejected",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_before_mismatch_names_row",
	},
	"TestLaneGenerateApiClaimRejectionsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_duplicate_claim",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_hard_coded_claim",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_invalid_replace_is_rejected",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_overloaded_function_is_rejected",
	},
	"TestLaneGenerateApiInsertDeclarationRejectionsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_insert_non_declaration_is_rejected",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_insert_existing_name_is_rejected",
	},
	"TestLaneGenerateApiInsertFieldRejectionsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_insert_invalid_fields_are_rejected",
	},
}

func TestLaneGenerateApiPropertyGuardsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiMemberGuardsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiInsertGuardsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiOwnerRejectionsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiLineRejectionsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiClaimRejectionsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiInsertDeclarationRejectionsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiInsertFieldRejectionsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiPythonGroupsCoverEveryTestOnce(test *testing.T) {
	test.Parallel()
	root := repositoryRoot(test)
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "stage3/lane/python_test.go"), nil, 0)
	if err != nil {
		test.Fatal(err)
	}
	active := map[string]bool{}
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "Test") || function.Body == nil {
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
			if directoryOK && moduleOK && directory.Value == `"stage3/lane"` && module.Value == `"test_generate_api"` {
				active[function.Name.Name] = true
			}
			return true
		})
	}
	counts := map[string]int{}
	for group, names := range generateApiPythonGroups {
		if !active[group] {
			test.Errorf("generator group %s has no wrapper", group)
			continue
		}
		for _, name := range names {
			counts[name]++
		}
	}
	for group := range active {
		if len(generateApiPythonGroups[group]) == 0 {
			test.Errorf("generator wrapper %s has no test names", group)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := boundedCommand(
		ctx,
		filepath.Join(repositoryRoot(test), "stage3/lane"),
		"python3",
		"-E",
		"-S",
		"-B",
		"-c",
		`
import json, unittest, test_generate_api
def names(suite):
    for test in suite:
        if isinstance(test, unittest.TestSuite):
            yield from names(test)
        else:
            yield test.id()
print(json.dumps(list(names(unittest.defaultTestLoader.loadTestsFromModule(test_generate_api)))))
`,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		test.Fatalf("discover generator tests: %v\n%s", err, output)
	}
	var discovered []string
	if err := json.Unmarshal(output, &discovered); err != nil {
		test.Fatal(err)
	}
	for _, name := range discovered {
		if counts[name] != 1 {
			test.Errorf("%s covered %d times, want exactly once", name, counts[name])
		}
		delete(counts, name)
	}
	for name := range counts {
		test.Errorf("generator group names undiscovered test %s", name)
	}
	test.Logf("%d generator tests covered by %d wrappers", len(discovered), len(active))
}

func TestLaneTablePython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_table")
}

func TestLaneArtifactPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_artifact")
}

func TestLaneCaseObserverPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_case_observer")
}

func TestLaneMeasurePlanPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_measure_plan")
}

func TestLaneProfilePython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_profile")
}

func TestLaneShardsPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_shards")
}

func TestLaneWorkerEnvironmentPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/lane", "test_worker_environment")
}

// stage3's other suites run from their own directories, which their imports name (caches, verdict's run and census,
// the scanner driver's runner), so each is a test of its own here (#mpesmv7).

func TestCachesPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3", "test_caches")
}

func TestVerdictPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/verdict", "test_verdict")
}

func TestScannerScratchRunnerPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/drivers/scanner", "test_scratch_runner")
}

// any-returns' known-answer fixture is a script, not a unittest suite: it inspects a three-callee fixture with stock
// TypeScript, whose typescript and @types/node are stage3/api's install, and asserts its own lib-to-tsc mutant is
// caught. Run without --evidence it writes nothing into the tree (#mpesmv7).
func TestAnyReturnsFixture(test *testing.T) {
	test.Parallel()
	runPythonCommand(
		test,
		"stage3/census/any-returns",
		[]string{"NODE_PATH=" + filepath.Join(repositoryRoot(test), "stage3/api/node_modules")},
		"test_fixture.py",
	)
}

func TestCensusPairPython(test *testing.T) {
	test.Parallel()
	runPython(test, "stage3/census/pair", "test_pair")
}

// runPython runs modules' unittest suites from directory, with a stage3 cache of the test's own, and fails with
// unittest's output when any of them fails.
func runPython(test *testing.T, directory string, modules ...string) {
	test.Helper()
	// Select by wrapper name while retaining the module literal for the file census.
	if directory == "stage3" && len(modules) == 1 && modules[0] == "test_apply" {
		modules = applyPythonGroups[test.Name()]
		if len(modules) == 0 {
			test.Fatalf("apply wrapper %s has no test names", test.Name())
		}
	}
	if directory == "stage3/lane" && len(modules) == 1 && modules[0] == "test_generate_api" {
		modules = generateApiPythonGroups[test.Name()]
		if len(modules) == 0 {
			test.Fatal("generator wrapper has no test names")
		}
	}

	runPythonCommand(test, directory, nil, append([]string{"-m", "unittest"}, modules...)...)
}

// runPythonCommand runs python3 with arguments from directory, environment added to the toolchain's, and fails with
// its output when it fails.
func runPythonCommand(test *testing.T, directory string, environment []string, arguments ...string) {
	test.Helper()
	root := repositoryRoot(test)
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	// -E and -S keep python off the box's PYTHON* settings and site directories, as the adapted tree's build runs it.
	command := boundedCommand(
		ctx,
		filepath.Join(root, directory),
		"python3",
		append([]string{"-E", "-S", "-B"}, arguments...)...,
	)
	// -B is this python's; the suites' own python children read PYTHONDONTWRITEBYTECODE, so none writes a
	// __pycache__ into the tree (stage3/oracle's, on Cloud, before).
	command.Env = append(command.Env, "STAGE3_CACHE="+test.TempDir(), "PYTHONDONTWRITEBYTECODE=1")
	command.Env = append(command.Env, environment...)
	// Ordinary apply clones the pinned TypeScript from the gate input, which the toolchain's environment doesn't carry.
	if source, ok := os.LookupEnv("ADAMIC_TYPESCRIPT_SOURCE"); ok {
		command.Env = append(command.Env, "ADAMIC_TYPESCRIPT_SOURCE="+source)
	}
	if output, err := command.CombinedOutput(); err != nil {
		test.Fatalf("%s in %s: %v\n%s", command, directory, err, output)
	}
}
