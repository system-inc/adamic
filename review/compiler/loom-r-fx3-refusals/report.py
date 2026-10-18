from pathlib import Path
import json,re,subprocess
out=Path('review/compiler/loom-r-fx3-refusals')
readers=json.loads((out/'consumer-inventory.json').read_text()); declared=json.loads((out/'declared-reader-selection.json').read_text())
passed={'internal/lower','internal/oracle','internal/ir','cmd/adamic'}
for file in ['consumer-results.json','consumer-final-results.json','consumer-rest-results.json','declared-reader-results.json','declared-reader-final-results.json','extra-reader-results.json']:
    if (out/file).exists():
        for r in json.loads((out/file).read_text()):
            if r['exit']==0:passed.add(r['package'][2:])
assert '--- PASS: TestApplyPython' in (out/'lane-pinned-input.log').read_text()
passed.add('stage3/lane')
extras=json.loads((out/'extra-reader-selection.json').read_text())
expected=set(readers)|{r['package'][2:] for r in declared}|{r['package'][2:] for r in extras}
assert expected<=passed,sorted(expected-passed)
leaves=[];skips=[]
for path in out.glob('*.log'):
    text=path.read_text(errors='replace')
    for name,seconds in re.findall(r'--- PASS: (\S+) \(([\d.]+)s\)',text):leaves.append(dict(log=path.name,test=name,seconds=float(seconds)))
    for name,seconds in re.findall(r'--- SKIP: (\S+) \(([\d.]+)s\)',text):skips.append(dict(log=path.name,test=name,seconds=float(seconds)))
assert all(r['seconds']<60 for r in leaves),sorted(leaves,key=lambda r:r['seconds'])[-5:]
(out/'test-leaf-seconds.json').write_text(json.dumps(leaves,indent=2)+'\n');(out/'skips.json').write_text(json.dumps(skips,indent=2)+'\n')
assert Path('internal/oracle/counts.md').read_bytes()==subprocess.check_output(['git','show','dbdf977f5b:internal/oracle/counts.md'])
audit=json.loads((out/'patch-audit.json').read_text());assert all(r['added_removed_lines_identical'] for r in audit['internal_patch'])
assert len(json.loads((out/'mutants.json').read_text()))==4
own=[r for r in leaves if r['log']=='lower-focus.log' and (r['test'].startswith('TestArrayNarrowing') or r['test'].startswith('TestFX5'))]
setup='\n'.join(line for line in (out/'setup.log').read_text().splitlines() if line.startswith('setup:'))
validation=json.loads((out/'validation.json').read_text())
commands='\n'.join('- '+ ' '.join(r['command'])+': exit '+str(r['exit'])+', command '+str(r['seconds'])+'s.' for r in validation)
text=f'''Rebased the three narrow refusals for items 127, 128 and 130, retaining the seven named native controls.
Commit: one branch HEAD supplied in the push report; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 088a27988a63171d8485e921973018df27688779.
Results: focused lower/oracle, {len(expected)} reader packages, guards, vet, counts and all seven Node/JavaScript/sanitized-native controls pass.
Mutants: every, findLast, literal and registration each fail their intended assertion; no build-error kill is counted.
Not covered: whole packages/full gate, ahra/Loom submission, WASI/macOS, optional estree library and cycle-ledger inputs.

Applied only the named commit's own change with git cherry-pick --no-commit, on the exact requested base. Conflict list: empty. No hunk needed manual resolution. Main's newer librarySpreadRefusal validation for non-object spreads and librarySpreads tracking remain in objectLiteral; the refusal changes merge automatically at elementType and arrayVisit. patch-audit.json confirms identical added/removed lines for every internal path versus the source commit's own diff, not its ancestors.

Runtime clearance: source runtime diff is empty; rebased runtime diff is empty. Zero runtime C lines changed, and there is no difference to list.

Counts: zero rows changed. The single completed Linux -update-counts regeneration passed in 52.521s with house clang 20.1.8 and rewrote identical bytes. counts-audit.json verifies equality to the exact base. The first attempt hit the aggregate 90.035s package timeout before TestMain could write any table; the four active leaves were only 0-1 seconds old. The completed attempt used a bounded 300s package timeout and -parallel 4. No observed test leaf exceeded 60s, and no over-budget leaf was rerun. No new oracle fixture is added: refused witnesses remain lower testdata, outside the top-level oracle corpus, so no oracle refused-directory relocation is needed. Assumption: one successful complete rewrite fulfills regeneration; the aborted process wrote no table and is retained as a failed attempt.

Guard results:

- TestCallTargetReaders passes in 1.23s with -timeout 90s; its prebuilt export product passes in 44.48s. No call-target reader or allowlist changed.
- TestEveryMapAProgramHoldsIsSorted passes. No IR Program field changes.
- TestRuntimeSetsAreEachUsed passes. No new runtime feature set or witness needed.
- TestNativeAgreementRowsNameRegisteredPrograms passes. All seven controls are registered by literal, unchanged from source.
- TestEveryTestIsParallelOrSaysWhy passes with main's existing baseline (7 current violations, 20 baseline entries). All 18 added tests call t.Parallel first; no baseline changed.
- TestEveryDuplicatedProductDeclarationAgrees passes. No product recipe or new compiler-consuming package is added, so compiler-dependencies.json remains unchanged. Its declared readers were also exercised; selection/results accompany this report.
- Quiet-tree traced native product test: 14 settled, 0 refused. Reader-only ADAMIC_BUILD_CACHE=read focused rows: PASS 1.282s. No row builds on this reader-only run.
- go vet ./internal/lower passes. Formatting and final source whitespace checks are recorded separately.

Mutants, rerun from the source report's patches through isolated Go overlays (production files were never mutated):

- every: remove only the receiver-union-layout guard. TestArrayNarrowingEveryRefused fails on acceptance, got <nil>.
- findLast: remove findLast from the result-representation guard. TestArrayNarrowingFindLastRefused fails on acceptance, got <nil>.
- literal: remove only the contextual union-literal guard. TestFX5NumberLiteralRefused fails on acceptance, got <nil>.
- registration: remove FX5MixedLiteral from nativeAgreementPrograms. TestNativeAgreementRowsNameRegisteredPrograms fails because its literal has no registered product.

Each exits 1 at the intended assertion, as mutants.json and mutant-*.log record. These are assertion proofs, not sanitizer or clang-warning claims. Mutant sources are .go.txt and exact patches are .patch.

The seven accepted controls (optional_number, optional_string, some_negated, find_last_index, preserved_union, boolean_callback, mixed-literal) separately match source Node byte for byte in generated JavaScript and native with ASan, UBSan and Linux LeakSanitizer enabled. Every run exits 0 with empty stderr. sanitizers.json records complete stdout, commands and wall times; the slowest standalone comparison, including its build, is 7.372s. No production runtime was edited.

The focused lower suite passes in 7.097s. Every added test's observed seconds:

'''+ '\n'.join('- '+r['test']+': '+str(r['seconds'])+'s.' for r in own)+f'''

Reader coverage: {len(readers)} Go import/test readers plus {len(declared)} additional readers declared by main's compiler dependency registry and {len(extras)} dynamic CLI/backend checks, {len(expected)} packages total. Each ran focused relevant tests or compile-only when it has no tests, never whole packages. consumer-selection.json, consumer-*-results.json and declared-reader-*.json give every command/result. Flow corpus coverage, shared setup and remainder checks pass; no new oracle fixture required full corpus expansion. Optional skips are listed in skips.json; they are not claimed as comparisons.

Focused validation commands (all stdout/stderr redirected directly to logs):

{commands}

Additional exact commands are in the runner scripts and result JSON. Counts command: timeout 360 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -v -timeout 300s -parallel 4 -args -update-counts. Product owner commands, the traced builder, read-only rows and corrected lane run are recorded in commands.txt. Every potentially long subprocess has a hard limit.

Setup succeeded. Linux, Go 1.27.1, clang 20.1.8, Node 24.19.0. nproc=5; cpu.max=400000 100000 (4 CPU quota). Build/test shells source /workspace/adamic-tools/env.sh, export GOPROXY=https://proxy.golang.org|direct and GOMAXPROCS=4; focused validation uses the local Go cache with ADAMIC_GOCACHE_OFF=1. Setup's shared-cache build itself succeeded. Timing lines:

```text
{setup}
```

Observed setup corrections and incomplete attempts:

- cmd/adamic's TestExplainChecksOutput initially lacked its sanitized runtime archive; the owner product was built. Its subprocesses also replace XDG_CACHE_HOME, hiding the default cache. Passing the supported ADAMIC_BUILD_CACHE_DIR explicitly made the original test pass (2.770s), without changing any test or product recipe. Initial failures are retained.
- Comments and regex option-dialect checks initially requested unsettled Go oracle owners. Their existing owner product tests were run; the original checks then pass. The admission reader likewise needed TestProduct_AdmissionCompiler before TestBuildRevisions.
- Stage3 lane's TestApplyPython initially lacked pinned npm packages, then the TypeScript gate checkout. npm ci --ignore-scripts --prefix stage3/api installed its unchanged lockfile. The public v6.0.3 checkout was cloned outside the repository and verified at 050880ce59e30b356b686bd3144efe24f875ebc8. With ADAMIC_TYPESCRIPT_SOURCE pointing to it, the unchanged test passes in 14.19s. No fixture data was invented or copied from cohere.
- The initial product trace refused settlement because the tree's status changed while concurrent evidence files were created. A quiet-tree run redirected output outside the repository and settled all 14 products; the subsequent read-only rows pass.
- The standalone sanitizer helper initially treated Node's experimental warning as program stderr, then lacked the adamic module loader for emitted JavaScript. It now uses the same oracle/node.mjs resolver and warning flag as the repository. Both incomplete attempts are retained and are not counted as passes.
- The initial counts timeout is described above. No table was hand-merged, no Mac table was used, and no test was removed or weakened.

No cohere code was copied. No ahra command, PR, full gate or Loom submission was run. Only this requested branch will be pushed, never main or another worker's branch. The exact base remains pinned even if origin/main subsequently moves.
'''
(out/'REPORT.md').write_text(text)
print(f'{len(expected)} reader packages covered; {len(own)} new tests; {len(skips)} skip events; maximum passing test time {max(r["seconds"] for r in leaves)}s')
