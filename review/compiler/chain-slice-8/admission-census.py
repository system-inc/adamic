import pathlib,subprocess,json,time,hashlib,concurrent.futures
root=pathlib.Path('/workspace/adamic'); out=root/'review/compiler/chain-slice-8/admission';out.mkdir(exist_ok=True)
paths=sorted(set(p for directory in ['internal/oracle/testdata','docs/step-18','stage3/fixtures'] for p in (root/directory).rglob('*.a')))
# Temporary TypeScript copies exercise the member's distinct TypeScript boundary.
copies=pathlib.Path('/tmp/adamic-gate/chain-slice-8-any');copies.mkdir(exist_ok=True)
for p in (root/'internal/load/testdata/0.1/refuse/checked_any').glob('*.a'):
 q=copies/(p.stem+'.ts');q.write_bytes(p.read_bytes());paths.append(q)
def check(pair):
 i,p=pair; row={'path':str(p.relative_to(root)) if p.is_relative_to(root) else 'checked .ts copy: '+p.name,'source_sha256':hashlib.sha256(p.read_bytes()).hexdigest()}
 for name,binary in [('main','/tmp/chain-slice-8-main'),('slice','/tmp/chain-slice-8')]:
  try:
   result=subprocess.run([binary,'c',str(p)],cwd=root,capture_output=True,timeout=12)
   row[name]={'exit':result.returncode,'admitted':result.returncode==0,'diagnostic':result.stderr.decode(errors='replace')}
  except subprocess.TimeoutExpired:row[name]={'timeout':True,'admitted':False}
 row['newly_admitted']=row['slice']['admitted'] and not row['main']['admitted'];row['regression']=row['main']['admitted'] and not row['slice']['admitted'];row['absolute_path']=str(p)
 return row
rows=[];start=time.monotonic()
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for row in pool.map(check,enumerate(paths)):
  rows.append(row)
  if len(rows)%25==0:
   (out/'census.json').write_text(json.dumps(rows,indent=2)+'\n');print(len(rows),'new',sum(r['newly_admitted'] for r in rows),'regressions',sum(r['regression'] for r in rows),flush=True)
(out/'census.json').write_text(json.dumps(rows,indent=2)+'\n')
summary={'programs':len(rows),'newly_admitted':sum(r['newly_admitted'] for r in rows),'regressions':sum(r['regression'] for r in rows),'seconds':time.monotonic()-start}
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n'); print(summary,flush=True)
