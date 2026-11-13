"""Render observed groups and exact per-root baseline comparisons."""
import json,os
from pathlib import Path
base=Path(__file__).resolve().parent;mode=os.environ['CENSUS_MODULE_FORMAT'];out=base/mode
read=lambda n:json.loads((out/n).read_text())
groups=read('grouped-first-stops.json');comparisons=read('comparison.json')
s=['# Next stops: '+mode,'','Measured compiler/apply commit: `516c3dedf370293d4e4f8f78ab4628d391aed5e4`. Roots and locations are relative to `src/compiler/`. Two ordinary observations per root agree except timing.','','| Roots | Location | Kind | Exact reason | Phase | Owner |','| ---: | --- | --- | --- | --- | --- |']
for g in groups:s.append('| '+' | '.join([str(g['count']),g['where'],g['kind'],g['reason'].replace('|','\\|').replace('<','&lt;').replace('>','&gt;'),g['phase'],g['owner']])+' |')
s+=['','## Root membership','']
for g in groups:s+=['- '+g['where']+' '+g['kind']+': '+', '.join('`'+r+'`' for r in g['roots'])]
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
s+=['','Files lowering whole: '+str(len(accepted))+'/79: '+', '.join(accepted)+'. No native-output success is claimed.','','## Comparison with f2546f1f','','Baseline: typescript/next-stops-f2546f1f at `2cda42f68690691c4ed080293684e1ff911adcfa`.','',str(sum(c['unchanged'] for c in comparisons))+' unchanged exact tuples; '+str(sum(not c['unchanged'] for c in comparisons))+' changed.','']
former={c['root'] for c in comparisons if c['before'][0]=='sys.ts:160:35'};assert len(former)==73
s+=['The former 73 sys.ts:160:35 roots now go to:','']
for g in groups:
 n=len(former.intersection(g['roots']))
 if n:s+=['- '+str(n)+' to '+g['where']+': '+g['reason']]
s+=['','| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |','| --- | --- | --- |']
for c in comparisons:s+=['| '+c['root']+' | '+'; '.join(c['before']).replace('|','\\|').replace('<','&lt;').replace('>','&gt;')+' | '+'; '.join(c['after']).replace('|','\\|').replace('<','&lt;').replace('>','&gt;')+' |']
old=json.loads((base/('464e5daa-'+mode+'-grouped.json')).read_text())
before={r:[g['where'],g['kind'],g['reason']] for g in old for r in g['roots']}
s+=['','## Comparison with 464e5daa','','Table at `c0d27049a616498d5619a9487536c83542c775ec`.','','| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |','| --- | --- | --- |']
for c in comparisons:s+=['| '+c['root']+' | '+'; '.join(before[c['root']]).replace('|','\\|').replace('<','&lt;').replace('>','&gt;')+' | '+'; '.join(c['after']).replace('|','\\|').replace('<','&lt;').replace('>','&gt;')+' |']
s+=['','## Adapted source at stops','']
for e in read('source-evidence.json'):
 if 'adapted_excerpt' in e:s+=['### '+e['where'],'','```typescript',e['adapted_excerpt'],'```','']
s+=['## Method','','One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from measured stage3/meter/owners.json or the emitting method.']
(out/'REPORT.md').write_text('\n'.join(s)+'\n')
