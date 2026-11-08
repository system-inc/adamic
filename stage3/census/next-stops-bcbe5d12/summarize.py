import json,re
from pathlib import Path
out=Path(__file__).resolve().parent
rows=sum([json.loads((out/f'group-{i}-rows.json').read_text()) for i in range(1,5)],[])
assert len(rows)==79
old=json.loads((out/'previous-grouped-first-stops.json').read_text())
previous={r:(g['where'],g['kind'],g['reason']) for g in old for r in g['roots']}
groups={};comparison=[]
forecasts={'core.ts:2116':'core.ts:2116 .compare value read','debug.ts:537':'debug.ts:537 cast','debug.ts:577':'debug.ts:577 cast','semver.ts:287':'semver.ts:287','performanceCore.ts:29':'performanceCore.ts:29','performance.ts:189':'performance.ts:189'}
for row in rows:
 root=row['roots'][0].split('/src/compiler/')[1];where=row.get('where','-').split('/src/compiler/')[-1];kind=row['kind'];reason=row.get('reason','-');key=(where,kind,reason)
 forecast=next((v for k,v in forecasts.items() if where.startswith(k+':')),None) if kind!='accepted' else 'not applicable, lowers whole'
 if forecast is None and kind!='accepted' and where!='-':
  match=re.match(r'(.*):(\d+):(\d+)$',where)
  source=Path('/workspace/bcbe5d12-adapted/src/compiler')/match[1]
  lines=source.read_text().splitlines();line=lines[int(match[2])-1]
  if match[1]=='core.ts' and '.compare' in line and 'new Intl.Collator' in line:
   forecast=forecasts['core.ts:2116']+' (shifted coordinates)'
  if match[1]=='debug.ts' and 'Object.create(Object.prototype) as FlowNode' in line:
   forecast=forecasts['debug.ts:537']+' (shifted coordinates)'
  if match[1]=='debug.ts' and 'Object.create(Array.prototype) as NodeArray<Node>' in line:
   forecast=forecasts['debug.ts:577']+' (shifted coordinates)'
 g=groups.setdefault(key,dict(where=where,kind=kind,reason=reason,roots=[],forecast=forecast or 'none, new',new_group=forecast is None and kind!='accepted'))
 g['roots'].append(root)
 comparison.append(dict(root=root,previous=list(previous[root]),current=list(key),change='unchanged' if previous[root]==key else 'moved',forecast=g['forecast']))
for g in groups.values():
 g['count']=len(g['roots']);g['unchanged']=sum(r['change']=='unchanged' for r in comparison if r['root'] in g['roots']);g['moved']=g['count']-g['unchanged'];g['new_tuple']=not any((g['where'],g['kind'],g['reason'])==(p['where'],p['kind'],p['reason']) for p in old)
(out/'grouped-first-stops.json').write_text(json.dumps(sorted(groups.values(),key=lambda g:(-g['count'],g['where'],g['kind'],g['reason'])),indent=2)+'\n')
(out/'comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
print('roots='+str(len(rows)))
