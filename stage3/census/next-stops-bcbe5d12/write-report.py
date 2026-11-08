import json
from pathlib import Path
out=Path(__file__).resolve().parent
def read(name):return json.loads((out/name).read_text())
def safe(s):return str(s).replace('|','\\|').replace('\r','').replace('\n',' ')
groups=read('grouped-first-stops.json');comparison=read('comparison.json');provenance=read('provenance.json');order=read('root-order.json');box=provenance['box']
pending={r['root']:r['classification'] for r in order.get('pending',[])}
checks={r['root'].split('src/compiler/')[-1]:r for r in order.get('readiness_node_checks',[])}
def root_label(root):
 label=pending.get(root)
 if root in checks:label='root-order artifact' if checks[root]['node']['throws'] else 'counts against root file'
 return '`'+root+'`'+(' ('+label+')' if label else '')
lines=['# Next stops on main bcbe5d12','','Task b5w3ycg. Measured, report only, not for landing.','',
 'Compiler and apply SHA: `'+provenance['compiler_sha']+'`. TypeScript 6.0.3 source pin: `'+provenance['typescript_pin']+'`.','',
 f"Box: `{box['hostname']}`; {box['nproc']} visible CPUs; cgroup CPU quota `{box['cpu_quota']}` (4 CPU equivalents); {box['memory']}. Managed environment: `ccarenv_b64_Y2NhcmVudl82MWJjNmEwMjE3Yzg4MTkxODEyY2I3ODNkM2JiMGQ2ZQ`.",'',
 'One completed full apply produced `/workspace/bcbe5d12-adapted`. All 79 ordinary independent compiler roots use this tree and the ordinary census tool built from bcbe5d12. The compiler and adapter input diff is clean; hashes are retained in the source and binary manifests. The ordinary census uses no overlay.','',
 '## Grouped first stops','',
 'Roots are relative to `src/compiler/`. Counts include every root, including readiness candidates. Groups are sorted largest first.','',
 '| Roots | Location | Kind | Exact reason | Forecast match | Root names |','| ---: | --- | --- | --- | --- | --- |']
for g in groups:lines.append('| '+ ' | '.join([str(g['count']),'`'+g['where']+'`',g['kind'],safe(g['reason']),safe(g['forecast']),', '.join(root_label(r) for r in g['roots'])])+' |')
accepted=[r['root'] for r in comparison if r['current'][1]=='accepted']
lines+=['','## Whole-root lowering','',f"{len(accepted)}/79 roots lower whole in the ordinary independent-root census: "+', '.join('`'+r+'`' for r in accepted)+'.',
 'This is ordinary Lower acceptance of each entire resolved root program. No native output is measured. Readiness candidates receive no additional lowering credit without the meter evidence.','',
 '## Comparison with af42d3dc','',
 'Baseline compiler/apply: `af42d3dc1261a3db536064c3dfbb5eaa7e8b6405`. The prior report and grouped JSON are retained beside this report. `comparison.json` records both exact tuples and the forecast for every root.','',
 '| Current group | Roots | Unchanged exact tuple | Moved exact tuple | New tuple in this table |','| --- | ---: | ---: | ---: | --- |']
for g in groups:lines.append('| '+ ' | '.join(['`'+g['where']+'` '+g['kind'],str(g['count']),str(g['unchanged']),str(g['moved']),'yes' if g['new_tuple'] else 'no'])+' |')
lines+=['','The collator operation moved from `core.ts:2116:26` to `core.ts:2103:26`; its kind and exact reason are unchanged. A moved coordinate alone is not a newly exposed compiler limitation. The `core.ts` root moved from the old generic-instantiation stop at `debug.ts:243:59` to the new method-value stop at `debug.ts:355:26`. The accepted group and the three readiness groups are unchanged.','',
 '## The forecasts','',
 'The forecast column applies to every root in its row; `comparison.json` gives one record per root. Matching uses the source operation when adaptations shift coordinates. Under the requested rule, every stopped group matching none of the six forecasts is called new below, even if its exact tuple was already present in the prior table. Accepted roots have no stop and no forecast match.','',
 '| Forecast | Matching roots | Measured location |','| --- | ---: | --- |']
for name in ['core.ts:2116','debug.ts:537','debug.ts:577','semver.ts:287','performanceCore.ts:29','performance.ts:189']:
 matches=[g for g in groups if g['forecast'].startswith(name)]
 lines.append('| `'+name+'` | '+str(sum(g['count'] for g in matches))+' | '+(', '.join('`'+g['where']+'`' for g in matches) or 'not reached as a first stop')+' |')
lines+=['','## Every group outside the forecasts: source and decision','']
excerpts=read('source-excerpts.json');emitters={e['where']:e for e in read('new-stop-emitters.json')}
for row in excerpts:
 e=emitters[row['where']]
 lines+=['### '+row['where'],'','Exact reason: '+safe(row['reason'])+'.','',
 'Source: `'+row['name']+'`, '+('whole enclosing function' if row['whole_function'] else 'lines around the stop')+'. The text file preserves the source bytes; line endings are normalized only in this display.','',
 '```ts',row['text'].replace('\r\n','\n').replace('\r',''),'```','',
 f"Emitter: `{e['file']}:{e['line']}` in `{e['function']}` (function begins at line {e['function_line']}). Caller: `{e.get('caller_file',e['file'])}:{e['caller_line']}`, `{e['caller']}`. "+e['note'],'']
lines+=['## Root-order artifacts','',
 'The classification follows `stage3/meter/report.py`: Node throws alone do not establish eligibility. The stop source must have attempted entry units and no entry readiness finding there. Artifact roots remain explicit in the first-stop table.','']
if order['status']=='incomplete':
 stop=read('ENTRY-STOP.json');st=stop['status'];last=stop['last_completed_timing']
 lines+=['STOPPING REASON: '+order['stopping_reason'],'',
 f"Entry child exit {st['exit_code']} after {st['seconds']:.3f} seconds: {st['stopping_reason']}. The ordinary 79-root census completed. The entry stream has {stop['raw_records']} record(s). Last completed timing: "+('`'+last['where']+'`' if last else 'none')+'. The interrupted active unit was not recorded.','',
 'Artifact classification is incomplete for the following roots; none is folded away or credited as lowering whole.','',
 '| Root | Independent stop | Node child exit | Node first error | Label |','| --- | --- | ---: | --- | --- |']
 for row in order['pending']:
  node=row.get('node') or {};lines.append('| '+ ' | '.join(['`'+row['root']+'`','`'+row['where']+'`',str(node.get('exit')),safe(node.get('first_error','unavailable')),row['classification']])+' |')
else:
 import sys
 sys.path.insert(0,'/workspace/adamic/stage3/meter');from report import root_order_table
 lines+=root_order_table('ordinary roots',dict(totals=dict(lowering=len(accepted),source_files=79),**order))
lines+=['','## Method and evidence','',
 'The roots are sorted and run sequentially in groups of 20, 20, 20 and 19. Each root invocation emits an independent observation and a single-root aggregate. The runner verifies all non-timing fields agree, retains both raw observations, and counts the first once. Non-source rows and a multi-root aggregate do not enter the denominator.','',
 'The separate entry probe follows the prior report: guarded meter overlay, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and a 1800-second elapsed bound. The supervisor also stops a child after 1800 seconds without output, records its actual exit code, kills its process group, and reaps it. Scratch Go sources use .go.txt. Node probes use the meter helper unchanged.','',
 'Preparation failures before the dependency checkout completed are retained in the raw archive. Their completed replacement builds exited 0. The complete apply and every completed root group were pushed to `typescript/next-stops-bcbe5d12-checkpoint`. The final requested branch contains one report commit on the pinned main; no force push or main push is used.','',
 "The first verification checked string ordering rather than the reference runner's Path ordering. That order-only assertion was corrected; the same 79 roots and raw streams passed. The archived initial verification failure is retained, and validation.json records the final successful checks.",'',
 'Raw streams, progress, timing, setup, apply, probe and build logs, plus the guarded overlay, are in `raw-streams.tar.gz`. Verify with `sha256sum -c raw-streams.tar.gz.sha256`. Direct JSON files retain per-root outcomes, comparisons, provenance, cleanup, and validation. Lowering source evidence is retained as .go.txt.','',
 '## Reproduction','',
 '```sh','source /workspace/adamic-tools/env.sh','bash cloud/setup.sh > setup.log 2>&1','npm ci --prefix stage3/api > npm-ci.log 2>&1',
 'export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source',
 '# The source checkout must be at 050880ce59e30b356b686bd3144efe24f875ebc8.',
 'python3 stage3/census/next-stops-bcbe5d12/supervise.py apply bash stage3/apply.sh --build-product /workspace/bcbe5d12-adapted > apply-runner.log 2>&1',
 'go build -o /workspace/scratch/b5w3ycg/census ./stage3/census/tool > census-build.log 2>&1',
 'python3 stage3/census/next-stops-bcbe5d12/run-groups.py > groups-runner.log 2>&1','```','']
text='\n'.join(lines)
assert '\u2013' not in text and '\u2014' not in text
(out/'REPORT.md').write_text(text)
print('report_groups='+str(len(groups)))
