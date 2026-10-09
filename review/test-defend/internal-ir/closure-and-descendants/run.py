import pathlib,json,difflib,subprocess,time,os
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-ir/closure-and-descendants'
plan=[dict(id='D1',row='TestClosureArgumentsCountTargets',file='internal/ir/call_targets.go',old='return p.ClosureArgumentLayout(call).Count',new='return !p.ClosureArgumentLayout(call).Count',menu='flip condition',aim='Invert the wrapper result while retaining correct argument layouts; the subsumer bypasses this wrapper.'),dict(id='D2',row='TestCallTargetsIncludeEveryDescendant',file='internal/ir/call_targets.go',old='\t\treturn targets\n',new='\t\treturn targets[1:]\n',menu='off-by-one bound',aim='Omit the first virtual implementation. The synthetic effect sets begin with nonthrowing candidates, so effects remain correct while target membership is wrong.')]
for m in plan:
 source=(root/m['file']).read_text(); assert source.count(m['old'])==1
 m['line']=source[:source.index(m['old'])].count('\n')+1
 (p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(source.splitlines(True),source.replace(m['old'],m['new']).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
(p/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')
rows=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')];results=[]
for m in plan:
 diff=p/(m['id']+'.diff');subprocess.run(['git','apply','--check',str(diff)],cwd=root,check=True);subprocess.run(['git','apply',str(diff)],cwd=root,check=True)
 try:
  start=time.monotonic()
  with (p/(m['id']+'.vet.log')).open('w') as log: vet=subprocess.run(['go','vet','./internal/ir/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defendir/cache/'+m['id']); command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/ir/','-run','.']
  with (p/(m['id']+'.log')).open('w') as log: run=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for line in (p/(m['id']+'.log')).read_text().splitlines():
   try: events.append(json.loads(line))
   except ValueError: pass
  outcomes={e['Test']:e['Action'] for e in events if e['Action'] in ['pass','fail','skip'] and e.get('Test') and '/' not in e['Test']}
  result=dict(mutant=m['id'],command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command)+' > '+m['id']+'.log 2>&1',exit=run.returncode,wall_seconds_including_vet_compile=time.monotonic()-start,rows_failed=[r for r,a in outcomes.items() if a=='fail'],rows_passed=[r for r,a in outcomes.items() if a=='pass'],rows_skipped=[r for r,a in outcomes.items() if a=='skip'],rows_unknown=[r for r in rows if r not in outcomes],fail_lines=[e['Output'].strip() for e in events if e.get('OutputType')=='error'],binary_seconds=[e['Elapsed'] for e in events if e['Action'] in ['fail','pass'] and not e.get('Test')])
  results.append(result);(p/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(result),flush=True)
 finally: subprocess.run(['git','apply','-R',str(diff)],cwd=root,check=True)
