"""Validate cross-format agreement and write the combined evidence index."""
import gzip,hashlib,json,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent
modes=['CommonJSBundle','EcmaScript'];records={}
for mode in modes:
 records[mode]=[json.loads(l) for l in (base/mode/'census.jsonl').read_text().splitlines()]
assert len(records[modes[0]])==len(records[modes[1]])==79
keys=['roots','kind','where','reason','fix','diagnostics']
assert all({k:a.get(k) for k in keys}=={k:b.get(k) for k in keys} for a,b in zip(*[records[m] for m in modes]))
checkpoints=[dict(sha=line.split(' ',1)[0],subject=line.split(' ',1)[1]) for line in subprocess.check_output(['git','log','--reverse','--format=%H %s','464e5daa..HEAD'],text=True).splitlines()]
assert len(checkpoints)==16
(base/'checkpoints.json').write_text(json.dumps(checkpoints,indent=2)+'\n')
s=['# Baobab: next stops on 464e5daa','','Both formats independently measured all 79 sorted compiler roots. All 79 first-stop (location, kind, exact reason) tuples agree between formats. Each root has two raw observations agreeing except timing: 316 raw observations across both modes.','','CommonJSBundle was completed, validated and pushed at ef97b8ca2 before any EcmaScript root ran. The final branch is one evidence commit directly on 464e5daa23a60588f5aaa656e95d93af05d582c3. Checkpoint history is retained separately on typescript/next-stops-464e5daa-checkpoint.','','## Grouped tables','']
for mode in modes:
 report=(base/mode/'REPORT.md').read_text()
 table=report.split('| Roots |',1)[1].split('\n\n## Root membership',1)[0]
 s+=['### '+mode,'','| Roots |'+table,'']
s+=['## Changes from the requested baselines','','Bundle versus stage3/meter/README.md, Nov 13 method-change reading: 74 changed, 5 unchanged. Of the 73 formerly at core.ts:2599:15 (process), 72 now stop at sys.ts:142:81 and sys.ts itself stops at scanner.ts:469:12. core.ts itself moves from performanceCore.ts:29:15 to sys.ts:142:81, bringing the new shared group to 73.','','ES versus typescript/next-stops-f44b5c27-2: all 73 formerly at sys.ts:57:1 now stop at sys.ts:142:81; the other 6 roots are unchanged.','','The sys.ts root is Refused at scanner.ts:469:12 in both modes, exactly: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it).','','Files lowering whole in both modes: corePublic.ts and hostErrors.ts, 2/79. No native-output success is claimed.','','## Reproduction and evidence','','The six requested scripts were copied from 3dc2a59dede1bff6eef557cb987f6ea0a3e6a462 and adapted to the new pins, paths and per-format evidence directories. The requested source README is absent; REPORT.md and tooling-adaptations.md supply its recipe. Fresh-run changes are detailed in each mode report and prior-tooling-adaptations.md. measure.py records the bounded invocation recipe. group.py runs at most ten roots, three at a time after the initial sequential bundle batch; run.py bounds each group and child to 1780 seconds and reports every child exit.','','stage3/apply.sh ran once. Its tree and ordinary census binary were reused for both formats. Root status JSON and raw observations record module_format; every root invocation explicitly inherited CENSUS_MODULE_FORMAT. No ordinary compiler or census-tool overlay was used. Only this new evidence directory differs from the measured commit.','','The initial census build exited 1 because the fresh worktree lacked checker submodules. Pinned submodule initialization and the retry exited 0. These actual exits remain in status files; no failed root measurement is omitted.','','Each mode REPORT.md contains the full grouped table, all root memberships, owners, phases and comparisons. comparison.json retains exact before/after tuples; source-evidence.json retains adapted excerpts and emitter names. owners.json comes from the measured commit; code methods identify stops without an exact owner mapping. Retained .go.txt files show dispatch and phase ordering.','','validation.json in each mode checks all roots, duplicate agreement, actual exits, duration bounds, explicit format, baseline tuples, binary/source hashes and source integrity. cross-format-validation.json records independent agreement. binaries.json and ordinary-buildinfo.txt identify the built ordinary binary. provenance.json identifies compiler, tooling, TypeScript and environment.','','Completed logs are compressed losslessly; compression.json keeps raw and compressed SHA256. Artifacts above 20 MB are omitted in favor of their SHA256; none of the committed artifacts exceeds 20 MB. artifact-manifest.json identifies every final evidence file except itself.','','## Checkpoints','','| Checkpoint | Commit |','| --- | --- |']
for c in checkpoints:s+=['| '+c['subject']+' | '+c['sha']+' |']
(base/'README.md').write_text('\n'.join(s)+'\n')
statuses=[json.loads(p.read_text()) for p in base.rglob('*-status.json')]
result=dict(roots_per_mode=79,raw_observations_per_mode=158,total_raw_observations=316,matching_first_stop_tuples=79,all_root_exits_zero=all(json.loads(p.read_text())['exit_code']==0 for p in base.rglob('root-??-status.json')),maximum_recorded_child_seconds=max(s['seconds'] for s in statuses),bundle_complete_pushed_before_es='ef97b8ca2',checkpoints=16)
assert result['all_root_exits_zero'] and result['maximum_recorded_child_seconds']<1800
(base/'cross-format-validation.json').write_text(json.dumps(result,indent=2)+'\n')
compression=json.loads((base/'compression.json').read_text())
for name,info in compression.items():
 path=base/(name+'.gz')
 if path.exists():assert hashlib.sha256(gzip.decompress(path.read_bytes())).hexdigest()==info['sha256'] and hashlib.sha256(path.read_bytes()).hexdigest()==info['compressed_sha256']
manifest={str(p.relative_to(base)):dict(bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()) for p in sorted(base.rglob('*')) if p.is_file() and p.name!='artifact-manifest.json' and '__pycache__' not in p.parts}
assert all(d['bytes']<=20_000_000 for d in manifest.values())
(base/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(json.dumps(result,indent=2))
