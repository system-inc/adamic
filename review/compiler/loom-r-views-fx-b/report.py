from pathlib import Path
import json,gzip,subprocess
p=Path('review/compiler/loom-r-views-fx-b')
base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
checks=['own-final','fixtures-final','flow','stage3','vet-final','maps','nullable-green','optional-green','generic-green','moved-rows','counts','readers']
results={n:json.loads((p/(n+'.json')).read_text()) for n in checks}
assert all(x['exit']==0 for x in results.values())
assert len([t for t in results['fixtures-final']['tests'] if '/' in t['Test']])==14
assert len(results['moved-rows']['tests'])==6
assert all(x['caught'] for x in json.loads((p/'mutants.json').read_text()))
assert len(json.loads((p/'counts-rows.json').read_text()))==20
leaves=[dict(check=n,**t) for n,x in results.items() for t in x['tests']]
(p/'leaf-seconds.json').write_text(json.dumps(leaves,indent=2)+'\n')
text=f'''Ported #1fk58py deliveries 134, 136 and 144, retaining main's newer lowering and splitting the scalar tests by backend.
Commit: one delivery commit on compiler/loom-r-views-fx-b; its SHA is in the final response. Base: {base}.
Commands and outputs: focused tests, 14 fixtures, flow, six moved-count programs, six mutants, counts once, 52 reader packages, reader guard and vet passed.
Mutants: receiver, destructuring, scalar conversion, nullable, generic and union receiver mutants all caught by behavioral assertions.
Not covered: whole-package/full repository gates, other numbered deliveries, ahra CLI, Loom submission or runtime C changes.

Source: 902c13f7ffb4070b0c2b79b28f4167e07a730bd5, parent bd90a433bf636b8325a591ab3e7487d457a69ab4. Fetched origin and the source, created the requested branch from exactly dbdf977f5b, and applied `git cherry-pick --no-commit 902c13f7ffb4070b0c2b79b28f4167e07a730bd5`. This uses the source's own parent-to-commit change. No main-to-source diff or later-main rebase was used. Conflict count: zero; no conflicted hunks. Git automatically merged internal/lower/object.go and internal/oracle/counts.md. Counts were subsequently regenerated, never hand-merged.

134 qualifies checked fields by receiver type and field name, preserving nullable, generic and union receiver boundaries. 136 carries receiver, contract and location metadata through destructuring. 144 validates optional scalar membership and kind before packing a value, preserving absent versus present zero/false. Main's bare-name record-storage boundary remains. No IR map field or package was added; main's SortedMap and deterministic gob tests pass. Compiler consumer declarations therefore need no addition. No product key or test output names a build directory or machine path; absolute input fixture resolution follows the existing harness. Evidence overlays record this local run's paths and are not products.

Conservative assumption retained from the source: the invalid object-union p49 and lying callable p51_wrong stop at their checked reads, with exit 70 and pinned diagnostics, although unchecked source Node continues. These are runtime checked failures. Every new fixture lowers, so no compiler-refused top-level oracle fixture or *_refused directory is needed. No cohere code was copied.

Runtime C versus pinned main: no files or lines changed under internal/native/runtime. Source's own runtime C change: also empty. Additional runtime C lines requiring clearance: none. Native backend Go emission changes only internal/native/view_unions_read.go.

The only adaptation beyond the source change is internal/oracle/scalar_union_views_test.go: the two combined misfit tests became six top-level tests, one per backend and scalar type, retaining Node control and exact failure assertions. The initial cold broad command reached its 90-second oracle timeout while unrelated existing runtime-mutant regressions were running. Its new TestScalarUnionViewMisfitBoolean passed in 63.56 seconds, exceeding the law; that function was split and never rerun. Existing unmodified TestViewFieldInheritedStaticReadiness (63.64), TestRequiredViewFieldOperandOnce (63.65), and TestCheckedViewFlagDowncast/tsc_source (64.62) were not rerun or edited. They shared cold runtime builds with other commands. This report does not claim the broad command passed. The final source-owned test layout passed with maximum 15.92 seconds per top-level test, setup included. Loom supplies the existing declared runtime products before running its cold test units.

An early reader build raced the scalar test rename and its generated _testmain.go referenced the old names. That stale command was stopped; the final 52-package command passed after edits completed. Early fixture selectors matched no fixture leaves; fixtures-final.json is the corrected run with all 14 leaves. Stage3 moved-count programs are not in that fixture family; moved-rows.json verifies all six directly. The inherited review/compiler/loom-views-fx-b/ evidence is source provenance, not a fresh result on this base.

Linux setup: `GOPROXY='https://proxy.golang.org|direct' timeout 600 bash cloud/setup.sh`, then source /workspace/adamic-tools/env.sh. nproc=5, cgroup cpu.max=400000 100000 (four CPUs), GOMAXPROCS=4 for verification. Go 1.27.1, Node 24.19.0, clang 20.1.8. Timing lines: Go 0.097s; Node 0.106s; markdown ready 0.234s (installed dependency validation 0.013s); clang 0.435s; submodules 1.987s; shared cache 2.937s; go build 67.968s; deferred test binaries 68.055s; build cache warm 68.056s; done 68.082s. Setup passed. Complete output is in raw-logs.json.gz.

Exact commands, exit codes, wall seconds and every test's seconds are in the per-command JSON files and leaf-seconds.json. All test output went directly to log files. Every Go test had an internal hard timeout; every runner also had a subprocess and outer timeout. No whole-package test run or full gate was used. Reader packages without matching tests were compiled and reported no tests to run. The closure includes Imports, TestImports and XTestImports and is in reader-packages.json.

| Check | Exit | Command wall seconds | Coverage |
|---|---:|---:|---|
'''
for n,x in results.items():text+=f"| {n}.json | {x['exit']} | {x['seconds']} | {len(x['tests'])} recorded test results |\n"
text+='''
Focused native TestObjectMetadataCacheIsPerEmission passed in 0.00s and JavaScript TestViewMixedUnionSelectsNumber in 0.18s during the initial command; both package processes completed successfully. TestRuntimeSetsAreEachUsed passed in 8.35s: no new or stale runtime set was introduced. The final TestCallTargetReaders passed in 6.95s with -timeout 90s. Vet passed on lower, oracle, native, javascript, flow and ir; final post-split vet of all four changed packages also passed without diagnostics.

The flow run covers its corpus metadata, remainder and view/predicate/interface/union units, plus each of the 14 new programs as a separate evidence-only top-level test calling all four flow checks. Maximum new flow test: 1.09s. Ordinary oracle fixtures compare source Node, JavaScript, release native and ASan/UBSan native, and check leaks for successful programs. Checked failures are held to their explicit diagnostic and exit 70. The six moved-count programs independently pass the same four lanes and leak checks through evidence-only top-level tests.

Final source-owned top-level seconds (package wall/compilation time is separately recorded above):

| Package | Test | Seconds |
|---|---|---:|
'''
for t in results['own-final']['tests']:
 if '/' not in t['Test']:text+=f"| {t['Package'].rsplit('/',1)[-1]} | {t['Test']} | {t.get('Elapsed',0):.2f} |\n"
text+='''
Mutants use independent Go overlays with sources ending .go.txt. Production sources are never mutated. All six expected commands exited 1 because their intended behavioral check failed, with no build failure, compiler warning or timeout counted as a kill. Exact commands, seconds and failing leaves are in *-mutant.json, mutants.json and scalar-mutant-final.json.

| Mutant | Catcher |
|---|---|
| Receiver keys collapse to receiver zero | lower TestFX6P37: JavaScript stdout differs from Node |
| Destructured ViewTypeID removed | lower TestFX6P53: JavaScript stdout differs from Node |
| Native optional-scalar validation removed | scalar misfit native and sanitized tests: expected checked exit 70, got 0; final split layout catches it in four leaves |
| Nullable receiver normalization removed | TestCheckedViewOptionalReadBoundary: optional checked read escaped |
| Generic receiver left rigid | TestCheckedViewUntaggedSourceFlows/generic/wrong: exit codes differ |
| Union receiver propagation removed | TestFixturesAssertions/assertions/19_identifier_kind.a/stage0: pinned gap changed |

The unmodified nullable, optional receiver, generic and stage3 catchers all passed separately. The scalar mutant was also run against the final six-test layout; its two JavaScript controls passed and four native/sanitized leaves failed as intended.

Counts were regenerated exactly once with Linux house clang 20.1.8 using `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -timeout 540s -args -update-counts`, outer limit 600s. Exit 0, 299.884s command wall time, 1,105 recorded test results. Regeneration made no further edit beyond the automatically applied source counts. Against pinned main: 14 added fixture rows, six modified rows, zero removed. Full rows are in counts-rows.json. Added rows are the nine fx6 fixtures and five scalar_union_views fixtures. Successful rows balance allocations with frees; checked failures stop before normal cleanup.

Existing moved rows, under stage3/interface-downcasts/:

| Fixture | Retains before -> after | Releases before -> after |
|---|---:|---:|
| visitor.a | 32 -> 28 | 40 -> 36 |
| lane1/interfaces-good.a | 6 -> 5 | 12 -> 11 |
| lane1/objects-good.a | 6 -> 4 | 12 -> 10 |
| untagged/fixtures/binding-name-source-good.a | 96 -> 84 | 144 -> 132 |
| untagged/fixtures/option-element-source-good.a | 40 -> 35 | 60 -> 55 |
| untagged/fixtures/structural-source-good.a | 16 -> 14 | 24 -> 22 |

Allocations, frees, peak and regions do not move in these six rows. Receiver-specific checks stop guarding unrelated same-name reads, removing matched retain/release pairs; total retains and releases each decrease by 26. Numeric changes are measured observations, and all six programs still agree with Node in the direct four-lane checks.

No PR, ahra command, Loom submission, main push or force push was made. The requested pinned main base is retained even if origin/main later moves.
'''
(p/'REPORT.md').write_text(text)
logs={f.name:f.read_text(errors='replace') for f in sorted(p.glob('*.log'))}
logs['stale-reader-build-observation']='The earlier reader build raced the rename: _testmain.go undefined: _test.TestScalarUnionViewMisfitNumber and _test.TestScalarUnionViewMisfitBoolean. Stopped; final command passed.'
with gzip.open(p/'raw-logs.json.gz','wt') as f:json.dump(logs,f)
print('report ready',len(text),len(logs))
