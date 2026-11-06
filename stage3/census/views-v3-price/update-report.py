#!/usr/bin/env python3
"""Render only completed observations; distinguish partial from closure totals."""
import hashlib,json,pathlib
p=pathlib.Path(__file__).resolve().parent
pieces=[json.loads(x.read_text()) for x in sorted((p/'pieces').glob('*.json'))] if (p/'pieces').exists() else []
source=json.loads((p/'source-summary.json').read_text()) if (p/'source-summary.json').exists() else None
plan=json.loads((p/'plan.json').read_text()) if (p/'plan.json').exists() else {}
complete=bool(plan) and len(pieces)==len(plan['files'])
lines=['# Views V3 checked element price','',f'**Exact latent closure totals: {"complete" if complete else "incomplete"}.** '+f'{len(pieces)} completed paired file pieces'+(f" of {len(plan['files'])}." if plan else '.'),'',
 f"**Source upper bound: {source['upper_bound']:,}; readonly subset: {source['readonly_subset']:,}.** Neither is the admission rule." if source else '**Source upper bound and readonly subset: pending completed main apply.**','',
 '| Latent measure'+(' (observed pieces only)' if not complete else '')+' | Main | Head |','|---|---:|---:|']
keys=['view_elements','view_element_checks','units','attempted','ir_observed','finding_free_attempted','finding_free_view_elements','finding_free_view_element_checks']
for k in keys:lines.append('| '+k+' | '+' | '.join(str(sum(r['sides'][s][k] for r in pieces)) if pieces else 'Not measured' for s in ['main','head'])+' |')
lines+=['','Task 6wfvfhm. Report only, not for landing. Box `4b7d656a5150`, Linux x86_64, 2026-11-06 UTC.',
'Head `9626d0113f08269e519ba216ff7f95ff84af16db`; main `2c3f322975b2dee68889e82eab0bbca324d417a9`.',
'Report base `12fec995dcadc7f1ae94e4ba63835884451378ae`; upstream TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`.',
'Both pinned compilers use their own apply.sh products. Only the report directory is committed.','',
'## Exact latent observation','',
'The runtime overlay counts ir.ViewElement and ir.ViewElementCheck by reflection using exact package and type names. It walks returned unit statements and all function bodies lowered in that unit result. Non-nil IR pointers, including ElementCheck, are visited once by identity. Main compiles without the optional types and reports zero. Counts are per isolated unit, so dependencies independently lowered in multiple units count in each observation. These are static IR nodes, not runtime execution counts.',
'',
'The meter retains partial IR after statement rollback and omits final whole-program validation and pruning. attempted excludes skips and deferred generics; ir_observed counts attempts whose snapshot reached the counter; finding_free_attempted counts attempted units with no lowering-phase finding. The finding-free columns supply a strict meaning of units lowered without claiming successful whole-program compilation.',
'',
'Each piece selects one file in the full tsc entry closure with LATENT_FULL=1 and LATENT_ONLY_FILE. Deferred instantiated witnesses have that piece context. The unmodified meter and counted binaries run with the identical selection. Removing only view_elements, view_element_checks and ir_measured from the raw counted bytes must reproduce the entire meter output byte for byte on both compilers. The ordinary fields, with only absolute adapted roots canonicalized, are passed to stage3/census/pair. Source bytes must match. All four runs and pairing share a 23-minute deadline; successful pieces are committed and plainly pushed.',
'',
'The head-owned fixtures prove (ViewElement, ViewElementCheck): per-read (2,0), map-good (1,1), for-of-good (1,1). Main gives (0,0) for all three. Both skip-kind mutants fail the same known-count assertions. A separate shared-pointer probe confirms one check reached three ways counts once, and its skip-check mutant fails. Every other latent field is byte-identical on both trees for the fixtures and mutants. counter-proof.json and counter-proof.tar.gz retain this proof.','',
'## Per-file latent counts','',
'| File | Main elements | Main checks | Head elements | Head checks | Main finding-free units | Head finding-free units |','|---|---:|---:|---:|---:|---:|---:|']
for r in pieces:
 a,b=r['sides']['main'],r['sides']['head'];lines.append(f"| {r['file']} | {a['view_elements']} | {a['view_element_checks']} | {b['view_elements']} | {b['view_element_checks']} | {a['finding_free_attempted']} | {b['finding_free_attempted']} |")
lines+=['','## Top ten observed latent files','']
for side in ['main','head']:
 lines += [side+':','']
 for r in sorted(pieces,key=lambda r:-(r['sides'][side]['view_elements']+r['sides'][side]['view_element_checks']))[:10]:
  v=r['sides'][side];lines.append(f"- {r['file']}: {v['view_elements']} elements, {v['view_element_checks']} checks")
lines+=['','## Broad source upper bound','',
'Stock TypeScript 6.0.3, with the options from void-intersections/count.cjs, enumerates read sites in the tsc entry closure on the completed main adapted tree. Array and readonly-array receivers include tuple receivers, union/intersection arms containing arrays and interfaces inheriting array bases. The readonly subset contains a readonly array or readonly tuple arm, including mixed unions. Neither column ports the checked-view admission rule. Index assignments using plain = are excluded; compound assignments and increments read and count. Repeated loop executions count as one static site. The seven fixture forms are index, at, for-of, spread, slice, map and forEach; named array handout methods are also inventoried by their method name.',
'',
'The source fixture includes mutable and readonly arrays, union, tuple, mutable tuple and intersection receivers. It expects 12 reads with 10 in the readonly subset. The index-skipping mutant fails the same assertion.','']
if source:
 lines += [f"Seven fixture forms alone: {source['seven_form_upper_bound']:,} reads; readonly subset {source['seven_form_readonly_subset']:,}. Additional named array roads account for the remainder of the broader total.",'']
 for title,key in [('Per form','forms'),('Top ten files','files'),('Per file','files')]:
  lines += ['### '+title,'','| '+('Form' if key=='forms' else 'File')+' | Upper bound | Readonly subset |','|---|---:|---:|']
  for name,r in list(source[key].items())[:10 if title=='Top ten files' else None]:lines.append(f"| {name} | {r['upper_bound']} | {r['readonly_subset']} |")
lines+=['','## Evidence and status','',
'Completed raw streams, meter baselines, paired deltas, logs and inner SHA256SUMS are packed in pieces/*.tar.gz. SHA256SUMS in this directory identifies every archive. The earlier raw-checkpoint.tar.gz records the superseded admission investigation and interrupted first apply; it is not current census evidence. Reproduction sources live in this directory; the meter and production compiler remain unchanged.','',
'Full closure totals remain unavailable until every planned file is complete.' if not complete else 'Every planned file is complete.','']
(p/'REPORT.md').write_text('\n'.join(lines))
(p/'SHA256SUMS').write_text(''.join(f'{hashlib.sha256(x.read_bytes()).hexdigest()}  {x.relative_to(p)}\n' for x in sorted(p.rglob('*.tar.gz'))))
