import pathlib,json,subprocess,time,os
p=pathlib.Path('review/test-defend/internal-native-radix/session-39e7366a');scope=json.loads((p/'scope.json').read_text());results=[]
for item in json.loads((p/'mutant-plan.json').read_text()):
 ident=item['mutant'];f=pathlib.Path(item['file']);original=f.read_text();assert original==subprocess.check_output(['git','show',scope['base']+':'+str(f)],text=True)
 try:
  f.write_text(original.replace(item['old'],item['new']))
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^TestRuntimeKeyKeepsBoundaries$'];env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-native-radix/cache/'+ident);t=time.monotonic()
  with (p/'logs'/(ident+'-added.log')).open('w') as log:code=subprocess.call(cmd,env=env,stdout=log,stderr=log)
  results.append({'mutant':ident,'command':cmd,'env':{'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},'exit':code,'wall_seconds':time.monotonic()-t});(p/'added-test-matrix.json').write_text(json.dumps(results,indent=2));assert code==0
 finally:f.write_text(original)
if 'TestRuntimeKeyKeepsBoundaries' not in scope['matrix_members']:
 scope['matrix_members'].append('TestRuntimeKeyKeepsBoundaries');scope['matrix_rows'].append('TestRuntimeKeyKeepsBoundaries')
scope['additional_matrix_reason']='The other new top-level test since audit, TestRuntimeKeyKeepsBoundaries, was run separately under each mutant despite being outside regexp execution reach.';(p/'scope.json').write_text(json.dumps(scope,indent=2))
