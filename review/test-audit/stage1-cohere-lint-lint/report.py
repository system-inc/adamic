import json,statistics,re
from pathlib import Path
E=Path(__file__).resolve().parent
runs=[json.loads(x) for x in (E/'runs.jsonl').read_text().splitlines()];R={r['label']:r for r in runs}
def events(label):
 out=[]
 for x in (E/(label+'.log')).read_text().splitlines():
  try:out.append(json.loads(x))
  except:pass
 return out
def group(t):
 if re.fullmatch('TestNodeTableIsLinkOnly_[0-9]{3}',t):return 'TestNodeTableIsLinkOnlyFamily'
 if re.fullmatch('TestOwnedWitnesses(_Setup|Union|_[0-9]{3})',t):return 'TestOwnedWitnesses'
 return t.split('/')[0]
prodrows=['TestThroughput','TestNodeTableIsLinkOnly','TestNodeTableIsLinkOnlyFamily','TestOwnedWitnessesAssignmentStable','TestOwnedWitnesses']
fast=['TestLegacyMutants','TestDecorationOptionMutant','TestCountGuardMutant','TestThroughput','TestNodeTableIsLinkOnlyUnionAndPlantedFailure','TestOwnedWitnessesAssignmentStable','TestOwnedWitnessesPlantedDisagreement']
matrix=[]
for m in json.loads((E/'production-plan.json').read_text())+json.loads((E/'witness-plan.json').read_text()):
 if m['id'] not in R:continue
 labels=[m['id']]+(['W4-Mutants-bounded'] if m['id']=='W4' else [])
 m['failed_tests']=sum([R[l]['failed_tests'] for l in labels],[]);m['failed_rows']=sorted(set(map(group,m['failed_tests'])));m['runs']=labels;matrix.append(m)
(E/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
configs=[('TestLegacyMutants','lint_test.go','Go cohere findings; negative mutant-agreement assertion','external-run','W1'),('TestDecorationOptionMutant','lint_test.go','Go cohere findings; negative decoration-mutant agreement assertion','external-run','W2'),('TestCountGuardMutant','lint_test.go','Go cohere ordinary output and count; negative count-only agreement assertion','external-run','W3'),('TestThroughput','lint_test.go','Go cohere count, when benchmark enabled; all session timing runs skipped','external-run',None),('TestMutants','lint_test.go','Go cohere findings; negative registered-rule mutant agreement assertions','external-run','W4'),('TestNodeTableIsLinkOnly','node_table_split_test.go','Self-authored corpus census and hash-shard union','self',None),('TestNodeTableIsLinkOnlyFamily','node_table_split_test.go','Same native port with plain versus junk-row option, no outside authority','self',None),('TestNodeTableIsLinkOnlyUnionAndPlantedFailure','node_table_split_test.go','Self-authored synthetic union corruption and disagreement','self','W5'),('TestOwnedWitnessesAssignmentStable','owned_witness_units_test.go','Self-authored stable hash assignment and empty corpus assertion','self',None),('TestOwnedWitnesses','owned_witness_units_test.go','Go cohere findings against Node, emitted JavaScript and native; self census checks','external-run',None),('TestOwnedWitnessesPlantedDisagreement','owned_witness_units_test.go','Self-authored synthetic native disagreement and exactly-one-shard failure','self','W6')]
rows=[]
for name,file,oracle,kind,wid in configs:
 prefix='warm-timing-' if name in ['TestNodeTableIsLinkOnlyFamily','TestOwnedWitnesses'] else 'timing-'
 ts=[R.get(prefix+name+'-'+str(i)) for i in [1,2,3]]
 printed=[next((float(m[1]) for e in events(r['label']) if (m:=re.match(r'ok\s+\S+\s+([0-9.]+)s',e.get('Output','')))),None) if r else None for r in ts]
 seconds=statistics.median(printed) if all(x is not None for x in printed) else None
 kills=[m['id'] for m in matrix if name in m['failed_rows'] and (m['id'].startswith('W') if wid else m['id'].startswith('M'))]
 unique=[m['id'] for m in matrix if m['id'].startswith('M') and m['failed_rows']==[name]] if not wid else []
 if wid:verdict='witness' if wid in kills else 'cannot-judge'
 elif name=='TestOwnedWitnesses':verdict='sacred' if unique else 'cannot-judge'
 elif name=='TestNodeTableIsLinkOnlyFamily':verdict='untrue' if len([m for m in matrix if m['id'].startswith('M')])==3 and not kills else 'cannot-judge'
 else:verdict='cannot-judge'
 last=wid if wid in kills else kills[-1] if kills else None;label='W4-Mutants-bounded' if last=='W4' else last;proof=None
 if label:
  for e in events(label):
   if group(e.get('Test',''))==name and re.search(r'\w+_test.go:\d+:',e.get('Output','')) and any(x in e['Output'] for x in ['survived','line ','caught','planted failure']):proof=e['Output'].strip().split('\n')[0];break
  if not proof:proof=next((e.get('Output','').strip() for e in events(label) if group(e.get('Test',''))==name and e.get('Output','').startswith('--- FAIL:')),None)
 vacuous=None
 if name in ['TestNodeTableIsLinkOnlyFamily','TestOwnedWitnesses'] and 'E1' in R:
  relevant=[t for t in R['E1']['passed_tests'] if group(t)==name];failed=[t for t in R['E1']['failed_tests'] if group(t)==name];vacuous=bool(relevant) and not failed
 elif wid and wid in kills:vacuous=False
 evidence=('ADAMIC_U110_WITNESS='+wid+'; ' if wid else 'selector='+str(last)+'; ')+' '.join(R[label]['command'])+' > '+label+'.log 2>&1; '+str(proof) if label else 'No proven failure; see timing and bounded matrix logs.'
 rows.append(dict(test=name,package='stage1/cohere/lint',file='stage1/cohere/lint/'+file,seconds=seconds,oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=unique,last_proven_fail=(last+': '+str(proof)) if last else None,verdict=verdict,subsumed_by=[],mutants_in_matrix=6 if wid else 3,vacuous=vacuous,bounded=True,matrix_rows=fast+(['TestMutants'] if wid=='W4' else []) if wid else prodrows,evidence=evidence))
(E/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps({'rows':rows,'matrix':[{k:m[k] for k in ['id','file','line','failed_rows']} for m in matrix],'recorded_wall':round(sum(r['wall_seconds'] for r in runs),3)},indent=2))
