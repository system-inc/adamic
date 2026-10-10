import pathlib,json,subprocess,time
root=pathlib.Path('/workspace/adamic');out=root/'review/compiler/chain-slice-8';rows=json.loads((out/'lower-results.json').read_text())
for i,r in enumerate(rows):
 if r['exit']==0:continue
 log=out/f'lower-{i:02}-repaired.jsonl';start=time.monotonic();cmd=['go','test','-p','4','./internal/lower','-run','^('+'|'.join(r['tests'])+')$','-count=1','-json','-timeout','90s']
 with log.open('w') as f:code=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=120).returncode
 r.update(exit=code,seconds=time.monotonic()-start,log=str(log.relative_to(root)))
 (out/'lower-results-final.json').write_text(json.dumps(rows,indent=2)+'\n');print(i,code,r['seconds'],flush=True)
raise SystemExit(any(r['exit'] for r in rows))
