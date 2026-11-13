"""Assemble independently validated measured reports and hash artifacts."""
import gzip,hashlib,json,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent;modes=['CommonJSBundle','EcmaScript']
records={m:[json.loads(l) for l in (base/m/'census.jsonl').read_text().splitlines()] for m in modes}
assert all(len(r)==79 for r in records.values())
keys=['kind','where','reason','fix','diagnostics']
matching=sum({k:a.get(k) for k in keys}=={k:b.get(k) for k in keys} for a,b in zip(*(records[m] for m in modes)))
assert all((base/m/'validation.json').exists() for m in modes)
s=['# Next stops on 516c3ded','','Measured commit: `516c3dedf370293d4e4f8f78ab4628d391aed5e4` (diagnostics/assignment-value-1). Parent: `751879544ea6e627e67c91d6ab3c4ba595be6356`. Baseline: `2cda42f68690691c4ed080293684e1ff911adcfa` on typescript/next-stops-f2546f1f. Additional per-root comparison with 464e5daa is included.','','Both formats independently measured all 79 compiler roots twice. '+str(matching)+'/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.','']
for m in modes:
 groups=json.loads((base/m/'grouped-first-stops.json').read_text());comparisons=json.loads((base/m/'comparison.json').read_text());former={c['root'] for c in comparisons if c['before'][0]=='sys.ts:160:35'};assert len(former)==73
 stops=[(len(former.intersection(g['roots'])),g) for g in groups if former.intersection(g['roots'])]
 s+=['The 73 former optional-Low stops in '+m+' go to '+str(len(stops))+' stop group(s):']
 for n,g in stops:s+=['- '+str(n)+' at '+g['where']+'; '+g['kind']+'; '+g['reason']]
 s+=['']
for m in modes:
 report=(base/m/'REPORT.md').read_text().splitlines()[1:];report=['#'+line if line.startswith('#') else line for line in report]
 s+=['## '+m,'','\n'.join(report)]
if (base/'sys-comparison.md').exists():s+=['## sys.ts full census','',''+(base/'sys-comparison.md').read_text()]
s+=['## Setup and limits','','Fresh detached scratch worktrees used the full measured and parent SHAs, with pinned cohere and TypeScript submodules initialized. The upstream clone confirms v6.0.3 is 050880ce59e30b356b686bd3144efe24f875ebc8. stage3/api npm ci exited 0. /tmp/adamic-gate was created before build. The initial parent local-submodule copy failed because file transport was disabled; the local-only retry enabled it and retained the same pinned SHAs. No compiler or adaptation changes were made. One completed apply and one ordinary build served both formats. The parent full census uses the same adapted source: the measured commits differ only in compiler code and tests.','','Each ordinary child and group has a 1780-second bound with actual exits recorded, at most ten roots per group, three children at a time. Duplicate observations agree except timing. This measures first stops, not successful native execution. Owners use the measured owners.json or emitting code method.']
checkpoints=[dict(sha=line.split(' ',1)[0],subject=line.split(' ',1)[1]) for line in subprocess.check_output(['git','log','--reverse','--format=%H %s','516c3dedf370293d4e4f8f78ab4628d391aed5e4..HEAD'],text=True).splitlines() if 'checkpoint' in line]
assert len(checkpoints)==16
(base/'checkpoints.json').write_text(json.dumps(checkpoints,indent=2)+'\n')
s+=['','## Checkpoints','','| Checkpoint | Commit |','| --- | --- |']
for c in checkpoints:s+=['| '+c['subject']+' | '+c['sha']+' |']
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
(base/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n');print(result)
