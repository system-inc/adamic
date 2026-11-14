from pathlib import Path
import json,sys
out=Path(__file__).resolve().parent
records=[json.loads(s) for s in Path(sys.argv[1]).read_text().splitlines()]
files=[r for r in records if str(r.get('file','')).endswith('/src/compiler/sourcemap.ts')]
assert len(files)==1,[(r.keys(),r.get('file')) for r in records]
r=files[0]
findings=r.get('findings',[])
print(json.dumps({'units':len(r.get('units',[])),'diagnostics':records[0].get('diagnostics'), 'findings':findings,'flushUnits':[u for u in r.get('units',[]) if 'flushMappingBuffer' in str(u)]},indent=2))
if len(sys.argv)>2:
 other=[json.loads(s) for s in Path(sys.argv[2]).read_text().splitlines()]
 after=next(r for r in other if str(r.get('file','')).endswith('/src/compiler/sourcemap.ts'))
 old=b"mappingCharCodes.length = 0;";new=b"mappingCharCodes.splice(0);"
 source=Path('/workspace/scratch/sourcemap-after/src/compiler/sourcemap.ts').read_bytes()
 cut=source.index(new)+len(new);delta=len(old)-len(new)
 def norm(f,green=False):
  # Different scratch roots are provenance, not a finding change. Keep all other fields.
  def rec(v):
   if isinstance(v,str):return v.replace('/workspace/scratch/sourcemap-baseline','<tree>').replace('/workspace/scratch/sourcemap-after','<tree>')
   if isinstance(v,dict):return {k: (x+delta if green and '/src/compiler/sourcemap.ts:' in f.get('where','') and k in ('start','end','site_start','site_end') and isinstance(x,int) and x>=cut else rec(x)) for k,x in v.items()}
   if isinstance(v,list):return [rec(x) for x in v]
   return v
  return json.dumps(rec(f),sort_keys=True)
 red=[f for f in findings if 'assigning a non-tuple array\'s length' in f.get('reason','')]
 assert len(red)==2 and len({f["where"] for f in red})==1,red
 b={norm(f) for f in findings};a={norm(f,True) for f in after['findings']}
 removed=[f for f in findings if f.get('where')==red[0]['where'] and (f.get('kind')=='Refused' or "assigning a non-tuple array's length" in f.get('text',''))]
 assert b-a=={norm(f) for f in removed},[json.loads(s) for s in b-a]
 assert not a-b,[json.loads(s) for s in a-b]
 print(json.dumps({'removed':removed,'offsetNormalizationBytes':delta,'newOrChangedFindings':0,'before':len(b),'after':len(a)},indent=2))
