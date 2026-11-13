"""Write only validated measured groups; owner evidence is resolved from source."""
import json,os
from pathlib import Path
base=Path(__file__).resolve().parent;mode=os.environ['CENSUS_MODULE_FORMAT'];out=base/mode
groups=json.loads((out/'grouped-first-stops.json').read_text());comparisons=json.loads((out/'comparison.json').read_text())
s=['# Baobab next stops: '+mode,'','Measured compiler/apply commit: `464e5daa23a60588f5aaa656e95d93af05d582c3`. Locations and roots are relative to `src/compiler/`. Two ordinary observations per root agree except timing.','','| Roots | Location | Kind | Exact reason | Phase | Owner |','| ---: | --- | --- | --- | --- | --- |']
for g in groups:s.append('| '+ ' | '.join([str(g['count']),g['where'],g['kind'],g['reason'].replace('|','\\|'),g['phase'],g['owner']])+' |')
s+=['','## Root membership','']
for g in groups:s+=['- '+g['where']+' '+g['kind']+': '+', '.join('`'+r+'`' for r in g['roots'])]
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
s+=['','Files lowering whole: '+str(len(accepted))+'/79: '+', '.join(accepted)+'. No native-output success is claimed.','','## Comparison','',('Bundle baseline: meter README Nov 13 method-change section, with exact raw head records from 3dd24184b9150872269ff035bd0f3814397ac437.' if mode=='CommonJSBundle' else 'ES baseline: typescript/next-stops-f44b5c27-2, commit 3dc2a59dede1bff6eef557cb987f6ea0a3e6a462.')]
s+=['',str(sum(c['unchanged'] for c in comparisons))+' unchanged exact tuples; '+str(sum(not c['unchanged'] for c in comparisons))+' changed. Full before/after tuples are in comparison.json.','']
if mode=='CommonJSBundle':
 former={c['root'] for c in comparisons if c['before'][0]=='core.ts:2599:15'};assert len(former)==73
 s+=['The former 73 process roots now go to:','']
 for g in groups:
  n=len(former.intersection(g['roots']))
  if n:s+=['- '+str(n)+' to '+g['where']+': '+g['reason']]
sys=next(c for c in comparisons if c['root']=='sys.ts');s+=['','sys.ts root: '+'; '.join(sys['after'])+'.','','## Method and limits','','One adapted tree built once with stage3/apply.sh; unchanged ordinary stage3/census/tool built at the measured commit; one root per bounded child; groups of ten (last nine), pushed to typescript/next-stops-464e5daa-checkpoint. The child environment explicitly sets CENSUS_MODULE_FORMAT='+mode+'. Status JSON records every actual child exit and elapsed bound. The initial build failed due to absent checker submodules; pinned submodule initialization and build retry are retained. No compiler overlay or implementation change.','','Fresh-run adaptations: copied the requested six Python scripts from 3dc2a59; repinned worktree, binary, tree and evidence paths; split format output directories; analyzer uses exact requested baselines; report and validation cover each format separately; checkpoint.py packages only completed logs and preserves hashes for oversized raw artifacts. Historical tooling notes are retained. The named source README does not exist; REPORT.md and tooling-adaptations.md supply its recipe.','','Owner and phase evidence: source-evidence.json and retained internal/lower excerpts. owners.json is copied from the measured commit; a code emitter is named where there is no exact mapping. Source and binary SHA256 identify measured bytes. Validation covers all roots, duplicate agreement, exit codes, time limits and source integrity.']
(out/'REPORT.md').write_text('\n'.join(s)+'\n')
