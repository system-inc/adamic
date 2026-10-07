import json,re,collections
from pathlib import Path
p=Path('/workspace/scratch/immortal-sites'); lines=(p/'runtime-parse.c').read_text().splitlines(); edges=json.loads((p/'sites.json').read_text()); buckets=collections.Counter(); samples=collections.defaultdict(list)
for e in edges:
 n=e['calls']; cost=e['cost']+[0]*5
 if cost[0]!=6*n or cost[1]!=2*n:continue
 source=e['source']; m=re.search(r'adamic_release\((\w+)\)',source); cat='unresolved'; defs=[]
 if m:
  var=m[1]; start=max(0,e['line']-1)
  while start>0 and not lines[start].startswith('static '):start-=1
  defs=[(i+1,l.strip()) for i,l in enumerate(lines[start:e['line']],start) if re.search(r'\b'+var+r'\s*=',l)]
  joined=' '.join(l for _,l in defs)
  if re.search(r'adamic_function_(73_Parser_kind|12_kind)\(',joined):cat='scanner.kind getter result'
  elif '"kind"' in joined:cat='kind field read'
  elif '&adamic_string_' in joined:cat='literal or literal-containing definition'
  elif defs:
   last=defs[-1][0]; window=' '.join(lines[max(start,last-6):last+1])
   if re.search(r'void \*'+var+r' = (\w+)->reference;',joined):
    slot=re.search(r'void \*'+var+r' = (\w+)->reference;',joined)[1]
    slotdefs=[l for l in lines[start:last] if re.search(r'\*'+slot+r' =',l)]
    if any('"kind"' in l for l in slotdefs):cat='old kind field on store'
   elif re.search(r'adamic_object_callee\([^;]*"kind"',window):cat='dynamic kind getter result'
  e['definitions']=defs
 e['category']=cat;buckets[cat]+=n;samples[cat].append(e)
print(dict(buckets),'total',sum(buckets.values()))
for cat,rows in samples.items():
 print('\n',cat)
 for e in sorted(rows,key=lambda e:e['calls'],reverse=True)[:12]:print(e['calls'],e['fn'],e['line'],e['source'],e.get('definitions'))
(p/'classified.json').write_text(json.dumps(dict(counts=buckets,sites=dict(samples)),indent=2)+'\n')
