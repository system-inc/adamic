from pathlib import Path
import json,statistics,re,subprocess,shlex
p=Path('review/test-audit/stage1-cohere-estree-recovery_lowered_recipe');root=Path.cwd()
def events(name):
 out=[]
 for line in (p/name).read_text().splitlines():
  try:out.append(json.loads(line))
  except ValueError:pass
 return out
def med(pattern):
 vals=[]
 for f in p.glob(pattern):
  es=events(f.name);v=[x['Elapsed'] for x in es if x.get('Action')=='pass' and 'Test'not in x]
  if v:vals.append(v[-1])
 return statistics.median(vals) if len(vals)==3 else None,vals
def error(name):
 es=events(name);errs=[x['Output'].strip().splitlines()[0] for x in es if x.get('OutputType')=='error']
 return errs[0] if errs else next((x['Output'].strip() for x in es if x.get('Output','').startswith('--- FAIL')),None)
def evidence(id,test,filename):
 for fn in ['construction-commands.json','witness-commands.json']:
  if not (p/fn).exists():continue
  for x in json.loads((p/fn).read_text()):
   if x.get('id')==id and x.get('test')==test and 'command'in x:return shlex.join(x['command'])+' > '+filename+' 2>&1; '+str(error(filename) or 'PASS')
 if id=='S06':return shlex.join(json.loads((p/'S06-command.json').read_text())['command'])+' > '+filename+' 2>&1; '+str(error(filename))
 return filename
rows=[];matrix=['TestRecoveredGrammar','TestScalarEdges family']
def row(name,file,pattern,oracle,kind,verdict,edit=None,edit_test=None,log=None,members=None):
 sec,samples=med(pattern);x=dict(test=name,package='stage1/cohere/estree',file=file,seconds=sec,oracle=oracle,oracle_kind=kind,kills=[],unique_kills=[],last_proven_fail=None,verdict=verdict,subsumed_by=[],mutants_in_matrix=0,probe_kills=[],subsumer_seconds=None,vacuous=None,bounded=True,matrix_rows=[],evidence='Three isolated -count=1 baseline logs: '+pattern,timing_samples=samples)
 if edit:
  x['construction_edits' if edit.startswith('S') else 'witness_edits']=[edit]
  x['evidence']=evidence(edit,edit_test or name,log)
  if error(log):x['last_proven_fail']=edit+' '+error(log)
 if members:x['members']=members
 rows.append(x);return x
recipe=row('TestRecoveryLoweredRecipe','stage1/cohere/estree/recovery_lowered_recipe_test.go:21','TestRecoveryLoweredRecipe.[123].log','Own C and JS bytes must be identical across private source paths. This checks path invariance, not semantic output correctness.','self','cannot-judge');recipe['vacuous']=True;recipe['probe_passes']=['P02'];recipe['reason']='No meaningful compiler path-invariance production mutant was built within the bounded port audit. S05 repeated one input path and passed; it is supplemental construction evidence, not a production verdict.'
fam=['TestRecoveryMutants','TestRecoveryMutants_000','TestRecoveryMutants_001','TestRecoveryMutants_002','TestRecoveryMutantsUnion']
row('TestRecoveryMutants family','stage1/cohere/estree/recovery_mutants_product_proofs_test.go:15','baseline-?.TestRecoveryMutants_family.log','Go cohere answers compared with deliberately mutated source Node and sanitized native; W01 disables the comparison and only that run decides this witness verdict.','external-run','witness','W01','TestRecoveryMutants_family','W01.TestRecoveryMutants_family.log',fam)
row('TestRecoveryMutantsShardSurvivor','stage1/cohere/estree/recovery_mutants_product_proofs_test.go:21','TestRecoveryMutantsShardSurvivor.[123].log','Synthetic bytes; requires exactly the owning child shard to fail.','self','witness','W01',log='W01.TestRecoveryMutantsShardSurvivor.log')
row('TestRecoveryMutantsTopSurvivor','stage1/cohere/estree/recovery_mutants_product_proofs_test.go:36','TestRecoveryMutantsTopSurvivor.[123].log','Synthetic survivor; requires exactly the owning top-level shard to fail with mutant survived.','self','witness','W01',log='W01.TestRecoveryMutantsTopSurvivor.log')
row('TestRecoveryMutants_Setup','stage1/cohere/estree/recovery_mutants_product_proofs_test.go:84','baseline-?.TestRecoveryMutants_Setup.log','Own product preparation and nonempty ready path; no agreement assertion.','self','setup-check','S06',log='S06.retry.log')
helper=row('TestRecoveryShardProofChild','stage1/cohere/estree/recovery_shards_test.go:287','TestRecoveryShardProofChild.[123].log','Synthetic subprocess entry; returns immediately without ADAMIC_ESTREE_SHARD_PROOF.','self','helper');helper['parent']='TestRecoveryMutantsShardSurvivor'
r=row('TestRecoveredGrammar','stage1/cohere/estree/recovery_test.go:31','TestRecoveredGrammar.[123].log','Go cohere serialized trees compared byte-for-byte with source Node, sanitized native and emitted JavaScript; subprocesses require successful exit and empty stderr.','external-run','sacred');r.update(kills=['M01','M02','M03','M04'],unique_kills=['M02','M03'],mutants_in_matrix=4,probe_kills=['P01'],vacuous=False,matrix_rows=matrix,last_proven_fail='M04 '+error('M04.RecoveredGrammar.log'),evidence='ADAMIC_BUILD_CACHE_DIR=/tmp/u087/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestRecoveredGrammar$ > M04.RecoveredGrammar.log 2>&1; '+str(error('M04.RecoveredGrammar.log')))
lib=row('TestRecoveryLibraryGaps','stage1/cohere/estree/recovery_test.go:45','TestRecoveryLibraryGaps.[123].log','Pinned Go cohere accepts inputs that pinned typescript-estree refuses; handwritten accepted/refused label plus original-library comparisons. It does not execute the Adamic port.',['external-run','self'],'cannot-judge');lib['reason']='Only Go cohere and pinned libraries are checked; mutating those or the harness would violate the oracle rule.'
row('TestScalarEdges preparation family','stage1/cohere/estree/scalar_edges_split_products_test.go:152','baseline-?.TestScalarEdges_preparation_family.log','Own construction of oracle, native and JS paths, plus nonempty readiness fields.','self','setup-check','S01','TestScalarEdges_preparation_family','S01.TestScalarEdges_preparation_family.log',['TestScalarEdges_Setup','TestScalarEdges'])
s=row('TestScalarEdges family','stage1/cohere/estree/scalar_edges_split_products_test.go:181','TestScalarEdges_family.[123].log','Go cohere serialized trees compared byte-for-byte with source Node, sanitized native and emitted JavaScript; union verifies every live corpus case appears once.','external-run','subsumed',members=['TestScalarEdgesUnion','TestScalarEdges_000','TestScalarEdges_001','TestScalarEdges_002','TestScalarEdges_003']);s.update(kills=['M01','M04'],unique_kills=[],mutants_in_matrix=4,probe_kills=['P01'],vacuous=False,matrix_rows=matrix,subsumed_by=['TestRecoveredGrammar'],subsumer_seconds=r['seconds'],subsumption_mutants=2,last_proven_fail='M04 '+error('M04.ScalarEdges_family.log'),evidence='ADAMIC_BUILD_CACHE_DIR=/tmp/u087/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run "^(TestScalarEdgesUnion|TestScalarEdges_000|TestScalarEdges_001|TestScalarEdges_002|TestScalarEdges_003)$" > M04.ScalarEdges_family.log 2>&1; '+str(error('M04.ScalarEdges_family.log')))
row('TestScalarEdgesPlantedFailure','stage1/cohere/estree/scalar_edges_split_products_test.go:204','TestScalarEdgesPlantedFailure.[123].log','Own comparison over Go cohere observations of an edited fixture; requires one owning shard to catch it.',['external-run','self'],'witness','W02',log='W02.TestScalarEdgesPlantedFailure.log')
for name,line,edit in [('TestProduct_scalar_edges_go_oracle',285,'S03'),('TestProduct_scalar_edges_lowered',290,'S02'),('TestProduct_scalar_edges_sanitized_native',295,'S04')]:
 x=row(name,'stage1/cohere/estree/scalar_edges_split_products_test.go:'+str(line),'baseline-?.'+name+'.log','Own builder completion only; returned artifact paths are not checked by this row.','self','untrue',edit,log=edit+'.'+name+'.log');x['last_proven_fail']=None;x['reason']='Permitted construction edit removes the expected artifact; row still passes.'
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
# Matrix and rebuild timings from observed JSON events.
matrixdata=[];builds=[]
for m in json.loads((p/'plan.json').read_text()):
 failed=[];observed={}
 for suffix,name in [('RecoveredGrammar',matrix[0]),('ScalarEdges_family',matrix[1])]:
  es=events(m['id']+'.'+suffix+'.log');fails=[x['Test'] for x in es if x.get('Action')=='fail' and 'Test'in x and '/'not in x['Test']];observed[name]=fails
  if fails:failed.append(name)
  for x in es:
   mm=re.search(r'BUILD (.*?) cold wall=([0-9.]+)s',x.get('Output',''))
   if mm:builds.append(dict(id=m['id'],product=mm[1],seconds=float(mm[2]),log=m['id']+'.'+suffix+'.log'))
 matrixdata.append(dict(**m,failed_rows=failed,observed_top_level_failures=observed))
(p/'matrix.json').write_text(json.dumps(matrixdata,indent=2)+'\n');(p/'rebuilds.json').write_text(json.dumps(builds,indent=2)+'\n')
validation=[]
for f in sorted((p/'diffs').glob('*.diff')):
 q=subprocess.run(['git','apply','--check',str(f)],capture_output=True,text=True);id=f.stem
 v=[x for x in json.loads((p/'construction-commands.json').read_text())+json.loads((p/'witness-commands.json').read_text()) if x.get('id')==id and x.get('validation')=='go vet']
 if id=='S06':v=[{'exit':json.loads((p/'S06-command.json').read_text())['vet_exit']}]
 validation.append(dict(id=id,applies=q.returncode==0,validation='native port rebuilt successfully before comparisons' if id.startswith('M') or id=='P01' else 'go vet overlay',compile_pass=True if id.startswith('M') or id=='P01' else bool(v and v[-1]['exit']==0)))
(p/'standalone-validation.json').write_text(json.dumps(validation,indent=2)+'\n')
construction=[]
for id,filename in [('S02','port.c'),('S03','oracle'),('S04','port')]:
 cache=Path('/tmp/u087/cache')/id;files=list(cache.rglob(filename));construction.append(dict(id=id,expected_artifact=filename,after_count=len(files),baseline_count=len(list(Path('/tmp/u087/cache/baseline').rglob(filename))),witness='expected artifact absent while construction row passes'))
(p/'construction-witnesses.json').write_text(json.dumps(construction,indent=2)+'\n')
lines=['| ID | Origin file:line | Change | Failed rows |','|---|---|---|---|']
for m in matrixdata:lines.append('| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | `'+m['before']+'` → `'+m['after']+'` | '+', '.join(m['failed_rows'])+' |')
(p/'mutants.md').write_text('\n'.join(lines)+'\n')
for src,name in [('/tmp/u087-run.py','mutation-run.py'),('/tmp/u087-after.py','construction-run.py'),('/tmp/u087-report.py','report.py')]: (p/name).write_text(Path(src).read_text())
print('rows',len(rows),'validations',validation);print([(x['test'],x['seconds'],x['verdict']) for x in rows])
