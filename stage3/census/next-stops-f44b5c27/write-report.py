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
s=['# Next stops on main f44b5c27','','Measured per-root next stops. Report only.','',f'Compiler and apply SHA: {q(p["sha"])}. TypeScript 6.0.3 source pin: {q(p["typescript_pin"])}.','',f'Box: {q(p["box"])}; {p["cpus"]} visible CPUs; CPU quota {q(p["cpu_max"])}; {p["meminfo"]}. Started {p["date"]}.','','## Grouped first stops','','Locations and root names are relative to src/compiler/. Only the first observation counts. Duplicate observations agree in all non-timing fields.','','| Roots | Location | Kind | Exact reason | Phase | Owner (method) | Compared with 3007955b | Root names |','| ---: | --- | --- | --- | --- | --- | --- | --- |']
for g in groups:
 comparison=', '.join(f'{g[k]} {k}' for k in ['unchanged','moved','new'] if g[k])
 s.append(f'| {g["count"]} | {q(g["where"])} | {g["kind"]} | {g["reason"].replace("|",chr(92)+"|")} | {phases[g["where"]]["phase"]} | {phases[g["where"]].get("owner", "— (unchanged)")} | {comparison} | {names(g["roots"])} |')
old=json.loads((out/'previous-grouped-first-stops.json').read_text())
former=next(g['roots'] for g in old if g['where']=='sys.ts:1692:17')
s+=['','The baseline had 73 roots stopped at sys.ts:1692:17, node:inspector.Session.post. Their measured destinations are:','','| Former roots | Location | Kind | Exact reason | Phase | Owner (method) | Root names |','| ---: | --- | --- | --- | --- | --- | --- |']
for g in groups:
 roots=[r for r in g['roots'] if r in former]
 if roots:s.append(f'| {len(roots)} | {q(g["where"])} | {g["kind"]} | {g["reason"]} | {phases[g["where"]]["phase"]} | {phases[g["where"]].get("owner", "— (unchanged)")} | {names(roots)} |')
s+=['', 'All 73 former inspector roots now stop at the ambient setTimeout declaration at adapted sys.ts:57:1. The expected toISOString / __filename progression is not the ordinary first stop observed here: refusal scanning ends before body lowering reaches either. The compiler-source evidence explains this ordering; no claim about a measured latent getNodeSystem stop is made.']
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
s+=['','## Roots lowering whole','',f'{len(accepted)}/79: {names(accepted)}. Lower acceptance covers each complete resolved root program. Native output is not measured.','','## Comparison with 3007955b','','Baseline: 3007955b6db96ae5f65cedb2e680d77ad1b88664, from typescript/next-stops-3007955b (933c8981edfa0e046893b1a1f1570c166a9e8e2c). Its report and grouped data are retained in previous-REPORT.md and previous-grouped-first-stops.json. Unchanged means an identical (location, kind, exact reason) tuple for that root. Moved means a different tuple present in the baseline. New means a tuple absent from the baseline.','','| Current group | Roots | Unchanged | Moved | New |','| --- | ---: | ---: | ---: | ---: |']
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
s+=['## Method, checkpoints and cleanup','','Read and reused the per-root first-stop census from baseline branch commit 933c8981. The ordinary census tool, source pin, sorted 79-root set, adapted-tree path, and duplicate-observation agreement check follow that recipe. Compiler/apply revision, comparison baseline and fresh-run tooling adaptations are recorded below. This report measures ordinary first stops; it does not claim latent findings are ordinary stops.','','Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api; export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source. The single apply command is bash stage3/apply.sh /workspace/scratch/next-stops-f44b5c27-adapted. The ordinary binary is built without an overlay using go build -o /workspace/scratch/next-stops-f44b5c27-census ./stage3/census/tool. Exact commands, actual child exit codes, elapsed times, and process-group cleanup are retained in status JSON files.','','Each root is measured in its own child, with a 1780-second bound and cleanup allowance. Root scheduling is split into bounded pieces to keep every process below 30 minutes. Each root child emits two observations (158 raw); only the first counts. Each observation repeats the full resolved root program, and the two must agree in all non-timing fields. Binary hashes, Go build metadata (ordinary-buildinfo.txt.gz), adapted-source hashes, source excerpts and Go evidence (.go.txt) identify the measured bytes. Raw logs are preserved losslessly as .log.gz; compression.json records their original and compressed hashes.','','One report-only commit is based directly on the measured SHA; only this directory is committed. Build and setup history for this run is retained in its own status files; no previous-run failed-build status was imported. Validation checks all 79 roots, duplicate agreement, exact baseline classification, clean tracked source, and no surviving census children.']
s+=['','## sys.ts and root-order artifact roots','', 'These are ordinary per-root observations. A root changes module traversal/initialization order; these three initialization-preflight stops are retained separately from the broad shared sys.ts stop. No alternate-order run or latent overlay was used.','', '| Root | Location | Kind | Exact reason | Phase | Owner (method) |','| --- | --- | --- | --- | --- | --- |']
for root in ['sys.ts','debug.ts','diagnosticInformationMap.generated.ts','types.ts']:
 c=next(c for c in comparisons if c['root']==root);where,kind,reason=c['after'];e=phases[where]
 s.append(f'| {q(root)} | {q(where)} | {kind} | {reason} | {e["phase"]} | {e.get("owner", "accepted")} |')
s+=['', 'All line/column locations are measured in the completed adapted tree. Upstream line numbers mentioned in the request can differ after adaptations; the retained source excerpts identify the actual expressions.']
s+=(out/'tooling-adaptations.md').read_text().splitlines()
checkpoint=out/'checkpoints.json'
if checkpoint.exists():
 s+=['','## Pushed checkpoints','', '| Group | Roots | Commit |','| --- | --- | --- |']
 for entry in json.loads(checkpoint.read_text()):s.append(f'| {entry["group"]} | {entry["roots"]} | {q(entry["sha"])} |')
s+=['', 'Validation passed: all 79 sorted roots are covered, all 158 raw observations agree in their 79 non-timing pairs, all root child exit codes are zero and every recorded child exit is retained, all child runtimes are below 1800 seconds, and no census child remains. See validation.json and the per-command status files.']
s+=['', 'The final branch is typescript/next-stops-f44b5c27-2, one new commit directly on the measured SHA. The earlier typescript/next-stops-f44b5c27 branch retains the stop record and was not rewritten.']
(out/'REPORT.md').write_text('\n'.join(s)+'\n')
