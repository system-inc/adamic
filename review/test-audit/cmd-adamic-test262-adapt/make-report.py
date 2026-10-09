from pathlib import Path
import subprocess,json,re,statistics,shutil,time
p=Path('review/test-audit/cmd-adamic-test262-adapt');rows=json.loads((p/'rows.json').read_text());statuses=json.loads((p/'matrix-status.json').read_text());times=json.loads((p/'timings.json').read_text());plan=json.loads((p/'plan.json').read_text());base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
for x in plan:x['changed_line']=x['line']+(1 if x['id'] in ['M8','M10','M11','M17'] else 0)
(p/'plan.json').write_text(json.dumps(plan,indent=2))
medians={r:statistics.median(x['seconds'] for x in times if x['test']==r) for r in rows};matrix=[];events={}
for s in statuses:
 es=[]
 for l in (p/(s['id']+'.log')).read_text().splitlines():
  try:es.append(json.loads(l))
  except ValueError:pass
 events[s['id']]=es
 fails=[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']]
 passes=[e['Test'] for e in es if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']]
 skips=[e['Test'] for e in es if e.get('Action')=='skip' and e.get('Test')]
 panic=any('panic:' in e.get('Output','') for e in es)
 assert not panic,(s['id'],'panic needs isolated reruns')
 matrix.append({'id':s['id'],'fail':fails,'pass':passes,'skip':skips,'binary_seconds':next((e['Elapsed'] for e in reversed(es) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None),'wall_seconds':s['seconds']})
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
def failure(mid,row):
 for e in events[mid]:
  name=e.get('Test','')
  if name==row or name.startswith(row+'/'):
   line=e.get('Output','').strip()
   if 'adapt_test.go:' in line:return name+': '+line
 return None
entry={r:'P1' for r in rows[:5]};entry.update(TestClassifyAdapt='P2',TestCrashPath='P3')
source=Path('cmd/adamic-test262/adapt_test.go').read_text();results=[]
for row in rows:
 kills=[m['id'] for m in matrix if m['id'].startswith('M') and row in m['fail']]
 unique=[m['id'] for m in matrix if m['id'] in kills and len(m['fail'])==1]
 subs=[other for other in rows if other!=row and kills and all(other in m['fail'] for m in matrix if m['id'] in kills)]
 sub=min(subs,key=lambda x:medians[x]) if subs else None
 verdict='sacred' if unique else ('untrue' if not kills else ('subsumed' if sub else 'overlapping'))
 last=kills[-1] if kills else None;probe=entry[row];probe_killed=next(row in m['fail'] for m in matrix if m['id']==probe)
 oracle='Self-written expected rewritten source and adaptation counts; no external execution or copied authority.'
 if row=='TestClassifyAdapt':oracle='Self-written program fragments and reported counts. M2 passes although a witness shows const value and no let value while reporting var-to-let=1; the expected transformation is incompletely checked.'
 if row=='TestCrashPath':oracle='Self-written exact path, colon-space and reason concatenation.'
 if row=='TestAdaptLeavesCheckoutText':oracle='Self-written unchanged class text and zero counts. M1 no-op passes. This row asserts no checkout-file IO and contains no rewritable construct; all 20 production mutants pass it, but its zero-value probe fails.'
 line=source[:source.index('func '+row+'(')].count('\n')+1
 evidence_id=last or probe;fail=failure(evidence_id,row)
 command='ADAMIC_MUTANT='+evidence_id+' timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > review/test-audit/cmd-adamic-test262-adapt/'+evidence_id+'.log 2>&1'
 results.append({'test':row,'package':'cmd/adamic-test262','file':'cmd/adamic-test262/adapt_test.go:'+str(line),'seconds':medians[row],'oracle':oracle,'oracle_kind':'self','kills':kills,'unique_kills':unique,'last_proven_fail':last+': '+failure(last,row) if last else None,'verdict':verdict,'subsumed_by':[sub] if verdict=='subsumed' else [],'mutants_in_matrix':20,'probe_kills':[probe] if probe_killed else [],'subsumer_seconds':medians[sub] if verdict=='subsumed' else None,'vacuous':not probe_killed,'bounded':False,'matrix_rows':[],'evidence':command+'; '+str(fail)+(' (probe only; no production failing line)' if not kills else '')})
(p/'results.json').write_text(json.dumps(results,indent=2))
setup=json.loads((p/'setup.json').read_text());vets=json.loads((p/'vet-status.json').read_text());summary={'base':base,'nproc':5,'setup':'skipped: warm env worked','npm_ci_seconds':next(s['seconds'] for s in setup if s['label']=='npm-ci'),'baseline_wall_seconds':next(s['seconds'] for s in setup if s['label']=='baseline'),'timing_21_runs_wall_seconds':sum(t['wall'] for t in times),'standalone_apply_vet_seconds':sum(v['seconds'] for v in vets),'matrix_wall_seconds':sum(s['seconds'] for s in statuses),'matrix_binary_seconds':sum(m['binary_seconds'] for m in matrix),'coverage_reach_seconds':(p/'reach.log').read_text(),'measurement_enabled_binary_seconds':9.626,'unique_verdict_counts':{v:sum(r['verdict']==v for r in results) for v in ['sacred','subsumed','untrue']}}
(p/'summary.json').write_text(json.dumps(summary,indent=2))
# Preserve scripts and all raw observations. The copied witness sources are instrumented production copies, not changes to the oracle.
shutil.copy('/tmp/u011-matrix.py',p/'replay.py');shutil.copy('/tmp/u011-report.py',p/'make-report.py')
report=f'''Unit u011, cmd/adamic-test262 adapter slice
Base origin/main {base}; all seven requested names remain in adapt_test.go and go test -list output. No family grouping, witness or helper classification applies to these seven. No requested row skipped.
CODE UNDER TEST: Go source adaptation and classification, plus crash path formatting. ORACLE: hand-written source strings, counts, fragments and crash-label formatting in adapt_test.go. Oracle kind self for every row. No Node, tsc, real test262 checkout or externally copied expected value decides these seven rows.
The clean whole-package baseline passed (51.483 binary seconds). Covered seven rows separately to list every reached function before writing mutants; reached-functions.txt, functions.txt and reach.out retain the evidence. compiler.init appears because it runs at package initialization; it was not mutated. All original production functions under the selected mutation sites appear in this coverage-backed list.
Twenty code-derived mutants were fixed before checking kills, spread across source entry, variable safety/scope, callback inference/type rendering, equality, throws, scanning, classification and crash formatting. Fixed menu only; no supplemental insertion mutant. One static selector compiled all alternatives; duplicate helper bodies are selector scaffolding, not distinct mutants. Every standalone diff omits the selector and changes only its production code. M1 is a meaningful no-op adapter returning original source, not an empty-answer probe. P1/P2/P3 return actual zero values and are separate.
Standalone diffs independently applied to the base and passed timeout 90 go vet ./cmd/adamic-test262/. Entry returns replace the entire function body in standalone diffs, avoiding unreachable code. In selector code they return conditionally at entry. No compiler internals, oracle, fixture, test, harness comparison or runtime C was mutated. Native cache-key instructions therefore did not apply. Source restoration was verified against the base.
Every mutant and probe ran the whole package with timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run ., with ADAMIC_MUTANT selecting the mode. Logs are *id*.log. Inactive selector control passed. No run timed out or panicked; every requested row completed in every run. No bounded narrowing was necessary.
Results: five sacred rows, StrictEq subsumed by ClassifyAdapt, LeavesCheckoutText untrue after 20 production mutants. StrictEq catches M1/M11/M12, all caught by ClassifyAdapt (median 0.009 vs StrictEq 0.008). This subsumption rests on three kills in a 20-mutant sample; it is a hint, not deletion evidence.
Survivors and one-line observed witnesses:
M15: function cb(x){{return x;}} [1].forEach(cb); becomes cb(x: number) clean, but stays cb(x) under M15.
M16: var x=1e+2; becomes let x=1e+2; clean, but stays var under M16.
M17: var x=1; class Box {{ value: number; }} stays var clean, but becomes let under M17.
See witness-control.json and witness-M15/M16/M17.json; corresponding Go run commands use copied instrumented production adapt.go and rewrite.go plus audit_selector.go and the observational main.go in witness-source/. These witnesses change no test or oracle.
Classification oracle-strength witness: classify-witness-control.json reports const_value=false, let_value=true, var_to_let_count=1; classify-witness-M2.json reports const_value=true, let_value=false, var_to_let_count=1. TestClassifyAdapt passes M2 in the actual matrix. Its positive assertions do not require literal let output despite counting var-to-let.
Probe ownership: P1 for the five direct adaptSource rows, P2 for TestClassifyAdapt's classify entry, P3 for TestCrashPath's withCrashPath entry. All fail their own probe, so every vacuous=false. Probe kills are excluded from all production kill/uniqueness/verdict calculations. LeavesCheckoutText's production last_proven_fail is null; P1 proves it can fail on an actual zero answer but does not rescue its untrue verdict under the stated rules.
Skip accounting: default whole runs skip only TestCompilerStartupMeasurement, outside this seven-row unit. Enabled ADAMIC_TEST262_MEASURE=1 separately; it passed in 9.626 binary seconds. measurement-functions.txt and measurement-reach.out show no mutated function is reached: only compiler.init, program, stripUseStrict, runCommandWithLimit, limitedBuffer.Write/String. Thus it is outside these mutants' reach; no requested SDK/corpus/opt-in row remains skipped. Subprocess helper entries return by design and are not counted as skips.
Where the brief was unclear, wrong or cost time:
1. The supplied historic revision differs from fresh origin/main. Followed the fresh-main instruction and verified all seven names; none moved or vanished.
2. The package is Go adapter code, not a stage1 port or compiler. Mutating a Node/test262 oracle would measure the wrong thing. The actual chosen functions and oracle were declared before mutation.
3. Self-written rewrites can embody language knowledge, but the tests neither run an outside implementation nor cite/copy an expected value from it. Calling them external-authority would lack evidence.
4. LeavesCheckoutText's name/comment promises checkout preservation, but it supplies only an in-memory class and checks returned text/counts. It does not create, read or check a checkout file. No-op adaptation passes; its class input has no declaration or expression eligible for these rewrites.
5. Empty adapted{{}} erases returned Source, whereas a no-op adapter returns Source unchanged. Those are different behaviors. The former probe fails Leaves; the latter production mutant passes. The required verdicts must not conflate them.
6. ClassifyAdapt accepts const while reporting var-to-let. Counts/fragments are a weaker oracle than an exact required transformation; the actual M2 witness demonstrates this limitation.
7. Whole-package correctness runs cost roughly 28 to 40 wall seconds each. Following the full-matrix rule for 20 mutants plus three probes used {summary['matrix_wall_seconds']:.3f} seconds including control. No run cooked, so the brief supplied no basis to narrow them. This was the main budget cost.
8. npm ci in stage3/api succeeded in {summary['npm_ci_seconds']:.3f} seconds, though the seven rows themselves do not use node_modules.
9. The package's opt-in startup measurement was outside the unit. It was separately enabled and its function coverage retained; default matrices still show its skip explicitly. It does not reach the mutation sites.
10. Go vet requires entry-return standalone diffs to remove unreachable original bodies. Whole-loop/statement deletions and direct body replacements compile independently; no verdict relies on a compilation failure.
11. Coverage percentages ending in zero were initially filtered incorrectly for the inventory display. The filter was corrected and the complete list saved before any mutant. Raw profile/function output remains authoritative.
12. Sacred here means package-unique for these observed production mutants, not repo-wide uniqueness. Other-package replay remains central work.
Timing: setup skipped, nproc 5. Clean baseline wall {summary['baseline_wall_seconds']:.3f} s. The 21 isolated timing runs used {summary['timing_21_runs_wall_seconds']:.3f} wall seconds; medians come only from each binary's ok line. Independent apply/vet validation {summary['standalone_apply_vet_seconds']:.3f} s. Full matrix wall {summary['matrix_wall_seconds']:.3f} s, binary sum {summary['matrix_binary_seconds']:.3f} s. No native rebuilds were required; Go compilation is included in wall-minus-binary overhead and was not separately isolated. See matrix-status.json and matrix.json for each run.
Not covered: other packages, repo-wide uniqueness, full semantic equivalence of all adapter transformations, a real test262 corpus run, and checkout-file IO. No external-authority claim was made.
Evidence only is committed under review/test-audit/cmd-adamic-test262-adapt and pushed to the requested branch. No main push or pull request.
'''
(p/'report.txt').write_text(report)
print(json.dumps(summary,indent=2));print([(r['test'],r['seconds'],r['kills'],r['unique_kills'],r['verdict'],r['last_proven_fail']) for r in results]);print('last mutant fails',[(m['id'],m['fail']) for m in matrix[-5:]])
assert len(matrix)==24
assert all(not r['vacuous'] for r in results)
