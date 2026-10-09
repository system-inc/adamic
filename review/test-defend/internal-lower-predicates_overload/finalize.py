import pathlib,json,shutil,subprocess,datetime
root=pathlib.Path('/workspace/adamic'); tmp=pathlib.Path('/tmp/defend042'); out=root/'review/test-defend/internal-lower-predicates_overload'; prefix='github.com/system-inc/adamic/internal/lower'
for p in tmp.iterdir():
 if p.is_file() and p.suffix in ('.log','.cover','.txt','.sh','.py'):shutil.copy2(p,out/p.name)
plan=json.loads((out/'plan.json').read_text()); census=[x for x in (tmp/'list.log').read_text().splitlines() if x.startswith('Test')]; matrix={}
for p in out.glob('*-matrix.log'):
 es=[]
 for l in p.read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 outcomes={x['Test']:x['Action'] for x in es if x.get('Test') and '/'not in x['Test'] and x.get('Action')in ['pass','fail','skip']}
 matrix[p.name.split('-')[0]]=dict(rows_failed=sorted(k for k,v in outcomes.items() if v=='fail'),rows_passed=sorted(k for k,v in outcomes.items() if v=='pass'),rows_skipped=sorted(k for k,v in outcomes.items() if v=='skip'),rows_unknown=sorted(set(census)-outcomes.keys()),completed=len(outcomes),cooked='test timed out' in p.read_text(),failing_output=[x.get('Output','').strip() for x in es if x.get('Test') and (': ' in x.get('Output','')) and (x['Test'].split('/')[0] in [k for k,v in outcomes.items() if v=='fail'])],binary_seconds=next((x.get('Elapsed') for x in reversed(es) if x.get('Action')in ('pass','fail') and not x.get('Test')),None))
(out/'matrix.json').write_text(json.dumps(matrix,indent=2))
rows=[]
for test,subsumer in [('TestIndirectPredicateOverloadIsPending','TestPredicateOverloadCallback'),('TestConditionAssertionAdmission','TestPredicateBodyProof'),('TestEveryNeedsCallbackEffects','TestPredicateCallbackContracts')]:
 attempts=[]; unique=None;evidence=[]
 for m in plan:
  if m['row']!=test or m['id'] not in matrix:continue
  column=matrix[m['id']];attempts.append(dict(mutant=m['id'],file_line=m['file']+':'+str(m['line']),change=m['old']+' -> '+m['new'],rows_failed=column['rows_failed']))
  if column['rows_failed']==[test] and not column['rows_unknown'] and not column['cooked']:unique=m['id']+' '+m['file']+':'+str(m['line'])
  if test in column['rows_failed']:
   evidence.append('ADAMIC_BUILD_CACHE_DIR=/tmp/defend042/cache/'+m['id']+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; '+next((line for line in column['failing_output'] if 'predicates_' in line),'See log'))
 rows.append(dict(test=test,package='internal/lower',prior_verdict='subsumed',subsumed_by=[subsumer],defense='defended' if unique else ('not defended' if len(attempts)==3 and all(not matrix[a['mutant']]['cooked'] and not matrix[a['mutant']]['rows_unknown'] for a in attempts) else 'cannot-judge'),unique_mutant=unique,attempts=attempts,evidence='\n'.join(evidence) or 'See completed whole-package outcomes in matrix.json.'))
(out/'rows.json').write_text(json.dumps(rows,indent=2))
for p in out.glob('*.diff'):subprocess.run(['git','apply','--check',str(p)],cwd=root,check=True)
report='''# Predicate and overload test defense

Starting origin/main: d29d80ceb5d5d42d9b0ffb7b528a57a2272c76f4. Audit source: origin/test-audit/internal-lower-predicates_overload, copied report, limitations, rows, plan and inventory included. All three requested rows and named subsumers remain present. Current package census has 276 top-level rows, versus the audit's 239. No test, harness, Node oracle or fixture was modified. Production sources restored; all saved diffs apply to starting origin/main and passed go vet ./internal/lower/.

## Code under test and oracle

Production Adamic Go lowering is the code under test: predicateArguments/indirect-overload admission, prefix runtime expression lowering, predicateFlowProof.statement, predicateRefusal, and provePredicate unsupported-target fallback. The indirect row checks a handwritten expected diagnostic substring, not diagnostic type or complete text. The every row checks a handwritten refusal substring. These are self oracles. The condition row now runs a complete accepted source program against Node, plus native and backend agreement, and is external-run. Its old audit version only checked a refusal/admission seam; prior vacuity and subsumption evidence does not describe the new runtime assertions. The proof-summary subsumer uses handwritten summary expectations.

## Disk, setup and baseline

Initial df: /tmp 2.8GB free on an 8.8GB filesystem; /workspace 18GB free. Removed named earlier unit scratch directories and adamic-gate under /tmp only. Locked permission-test fixtures required restoring owner write/traverse bits before removal. Final initial-cleanup /tmp 8.8GB free. The requested 15GB /tmp free threshold is impossible with this mount capacity, but disk stayed above 8GB throughout these runs. Never removed repository or tools. Warm env.sh worked (Go 1.27.1); setup skipped; nproc=5. npm ci in stage3/api added three packages in 897ms. A first census command accidentally used the npm directory; corrected to repository root before the meaningful baseline. Clean whole package passed in 88.777 binary seconds with 274 passes and two top-level skips. Two opt-in project inventories skip: OriginalCycleLedger and OptionalWideningCensus. MixedUnionContractGraph also skips an unsupported array subcase. All requested rows ran and passed.

## Coverage and aimed differences

Six isolated coverprofile runs, coverpkg=./internal/lower, all passed. Exact commands, full profiles and coverage-differences.json retained. Indirect versus OverloadCallback: six exclusive blocks including predicates.go:481-483, the checked indirect overload rejection. Condition versus BodyProof: 1,228 exclusive blocks because complete Lower reaches emission-ready expression/function/control code, while BodyProof directly invokes provePredicate. Every versus CallbackContracts: 30 exclusive blocks including unsupported loop dispatch and nonprimitive target fallback. Exclusive coverage relative to one subsumer is a lead, not uniqueness against all tests. Whole-package mutation replay settles observed uniqueness.

D1 changes only the exclusive indirect-call diagnostic constant. This proves the row uniquely guards the capability-gap distinction; it does not prove execution correctness or a stronger diagnostic-type assertion. A1 restricts the existing logical-negation lowering condition so boxed union operands fall through to a capability gap. The condition program passes unknown to an assertion, so its production runtime negation fails; the proof-only row never lowers that expression. This is a production condition change, not an inserted test-specific name selector. Both columns completed every current top-level row; exact passed lists are in matrix.json.

E1 changes the default unsupported-control-flow result to an early empty-path success. Every still fails a subsequent proof boundary as expected, so its test passes; UnprovenPredicateReturnsAreRefused catches the changed loop diagnostic. E2 accepts the unsupported target fallback, aimed at the array claim. E3 inverts the failed-summary handoff in predicateRefusal. Outcomes are recorded in rows.json and matrix.json. These are admission faults, not empty-answer entry probes. Each mutant has a distinct compiler build cache. The plan was saved before observing its outcomes; later candidates for a row already uniquely defended were not executed.

## Limits and owner findings

Uniqueness is for the observed current package with the recorded opt-in skips, not repo-wide. No other package was run. None of these names is a cost row, and none is a Node/native executor twin. Unsupported-loop and array-target proof faults may also affect other predicate rows; that is precisely why the whole package was replayed. An attempted mutant that does not fail Every is not evidence of vacuity.

The current condition-admission row gained runtime agreement since the audit. The prior evidence is stale in that material respect. The indirect oracle checks only a diagnostic substring, so its unique defense has that limited strength. The callback-effects row requires a refusal for an effectful array callback; any other refusal with the same generic substring can pass for the wrong reason. This is a specificity limitation, not a promise of performance that lacks a threshold. Its name promises callback-effect protection, but its assertions do not independently distinguish the precise cause of refusal.

Tool and publication issues: the 15GB /tmp instruction exceeds filesystem capacity; read-only default sandbox requires approved write executions; shell Git publication needs credentials not present in the runtime. Earlier unit's oversized connector upload stalled, so publication will use compact evidence rather than a large binary payload. Full execution timings are in runs.json. All per-test coverage and native runtime logs go to files.
'''
report+='\nFinished '+datetime.datetime.now(datetime.timezone.utc).isoformat()+'. Results: '+', '.join(r['test']+' '+r['defense'] for r in rows)+'.\n'
(out/'report.md').write_text(report)
print(json.dumps(rows,indent=2))
