import pathlib,json,subprocess,os,time
root=pathlib.Path(__file__).resolve().parent
rows=['TestSyntaxRefusals','TestCookedSurrogates','TestSyntaxGrammar','TestGeneratedAgreement','TestThroughput','TestLossyInputRefusal','TestCookedSurrogateMutant']
results=[]
for m in json.loads((root/'plan.json').read_text()):
 path=pathlib.Path(m['file']);original=path.read_text()
 try:
  assert original.count(m['from'])==1
  path.write_text(original.replace(m['from'],m['to'],1))
  (root/(m['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',str(path)]))
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-estree/cache/'+m['mutant']
  for row in rows:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^'+row+'$']
   start=time.monotonic()
   with (root/(m['mutant']+'-'+row+'.log')).open('w') as f:done=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
   events=[]
   for l in (root/(m['mutant']+'-'+row+'.log')).read_text().splitlines():
    try:events.append(json.loads(l))
    except ValueError:pass
   results.append({'mutant':m['mutant'],'row':row,'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'seconds':round(time.monotonic()-start,3),'exit':done.returncode,'failures':[e.get('Test') for e in events if e.get('Action')=='fail' and e.get('Test')],'output':[e['Output'].strip() for e in events if '.go:' in e.get('Output','') and ('timeout=' in e['Output'] or 'Go ' in e['Output'] or 'port ' in e['Output'] or 'mutant anchor' in e['Output'])]})
   (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
 finally:path.write_text(original)
