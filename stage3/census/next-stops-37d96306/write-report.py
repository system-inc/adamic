"""Render the measured answer, changed stops, grouped head table and membership."""
import collections,json
from pathlib import Path
base=Path(__file__).resolve().parent
read=lambda n:json.loads((base/n).read_text())
cs=read('comparison.json');groups=read('head/CommonJSBundle/grouped-first-stops.json');changed=[c for c in cs if not c['unchanged']]
phrase='one keeps something weakly that the other keeps strongly'
moved=[c for c in changed if c['after'][1]=='NotYet' and phrase in c['after'][2]]
escape=lambda s:str(s).replace('|','\\|').replace('<','&lt;').replace('>','&gt;')
fmt=lambda t:escape('; '.join(t))
s=[f'**{len(moved)} roots move to the sameKeepingWalk stop** ("{phrase}").','',f'{len(changed)}/79 roots change first-stop tuples between main `89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825` and head `37d96306133ed2782184bee1bd8406d61b83a858`; {79-len(changed)} are unchanged. CommonJSBundle only. The baseline is freshly built and measured at main itself, beside the head.','']
if changed:
 s+=['| Root | Main first stop (location; kind; exact reason) | Head first stop (location; kind; exact reason) |','| --- | --- | --- |']
 for c in changed:s+=['| '+c['root']+' | '+fmt(c['before'])+' | '+fmt(c['after'])+' |']
else:s+=['No root has a different first stop.']
s+=['','## Where the 73 go','']
former=set(read('former-73.json'));assert len(former)==73
s+=['These are the same 73 roots grouped at sys.ts:1387:5 in the copied census. Membership is pinned in [former-73.json](former-73.json). Both current revisions were measured independently:','', '| Revision | Roots | First stop | Exact reason |','| --- | ---: | --- | --- |']
for rev in ['base','head']:
 for g in read(rev+'/CommonJSBundle/grouped-first-stops.json'):
  n=len(former.intersection(g['roots']))
  if n:s+=['| '+rev+' | '+str(n)+' | '+g['where']+' ('+g['kind']+') | '+escape(g['reason'])+' |']
s+=['','## Head grouped first-stop table','','Locations and root names are relative to the shared adapted `src/compiler/` tree.','','| Roots | Location | Kind | Exact reason | Phase |','| ---: | --- | --- | --- | --- |']
for g in groups:s+=['| '+' | '.join(map(escape,[g['count'],g['where'],g['kind'],g['reason'],g['phase']]))+' |']
s+=['','## Root membership','']
for g in groups:s+=['- `'+g['where']+'` ('+g['kind']+'): '+', '.join('`'+r+'`' for r in g['roots'])+'.']
s+=['','## Adapted source at changed stops','']
seen=set()
for rev,field in [('base','before'),('head','after')]:
 for e in read(rev+'/CommonJSBundle/source-evidence.json'):
  key=(e['where'],e['kind'],e['reason'])
  if any(tuple(c[field])==key for c in changed) and key not in seen:
   seen.add(key);s+=['### '+rev+': '+e['where'],'','Exact reason: '+e['reason'],'','Phase: '+e['phase']+'. Emitter: '+e['emitter']+'.','']
   if 'adapted_line' in e:s+=['Quoted adapted source line:','','```typescript',e['adapted_line'],'```','']
if not seen:s+=['There are no changed stops between these revisions, so there are no changed-stop source lines to quote. All observed stop excerpts are retained below.','']
s+=['## Stop phases and source evidence','','The new `sameKeepingWalk` guard is a lowering-phase check, before `present()`. A lowering-phase stop can become first only where a root has no earlier scan stop. `lowerWithElementKey` runs initialization checks and the refusal scan before declaration and expression lowering. The table reports the phase of the actual first diagnostic; it does not establish whether the new guard would be encountered after that diagnostic was removed. No stop-removal experiment was performed.','','At head, `internal/lower/expression.go:413-414` returns true on `visited[t]` or the 64-type limit. Lines 467-485 add the union guard, which returns false. The caller at lines 2224-2231 emits `a <own type> seen as a <slot type> (one keeps something weakly that the other keeps strongly)`. This wording already existed on main. The census measures changes to the first-stop tuple; it does not instrument the guard.','']
for e in read('head/CommonJSBundle/source-evidence.json'):
 s+=['### '+e['where']+' ('+e['kind']+')','','Phase: '+e['phase']+'. Emitter: '+e['emitter']+'.','']
 if e.get('owner_basis'):s += [e['owner_basis'],'']
 if 'adapted_excerpt' in e:s+=['```typescript',e['adapted_excerpt'],'```','']
s+=['## Setup and validation','','The two measurement worktrees were created fresh with `git worktree add --detach` under `/workspace/scratch/next-stops-37d96306/`, separately at each requested SHA. Both ran `git submodule update --init --recursive --depth 1`. TypeScript v6.0.3 was pinned at `050880ce59e30b356b686bd3144efe24f875ebc8`; `npm ci` in head/stage3/api exited 0. `git diff 89614cc0 37d96306 -- stage3` is empty: stage3 is identical. The adapted source was built once and reused for both ordinary census binaries, each built from its own compiler worktree. No compiler or adaptation source was changed. The shared checkout has no tracked changes.','','GOTMPDIR, TMPDIR, GOCACHE, GOMODCACHE, npm cache and adaptation product storage are in scratch. Binary paths, hashes and sizes are recorded in binaries.json; binaries themselves are excluded. Any captured Go source has a `.go.txt` suffix. The adapted source manifest was frozen before measurement and checked afterward.','','Each revision ran all 79 ordinary file roots, one root per child, in groups of at most ten, with three children at a time. Each child and group has a 1780-second timeout, allowing at most ten seconds for termination before kill, and actual exit codes and elapsed times recorded. All root children produce two observations equal in every field except timing: 316 raw observations in total. CommonJSBundle only; no separate sys.ts single-file run. This is a first-stop measurement and makes no claim about native execution.','','The first setup checkpoint compressed two setup logs while their processes were active. Both complete logs were recovered from still-open file descriptors and replaced at the next checkpoint; checkpoint.py now compresses only logs with completed status records. No measurement was rerun for this.','','## Evidence','','[comparison.json](comparison.json) lists both stops for every root. `base/CommonJSBundle/` and `head/CommonJSBundle/` retain all paired observations, root and group statuses, grouped tables and source evidence. [validation.json](validation.json) records validation; [artifact-manifest.json](artifact-manifest.json) hashes artifacts; [compression.json](compression.json) hashes raw and compressed logs. [provenance.json](provenance.json) identifies the input commits. Checkpoint history is recorded in checkpoints.json.']
(base/'REPORT.md').write_text('\n'.join(s)+'\n')
print('report:',len(moved),'move to sameKeepingWalk;',len(changed),'changed roots')
