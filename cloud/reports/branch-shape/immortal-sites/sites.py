import re,json,collections
from pathlib import Path
p=Path('/workspace/scratch/immortal-sites')
sym={'fn':{},'fl':{}}; fn=fl=callee='?'; pos=[0,0]; pending=None; edges=[]
for row in (p/'current.callgrind').read_text().splitlines():
 if row.startswith(('fn=','fl=','fi=','fe=','cfn=','cfl=','cfi=','cfe=','jfi=','jfl=','jfn=')):
  k,v=row.split('=',1); kind='fn' if k.endswith('fn') else 'fl'; m=re.fullmatch(r'\((\d+)\)(?: (.*))?',v)
  if m:
   i,n=m.groups()
   if n is not None:sym[kind][i]=n
   v=sym[kind][i]
  if k=='fn':fn=v
  elif k in ('fl','fi','fe'):fl=v
  elif k=='cfn':callee=v
 elif row.startswith('calls='):pending=(callee,int(row.split()[0][6:]))
 elif row and row[0] in '0123456789+-*':
  fields=row.split()
  for i,v in enumerate(fields[:2]):
   if v!='*':pos[i]=pos[i]+int(v,0) if v[0] in '+-' else int(v,0)
  if pending:
   target,n=pending; pending=None
   if target=='adamic_release':edges.append(dict(fn=fn,file=fl,address=hex(pos[0]),line=pos[1],calls=n,cost=list(map(int,fields[2:]))))
lines=(p/'runtime-parse.c').read_text().splitlines()
for e in edges:
 e['source']=lines[e['line']-1].strip() if 'runtime-parse.c' in e['file'] and 0<e['line']<=len(lines) else ''
(p/'sites.json').write_text(json.dumps(edges,indent=2)+'\n')
print('calls',sum(e['calls'] for e in edges),'edges',len(edges))
for e in sorted(edges,key=lambda e:e['calls'],reverse=True)[:45]:print(e)
