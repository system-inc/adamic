import hashlib,json,os,subprocess,time
from pathlib import Path
e=Path(__file__).resolve().parent;art=e/'artifacts';rows=json.loads((e/'run-times.json').read_text());env={**os.environ,'GOMAXPROCS':'4'};truth=rows[0]['stdout_sha256'];inputs=art/'cases.txt'
(e/'bulk-sanitizer-limit.json').write_text(json.dumps(dict(side='base',status='timeout',limit_seconds=180,stderr_bytes=(art/'base-sanitized.stderr').stat().st_size,note='Exploratory single-process 3919-case benchmark, not a package test or hash shard. Actual full-package sanitizer observations completed on both revisions.'),indent=2)+'\n')
for round_ in range(3):
 for side in (['base','v4'] if round_%2==0 else ['v4','base']):
  if any(r['side']==side and r['mode']=='release' and r['round']==round_ for r in rows):continue
  root='/tmp/adamic-v4-2-base' if side=='base' else '/tmp/adamic-json-82';output=art/(side+'-release.stdout');error=art/(side+'-release.stderr');start=time.monotonic()
  with output.open('wb') as out,error.open('wb') as err:r=subprocess.run([str(art/side/'release'),'--cases',str(inputs)],cwd=root,env=env,stdout=out,stderr=err,timeout=180)
  row=dict(side=side,mode='release',round=round_,seconds=time.monotonic()-start,exit=r.returncode,stdout_bytes=output.stat().st_size,stdout_sha256=hashlib.sha256(output.read_bytes()).hexdigest(),stderr_bytes=error.stat().st_size);rows.append(row);(e/'run-times.json').write_text(json.dumps(rows,indent=2)+'\n');print(row,flush=True);assert r.returncode==0 and row['stderr_bytes']==0 and row['stdout_sha256']==truth,row
(e/'measurement-input.json').write_text(json.dumps(dict(cases=3919,bytes=inputs.stat().st_size,sha256=hashlib.sha256(inputs.read_bytes()).hexdigest()),indent=2)+'\n')
