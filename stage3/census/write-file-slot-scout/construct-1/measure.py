from pathlib import Path
import subprocess,os,json,time,concurrent.futures,collections
r=Path('/workspace/scratch/write-file-construct-run');bases={'control':r/'control-adapted','candidate':r/'lane/adapted-tree'}
while not (r/'root-census').exists():time.sleep(3)
summary={};start=time.monotonic()
for n,tree in bases.items():
 if n=='candidate':
  while not (r/'lane/verdict.txt').exists() and not (r/'lane.exit').exists():time.sleep(5)
  assert (tree/'patch-set.md').exists()
 roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 def run(i):
  chosen=roots[i::4];manifest=r/f'root-{n}-{i}.manifest.json';manifest.write_text(json.dumps([str(p) for p in chosen],indent=2)+'\n')
  env=dict(os.environ,CENSUS_MODULE_FORMAT='CommonJSBundle',SCOUT_ROOT_MANIFEST=str(manifest),TMPDIR=str(r/'tmp'),GOTMPDIR=str(r/'tmp'),GOCACHE='/workspace/scratch/write-file-slot-measure/gocache',GOMEMLIMIT='2GiB')
  env.pop('ADAMIC_PROGRAM_REGION',None);env.pop('ADAMIC_OWNERSHIP_QUERY',None)
  with (r/f'root-{n}-{i}.log').open('w') as log:p=subprocess.run([str(r/'root-census'),str(tree/'src/compiler'),str(r/f'root-{n}-{i}.jsonl')],env=env,stdout=log,stderr=subprocess.STDOUT)
  assert p.returncode==0,(n,i,p.returncode)
  rows=[json.loads(s) for s in (r/f'root-{n}-{i}.jsonl').read_text().splitlines()];assert len(rows)==len(chosen);assert [x['roots'] for x in rows]==[[str(p)] for p in chosen];return rows
 (r/'root-profile-status.json').write_text(json.dumps({'phase':n})+'\n')
 with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=[x for shard in pool.map(run,range(4)) for x in shard]
 rows.sort(key=lambda x:x['roots'][0]);(r/f'roots-{n}.json').write_text(json.dumps(rows,indent=2)+'\n')
 counts=collections.Counter((x['kind'],x.get('where','').replace(str(tree)+'/', ''),x.get('reason','')) for x in rows)
 summary[n]={'roots':79,'groups':[{'count':v,'kind':k[0],'where':k[1],'reason':k[2]} for k,v in sorted(counts.items(),key=lambda kv:(-kv[1],kv[0]))]};print(n,json.dumps(summary[n]),flush=True)
summary['wallSeconds']=time.monotonic()-start;(r/'root-profile.json').write_text(json.dumps(summary,indent=2)+'\n');(r/'root-profile-status.json').write_text('{"phase":"complete"}\n')
