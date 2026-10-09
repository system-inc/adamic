from pathlib import Path
import subprocess,json,concurrent.futures
out=Path('review/compiler/lowering-gaps-3');rows=json.loads((out/'admission-before.json').read_text());pending=[r for r in rows if r['exit']=='timeout']
def probe(row):
 try:
  r=subprocess.run(['/tmp/adamic-gaps3-main','c',row['path']],stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=90)
  row=dict(path=row['path'],exit=r.returncode,diagnostic=r.stderr.decode(errors='replace')[:2000]);print(row['path'],row['exit'],flush=True);return row
 except subprocess.TimeoutExpired:print(row['path'],'timeout',flush=True);return row
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:resolved={r['path']:r for r in pool.map(probe,pending)}
rows=[resolved.get(r['path'],r) for r in rows];(out/'admission-before.json').write_text(json.dumps(rows,indent=2)+'\n');print('remaining',sum(r['exit']=='timeout' for r in rows),flush=True)
