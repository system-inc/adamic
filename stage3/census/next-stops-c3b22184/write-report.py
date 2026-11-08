import json
from pathlib import Path
out=Path(__file__).resolve().parent
def read(n):return json.loads((out/n).read_text())
def safe(s):return str(s).replace('|','\\|').replace('\r','').replace('\n',' ')
groups=read('grouped-first-stops.json');comparison=read('comparison.json');provenance=read('provenance.json');order=read('root-order.json');box=provenance['box']
pending={r['root']:r['classification'] for r in order.get('pending',[])};checks={r['root'].split('src/compiler/')[-1]:r for r in order.get('readiness_node_checks',[])}
def label(root):
 value=pending.get(root)
 if root in checks:value='root-order artifact' if checks[root]['node']['throws'] else 'counts against root file'
 return '`'+root+'`'+(' ('+value+')' if value else '')
lines=['# Next stops on main c3b22184','','Adamic, #baobab. Measured census, report only, not for landing.','',
 'Compiler/apply SHA: `'+provenance['compiler_sha']+'`. TypeScript 6.0.3 source pin: `'+provenance['typescript_pin']+'`.','',
 f"Box: `{box['hostname']}`; {box['nproc']} visible CPUs; CPU quota `{box['cpu_quota']}`; {box['memory']}.",'',
 'All 79 compiler roots use one completed full apply at `/workspace/c3b22184-adapted` and the ordinary census tool built from the pinned main. Compiler and adapter source diffs are clean. Hash manifests identify the tool and adapted source bytes. The ordinary census uses no overlay.','',
 '## Files lowering whole','']
accepted=[r['root'] for r in comparison if r['current'][1]=='accepted']
lines+=[str(len(accepted))+'/79 roots lower whole in the ordinary independent-root census: '+(', '.join('`'+r+'`' for r in accepted) or 'none')+'.',
 'This is Lower acceptance of each entire resolved root program. No native output is measured. Artifact roots receive no additional credit without the meter evidence.','',
 '## Grouped first stops','',
 'Root names are relative to `src/compiler/`. All roots are retained, including readiness candidates. Groups are sorted largest first. The accepted row is retained for coverage.','',
 '| Roots | Stop | Kind | Construct | Exact reason | Compared with bcbe5d12 | Forecast | Root names |',
 '| ---: | --- | --- | --- | --- | --- | --- | --- |']
for g in groups:lines.append('| '+' | '.join([str(g['count']),'`'+g['where']+'`',g['kind'],'`'+safe(g['construct']).replace('`','')+'`',safe(g['reason']),g['comparison'],safe(g['forecast']),', '.join(label(r) for r in g['roots'])])+' |')
lines+=['','## Comparison with bcbe5d12','',
 'Baseline compiler/apply: `bcbe5d1244ee8a218d860415dea19cca114bae46`, report commit `30057ecaa2d71ee183e20f32ad7f3e4def6b5bb2`. The exact prior report and grouped JSON are retained. `comparison.json` contains old and current location/kind/reason tuples, constructs, and forecasts for every root.','',
 'The comparison column says unchanged when every root retains its exact tuple, moved when roots change to another forecasted or already known stop, and new for a previously unseen tuple outside the forecasts. The separate forecast column labels every forecast miss new, including unchanged baseline groups.','',
 '| Current group | Roots | Exact tuple unchanged | Roots moved | New tuple in this table |','| --- | ---: | ---: | ---: | --- |']
for g in groups:lines.append('| '+' | '.join(['`'+g['where']+'` '+g['kind'],str(g['count']),str(g['unchanged']),str(g['moved']),'yes' if g['new_tuple'] else 'no'])+' |')
old73=set(next(g for g in read('previous-grouped-first-stops.json') if g['count']==73)['roots'])
lines+=['','Where the prior 73 collator roots stop now:','', '| Current stop | Roots from the old collator group |','| --- | ---: |']
for g in groups:
 count=len(old73 & set(g['roots']))
 if count:lines.append('| `'+g['where']+'` '+g['kind']+' | '+str(count)+' |')
lines+=['','## Forecasts from #b5w3ycg','',
 'Matching identifies the source construct when coordinates shift. `forecast-anchors.json` records the old source lines and verifies their files against the bcbe5d12 source manifest. Named overload walls match the enclosing declaration and the overload-result reason. Every stop matching none is new under the requested forecast rule.','',
 '| Forecast | Matching roots | Measured stop |','| --- | ---: | --- |']
forecasts=['debug.ts:537 cast','debug.ts:577 cast','semver.ts:287','performanceCore.ts:29','performance.ts:189','tryGetPerformanceHooks performance.now presence test (Lower a3)','findAncestor overload-result wall','serializePropertySymbolsForClassOrInterface overload-result wall']
for forecast in forecasts:
 matches=[g for g in groups if g['forecast']==forecast]
 lines.append('| '+forecast+' | '+str(sum(g['count'] for g in matches))+' | '+(', '.join('`'+g['where']+'`' for g in matches) or 'not reached as a first stop')+' |')
lines+=['','## Every group outside the forecasts: source and lowering decision','']
emitters={e['where']:e for e in read('new-stop-emitters.json')}
for row in read('source-excerpts.json'):
 e=emitters[row['where']]
 lines+=['### '+row['where'],'','Construct: `'+row['construct'].replace('`','')+'`','', 'Exact reason: `'+safe(row['reason'])+'`.','',
 'Exact source excerpt: `'+row['name']+'`, '+('whole enclosing function' if row['whole_function'] else 'lines around the stop')+'. Display line endings are normalized; the text file retains the extracted bytes.','',
 '```ts',row['text'].replace('\r\n','\n').replace('\r',''),'```','',
 f"Emitter: `{e['file']}:{e['line']}`, `{e['function']}` (function starts at line {e['function_line']}). Caller: `{e.get('caller_file',e['file'])}:{e['caller_line']}`, `{e['caller']}`. "+e['note'],'']
lines+=['## Root-order artifacts','',
 'Classification follows the current main revision of `stage3/meter/report.py`. Node throws alone are insufficient: the stop source must have attempted entry units and no entry readiness finding there. Candidate roots stay visible in the grouped table.','']
if order['status']=='incomplete':
 stop=read('ENTRY-STOP.json');status=stop['status'];last=stop.get('last_completed_timing')
 lines+=['STOPPING REASON: '+order['stopping_reason'],'',
 f"Entry child exit {status['exit_code']} after {status['seconds']:.3f} seconds: {status['stopping_reason']}. The ordinary 79-root census completed. Entry stream records: {stop['raw_records']}. Last completed timing: "+('`'+last['where']+'`' if last else 'none')+'. The interrupted active unit was not recorded.','',
 '| Root | Independent stop | Node exit | Node first error | Label |','| --- | --- | ---: | --- | --- |']
 for row in order['pending']:
  node=row.get('node') or {};lines.append('| '+' | '.join(['`'+row['root']+'`','`'+row['where']+'`',str(node.get('exit')),safe(node.get('first_error','unavailable')),row['classification']])+' |')
else:
 import sys
 sys.dont_write_bytecode=True;sys.path.insert(0,'/workspace/adamic/stage3/meter')
 from report import root_order_table,artifact_lowers_whole
 credited=[r['root'] for r in order['root_order_artifacts'] if artifact_lowers_whole(r)]
 lines+=root_order_table('ordinary roots',dict(totals=dict(lowering=len(accepted)+len(credited),source_files=79),**order))
 if credited:lines+=['','Additional meter-proven artifact programs lowering whole: '+', '.join('`'+r+'`' for r in credited)+'.']
lines+=['','## Method and evidence','',
 'The same recipe as next-stops-bcbe5d12: source the tools environment, run cloud/setup.sh, npm ci --prefix stage3/api, and export ADAMIC_TYPESCRIPT_SOURCE to the pinned pristine source. Complete a fresh full apply. Run the ordinary per-file census sequentially in groups of 20, 20, 20 and 19, using the same Path sorting as the reference runner. Each invocation emits two single-root observations; all non-timing fields must agree, and only the first counts.','',
 'The separate entry probe uses the meter guarded overlay, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and the prior 1800-second elapsed bound. A supervisor also stops any child after 1800 seconds without output, records the actual child exit code, kills its process group, and reaps it. Physical scratch Go files use .go.txt. Node readiness probes use the meter helper.','',
 'The completed apply and each completed root group were pushed to `typescript/next-stops-c3b22184-checkpoint`. The requested branch contains one report commit directly on c3b22184. No force push or main push is used.','',
 'Raw streams, timing, progress, setup, apply, tool-build, probe logs, and the guarded overlay are packed in `raw-streams.tar.gz`, with `raw-streams.tar.gz.sha256` and a member hash manifest. Verify with `sha256sum -c raw-streams.tar.gz.sha256`. Direct JSON retains the complete root ledger, comparison, forecasts, provenance, process cleanup, and final verification.','',
 'Reproduction commands and actual invocations are in prepare.py, measure.py, run-groups.py, prepare-entry.py, and run-entry.py. All log destinations and child exit codes are retained.','']
text='\n'.join(lines);assert '\u2013' not in text and '\u2014' not in text
(out/'REPORT.md').write_text(text);print('report_groups='+str(len(groups)))
