import json,sys,collections,pathlib
before,after,out=sys.argv[1:]
def read(file):
 rows=[]
 for line in open(file):
  d=json.loads(line)
  if 'file' not in d:continue
  tree=d['file'].split('/src/')[0]
  for row in d['findings']:
   row=json.loads(json.dumps(row).replace(tree,'<tree>'))
   # Line/column, phase, kind, unit and full diagnostic text identify findings.
   # Byte spans after the shortened condition move mechanically; retain them in raw JSONL.
   row.pop('start',None);row.pop('end',None)
   rows.append(row)
 return rows
b,a=read(before),read(after)
key=lambda r:json.dumps(r,sort_keys=True)
bc,ac=collections.Counter(map(key,b)),collections.Counter(map(key,a))
gone=[json.loads(k) for k,n in (bc-ac).items() for _ in range(n)]
new=[json.loads(k) for k,n in (ac-bc).items() for _ in range(n)]
identity=lambda r:tuple(r.get(k) for k in ['where','phase','kind','unit'])
changed=[]
for old in gone[:]:
 match=next((n for n in new if identity(n)==identity(old)),None)
 if match is not None:
  changed.append({'before':old,'after':match});gone.remove(old);new.remove(match)
result={'before':len(b),'after':len(a),'new':new,'gone':gone,'changed':changed,'counts':{'new':len(new),'gone':len(gone),'changed':len(changed)},'identity':'full fields and diagnostic text; absolute tree prefix normalized; mechanical byte spans omitted (raw retained)'}
pathlib.Path(out).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result['counts']))
