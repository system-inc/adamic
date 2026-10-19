import pathlib,json,subprocess,collections,datetime
out=pathlib.Path('review/compiler/loom-r-cycle-targets-ac-m2');base='5565a7cafb0ee8d538f8a5ece2304d5aa4254145'
r=json.loads((out/'test-summary.json').read_text()); by={x['log']:x for x in r}
text=f'''Fixed false closure-array cycle refusals and re-cut compiler/loom-r-cycle-targets-ac-m2 onto newest fetched main.
Source ca09c7495135717d4c79007ee6d4da560c255a26; base {base}; delivery commit SHA is in the final response.
Full historical fuzz comparison, required package runs, counts, guards, vet, gofmt and a-check are recorded below.
All six original implementation mutants, the retained Effects mutant, five new producer mutants and both header mutants are caught.
Not covered: full repository gate, full-tree a-check, ahra/Loom submission, or a renewed whole admission census; built-in WASI opt-in tests and benchmarks remain skipped.

Reproduction and cause

Run the two reported tests first at exact ca09c749 and exact parent 67733635b0252eb9d3bb3645c72262c7b18361e1, then the complete internal/fuzz package at each exact revision in isolated detached worktrees. Both raw outputs are retained, as are the focused outputs. The parent passes its whole package in 70.442s (longest leaf 45.65s). The source fails in 37.949s. In addition to the two reported failures, the full source run fails TestOwnershipShapes and TestOverridesShapesAndLower. All are compiler refusals rather than changed Node behavior.

Representative source diagnostic (ephemeral source paths omitted here only; the logs retain the complete output):

    fuzz_test.go:88: seed 31: stage 0 didn't lower it:
    program.a:42:7: Adamic 0.1 refuses (() => number)[], an array whose
    elements can reach back to an array like it: a cycle reference counting
    can't free, and the write at program.a:522:3 may close one ... (adamic/cycle-capable)

Parent output:

    PASS
    ok  github.com/system-inc/adamic/internal/fuzz 70.416s

The source solver pools directly stored function-valued array elements. It also marked the entire pool Unknown for unrelated arrays: regex results, scalar parallelMap results, unbound parameters whose direct elements cannot be functions, and the synthetic array parameter of a widened intrinsic callback adapter. The last parameter has an actual element ABI, but that evidence was erased. One unknown invocation then conservatively poisoned these adapter inputs too. The cycle finder consequently rejected safe closure arrays elsewhere in the generated program.

The original failing seed 3 is also Agreed against source Node, JavaScript and sanitized native: all three exit 0 with identical stdout (seed3-node-native.log/json). The first CLI attempt used a relative workspace that could not resolve its compiler executable; that setup failure is retained in seed3-relative-work.log. The successful run uses an absolute evidence workspace. Disposable binaries and the duplicate generated program.a are removed after proving the source bytes match the retained seed3.a.

The narrow fix preserves ir.Local.ArrayElement from checked declarations and the intrinsic callback adapter's supplied ABI. Array parameters with a known non-callable element representation no longer supply Unknown directly to the function-element pool. Zero (no evidence), Closure and Union element representations remain conservative, for both unbound inputs and unknown callers. Regex array results contain strings/scalar metadata; parallelMap is exempted only for the enumerated scalar result representations. Non-scalar parallel results remain Unknown. Nested objects/collections keep their separate field/element cycle edges. The closure-cycle algorithm, captures, foreign boundary poisoning and empty-target safeguards remain intact.

Assumption: only immediate function-valued elements belong in this pooled function-slot graph. The checked element representation and intrinsic supplied ABI prove that boundary; a nominal callback signature does not supply a closure target. Existing fuzz expectations and generator sources are unchanged. No check is weakened to accept a new refusal.

New coverage: TestClosureTargetsIgnoreNonCallableArrayInputs pins the bounded producer set despite unrelated inputs/results; TestClosureTargetsKeepCallableArrayInputsUnknown pins callable, union and unproved inputs and an unmodelled callable parallel result. TestClosureArraysWithUnrelatedProducersAgreeWithNode compiles a small combination of the generated shapes and requires Agreed against source Node, emitted JavaScript and native under AddressSanitizer, UndefinedBehaviorSanitizer and leak checking. Expected stdout is 7, 1, 2,3, 1,2 on four lines. Its immutable parallel input respects the existing task boundary. New tests call t.Parallel first and are not split or deleted to conceal a slow leaf.

Re-cut and conflicts

Fetch origin and select newest origin/main, exactly {base}. Apply only ca09c749's own parent-to-tip change with git merge-tree --write-tree --merge-base=ca09c749~1 origin/main ca09c749 and git cherry-pick --no-commit ca09c749. This newest base merges without conflicts; merge-tree.log and apply.log retain the result. No newer compiler infrastructure is reverted. The cwd expected-refusal header is removed because the branch now admits it; that expectation cannot coexist with the change. Main's slice 2 Effects result identity remains in the shared producer solver. Restore main's complete counts.md before measuring, discarding the automatic textual table merge.

Runtime clearance

Neither ca09c749's own change nor this re-cut changes internal/native/runtime/*.c or *.h. Exact signed extra runtime lines against main: none. runtime-beyond-source.patch is empty. No runtime clearance is needed for an additional line.

Counts

The initial anchored ^TestCountsAreRecorded$ command selected no tests on main's new sharded implementation and wrote no table. The actual regeneration uses the documented unanchored selector:

    timeout 630 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -parallel=4 -json -timeout=600s -args -update-counts

PASS in 154.846s, longest leaf 18.89s, on Linux with house clang 20.1.8. The table is generated once successfully, never edited by hand. Eight existing rows change and eleven rows are added; none are removed. Every numeric delta, and all additions, exactly match ca09c749's own parent-to-tip table delta. The fuzz fix adds no additional oracle count movement. counts-rows.json compares every row with main and source; all same_as_source_delta entries are true.

'''
text+=(out/'counts-rows.md').read_text()
text+='''
Validation commands and outputs

Historical commands, at each exact detached revision:

    timeout 150 go test ./internal/fuzz -run '^(TestUndefinedNumbersShapes|TestGeneratedProgramsCheckAndLower)$' -count=1 -parallel=4 -json -timeout=120s
    timeout 630 go test ./internal/fuzz -count=1 -parallel=4 -json -timeout=600s

Final full-package commands are in final-package-results.json and package-results-before-disk-recovery.json; every process writes raw JSON directly to its log. They run go test ./internal/fuzz, ./internal/ir, ./internal/lower, ./internal/native, ./internal/javascript, ./internal/flow, ./internal/fresh and ./internal/oracle, with -count=1 -parallel=4 -json and a hard package timeout. No full gate runs.

| Package | Result | Go package seconds | Maximum leaf seconds |
|---|---|---:|---:|
'''
for name in ['internal/fuzz','internal/ir','internal/lower','internal/native','internal/javascript','internal/flow','internal/fresh','internal/oracle']:
 entry=next((e for e in r if any(p.get('Package','').endswith('/'+name)for p in e['packages']) and e['log'].startswith('final-package')),None)
 if entry:
  p=next(p for p in entry['packages']if p.get('Package','').endswith('/'+name));text+=f"| {name} | {p['Action']} | {p.get('Elapsed')} | {entry['max_leaf_seconds']} |\n"
 else:text+=f'| {name} | IN PROGRESS | | |\n'
text+='\nNew test seconds (the three tests added by this fix; new-test-seconds.json includes all eleven added tests in the complete re-cut):\n\n'
for entry in r:
 if entry['log'].startswith('final-package'):
  for e in entry['new_tests']:text+=f"- {e['Test']}: {e.get('Elapsed',0):.2f}s, {e['Action']}.\n"
text+='''
The explicit guards command uses -timeout=90s and selects TestCallTargetReaders.*, TestEveryMapAProgramHoldsIsSorted, TestNativeAgreementRowsNameRegisteredPrograms, TestRuntimeSetsAreEachUsed and TestEveryTestIsParallelOrSaysWhy across internal/ir, internal/lower, internal/oracle and cmd/adamic-gate. All four packages pass; longest guard leaf 2.86s. Full IR also covers the export-input and reader-product checks. No new map is introduced; the solver's held map remains SortedMap. There is no new compiler-consuming package. New products/tests introduce no build-directory or machine-path name. Existing runtime-set witnesses and literal native-agreement registrations remain valid.

Go vet passes for internal/fuzz, internal/ir, internal/lower, internal/oracle, internal/native, internal/javascript, internal/flow and internal/fresh. gofmt -l over every changed Go file and git diff --check over internal/docs produce empty output. No test is removed or given a new skip; pre-existing platform/opt-in skips are retained in test-summary.json and the raw logs. test-summary.json records all observed final and historical leaf times; no leaf reaches 60s.

Native records 94 built-in skips; native-skips.json retains each exact reason. No new skip or selector excludes an ordinary native test.

A-check

Read stage3/a-check-headers/README.md whole. Build this branch's own compiler and invoke the unchanged pinned Gate.aCheck predicate from fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py. All six changed .a files pass the predicate in place. Status counts: checked 3, NotYet 0, refused 3, checker error 0. The three source refusal files are byte-identical to ca09c749, headers included. The cwd source body is also unchanged; only its obsolete expected-refusal line is removed. No header is added. Remove the stale first-line refusal header from stage3/namespace-init-sys/cwd.a, which this change admits. Its program body is byte-identical and its emitted C is byte-identical before/after removal (cwd-header-removal.json), so its already-measured count row is unchanged. Grep all test/testdata path readers: no pinned position changes. The lower and oracle cwd readers and all reader/policy guards pass again after removal (header-reader-checks.log). cwd-stale-header.json first proves the old header fails the unchanged a-check predicate on the actual clean compile.

| Path | Compiler status |
|---|---|
'''
for a in json.loads((out/'a-check-status.json').read_text()):text+=f"| {a['path']} | {a['status']} |\n"
text+='''
The refused direct/joined files retain the captured-memo refusal header; map retains the Map-value cycle refusal header. The three clean files have no header. check-headers.py invokes the real compiler for each file, then replays its actual diagnostics through the unchanged predicate with in-memory removed-header and wrong-reason mutants. Both fail; restored headers pass. No fixture is mutated on disk during the proof.

Mutants and catchers

| Mutant | Catcher | Result |
|---|---|---|
'''
for m in json.loads((out/'mutants.json').read_text()):text+=f"| {m['mutant']} | {m['catcher']}"+((' and '+m['additional_catcher'])if'additional_catcher'in m else'')+f" | {'caught'if m['caught']else'FAILED'} |\n"
text+='''
These twelve implementation mutants fail their intended assertions, not compilation, timeout or process setup. Original admission type matching is caught independently in lower and oracle (type-matched-captures-both.log). The two header mutants are recorded separately in header-mutants.json. Mutants use Go overlays, with .go.txt and .patch evidence, so production sources remain unchanged while other validation runs.

Environment and recoveries

Set GOPROXY=https://proxy.golang.org|direct; run bash cloud/setup.sh --wasi-sdk and source its printed /workspace/adamic-tools/env.sh. Setup succeeds. Timing lines are retained in setup.log: go ready 0.025s, node ready 0.025s, markdown ready 0.077s, submodules 0.083s, clang 0.181s, WASI SDK 0.275s, shared cache 1.223s, go build 124.024s, build cache warm 124.208s, total 124.236s. nproc is 5; cpu.max is 400000 100000, so the instance has a 4-CPU quota. House clang is exactly 20.1.8; WASI SDK 27 is present. Tests use GOMAXPROCS=4, -parallel=4 and ADAMIC_GOCACHE_OFF=1 with the external GOCACHEPROG unset; ordinary Go/product caches remain enabled. The pre-existing worker GCC wrapper supplies LLVM's public sanitizer header to genuine /usr/bin/gcc; no runtime/compiler source is changed for it.

Historical worktrees initially lacked the nested TypeScript submodule; both setup-failure outputs are retained separately. Their cohere submodules reference the exact already-initialized TypeScript revision d92d9bfee114c80be2c375d72edae966176e3a4f, matching the submodule pointer. No cohere implementation is copied into Adamic. Full historical tests then run and their worktrees are removed after saving the outputs.

The first full IR/explicit reader runs fail only with 'no space left on device' while materializing reader export archives. The raw failed logs are retained as ir-disk-full.log and guards-disk-full.log. Reclaim disk by hardlinking byte-identical completed published export archives, checking full SHA-256, size, mode and ownership, preserving every product path/content; also remove two abandoned Go process scratch directories from 19:43 which no live process references. No standard or product cache is flushed. Full IR and explicit guards then pass. No test leaf exceeds its budget or is rerun after exceeding 60s.

Delivery

One commit on the stated main base preserves the original message body, adds the requested re-cut/fix line and conservative assumption, and retains the Ahra co-author trailer. Push only compiler/loom-r-cycle-targets-ac-m2, without force. No PR, main push, old-branch push, ahra invocation or Loom submission. Delivery exceeds the 30-minute target while completing the exact historical comparisons, disk recovery and mandatory full-package validation.
'''
(out/'REPORT.md').write_text(text)
