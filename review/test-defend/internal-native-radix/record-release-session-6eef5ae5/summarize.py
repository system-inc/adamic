import pathlib,json,subprocess,shlex
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix');runs=json.loads((p/'runs.json').read_text());plan=json.loads((p/'plan.json').read_text());scope=json.loads((p/'matrix-rows.json').read_text());matrix=[]
for r in runs:
 es=[]
 for l in (p/r['log']).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 errors={}
 for e in es:
  if e.get('OutputType')=='error':errors.setdefault(e.get('Test'),[]).append(e.get('Output','').strip())
 failed=[n for n in r['failed'] if '/' not in n];passed=[n for n in r['passed'] if '/' not in n];infra=[n for n in failed if any('text file busy' in s for s in errors.get(n,[]))];build=[n for n in failed if any('clang failed' in s or 'must compile:' in s for s in errors.get(n,[]))]
 matrix.append(dict(r,failed=failed,passed=passed,errors_by_row=errors,infrastructure_failures=infra,build_failures=build,unknown=[n for n in scope if n not in failed+passed],binary_seconds=next((e.get('Elapsed') for e in es if not e.get('Test') and e.get('Action') in ['pass','fail']),None)))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
results=[]
for name,ids,prior in [('TestRecordBenchmark',['B1','B2','B3'],'TestRuntimeStringEquality'),('TestRuntimeReleasePaths',['R1','R2','R3'],'TestRegExpIteratorResultShape')]:
 attempts=[];proof=[];valid=0;unique=[]
 for id in ids:
  m=next(x for x in plan if x['id']==id);cells=[x for x in matrix if x['id']==id];failed=sorted(set(n for x in cells for n in x['failed'] if n not in x['infrastructure_failures']+x['build_failures']));passed=sorted(set(n for x in cells for n in x['passed']));unknown=sorted(set(n for x in cells for n in x['unknown'])-set(failed+passed));attempts.append(dict(mutant=id,file_line=m['file']+':'+str(m['line']),change=m['old']+' -> '+m['new'],rows_failed=failed,rows_passed=passed,rows_unknown=unknown,aim=m['aim'],subject_result='fail' if name in failed else 'pass' if name in passed else 'unknown'))
  if name in passed or (name in failed and any(n!=name for n in failed)):valid+=1
  if failed==[name] and not unknown:unique.append(id)
  for x in cells:
   if name in x['failed'] and name not in x['infrastructure_failures']+x['build_failures']:
    proof.append(dict(mutant=id,command='ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/workspace/scratch/defend-radix/cache/'+id+' '+shlex.join(x['command'])+' > '+x['log']+' 2>&1',line=x['errors_by_row'].get(name,['--- FAIL: '+name])[0],log=x['log']));break
 results.append(dict(test=name,package='internal/native',prior_verdict='subsumed',subsumed_by=[prior],defense='defended' if unique else 'not defended' if valid==3 else 'cannot-judge',unique_mutant=unique[0] if unique else None,attempts=attempts,evidence=proof,bounded=True,matrix_rows=scope))
(p/'results.json').write_text(json.dumps(results,indent=2));print([(x['test'],x['defense'],[(a['mutant'],a['subject_result'],a['rows_failed']) for a in x['attempts']]) for x in results])
