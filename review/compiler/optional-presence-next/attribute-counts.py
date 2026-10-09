import json,subprocess,pathlib,time,os
out=pathlib.Path('review/compiler/optional-presence-next');paths=[r['path'] for r in json.loads((out/'counts-rows.json').read_text()) if r['merged']!=r['measured']];rows=[];scratch=pathlib.Path('/workspace/scratch/optional-counts');scratch.mkdir(exist_ok=True)
for path in paths:
 row={'path':path}
 for label,compiler in [('base','/workspace/scratch/optional-base-adamic'),('chain','/workspace/scratch/optional-chain-adamic')]:
  name=path.replace('/','-');binary=str(scratch/(label+'-'+name));log=out/(label+'-'+name+'.log');start=time.monotonic()
  with log.open('w') as f:
   build=subprocess.run(['timeout','60',compiler,'build',path,'-o',binary,'--count'],stdout=f,stderr=subprocess.STDOUT,timeout=65);assert build.returncode==0,(path,label,log.read_text())
   run=subprocess.run(['/bin/sh','-c','ulimit -s 8192 && exec "$0"',binary],capture_output=True,text=True,timeout=60);f.write(run.stdout+run.stderr)
  line=next(s for s in run.stderr.splitlines() if s.startswith('adamic: counts:'));row[label]=dict(counts=line,exit=run.returncode,seconds=round(time.monotonic()-start,3))
 rows.append(row);(out/'counts-parent-observations.json').write_text(json.dumps(rows,indent=2)+'\n');print(row,flush=True)
