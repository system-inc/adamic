"""Render observed groups and exact per-root baseline comparisons."""
import json,os
from pathlib import Path
base=Path(__file__).resolve().parent;mode=os.environ['CENSUS_MODULE_FORMAT'];out=base/mode
read=lambda n:json.loads((out/n).read_text())
groups=read('grouped-first-stops.json');comparisons=read('comparison.json')
s=['# Next stops: '+mode,'','Measured compiler/apply commit: `7534c1d7575db02bb7b6512119e0685987b7e333`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.','','| Roots | Location | Kind | Exact reason | Phase | Owner |','| ---: | --- | --- | --- | --- | --- |']
lead=['## Stops and movement from ff4ce03f','']
for g in groups:
 members=[c for c in comparisons if c['root'] in g['roots']]
 moved=sum(not c['unchanged'] for c in members)
 lead+=['- '+str(g['count'])+' roots: '+g['where']+'; '+g['kind']+'; '+g['reason']+'. '+str(moved)+' moved, '+str(len(members)-moved)+' unchanged.']
lead+=['','## Adapted source at new stops','']
for e in read('source-evidence.json'):
 if any(not c['unchanged'] and c['after']==[e['where'],e['kind'],e['reason']] for c in comparisons):
  lead+=['### '+e['where'],'','```typescript',e['adapted_line'],'```','','Phase: '+e['phase']+'. Owner: '+e['owner']+'.','',e.get('owner_basis',''),'']
s=s[:4]+lead+s[4:]
for g in groups:s.append('| '+' | '.join([str(g['count']),g['where'],g['kind'],g['reason'].replace('|','\\|').replace('<','&lt;').replace('>','&gt;'),g['phase'],g['owner']])+' |')
s+=['','## Root membership','']
for g in groups:s+=['- '+g['where']+' '+g['kind']+': '+', '.join('`'+r+'`' for r in g['roots'])]
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
s+=['','Files lowering whole: '+str(len(accepted))+'/79: '+', '.join(accepted)+'. No native-output success is claimed.','','## Comparison with ff4ce03f','','Baseline: typescript/next-stops-ff4ce03f at `2147e992cf42301c1e20c64c8d39c9ab027c65d7`.','',str(sum(c['unchanged'] for c in comparisons))+' unchanged exact tuples; '+str(sum(not c['unchanged'] for c in comparisons))+' changed.','']
former={c['root'] for c in comparisons if c['before'][0]=='sys.ts:514:17'};assert len(former)==73
s+=['The former 73 sys.ts:514:17 roots now go to:','']
for g in groups:
 n=len(former.intersection(g['roots']))
 if n:s+=['- '+str(n)+' to '+g['where']+': '+g['reason']]
s+=['','| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |','| --- | --- | --- |']
for c in comparisons:s+=['| '+c['root']+' | '+'; '.join(c['before']).replace('|','\\|').replace('<','&lt;').replace('>','&gt;')+' | '+'; '.join(c['after']).replace('|','\\|').replace('<','&lt;').replace('>','&gt;')+' |']
s+=['','## Adapted source at stops','']
for e in read('source-evidence.json'):
 if 'adapted_excerpt' in e:s+=['### '+e['where'],'','```typescript','\n'.join(line.rstrip() for line in e['adapted_excerpt'].splitlines()),'```','']
for e in read('source-evidence.json'):
 if e.get('expression_change_reach'):s+=['### ff4ce03f never-read reach: '+e['where'],'',e['expression_change_reach'],'']
s+=['## Method','','One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from emitting source and stage3/adapt/**/owners.json searches retained in ownership-search.json.']
(out/'REPORT.md').write_text('\n'.join(s)+'\n')
