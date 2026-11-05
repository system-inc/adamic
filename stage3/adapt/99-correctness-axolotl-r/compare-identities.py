from pathlib import Path
import json,collections,hashlib
r=Path('/workspace/scratch/99r'); streams={}
for label in ['reference','counted','after']:
 report=json.loads((r/(label+'-oracle/report.json')).read_text());assert report['counts']=={'passing':106366,'failing':1,'pending':0};assert report['baseline_diffs']==['api/typescript.d.ts'];assert report['workers']==8 and report['runners']=='all' and report['tests'] is None
 streams[label]=collections.Counter((row['event'],row['title']) for line in (r/(label+'-identities.jsonl')).open() if (row:=json.loads(line)))
 assert all(n%2==0 for n in streams[label].values())
assert streams['reference']==streams['counted']==streams['after']
result={'allThreeMultisetsEqual':True,'rawPassEvents':212732,'rawFailEvents':2,'passing':106366,'failing':1,'pending':0,'added':0,'removed':0,'normalizedSha256':hashlib.sha256((r/'normalized-identities.jsonl').read_bytes()).hexdigest()}
(r/'identity-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
