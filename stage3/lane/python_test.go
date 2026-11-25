package lane

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
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
		"test_apply.Family25Tests.test_call_written_read_adapter",
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
	command := boundedPython(ctx, filepath.Join(root, "stage3"), "-E", "-S", "-B", "-c", `
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

// Keep each group below 25 seconds on one processor, including cases that invoke
// the generator more than once. Every case retains its own proof tree and process.
var generateApiPythonGroups = map[string][]string{
	"TestLaneGenerateApiAfterOptionalCompositionPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_after_optional_composes_and_guard_accepts",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_after_optional_requires_adaptation_20_suffix",
	},
	"TestLaneGenerateApiAfterOptionalRejectionsPython": {
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_after_optional_preserves_suffix",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_after_optional_requires_property_signature",
		"test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_after_optional_rejects_replace",
	},
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

func TestLaneGenerateApiAfterOptionalCompositionPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiAfterOptionalRejectionsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiPropertyGuardsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiMemberGuardsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiInsertGuardsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiOwnerRejectionsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiLineRejectionsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiClaimRejectionsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiInsertDeclarationRejectionsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiInsertFieldRejectionsPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_generate_api")
}

func TestLaneGenerateApiPythonGroupsCoverEveryTestOnce(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "stage3/lane/python_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
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
			t.Errorf("generator group %s has no wrapper", group)
			continue
		}
		for _, name := range names {
			counts[name]++
		}
	}
	for group := range active {
		if len(generateApiPythonGroups[group]) == 0 {
			t.Errorf("generator wrapper %s has no test names", group)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := boundedPython(ctx, filepath.Join(repositoryRoot(t), "stage3/lane"), "-E", "-S", "-B", "-c", `
import json, unittest, test_generate_api
def names(suite):
    for test in suite:
        if isinstance(test, unittest.TestSuite):
            yield from names(test)
        else:
            yield test.id()
print(json.dumps(list(names(unittest.defaultTestLoader.loadTestsFromModule(test_generate_api)))))
`)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("discover generator tests: %v\n%s", err, output)
	}
	var discovered []string
	if err := json.Unmarshal(output, &discovered); err != nil {
		t.Fatal(err)
	}
	for _, name := range discovered {
		if counts[name] != 1 {
			t.Errorf("%s covered %d times, want exactly once", name, counts[name])
		}
		delete(counts, name)
	}
	for name := range counts {
		t.Errorf("generator group names undiscovered test %s", name)
	}
	t.Logf("%d generator tests covered by %d wrappers", len(discovered), len(active))
}

func TestLaneSanctionedApiFreshnessPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_sanctioned_api_freshness")
}

func TestUpstreamLedgerPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/lane", "test_upstream_ledger")
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
	t.Logf("%s", runPythonCommand(t, "stage3/verdict", nil, "test_verdict.py"))
}

func TestVerdictAcceptanceRegenerationPython(t *testing.T) {
	t.Parallel()
	source := verdictRegenerationSource(t)
	t.Logf("%s", runPythonCommand(t, "stage3/verdict", source, "test_regeneration.py", "Regeneration.test_acceptance_regenerates"))
}

func TestVerdictUpstreamRegenerationPython(t *testing.T) {
	t.Parallel()
	source := verdictRegenerationSource(t)
	t.Logf("%s", runPythonCommand(t, "stage3/verdict", source, "test_regeneration.py", "Regeneration.test_upstream_regenerates"))
}

func verdictRegenerationSource(t *testing.T) []string {
	t.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout")
	}
	source, err := filepath.Abs(source)
	if err != nil {
		t.Fatal(err)
	}
	return []string{"ADAMIC_TYPESCRIPT_SOURCE=" + source}
}

func TestScannerScratchRunnerPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/drivers/scanner", "test_scratch_runner")
}

// any-returns' known-answer fixture is a script, not a unittest suite: it inspects a three-callee fixture with stock
// TypeScript, whose typescript and @types/node are stage3/api's install, and asserts its own lib-to-tsc mutant is
// caught. Run without --evidence it writes nothing into the tree (#mpesmv7).
func TestAnyReturnsFixture(t *testing.T) {
	t.Parallel()
	runPythonCommand(t, "stage3/census/any-returns", []string{"NODE_PATH=" + filepath.Join(repositoryRoot(t), "stage3/api/node_modules")}, "test_fixture.py")
}

func TestCensusPairPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/census/pair", "test_pair")
}

func TestFirstStopsDiffPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/census/firststops", "test_fsdiff")
}

func TestFirstLinkReportPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/firstlink", "test_report")
}

// The census replay's entry reach keeps the entry first, read from its source (#4ft3dyf). The replay's integration
// suite builds the census products and isn't wrapped yet (#pn630c6).
func TestReplayOrderPython(t *testing.T) {
	t.Parallel()
	runPython(t, "stage3/census/latent/replay", "test_replay_order")
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
	if directory == "stage3/lane" && len(modules) == 1 && modules[0] == "test_generate_api" {
		modules = generateApiPythonGroups[t.Name()]
		if len(modules) == 0 {
			t.Fatal("generator wrapper has no test names")
		}
	}

	runPythonCommand(t, directory, nil, append([]string{"-m", "unittest"}, modules...)...)
}

func pythonUnitBudget(directory string) time.Duration {
	if directory == "stage3/verdict" {
		return 30 * time.Second
	}
	return 55 * time.Second
}

func pythonUnitContext(directory string) (context.Context, context.CancelFunc, time.Duration) {
	limit := pythonUnitBudget(directory)
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	return ctx, cancel, limit
}

func TestPythonUnitBudget(t *testing.T) {
	t.Parallel()
	for directory, want := range map[string]time.Duration{
		"stage3/verdict": 30 * time.Second,
		"stage3/lane":    55 * time.Second,
	} {
		t.Run(directory, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				ctx, cancel, limit := pythonUnitContext(directory)
				defer cancel()

				<-ctx.Done()
				elapsed := time.Since(start)
				if limit != want || elapsed != want || ctx.Err() != context.DeadlineExceeded {
					t.Fatalf("unit budget %s, context expired after %s with %v; want %s and context deadline exceeded",
						limit, elapsed, ctx.Err(), want)
				}
				t.Logf("unit budget %s, context expired after %s with %v", limit, elapsed, ctx.Err())
			})
		})
	}
}

// The test clock advances after runPythonCommand creates its context and
// before capture begins. The expired context keeps the command from starting.
type expiredPythonUnit struct {
	testing.TB
	diagnostics   chan string
	contextBudget time.Duration
}

func (unit *expiredPythonUnit) observePythonUnitContext(ctx context.Context) {
	if deadline, ok := ctx.Deadline(); ok {
		unit.contextBudget = time.Until(deadline)
	}
	if unit.contextBudget != 30*time.Second {
		runtime.Goexit()
	}
}

func (unit expiredPythonUnit) TempDir() string {
	time.Sleep(30*time.Second + time.Nanosecond)
	return unit.TB.TempDir()
}

func (unit expiredPythonUnit) Fatalf(format string, args ...any) {
	unit.diagnostics <- fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func TestRunPythonCommandVerdictBudget(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		unit := expiredPythonUnit{TB: t, diagnostics: make(chan string, 1)}
		done := make(chan struct{})
		go func() {
			defer close(done)
			runPythonCommand(&unit, "stage3/verdict", nil, "-c", "raise AssertionError('expired command started')")
		}()
		<-done
		if unit.contextBudget != 30*time.Second {
			t.Fatalf("wrapper context deadline is %s from creation, want 30s", unit.contextBudget)
		}
		t.Logf("wrapper context deadline is %s from creation", unit.contextBudget)
		select {
		case report := <-unit.diagnostics:
			const want = "exceeded its 30s unit budget: context deadline exceeded"
			if !strings.Contains(report, want) || pythonUnitBudget("stage3/verdict") != 30*time.Second {
				t.Fatalf("want %q, got %q", want, report)
			}
			t.Log(strings.TrimSpace(report))
		default:
			t.Fatal("expired wrapper returned without a budget diagnostic")
		}
	})
}

// runPythonCommand runs python3 with arguments from directory, environment added to the toolchain's, and fails with
// its output when it fails.
func runPythonCommand(t testing.TB, directory string, environment []string, arguments ...string) []byte {
	t.Helper()
	root := repositoryRoot(t)
	ctx, cancel, limit := pythonUnitContext(directory)
	defer cancel()
	if observer, ok := t.(interface{ observePythonUnitContext(context.Context) }); ok {
		observer.observePythonUnitContext(ctx)
	}
	// -E and -S keep python off the box's PYTHON* settings and site directories, as the adapted tree's build runs it.
	command := boundedPython(ctx, filepath.Join(root, directory), append([]string{"-E", "-S", "-B"}, arguments...)...)
	// -B is this python's; the suites' own python children read PYTHONDONTWRITEBYTECODE, so none writes a
	// __pycache__ into the tree (stage3/oracle's, on Cloud, before).
	command.Env = append(command.Env, "STAGE3_CACHE="+t.TempDir(), "PYTHONDONTWRITEBYTECODE=1")
	// Ordinary apply clones the pinned TypeScript from the gate input, which the toolchain's environment doesn't carry.
	if source, ok := os.LookupEnv("ADAMIC_TYPESCRIPT_SOURCE"); ok {
		command.Env = append(command.Env, "ADAMIC_TYPESCRIPT_SOURCE="+source)
	}
	command.Env = append(command.Env, environment...)
	output, err := capturePythonCommand(ctx, command)
	if ctx.Err() != nil {
		t.Fatalf("%s in %s exceeded its %s unit budget: %v\n%s", command, directory, limit, ctx.Err(), output)
	}
	if err != nil {
		t.Fatalf("%s in %s: %v\n%s", command, directory, err, output)
	}
	return output
}

// Capture the streams within the same deadline as the command and its descendants.
func capturePythonCommand(ctx context.Context, command *exec.Cmd) ([]byte, error) {
	stdout, stderr, err := childguard.Capture(ctx, command)
	return append(stdout, stderr...), err
}

func TestPythonWrapperDeadlineWithInheritedStderr(t *testing.T) {
	t.Parallel()
	const marker = "child inherited stderr"
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	command := exec.CommandContext(ctx, "sh", "-c", `(printf '%s\n' "$1" >&2; exec sleep 4) & wait`, "stderr-holder", marker)
	start := time.Now()
	output, err := capturePythonCommand(ctx, command)
	elapsed := time.Since(start)
	if err == nil || ctx.Err() != context.DeadlineExceeded || !strings.Contains(string(output), marker) {
		t.Fatalf("expected expired deadline and child stderr marker, got %v, %v, %q", err, ctx.Err(), output)
	}
	if elapsed >= 2*time.Second {
		t.Fatalf("500ms deadline took %s while descendant held stderr: %s", elapsed, output)
	}
	t.Logf("500ms deadline returned in %s with %q", elapsed, strings.TrimSpace(string(output)))
}
