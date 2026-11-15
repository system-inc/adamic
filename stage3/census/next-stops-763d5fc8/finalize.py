"""Assemble measured reports, validate paired formats, and hash evidence."""
import gzip,hashlib,json,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent
modes=['CommonJSBundle','EcmaScript']
records={m:[json.loads(l) for l in (base/m/'census.jsonl').read_text().splitlines()] for m in modes}
assert all(len(r)==79 for r in records.values())
keys=['kind','where','reason','fix','diagnostics']
assert [r['roots'] for r in records[modes[0]]]==[r['roots'] for r in records[modes[1]]]
matching=sum({k:a.get(k) for k in keys}=={k:b.get(k) for k in keys} for a,b in zip(*(records[m] for m in modes)))
assert all((base/m/'validation.json').exists() for m in modes)
for m in modes:
 comparisons=json.loads((base/m/'comparison.json').read_text())
 former=[c for c in comparisons if c['before'][0]=='sys.ts:1385:31']
 assert len(former)==73

p=json.loads((base/'provenance.json').read_text())
s=['# Next stops on 763d5fc8', '', 'Measured commit: `'+p['sha']+'`. Parent: `'+p['parent']+'`. Baseline: typescript/next-stops-55c375d9 at `'+p['baseline']+'` (both formats).', '', '## Stops and movement from 55c375d9', '']
for m in modes:
 cs=json.loads((base/m/'comparison.json').read_text())
 s += ['### '+m,'']
 for g in json.loads((base/m/'grouped-first-stops.json').read_text()):
  members=[c for c in cs if c['root'] in g['roots']]
  moved=sum(not c['unchanged'] for c in members)
  s += ['- '+str(g['count'])+(' root: ' if g['count']==1 else ' roots: ')+g['where']+'; '+g['kind']+'; '+g['reason']+'. '+str(moved)+' moved, '+str(len(members)-moved)+' unchanged from 55c375d9.']
 s += ['']
s += ['Both formats independently measured all 79 compiler roots twice. '+str(matching)+'/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.','']
forecast_sites=['sys.ts:1500:57','sys.ts:1598:57','sys.ts:1532:13','sys.ts:1533:13','sys.ts:1588:192']
s += ['The forecast was clearTimeout (task #7yz6rtp), watchFile/watchDirectory (task #cyjqp5f), or process.recordreplay (task #b7phwab), whichever was first in source order and still on main. Forecast locations in the 55c375d9 tree: '+', '.join(forecast_sites)+'.','']
for m in modes:
 cs=json.loads((base/m/'comparison.json').read_text())
 from collections import Counter
 stops=Counter(c['after'][0] for c in cs if c['before'][0]=='sys.ts:1385:31')
 assert sum(stops.values())==73
 for stop,n in stops.items():
  s += [m+': '+str(n)+' of the former 73 roots meet '+stop+' first. '+('The forecast did not hold: the writeFile assignment in patchWriteFileEnsuringDirectory stops them before any of the forecast sites.' if stop=='sys.ts:1387:5' else 'See the measured source below for the forecast assessment.'),'']
s += ['## Adapted source at new stops','']
seen=set()
for m in modes:
 cs=json.loads((base/m/'comparison.json').read_text())
 new={tuple(c['after']) for c in cs if not c['unchanged']}
 for e in json.loads((base/m/'source-evidence.json').read_text()):
  key=(e['where'],e['kind'],e['reason'])
  if key in new and key not in seen:
   seen.add(key);s+=['### '+e['where'],'','```typescript',e['adapted_line'],'```','','Adapted context:','','```typescript','\n'.join(line.rstrip() for line in e['adapted_excerpt'].splitlines()),'```','','Exact reason: '+e['reason']+'.','','Phase: '+e['phase']+'. Owner: '+e['owner']+'. Emitter: '+e['emitter']+'.','']
   if e.get('owner_basis'):s += ['Ownership evidence: '+e['owner_basis']+'.','']
   if e.get('expression_change_reach'):s += ['Prior lowering change reach: '+e['expression_change_reach'],'']
if not seen:s+=['No new stop tuples relative to the 55c375d9 baseline.','']
for m in modes:
 report=(base/m/'REPORT.md').read_text().splitlines()[1:]
 s += ['## '+m,'','\n'.join('#'+line if line.startswith('#') else line for line in report)]
s += ['', 'The first 50-root checkpoint push lost GitHub credentials (exit 128); the exact failure and historical partial report are retained in checkpoint-push-failure.json and BLOCKED-AT-50.md. The user-requested retry succeeded, after which measurement resumed without repeating completed roots.', '', '## Setup and limits','','A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. An initial build attempted before nested submodule checkout completed exited 1 because cohere/TypeScript/tsc/go.mod was absent; its exact log and exit are retained as build-before-submodules. The retry after submodule initialization exited 0. The upstream clone confirmed v6.0.3 at '+p['upstream']+'. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. Both apply and build completed before measurement. All npm cache writes used scratch. The shared checkout’s tracked files were unchanged. No compiler or adaptation changes were made. One completed apply and one successful ordinary build served both formats. The separate sys.ts single-file census was skipped.','','Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.']
checkpoints=[dict(sha=line.split(' ',1)[0],subject=line.split(' ',1)[1]) for line in subprocess.check_output(['git','log','--reverse','--format=%H %s',p['sha']+'..HEAD'],text=True).splitlines() if 'checkpoint' in line]
assert len(checkpoints)==16
(base/'checkpoints.json').write_text(json.dumps(checkpoints,indent=2)+'\n')
s += ['','## Evidence','','[artifact-manifest.json](artifact-manifest.json) records committed artifact sizes and SHA256 values. [oversized-artifacts.json](oversized-artifacts.json) indexes excluded binaries by SHA256; no file over 20 MB is committed. Raw paired observations, actual exits, compressed logs with hashes, source excerpts and per-root baseline comparisons are retained beside this report.','','## Checkpoints','','| Checkpoint | Commit |','| --- | --- |']
for c in checkpoints:s += ['| '+c['subject']+' | '+c['sha']+' |']
(base/'REPORT.md').write_text('\n'.join(s)+'\n')
statuses=[json.loads(p.read_text()) for p in base.rglob('*-status.json')]
result=dict(roots_per_mode=79,raw_observations=316,matching_first_stop_tuples=matching,maximum_child_seconds=max(s['seconds'] for s in statuses),bundle_complete_commit=json.loads((base/'bundle-complete.json').read_text())['commit'])
assert result['maximum_child_seconds']<1800
(base/'cross-format-validation.json').write_text(json.dumps(result,indent=2)+'\n')
for name,info in json.loads((base/'compression.json').read_text()).items():
 p=base/(name+'.gz')
 if p.exists():assert hashlib.sha256(gzip.decompress(p.read_bytes())).hexdigest()==info['sha256']
manifest={str(p.relative_to(base)):dict(bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()) for p in sorted(base.rglob('*')) if p.is_file() and p.name!='artifact-manifest.json' and '__pycache__' not in p.parts}
assert all(d['bytes']<=20_000_000 for d in manifest.values())
(base/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(result)
