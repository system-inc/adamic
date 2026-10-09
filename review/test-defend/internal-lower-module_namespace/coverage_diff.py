import pathlib,json,re
p=pathlib.Path('review/test-defend/internal-lower-module_namespace')
def lines(name):
 out=set()
 for s in (p/'coverage'/(name+'.out')).read_text().splitlines()[1:]:
  a,_,n=s.split(); file,span=a.rsplit(':',1); start,end=span.split(','); lo=int(start.split('.')[0]);hi=int(end.split('.')[0])
  if int(n)>0:
   out.update((file,i) for i in range(lo,hi+1))
 return out
out=[]
for row in json.loads((p/'audit-rows.json').read_text()):
 name=row['test']; other=row['subsumed_by'][0] if row['subsumed_by'] else 'REST-'+name
 if not (p/'coverage'/(other+'.out')).exists():continue
 exclusive=sorted(lines(name)-lines(other))
 out.append({'test':name,'comparison':other,'exclusive_lines':[f'{f}:{i}' for f,i in exclusive]})
 print(name,len(exclusive))
 for f,i in exclusive:
  if any(s in f for s in ['namespace','modules.go','load_time']): print(' ',f.rsplit('/',1)[-1],i)
(p/'coverage-exclusive.json').write_text(json.dumps(out,indent=2))
