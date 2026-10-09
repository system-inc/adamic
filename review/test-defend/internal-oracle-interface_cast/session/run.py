import pathlib,json,subprocess,os,time
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-oracle-interface_cast/session'); menu=json.loads((p/'menu.json').read_text()); selector='^(TestInterfaceCast.*|TestCheckedCast.*|TestUncheckableCastAdmission|Test.*View.*|TestNarrowedUnion.*|TestNarrowedFieldUsesSharedReadiness)$'; out=[]
for m in [{'mutant':'clean'}]+menu:
 mid=m['mutant'];file=pathlib.Path(m['file']) if 'file' in m else None;orig=file.read_text() if file else None
 try:
  if file:file.write_text(orig.replace(m['old'],m['new'],1))
  vetcmd=['go','vet','./internal/lower/','./internal/native/'];
  if file:
   with (p/(mid+'-vet.log')).open('w') as f:v=subprocess.run(vetcmd,stdout=f,stderr=subprocess.STDOUT)
   if v.returncode:out.append({'mutant':mid,'compile_exit':v.returncode});continue
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',selector];start=time.monotonic();env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/interface-defense/cache/'+mid}
  with (p/(mid+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
  statuses={};errors=[];cooked=False
  for l in (p/(mid+'.log')).read_text().splitlines():
   try:e=json.loads(l)
   except:continue
   if 'panic: test timed out' in e.get('Output',''):cooked=True
   n=e.get('Test','')
   if n and '/' not in n and e.get('Action') in ['pass','fail','skip']:statuses[n]=e['Action']
   if n and e.get('Action')=='output' and '.go:' in e.get('Output','') and ('differ' in e['Output'] or 'want' in e['Output'] or 'survived' in e['Output']):errors.append({'test':n,'line':e['Output'].strip()})
  row={'mutant':mid,'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'exit':r.returncode,'wall':time.monotonic()-start,'cooked':cooked,'statuses':statuses,'errors':errors};out.append(row);(p/'matrix.json').write_text(json.dumps(out,indent=2));print(mid,r.returncode,row['wall'],[k for k,v in statuses.items() if v=='fail'],flush=True)
  if mid=='clean' and r.returncode:break
 finally:
  if file:file.write_text(orig)
print('completed',flush=True)
