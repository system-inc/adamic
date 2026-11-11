#!/usr/bin/env python3
"""Render exact refusal reasons/source slices with per-case readings and owner units."""
import collections,gzip,json,sys,re
from pathlib import Path
root=Path(__file__).resolve().parent

def load(p):
 if not p.exists():p=Path(str(p)+'.gz')
 op=gzip.open if p.suffix=='.gz' else open
 with op(p,'rt') as f:return json.load(f)
def rows(p):
 if not p.exists():p=Path(str(p)+'.gz')
 op=gzip.open if p.suffix=='.gz' else open
 with op(p,'rt') as f:return [json.loads(x) for x in f if x.strip()]
def short(s):return s.replace('/workspace/roads-adapted/','')
def reading(r):
 steps=r.get('steps',[]);identities=[s['source'] for s in steps if s['source']==s['target']]
 if r['source']==r['target']:identities.insert(0,r['source'])
 trusted={'Modifier','BindingPattern','Node','Node | undefined','Declaration','Symbol','HasJSDoc','EmitNode | undefined','JSDoc','BindingElement | ParameterDeclaration | VariableDeclaration'}
 declared=set(re.findall(r'\b(?:interface|type|enum|class)\s+(\w+)',Path('/workspace/roads-adapted/src/compiler/types.ts').read_text()))
 identities=[x for x in identities if x in trusted or (' | ' in x and all(part in declared|{'undefined','null'} for part in x.split(' | ')))]
 terminal=steps[-1] if steps else {'source':r['source'],'target':r['target']}
 if identities and any(s['road']=='union arms' for s in steps):
  same='Modifier' if 'Modifier' in identities else identities[0]
  return 'guard bug: identity/union cross-product',f"The reported path first relates `{same}` to the same `{same}` type. That identical slot admits exactly the source's values; writing through it adds no value its source reader cannot take. The leaf `{terminal['source']} as {terminal['target']}` is disjoint, but the head manufactured that requirement while traversing union alternatives under this unchanged slot. It is the identity-union failure reproduced by identity.ts.txt, not evidence that this view introduces that incompatible store. This judges the reported refusal path; it does not claim that all other paths in the view are harmless."
 if len(steps)==1 and steps[0]['step']=='a source arm':
  if r.get('narrows') and 'SyntaxKind.' in r['text']:
   return 'guard bug: excluded discriminator arm',f"This predicate checks node kinds and preserves their discriminated alternatives. The concrete recorded target is `{r['target']}`; the rejected `{terminal['source']}` has a different readonly kind. It cannot enter this particular target alternative. If the predicate also admits that source kind, it belongs to its own target alternative instead. The leaf types are disjoint, but the guard incorrectly requires every source alternative to fit this individual target, losing that correlation. No write through this target alternative is a store into the other kind."
  if r['text']=='test(value)' and r['target'] in {'ObjectLiteralExpression','NamespaceImport','NamespaceExport'}:
   return 'guard bug: checked tryCast alternative',f"The site is tryCast's `test(value)`. The enclosing caller supplies the corresponding node-kind predicate before tryCast returns `{r['target']}` (for example getTsConfigObjectLiteralExpression supplies isObjectLiteralExpression, and getNamespaceDeclarationNode supplies isNamespaceImport/isNamespaceExport). `{terminal['source']}` is an excluded alternative, not a receiver admitted by that successful check. The leaf types are disjoint; demanding this cross-arm pair is a false refusal."
 return 'needs reading','No automatic judgment. This case must be read before the report is final.'

def main():
 prefix=sys.argv[1] if len(sys.argv)>1 else 'guard';full=prefix!='guard'
 records=[]
 if full:
  for side in ['parent','head']:
   for r in rows(root/'raw'/f'{prefix}-{side}-records.jsonl'):
    if r['kind']=='GuardRefusal':records.append(dict(r,side=side))
 else:records=[dict(r,side='head') for r in load(root/f'{prefix}-new-records.json')]
 new=load(root/f'{prefix}-new-records.json');newkeys={(r['where'],r['reason'],r['text']) for r in new}
 groups=collections.defaultdict(list)
 for r in records:groups[(r['road'],r['where'],r['source'],r['target'],r['reason'],r['text'])].append(r)
 manual_path=root/'manual-readings.json';manual=json.loads(manual_path.read_text()) if manual_path.exists() else {}
 lines=[f'# {prefix}: '+('every guard finding in full mode' if full else 'every new first-error guard finding'),'', 'Reasons and type names are verbatim. Source slices retain their text, with CRLF displayed as LF. Repeated observations are grouped only when site, source, target, reason and source text agree; every owner unit is listed. Road is the first new mechanism in the actual path; the exact reason preserves every subsequent mechanism.', '']
 index=0;unread=[];counts=collections.Counter();catalogue=[]
 for road in ['elements','index values','union arms','base pair check']:
  lines+=['## '+road,'']
  selected=sorted((k,v) for k,v in groups.items() if k[0]==road)
  if not selected:lines+=['None.',''];continue
  for key,observations in selected:
   r=observations[-1];index+=1;identifier=f'{prefix}-{index:04d}'
   isnew=(r['where'],r['reason'],r['text']) in newkeys
   category,text=reading(r) if isnew else ('unchanged/base finding','Quoted for the complete full-mode check; this is not a newly added refusal.')
   manualkey=' | '.join([r['where'],r['source'],r['target'],r['reason']])
   if manualkey in manual:category,text=manual[manualkey]['category'],manual[manualkey]['reading']
   counts[category]+=1
   if category=='needs reading':unread.append(dict(r,case=identifier,manual_key=manualkey))
   units=sorted({x['unit'] for x in observations});sides=collections.Counter(x['side'] for x in observations)
   lines += [f'### {identifier}: {short(r["where"])}','',f'Observations: {dict(sides)}. New in head: {isnew}.','', 'Source and target as named by the guard:', '', '~~~~text', r['source']+' as '+r['target'], '~~~~','', 'Exact reason:', '', '~~~~text',r['reason'],'~~~~','', 'Source text:', '', '~~~~typescript',r['text'].replace('\r\n','\n'),'~~~~','',f'Reading: **{category}.** {text}','', 'Owner units:', '']
   lines += ['- `'+short(u)+'`' for u in units];lines += ['']
   catalogue.append({'case':identifier,'road':road,'where':r['where'],'source':r['source'],'target':r['target'],'new':isnew,'reading_category':category,'reading':text,'units':units,'observations':dict(sides),'reason':r['reason']})
 (root/f'{prefix}-findings.md').write_text('\n'.join(lines)+'\n')
 (root/f'{prefix}-readings.json').write_text(json.dumps(catalogue,indent=2)+'\n')
 (root/f'{prefix}-unread.json').write_text(json.dumps(unread,indent=2)+'\n')
 print(json.dumps({'prefix':prefix,'cases':index,'categories':dict(counts),'needs_reading':len(unread)}))
if __name__=='__main__':main()
