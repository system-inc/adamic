import pathlib,json,re,statistics,subprocess
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-oracle-checked_views_flag_downcast';rows=json.loads((e/'requested-rows.json').read_text());wit=json.loads((e/'witness-rows.json').read_text());prod=json.loads((e/'production-rows.json').read_text())
def events(name):
 out=[]
 for l in (e/name).read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
def fails(name):return [x['Test'] for x in events(name) if x.get('Action')=='fail' and x.get('Test') in rows]
def evidence(log,row):
 ev=events(log);lines=[x.get('Output','').strip() for x in ev if x.get('Test','').startswith(row) and x.get('Action')=='output'];bad=[s for s in lines if '_test.go:' in s and 'gate cache' not in s and ' caught' not in s and any(k in s for k in ['surviv','differ','diagnostic','expected','not yet','NotYet','panic:'])] or [s for s in lines if 'FAIL:' in s];return next((s for s in bad if 'FAIL:' not in s),next((s for s in bad),'PASS observed'))
raw={m:fails(m+'.log') for m in ['M1','M2','M3','M4']};matrix={m:[x for x in f if x in prod] for m,f in raw.items()};(e/'matrix.json').write_text(json.dumps({'rows':rows,'raw_failures':raw,'production_failures':matrix,'outside_rows':'unknown','witness_failures_excluded':wit},indent=2))
seconds={}
for row in rows:
 vals=[]
 for i in range(1,4):
  vals.append(float(re.search(r'([0-9.]+)s', (e/f'timing-{row}-{i}.log').read_text()).group(1)))
 seconds[row]=statistics.median(vals)
subs={rows[3]:[rows[12]],rows[4]:[rows[11]],rows[11]:[rows[4]],rows[12]:[rows[3],rows[11]]};report=[]
files=list((r/'internal/oracle').glob('*_test.go'))
for row in rows:
 f=next(p for p in files if re.search(r'func '+row+r'\(',p.read_text()));line=next(i for i,l in enumerate(f.read_text().splitlines(),1) if 'func '+row+'(' in l)
 kills=[m for m,fs in matrix.items() if row in fs];w=row in wit;wf=row in fails('W1.log')
 verdict=('witness' if wf else 'untrue') if w else ('overlapping' if len(subs.get(row,[]))>1 else 'subsumed' if row in subs else 'untrue')
 log='W1.log' if w else (kills[-1]+'.log' if kills else 'M1.log');id='W1' if w else (kills[-1] if kills else None)
 oracle='Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison' 
 kind=['external-run','self']
 if row==rows[7]:oracle='Handwritten physical representation outputs and sanitizer expectations; no external reference run';kind='self'
 if row in [rows[4],rows[11]]:oracle='Node source execution compared with native and JavaScript outputs, plus sanitizer/leak checks';kind='external-run'
 if row==rows[5]:oracle='Node success exit only; handwritten views-v3 refusal substring. The source check can pass for a wrong successful output.'
 probes=[] if w else ['PLower']+([] if row==rows[5] else ['PC','PJS'])
 cmd='ADAMIC_MUTANT='+str(id)+' ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/'+str(id)+' timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '+repr('^('+'|'.join(wit if w else rows)+')$')
 obj=dict(test=row,package='internal/oracle',file=str(f.relative_to(r))+':'+str(line),seconds=seconds[row],oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=[],last_proven_fail=(id+': '+evidence(log,row)) if (wf if w else kills) else None,verdict=verdict,subsumed_by=subs.get(row,[]),mutants_in_matrix=4,probe_kills=probes,subsumer_seconds=seconds[subs[row][0]] if verdict=='subsumed' else None,vacuous=None if w else False,bounded=True,matrix_rows=rows,evidence=cmd+' > '+log+' 2>&1; '+evidence(log,row),witness=w)
 if row in [rows[3],rows[12]]:obj['vacuous_subcases']='PC/PJS preserve lowering-refusal cases which do not reach those entries; no positive executed case survived its Lower probe.'
 report.append(obj)
(e/'rows.json').write_text(json.dumps(report,indent=2))
clean=json.loads((e/'clean-runs.json').read_text());audit=json.loads((e/'audit-runs.json').read_text());timing={'clean_steps':clean,'audit_steps':audit,'isolated_timing_wall_total':sum(x['wall'] for x in clean if x['log'].startswith('timing-')),'audit_wall_total':sum(x['wall'] for x in audit)};(e/'timings.json').write_text(json.dumps(timing,indent=2))
summary='''u057 audited all 13 requested rows at 68db8ddd145281a62655452496bdc32ef848bdf3; none moved or vanished.
Whole-package baseline timed out at 90 seconds; clean bounded baseline and final rerun passed.
Four vetted production mutants: two caught and two survived the bounded matrix; no unique production kills.
Six witnesses proved; two witness rows and one production row are untrue under this experiment.
Five production rows failed their empty-entry probes; evidence is on the requested audit branch.
'''
menu=json.loads((e/'menu.json').read_text());table='| ID | Origin file:line | Change | Observed failing rows |\n|---|---|---|---|\n'
for m in menu:table+='| '+m['id']+' | '+m['file']+':'+','.join(map(str,m['lines']))+' | '+m['change']+' | '+(', '.join(raw[m['id']]) or 'none')+' |\n'
text=summary+'\n```json\n'+json.dumps(report,indent=2)+'\n```\n\n'+table+'''
Raw failures above include witness precondition failures. matrix.json separately records the five production rows. W1 disables disagreement at internal/oracle/oracle_test.go:716. Six witnesses fail; tuple-identity and callable-producer still pass, so their verdict is untrue. Witness production failures never count as kills.

Survivors: see observation-clean.log, observation-M1.log and observation-M4.log and the saved observation source. M1 changes the helper's unsupported-demand result on synthetic IR; the public array fixture is refused earlier. M4 drops ordinary literal restriction checking; the independent admitted source changes a field after downcast. These observations are supplemental witnesses, not additional mutants or unit rows.

Brief ambiguities and costs:
- The supplied file commit is older than current origin/main. Scope was verified using the actual list at the recorded starting commit.
- Three mixed top-level rows contain both control runs and built-in mutants. They are classified as witnesses as whole rows, so production failures cannot decide their verdicts.
- The compiler instruction limits rebuilds to four mutants, while three per row would require 39. Four spread mutations were selected before failures were inspected.
- M2 is a diagnostic constant change in both canonical paths, not a behavioral runtime change. Subsumption is a hint over one observed mutant each, not a deletion recommendation.
- The package baseline exceeded its test-binary budget. All matrices ran exactly the 13 listed rows; package and repository uniqueness remain unknown.
- ObjectPrimitiveSource's comment-boolean, comment-good and comment-flags-wrong subcases hard-skip pending views-v3. No requested top-level row skipped. Outside-slice WASI and opt-in skips observed in baseline are recorded in baseline.log; that interrupted run is not a complete skip inventory.
- Coverage lists reached Go CUT functions. Runtime C calls were not traced. No runtime C or TypeScript oracle edits were used.
- Lower's empty-program probe can panic in emitters. Each row was isolated, so later rows were not falsely recorded as failures from a prior panic. Probe failures do not establish useful production discrimination.
- PC/PJS do not affect negative subcases already refused by lowering. They are not evidence of an executed positive case accepting empty output.
- All handwritten refusal and panic strings are self oracles, even when the same row also runs Node. The array boundary source checks only success exit, a weaker pin than full output.
- The two surviving witnesses need their own assertion that the comparison rejected a counterfactual. Their passing W1 runs are concrete evidence, not a production-mutant verdict.
- No other package suite or full repository gate was run. Native products used distinct caches for each compiler mutant. Cold matrix times include rebuild and test execution; exclusive build time is not separately measurable from these logs.

Timing: tools were warm, setup.sh skipped; nproc 5. npm-ci.log records dependency installation. timings.json records command wall durations, isolated own-binary medians are in rows.json, and each native rebuild matrix's wall time is in audit-runs.json. The whole-package baseline cooked at 90.091 seconds; bounded baseline's own line was 7.19 seconds. All standalone production and probe diffs and W1 passed go vet; final bounded baseline and final vet passed after restoration.
'''
(e/'REPORT.md').write_text(text)
for p in ['/tmp/u057-clean.py','/tmp/u057-audit.py','/tmp/u057-observe.py','/tmp/u057-report.py']:(e/pathlib.Path(p).name).write_text(pathlib.Path(p).read_text())
print(json.dumps({'seconds':seconds,'raw':raw,'witness_fails':fails('W1.log'),'isolated_wall':timing['isolated_timing_wall_total'],'audit_wall':timing['audit_wall_total']}))
