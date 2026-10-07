import json
from pathlib import Path
names=['allocations','frees','retains','releases','peak live','in regions']
def rows(path):
 d={}
 for line in Path(path).read_text().splitlines():
  fields=[s.strip() for s in line.split('|')[1:-1]]
  if len(fields)==7 and all(s.isdigit() for s in fields[1:]):
   assert fields[0] not in d, fields[0]
   d[fields[0]]=list(map(int,fields[1:]))
 return d
old=rows('/workspace/scratch/seat16-runtime/area-counts.md');new=rows('internal/oracle/counts.md');assert not old.keys()-new.keys(), old.keys()-new.keys()
changed={p:{names[i]:[a,b] for i,(a,b) in enumerate(zip(old[p],v)) if a!=b} for p,v in new.items() if p in old and v!=old[p]}
rises={p:{k:v for k,v in changes.items() if v[1]>v[0]} for p,changes in changed.items()};rises={p:v for p,v in rises.items() if v}
result={'baseline_rows':len(old),'final_rows':len(new),'added':{p:v for p,v in new.items() if p not in old},'changed':changed,'rises':rises}
Path('/workspace/scratch/seat16-runtime/count-diff.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
