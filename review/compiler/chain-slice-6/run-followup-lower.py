import subprocess,pathlib,re,json,time,sys
root=pathlib.Path.cwd(); out=root/'review/compiler/chain-slice-6/followup-lower';out.mkdir(exist_ok=True)
with (out/'lower-list.log').open('w') as log:
 p=subprocess.run(['go','test','./internal/lower','-list','^Test','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=180)
assert p.returncode==0,'listing failed'
names=[s for s in (out/'lower-list.log').read_text().splitlines() if re.fullmatch(r'Test\w+',s)]
assert names and len(set(names))==len(names)
# Named exact selectors partition every top-level test once; no broad package run.
groups=[names[i:i+20] for i in range(0,len(names),20)]
(out/'lower-shards.json').write_text(json.dumps(groups,indent=2)+'\n')
start_index=int(sys.argv[1]) if len(sys.argv)>1 else 0
results=json.loads((out/'lower-results.json').read_text())[:start_index] if start_index else []
for i,group in list(enumerate(groups))[start_index:]:
 command=['go','test','./internal/lower','-run','^('+'|'.join(group)+')$','-count=1','-json','-timeout','90s']
 start=time.monotonic()
 with (out/f'lower-{i:02}.jsonl').open('w') as log:
  p=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=100)
 results.append({'shard':i,'command':command,'exit':p.returncode,'seconds':round(time.monotonic()-start,3)})
 (out/'lower-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(f'shard {i}: exit={p.returncode}',flush=True)
 assert p.returncode==0,f'shard {i} failed'
verdicts={}
for i in range(len(groups)):
 for line in (out/f'lower-{i:02}.jsonl').read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if r.get('Action') in ['pass','skip','fail'] and r.get('Test') in names:
   assert r['Test'] not in verdicts,r['Test']
   verdicts[r['Test']]=r['Action']
assert set(verdicts)==set(names),(set(names)-set(verdicts))
(out/'lower-union.json').write_text(json.dumps({'enumerated':len(names),'pass':sum(v=='pass' for v in verdicts.values()),'skip':sum(v=='skip' for v in verdicts.values()),'fail':sum(v=='fail' for v in verdicts.values()),'verdicts':verdicts},indent=2)+'\n')
