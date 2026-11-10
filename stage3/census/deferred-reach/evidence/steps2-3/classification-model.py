import json
from pathlib import Path
from collections import Counter,deque
w=Path('/workspace/scratch/abrq61p');rel=lambda s:'src/'+s.split('/src/',1)[1] if '/src/' in s else s
rows=[json.loads(x) for x in (w/'base.jsonl').read_text().splitlines()]
units={rel(u['where']):u for r in rows for u in r.get('units',[])}
refs={u['where']:u for u in json.loads((w/'classification-references.json').read_text())['units']}
primary_files={k.rsplit(':',2)[0] for k in units}
first={}
for r in rows:
 for f in r.get('findings',[]):
  if f.get('phase') in ('lowering','deferred'):
   first.setdefault(rel(f['unit']),{k:rel(v) if k in ('unit','where') else v for k,v in f.items()})

def census_owner(call):
 top=call['top_unit'];unit=units.get(top)
 generic=next((f['where'] for f in call.get('functions',[]) if f['kind']=='FunctionDeclaration' and f['type_parameters']),None)
 if call.get('external_callback'):return 'outside',None
 if generic:return 'generic',generic
 if unit and unit['status'] in ('attempted','panic','speculative'):
  if unit['kind'] in ('KindModuleDeclaration','KindClassDeclaration') and call.get('functions'):return 'outside',None
  return 'root',top
 return 'outside',None

def callers(where):
 # Reverse generic-call graph. Nongeneric nested declarations are eagerly
 # prepared by their containing attempt; generic boundaries need instantiation.
 roots={};outside=[];todo=deque([(where,[])]);seen=set()
 while todo:
  node,chain=todo.popleft()
  if node in seen:continue
  seen.add(node)
  for c in refs.get(node,{}).get('calls',[]):
   role,next_node=census_owner(c);edge=dict(c,target=node);path=[edge,*chain]
   if role=='root':
    item=dict(unit=next_node,entry_call=c['call_site'],route=path,stop=first.get(next_node))
    old=roots.get(next_node)
    key=lambda x:(len(x['route']),x['entry_call'],json.dumps(x['route'],sort_keys=True))
    if old is None or key(item)<key(old):roots[next_node]=item
   elif role=='generic' and next_node in refs and next_node.rsplit(':',2)[0] in primary_files:
    if next_node not in seen:todo.append((next_node,path))
    else:outside.append(dict(edge,reason='deferred recursive boundary'))
   else:outside.append(dict(edge,reason='no independently lowered caller body' if not c.get('external_callback') else 'host callback registration'))
 return sorted(roots.values(),key=lambda x:x['unit']),outside

requests=[]
for line in (w/'head-requests.jsonl').read_text().splitlines():
 try:requests.append({k:rel(v) for k,v in json.loads(line).items()})
 except json.JSONDecodeError:pass
actual={}
for r in requests:
 actual.setdefault(r['target'],[]).append(r)
 target=refs.get(r['target'],{})
 implementation=target.get('implementation')
 if target.get('implementation_generic') and implementation and implementation != r['target']:
  actual.setdefault(implementation,[]).append(dict(r,via_signature=r['target']))
result=[]
for where,u in units.items():
 if u['status']!='deferred_generic_instantiation':continue
 info=refs[where];root_callers,outside=callers(where) if info['has_body'] and not u['instantiated'] else ([],[]);observed=actual.get(where,[])
 if not info['has_body']:classification='signature only, no body'
 elif u['instantiated']:classification='reached'
 elif observed:classification=None
 elif not info['references']:classification='unreached with no call anywhere'
 elif not root_callers:classification='unreached, called only from outside the census'
 elif all(c['stop'] is not None for c in root_callers):classification='unreached hidden behind a census caller\'s earlier stop'
 else:classification=None
 result.append(dict(where=where,name=u.get('name'),has_body=info['has_body'],raw_instantiated=u['instantiated'],classification=classification,census_callers=root_callers,outside_callers=outside,outside_caller_files=sorted({x['call_site'].rsplit(':',2)[0] for x in outside}),requests=observed,reference_count=len(info['references']),direct_or_alias_calls=info['calls']))
(w/'classification-draft.json').write_text(json.dumps(result,indent=2)+'\n')
for title,subset in [('all',result),('core',[x for x in result if x['where'].startswith('src/compiler/core.ts:')])]:
 print(title,Counter(x['classification'] for x in subset))
print('unclassified',[(r['where'],r['name'],len(r['requests'])) for r in result if r['classification'] is None])
