import pathlib,json,statistics,subprocess,os,time,gzip,difflib
P=pathlib.Path(__file__).parent
rows=json.loads((P/'rows.json').read_text()); times=json.loads((P/'times.json').read_text()); runs=json.loads((P/'matrix-runs.json').read_text()); mutants=json.loads((P/'mutants.json').read_text())
def events(log):
 result=[]
 for line in (P/'logs'/log).read_text().splitlines():
  try: result.append(json.loads(line))
  except ValueError: pass
 return result
matrix={}
for id,r in runs.items():
 matrix[id]=sorted({e['Test'].split('/')[0] for e in events(r['log']) if e['Action']=='fail' and 'Test' in e})
 assert all(any(e.get('Test')==row['test'] and e['Action'] in ('pass','fail','skip') for e in events(r['log'])) for row in rows), 'incomplete unit matrix'
 if 'individual' in r:
  assert all(any(e.get('Test')==name and e['Action'] in ('pass','fail','skip') for e in events(run['log'])) for name,run in r['individual'].items()), 'incomplete individual rerun'
(P/'kill-matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
killsets={name:{id for id,names in matrix.items() if name in names} for names in matrix.values() for name in names}
# Choose a known timed unit subsumer when available; otherwise one named package row.
subsumers={}
for row in rows:
 name=row['test']; kills=killsets.get(name,set()); unique=[id for id in kills if len(matrix[id])==1]
 if not kills or unique: continue
 candidates=[other for other,k in killsets.items() if other!=name and kills<=k]
 if candidates:
  subsumers[name]=min(candidates,key=lambda x:(x not in times,statistics.median(times[x]) if x in times else 0,x))
for name in sorted(set(subsumers.values())-times.keys()):
 vals=[]
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+name+'$']
  log='subsumer-'+name+'-'+str(i+1)+'.log'
  with (P/'logs'/log).open('w') as out: r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
  assert r.returncode==0,name
  vals.append(next(e['Elapsed'] for e in events(log) if e['Action']=='pass' and 'Test' not in e))
 times[name]=vals
(P/'all-times.json').write_text(json.dumps(times,indent=2)+'\n')
probes=json.loads((P/'probe-runs.json').read_text())
oracles={
'TestArgumentsLengthRefusals':'Self-pinned Refused type and exact diagnostic suffix.',
'TestArgumentsLengthReadNeighbors':'Self: only absence of lowering error; ignores returned IR.',
'TestArgumentsLengthRefusalFixtures':'Self-pinned Refused type and exact diagnostic suffix for .a and .ts.',
'TestArrayPredicatePreservesDeclaredElementContract':'Self: only absence of lowering error; ignores returned IR.',
'TestArrayPredicateCannotInventAnElementContract':'Self-pinned Refused type and return-is-not-proven substring.',
'TestArrayPredicateDoesNotMisclassifyNativeTuples':'Self-pinned NotYet type and tuple-or-erased substring.',
'TestUnknownArrayPredicateRefusesUnrepresentedObservations':'Self: any NotYet suffices; M06 changes the failure reason without failing this row.',
'TestArrayPredicateCoexistsWithUnknownReflection':'Self: only absence of lowering error; ignores returned IR.',
'TestUncheckableCastsStayRefused':'Self-pinned Refused and repair substrings; does not inspect runtime behavior.',
'TestCheckedCastProofAndElision':'Self: IR check counts only; accepts M10 wrong numeric tag (0 becomes 1).',
'TestCensusMarkerResultIsAssignable':'Self: any Refused suffices; does not pin result-relation reason.',
'TestCensusMarkerZeroCallIsNotAssumedSafe':'Self-pinned NotYet and erased-never-rest-call substring.',
'TestOptionalFunctionValueRelation':'Self-pinned Refused and optional-number parameter relation substring.',
'TestOverloadedShorthandFunctionValueStaysNotYet':'Self-pinned NotYet and overloaded-function-as-value substring.',
'TestCensusOverloadBinderGuards':'Self-pinned nonnil error and implementation constraint/parameter/result substrings.'}
def fail_line(log,name):
 ev=events(log)
 lines=[e.get('Output','').strip() for e in ev if e.get('Test','').split('/')[0]==name and e['Action']=='output' and ('.go:' in e.get('Output','') or 'panic:' in e.get('Output',''))]
 return lines[0] if lines else next((e.get('Output','').strip() for e in ev if e.get('Test','')==name and 'FAIL' in e.get('Output','')), 'no failing assertion line recorded')
results=[]
for index,row in enumerate(rows,1):
 name=row['test']; kills=sorted(killsets.get(name,set())); unique=[id for id in kills if len(matrix[id])==1]
 subsumer=subsumers.get(name)
 if unique: verdict='sacred'
 elif not kills: verdict='untrue'
 elif subsumer: verdict='subsumed'
 else: verdict='overlapping'
 peers=[subsumer] if subsumer else []
 if verdict=='overlapping':
  remaining=set(kills)
  while remaining:
   candidate=min((other for other in killsets if other!=name and killsets[other]&remaining),key=lambda x:(-len(killsets[x]&remaining),x not in times,statistics.median(times[x]) if x in times else 0,x))
   peers.append(candidate); remaining-=killsets[candidate]
 last=kills[-1] if kills else None
 evidence=('ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u026/cache/'+last+' '+runs[last]['command']+'; '+fail_line(runs[last]['log'],name)) if last else 'No production mutant failed this row in 17 whole-package runs; see kill-matrix.json.'
 result=dict(row='R%02d'%index,test=name,package='internal/lower',file=row['file']+':'+str(row['line']),seconds=statistics.median(times[name]),oracle=oracles[name],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=(last+': '+fail_line(runs[last]['log'],name)) if last else None,verdict=verdict,subsumed_by=peers,mutants_in_matrix=17,probe_kills=['P01'] if probes[name]['exit']!=0 else [],subsumer_seconds=statistics.median(times[subsumer]) if subsumer else None,vacuous=probes[name]['exit']==0,bounded=False,evidence=evidence)
 results.append(result)
(P/'results.json').write_text(json.dumps(results,indent=2)+'\n')
(P/'matrix-rows.json').write_text(json.dumps(sorted({e['Test'].split('/')[0] for e in events('M01.log') if e['Action']=='run' and 'Test' in e}),indent=2)+'\n')
summary={ 'start_commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'nproc':5,'baseline_binary_seconds':33.494,'measurement_wall_seconds':float((P/'measurement-wall.txt').read_text()),'validation_wall_seconds':sum(r['wall'] for r in json.loads((P/'validation.json').read_text()).values()),'matrix_wall_seconds':sum(r['wall'] for r in runs.values()),'individual_rerun_wall_seconds':sum(x['wall'] for r in runs.values() for x in r.get('individual',{}).values()),'probe_wall_seconds':sum(r['wall'] for r in probes.values()),'skips': sorted({e['Test'] for e in events('M01.log') if e['Action']=='skip' and 'Test' in e}), 'survivors':[id for id,names in matrix.items() if not names], 'unit_survivors':[id for id,names in matrix.items() if not any(n in {r['test'] for r in rows} for n in names)]}
(P/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2)); print('Verdicts:',[(r['row'],r['verdict'],r['subsumed_by'],r['vacuous']) for r in results])
