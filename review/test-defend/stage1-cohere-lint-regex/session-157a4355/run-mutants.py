import pathlib,subprocess,json,os,time
root=pathlib.Path(__file__).resolve().parent
results=[]
for m in json.loads((root/'plan.json').read_text()):
 path=pathlib.Path(m['file']);original=path.read_text()
 try:
  assert original.count(m['from'])==1,m
  path.write_text(original.replace(m['from'],m['to'],1))
  (root/(m['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',str(path)]))
  with (root/(m['mutant']+'-vet.log')).open('w') as f:
   check=subprocess.run(['go','vet',m['package']],stdout=f,stderr=subprocess.STDOUT)
  assert check.returncode==0,m
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-lint-regex/cache/'+m['mutant']
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/regex/','-run','.']
  start=time.monotonic()
  with (root/(m['mutant']+'.log')).open('w') as f:result=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
  events=[]
  for line in (root/(m['mutant']+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  results.append({**m,'command':cmd,'build_cache_dir':env['ADAMIC_BUILD_CACHE_DIR'],'seconds':round(time.monotonic()-start,3),'exit':result.returncode,'rows_failed':sorted({e['Test'].split('/')[0] for e in events if e.get('Test') and e.get('Action')=='fail'}),'rows_passed':sorted({e['Test'] for e in events if e.get('Test') and '/' not in e['Test'] and e.get('Action')=='pass'}),'failure_output':[e['Output'].strip() for e in events if ('regex_test.go:128:' in e.get('Output','') or 'shapes_test.go:45:' in e.get('Output','') or 'SyntaxError:' in e.get('Output',''))]})
  (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
 finally:path.write_text(original)
