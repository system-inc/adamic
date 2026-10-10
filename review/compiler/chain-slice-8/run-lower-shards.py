import pathlib,re,subprocess,json,time
root=pathlib.Path('/workspace/adamic'); out=root/'review/compiler/chain-slice-8'
names=sorted(set(n for p in (root/'internal/lower').glob('*_test.go') for n in re.findall(r'^func (Test\w+)\(t \*testing.T\)',p.read_text(),re.M)))
results=[]
for i in range(0,len(names),12):
 selected=names[i:i+12]; log=out/f'lower-{i//12:02}.jsonl'; start=time.monotonic()
 cmd=['go','test','-p','4','./internal/lower','-run','^('+'|'.join(selected)+')$','-count=1','-json','-timeout','90s']
 with log.open('w') as f:
  try: code=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=120).returncode
  except subprocess.TimeoutExpired: code=124
 results.append(dict(tests=selected,exit=code,seconds=time.monotonic()-start,log=str(log.relative_to(root))))
 (out/'lower-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(i//12,code,results[-1]['seconds'],flush=True)
raise SystemExit(any(r['exit'] for r in results))
