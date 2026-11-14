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
 former=[c for c in comparisons if c['before'][0]=='sys.ts:618:33']
 assert len(former)==73 and all(c['after'][0]=='sys.ts:790:87' for c in former)

p=json.loads((base/'provenance.json').read_text())
s=['# Next stops on 8e3a784a', '', 'Measured commit: `'+p['sha']+'`. Parent: `'+p['parent']+'`. Baseline: typescript/next-stops-7534c1d7 at `'+p['baseline']+'` (both formats).', '', '## Stops and movement from 7534c1d7', '']
for m in modes:
 cs=json.loads((base/m/'comparison.json').read_text())
 s += ['### '+m,'']
 for g in json.loads((base/m/'grouped-first-stops.json').read_text()):
  members=[c for c in cs if c['root'] in g['roots']]
  moved=sum(not c['unchanged'] for c in members)
  s += ['- '+str(g['count'])+(' root: ' if g['count']==1 else ' roots: ')+g['where']+'; '+g['kind']+'; '+g['reason']+'. '+str(moved)+' moved, '+str(len(members)-moved)+' unchanged from 7534c1d7.']
 s += ['']
s += ['Both formats independently measured all 79 compiler roots twice. '+str(matching)+'/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.','','Neither sys.ts:515:39 (#036ve6w) nor sys.ts:594:5 (#trxxkge) is first for the former 73 roots: all stop at sys.ts:790:87 during refusal_scan, before body lowering.', '', '## Adapted source at new stops','']
seen=set()
for m in modes:
 cs=json.loads((base/m/'comparison.json').read_text())
 new={tuple(c['after']) for c in cs if not c['unchanged']}
 for e in json.loads((base/m/'source-evidence.json').read_text()):
  key=(e['where'],e['kind'],e['reason'])
  if key in new and key not in seen:
   seen.add(key);s+=['### '+e['where'],'','```typescript',e['adapted_line'],'```','','Adapted context:','','```typescript','\n'.join(line.rstrip() for line in e['adapted_excerpt'].splitlines()),'```','','Exact reason: '+e['reason']+'.','','Phase: '+e['phase']+'. Owner: '+e['owner']+'. Emitter: '+e['emitter']+'.','']
   if e.get('owner_basis'):s += ['Ownership evidence: '+e['owner_basis']+'.','']
   if e.get('expression_change_reach'):s += ['Reach of 7534c1d7 never-read lowering change: '+e['expression_change_reach'],'']
if not seen:s+=['No new stop tuples relative to the 7534c1d7 baseline.','']
for m in modes:
 report=(base/m/'REPORT.md').read_text().splitlines()[1:]
 s += ['## '+m,'','\n'.join('#'+line if line.startswith('#') else line for line in report)]
s += ['','## Setup and limits','','A fresh detached scratch worktree used the measured SHA, with recursive submodules initialized. The upstream clone confirmed v6.0.3 at '+p['upstream']+'. stage3/api npm ci exited 0. Go/adamic caches and temporary measurement products used writable scratch. Both apply and build completed before measurement. All npm cache writes used scratch. The shared checkout’s tracked files were unchanged. No compiler or adaptation changes were made. One completed apply and one successful ordinary build served both formats. The separate sys.ts single-file census was skipped.','','Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.']
s += ['', 'Git HTTPS credentials stopped working after the EcmaScript 60-root checkpoint. The 70- and 79-root Git pushes exited 128: `fatal: could not read Username for \'https://github.com\': No such device or address`. The connected GitHub tool published the remaining objects and advanced refs without force; each published checkpoint tree was verified identical to its local tree. publication.json maps the local and published checkpoint commits. The measured SHA includes enum-never commit 1d3b46dc9e85f9c1a555fcae39f1c125c8a181f9 (verified ancestry).', '']
checkpoints=[dict(sha=line.split(' ',1)[0],subject=line.split(' ',1)[1]) for line in subprocess.check_output(['git','log','--reverse','--format=%H %s',p['sha']+'..HEAD'],text=True).splitlines() if 'checkpoint' in line]
publication=json.loads((base/'publication.json').read_text())
for c in checkpoints:
 if c['sha'] in publication['checkpoint_commit_map']:
  c['local_sha']=c['sha'];c['sha']=publication['checkpoint_commit_map'][c['sha']]
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
