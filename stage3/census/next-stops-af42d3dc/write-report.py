import json
from pathlib import Path
out=Path(__file__).resolve().parent
load=lambda name:json.loads((out/name).read_text())
groups=load('grouped-first-stops.json');artifacts=load('root-order.json');p=load('provenance.json');excerpts=load('source-excerpts.json');emitters=load('new-stop-emitters.json')
def esc(s):return str(s).replace('|','&#124;').replace('\n',' ')
a=[ '# Next stops after the core.ts wave', '', 'Task b5w3ycg. Report only, not for landing.', '',f"Main/compiler/apply SHA: `{p['main_sha']}`. TypeScript 6.0.3: `{p['typescript_pin']}`.", '', 'Both requested runs use the same completed full apply. No ordinary compiler overlay, adaptation subset or partial-root fallback is used.', '', '## Grouped first stops', '', 'Roots are relative to `src/compiler/`. Every one of the 79 sorted roots is retained.', '', '| Count | Location | Kind | Exact reason | Root names |', '| ---: | --- | --- | --- | --- |']
artifact_roots={r['root'].removeprefix('src/compiler/') for r in artifacts.get('root_order_artifacts',[])}
pending={r['root'] for r in artifacts.get('pending',[])}
for g in groups:
 names=[f'`{r}`'+(' (root-order artifact)' if r in artifact_roots else ' (artifact classification incomplete)' if r in pending else '') for r in g['roots']]
 a.append('| '+ ' | '.join([str(g['count']),'`'+g['where']+'`',g['kind'],esc(g['reason']),', '.join(names)])+' |')
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
a+=['', '## Whole-root lowering', '',f"{len(accepted)}/79 roots lower whole in the ordinary independent-root census: "+', '.join('`'+r+'`' for r in accepted)+'.', 'This credits ordinary Lower acceptance of the entire resolved root program. Native emission and a multi-root aggregate are not measured.', '', '## Comparison', '', 'The 37f8314b report had 73 roots at core.ts:1049:34, Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can\'t hold. The 4186581e state had 73 roots at core.ts:1326:26, getOwnValues, Refused: a cast the runtime can\'t check.', '']
comparison=load('comparison.json');changed=[r for r in comparison if r['previous']!=r['current']]
for g in groups:
 count=sum(r['root'] in g['roots'] for r in changed)
 a.append(f"- {g['count']} root(s) at `{g['where']}` ({g['kind']}): {count} changed from the 37f8314b table; {g['count']-count} unchanged.")
a+=['','`comparison.json` retains the exact previous and current tuple for each root. `previous-REPORT.md` and `previous-grouped-first-stops.json` retain the comparison baseline.','', '## Every new first-stop group: source and decision','']
for x in excerpts:
 a+=['### '+x['where'],'', 'Exact reason: '+x['reason']+'.','',f"Exact adapted source, `{x['name']}`. "+('The whole enclosing function follows.' if x['whole_function'] else f"The exact site line follows; enclosing function {x['scope_name']} spans lines {x['scope_start']} to {x['scope_end']}."),'','```ts',x['text'],'```','']
 for e in emitters[x['where']]:a.append(f"Emitter: `{e['file']}:{e['line']}`, `{e['function']}`. {e.get('note','')}")
a+=['','## Root-order artifacts','']
if artifacts['status']=='complete':
 a+=[f"Classification uses unmodified `stage3/meter/report.py:root_order_checks` and `artifact_lowers_whole`. Meter-adjusted lowering: {artifacts['totals']['lowering']}/79. Root-order artifacts: {len(artifact_roots)}.",'','| Root | Ordinary stop | Node result | Entry result |','| --- | --- | --- | --- |']
 for r in artifacts['readiness_node_checks']:
  entry=r['entry_next_stop'];desc=(entry['kind']+' '+entry['where']+': '+entry['reason']) if entry else ('lowers whole' if r['entry_program_attempted'] else 'incomplete program coverage')
  a.append('| '+' | '.join([r['root'],r['where'],esc(str(r['node']['exit'])+': '+r['node']['first_error']),esc(desc)])+' |')
else:
 a+=['STOPPING REASON: '+artifacts['reason'],'','No artifact label or whole-root lowering credit is inferred from an incomplete entry probe. The ordinary 79-root census is complete. `root-order.json` retains the pending roots and the child\'s actual exit code.']
if artifacts['status']=='incomplete':
 stop_path=out/'ENTRY-STOP.json'
 if stop_path.exists():
  stop=load('ENTRY-STOP.json');last=stop['last_completed_timing']
  a.append(f"Entry child exit {stop['status']['exit_code']} after {stop['status']['seconds']:.3f} seconds: {stop['status']['stopping_reason']}. Last completed timing: `{last['where'] if last else 'none'}`. The interrupted active unit was not recorded. Partial stream records: {stop['raw_records']}. See ENTRY-STOP.json. `entry-probe.tar.gz` retains the exact stopped checkpoint bundle, with its own SHA-256 file.")
 a.append('')
 for row in artifacts['pending']:
  node=row.get('node')
  if node:a.append(f"- `{row['root']}`: artifact classification incomplete; Node child exit {node['exit']}, {node['first_error']}.")
a+=['','## Method, limits and evidence','', 'The ordinary census was built from the pinned main compiler. Four root groups contain 20, 20, 20 and 19 roots. Each root invocation emits the independent observation and an identical single-root aggregate. The runner verifies all non-timing fields agree and counts the first observation only. Non-TypeScript rows and a 79-root aggregate are excluded.', '', 'The separate entry probe uses the meter guarded entry overlay with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1. Physical scratch Go files use .go.txt; virtual overlay paths preserve the original compiler paths. The probe has an 1800-second bound. `supervise.py` records actual child exit codes, enforces 30 minutes without output, kills the entire child process group and reaps the direct child.', '', 'Completed apply and each completed group were pushed to the append-only checkpoint branch. The final branch contains one report commit on main. `CHECKPOINT.json` records the last checkpoint. No force push or main push was used.', '', 'Raw ordinary/entry streams, progress, timing and preparation logs are packed in `raw-streams.tar.gz`; verify with `sha256sum -c raw-streams.tar.gz.sha256`. Summaries, provenance, source/product manifests, emitter evidence and validation remain directly readable. The core.ts report is in `../core-sites-af42d3dc/REPORT.md`.', '', '## Reproduction', '', '```sh','source /workspace/adamic-tools/env.sh','bash cloud/setup.sh > setup.log 2>&1','npm ci --prefix stage3/api > npm-ci.log 2>&1','export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source','python3 stage3/census/next-stops-af42d3dc/supervise.py apply bash stage3/apply.sh --build-product /workspace/after-wave-adapted > apply-runner.log 2>&1','python3 stage3/census/next-stops-af42d3dc/prepare-tools.py > prepare-tools.log 2>&1','python3 stage3/census/core-sites-af42d3dc/run-census.py > core-runner.log 2>&1','python3 stage3/census/next-stops-af42d3dc/run-groups.py > groups.log 2>&1','python3 stage3/census/next-stops-af42d3dc/run-entry.py > entry-runner.log 2>&1','```','']
(out/'REPORT.md').write_text('\n'.join(a));print('report written')
