from pathlib import Path
import json,difflib,re,collections
s=Path('/tmp/watch-file-check-cead')
def rows(side):
 r=[json.loads(l) for l in (s/('census-'+side)/'raw.jsonl').read_text().splitlines()]
 assert r[0]['latent_mode']=='full' and r[0]['module_format']=='CommonJSBundle' and r[-1]['status']=='relation_report'
 return r
base=rows('main');base_text=(s/'sys-main.ts').read_text().splitlines()
def collect(side):
 candidate=(s/('sys-'+side+'.ts')).read_text().splitlines();mapping={}
 for tag,a1,a2,b1,b2 in difflib.SequenceMatcher(a=base_text,b=candidate,autojunk=False).get_opcodes():
  if tag=='equal' or tag=='replace' and a2-a1==b2-b1:
   mapping.update({j+1:a1+j-b1+1 for j in range(b1,b2)})
 def norm(value):
  value=value.replace('/tmp/watch-functions-1532/main-tree/src/compiler/','')
  return re.sub(r'sys\.ts:(\d+):(\d+)',lambda m:'sys.ts:'+str(mapping.get(int(m[1]),'NEW_'+m[1]))+':'+m[2],value)
 unique={}
 stream=rows(side)
 for row in stream:
  for f in row.get('findings',[]):
   normal={k:norm(f.get(k,'')) for k in ('where','phase','kind','reason','text')}
   key=tuple(normal.values())
   if key not in unique:unique[key]={'normalized':normal,'observations':[]}
   unique[key]['observations'].append(f)
 return stream,unique
_,before=collect('main')
for side in ['after']:
 if not (s/('census-'+side)/'result.json').exists():continue
 stream,after=collect(side);gone={k:before[k] for k in before.keys()-after.keys()};new={k:after[k] for k in after.keys()-before.keys()}
 group=lambda item:tuple(item['normalized'][k] for k in ('where','phase','kind'))
 changed=[]
 for location in set(map(group,gone.values()))&set(map(group,new.values())):
  old={k:v for k,v in gone.items() if group(v)==location};fresh={k:v for k,v in new.items() if group(v)==location}
  changed.append({'location':location,'before':list(old.values()),'after':list(fresh.values())})
  for k in old:del gone[k]
  for k in fresh:del new[k]
 target=lambda d,line:[v for v in d.values() if v['normalized']['where']==f'sys.ts:{line}:13' and v['normalized']['phase']=='refusal_scan']
 delta={'variant':side,'uniqueBefore':len(before),'uniqueAfter':len(after),'new':list(new.values()),'gone':list(gone.values()),'changed':changed,'target1532Gone':bool(target(before,1532)) and not target(after,1532),'sibling1533Unchanged':target(before,1533)[0]['normalized']==target(after,1533)[0]['normalized'],'checkerDiagnostics':stream[0]['diagnostic_sites'],'unitStatuses':{label:dict(collections.Counter(u['status'] for r in data for u in r.get('units',[]))) for label,data in [('before',base),('after',stream)]}}
 (s/('delta-'+side+'.json')).write_text(json.dumps(delta,indent=2)+'\n')
 print(json.dumps({k:len(v) if k in ('new','gone','changed') else v for k,v in delta.items() if k not in ('checkerDiagnostics','unitStatuses') and k not in ('uniqueBefore','uniqueAfter')}))
 for change in delta['new']: print('NEW',json.dumps(change['normalized']))
 for change in changed:print('CHANGED',json.dumps(change['location']))
