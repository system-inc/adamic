from pathlib import Path
import re,json
root=Path('/workspace/scratch/cast-tags/work/stage3/adapt')
names=['CodeBlock','ExceptionBlock','LabeledBlock','SwitchBlock','LoopBlock','WithBlock','TypePredicate','TypePredicateBase','ThisTypePredicate','IdentifierTypePredicate','AssertsThisTypePredicate','AssertsIdentifierTypePredicate','Mutable','createBinder','bindContainer','createNodeFactory','createUnionOrIntersectionTypeNode','createUnionTypeNode','createIntersectionTypeNode','createSyntacticTypeNodeBuilder','tryReuseExistingTypeNode','visitExistingNodeTreeSymbolsWorker']
rows=[]
for p in sorted(root.rglob('*')):
 if p.name not in ['owners.json','targets.json','adapt.cjs']:continue
 text=p.read_text(); hits=[n for n in names if re.search(r'\b'+n+r'\b',text)]
 if not hits:continue
 row={'file':str(p.relative_to(root)),'hits':hits,'pins':[]}
 if p.suffix=='.json':
  def scan(x):
   if isinstance(x,dict):
    identity={k:v for k,v in x.items() if k in ['file','owner','path','kind','ownerBeforeSha256','ownerAfterSha256','beforeSha256','afterSha256']}
    if any(n in str(identity) for n in names):row['pins'].append(identity)
    for v in x.values():scan(v)
   elif isinstance(x,list):
    for v in x:scan(v)
  scan(json.loads(text))
 rows.append(row)
Path('/workspace/scratch/cast-tags/evidence/pin-scout.json').write_text(json.dumps(rows,indent=2)+'\n')
for row in rows:
 print(row['file'],','.join(row['hits']))
 for pin in row['pins']:print(' ',pin)
