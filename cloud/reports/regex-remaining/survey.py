#!/usr/bin/env python3
"""Run independent serial runner shards; never share the TypeScript IPC stream."""
import argparse,concurrent.futures,json,pathlib,subprocess,collections
p=argparse.ArgumentParser();p.add_argument('--runner',required=True);p.add_argument('--root',required=True);p.add_argument('--test262',required=True);p.add_argument('--work',required=True);p.add_argument('--shards',type=int,default=3);a=p.parse_args()
prefixes=['built-ins/RegExp','language/literals/regexp','annexB/built-ins/RegExp','annexB/language/literals/regexp']
corpus=pathlib.Path(a.test262)/'test';paths=sorted(str(f.relative_to(corpus)) for prefix in prefixes for f in (corpus/prefix).rglob('*.js') if '_FIXTURE' not in f.name)
assert len(paths)==len(set(paths));assert a.shards>0
base=pathlib.Path(a.work);base.mkdir(parents=True,exist_ok=True)
def run(i):
 work=base/f'shard-{i}';command=[a.runner,'--adapt','--json','--jobs','1','--timeout','60s','--root',a.root,'--test262',a.test262,'--work',str(work),*paths[i::a.shards]]
 (base/f'shard-{i}.command.json').write_text(json.dumps(command,indent=2)+'\n')
 with (base/f'shard-{i}.json').open('w') as out,(base/f'shard-{i}.log').open('w') as err:
  subprocess.run(command,stdout=out,stderr=err,check=True)
 return [json.loads(s) for s in (work/'results.jsonl').read_text().splitlines()]
with concurrent.futures.ThreadPoolExecutor(max_workers=a.shards) as pool:rows=[r for result in pool.map(run,range(a.shards)) for r in result]
assert len(rows)==len(paths) and {r['path'] for r in rows}==set(paths)
(base/'results.jsonl').write_text(''.join(json.dumps(r)+'\n' for r in sorted(rows,key=lambda r:r['path'])))
print(collections.Counter(r['kind'] for r in rows))
if any(r['kind'] in ['fail','crashed','unrun'] for r in rows):raise SystemExit(1)
