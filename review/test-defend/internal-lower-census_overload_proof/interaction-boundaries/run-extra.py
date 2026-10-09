import pathlib,json,difflib,subprocess,time,os
root=pathlib.Path('/workspace/adamic'); p=root/'review/test-defend/internal-lower-census_overload_proof/interaction-boundaries'
plan=[{'id': 'D4', 'row': 'TestCensusPredicateMarkerKeepsProofBoundaries', 'file': 'internal/lower/census_small.go', 'old': 'return parameter.Flags()&checker.TypeFlagsNever != 0', 'new': 'return parameter.Flags()&(checker.TypeFlagsNever|checker.TypeFlagsNumber) != 0', 'menu': 'change constant', 'aim': 'Classify number-rest slots as erased in both marker admission and function assignment, avoiding the independent refusal that masked D3.'}]
for m in plan:
 source=(root/m['file']).read_text(); assert source.count(m['old'])==1
 m['line']=source[:source.index(m['old'])].count('\n')+1
 diff=''.join(difflib.unified_diff(source.splitlines(True),source.replace(m['old'],m['new']).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (p/(m['id']+'.diff')).write_text(diff)
(p/'plan-extra.json').write_text(json.dumps(plan,indent=2)+'\n')
names=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]
results=json.loads((p/'matrix.json').read_text())
for m in plan:
 diff=p/(m['id']+'.diff'); subprocess.run(['git','apply','--check',str(diff)],cwd=root,check=True); subprocess.run(['git','apply',str(diff)],cwd=root,check=True)
 try:
  start=time.monotonic()
  with (p/(m['id']+'.vet.log')).open('w') as log: vet=subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend027/cache/'+m['id'])
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.']
  with (p/(m['id']+'.log')).open('w') as log: run=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for line in (p/(m['id']+'.log')).read_text().splitlines():
   try: events.append(json.loads(line))
   except ValueError: pass
  outcomes={e['Test']:e['Action'] for e in events if e['Action'] in ['pass','fail','skip'] and e.get('Test') and '/' not in e['Test']}
  result=dict(mutant=m['id'],command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command),exit=run.returncode,wall_seconds=time.monotonic()-start,rows_failed=[r for r,a in outcomes.items() if a=='fail'],rows_passed=[r for r,a in outcomes.items() if a=='pass'],rows_skipped=[r for r,a in outcomes.items() if a=='skip'],rows_unknown=[r for r in names if r not in outcomes],fail_lines=[e['Output'].strip() for e in events if e.get('Output') and (': want ' in e['Output'] or 'exception escaped' in e['Output'])])
  results.append(result); (p/'matrix.json').write_text(json.dumps(results,indent=2)+'\n'); print(json.dumps({k:v for k,v in result.items() if k not in ['rows_passed']}),flush=True)
 finally: subprocess.run(['git','apply','-R',str(diff)],cwd=root,check=True)
