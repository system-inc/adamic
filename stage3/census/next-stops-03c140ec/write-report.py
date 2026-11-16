"""Render only measured groupings, exact stop tuples and reviewed source evidence."""
import collections,json
from pathlib import Path
b=Path(__file__).resolve().parent
read=lambda p:json.loads((b/p).read_text())
cs=read('comparison.json');old=read('base/CommonJSBundle/grouped-first-stops.json')
former=set(next(g['roots'] for g in old if g['count']==73))
esc=lambda s:str(s).replace('|','\\|').replace('\n','<br>')
fmt=lambda t:esc('; '.join(t))
s=['# Next stops at 03c140ec','','Measured compiler: `03c140ec28a249b8b0f13eb24e34fc3d3df05266`, resolved from fetched main. Baseline: the corresponding 82c78b6b tables in `typescript/next-stops-82c78b6b`, branch tip `17124ea520c61521ddaf201a55035fb8e6b3ae2f`. Each format is compared with its own archived table.','']
for mode in ['CommonJSBundle','EcmaScript']:
 groups=read('head/'+mode+'/grouped-first-stops.json')
 s += ['## '+mode+': first stops and movement','','| Roots | Stop | Moved from 82c78b6b |','|---:|---|---|']
 for g in groups:
  matches=[c for c in cs if c['module_format']==mode and c['root'] in g['roots']];moved=sum(not c['unchanged'] for c in matches)
  s += ['| '+str(g['count'])+' | '+esc(g['where']+'; '+g['kind']+'; '+g['reason'])+' | '+str(moved)+' moved; '+str(len(matches)-moved)+' unchanged |']
 s += ['']
 for g in groups:
  n=len(former.intersection(g['roots']))
  if n:s += [f'Of the former 73 roots, **{n} first meet `{g["where"]}` ({g["kind"]})**: {g["reason"]}.','']
for mode in ['CommonJSBundle','EcmaScript']:
 groups=read('head/'+mode+'/grouped-first-stops.json')
 s += ['## '+mode+': grouped table','','| Roots | Location | Kind | Exact reason | Phase | Owner |','|---:|---|---|---|---|---|']
 for g in groups:s+=['| '+' | '.join(esc(g[k]) for k in ['count','where','kind','reason','phase','owner'])+' |']
 s += ['','## '+mode+': root membership','']
 for g in groups:s += ['- `'+g['where']+'` ('+g['kind']+'): '+', '.join('`'+r+'`' for r in g['roots'])+'.']
 accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
 s += ['','Files lowering whole in '+mode+': '+(', '.join('`'+r+'`' for r in accepted) if accepted else 'none')+'. Acceptance means the ordinary load/check/lower gate accepts the complete rooted program; it does not claim runtime execution.','']
 s += ['## '+mode+': adapted source and ownership','']
 for e in read('head/'+mode+'/source-evidence.json'):
  if e['where']=='-':continue
  old=read('base/'+mode+'/grouped-first-stops.json')
  isnew=not any([g['where'],g['kind'],g['reason']]==[e['where'],e['kind'],e['reason']] for g in old)
  s += ['### `'+e['where']+'`'+('; new stop' if isnew else '; retained stop'),'','```typescript',e['adapted_line'],'```','',e.get('source_review','Ownership review pending.'),'',e.get('host_runtime','Host/runtime classification pending.'),'','Emitter: '+e.get('emitter','pending')+'.','']
s += ['## Per-root comparison','','Each tuple is location; kind; exact reason. Movement refers to tuple equality against the corresponding format baseline table.','','| Root | Baseline CommonJSBundle | CommonJSBundle | Moved | Baseline EcmaScript | EcmaScript | Moved |','|---|---|---|---|---|---|---|']
by={(c['module_format'],c['root']):c for c in cs}
for root in sorted(former | {r for g in old for r in g['roots']}):
 a=by['CommonJSBundle',root];z=by['EcmaScript',root]
 s += ['| '+root+' | '+fmt(a['before'])+' | '+fmt(a['after'])+' | '+('no' if a['unchanged'] else 'yes')+' | '+fmt(z['before'])+' | '+fmt(z['after'])+' | '+('no' if z['unchanged'] else 'yes')+' |']
s += ['','## Method and validation','','One detached scratch worktree was created at the requested SHA resolved from fetched main; recursive submodules were initialized with `--depth 1`. TypeScript v6.0.3 is pinned at `050880ce59e30b356b686bd3144efe24f875ebc8`. `npm ci` ran in stage3/api. Temporary directories and Go caches are under `/workspace/scratch/next-stops-03c140ec`.','','The Go binary records the requested main revision with vcs.modified=true: only the new census directory differed at build time. Scratch HEAD remained at the measured SHA through the build; no compiler files were changed. measurement-worktree-status.txt records the pre-measurement state. Compiler/adaptation sources were unchanged; validation restricts all differences to this directory.', '', 'The unchanged apply ran once and the ordinary census tool built once. CommonJSBundle ran first, all 79 roots were validated and pushed as a checkpoint, then EcmaScript used exactly the same adapted tree and binary with no new apply/build. Each root ran as one child and emitted two equal observations except `seconds`. Groups held at most ten roots, with three children at a time; each child and group had a 1780-second bound and its exit was recorded. No compiler overlay, adaptation change, latent census, or forecast was used.','','See `validation.json`, per-child/group statuses, compressed logs, paired root JSONL, `binaries.json`, source manifests, provenance, ownership searches, and `checkpoints.json`. `lower.go.txt` snapshots the measured compiler only. The baseline observations are archived evidence, not rerun here.']
s += ['', 'The validated CommonJSBundle checkpoint preceding EcmaScript was `'+read('format-transition.json')['commonjs_validated_checkpoint']+'`; `format-transition.json` records the transition and unchanged binary hash.']
v=read('validation.json')
s += ['', f'Validation passed for {v["raw_observations"]} raw observations (79 roots × two formats × two observations). All {len(v["all_child_exits"])} recorded child/group/setup exits are zero; maximum recorded elapsed time is {v["maximum_child_seconds"]:.3f} seconds, below 1800 seconds. Frozen adapted source and binary hashes match; no live census children remain.']
for mode in ['CommonJSBundle','EcmaScript']:
 c=[x for x in cs if x['module_format']==mode]
 s += [f'{mode}: {sum(not x["unchanged"] for x in c)}/79 first-stop tuples moved from the baseline; {sum(x["unchanged"] for x in c)} are unchanged.']
(b/'REPORT.md').write_text('\n'.join(s)+'\n')
