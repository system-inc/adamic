import json,collections,pathlib
root=pathlib.Path('/tmp/invoke-map-4pc8wd8/evidence')
def read(name):return json.loads((root/name).read_text())
def rows(name):
 rs=[json.loads(l) for l in (root/name).read_text().splitlines()]
 assert rs[0]['latent_mode']=='full' and rs[0]['module_format']=='CommonJSBundle'
 assert rs[-1]['status']=='relation_report'
 return rs
old,new=rows('main-census.jsonl'),rows('amendment-census.jsonl')
def findings(rs):return [f for r in rs for f in r.get('findings',[])]
a,b=findings(old),findings(new)
# Site identity retains phase, attempting unit, exact reason/text, and boundary spans.
# Byte offsets after the shorter annotation are mapped back to main's source.
oldsrc=(root/'main-sys.ts').read_bytes();newsrc=(root/'amendment-sys.ts').read_bytes()
pivot=oldsrc.index(b'const invokeMap = new Map<Path, string[]>();');delta=len(oldsrc)-len(newsrc)
def normalized(f,after=False):
 f=dict(f)
 if after and f.get("where", "").rsplit(":",2)[0].endswith("/compiler/sys.ts"):
  for key in ('start','end'):
   if isinstance(f.get(key),int) and f[key]>pivot:f[key]+=delta
 return f
ac=collections.Counter(json.dumps(normalized(f),sort_keys=True) for f in a)
bc=collections.Counter(json.dumps(normalized(f,True),sort_keys=True) for f in b)
gone=[json.loads(k) for k,n in (ac-bc).items() for _ in range(n)]
newfind=[json.loads(k) for k,n in (bc-ac).items() for _ in range(n)]
result={'main_observations':len(a),'amendment_observations':len(b),'offset_mapping':{'pivot':pivot,'main_minus_amendment_bytes':delta},'new':newfind,'gone':gone,'changed':[], 'note':'Changed means a matched site/phase/unit with altered finding content; byte offsets normalized for the annotation length.'}
# Pair replacements by site, phase, attempting unit and kind.
def key(f):return tuple(f.get(k) for k in ('where','phase','unit','kind'))
for g in gone[:]:
 match=next((f for f in newfind if key(f)==key(g)),None)
 if match:
  result['changed'].append({'main':g,'amendment':match});gone.remove(g);newfind.remove(match)
(root/'census-delta.json').write_text(json.dumps(result,indent=2)+'\n')
lines=[]
for label in ('new','gone','changed'):
 lines.append(f'{label}: {len(result[label])}')
 for f in result[label]:lines.append(json.dumps(f,ensure_ascii=False,sort_keys=True))
(root/'census-delta.txt').write_text('\n'.join(lines)+'\n')
print('\n'.join(lines))
assert not result['new'] and not result['changed']
assert any(':790:87' in f.get('where','') for f in result['gone'])
assert not any(':790:87' in f.get('where','') for f in b)
