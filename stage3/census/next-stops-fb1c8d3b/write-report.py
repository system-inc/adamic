"""Write the measured first-stop report; no predicted stops are admitted."""
import collections,json
from pathlib import Path
out=Path(__file__).resolve().parent
p=json.loads((out/'provenance.json').read_text())
groups=json.loads((out/'grouped-first-stops.json').read_text())
comparisons=json.loads((out/'comparison.json').read_text())
phases=json.loads((out/'stop-phases.json').read_text())
excerpts=json.loads((out/'source-excerpts.json').read_text())
q=lambda s:'`'+s+'`'
names=lambda roots:', '.join(map(q,roots))
s=['# Next stops on main fb1c8d3b','','Measured per-root next stops. Report only.','',f'Compiler and apply SHA: {q(p["sha"])}. TypeScript 6.0.3 source pin: {q(p["typescript_pin"])}.','',f'Box: {q(p["box"])}; {p["cpus"]} visible CPUs; CPU quota {q(p["cpu_max"])}; {p["meminfo"]}. Started {p["date"]}.','','## Grouped first stops','','Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.','','| Roots | Location | Kind | Exact reason | Phase | Compared with b8d5f493 | Root names |','| ---: | --- | --- | --- | --- | --- | --- |']
for g in groups:
 comparison=', '.join(f'{g[k]} {k}' for k in ['unchanged','moved','new'] if g[k])
 s.append(f'| {g["count"]} | {q(g["where"])} | {g["kind"]} | {g["reason"].replace("|",chr(92)+"|")} | {phases[g["where"]]["phase"]} | {comparison} | {names(g["roots"])} |')
old=json.loads((out/'previous-grouped-first-stops.json').read_text())
former=next(g['roots'] for g in old if g['where']=='tracing.ts:381:1')
s+=['','The baseline had 73 roots stopped at tracing.ts:381:1, the ambient declare const require. Their measured destinations are:','','| Former roots | Location | Kind | Exact reason | Phase | Root names |','| ---: | --- | --- | --- | --- | --- |']
for g in groups:
 roots=[r for r in g['roots'] if r in former]
 if roots:s.append(f'| {len(roots)} | {q(g["where"])} | {g["kind"]} | {g["reason"]} | {phases[g["where"]]["phase"]} | {names(roots)} |')
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
s+=['','## Roots lowering whole','',f'{len(accepted)}/79: {names(accepted)}. Lower acceptance covers each complete resolved root program. Native output is not measured.','','## Comparison with b8d5f493','','Baseline: b8d5f493df0227efcb33c6040ba40c906eec73c9, from typescript/next-stops-b8d5f493 (09998fc3d1fb2b1d18cd88242506bc12b2d105f4). Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.','','| Current group | Roots | Unchanged | Moved | New |','| --- | ---: | ---: | ---: | ---: |']
for g in groups:s.append(f'| {q(g["where"])} {g["kind"]} | {g["count"]} | {g["unchanged"]} | {g["moved"]} | {g["new"]} |')
s+=['', 'Total: '+', '.join(f'{sum(g[k] for g in groups)} {k}' for k in ['unchanged','moved','new'])+'. Exact before and after reasons are in comparison.json.','','| Root | Classification | Before location / kind | After location / kind |','| --- | --- | --- | --- |']
for c in comparisons:s.append(f'| {q(c["root"])} | {c["classification"]} | {q(c["before"][0])} {c["before"][1]} | {q(c["after"][0])} {c["after"][1]} |')
s+=['','## Stop source and lowering emitters','']
for g in groups:
 if g['where']=='-':continue
 evidence=phases[g['where']]
 s+=['### '+g['where'],'',g['kind']+': '+q(g['reason'])+'. Phase: '+evidence['phase']+'. '+evidence['evidence'],'']
 excerpt=next(e for e in excerpts if e['where']==g['where'])
 s+=['```ts',excerpt['text'].rstrip(),'```','']
s+=['## Method, checkpoints and cleanup','','Read and reused the per-root first-stop census from 09998fc3. The ordinary census tool, source pin, sorted 79-root set, adapted-tree path, and duplicate-observation agreement check follow that recipe. Only compiler/apply revision and the comparison baseline changed. This report measures ordinary first stops; it does not claim latent findings are ordinary stops.','','Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The single apply command is bash stage3/apply.sh /workspace/scratch/next-stops-adapted. The ordinary binary is built without an overlay using go build -o /workspace/scratch/next-stops-census ./stage3/census/tool. Exact commands, actual child exit codes, elapsed times, and process-group cleanup are retained in status JSON files.','','Each root is measured in its own child, with a 1780-second bound and cleanup allowance. Root scheduling is split into bounded pieces to keep every process below 30 minutes. Each root emits two observations (158 raw); only the first counts. Binary hashes, Go build metadata (ordinary-buildinfo.txt.gz), adapted-source hashes, source excerpts and Go evidence (.go.txt) identify the measured bytes. Raw logs are preserved losslessly as .log.gz; compression.json records their original and compressed hashes.','','One report-only commit is based directly on the measured SHA; only this directory is committed. Validation checks all 79 roots, duplicate agreement, exact baseline classification, clean tracked source, and no surviving census children.']
(out/'REPORT.md').write_text('\n'.join(s)+'\n')
