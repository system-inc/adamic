import json,re
from pathlib import Path
out=Path('/workspace/scratch/path-adapt-final')
def read(label):
 records=[json.loads(l) for l in (out/('census-'+label+'.jsonl')).read_text().splitlines()]
 findings=[]
 for record in records:
  for f in record.get('findings',[]):
   def loc(where):return re.sub(r'^.*?/src/', 'src/',where)
   def unshift(where):
    where=loc(where)
    if label=='after' and '/path.ts:' in where:
     m=re.search(r':(\d+):(\d+)$',where)
     if m and int(m[1])>=502:where=where[:m.start()]+':'+str(int(m[1])-2)+':'+m[2]
    return where
   findings.append((f['kind'],unshift(f['where']),f['reason']))
 return records,sorted(set(findings))
before,b=read('before');after,a=read('after')
removed=[x for x in b if x not in a];added=[x for x in a if x not in b]
result={'beforeHeader':before[0],'afterHeader':after[0],'removedUniqueFindings':removed,'addedUniqueFindings':added,'afterCoordinatesNormalizedByTwoLines':True}
(out/'census-delta.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
