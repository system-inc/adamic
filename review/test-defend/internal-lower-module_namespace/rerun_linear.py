import pathlib,subprocess,json,time,os
p=pathlib.Path('review/test-defend/internal-lower-module_namespace');scope=json.loads((p/'scope.json').read_text());item=next(r for r in json.loads((p/'mutant-plan.json').read_text()) if r['mutant']=='D06')
file=pathlib.Path(item['file']);original=file.read_text();assert original==subprocess.check_output(['git','show',scope['base']+':'+str(file)],text=True)
regexes=['^TestNamespaceCallGraphLinearWork$/^12$','^('+'|'.join(n for n in scope['package_tests'] if n!='TestNamespaceCallGraphLinearWork')+')$']
results=[]
try:
 file.write_text(original.replace(item['old'],item['new'],1))
 for label,regex in zip(['D06-depth12','D06-rest'],regexes):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',regex];env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-module-namespace/cache/D06');t=time.monotonic()
  with (p/'logs'/(label+'.log')).open('w') as log: code=subprocess.call(cmd,env=env,stdout=log,stderr=log)
  events=[]
  for s in (p/'logs'/(label+'.log')).read_text().splitlines():
   try:events.append(json.loads(s))
   except ValueError:pass
  results.append(dict(label=label,command=cmd,env={'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},exit=code,wall_seconds=time.monotonic()-t,rows_failed=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test'))),rows_passed=sorted(e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']),errors=[e for e in events if e.get('OutputType')=='error'],package_result=[e for e in events if e.get('Action') in ['fail','pass'] and not e.get('Test')]))
  (p/'D06-bounded-matrix.json').write_text(json.dumps(results,indent=2));print(label,code,round(results[-1]['wall_seconds'],2),results[-1]['rows_failed'],flush=True)
finally:file.write_text(original)
