"""Write only validated measured groups; owner evidence is resolved from source."""
import json,os
from pathlib import Path
base=Path(__file__).resolve().parent;mode=os.environ['CENSUS_MODULE_FORMAT'];out=base/mode
groups=json.loads((out/'grouped-first-stops.json').read_text());comparisons=json.loads((out/'comparison.json').read_text())
s=['# Baobab next stops: '+mode,'','Measured compiler/apply commit: `f2546f1fcb7c76f8dd549094401246296b6998a7`. Locations and roots are relative to `src/compiler/`. Two ordinary observations per root agree except timing.','','| Roots | Location | Kind | Exact reason | Phase | Owner |','| ---: | --- | --- | --- | --- | --- |']
for g in groups:s.append('| '+ ' | '.join([str(g['count']),g['where'],g['kind'],g['reason'].replace('|','\\|'),g['phase'],g['owner']])+' |')
s+=['','## Root membership','']
for g in groups:s+=['- '+g['where']+' '+g['kind']+': '+', '.join('`'+r+'`' for r in g['roots'])]
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
s+=['','Files lowering whole: '+str(len(accepted))+'/79: '+', '.join(accepted)+'. No native-output success is claimed.','','## Comparison','','Baseline: typescript/next-stops-464e5daa at c0d27049a616498d5619a9487536c83542c775ec.','',str(sum(c['unchanged'] for c in comparisons))+' unchanged exact tuples; '+str(sum(not c['unchanged'] for c in comparisons))+' changed.','']
former={c['root'] for c in comparisons if c['before'][0]=='sys.ts:142:81'};assert len(former)==73
s+=['The former 73 PollingInterval roots now go to:','']
for g in groups:
 n=len(former.intersection(g['roots']))
 if n:s+=['- '+str(n)+' to '+g['where']+': '+g['reason']]
s+=['','| Root | Before (location; kind; exact reason) | After (location; kind; exact reason) |','| --- | --- | --- |']
for c in comparisons:s+=['| '+c['root']+' | '+'; '.join(c['before']).replace('|','\\|')+' | '+'; '.join(c['after']).replace('|','\\|')+' |']
s+=['','## Adapted source at stops','']
for e in json.loads((out/'source-evidence.json').read_text()):
 if 'adapted_excerpt' in e:s+=['### '+e['where'],'','```typescript',e['adapted_excerpt'],'```','']
s+=['## Method','','One apply and one ordinary census build, reused across formats. Groups of at most ten roots run three children at a time. Each group and child is bounded to 1780 seconds with actual exits retained. Every root has two observations equal except timing. Owners come from the measured owners.json or the emitting method.']
(out/'REPORT.md').write_text('\n'.join(s)+'\n')
