import pathlib,time,json,subprocess,os
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-oracle-interface_cast/session')
while 'completed' not in (p/'new-driver.log').read_text():time.sleep(1)
menu=json.loads((p/'menu.json').read_text());out=[]
base=json.loads((p/'matrix.json').read_text())[0]['command'][-1]
follow=[('D5',n) for n in ['TestInterfaceCastOracle','TestInterfaceCastImportedConstruction','TestInterfaceCastScalarTags','TestInterfaceCastChecksMalformedRead']]
follow.append(('D8',base))
for mid,sel in follow:
 m=next(x for x in menu if x['mutant']==('D1' if mid=='D8' else mid));file=pathlib.Path(m['file']);orig=file.read_text(); new='for _, member := range declared.Types()[:min(len(declared.Types()), 3)] {' if mid=='D8' else m['new']; selector=sel if sel.startswith('^') else '^'+sel+'$'
 try:
  file.write_text(orig.replace(m['old'],new,1))
  if mid=='D8':(p/'D8.diff').write_bytes(subprocess.check_output(['git','diff','--',m['file']]))
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',selector];start=time.monotonic();log=mid+'-'+('bounded' if mid=='D8' else sel)+'.log'
  with (p/log).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/interface-defense/cache/'+mid})
  events=[]
  for l in (p/log).read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  out.append({'mutant':mid,'command':cmd,'exit':r.returncode,'wall':time.monotonic()-start,'log':log,'statuses':{e['Test']:e['Action'] for e in events if e.get('Test') and '/' not in e['Test'] and e['Action'] in ['pass','fail','skip']},'panic':[e.get('Output','') for e in events if 'panic:' in e.get('Output','')],'errors':[{'test':e.get('Test'), 'line':e.get('Output','').strip()} for e in events if '.go:' in e.get('Output','') and e.get('Action')=='output']});(p/'follow-matrix.json').write_text(json.dumps(out,indent=2));print(mid,sel,r.returncode,flush=True)
 finally:file.write_text(orig)
print('completed',flush=True)
