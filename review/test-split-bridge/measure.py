#!/usr/bin/env python3
import json,os,re,subprocess,sys,time
from pathlib import Path
root=Path(__file__).resolve().parents[2]
mode=sys.argv[1]
logs=Path('/tmp/test-split-bridge')/mode;logs.mkdir(parents=True,exist_ok=True)
env={**os.environ,'GOMAXPROCS':'4','ADAMIC_GATE_UNCACHED':'1'}
if mode=='after-corpus':env['ADAMIC_TSGO_CORPUS']='/tmp/test-split-bridge-typescript'
rows=[]
for package in ['bridge/tsgo','bridge/tsgo/checker']:
 label=package.replace('/','-')
 listing=logs/(label+'-list.jsonl')
 with listing.open('w') as out:
  result=subprocess.run(['go','test','./'+package,'-json','-list','^Test','-count=1'],cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 assert result.returncode==0,listing
 names=[]
 for line in listing.read_text().splitlines():
  try:record=json.loads(line)
  except json.JSONDecodeError:continue
  name=record.get('Output','').strip()
  if re.fullmatch(r'Test\w+',name):names.append(name)
 for name in names:
  command=['go','test','./'+package,'-json','-run','^'+name+'$','-count=1','-parallel=4','-timeout='+('30m' if mode=='before' else '30s')]
  logfile=logs/(label+'-'+name+'.jsonl'); start=time.monotonic()
  with logfile.open('w') as out:run=subprocess.run(command,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  events=[]
  for line in logfile.read_text().splitlines():
   try:events.append(json.loads(line))
   except json.JSONDecodeError:pass
  terminals=[e for e in events if e.get('Test')==name and e['Action'] in ('pass','fail','skip')]
  assert len(terminals)==1,(name,logfile)
  invocations=[e for e in events if not e.get('Test') and e['Action'] in ('pass','fail','skip')]
  row={'package':package,'test':name,'seconds':terminals[0]['Elapsed'],'invocation_seconds':invocations[-1]['Elapsed'],'wall_seconds':time.monotonic()-start,'action':terminals[0]['Action'],'exit':run.returncode,'command':command,'log':str(logfile)}
  rows.append(row)
  (root/'review/test-split-bridge'/ (mode+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
  print(f"{package}/{name}: {row['seconds']:.3f}s {row['action']}",flush=True)
  assert row['exit']==0,row
  if mode.startswith('after'):assert row['invocation_seconds']<30,row
