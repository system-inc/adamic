#!/usr/bin/env python3
import collections
import json
from pathlib import Path

out = Path(__file__).resolve().parent
repo = out.parents[2]
tree = Path('/tmp/next-stops-f6c44ef5/adapted')
def site(value):
    return 'src/' + value.rsplit('/src/', 1)[1] if '/src/' in value else value
records = [json.loads(line) for line in (out / 'census.jsonl').read_text().splitlines()]
roots = [r for r in records if len(r['roots']) == 1 and Path(r['roots'][0]).suffix == '.ts']
expected = {str(p) for p in (tree / 'src/compiler').rglob('*.ts')}
assert len(roots) == 79 and {r['roots'][0] for r in roots} == expected
assert len(records) == 83
assert len([r for r in records if len(r['roots']) != 1]) == 1
old = json.loads((repo / 'stage3/adapt/99-correctness-axolotl-q/census-results.json').read_text())
old = {r['root']: r for r in old['results']}
results = [dict(root=site(r['roots'][0]), kind=r['kind'], where=site(r.get('where', '')), reason=r.get('reason', ''), seconds=r['seconds']) for r in roots]
groups = collections.defaultdict(list)
for r in results:
    groups[r['where'], r['kind'], r['reason']].append(r['root'])
ordered = sorted(groups, key=lambda k: (-len(groups[k]), k))
comparison = []
for r in results:
    before = old[r['root']]
    same = all(before[k] == r[k] for k in ('kind', 'where', 'reason'))
    comparison.append(dict(root=r['root'], classification='unchanged' if same else 'moved', before={k: before[k] for k in ('where','kind','reason')}, after={k:r[k] for k in ('where','kind','reason')}))
assert collections.Counter(r['classification'] for r in comparison) == {'moved':73,'unchanged':6}
assert all(r['before']['where']=='src/compiler/core.ts:736:36' and r['after']['where']=='src/compiler/core.ts:810:89' for r in comparison if r['classification']=='moved')
accepted = [r['root'] for r in results if r['kind']=='accepted']
assert accepted == ['src/compiler/corePublic.ts', 'src/compiler/hostErrors.ts']
summary = dict(roots=79, lowers_whole=accepted, groups=[dict(where=k[0],kind=k[1],reason=k[2],count=len(groups[k]),roots=groups[k]) for k in ordered], results=results, comparison=comparison)
(out / 'results.json').write_text(json.dumps(summary,indent=2)+'\n')
box = json.loads((out / 'box.json').read_text())
lines = ['# Next stops after brand-mint admission', '',
'All 73 roots that stopped at `core.ts:736:36` in 99q now stop at `src/compiler/core.ts:810:89`, Refused: "a cast the runtime can\'t check". No root still stops at core.ts:736. The next refusal is the comparator fallback in `sortAndDeduplicate`, before ordinary lowering reaches the admitted brand values. This census does not observe a NotYet at those values.', '',
'Compiler and adaptation SHA: `f6c44ef5f5d3908e134e2f098bf5c9cabdff424f`, fetched as origin/main. Report-only branch: `typescript/next-stops-f6c44ef5`.', '',
'TypeScript: `v6.0.3`, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Both measurement tools and the single adapted tree come from the measured Adamic SHA. The ordinary tool uses no overlay, suppression, backend change, or partial-root fallback.', '',
f"Box: `{box['hostname']}`; {box['nproc']} available CPUs; quota {box['cpu_quota']['quota_us']}/{box['cpu_quota']['period_us']} microseconds; memory {box['total_memory']}. Full load metadata is in box.json. Times describe this box only.", '',
f"The complete ordinary run took {sum(r['seconds'] for r in records):.3f} recorded seconds ({sum(r['seconds'] for r in records)/60:.2f} minutes). It finished before the 30-minute timeout; no split was needed. All 79 roots have exactly one row. The three JSON-input rows and the aggregate 79-root single-entry error remain in census.jsonl, outside the table.", '',
'## First stops', '',
'Root names below are relative to `src/compiler/`. Groups use the exact (location, kind, reason). Artifact labels are supplemental; no root is folded away.', '',
'| Count | Location | Kind | Reason | Root names |', '| ---: | --- | --- | --- | --- |']
for k in ordered:
    names=', '.join('`'+n.removeprefix('src/compiler/')+'`' for n in groups[k])
    location = '`'+k[0]+'`' if k[0] else '-'
    reason = k[2] or '-'
    lines.append(f'| {len(groups[k])} | {location} | {k[1]} | {reason} | {names} |')
lines += ['', '**2/79 roots lower whole:** `corePublic.ts`, `hostErrors.ts`. This means the ordinary Lower call succeeded; no native output is claimed.', '',
'## Comparison with 99q', '',
'The per-root comparison in results.json uses 99q\'s census-results.json as well as its CENSUS.md table. Compiler and adaptation changes are measured together; this is not a causal isolation of one commit.', '',
'| 99q group | Current group | Classification |', '| --- | --- | --- |',
'| 73 Refused at core.ts:736:36 | The same 73 roots, Refused at core.ts:810:89, same reason | Moved; the core.ts:810:89 location group is new |',
'| 2 lower whole | corePublic.ts and hostErrors.ts | Unchanged |',
'| 2 enum-before-init | diagnosticInformationMap.generated.ts at commandLineParser.ts:293:35; types.ts at diagnosticInformationMap.generated.ts:13:45 | Unchanged, kept as two exact-location groups |',
'| 1 at semver.ts:71:9 | debug.ts, same NotYet and reason | Unchanged |',
'| 1 at debug.ts:243:59 | core.ts, same NotYet and reason | Unchanged |', '',
'No additional root or reason group is new. In total: 73 moved, 6 unchanged.', '',
'## Biggest group source and emitter', '',
'The ordinary path in `internal/lower/lower.go` runs module refusal scans before `declareModule` and function-body lowering. Recognition admits the mint casts to that scan, but the later comparator refusal stops this program before branded-value lowering. The tests of isolated mint examples and these full-root first stops therefore observe different phases.', '',
'Exact adapted `src/compiler/core.ts:807-811` (line 810 is the stop):', '', '```ts']
source=(tree/'src/compiler/core.ts').read_text().splitlines()
lines += source[806:811] + ['```', '',
'Column 89 starts `compareStringsCaseSensitive as any as Comparer<T>`. The full return line is retained above.', '',
'The reason is constructed in `internal/lower/cast_proof.go:180`, inside `castProof`. The cast scan calls it through `castRefusal` at line 150; the scan rule is registered at line 164. The `any` source/target branch returns this refusal at lines 185-186. Brand-mint recognition is checked earlier, at lines 176-178.', '', '```go',
'    refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime can\'t check", Fix: l.castRepair(node)}',
'```', '',
'Located with `rg -n -F "a cast the runtime can" internal/lower`; the source file and its SHA are pinned by the compiler commit.', '',
'## Evidence', '',
'The ordinary binary build metadata records Go 1.27.1 and f6c44ef5. Its vcs.modified flag is true with the untracked report directory present; source-diff.txt is empty and confirms the measured tracked sources are unchanged.', '',
'`census.jsonl` and `census.log.gz` retain the complete ordinary stream and stderr. `results.json` retains normalized roots, groups, timings, and the per-root 99q comparison. `provenance.json` records compiler, TypeScript pin, binary hashes, and key adapted source hashes. `COMMANDS.md` records preparation and invocation commands. Build and apply logs are retained alongside these files.', '']
(out/'REPORT.md').write_text('\n'.join(lines))
print('Validated 79 roots: 73 moved, 6 unchanged; 2 lower whole.')

proof = json.loads((out / 'artifact-proof.json').read_text())
assert proof['source_files'] == 79 and len(proof['root_order_artifacts']) == 3
marked = {a['root'] for a in proof['root_order_artifacts']}
summary['root_order_artifacts'] = proof['root_order_artifacts']
summary['artifact_certification'] = proof['certification']
summary['entry_program_lowering_certified'] = False
(out / 'results.json').write_text(json.dumps(summary,indent=2)+'\n')
report = (out / 'REPORT.md').read_text()
for root in marked:
    short = root.removeprefix('src/compiler/')
    report = report.replace('`'+short+'` |', '`'+short+'` (root-order artifact) |', 1)
report += '\n## Root-order artifact labels\n\n'
report += 'The exact candidate expression from `stage3/meter/report.py:root_order_checks` was evaluated with proven inputs, then its Node-throws classification was applied. All three readiness stops are **root-order artifacts**. Their original first stops and root counts remain in the table. This does not certify whole-program lowering for their entry reaches.\n\n'
report += 'The proof avoids an unrelated complete body census. Native full-mode unit timing rows witness attempted units in every stop file used by a readiness candidate: semver.ts, commandLineParser.ts, and diagnosticInformationMap.generated.ts. The timer runs only after the unit is marked attempted; the only later status change is panic, which the meter also accepts. Additional attempted files cannot change candidate membership because every ordinary readiness stop file is already witnessed.\n\n'
report += 'The readiness-finding set is provably empty in this guarded entry binary. Both readiness reasons are emitted only inside `namespaceInitialization` in `internal/lower/namespaces.go` (lines 255, 257, 306, and 309). Its sole production call is in lower.go:90, removed by the meter overlay when LowerWithOptions is replaced with its no-output error. The active overlay has only the function definition, no call or method-value reference. The guarded entry binary contains neither the function symbol nor either reason literal; the ordinary binary contains all of them. `artifact-proof.json` records these checks, binary hashes are in provenance.json, and the symbol-table evidence is retained. No finding or first-stop record is synthesized.\n\n'
report += '| Root | Original first stop | Node result | Meter label |\n| --- | --- | --- | --- |\n'
for row in proof['readiness_node_checks']:
    error = row['node']['first_error'].replace('|','&#124;')
    report += '| `'+row['root'].removeprefix('src/compiler/')+'` | `'+row['where']+'` | exit '+str(row['node']['exit'])+': '+error+' | '+row['classification']+' |\n'
report += '\nNode v24.19.0 and TypeScript 6.0.3 were used through the unchanged meter node_readiness.py helper. Native attempted-unit witnesses, the exact candidate expression, all probe results, and the reachability certificate are retained in artifact-proof.json. The incomplete supplementary body scans were stopped and are not represented as complete entry censuses. Their original data, including the timing witnesses, are in auxiliary-entry-attempts.tar.gz. Symbol tables are entry-symbols.txt.gz and census-symbols.txt.gz. The 79-root ordinary census is complete.\n'
(out / 'REPORT.md').write_text(report)
print('Certified labels retained on all 3 artifact roots; no roots folded away.')
