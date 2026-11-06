from pathlib import Path
import json
repo=Path('/workspace/adamic'); out=repo/'stage3/census/next-stops-7e3c0682'
groups=json.loads((out/'grouped-first-stops.json').read_text())
checkpoint=json.loads((out/'CHECKPOINT.json').read_text())
box=json.loads((out/'box.json').read_text())
comparisons=json.loads((out/'comparison.json').read_text())
artifact=json.loads((out/'root-order.json').read_text())
incomplete=artifact.get('status')=='incomplete'
pending={row['root']:row for row in artifact.get('pending',[])}
artifacts={row['root'].split('src/compiler/',1)[-1]:row for row in artifact['root_order_artifacts']}
accepted=[name for g in groups if g['kind']=='accepted' for name in g['roots']]
old_cast=[c for c in comparisons if c['before'][0]=='core.ts:810:89']
moved=sum(len(c['roots']) for c in old_cast if c['after'][0]!='core.ts:810:89')
text=[f'# Next stops on 7e3c0682', '',
 f'{moved}/73 roots moved off core.ts:810:89. The new biggest group is {groups[0]["count"]} roots at `{groups[0]["where"]}`, {groups[0]["kind"]}: {groups[0]["reason"]}.', '',
 f'Compiler and apply source: `{checkpoint["compiler_commit"]}`. TypeScript 6.0.3 pin: `{checkpoint["typescript_pin"]}`.',
 f'Box: `{box["hostname"]}`; {box["nproc"]} visible CPUs; cgroup quota {box["cpu_quota"]["quota_us"]}/{box["cpu_quota"]["period_us"]} microseconds; memory {box["total_memory"]}. Load start: {box["loadavg_start"]}; load end: {box.get("loadavg_end")}.', '',
 '## Grouped first stops', '',
 'All root names below are relative to `src/compiler/`. Root-order artifacts retain their ordinary first-stop rows; they are never folded into accepted or omitted.', '',
 '| Roots | Location | Kind | Exact reason | Root names | Comparison with 99w |',
 '| ---: | --- | --- | --- | --- | --- |']
for g in groups:
 names=', '.join('`'+name+'`'+(' (root-order artifact)' if name in artifacts else ' (artifact classification incomplete)' if name in pending else '') for name in g['roots'])
 matching=[c for c in comparisons if tuple(c['after'])==(g['where'],g['kind'],g['reason'])]
 status='unchanged' if all(c['status']=='unchanged' for c in matching) else 'new stop; roots moved from 99w'
 text.append(f'| {g["count"]} | `{g["where"]}` | {g["kind"]} | {g["reason"]} | {names} | {status} |')
text += ['', '## Whole-root lowering', '', f'{len(accepted)}/79 roots lower whole in the ordinary independent-root census: '+', '.join('`'+s+'`' for s in accepted)+'.',
 'This means ordinary Lower accepted each entire root program. It does not claim native emission or that the aggregate 79-root program lowers.', '',
 '## Comparison with 99w', '',
 '| 99w group | Roots | Status now |', '| --- | ---: | --- |']
for c in comparisons:
 before=c['before']; after=c['after']
 desc=c['status']+(' to `'+after[0]+'`, '+after[1]+': '+after[2] if before!=after else '')
 text.append(f'| `{before[0]}`, {before[1]}: {before[2]} | {len(c["roots"])} | {desc} |')
text += ['', 'The old core.ts:810 group is gone as a first stop. The core.ts:966:72 group is new. The two accepted roots and the four singleton stop groups are compared above using every exact root path from 99w\'s census-results.json, including duplicate basenames in different subdirectories.', '',
 '## Root-order artifacts', '',
 ('STOPPING REASON: The meter entry census was terminated at its 1800-second bound after emitting only its header. The 79-root ordinary census is complete. Full meter artifact classification is incomplete because attempted-source and entry-finding evidence was not emitted. No artifact is inferred from Node throws alone. The three Node-throwing readiness roots remain separate ordinary first-stop rows, with incomplete labels. `root-order.json` retains the gap and Node results.' if incomplete else 'Classification uses the unmodified `stage3/meter/report.py` artifact rule: the entry census must attempt the stop file without a readiness finding there, and the meter Node probe must throw. `root-order.json` and `root-order.md` retain the exact classifications, entry evidence, and Node results. Ordinary readiness stops remain in the grouped table.'), '']
for root,row in pending.items():
 text.append(f'- `{root}`: artifact classification incomplete; Node exit {row["node"]["exit"]}, `{row["node"]["first_error"]}`. No whole-root lowering credit is assigned.')
for root,row in artifacts.items():
 stop=row['entry_next_stop']
 text.append(f'- `{root}`: root-order artifact; Node exit {row["node"]["exit"]}, `{row["node"]["first_error"]}`. Entry next stop: '+('none' if stop is None else f'{stop["kind"]} `{stop["where"]}`: {stop["reason"]}')+f'. Complete program attempted: {row["entry_program_attempted"]}.')
text += ['', ('The official meter artifact total is undetermined. Only the two ordinary accepted roots are credited as whole-root lowerings. Artifact roots cannot gain whole-root credit without complete entry-program evidence.' if incomplete else f'Meter root-order-only lowering total: {artifact["totals"]["lowering"]}/79. Artifact classification does not count a root as lowering whole unless the entire resolved program was attempted and the entry census has no finding in that program.'), '',
 '## New biggest stop', '', 'Exact adapted source, `src/compiler/core.ts:958-971` (the retained `core.ts-958-971.txt` preserves source line endings):', '', '```ts', (out/'core.ts-958-971.txt').read_text().rstrip(), '```', '',
 'The stop is line 966, column 72: the first argument `xs` in `append(xs, ys)`. The exact reason is assembled by `wideningRefusal` in `internal/lower/invariance.go:715`, with `Refused` returned at line 729. `reason-grep.txt` retains the grep hits and `invariance.go.txt` retains lines 699-730. These are the compiler files at 7e3c0682.', '',
 '## Reproduction and evidence', '',
 f'Setup took {checkpoint["setup_seconds"]} seconds; apply took {checkpoint["apply_seconds"]} seconds on this box. Product key: `{checkpoint["product_key"]}`. Manifest SHA-256: `{checkpoint["manifest_sha256"]}`. `adapted-product.manifest.json` retains the exact manifest. `binaries.sha256` identifies the unchanged ordinary census and the separate meter entry-census binaries.', '',
 '```sh', 'source /workspace/adamic-tools/env.sh', 'bash cloud/setup.sh > /workspace/scratch/next-stops/setup.log 2>&1', 'npm ci --prefix stage3/api > /workspace/scratch/next-stops/npm-ci.log 2>&1', 'export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source', 'stage3/apply.sh --build-product /workspace/next-stops-adapted > /workspace/scratch/next-stops/apply.log 2>&1', 'go build -o /workspace/scratch/next-stops/census ./stage3/census/tool > /workspace/scratch/next-stops/build.log 2>&1', 'python3 stage3/census/next-stops-7e3c0682/run-groups.py > /workspace/scratch/next-stops/groups.log 2>&1', '```', '',
 'The missing upstream source checkout was cloned at tag v6.0.3 and its HEAD verified against the exact pin. The 79 sorted TypeScript roots use one completed adapted tree. Groups contain 20, 20, 20, and 19 roots. The unchanged census CLI accepts one file as its source argument and emits a per-file row plus a single-root aggregate row. Both raw rows are retained in each group JSONL; the runner checks their non-timing fields are identical and counts only the first. Thus no root is counted twice. No non-TypeScript inputs or multi-root aggregate are mixed into the 79-root table.', '',
 'The ordinary census has no overlay, suppression, compiler edits, adaptation edits, or partial-root fallback. The separate meter entry census uses the meter\'s standard measurement overlay for artifact classification only. `build-entry.py` reproduces it with physical scratch Go filenames ending in .go.txt; virtual Go paths and overlay contents are unchanged. Entry raw evidence is retained compressed beside this report, along with Node probes and source excerpts.', '',
 'Each completed group was committed and pushed immediately. All census invocations exited or were reaped before completion; no census process remains. The branch is report only, with required checkpoint history preserved and no push to main.']
result='\n'.join(text)+'\n'
assert '\u2013' not in result and '\u2014' not in result
(out/'REPORT.md').write_text(result)
