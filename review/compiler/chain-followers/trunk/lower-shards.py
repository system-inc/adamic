import subprocess,pathlib,json
out=pathlib.Path(__file__).resolve().parent
with (out/'lower-list.log').open('w') as log:r=subprocess.run(['go','test','./internal/lower','-list','^Test','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=150)
assert r.returncode==0
names=[s for s in (out/'lower-list.log').read_text().splitlines() if s.startswith('Test')]
results=json.loads((out/'lower-shards.json').read_text())[:3]
for i in range(105,len(names),35):
 command=['go','test','./internal/lower','-run','^('+'|'.join(names[i:i+35])+')$','-count=1','-timeout','90s']
 with (out/f'lower-shard-{i//35:02d}.log').open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=150)
 results.append(dict(command=command,exit=r.returncode));(out/'lower-shards.json').write_text(json.dumps(results,indent=2))
 print(i//35,r.returncode,flush=True)
 if r.returncode:raise SystemExit(1)
