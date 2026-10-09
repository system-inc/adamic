import pathlib,time,json,subprocess,os
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-oracle-interface_cast/session')
while 'completed' not in (p/'driver.log').read_text():time.sleep(1)
menu=json.loads((p/'menu.json').read_text());out=[]; selector='^(TestFractionalPowersReachRuntime|TestReviewPrograms.*)$'
for m in [{'mutant':'clean'}]+menu:
 mid=m['mutant'];file=pathlib.Path(m['file']) if 'file' in m else None;orig=file.read_text() if file else None
 try:
  if file:file.write_text(orig.replace(m['old'],m['new'],1))
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',selector]; start=time.monotonic()
  with (p/(mid+'-new.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/interface-defense/cache/'+mid})
  statuses={};errors=[];cooked=False
  for l in (p/(mid+'-new.log')).read_text().splitlines():
   try:e=json.loads(l)
   except:continue
   if 'panic: test timed out' in e.get('Output',''):cooked=True
   n=e.get('Test','')
   if n and '/' not in n and e.get('Action') in ['pass','fail','skip']:statuses[n]=e['Action']
   if n and '.go:' in e.get('Output','') and e.get('Action')=='output':errors.append({'test':n,'line':e['Output'].strip()})
  out.append({'mutant':mid,'command':cmd,'exit':r.returncode,'wall':time.monotonic()-start,'cooked':cooked,'statuses':statuses,'errors':errors});(p/'new-matrix.json').write_text(json.dumps(out,indent=2));print(mid,r.returncode,flush=True)
  if mid=='clean' and r.returncode:break
 finally:
  if file:file.write_text(orig)
print('completed',flush=True)
