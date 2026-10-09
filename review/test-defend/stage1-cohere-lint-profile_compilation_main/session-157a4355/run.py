import pathlib,subprocess,json,time,os,sys
p=pathlib.Path('/tmp/defend-lint-profile'); root=pathlib.Path('/workspace/adamic'); plan=json.loads((p/'plan.json').read_text())
pattern='^(TestProfileCompilation.*|TestProduct_ProfileCompilation.*)$'
for m in plan:
 f=root/m['file']; old=f.read_bytes(); new=old.replace(m['before'].encode(),m['after'].encode()); assert new!=old
 try:
  f.write_bytes(new)
  (p/(m['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',m['file']],cwd=root))
  with (p/(m['mutant']+'-vet.log')).open('wb') as log: vet=subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR=str(p/'cache'/m['mutant']))
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',pattern]
  start=time.monotonic()
  with (p/(m['mutant']+'.log')).open('wb') as log: r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for l in (p/(m['mutant']+'.log')).read_text().splitlines():
   try: events.append(json.loads(l))
   except: pass
  result={'mutant':m['mutant'],'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'wall_seconds':time.monotonic()-start,'exit':r.returncode,'rows_failed':[e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e], 'rows_passed':[e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e], 'skips':[e['Test'] for e in events if e.get('Action')=='skip' and 'Test' in e], 'binary_seconds':events[-1].get('Elapsed') if events else None}
  (p/(m['mutant']+'.json')).write_text(json.dumps(result,indent=2)); print(json.dumps(result),flush=True)
 finally:
  f.write_bytes(old)
