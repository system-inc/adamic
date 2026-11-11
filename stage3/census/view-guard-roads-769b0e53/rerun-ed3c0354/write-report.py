#!/usr/bin/env python3
"""Render the latest report only after analysis is complete; refusals require explicit readings."""
import json,posixpath,re
from pathlib import Path
D=Path(__file__).resolve().parent;P=D.parent
s=json.loads((D/'summary.json').read_text());v=json.loads((D/'normal-verification.json').read_text());cases=json.loads((D/'cases.json').read_text());readings={}
if cases:
 readings=json.loads((D/'readings.json').read_text());assert all(str(c['id']) in readings for c in cases)
runs=[]
for title,prefix in [('Head build','build'),('Head first-error tsc entry','first'),('Identity fixture','identity-real'),('Identity forced-refusal control','identity-mutant'),('Parent diagnostic build','build-parent-diagnostics'),('Parent identity fixture','parent-identity-real'),('Parent identity forced-refusal control','parent-identity-mutant')]:
 r=json.loads((D/f'raw/{prefix}.json').read_text());runs.append(f"| {title} | {r['child_exit']} | {r['elapsed_seconds']:.3f} | {r['peak_rss_kib']} |")
for title,prefix in [('Reused parent first-error guard census','guard-parent'),('Reused parent seam-free baseline','baseline-parent')]:
 r=json.loads((P/f'raw/{prefix}.json').read_text());runs.append(f"| {title} | {r['child_exit']} | {r['elapsed_seconds']:.3f} | {r['peak_rss_kib']} |")
roadrows=[]
for road in ['elements','index values','union arms','base pair check']:
 roadrows.append(f"| {road} | 0 | {s['roads_observations'].get(road,0)} | {s['roads_findings']['raw_findings'].get(road,0)} | {s['roads_findings']['distinct_reads'].get(road,0)} |")
counts=s['pair_counts'];catalog=['# Every current guard refusal','']
for c in cases:
 r=c['record'];catalog += [f"## Case {c['id']}: {r['where']}",'',f"Road: {r['road']}. Source: `{r['source']}`. Target: `{r['target']}`. Observations: {c['observations']}.",'','Exact reason:','',r['reason'],'','Source text:','','```typescript',r['text'],'```','',readings[str(c['id'])],'','Owner units:','']+['- `'+u+'`' for u in c['owner_units']]+['','Native root and path diagnostics:','','```json',json.dumps(r['diagnostics'],indent=2),'```','']
if not cases:catalog+=['No guard refusals. There are no remaining or new pairs to read, no split crossings to classify, and no checker-identical pointer-distinct refusing pairs.','']
(D/'FINDINGS.md').write_text('\n'.join(catalog))
if not cases:full='No full-mode work was started for ed3c0354: the first-error pair has no new findings, so the requested conditional full-mode check does not apply. The archived 769b0e53 full-mode results describe only that earlier head.'
else:full='The first-error result was announced before any full-mode work. See FULL.md for full-mode status and every finding in selected files.'
text=f'''# View guard roads ed3c0354 against 16429a30

Task 3zex15c. Target: `ed3c0354e2d3f97663778e91aac7f100a47b4e18`, found on `memory/view-guard-roads-3`. The compiler was built from that exact checkout. The parent and its cached streams remain `16429a30`; the same parent-adapted TypeScript 6.0.3 tree is used on both sides.

## First-error result

{len(cases)} raw guard findings and {counts['distinct_reads']['new_findings']} distinct findings are added against the parent. {s['refusal_observations']} refusal observations occur across {s['guard_stopped_units']} units. There are {s['probe_errors']} probe errors and {s['diagnostic_errors']} diagnostic errors. All {s['views']} recorded views in {s['observed_units']} attempted units have the same per-unit population as the parent. The complete normal inventory also retains 504 deferred generic units.

| First widened road | Parent observations | Head observations | New raw findings | New distinct findings |
| --- | ---: | ---: | ---: | ---: |
{chr(10).join(roadrows)}

Of the earlier 289 exact raw findings, **{s['prior_289_exact_findings_remain']} remain and {s['prior_289_exact_findings_gone']} are gone**. There are **{s['new_exact_findings_not_in_289']} new findings outside those 289**. The independent view-location/source/target comparison finds {s['prior_view_sites_still_refuse']} previously refusing views still refusing and {s['new_view_sites_not_in_prior']} new refusing views; this avoids counting a changed first reason as a cleared view. Of the earlier 114 units with no ordinary finding, **{s['prior_114_still_guard_stopped']} still have a latent guard stop** and {s['prior_114_cleared']} have none. Current ordinary-finding-free guard-stopped units total {s['ordinary_finding_free_guard_stopped_units']}.

The stage3/census/pair raw/distinct results are preserved in [delta.md](delta.md) and delta.json. Parent-to-head cleared findings, moved findings and changed guard-first stops in already guard-blocked units are all zero because the parent guard had no refusals. The comparison to the earlier head is separate from those parent-to-current counts.

## Remaining and new refusals

[Every refusal and its source reading](FINDINGS.md) is listed by road, with exact reason, source text, root types and native path diagnostics. Refusing observations containing an actual split crossing without literal-field separation: **{s['split_crossing_refusal_observations']}**. Refusing observations containing a checker-identical but pointer-distinct pair at the root or any path step: **{s['checker_identical_distinct_object_refusal_observations']}**. Separate machine-readable lists are split-crossing-cases.json and checker-identical-distinct-object-cases.json. These categories can overlap.

The identity fixture [identity.ts.txt](../identity.ts.txt) records one view and zero real refusals at this head, as at the parent. The scratch mutant that refuses every pair emits one refusal on each side, proving both observers are active. The complete extended observer source is byte-identical on both sides and builds on both. All four fixture runs preserve the cached parent's ordinary bytes apart from checker-cache telemetry. See identity-result.json and raw/identity-*. The production lowering source was never modified by the measurement seam.

## Ordinary outcomes and measurement validity

The normal first-error output retains all **{v['ordinary_findings']} ordinary finding observations**. Serialized finding arrays are byte-identical to the seam-free cached baseline (SHA-256 `{v['ordinary_findings_sha256']}`). Every header, file, unit and finding byte outside the relation_report telemetry is identical (SHA-256 `{v['non_telemetry_bytes_sha256']}`). The excluded telemetry differs only in checker relation-cache sizes. There is no ordinary regression, changed ordinary first stop or new panic. normal-verification.json contains the assertions.

[The method](METHOD.md) specifies exactly what the seam populates and what it does not model. This is **units only**: it does not model whole-program graph membership. A latent refusal does not establish a production whole-program stop. The only observer extension is post-refusal identity/literal-field metadata at the root and every real failure-path step, plus recognition of the new split-arm road label. The identical extended observer is also built on the parent and checked with the fixture control. The parent census has no refusals, so its cached stream needs no supplemental annotations.

## Full mode

{full}

## Time and peak memory

Each wrapper records the child's own exit code and wait4 peak RSS. Times include checker loading and census startup; the build row measures the go child and its compiler descendants. Parent census measurements are reused, not rerun. The 56.464-second parent diagnostic build overlapped the head census; these operational timings are not a controlled speed comparison.

| Run | Child exit | Elapsed seconds | Peak RSS KiB |
| --- | ---: | ---: | ---: |
{chr(10).join(runs)}

The superseded 8e57a877 first-error run was stopped on the target change after 273.022 seconds and 1170 timed units (child exit -15, peak RSS 1328768 KiB); there is no completed pair for it. No d4e6f92c or d3757b95 census ran. No superseded or current census process is left running after completion.

Raw streams and per-unit timings are beside this report in raw/; files over 1 MB are gzipped. Exact overlay sources use .go.txt names. analyze.py, write-report.py, provenance.json and the archived parent evidence make the comparison reproducible. The earlier completed census remains [REPORT-769b0e53.md](../REPORT-769b0e53.md).
'''
assert '\u2013' not in text and '\u2014' not in text
(D/'REPORT.md').write_text(text)
print('Rendered',len(cases),'cases')

roottext=re.sub(r"\]\(([^)]+)\)",lambda m:"]("+posixpath.normpath("rerun-ed3c0354/"+m.group(1))+")",text)
roottext=roottext.replace("Task 3zex15c.","Unlinked artifact names below are relative to `rerun-ed3c0354/`.\n\nTask 3zex15c.",1)
(P/'REPORT.md').write_text(roottext)
