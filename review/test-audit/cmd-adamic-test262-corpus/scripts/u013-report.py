import pathlib,json,time,statistics,subprocess,os,re
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/cmd-adamic-test262-corpus';rows=json.load(open(E/'scope.json'));plans=json.load(open(E/'plan.json'));supp={'M06','M13'}
def events(file):
 out=[]
 for line in file.read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
def failures(file):return sorted({e['Test'] for e in events(file) if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']})
def failing_line(file,row):
 for e in events(file):
  if e.get('Test','').split('/')[0]==row and e.get('Action')=='output' and re.search(r'\w+_test.go:\d+:',e.get('Output','')):return e['Output'].strip()
 return None
while not (E/'witness.done').exists():time.sleep(1)
allrows=[e['Output'].strip() for e in events(E/'baseline.log') if e.get('Output','').startswith('Test')]
# plain -list is authoritative
allrows=[x for x in (E/'test-list.log').read_text().splitlines() if x.startswith('Test')]
med={r:statistics.median([next(e['Elapsed'] for e in reversed(events(E/f'timing-{r}-{i}.log')) if e.get('Action')=='pass' and not e.get('Test')) for i in range(3)]) for r in rows}
mat={p['id']:failures(E/(p['id']+'.log')) for p in plans};(E/'matrix.json').write_text(json.dumps(mat,indent=2))
probes=json.load(open(E/'probes.json'));pmat={p[0]:failures(E/(p[0]+'.log')) for p in probes};(E/'probe-matrix.json').write_text(json.dumps(pmat,indent=2))
# all package timings are already in each full-column log
external={'TestMiniRunner','TestLargeCompilerOutputIsComplete','TestEditCacheSeparation','TestWorkerLazyFallback','TestLoweringSourceEdit'}
oracles={
'TestClassifyCorpus':'self: handwritten .want classification and negative metadata; skip reasons checked by substring only',
'TestVerdictCorpus':'self: JSON executions and expected outcome labels; Node is not run by this row',
'TestNormalizeReason':'self: handwritten diagnostic grouping and panic labels; no tsc comparison',
'TestRewriteHarnessCalls':'self: handwritten strings and substring checks',
'TestFrontmatterShapes':'self: handwritten frontmatter fields and classification',
'TestMiniRunner':'Node runs and is compared with native execution; self-written aggregate counts and refusal label',
'TestLargeCompilerOutputIsComplete':'Node runs and is compared with native execution; self-written minimum C size and final newline',
'TestOutputOverflowIsReported':'self: 6 consumed bytes, abc retained, overflow flag',
'TestEditCacheSeparation':'Node runs and is compared with native execution; self-written cache hit invariants',
'TestNodeHarnessIdentity':'self: identity equality and inequality after fixture source edits',
'TestWorkerLazyFallback':'Node runs and is compared with native execution; fallback itself checks only exit zero and nonempty C, not equivalence of C',
'TestLoweringSourceEdit':'Node runs and is compared with native execution; self-written cache invalidation invariants',
'TestRunnerLocationHelper':'helper subprocess entry for TestRunnerLocationIdentity; default timing measures immediate return',
'TestRunnerLocationIdentity':'self: byte-for-byte equality of helper output after executable relocation',
'TestCompilerStartupMeasurement':'self: subprocess Adamic C must equal internal/native.C output from Adamic Load and Lower; startup loop has no performance threshold'}
result=[]
for r in rows:
 kills=[id for id,rs in mat.items() if r in rs and id not in supp];unique=[id for id in kills if len(mat[id])==1];others=[];subs=None
 if r=='TestRunnerLocationHelper':verdict='helper'
 elif unique:verdict='slow-worthy' if med[r]>60 else 'sacred'
 elif not kills:
  verdict='cannot-judge' if any(r in mat[id] for id in supp) else 'untrue'
 else:
  candidates=[q for q in allrows if q!=r and all(q in mat[id] for id in kills)]
  if candidates:
   q=min(candidates,key=lambda q:med.get(q,float('inf')))
   if q not in med:
    values=[]
    for i in range(3):
     log=E/f'timing-{q}-{i}.log'
     with log.open('w') as out:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','^'+q+'$'],cwd=R,stdout=out,stderr=subprocess.STDOUT)
     values.append(next(e['Elapsed'] for e in reversed(events(log)) if e.get('Action')=='pass' and not e.get('Test')))
    med[q]=statistics.median(values)
   verdict='subsumed';others=[q];subs=med[q]
  else:verdict='overlapping';others=sorted({q for id in kills for q in mat[id] if q!=r})
 own=[p[0] for p in probes if r in p[4]];pk=[id for id in own if r in pmat[id]];vac=None if not own else not pk
 last=kills[-1] if kills else None;fl=failing_line(E/(last+'.log'),r) if last else None
 file='performance_test.go' if r=='TestCompilerStartupMeasurement' else 'edit_test.go' if r in ['TestEditCacheSeparation','TestNodeHarnessIdentity','TestWorkerLazyFallback','TestLoweringSourceEdit','TestRunnerLocationHelper','TestRunnerLocationIdentity'] else 'corpus_test.go'
 o=dict(test=r,package='cmd/adamic-test262',file='cmd/adamic-test262/'+file,seconds=round(med[r],3),oracle=oracles[r],oracle_kind=['external-run','self'] if r in external else 'self',kills=kills,unique_kills=unique,last_proven_fail=(last+': '+fl) if fl else None,verdict=verdict,subsumed_by=others,mutants_in_matrix=15,probe_kills=pk,subsumer_seconds=subs,vacuous=vac,bounded=False,matrix_rows=[],evidence=('ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT='+last+' timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > '+last+'.log 2>&1; '+fl) if fl else 'No admissible mutant failure proved; see supplemental and probe evidence',supplemental_kills=[id for id in supp if r in mat[id]])
 if r=='TestNodeHarnessIdentity':o['limitation']='Only supplemental M13 caught this row. It cannot support a verdict under the fixed menu.';o['vacuous_subcases']=['compiler-only source edit preserves Node harness identity: the empty identity passes this first assertion; execution edit then fails']
 if r=='TestRunnerLocationHelper':o['parent']='TestRunnerLocationIdentity';o['evidence']='Three default runs return immediately; parent subprocess evidence is in TestRunnerLocationIdentity';o['mutants_in_matrix']=0
 result.append(o)
(E/'rows.json').write_text(json.dumps(result,indent=2));(E/'medians.json').write_text(json.dumps(med,indent=2))
# clean timed compile and standalone diff checks
start=time.monotonic()
with (E/'clean-build.log').open('w') as log:rc=subprocess.run(['timeout','90','go','test','-c','-o','/tmp/u013-clean.test','./cmd/adamic-test262/'],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode
(E/'clean-build-time.json').write_text(json.dumps(dict(seconds=time.monotonic()-start,exit=rc)))
with (E/'apply-check.log').open('w') as log:
 for p in plans:
  rc=subprocess.run(['git','apply','--check','--cached',str(E/'diffs'/(p['id']+'.diff'))],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode
  assert rc==0,p['id']
(E/'validation.json').write_text(json.dumps({'standalone_diffs':len(plans),'apply_against_starting_index':'passed','go_vet_each_standalone':'passed, matrix.done is written only after all successful validations','production_restored':True}))
(E/'report.done').write_text('done')
