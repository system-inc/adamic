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
p=json.loads((base/'provenance.json').read_text())
s=['# Next stops on 5c8338c4','','Measured commit: `'+p['sha']+'` (main, "stage3/adapt: admit undefined slots in the changed-file polling queue"). Parent: `'+p['parent']+'`. Baseline: `'+p['baseline']+'` on typescript/next-stops-5d334653.','','Both formats independently measured all 79 compiler roots twice. '+str(matching)+'/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.','']
for m in modes:
 cs=json.loads((base/m/'comparison.json').read_text())
 former=[c for c in cs if c['before'][0]=='sys.ts:290:19']
 assert len(former)==73
 changed=sum(not c['unchanged'] for c in cs)
 s += [m+': '+str(changed)+' changed exact tuples; '+str(79-changed)+' unchanged from 5d334653.','']
 accepted=[r for g in json.loads((base/m/'grouped-first-stops.json').read_text()) if g['kind']=='accepted' for r in g['roots']]
 s += ['Files lowering whole in '+m+': '+str(len(accepted))+'/79: '+', '.join(accepted)+'.','']
s += ["The per-root comparisons below record every changed and unchanged exact tuple against 5d334653.",'']
s += ['','## Adapted source at new stops','']
seen=set()
for m in modes:
 cs=json.loads((base/m/'comparison.json').read_text())
 new={tuple(c['after']) for c in cs if not c['unchanged']}
 for e in json.loads((base/m/'source-evidence.json').read_text()):
  key=(e['where'],e['kind'],e['reason'])
  if key in new and key not in seen:
   seen.add(key);s+=['### '+e['where'],'','```typescript',e['adapted_line'],'```','','Exact reason: '+e['reason']+'.','','Phase: '+e['phase']+'. Owner: '+e['owner']+'. Emitter: '+e['emitter']+'.','']
   if e.get('owner_basis'):s += ['Ownership evidence: '+e['owner_basis']+'.','']
if not seen:s+=['No new stop tuples relative to the 5d334653 baseline.','']
for m in modes:
 report=(base/m/'REPORT.md').read_text().splitlines()[1:]
 s += ['## '+m,'','\n'.join('#'+line if line.startswith('#') else line for line in report)]
s += ['## Setup and limits','','A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. The upstream clone confirmed v6.0.3 at '+p['upstream']+'. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. All npm cache writes used scratch. The shared checkout’s tracked files were unchanged. No compiler or adaptation changes were made. One completed apply and one ordinary build served both formats. The separate sys.ts single-file census was skipped.','','Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.']
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
