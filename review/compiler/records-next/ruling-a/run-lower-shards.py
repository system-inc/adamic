import subprocess, pathlib, re, json
root=pathlib.Path('/workspace/adamic'); out=root/'review/compiler/records-next/ruling-a'
with (out/'lower-list.log').open('w') as log:
 r=subprocess.run(['go','test','./internal/lower','-list','^Test'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=85)
assert r.returncode==0
names=[x for x in (out/'lower-list.log').read_text().splitlines() if re.fullmatch('Test\w+',x)]
results=[]
for i in range(12):
 selected=names[i::12]; command=['go','test','./internal/lower','-run','^('+'|'.join(selected)+')$','-count=1','-timeout','85s']
 with (out/f'lower-shard-{i+1:02}.log').open('w') as log:
  try: code=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=89).returncode
  except subprocess.TimeoutExpired: code=124
 results.append({'shard':i+1,'tests':selected,'exit':code});(out/'lower-shards.json').write_text(json.dumps(results,indent=2)+'\n')
 print('lower shard',i+1,'exit',code,flush=True)
assert all(x['exit']==0 for x in results)
