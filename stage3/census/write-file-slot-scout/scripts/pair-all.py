from pathlib import Path
import json,re,collections
r=Path('/workspace/scratch/write-file-slot-measure')
def norm(s):
 s=re.sub(r'^.*?\.(?:ts|a):\d+:\d+:\s*','',s)
 s=re.sub(r'/workspace/scratch/write-file-slot-measure/(?:main|a-construction|b-slot)/','<tree>/',s)
 return re.sub(r'(?<=\.ts):\d+:\d+',':<pos>',s)
def read(n):
 rows=list(map(json.loads,(r/(n+'.jsonl')).read_text().splitlines()));f=[f for x in rows for f in x.get('findings',[])]
 # Preserve every observation, including Boundary records and repeated attempts.
 counts=collections.Counter((x['kind'],x['phase'],x['reason'],norm(x['text'])) for x in f)
 return f,counts
base,b=read('main');out={}
for n in ['main','a-construction','b-slot']:
 f,c=read(n);out[n]={'rawRecords':len(f),'distinctNormalizedTuples':len(c),'added':[{'key':list(k),'count':v} for k,v in (c-b).items()],'removed':[{'key':list(k),'count':v} for k,v in (b-c).items()]}
print(json.dumps(out,indent=2));(r/'all-record-pairs.json').write_text(json.dumps(out,indent=2)+'\n')
