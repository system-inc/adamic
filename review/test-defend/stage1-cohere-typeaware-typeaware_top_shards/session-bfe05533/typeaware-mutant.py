import pathlib,subprocess,json,time,os
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/stage1-cohere-typeaware-typeaware_top_shards/session-bfe05533';plan=json.loads((P/'mutant-plan.json').read_text())[0];f=R/plan['file'];original=f.read_text();assert original.count(plan['old'])==1;env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/workspace/defend-typeaware-cache/D1';runs=[]
def run(id,rows):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run','^('+'|'.join(rows)+')$'];t=time.monotonic()
 with (P/(id+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=R,env=env,stdout=log,stderr=subprocess.STDOUT)
 es=[]
 for line in (P/(id+'.log')).read_text().splitlines():
  try:es.append(json.loads(line))
  except:pass
 states={e['Test']:e['Action'] for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']};fails=[n for n,a in states.items() if a=='fail'];item=dict(id=id,rows=rows,states=states,failed=fails,exit=r.returncode,seconds=time.monotonic()-t,cache=env['ADAMIC_BUILD_CACHE_DIR'],command=cmd,binary_seconds=next((e.get('Elapsed') for e in reversed(es) if not e.get('Test') and e.get('Action') in ['pass','fail']),None));runs.append(item);(P/'mutant-runs.json').write_text(json.dumps(runs,indent=2));print(id,item['seconds'],len(states),fails,flush=True)
 return item
try:
 f.write_text(original.replace(plan['old'],plan['new']))
 for id,name in [('D1-build-asan','TestProduct_typeaware_native_asan'),('D1-build-native','TestProduct_typeaware_native'),('D1-setup','TestTypeAwareAgreementAndMutants_Setup')]:
  r=run(id,[name])
  if r['exit']:raise RuntimeError('cooked or red prebuild '+id)
 groups=json.loads((P/'baseline-groups.json').read_text())
 for i,g in enumerate(groups):
  r=run('D1-matrix-'+str(i),g['rows'])
  if len(r['states'])!=len(g['rows']):raise RuntimeError('incomplete matrix '+r['id'])
finally:f.write_text(original)
