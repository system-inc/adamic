from pathlib import Path
import subprocess,os,json,time,concurrent.futures,collections
r=Path('/workspace/scratch/write-file-slot-aprime');bases={'main':Path('/workspace/scratch/write-file-slot-measure/main'),'candidate':r/'candidate'}
status=r/'root-profile-status.json'
status.write_text(json.dumps({'phase':'waiting_for_uninstrumented_lane_and_full_census'})+'\n')
while not (r/'candidate-lane/verdict.txt').exists() or not (r/'full-sys.exit').exists():time.sleep(5)
summary={};start=time.monotonic()
for n,tree in bases.items():
 roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 def run(i):
  chosen=roots[i::4];manifest=r/f'root-{n}-{i}.manifest.json';manifest.write_text(json.dumps([str(p) for p in chosen],indent=2)+'\n')
  env=dict(os.environ,CENSUS_MODULE_FORMAT='CommonJSBundle',SCOUT_ROOT_MANIFEST=str(manifest),TMPDIR=str(r/'tmp'),GOTMPDIR=str(r/'tmp'),GOCACHE='/workspace/scratch/write-file-slot-measure/gocache',GOMEMLIMIT='3GiB')
  env.pop('ADAMIC_PROGRAM_REGION',None);env.pop('ADAMIC_OWNERSHIP_QUERY',None)
  cmd=[str(r/'root-census'),str(tree/'src/compiler'),str(r/f'root-{n}-{i}.jsonl')]
  with (r/f'root-{n}-{i}.log').open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  assert p.returncode==0,(n,i,p.returncode)
  rows=[json.loads(s) for s in (r/f'root-{n}-{i}.jsonl').read_text().splitlines()];assert len(rows)==len(chosen);assert [x['roots'] for x in rows]==[[str(p)] for p in chosen]
  return rows
 status.write_text(json.dumps({'phase':n,'started':start})+'\n')
 with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=[x for shard in pool.map(run,range(4)) for x in shard]
 rows.sort(key=lambda x:x['roots'][0]);assert len(rows)==79
 (r/f'roots-{n}.json').write_text(json.dumps(rows,indent=2)+'\n')
 counts=collections.Counter((x['kind'],x.get('where','').replace(str(tree)+'/', ''),x.get('reason','')) for x in rows)
 summary[n]={'roots':79,'groups':[{'count':v,'kind':k[0],'where':k[1],'reason':k[2]} for k,v in sorted(counts.items(),key=lambda kv:(-kv[1],kv[0]))]};print(n,json.dumps(summary[n]),flush=True)
main=json.loads((r/'roots-main.json').read_text());candidate=json.loads((r/'roots-candidate.json').read_text());m={Path(x['roots'][0]).relative_to(bases['main']/'src/compiler').as_posix():x for x in main};c={Path(x['roots'][0]).relative_to(bases['candidate']/'src/compiler').as_posix():x for x in candidate}
former=[k for k,x in m.items() if x.get('where','').endswith('/sys.ts:1387:5') and 'writeFile would lose' in x.get('reason','')]
summary['former1387']={'count':len(former),'routes':[{'root':k,'afterKind':c[k]['kind'],'afterWhere':c[k].get('where','').replace(str(bases['candidate'])+'/',''),'afterReason':c[k].get('reason','')} for k in former]}
summary['wallSeconds']=time.monotonic()-start;(r/'root-profile.json').write_text(json.dumps(summary,indent=2)+'\n');status.write_text(json.dumps({'phase':'complete','wallSeconds':summary['wallSeconds']})+'\n');print('COMPLETE',len(former),'former target roots',flush=True)
