import gzip
import hashlib
import json
from pathlib import Path
lane=Path(__file__).resolve().parent
root=lane.parents[3]
source=root/'stage3/interface-downcasts/lazy/census/read-demand-pairs.json.gz'
rows=[r for r in json.loads(gzip.decompress(source.read_bytes())) if 'object plus primitive union' in r['families']]
rows.sort(key=lambda r:(-r['read_count'],r['type_id'],r['field']))
assert len(rows)==42 and sum(r['read_count'] for r in rows)==181
for index,row in enumerate(rows,1):row['rank']=index+0;row['status']='pending original-pair witness'
result={'basis':'lazy static candidates; exact production reachability unmeasured','integration_sha':'ba59427ccc7afecae29a305c41e6e9c7867e5610','source_sha256':hashlib.sha256(source.read_bytes()).hexdigest(),'pairs':42,'candidate_reads':181,'completed_pairs':0,'completed_reads':0,'rows':rows}
(lane/'lazy-pair-progress.json').write_text(json.dumps(result,indent=2)+'\n')
print('42 candidate pairs / 181 candidate reads; exact reachability unmeasured')
for row in rows: print(row['rank'],row['read_count'],row['type'],row['field'],row['declared_type'])
