import pathlib,json,subprocess,os,time
p=pathlib.Path('review/test-audit/stage1-typescript-parser-jsx_rejection')
rows=json.loads((p/'rows.json').read_text()); mutations=json.loads((p/'mutants.json').read_text())
jsx=rows[:2]+rows[3:5]
selection={'M1':jsx,'M2':jsx,'M3':jsx,'M4':rows[7:9],'P1':rows[:2]+rows[3:5]+rows[7:9],'W1':[rows[2],rows[5]],'W2':[rows[9]],'S1':[rows[6]]}
status=json.loads((p/'run-status.json').read_text())
mutations=[m for m in mutations if m['id'] in ['P1','W1','W2','S1']]
env=os.environ.copy()
with (p/'clean-build.log').open('w') as log:
 cleanBuild=subprocess.run(['timeout','90','go','run',str(p/'build_probe.go'),'/tmp/u156/bin/clean'],stdout=log,stderr=subprocess.STDOUT,env=env)
if cleanBuild.returncode:raise RuntimeError('clean native build failed')
fixture=pathlib.Path('/tmp/u156/escaped.tsx');fixture.write_text(r'const x = <Fo\u006f/>;')
witness=[]
result=subprocess.run(['/tmp/u156/bin/clean',str(fixture),'--whole'],capture_output=True,text=True)
witness.append({'id':'clean','exit':result.returncode,'stdout':result.stdout,'stderr':result.stderr})
(p/'M2-behavior.json').write_text(json.dumps(witness,indent=2))
for m in mutations:
 file=pathlib.Path(m['file']);clean=file.read_text();assert clean.count(m['old'])==1
 try:
  file.write_text(clean.replace(m['old'],m['new']))
  env=os.environ.copy();env['ADAMIC_PARSER_BENCH']='1';env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u156-typescript'
  if m['id'].startswith('M') or m['id']=='P1':env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u156/cache/'+m['id']
  if not m.get('harness'):
   with (p/(m['id']+'-build.log')).open('w') as log:
    build=subprocess.run(['timeout','90','go','run',str(p/'build_probe.go'),'/tmp/u156/bin/'+m['id']],stdout=log,stderr=subprocess.STDOUT,env=env)
   if build.returncode:raise RuntimeError('native build failed '+m['id'])
   if m['id']=='M2':
    result=subprocess.run(['/tmp/u156/bin/M2',str(fixture),'--whole'],capture_output=True,text=True)
    witness.append({'id':'M2','exit':result.returncode,'stdout':result.stdout,'stderr':result.stderr})
    (p/'M2-behavior.json').write_text(json.dumps(witness,indent=2))
  if m.get('harness'):
   with (p/(m['id']+'-vet.log')).open('w') as log:
    v=subprocess.run(['timeout','90','go','vet','./stage1/typescript/parser/'],stdout=log,stderr=subprocess.STDOUT,env=env)
   if v.returncode:raise RuntimeError('vet failed '+m['id'])
  for name in selection[m['id']]:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run','^'+name+'$']
   start=time.time()
   with (p/(m['id']+'-'+name+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
   status.append({'id':m['id'],'test':name,'exit':r.returncode,'wall':time.time()-start,'command':' '.join(cmd),'cache':env.get('ADAMIC_BUILD_CACHE_DIR')})
   (p/'run-status.json').write_text(json.dumps(status,indent=2))
   print(m['id'],name,r.returncode,round(time.time()-start,2),flush=True)
 finally:file.write_text(clean)
