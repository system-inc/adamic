"""Render measured CommonJSBundle first stops and exact root comparisons."""
import json
from pathlib import Path
base=Path(__file__).resolve().parent;out=base/'CommonJSBundle'
read=lambda n:json.loads((out/n).read_text())
groups=read('grouped-first-stops.json');cs=read('comparison.json');evidence=read('source-evidence.json');p=json.loads((base/'provenance.json').read_text())
bind=[g for g in groups if g['reason'].startswith('bind of a function that declares ')]
s=['# CommonJSBundle next stops at ba619e28','','## Roots first stopped by the new bind reason','']
if not bind:s+=['None of the 79 roots first stops with the new bind reason.','']
for g in bind:
 s+=['- Roots: '+', '.join('`'+r+'`' for r in g['roots'])+'. Site: `'+g['where']+'`. Reason: '+g['reason']+'.']
 e=next(e for e in evidence if (e['where'],e['kind'],e['reason'])==(g['where'],g['kind'],g['reason']))
 s+=['','Quoted adapted source line:','','```typescript',e['adapted_line'],'```','']
s+=['## Every root whose first stop differs from the comparison','', 'Comparison: `'+p['comparison_branch']+'` at `'+p['baseline']+'`. '+p['comparison_note'],'']
changed=[c for c in cs if not c['unchanged']]
if not changed:s+=['None. All 79 first-stop tuples are unchanged.','']
else:
 s+=['| Root | Before: site; kind; exact reason | After: site; kind; exact reason |','| --- | --- | --- |']
 for c in changed:s+=['| '+c['root']+' | '+'; '.join(c['before']).replace('|','\\|')+' | '+'; '.join(c['after']).replace('|','\\|')+' |']
 s+=['']
s+=['## Grouped first-stop table','','| Roots | Site | Kind | Exact reason | Phase | Owner |','| ---: | --- | --- | --- | --- | --- |']
for g in groups:s+=['| '+' | '.join([str(g['count']),g['where'],g['kind'],g['reason'].replace('|','\\|'),g['phase'],g['owner']])+' |']
s+=['','## Group membership','']
for g in groups:s+=['- `'+g['where']+'`, '+g['kind']+': '+', '.join('`'+r+'`' for r in g['roots'])+'.']
s+=['','## All 79 root comparisons','','| Root | Before: site; kind; exact reason | After: site; kind; exact reason | Changed |','| --- | --- | --- | --- |']
for c in cs:s+=['| '+c['root']+' | '+'; '.join(c['before']).replace('|','\\|')+' | '+'; '.join(c['after']).replace('|','\\|')+' | '+('no' if c['unchanged'] else 'yes')+' |']
s+=['','## Source at measured stops','']
for e in evidence:
 if 'adapted_line' in e:s+=['### '+e['where'],'','```typescript',e['adapted_line'],'```','']
s+=['## Method and limits','','Measured compiler and existing adaptation inputs: `'+p['sha']+'` (task #04f5s0s). A fresh detached worktree initialized recursive submodules. TypeScript source was cloned at `'+p['upstream']+'` (v6.0.3); stage3/api `npm ci` passed. Existing `stage3/apply.sh` prepared the ordinary input tree without changing adaptations. One ordinary census binary was built and reused. CommonJSBundle only; no separate sys.ts single-file census.','','All 79 sorted compiler roots ran in separate child processes, with two observations per root equal except timing. Groups contain at most ten roots; three groups run concurrently, each preserving the copied three-root-child concurrency. Each root and group has a 1780-second bound with exit and cleanup recorded. Source hashes and the binary hash were checked. No compiler or adaptation changes were made. This measures first stops and lowering acceptance, not successful native execution.','','Paths and sites are relative to `src/compiler/`. Comparisons use exact (site, kind, reason) tuples. Raw paired observations, statuses, compressed logs, source evidence, and the full root comparison are retained.']
text='\n'.join(s)+'\n';(out/'REPORT.md').write_text(text);(base/'REPORT.md').write_text(text)
