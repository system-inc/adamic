#!/usr/bin/env python3
"""Render only completed observations; distinguish partial from closure totals."""
import hashlib,json,pathlib
p=pathlib.Path(__file__).resolve().parent
pieces=[json.loads(x.read_text()) for x in sorted((p/'pieces').glob('*.json'))] if (p/'pieces').exists() else []
pieces=[r for group in pieces for r in group.get('records',[group])]
declaration_plan=json.loads((p/'declaration-plan/INDEX.json').read_text()) if (p/'declaration-plan/INDEX.json').exists() else None
declaration_state=json.loads((p/'declaration-plan/RESULTS.json').read_text()) if (p/'declaration-plan/RESULTS.json').exists() else None
if declaration_state:
 for result in declaration_state['results'].values():
  if result['status']=='complete' and not result.get('seed'):pieces.extend(result['records'])
 aggregate={}
 for record in pieces:
  name=record['file']
  if name not in aggregate:aggregate[name]=record
  else:
   for side in ['main','head']:
    for key,value in record['sides'][side].items():
     if type(value) is int:aggregate[name]['sides'][side][key]=aggregate[name]['sides'][side].get(key,0)+value
 pieces=list(aggregate.values())
source=json.loads((p/'source-summary.json').read_text()) if (p/'source-summary.json').exists() else None
plan=json.loads((p/'plan.json').read_text()) if (p/'plan.json').exists() else {}
assert len({r['file'] for r in pieces})==len(pieces),'duplicate observed file'
complete=bool(plan) and {r['file'] for r in pieces}==set(plan['files'])
if declaration_plan:complete=complete and sum(r['sides']['main']['units'] for r in pieces)==declaration_plan['units']
lines=['# Views V3 checked element price','',f'**Exact latent closure totals: {"complete" if complete else "incomplete"}.** '+(f"{sum(r['sides']['main']['units'] for r in pieces)} of {declaration_plan['units']} full-mode units paired across {len(pieces)} observed files." if declaration_plan else f'{len(pieces)} completed paired file pieces'+(f" of {len(plan['files'])}." if plan else '.')),'',
 f"**Source upper bound (seven forms): {source['seven_form_upper_bound']:,}; readonly subset: {source['seven_form_readonly_subset']:,}.** Broader named roads: {source['upper_bound']:,} / {source['readonly_subset']:,}. Neither is the admission rule." if source else '**Source upper bound and readonly subset: pending completed main apply.**','',

 '| Latent measure'+(' (observed pieces only)' if not complete else '')+' | Main | Head |','|---|---:|---:|']
keys=['view_elements','view_element_checks','units','attempted','panicked_units','ir_observed','finding_free_attempted','finding_free_view_elements','finding_free_view_element_checks']
for k in keys:lines.append('| '+k+' | '+' | '.join(str(sum(r['sides'][s].get(k,0) for r in pieces)) if pieces else 'Not measured' for s in ['main','head'])+' |')
lines+=['','Task 6wfvfhm. Report only, not for landing. Box `4b7d656a5150`, Linux x86_64, 2026-11-06 UTC.',
'Head `9626d0113f08269e519ba216ff7f95ff84af16db`; main `2c3f322975b2dee68889e82eab0bbca324d417a9`.',
'Report base `12fec995dcadc7f1ae94e4ba63835884451378ae`; upstream TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`.',
'Both pinned compilers use their own apply.sh products; all 746 src files are byte-identical. Both complete tsc entry programs are checker-clean with zero diagnostics. Only the report directory is committed.','',
'## Exact latent observation','',
'The runtime overlay counts ir.ViewElement and ir.ViewElementCheck by reflection using exact package and type names. It walks returned unit statements and all function bodies lowered in that unit result. Non-nil IR pointers, including ElementCheck, are visited once by identity. Main compiles without the optional types and reports zero. Counts are per isolated unit, so dependencies independently lowered in multiple units count in each observation. These are static IR nodes, not runtime execution counts.',
'',
'The meter retains partial IR after statement rollback and omits final whole-program validation and pruning. attempted excludes skips and deferred generics; ir_observed counts attempts whose snapshot reached the counter; finding_free_attempted counts attempted units with no lowering-phase finding. The finding-free columns supply a strict meaning of units lowered without claiming successful whole-program compilation.',
'',
'The four posted whole-file seed observations selected files in the full tsc entry closure with LATENT_FULL=1 and LATENT_ONLY_FILE. Deferred instantiated witnesses have that piece context. Their unit sets do not overlap. New declaration pieces can repeat a file, but never a completed unit. For those seed observations, the unmodified meter and counted binaries ran with identical selection. Removing only view_elements, view_element_checks and ir_measured from the raw counted bytes must reproduce the entire meter output byte for byte on both compilers. The ordinary fields, with only absolute adapted roots canonicalized, are passed to stage3/census/pair. Source bytes must match. Each binary run has a 23-minute deadline and is committed and plainly pushed as its own raw piece. The group is paired after all four raw runs are complete; then its paired archive retains the raw records and leg timings, replacing the redundant standalone archives.',
'',
'The head-owned fixtures prove (ViewElement, ViewElementCheck): per-read (2,0), map-good (1,1), for-of-good (1,1). Main gives (0,0) for all three. Both skip-kind mutants fail the same known-count assertions. A separate shared-pointer probe confirms one check reached three ways counts once, and its skip-check mutant fails. Every other latent field is byte-identical on both trees for the fixtures and mutants. counter-proof.json and counter-proof.tar.gz retain this proof.','',
'## Per-file latent counts','',
'| File | Units paired / planned | Main elements | Main checks | Head elements | Head checks | Main attempted / free / panic | Head attempted / free / panic |','|---|---:|---:|---:|---:|---:|---:|---:|']
for r in pieces:
 a,b=r['sides']['main'],r['sides']['head'];planned=len(declaration_plan['inventory'][r['file']]['units']) if declaration_plan else a['units'];lines.append(f"| {r['file']} | {a['units']} / {planned} | {a['view_elements']} | {a['view_element_checks']} | {b['view_elements']} | {b['view_element_checks']} | {a['attempted']} / {a['finding_free_attempted']} / {a.get('panicked_units',0)} | {b['attempted']} / {b['finding_free_attempted']} / {b.get('panicked_units',0)} |")
lines+=['','## Top ten observed latent files','']
for side in ['main','head']:
 lines += [side+':','']
 for r in sorted(pieces,key=lambda r:-(r['sides'][side]['view_elements']+r['sides'][side]['view_element_checks']))[:10]:
  v=r['sides'][side];lines.append(f"- {r['file']}: {v['view_elements']} elements, {v['view_element_checks']} checks")
lines+=['','## Broad source upper bound','',
'Stock TypeScript 6.0.3, with the options from void-intersections/count.cjs, enumerates read sites in the tsc entry closure on the completed main adapted tree. Array and readonly-array receivers include tuple receivers, union/intersection arms containing arrays and interfaces inheriting array bases. The readonly subset contains a readonly array or readonly tuple arm, including mixed unions. Neither column ports the checked-view admission rule. These count source sites, not a numeric limit on the sum of repeated dependency IR across isolated units. Index assignments using plain = are excluded; compound assignments and increments read and count. Repeated loop executions count as one static site. The seven fixture forms are index, at, for-of, spread, slice, map and forEach; named array handout methods are also inventoried by their method name.',
'',
'The source fixture includes mutable and readonly arrays, union, tuple, mutable tuple and intersection receivers. It expects 12 reads with 10 in the readonly subset. The index-skipping mutant fails the same assertion.','']
if source:
 lines += [f"Seven fixture forms alone: {source['seven_form_upper_bound']:,} reads; readonly subset {source['seven_form_readonly_subset']:,}. Additional named array roads account for the remainder of the broader total.",'']
 for title,key in [('Per form','forms'),('Top ten seven-form files','seven_form_files'),('Top ten broader-road files','files'),('Per file','files')]:
  if title=='Per file':
   lines += ['### Per file','','| File | Seven forms | Seven readonly | Broader roads | Broader readonly |','|---|---:|---:|---:|---:|']
   for name,r in source['seven_form_files'].items():
    broad=source['files'][name];lines.append(f"| {name} | {r['upper_bound']} | {r['readonly_subset']} | {broad['upper_bound']} | {broad['readonly_subset']} |")
   continue
  lines += ['### '+title,'','| '+('Form' if key=='forms' else 'File')+' | Upper bound | Readonly subset |','|---|---:|---:|']
  for name,r in list(source[key].items())[:10 if title.startswith('Top ten') else None]:lines.append(f"| {name} | {r['upper_bound']} | {r['readonly_subset']} |")
lines+=['','## Evidence and status','',
'Completed raw streams, meter baselines, paired deltas, logs and inner SHA256SUMS are packed in pieces/*.tar.gz. SHA256SUMS in this directory identifies every archive. The earlier raw-checkpoint.tar.gz records the superseded admission investigation and interrupted first apply; it is not current census evidence. Reproduction sources live in this directory; the meter and production compiler remain unchanged.','',
'Full closure totals remain unavailable until every planned file is complete.' if not complete else 'Every planned file is complete.','']
if declaration_plan:
 lines += ['## Declaration partition','',f"The frozen shared plan has {declaration_plan['top_units']} top-level declarations owning {declaration_plan['units']} full-mode units, including nested declarations, in {len(declaration_plan['pieces'])} initial pieces. Go inventories from both compilers match the independent stock TypeScript inventory. Each unit belongs to exactly one contiguous top-level interval. Audits reject lost, duplicated and moved units, including a move that preserves the global inventory.",'',
 'The posted four whole-file observations are seed leaves and are never remeasured. New pieces use the scheduling-only selector in full mode; first-error mode is not substituted. Selection-off output is byte-identical to the previously proven counted and original meter fixture streams on both compilers, and disjoint selected fixture units and lowering findings exactly reconstruct the whole fixture. New chunks run main then head and are paired immediately. They do not independently rerun a third and fourth meter binary: the original field-identity proof and the new selector-off proof are retained in declaration-selector-proof.tar.gz. This distinction avoids claiming a per-piece four-run identity test that was not performed.','',
 'Each new paired piece has a combined ten-minute limit, with raw output and timers saved even on interruption. An interrupted interval is never rerun unchanged: its immutable refinement partitions its top-level atoms and all owned units between two contiguous children. A single atomic declaration that times out is reported blocked rather than silently omitted. The shared root INDEX stays frozen and active refinements must preserve its complete inventory. Raw piece payloads are packed into rolling declaration-data archives to keep the file count bounded.','']
 if declaration_state and declaration_state.get('stop'):lines += ['Current stop: '+declaration_state['stop']['reason']+' at '+declaration_state['stop']['utc']+'. Next unrun piece: '+declaration_state['stop']['next_piece']+'.','']
elif (p/'STOPPED.md').exists():lines += ['## Stop point','',(p/'STOPPED.md').read_text().split('\n',1)[1]]
(p/'REPORT.md').write_text('\n'.join(lines))
(p/'SHA256SUMS').write_text(''.join(f'{hashlib.sha256(x.read_bytes()).hexdigest()}  {x.relative_to(p)}\n' for x in sorted(p.rglob('*.tar.gz'))))
