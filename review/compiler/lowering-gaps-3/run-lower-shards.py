from pathlib import Path
import subprocess,json,concurrent.futures
out=Path('review/compiler/lowering-gaps-3')
names=[s for s in (out/'lower-test-list.log').read_text().splitlines() if s.startswith('Test')]
shards=[names[i::12] for i in range(12)]
def run(pair):
 i,names=pair;cmd=['timeout','120','go','test','./internal/lower','-run','^('+'|'.join(names)+')$','-count=1','-timeout=90s','-json']
 with (out/('lower-shard-'+str(i)+'.jsonl')).open('wb') as log: p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 return dict(shard=i,tests=names,exit=p.returncode)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:results=list(pool.map(run,enumerate(shards)))
(out/'lower-shards.json').write_text(json.dumps(results,indent=2)+'\n')
print('shards',len(results),'tests',len(names),'failures',[r['shard'] for r in results if r['exit']])
