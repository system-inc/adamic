import concurrent.futures, json, pathlib, subprocess, sys, time
out=pathlib.Path(__file__).parent
mode,binary=sys.argv[1:]
file=out/('admission-'+mode+'.json')
rows=json.loads(file.read_text())
pending=[r for r in rows if r['exit']=='timeout']
def probe(row):
 start=time.monotonic()
 try:
  p=subprocess.run([binary,'c',row['path']],stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=90)
  result=dict(path=row['path'],exit=p.returncode,diagnostic=p.stderr.decode(errors='replace')[:2500],seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired:result=row
 print(mode,row['path'],result['exit'],flush=True)
 return result
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 resolved={r['path']:r for r in pool.map(probe,pending)}
rows=[resolved.get(r['path'],r) for r in rows]
file.write_text(json.dumps(rows,indent=2)+'\n')
print(mode,'admitted',sum(r['exit']==0 for r in rows),'timeouts',sum(r['exit']=='timeout' for r in rows),flush=True)
if (out/'admission-after.json').exists():
 before={r['path']:r for r in json.loads((out/'admission-before.json').read_text())}
 after=json.loads((out/'admission-after.json').read_text())
 result=dict(new=[r['path'] for r in after if r['exit']==0 and before[r['path']]['exit']!=0],regressions=[r['path'] for r in after if r['exit']!=0 and before[r['path']]['exit']==0],timeouts=[r['path'] for r in after if r['exit']=='timeout' or before[r['path']]['exit']=='timeout'])
 (out/'admission-delta.json').write_text(json.dumps(result,indent=2)+'\n')
 print(result,flush=True)
