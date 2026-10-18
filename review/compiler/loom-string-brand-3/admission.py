import concurrent.futures,hashlib,json,os,subprocess,tempfile,time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-string-brand-3';baseline={}
for name in ['expression.go','refusals.go','cast.go','cast_proof.go']:
 path='internal/lower/'+name
 target=out/(name+'.main.txt');target.write_text(subprocess.check_output(['git','show','origin/main:'+path],text=True));baseline[str(root/path)]=str(target)
overlay=out/'main-overlay.json';overlay.write_text(json.dumps({'Replace':baseline}))
with tempfile.TemporaryDirectory(prefix='string-brand-admission-') as directory:
 binaries={phase:str(Path(directory)/phase) for phase in ['main','after']};builds=[]
 for phase,binary in binaries.items():
  cmd=['go','build']+(['-overlay',str(overlay)] if phase=='main' else [])+['-o',binary,'./cmd/adamic'];start=time.monotonic()
  with (out/('build-'+phase+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=180)
  builds.append(dict(command=cmd,exit=r.returncode,seconds=round(time.monotonic()-start,3)));assert r.returncode==0,phase
 (out/'admission-builds.json').write_text(json.dumps(builds,indent=2));hashes={phase:hashlib.sha256(Path(binary).read_bytes()).hexdigest() for phase,binary in binaries.items()}
 files=sorted(p for p in (root/'internal/oracle/testdata').rglob('*') if p.suffix in ['.a','.ts'] and not p.name.endswith('.d.ts'))
 def measure(path):
  row={'path':str(path.relative_to(root))}
  for phase,binary in binaries.items():
   r=subprocess.run([binary,'c',str(path)],capture_output=True,timeout=20);row[phase]=dict(exit=r.returncode,stderr=r.stderr.decode(errors='replace'))
  return row
 with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
  rows=[]
  for row in pool.map(measure,files):
   rows.append(row)
   if len(rows)%100==0:print('compared',len(rows),'of',len(files),flush=True)
 new=[r['path'] for r in rows if r['main']['exit']!=0 and r['after']['exit']==0];lost=[r['path'] for r in rows if r['main']['exit']==0 and r['after']['exit']!=0]
 reasons=['a value of type __String | undefined','a function returning __String | undefined'];counts={phase:{reason:sum("can't lower "+reason+' yet' in r[phase]['stderr'] for r in rows) for reason in reasons} for phase in ['main','after']}
 result=dict(files=len(rows),newly_admitted=new,lost_admissions=lost,first_error_counts=counts,binary_sha256=hashes,rows=rows);(out/'admission.json').write_text(json.dumps(result,indent=2));print(json.dumps({k:v for k,v in result.items() if k!='rows'},indent=2),flush=True)
 assert len(new)==6 and not lost,(new,lost)
 assert all(Path(p).name.startswith('string_brand_') for p in new)
