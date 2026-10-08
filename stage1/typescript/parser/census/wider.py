"""Run disjoint corpus shards concurrently; no repository scripts or installs."""
import pathlib,subprocess,concurrent.futures,json,gzip,shutil
base=pathlib.Path(__file__).resolve().parent
rows=json.loads((base/'fetch-results.json').read_text());paths=[f for row in rows for f in row.get('files',[])]
# Balance file count; the manifest remains the definitive inventory.
shards=[paths[i::4] for i in range(4)]
def run(i):
 mf=base/f'public-{i}.manifest';mf.write_text('\n'.join(shards[i])+'\n')
 with (base/f'public-{i}.log').open('w') as out:
  p=subprocess.run(['python3',str(base/'run.py'),str(mf),str(base/f'public-{i}.json')],stdout=out,stderr=subprocess.STDOUT)
 if p.returncode:raise RuntimeError(f'Shard {i} failed')
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:list(pool.map(run,range(4)))
result={'files':0,'parsed':0,'identical':0,'shape_identical':0,'classes':{},'failures':[],'records':[]}
for i in range(4):
 d=json.loads((base/f'public-{i}.json').read_text())
 for k in ['files','parsed','identical','shape_identical']:result[k]+=d[k]
 for k in ['failures','records']:result[k]+=d[k]
 for c,v in d['classes'].items():
  old=result['classes'].get(c);count=(old['files'] if old else 0)+v['files']
  if old is None or v['shortest_bytes']<old['shortest_bytes']:
   result['classes'][c]=v.copy();folder=base/'public/examples';folder.mkdir(parents=True,exist_ok=True)
   for src in (base/f'public-{i}/examples').glob(c.replace('/','-')+'.*'):shutil.copyfile(src,folder/src.name)
  result['classes'][c]['files']=count
result['classes']=dict(sorted(result['classes'].items(),key=lambda x:(-x[1]['files'],x[0])))
(base/'public.json').write_text(json.dumps(result,indent=2)+'\n')
print({k:v for k,v in result.items() if k not in ['records','classes','failures']},flush=True)
