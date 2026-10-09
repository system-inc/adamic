import pathlib,json,re,statistics,subprocess,time,os,shutil
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/internal-lower-prototype';rows=json.loads((p/'requested-rows.json').read_text());menu=json.loads((p/'menu.json').read_text());prod=[m for m in menu if m['id']!='M04'];probes=json.loads((p/'probes.json').read_text())
def events(f):
 out=[]
 for l in (p/f).read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
def failures(f):return [x['Test'] for x in events(f) if x.get('Action')=='fail' and 'Test' in x and '/' not in x['Test']]
def seconds(f):return float(re.findall(r'\bok\s+\S+\s+(\d+(?:\.\d+)?)s',(p/f).read_text())[-1])
mat={m['id']:failures(m['id']+'.log') for m in menu};allrows=[x['Test'] for x in events('baseline.log') if x.get('Action') in ['pass','fail','skip'] and 'Test' in x and '/' not in x['Test']];(p/'matrix.json').write_text(json.dumps(mat,indent=2));(p/'matrix-rows.json').write_text(json.dumps(allrows,indent=2))
assert all(t in (p/'list.log').read_text().splitlines() for t in rows)
# Preserve unknowns on a genuinely aborted package run, using isolated row evidence.
bounded={};ks={t:[] for t in rows}
for m in prod:
 id=m['id'];es=events(id+'.log');term=[x for x in es if x.get('Action') in ['pass','fail','skip'] and 'Test' in x and '/' not in x['Test']];bounded[id]=len(term)!=len(allrows)
 for t in rows:
  af=p/(id+'-alone-'+t+'.log')
  if bounded[id] and af.exists():
   if failures(af.name) or any(x.get('Action')=='fail' and 'Test' not in x for x in events(af.name)):ks[t].append(id)
  elif t in mat[id]:ks[t].append(id)
(p/'bounded-mutants.json').write_text(json.dumps(bounded,indent=2));times={t:statistics.median(seconds('timing-'+t+'-'+str(i)+'.log') for i in range(1,4)) for t in rows}
# Time only outside rows needed to establish an observed subsumer.
outside=set()
for t in rows:
 if not ks[t] or any(mat[m]==[t] and not bounded[m] for m in ks[t]):continue
 candidates=[q for q in allrows if q!=t and all(q in mat[m] for m in ks[t])]
 if candidates and not any(q in times for q in candidates):outside.update(candidates)
runs=[]
for t in sorted(outside):
 for i in range(1,4):
  log='timing-'+t+'-'+str(i)+'.log';cmd=['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+t+'$'];start=time.monotonic()
  with (p/log).open('w') as f:q=subprocess.run(cmd,cwd=r,stdout=f,stderr=subprocess.STDOUT)
  runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));assert q.returncode==0
 times[t]=statistics.median(seconds('timing-'+t+'-'+str(i)+'.log') for i in range(1,4))
(p/'subsumer-runs.json').write_text(json.dumps(runs,indent=2));(p/'median-seconds.json').write_text(json.dumps(times,indent=2))
ors=[('Node enumerates Object.prototype; inherited-read refusal class and nil result are handwritten.',['external-run','self']),('Handwritten Refused class and nonempty repair text.','self'),('Node enumerates members; handwritten CheckError expectation from TypeScript checker runs. No lowering entry is reached.',['external-run','self']),('Handwritten NotYet class and reason substrings.','self'),('Handwritten NotYet class only; alternate NotYet reasons can pass.','self'),('Handwritten prototype-chain refusal and repair substring.','self'),('Handwritten Refused class and diagnostic substrings.','self'),('Both native C and JavaScript artifacts must equal our own operand-only output. Shared wrong output can compare equal.','self'),('Handwritten .a refusal text, one eager .ts assertion, and no delayed read checks in IR. Historical test name no longer describes its entire checker.','self'),('Handwritten acceptance of either NotYet or Refused, without checking the reason. M17 changes the first refusal reason while this row passes.','self'),('Node 24 computes RegExp.source for each input; lone-surrogate byte check is handwritten.',['external-run','self']),('Handwritten representation kind/known pairs and checker-assignability preconditions.','self'),('Handwritten Refused class, location, directive name, and repair text.','self'),('Handwritten absence of an error; it does not assert an IR answer exists.','self')]
out=[]
for i,t in enumerate(rows):
 k=ks[t];uq=[m for m in k if mat[m]==[t] and not bounded[m]];c=[q for q in allrows if q!=t and k and all(q in mat[m] for m in k)];sub=min([q for q in c if q in times],key=times.get) if any(q in times for q in c) else None
 v='cannot-judge' if i==2 else 'sacred' if uq else 'subsumed' if sub else 'overlapping' if k else 'untrue';mid=(uq or k or [None])[-1];file=next(f for f in ['prototype_test.go','proven_relations_test.go','readiness_test.go','regexp_test.go','representation_clock_source_test.go','suppression_directives_test.go'] if 'func '+t+'(' in (r/'internal/lower'/f).read_text());s=(r/'internal/lower'/file).read_text();ln=s[:s.index('func '+t+'(')].count('\n')+1;failure=None;log=None
 if mid:
  log=mid+'-alone-'+t+'.log' if bounded[mid] else mid+'.log';es=events(log);failure=next((x['Output'].strip() for x in es if (x.get('Test')==t or x.get('Test','').startswith(t+'/')) and re.match(r'\s*\w+_test.go:\d+:',x.get('Output',''))),None)
  if not failure:failure=next((x['Output'].strip() for x in es if x.get('Output','').startswith('--- FAIL: '+t)), 'package failed while running '+t)
 pp=[];own=[];passing=[]
 for pr in probes:
  if t in pr['rows']:
   id=pr['id'];fn=id+'-'+t+'.log';es=events(fn);bad=bool(failures(fn)) or any(x.get('Action')=='fail' and 'Test' not in x for x in es);own.append(bad)
   if bad:pp.append(id)
   else:passing.append(id)
 if i==3:pp.append('M04')
 cov=[]
 if v=='overlapping':
  remaining=set(k)
  for q in sorted([q for q in allrows if q!=t],key=lambda q:-len(set(k)&{m for m in k if q in mat[m]})):
   covered=remaining&{m for m in k if q in mat[m]}
   if covered:cov.append(q);remaining-=covered
   if not remaining:break
  assert not remaining
 d=dict(test=t,package='internal/lower',file='internal/lower/'+file+':'+str(ln),seconds=times[t],oracle=ors[i][0],oracle_kind=ors[i][1],kills=k,unique_kills=uq,last_proven_fail=(mid+': '+failure) if mid else None,verdict=v,subsumed_by=[sub] if sub else cov,mutants_in_matrix=19,probe_kills=pp,subsumer_seconds=times[sub] if sub else None,vacuous=None if not own else not any(own),bounded=any(bounded.values()),evidence=('ADAMIC_MUTANT='+mid+' ADAMIC_BUILD_CACHE_DIR=/tmp/u043/cache/'+mid+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > '+log+' 2>&1; '+failure) if mid else ('No lowering function reached; no checker/oracle mutation performed.' if i==2 else 'All 19 production matrix runs passed this row; see matrix.json and M*.log.'))
 if d['bounded']:d['matrix_rows']=rows
 if sub:d['subsumption_kills']=len(k)
 if i==11:d['vacuous_subcases']=['PRepresentation: source and lengthSource still return expected (0,false); objectSource, callableSource, arraySource and stringSource fail.']
 if i==13:d['evidence']+=' PLower-TestSuppressionDirectiveSoundNeighbors.log: passes with nil program and nil error.'
 if i==2:d['cannot_judge_reason']='The requested row exercises load.Load and the external checker rather than internal/lower. The frozen production menu changes lowering only; a meaningful lowering mutant cannot reach it.'
 out.append(d)
(p/'rows.json').write_text(json.dumps(out,indent=2));(p/'probe-matrix.json').write_text(json.dumps({pr['id']:{t:(pr['id'] in next(x for x in out if x['test']==t)['probe_kills']) for t in pr['rows']} for pr in probes},indent=2))
summary='\n'.join(['u043: all 14 requested rows remain present at origin/main 5deb11d433947539ab6448558b8ab1f1904393a5.','Clean whole-package baseline passed in 24.211 binary seconds; warm tools; nproc=5.',str(len(prod))+' production mutants plus M04 supplemental empty-hazard probe; whole-package matrices saved.','Verdicts: '+', '.join(str(sum(x['verdict']==v for x in out))+' '+v for v in ['sacred','subsumed','overlapping','untrue','cannot-judge'])+'.','Evidence pushed on test-audit/internal-lower-prototype under review/test-audit/internal-lower-prototype/.'])
report=summary+'\n\n```json\n'+json.dumps(out,indent=2)+'\n```\n\n| ID | Starting file:line | Change | Failed rows |\n| --- | --- | --- | --- |\n'
for m in menu:report+='| '+m['id']+(' supplemental probe' if m['id']=='M04' else '')+' | '+m['file']+':'+str(m['line'])+' | '+m['change']+' | '+', '.join(mat[m['id']])+' |\n'
report+='\nSurvivors: '+', '.join(m['id'] for m in prod if not mat[m['id']])+'\n'
report+='''
The CUT is Adamic lowering: inherited-member/property/element guards, prototype dispatch and hazard detection, representation mapping, relation proofs, eager assertion accounting, directive refusal, and RegExp source escaping. The oracle is Node for prototype membership and RegExp sources, and handwritten expected error classes, diagnostic text, IR facts, and own-artifact equality elsewhere. The reached-function inventory was generated before planting mutants from slice coverage. Every production standalone diff was vetted separately; the source switch was built once and uses per-mutant ADAMIC_BUILD_CACHE_DIR directories. No .a port or external checker was mutated.

The brief cites 8de93800f4 but requires fresh origin/main. This audit uses 5deb11d433947539ab6448558b8ab1f1904393a5, and all file:line references and diffs use that commit. None of the 14 rows moved or vanished. Their bodies have different assertions and are separate rows, not input-only wrappers sharing one family checker. No requested witness, setup-only check, or subprocess helper was found.

The whole baseline fit under 90 seconds, so production matrices were not narrowed just because internal/lower is named as a big package. Package-unique kills require one failed top-level row in a complete whole-package run. Kills outside these runs and repo-wide uniqueness remain unmeasured. Out-of-slice subsumers were timed when needed; choosing a subsumer is a small-matrix hint, not a deletion recommendation.

M04 was in the initial fixed change-constant menu, but its unconditional empty hazard answer is reported as a supplemental empty-answer probe. It is excluded from production kills, unique kills, and verdicts. Direct-entry probes are separate P*.diff files: Lower, property, elementAccess, representation, and escapeRegexSource. They ran callers separately to prevent a nil-program panic from hiding later rows. Probe panics are observed failures. No probe was planted in load.Load, because it is preparation for the lowering rows; the nullish row itself reaches only that checker path and has vacuity null.

The nullish row cannot be judged by a lowering mutant. Its time and baseline behavior are measured, but no mutation in the external checker was performed. The suppression sound-neighbor row accepts nil program/nil error, demonstrating vacuity. Representation's two unresolved-parameter checks pass its empty probe while its four positive representations fail. Most refusal rows use our own expected text or error classes. RegExpNativeRefusals accepts either Refused or NotYet without pinning the reason; the M17 witness demonstrates an alternate refusal passing this row. Erasure compares two outputs of the same lowerer, so shared wrong artifacts can agree.

Baseline skips outside the requested rows were TestOriginalCycleLedger (requires pinned pristine TypeScript 6.0.3 with generated diagnostics), TestOptionalWideningCensus (requires a caller-selected project config and output path), and one TestMixedUnionContractGraph subcase. No requested row skipped. These are corpus/inventory opt-ins rather than missing tools for requested rows; their corpus runs remain uncovered. stage3/api npm ci ran first and reported three packages in 357 ms. Tools were warm and cloud/setup.sh was not run. Fetch/list/baseline compilation wall timings were not separately captured.

The initial switch instrumentation had a syntax error in a guarded dropped block. This was corrected before the first matrix; initial-switch-error.log preserves that failed build preparation. The switch also selects M02's constant exactly as its standalone diff specifies. Production source is restored before final clean package run and vet. Evidence .diff files naturally contain whitespace-bearing context; whitespace diagnostics on them as newly added text are distinct from the restored source diff check.

The audit does not measure mutation adequacy for the entire compiler, error-message authority against ECMAScript, or repo-wide uniqueness. No other package's suite was run. The 20-change cap and supplemental-probe exclusion leave 19 production mutants, below three per row. Untrue means no fixed-menu production mutant made that row fail in this session, not that no possible mutant can ever fail it. Expected values without checked outside authority are classified self.
'''
mr=json.loads((p/'mutant-runs.json').read_text());cr=json.loads((p/'clean-runs.json').read_text());report+='\nMeasured timing: switch build '+str(next(x['wall'] for x in mr if x['log']=='switch-build.log'))+' seconds; standalone vet '+str(round(sum(x['wall'] for x in mr if '-vet.log' in x['log']),3))+' seconds; matrix/probe command wall total '+str(round(sum(x['wall'] for x in mr if '-vet.log' not in x['log'] and x['log'] not in ['switch-build.log','switch-gofmt.log']),3))+' seconds; clean timings/coverage command wall total '+str(round(sum(x['wall'] for x in cr),3))+' seconds. Individual rebuild/run timings are in mutant-runs.json; all three clean timings per requested row are retained.\n'
(p/'REPORT.md').write_text(report)
for name in ['clean','mutants','report']:
 path=pathlib.Path('/tmp/u043-'+name+'.py')
 if path.exists():shutil.copyfile(path,p/(name+'-script.py'))
print(summary,flush=True)
