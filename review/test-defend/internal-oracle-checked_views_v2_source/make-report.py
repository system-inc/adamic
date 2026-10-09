import pathlib,json,shlex,subprocess
p=pathlib.Path('review/test-defend/internal-oracle-checked_views_v2_source'); r=json.loads((p/'results.json').read_text()); plan=json.loads((p/'matrix-plan.json').read_text()); scope=json.loads((p/'scope.json').read_text()); exclusive=json.loads((p/'exclusive-coverage.json').read_text()); target='TestClosureMergeRefusals'
mut=[x for x in r if x['label'].startswith('D1-batch')]; clean=[x for x in r if x['label'].startswith('clean-batch')]
failed=sorted({n for x in mut for n in x['failed'] if '/' not in n}); passed=sorted({n for x in mut for n in x['passed'] if '/' not in n}); skipped=sorted({n for x in mut for n in x['skipped'] if '/' not in n}); assert failed==[target] and len(passed)==189 and len(skipped)==7
assert all(x['exit']==0 for x in clean);assert all(not x['timeout'] for x in mut);assert set(scope['added'])<=set(passed)
(p/'passed-rows.json').write_text(json.dumps({'top_level_passed':passed,'top_level_failed':failed,'top_level_skipped':skipped,'observed_passed_subcases':sorted({n for x in mut for n in x['passed'] if '/' in n}),'observed_skipped_subcases':sorted({n for x in mut for n in x['skipped'] if '/' in n}),'unknown':'TestCountsAreRecorded; unselected TestNativeAgreesWithNode fixtures; opted-out WASI and missing acceptance fixtures; external packages'},indent=2)+'\n')
confirmation=json.loads((p/'confirmation-command.json').read_text()); events=[json.loads(l) for l in (p/'D1-confirmation.log').read_text().splitlines() if l.startswith('{')]; fails=[x.get('Test') for x in events if x['Action']=='fail' and x.get('Test')];assert set(n for n in fails if '/' not in n)=={target}
line=next(x['Output'].strip() for x in events if 'diagnostic: got' in x.get('Output',''))
command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-closure/cache/D1 '+shlex.join(confirmation['command'])
row={'test':target,'package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':['TestCheckedViewUntaggedArrayPending'],'defense':'defended','unique_mutant':'D1 internal/lower/locals.go:14','attempts':[{'mutant':'D1','file_line':'internal/lower/locals.go:14','change':'Change the production var-refusal What constant from "var" to "let"; retain Refused type, source location and repair.','rows_failed':failed}],'evidence':command+'; '+line,'bounded':True,'matrix_rows':plan['top_level_rows'],'passed_rows_file':'passed-rows.json','skipped':skipped,'corpus_subset':plan['corpus_members']}
(p/'rows.json').write_text(json.dumps([row],indent=2)+'\n')
for i,pattern in enumerate(plan['batch_patterns']):(p/('matrix-batch-'+str(i+1)+'.regex')).write_text(pattern+'\n')
base=(p/'base.txt').read_text().strip(); baseline=[json.loads(l) for l in (p/'baseline.log').read_text().splitlines() if l.startswith('{')]; elapsed=baseline[-1]['Elapsed']; assert not any(x['Action']=='fail' and x.get('Test') for x in baseline)
assert any('panic: test timed out' in x.get('Output','') for x in baseline)
text=f'''Defended in the bounded matrix: D1 fails only TestClosureMergeRefusals.
189 other top-level rows pass; seven skip; 12 var-containing corpus fixtures pass.
Keep the diagnostic guard. Complete package and repo-wide uniqueness remain unknown.

# Closure refusal defense

Starting origin/main: {base}. Audit base c0a7667baadaf161d6bb9066e0838b32774caa6c. Evidence branch test-defend/internal-oracle-checked_views_v2_source. Read CLAUDE.md, full fetched audit REPORT.md and row reports, target and subsumer test files, locals.variables, Lower/refuse orchestration, diagnostic formatting, optional widening refusal, and fixture sources/pins. README.md, docs/0.1.md and docs/memory.md are unchanged from the prior warm unit. Audit copies are preserved. Full refspec fetch created the named remote-tracking ref rather than only FETCH_HEAD.

CODE UNDER TEST: Adamic lowering, especially lowering.variables and the returned Refused diagnostic. ORACLE: unchanged source executed by Node, plus self-written .refused snapshots for Adamic diagnostic type/location/construct/repair. Node checks exit zero and empty stderr; this row does not compare source stdout. These are external-run and self oracles, not an externally validated diagnostic text. No oracle, fixture, snapshot, harness or test was edited.

## Baseline and current scope

First df /tmp /workspace: 8.6G and 14G free respectively. Removed only previous-unit /tmp/defend-stack and /tmp/adamic-gate, then recreated traversable TMPDIR. Second df: /tmp 8.6G free; /workspace unchanged. The /tmp mount totals only 8.8G, so the requested 15G cannot fit there. No full-disk error occurred. DISK.md and disk-final.log retain the figures.

Warm /workspace/adamic-tools/env.sh worked; setup skipped. nproc=5. npm ci in stage3/api completed before baseline. Optional native splitting remained unset; no setup-mode regression was introduced. Current list has {len(scope['current'])} top-level tests, versus {len(scope['audit'])} at audit; none vanished. Added and replayed: {', '.join(scope['added'])}.

The clean whole-package binary exceeded its 90-second budget at {elapsed}s. It had no observed failing test before timeout. No mutant was planted then. Clean bounded matrix ran every standalone top-level test except the corpus and counted-corpus aggregators, plus 12 corpus fixtures whose source contains a conservative lexical var token. All five clean batches passed, with seven top-level skips listed below. This broad matrix avoids treating newly added or unrelated standalone tests as unseen. No clean narrowed batch was red.

## Coverage and aimed distinction

Per-test commands used timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./internal/lower -coverprofile=<closure.cover or array.cover> ./internal/oracle/ -run '<anchored name>' with ADAMIC_GATE_UNCACHED=1 and a coverage cache. Both passed: closure binary 2.614s; array binary 0.173s. Profiles and function inventories are saved.

The closure row has {len(exclusive['closure_only'])} positive blocks not covered by its reported subsumer; the reverse has {len(exclusive['array_only'])}. The decisive target-only block is internal/lower/locals.go:14.3,15.1, constructing the non-block-scoped var refusal. The target also exclusively reaches the optional-widening refusal return. The subsumer loads array-union fixtures, asserts NotYet for unavailable element-kind/never-array metadata, then skips execution. It does not assert this var diagnostic.

The target's scanner fixture has a real var declaration at 69:5. Its twelve other inputs pin absent errno fields in an Error-to-ErrnoException structural relation. Its oracle asserts the complete diagnostic, while the audit's sole catch was a broad condition flip that refused const/let earlier. The aimed distinction is the forbidden construct name, not a file-path-specific mutation.

## Mutant and replay

D1 changes just the production diagnostic constant at starting-main internal/lower/locals.go:14: What: "var" becomes What: "let". Refusal type, condition, location and repair remain unchanged. This is the permitted constant-change menu item and changes real compiler output for the existing scanner program. It introduces no test-triggered branch or supplemental insertion. D1.diff is standalone against starting main and applies cleanly. go vet ./internal/lower/ passed with D1; logs and command are saved. No native rebuild is needed for this refusal path, which stops before IR/code generation. Broader matrix native build work is included in wall timings.

Only TestClosureMergeRefusals fails, on its scanner subcase. Exact line:

{line}

The short isolated confirmation command, executed after the broad matrix:

{command} > D1-confirmation.log 2>&1

The same failure was observed in D1-batch-1.log. D1-batch-2 through D1-batch-5 pass. The former subsumer passes all its admission assertions before its five known execution skips. All six added top-level tests pass. 189 other top-level rows pass, seven skip; every planned row has an observed terminal event. The 12 selected TestNativeAgreesWithNode corpus inputs all pass. Exact passed rows and subcases: passed-rows.json. Exact patterns and commands: matrix-plan.json, matrix-batch-N.regex and results.json. Per-mutant ADAMIC_BUILD_CACHE_DIR=/tmp/defend-closure/cache/D1 and ADAMIC_GATE_UNCACHED=1 were set throughout.

Verdict: defended, bounded. The skipped tests, count-table aggregator, unsampled corpus fixtures and external packages remain unknown, so this is not a complete-package uniqueness claim. The lexical source scan is a conservative lead, not a proof of transitive imported-source reachability. No matrix panic aborted later tests, and no narrowed matrix exceeded the budget. D2/D3 were fixed in plan.json before results but not executed because D1 established a distinct catch. There are no executed survivors or empty-answer/harness probes.

Production source was restored in finally blocks after both executions. The final clean target/subsumer run passes. git diff --exit-code -- internal and git apply --check D1.diff pass. No test deletion, rewrite or weakening, production commit, main push or pull request.

## Brief ambiguities, costs and findings

* The 15 GB threshold is unattainable on /tmp's entire 8.8 GB mount; /workspace has 14G free after authorized /tmp cleanup. Logs show no disk error.
* The audit's rows.json is only a name list. The actual row findings are in rows-report.json, also embedded in REPORT.md. All were fetched and retained.
* The required current origin/main differs from the audit base. The target now contains 13 refusal witnesses. Six new package tests were included; none vanished.
* A whole-package run cooks at 90 seconds. The replay was broadened to 196 standalone tests and a source-selected 12-fixture corpus subset, split across five batches. TestCountsAreRecorded and remaining corpus inputs were omitted and are unknown, not passes. The count-table test cannot be meaningfully sampled without making its final table incomplete.
* Seven top-level rows skipped: {', '.join(skipped)}. Entries acceptance awaits missing acceptance fixtures; Stage3FixtureHook is a subprocess entry; five WASI rows remain opt-in. None contributes a passing uniqueness comparison. Additional review/array subcase skips are all listed in passed-rows.json. No opt-in WASI toolchain installation was attempted.
* Coverage is exclusive only relative to the reported subsumer, not the whole package. D1's observed matrix, rather than exclusive coverage alone, establishes the distinct catch.
* This is a diagnostic mutation, not an executable miscompile. That is appropriate to a refusal row that explicitly pins construct and repair. No Node stdout agreement or compiled closure behavior is claimed.
* The brief permits at most three attempts; three honest failures are needed for not defended. A distinct first catch makes further aimed mutations unnecessary. This row is neither an executor twin nor a cost row.
* Owner finding: the assertions check refusal receipts and source exit/stderr. They do not demonstrate native closure merging or capture semantics. The scanner's first var refusal prevents reaching much of its closure implementation; filesystem inputs stop at optional-widening refusals. Keeping these diagnostic guards does not establish those later behaviors.
* The npm install's separate wall time was not instrumented. Native builds and execution are combined in recorded command times, not claimed as separate compiler-only measurements.

Timing: warm setup skipped; whole baseline {elapsed}s binary. Clean bounded command walls total {sum(x['wall_seconds'] for x in clean):.3f}s; D1 bounded matrix {sum(x['wall_seconds'] for x in mut):.3f}s; D1 vet {next(x['wall_seconds'] for x in r if x['label']=='D1-vet'):.3f}s; short confirmation {confirmation['wall_seconds']:.3f}s; restored clean command {next(x['wall_seconds'] for x in r if x['label']=='restored'):.3f}s. Individual command and binary timings are in results.json.
'''
assert '\u2014' not in text
(p/'REPORT.md').write_text(text)
print('failed',failed,'passed',len(passed),'skipped',len(skipped),'coverage exclusive',len(exclusive['closure_only']),len(exclusive['array_only']))
