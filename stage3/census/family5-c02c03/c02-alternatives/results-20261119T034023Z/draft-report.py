import json,collections
from pathlib import Path
W=Path('/tmp/family5-c02-alternatives');state=json.loads((W/'state.json').read_text())
lines=['# C02 whole reach and scratch alternatives','',f"Base `{state['main']}`; C02 `{state['c02']}`. TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. No adaptation, ledger or internal/ edit, and no landing push. All measurements ran sequentially on the same box. Candidate changes exist only in scratch trees; patches and manifests are evidence.",'','## Whole-entry first-error reach','','[Every C02 callable-field site and every new, gone, changed or reached finding, with full before/after words](../whole-entry/REPORT.md). Both raw streams were pushed immediately after completion, before the next side started.','', '| Side | Real exit | Wall seconds | Stream commit |','|---|---:|---:|---|']
for side in ['base','c02']:
 p=W/'entry'/side/'status.json'
 if p.exists():
  s=json.loads(p.read_text());lines.append(f"| {side} | {s['exit_code']} | {s['wall_seconds']} | `{state['entry_pushes'].get(side,'not pushed')}` |")
p=W/'entry/c02/callable-sites.json'
if p.exists():
 d=json.loads(p.read_text());lines+=['',f"At C02, `{d['phrase']}` appears in {d['events']} events: {d['primary_distinct_positions']} distinct primary diagnostic positions, {d['all_distinct_positions']} positions including Boundary mirrors."]
p=W/'comparisons/whole-entry/comparison.json'
if p.exists():lines+=['','```json',json.dumps(json.loads(p.read_text())['files'],indent=2),'```']
lines+=['','Whole-entry pairing: **new 0, gone 14, changed 1, reached 0**. The callable refusal is at utilities.ts:5339:11 in getFunctionFlags (unit 5317:1), replacing the pre-existing NotYet about checked-view modifiers. This is one observable site in the requested first-error pass, not a global full-mode count.','','## Scratch candidates','','(a): `/** @internal */ getPositionOfLineAndCharacter?(line: number, character: number, allowEdits?: true): number;` on compiler SourceFile, types.ts:4326 owner.','','(b): `/** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;` on that same SourceFile owner.','','(c): the 14 source expressions gain `satisfies SourceFileLike`; type-only imports are added in checker, parser and program. No declaration change. Stock TypeScript retains the expression’s source type; the typing evidence below reports it rather than assuming this is a narrower view.','','Compiler stock checks use all 79 roots in both the inherited and explicit strict configurations. A separate stock services-source program tests the merge with services/types.ts:172, without project-reference declaration output hiding it. Its baseline and introduced diagnostics are retained. Commented and comment-free ES2024/ESNext JavaScript emits are compared on every changed source; actual emitted bundles and public declarations are compared separately. Missing harness bundles are identified as not built by the local task, not claimed identical.','','| Candidate | Stock 79 | Strict 79 | Services probe | Source JS | Emitted JS | Public build | Public cmp |','|---|---:|---:|---:|---:|---:|---:|---:|']
candidate_table_end=len(lines);candidate_rows=[]
for name in ['a','b','c']:
 d=state.get('candidates',{}).get(name,{})
 def code(label):
  p=(W/'node-controls'/name/'status.json') if label=='services-merge' and (W/'node-controls'/name/'status.json').exists() else W/'candidates'/name/label/'status.json';return json.loads(p.read_text())['exit_code'] if p.exists() else '—'
 candidate_rows.append('| '+name+' | '+' | '.join(str(code(x)) for x in ['stock','stock-strict','services-merge','javascript','emitted-bundles','public-build','public-compare'])+' |')
 for label in ['stock','stock-strict','services-merge']:
  p=(W/'node-controls'/(name+'.json')) if label=='services-merge' and (W/'node-controls'/(name+'.json')).exists() else W/'candidates'/name/(label+'.json')
  if not p.exists():continue
  r=json.loads(p.read_text());lines += ['',f"{name} / {label}: {r['root_count']} roots, {r['diagnostic_count']} diagnostics."]
  for diag in r['diagnostics']:lines+=['',f"- TS{diag['code']} `{diag.get('file')}:{diag.get('line')}:{diag.get('column')}`: {diag['message']}"]
 p=W/'candidates'/name/'emitted-bundles.json'
 if p.exists():
  r=json.loads(p.read_text());lines+=['',name+' emitted JavaScript: '+str(r['compared'])+' files compared, '+str(r['different'])+' differences. Baseline-only files not produced in this build: `'+json.dumps(r['baseline_only_not_built_by_local_task'])+'`. A failed public build does not establish identity for files it failed to produce.']
 p=W/'candidates'/name/'public.json'
 if p.exists():lines+=['',name+' public API:','```json',p.read_text(),'```']
 p=(W/'node-controls'/(name+'-delta.json')) if (W/'node-controls'/(name+'-delta.json')).exists() else W/'candidates'/name/'services-merge-delta.json'
 if p.exists():lines+=['',name+' services merge relative to the stock baseline:','```json',p.read_text(),'```']
 if d and not d.get('eligible_for_full_census'):lines+=['',name+' full census not run: the compiler input is not checker-clean, or (a) was killed by its introduced merge error, as instructed.']
lines[candidate_table_end:candidate_table_end]=candidate_rows
lines+=['','The first unified services-source probes inherited services’ empty types list and therefore reported 64 missing-Node-global diagnostics at base. Those raw attempts are retained. The table and merge diagnostics above use the corrected Node-global controls when present; their baseline is reported separately below.']
if (W/'node-controls/base.json').exists():
 r=json.loads((W/'node-controls/base.json').read_text());lines+=['','Corrected services baseline: '+str(r['root_count'])+' roots, '+str(r['diagnostic_count'])+' diagnostics; real exit '+str(state['node_controls']['base']['exit_code'])+'.']
lines+=['','## Every proposed site annotation in (c)','','| Original site | Source expression | Proposed expression | Type-only |','|---|---|---|---|']
p=W/'candidates/c/edits.json'
if p.exists():
 for e in json.loads(p.read_text()):lines.append(f"| `{e['file']}:{e['line']}:{e['column']}` | `{e['before']}` | `{e['after']}` | yes |")
p=W/'candidates/c/stock.json'
if p.exists():
 lines+=['','Stock TypeScript typing for the annotated expressions:','```json',json.dumps(json.loads(p.read_text()).get('satisfies_views',[]),indent=2),'```']
lines+=['','## Full single-file censuses','','Physical S/T/P site counts are deduplicated by diagnostic position. Pairing grades count primary events by phase/unit; a full census can revisit one position in several units. Boundary mirrors are counted separately.\n\nAll full runs use the tsc.ts entry, exact absolute LATENT_ONLY_FILE, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1 and CommonJSBundle. utilities.ts, parser.ts, program.ts and checker.ts are the four distinct files holding the 14 G02 sites; utilities.ts also contains getFunctionFlags. Complete streams, complete stderr, real child exits, source and tool hashes are retained. The previously completed isolated utilities baseline is reused only when its tool, types.ts and utilities.ts hashes match. It is not counted as a new run.','']
lines+=['| Candidate | G02 before | G02 after | New | Gone | Changed | Reached | All stock/emit/API prerequisites green |','|---|---:|---:|---:|---:|---:|---:|---|']
for name in ['a','b','c']:
 p=W/'comparisons'/('candidate-'+name)/'comparison.json'
 if not p.exists():lines.append('| '+name+' | — | — | — | — | — | — | no completed census comparison |');continue
 q=json.loads(p.read_text());g=collections.Counter(x['grade'] for x in q['changes'] if x.get('after',x.get('before',{})).get('kind')!='Boundary');lines.append('| '+name+' | '+str(sum(f['target_before_distinct_sites'] for f in q['files'].values()))+' | '+str(sum(f.get('target_after_equivalent_fact_distinct_sites',f['target_after_distinct_sites']) for f in q['files'].values()))+' | '+' | '.join(str(g[k]) for k in ['new','gone','changed','reached'])+' | '+str(state['candidates'][name]['landing_prerequisites_green'])+' |')
viable=[]
for name,d in state.get('candidates',{}).items():
 p=W/'comparisons'/('candidate-'+name)/'comparison.json'
 if not p.exists():continue
 result=json.loads(p.read_text());lines+=['',name+':','```json',json.dumps(result['files'],indent=2),'```']
 primary=collections.Counter(x['grade'] for x in result['changes'] if x.get('after',x.get('before',{})).get('kind')!='Boundary');mirrors=collections.Counter(x['grade'] for x in result['changes'] if x.get('after',x.get('before',{})).get('kind')=='Boundary');lines+=['','Primary (non-Boundary) event grades: `'+json.dumps(dict(primary))+'`; Boundary mirror grades: `'+json.dumps(dict(mirrors))+'`.']
 outside=collections.Counter(x['grade'] for x in result['changes'] if x['grade'] in ['new','changed'] and not x.get('within_original_target_sites',False) and x.get('after',x.get('before',{})).get('kind')!='Boundary');lines+=['','New/changed primary events outside the original G02 sites: `'+json.dumps(dict(outside))+'`.']
 target=sum(x.get('target_after_equivalent_fact_distinct_sites',x['target_after_distinct_sites']) for x in result['files'].values());growth=[x for x in result['changes'] if x['grade'] in ['new','changed']]
 if len(result['files'])==4 and target==0 and not growth and d.get('landing_prerequisites_green'):viable.append(name)
 for i,ch in enumerate(result['changes'],1):
  lines+=['',f"### {name}.{i}: {ch['grade']}",'', 'Scope: '+('original G02 site or its diagnostic mirror' if ch.get('within_original_target_sites') else 'outside original G02 sites')+'.']
  for side in ['before','after']:
   f=ch.get(side)
   if f:lines += [f"{side}: `{f['where']}`; unit `{f.get('unit','')}`; {f['kind']} / {f['phase']}.",'','Reason: '+f['reason'],'','```text',f['text'],'```']
  if ch.get('reach_proof'):lines+=['','```json',json.dumps(ch['reach_proof'],indent=2),'```']
lines+=['','## Growth outside G02 in candidate (c)','','The nine primary new findings below are outside the original 14 G02 sites. They are new, not reached: the annotated relation still fails, so the G02 presence obligation has not been discharged. The full-event list above also retains their rollback Boundary mirrors. Coordinates below are in the scratch candidate after insertion.','','| Site | Kind / phase | Full message |','|---|---|---|']
p=W/'comparisons/candidate-c/comparison.json'
if p.exists():
 for ch in json.loads(p.read_text())['changes']:
  f=ch.get('after',{})
  if ch['grade'] in ['new','changed'] and not ch.get('within_original_target_sites') and f.get('kind')!='Boundary':
   lines.append('| `'+f['where']+'` | '+f['kind']+' / '+f['phase']+' | '+f['text'].replace('|','\\|')+' |')
lines+=['','## Every child exit','','| Receipt | Real exit | Wall seconds |','|---|---:|---:|']
for p in sorted(W.rglob('status.json')):
 if 'entry-pair-' in str(p):continue
 s=json.loads(p.read_text());lines.append(f"| {p.parent.relative_to(W)} | {s['exit_code']} | {s.get('wall_seconds','—')} |")
lines+=['','For (c), distinguish disappearance of the literal old reason from proof of the shape. The explicit satisfies relation uses `an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type`. This is the same unresolved S/T/P presence obligation. Such a replacement is graded changed, not reached. The summary G02-after column counts both diagnostic forms; the JSON also retains the literal-old-reason count separately. Stock TypeScript still reports SourceFile as the result type.','','## Result','']
if state.get('stop'):lines += ['Stopped: '+state['stop']+'. Uncompleted lines are not claimed green.']
elif state.get('alternatives_complete') and not viable:
 lines += ['No candidate meets all requested checks. (a) is killed by TS2386; (c) keeps all 14 missing-property obligations under the explicit-relation diagnostic and adds nine primary findings outside G02. Distinguish the lowering result from source/API legality: (b) clears all 14 G02 sites with zero new, changed or reached findings in the four measured files, but its optional property conflicts with the services required method and the public build fails. This is not a landing-ready declaration. C02 does not qualify as an established-shape adaptation under this delegation; return the cause to Compiler for a different road. No landing was pushed.']
elif viable:lines += ['Candidates meeting the measured four-file and stock/emit criteria: '+', '.join(viable)+'. This is a scout measurement, not a landing. Nothing was pushed to a landing branch.']
else:lines += ['Measurements are incomplete; no conclusion for unrun lines.']
(W/'REPORT.md').write_text('\n'.join(lines)+'\n');print('Drafted',W/'REPORT.md')
