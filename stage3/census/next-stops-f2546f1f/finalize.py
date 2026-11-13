"""Assemble independently validated format reports and artifact hashes."""
import gzip,hashlib,json,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent
modes=['CommonJSBundle','EcmaScript']
records={m:[json.loads(l) for l in (base/m/'census.jsonl').read_text().splitlines()] for m in modes}
assert all(len(r)==79 for r in records.values())
keys=['roots','kind','where','reason','fix','diagnostics']
matching=sum({k:a.get(k) for k in keys}=={k:b.get(k) for k in keys} for a,b in zip(*(records[m] for m in modes)))
assert all((base/m/'validation.json').exists() for m in modes)
s=['# Next stops on f2546f1f','','Measured commit: f2546f1fcb7c76f8dd549094401246296b6998a7. Baseline: completed 464e5daa table at c0d27049a616498d5619a9487536c83542c775ec.','','Both formats independently measured all 79 compiler roots, twice each. '+str(matching)+'/79 first-stop tuples agree across formats. CommonJSBundle was validated and pushed before EcmaScript began. No new apply or build occurred between formats.','']
s+=['All 73 former PollingInterval roots move to one shared new stop in each format: sys.ts:160:35, Refused: optional property Low in Partial<Levels> absent from structural source {}, which can hide fields. The other six roots are unchanged. Files lowering whole in both formats: corePublic.ts and hostErrors.ts (2/79).','']
for m in modes:
 comparisons=json.loads((base/m/'comparison.json').read_text());former=[c for c in comparisons if c['before'][0]=='sys.ts:142:81'];assert len(former)==73 and sum(not c['unchanged'] for c in comparisons)==73
 assert all(c['after']==['sys.ts:160:35','Refused','optional property Low in Partial<Levels> absent from structural source {}, which can hide fields'] for c in former)
 report=(base/m/'REPORT.md').read_text().splitlines()[1:];report=['#'+line if line.startswith('#') else line for line in report]
 s+=['## '+m,'','\n'.join(report)]
s+=['## Setup and limits','','A fresh detached worktree at the measured commit initialized both pinned submodules. The dedicated upstream clone confirms v6.0.3 is 050880ce59e30b356b686bd3144efe24f875ebc8; stage3/api npm ci exited 0. The first build exited 1 because the tools environment referenced missing /tmp/adamic-gate. Creating that directory and retrying the unchanged build exited 0. Apply ran once and exited 0. Failed and successful command statuses are retained. The largest recorded completed command duration is below 30 minutes. No compiler or adaptation change was made.','','## Interrupted attempt','','The environment/session reset removed the outer runner during CommonJSBundle roots 51–60. Its orphaned processes were terminated; the incomplete attempt is preserved under interrupted-CommonJSBundle-51-60 with no fabricated group exit. The full group was repeated under the bounded runner; the primary table uses that completed pair per root. Defunct process entries are excluded from the live-child check.','','## Evidence','','The copied recipe runs each child and group with a 1780-second bound, at most ten roots per group and three children at a time. Status files retain actual exits. Source/binary hashes, duplicate checks, per-root before/after comparisons, emitter evidence, and checkpoint history are retained beside this report. Only this census directory differs from the measured compiler.']
checkpoints=[dict(sha=line.split(' ',1)[0],subject=line.split(' ',1)[1]) for line in subprocess.check_output(['git','log','--reverse','--format=%H %s','f2546f1f..HEAD'],text=True).splitlines() if 'checkpoint' in line]
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
(base/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(result)
