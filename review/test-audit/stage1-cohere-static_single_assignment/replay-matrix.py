import pathlib,subprocess,os,json,time,difflib
repo=pathlib.Path('/workspace/adamic'); root=repo/'review/test-audit/stage1-cohere-static_single_assignment'
plan=json.loads((root/'plan.json').read_text()); results=[]
for m in plan['mutants']:
 p=repo/m['file']; original=p.read_text(); changed=original.replace(m['from_'],m['to'],1)
 (root/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 p.write_text(changed)
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u136/cache/'+m['id'];env['AUDIT_CLANG_LOG']=str(root/'logs'/(m['id']+'-clang.jsonl'));env['PATH']='/tmp/u136/bin:'+env['PATH']
 try:
  if m['file'].endswith('.go'):
   with (root/'logs'/(m['id']+'-vet.log')).open('w') as f: rc=subprocess.run(['go','vet','./internal/lower/'],stdout=f,stderr=subprocess.STDOUT,env=env).returncode
   if rc: raise Exception('vet failed')
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/static_single_assignment/','-run','.']
  t=time.monotonic()
  with (root/'logs'/(m['id']+'.log')).open('w') as f: rc=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
  results.append(dict(id=m['id'],command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command)+' > '+str(root/'logs'/(m['id']+'.log'))+' 2>&1',wall_seconds=time.monotonic()-t,returncode=rc))
  (root/'matrix-runs.json').write_text(json.dumps(results,indent=2)+'\n')
 finally: p.write_text(original)
