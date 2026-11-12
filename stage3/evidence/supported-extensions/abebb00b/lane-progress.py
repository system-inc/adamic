import json,collections
from pathlib import Path
p=Path(__file__).resolve().parent
counts=collections.Counter();fail=[];end=None;stress=None
for line in (p/'lane-whole.log').read_text().splitlines():
 try:e=json.loads(line)
 except json.JSONDecodeError:continue
 t=e.get('Test','')
 if t.startswith('TestShard_') and '/' not in t:counts[e['Action']]+=1
 if t=='TestShard_Stage3LaneFile000' and e.get('Action') in ['pass','fail']:stress=e['Action']
 if e.get('Action')=='fail':fail.append(t or e.get('Package'))
 if e.get('Action') in ['pass','fail'] and not t:end=e['Action']
print(json.dumps({'shards':dict(counts),'failures':fail,'packageVerdict':end,'stressShardVerdict':stress}))
