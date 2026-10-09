import pathlib,json,subprocess,shlex
p=pathlib.Path('/workspace/adamic/review/test-defend/stage1-typescript-parser-jsx_rejection');runs=json.loads((p/'runs.json').read_text());plan=json.loads((p/'plan.json').read_text());subjects=['TestJsxMemberNameRejection','TestJsxNode','TestJsxNative'];allrows=json.loads((p/'matrix-rows.json').read_text());matrix=[]
for r in runs:
 events=[]
 for l in (p/r['log']).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 errors={}
 for e in events:
  if e.get('OutputType')=='error':errors.setdefault(e.get('Test'),[]).append(e.get('Output','').strip())
 failed=[t for t in r['failed'] if '/' not in t];passed=[t for t in r['passed'] if '/' not in t]
 anchor=[t for t in failed if any('must change exactly one site' in s or 'exactly one site' in s for s in errors.get(t,[]))]
 compilefail=[t for t in failed if any('native build:' in s or 'NotYet' in s or 'Refused:' in s for s in errors.get(t,[]))]
 matrix.append(dict(r,failed=failed,passed=passed,anchor_failures=anchor,possible_build_failures=compilefail,errors_by_row=errors,unknown=[t for t in allrows if t not in failed+passed] if r['label']=='matrix' else []))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
result=[]
for name,ids,subs in [(subjects[0],['D1','D2','D3'],['TestJsxNode']),(subjects[1],['D4','D5','D6'],['TestJsxNative']),(subjects[2],['D4','D5','D6'],['TestJsxNode'])]:
 attempts=[];proof=[]
 for id in ids:
  m=next(x for x in plan if x['id']==id);cells=[x for x in matrix if x['id']==id];failed=sorted(set(t for x in cells for t in x['failed'] if t not in x['anchor_failures']+x['possible_build_failures']));passed=sorted(set(t for x in cells for t in x['passed']))
  a=dict(mutant=id,file_line=m['file']+':'+str(m['line']),change=m['old']+' -> '+m['new'],rows_failed=failed,rows_passed=passed,aim=m['aim']);attempts.append(a)
  for x in cells:
   if name in x['failed'] and name not in x['anchor_failures']+x['possible_build_failures']:
    proof.append(dict(mutant=id,command=shlex.join(x['command'])+' > '+x['log']+' 2>&1',line=x['errors_by_row'].get(name,['--- FAIL: '+name])[0],log=x['log']));break
 unique=[a for a in attempts if a['rows_failed']==[name] and not any(x['unknown'] for x in matrix if x['id']==a['mutant'])]
 valid_attempts=all(name in a['rows_failed'] and any(t!=name for t in a['rows_failed']) for a in attempts)
 defense='defended' if unique else 'not defended' if valid_attempts else 'cannot-judge'
 result.append(dict(test=name,package='stage1/typescript/parser',prior_verdict='subsumed',subsumed_by=subs,defense=defense,unique_mutant=unique[0]['mutant']+' '+unique[0]['file_line'] if unique else None,attempts=attempts,evidence=proof,coverage='coverage-diffs.json; V8 profiles are supplemental and transformed-source offsets are not exact original line coverage.'))
(p/'results.json').write_text(json.dumps(result,indent=2));print([(r['test'],r['defense'],[(a['mutant'],a['rows_failed']) for a in r['attempts']]) for r in result])
