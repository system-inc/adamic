from pathlib import Path
from collections import Counter, defaultdict
import json,gzip
out=Path(__file__).resolve().parent

def load(case,side):
 p=out/(case+'-'+side+'.jsonl')
 if not p.exists():p=Path(str(p)+'.gz')
 opener=gzip.open if str(p).endswith('.gz') else open
 with opener(p,'rt') as f:return [json.loads(s) for s in f if s.strip()]
def findings(rows):return [f for r in rows for f in r.get('findings',[]) or []]
def short(s):return s.replace('/workspace/never-call-adapted/src/','src/')
def esc(s):return str(s).replace('|','\\|').replace('\n','<br>')
def quote(s):return '"'+esc(s)+'"'
def unitlabel(u):
 names=json.loads((out/'entry-unit-names.json').read_text())
 return esc(u.get('name') or names.get(short(u['where']), 'unnamed unit'))+' `'+short(u['where'])+'`'
def reason(f):return ('`'+short(f['where'])+'` '+f['kind']+': '+quote(f['reason'])) if f else 'No lowering stop'
def heading(s):lines.extend(['',s,''])
def table(headers,rows):
 if lines and lines[-1]!='':lines.append('')
 lines.extend(['| '+' | '.join(headers)+' |','| '+' | '.join('---' for _ in headers)+' |'])
 lines.extend('| '+' | '.join(str(v) for v in r)+' |' for r in rows)
 lines.append('')
full={c:{s:load(c,s) for s in ['parent','head']} for c in ['debug-full','semver-full']}
entry_ready=(out/'tsc-entry-ordered-first-paired.json').exists()
cases=['debug-full','semver-full']+(['tsc-entry-ordered-first'] if entry_ready else [])
deltas={c:json.loads((out/(c+'-paired.json')).read_text()) for c in cases}
lines=['# Paired never-call census, task 8n7eet3','',
'Parent `bdbc73864f82dbc0c51434decf0ffe96d4782767` against head `392cb60b7a967a4515f4f0041592359b57746955`, measured November 10, 2026. Report only; no compiler change is included. The remote optimizer branch was checked before setup and still pointed at the requested head.','',
'The seven occurrences at debug.ts:227 disappear in the semver full run: six have the Version-fields target and one has the Comparator-array target. They deduplicate to two diagnostic findings. The standalone debug full run has no changes because those generic declarations have no concrete instantiation witness in that selected file. The full Version run retains its static-initialization refusal at semver.ts:47:41, along with its other existing stops.']
if entry_ready:
 d=deltas['tsc-entry-ordered-first'];raw=d['raw_findings'];distinct=d['distinct_reads']
 extra=json.loads((out/'audit.json').read_text()).get('extra_new_site_regressions_distinct_reads',[]) if (out/'audit.json').exists() else []
 lines[4:4]=[f"Complete entry program: {len(raw['lowered'])} raw / {len(distinct['lowered'])} distinct findings clear; {len(raw['new_findings'])} raw / {len(distinct['new_findings'])} distinct findings are new; {len(raw['moved'])} raw / {len(distinct['moved'])} distinct old reasons change at a site. {len(distinct['unit_first_stop_delta'])} unit first stops change. There are {len(distinct['regressions'])+len(extra)} distinct regression records under the requested definition.",'']
if entry_ready:
 units=deltas['tsc-entry-ordered-first']['distinct_reads']['unit_first_stop_delta']
 cleared_units=[r['unit'].get('name','') for r in units if r['base'] and not r['head']]
 lines[6:6]=['Finding-free head units: '+', '.join('`'+name+'`' for name in cleared_units)+'. The other '+str(sum(bool(r['base'] and r['head']) for r in units))+' changed units remain blocked. Version advances past the checked predicate, but its measured head first stop is debug.ts:252:33, "a value of type never", in both census modes. The later static-initialization stop is still visible in full mode; it is not the first-error head stop.','']
if not entry_ready:lines+=['','Status: full runs complete; the corrected complete entry census is in progress. No entry-program result is claimed yet.']
heading('## Pins, scope and guards')
lines+=['One adapted tree was built from the parent with `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, `ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-6.0.3`, and `bash stage3/apply.sh --build-product /workspace/never-call-adapted`. Stock source is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both compiler commits use cohere `5360753e19e4417675a90e92a1823bea66b13918` and checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go is 1.27.1, Node is v24.19.0. Each side was built from its own SHA with its own scratch overlay. Compiler source stayed unchanged.','',
'Full runs pass the adapted `src/compiler` directory, set `LATENT_FULL=1`, and select exactly one adapted file with `LATENT_ONLY_FILE`. The entry runs load only `src/tsc/tsc.ts` as the original root and use the scratch `Program.LatentReplayEntryReach()` hook, corrected to preserve the original entry as `Files()[0]`, to expand the observation ledger to its 81 resolved source files. The original root remains unchanged. Both entry driver copies are retained as `.go.txt` evidence.','',
'All reported runs set `LATENT_ASSERT_NO_OUTPUT=1`. The ordinary loader is disabled, ordinary lowering must return nil IR and the measurement-only error, and no backend runs. Unit and prepass timings are separate streams. Every child exit is recorded, and the wrapper stops a run after 30 minutes without a change in its output evidence.','',
'The requested reference `stage3/census/core-sites-af42d3dc/REPORT.md` is absent from the parent commit. The specified full-mode flags were used directly. The actual headers report `checker_rejected: false`, with no checker diagnostics. The meter retains its fixed label "measured on a checker-rejected program"; that label does not override the observed checker status. These are lowering censuses, not successful native compilations.','',
'The first entry invocation loaded the program but recorded only the root file. Those two preliminary runs are retained under `tsc-first-*` as excluded diagnostic evidence. The corrected census is exclusively `tsc-entry-ordered-first-*`; its ledger must match the independently enumerated 81-file `closure.json`.']
heading('## Times and peak memory')
rows=[]
for c in cases:
 for side in ['parent','head']:
  t=json.loads((out/(c+'-'+side+'.timing.json')).read_text())
  rows.append([c,side,f"{t['seconds']:.3f}",f"{t['peak_rss_kib']:,}",t['child_exit_code'],str(t['silent_cutoff']).lower()])
table(['Run','Side','Elapsed seconds','Peak RSS KiB','Child exit','Silence cutoff'],rows)
lines+=['The four full runs were sequential. The final two entry runs overlapped to stay within the session window; their wall times include that concurrency and are not an isolated performance comparison. Full-run elapsed time is wrapper wall time, with up to one second of completion-poll overhead; full-run peak RSS uses Linux getrusage(RUSAGE_CHILDREN).ru_maxrss. The entry launchers were interrupted during an environment reconnection while the compilers survived. An exit observer recovered entry wall time from kernel process start ticks and the boot clock, actual wait status from the kernel, and lifetime peak RSS from /proc/PID/status VmHWM at PTRACE_EVENT_EXIT. It forwards ordinary signals unchanged and does not enable EXITKILL. Entry times include this late exit observation and concurrency, so they are not an isolated performance benchmark. A controlled 64 MiB observer probe returned the expected exit 7 and reported 71,552 KiB peak RSS, exactly matching getrusage for that child. All RSS values describe each compiler process, not the container. Setup and builds are excluded.']
heading('## Cleared, new and changed counts')
rows=[]
for c,d in deltas.items():
 for name in ['raw_findings','distinct_reads']:
  v=d[name];rows.append([c,name,len(v['lowered']),len(v['moved']),len(v['new_findings']),len(v['new_sites']),len(v['unit_first_stop_delta']),len(v['regressions'])])
table(['Comparison','View','Cleared','Changed reason at site','New findings','New sites','Changed unit first stop','Regressions'],rows)
lines+=['These are the pair tool\'s exact counts, including measurement `Boundary` records. "Cleared" means an old finding has no remaining finding at its identity. "Changed reason at site" means an old finding disappeared but another remains there. "New findings" includes exposed reasons in already-blocked units; it is not a regression count. Raw identity preserves kind, source identity, reason and text. Distinct identity uses the outermost contiguous access chain and exact reason.','',
'For semver full mode, the four cleared raw findings are two NotYet reasons at debug.ts:227:5 and two Boundary records at debug.ts:235:9. They become three distinct entries: two diagnostic reasons and one boundary reason. The seven original NotYet occurrences and seven accompanying boundary occurrences are removed. No new full-mode finding appears in either selected-file comparison.']
heading('## debug.ts: assertion helpers and callers')
lines+=['Standalone debug full mode records 81 declaration units and 22 finding occurrences on each side, with 19 raw and 19 distinct entries including boundaries. No finding occurs at debug.ts:227 or in lines 227 through 250 in that selected-file run. The six relevant declaration units are deferred on both sides, with `instantiated: false`; absence of findings there is not a claim that their bodies lowered.']
bu={u['where']:u for r in full['debug-full']['parent'] for u in r.get('units',[]) or []};hu={u['where']:u for r in full['debug-full']['head'] for u in r.get('units',[]) or []}
table(['Declaration','Unit','Parent status','Head status','Parent / head reason'],[[u.get('name'),short(w),u['status'],hu[w]['status'],'No finding; generic instantiation deferred on both sides'] for w,u in bu.items() if u.get('name') in ['assertIsDefined','assertEachIsDefined','checkDefined','checkEachDefined']])
inventory=json.loads((out/'inventory.json').read_text())
table(['Direct call in debug.ts','Caller','Call text','Parent / head finding at this local call site'],[[r['where'],r['owner'],'`'+esc(r['expression'])+'`','None in standalone debug full mode'] for r in inventory['debug.ts']['calls']])
lines+=['The semver-selected run supplies concrete caller witnesses and records every occurrence below at debug.ts:227:5. "Absent" means that exact site has no finding in that caller attempt on the head. All rows have phase `lowering`.']
pf=findings(full['semver-full']['parent']);hf=findings(full['semver-full']['head']);units={u['where']:u for r in full['semver-full']['parent'] for u in r.get('units',[]) or []}
rows=[]
for f in pf:
 if '/debug.ts:227:' in f['where']:
  after=[g for g in hf if g['unit']==f['unit'] and g['where']==f['where']]
  rows.append([units.get(f['unit'],{}).get('name','')+' `'+short(f['unit'])+'`',short(f['where']),f['kind']+': '+quote(f['reason']),'; '.join(g['kind']+': '+quote(g['reason']) for g in after) or 'Absent'])
table(['Attempting caller unit','Site','Parent reason','Head reason'],rows)
lines+=['The associated checkDefined call at debug.ts:235:9 loses seven boundary occurrences with reason "failed statement; state rolled back; continued at next statement or declaration". Neither full stream has a finding inside assertEachIsDefined (239 through 245) or at its call in checkEachDefined (248). Every full-mode finding, including unchanged callers and dependency sites, is listed below.']
heading('## semver.ts: Version and Version.zero')
lines+=['Version spans lines 46 through 139; Version.zero is declared at line 47. Full mode continues after failed statements, so its first recorded failure and its later static-field failure are separate observations. The Version unit\'s first recorded failure changes from the checked Version-fields predicate at debug.ts:227:5 to debug.ts:252:33, "a value of type never". The static initializer still reports semver.ts:47:41, "a constructor object escaping before static fields are initialized", on both sides. This is the later static stop referenced as #1451npc in the request. The entry first-error Version unit instead stops at debug.ts:252:33 on the head, so the prediction of the static refusal as its next first stop is not reproduced.']
rows=[]
keys=sorted({(f['where'],f['kind'],f['reason']) for f in pf+hf if '/semver.ts:' in f['where'] and 46<=int(f['where'].rsplit(':',2)[1])<=139})
for where,kind,why in keys:
 a=[f for f in pf if (f['where'],f['kind'],f['reason'])==(where,kind,why)];b=[f for f in hf if (f['where'],f['kind'],f['reason'])==(where,kind,why)]
 rows.append([short(where),kind,quote(why) if a else 'Absent',quote(why) if b else 'Absent',str(len(a))+' / '+str(len(b))])
table(['Every finding site in Version','Kind','Parent reason','Head reason','Occurrences parent / head'],rows)
lines+=['The dependency sites reached while attempting Version itself are also exhaustive here:']
a=[f for f in pf if f['unit'].endswith('/semver.ts:46:1') and '/semver.ts:' not in f['where']];b=[f for f in hf if f['unit'].endswith('/semver.ts:46:1') and '/semver.ts:' not in f['where']]
rows=[]
for where,kind,why in sorted({(f['where'],f['kind'],f['reason']) for f in a+b}):
 rows.append([short(where),kind,quote(why) if any((f['where'],f['kind'],f['reason'])==(where,kind,why) for f in a) else 'Absent',quote(why) if any((f['where'],f['kind'],f['reason'])==(where,kind,why) for f in b) else 'Absent'])
table(['Dependency site','Kind','Parent reason','Head reason'],rows)
heading('## Other full-mode movements')
lines+=['There are no other changed finding sites in either full-mode stream beyond debug.ts:227:5 and its boundary site debug.ts:235:9. Six semver unit first stops change because that earlier finding disappears; all remain blocked.']
table(['Unit','Parent first recorded stop','Head first recorded stop'],[[unitlabel(r['unit']),reason(r['base']),reason(r['head'])] for r in deltas['semver-full']['raw_findings']['unit_first_stop_delta']])
if entry_ready:
 heading('## Complete tsc entry program, first-error mode')
 d=deltas['tsc-entry-ordered-first'];v=d['distinct_reads']
 lines+=['Both streams retain the single original tsc entry root and observe exactly the 81-file closure. Pairing validates identical measurement headers, unit identities and attempted eligibility. A changed first stop in an already-blocked unit is listed separately below and does not count as a regression.']
 table(['Every moved unit','Parent first stop','Head first stop'],[[unitlabel(r['unit']),reason(r['base']),reason(r['head'])] for r in v['unit_first_stop_delta']])
 heading('### All entry findings in the requested assertion and Version scopes')
 parent_findings=findings(load('tsc-entry-ordered-first','parent'));head_findings=findings(load('tsc-entry-ordered-first','head'))
 def in_requested_scope(f):
  parts=f['where'].rsplit(':',2)
  if len(parts)!=3 or not parts[1].isdigit():return False
  file,line,_=parts;line=int(line)
  return (file.endswith('/debug.ts') and 227<=line<=250) or (file.endswith('/semver.ts') and 46<=line<=139)
 scoped=[f for f in parent_findings+head_findings if in_requested_scope(f)]
 rows=[]
 for where,kind,why in sorted({(f['where'],f['kind'],f['reason']) for f in scoped}):
  a=[f for f in parent_findings if (f['where'],f['kind'],f['reason'])==(where,kind,why)];b=[f for f in head_findings if (f['where'],f['kind'],f['reason'])==(where,kind,why)]
  rows.append([short(where),kind,quote(why) if a else 'Absent',quote(why) if b else 'Absent',str(len(a))+' / '+str(len(b)),'<br>'.join(sorted({short(f['unit']) for f in a+b}))])
 table(['Site','Kind','Parent reason','Head reason','Occurrences parent / head','Every attempting caller unit'],rows)
 heading('### Every changed entry-program finding')
 rows=[]
 for item in v['lowered']:
  f=item['base'];rows.append(['Cleared',f['phase'],unitlabel({'where':f['unit']}),reason(f),'Absent at this finding identity'])
 for item in v['moved']:
  f=item['base'];rows.append(['Changed reason',f['phase'],unitlabel({'where':f['unit']}),reason(f),'<br>'.join(reason(g) for g in item['head'])])
 for item in v['new_findings']:
  f=item['finding'];rows.append(['New finding',f['phase'],unitlabel({'where':f['unit']}),'No old finding with this identity and reason',reason(f)])
 table(['Finding delta','Phase','Attempting unit or scan owner','Parent','Head'],rows)
 heading('### Predicate or never changes outside debug.ts and semver.ts')
 affected=[]
 for r in v['unit_first_stop_delta']:
  if any('/'+name+':' in r['unit']['where'] for name in ['debug.ts','semver.ts']):continue
  if any(f and any(t in f['reason'].lower() for t in ['predicate','never']) for f in [r['base'],r['head']]):affected.append(r)
 if affected:
  table(['Unit','Whole old reason and site','Whole new reason and site'],[[unitlabel(r['unit']),reason(r['base']),reason(r['head'])] for r in affected])
 else:lines+=['None.']
 lines+=['One additional movement outside the requested files is getRawLiteral in transformers/taggedTemplate.ts:110:1. It remains blocked: the old stop is factory/nodeFactory.ts:6111:21, "writing a checked object field without its source-slot type certificate"; the new stop is utilities.ts:10656:6, "a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace". Both whole reasons also appear in the complete first-stop table.']
 heading('### Every regression')
 audit=json.loads((out/'audit.json').read_text()) if (out/'audit.json').exists() else {}
 extra=audit.get('extra_new_site_regressions_distinct_reads',[])
 if extra:
  lines+=['Additional finding-free site regressions from the independent all-phase audit:','```json',json.dumps(extra,indent=2),'```','']
 if v['regressions']:
  for r in v['regressions']:lines+=['```json',json.dumps(r,indent=2),'```','']
 elif not extra:lines+=['None: no previously finding-free attempted unit acquires a stop, no finding-free scan site acquires a finding, and no new panic appears. New lowering sites exposed solely by a changed first stop in an already-blocked unit are reported as stop changes, as requested. The independent audit also checks new sites in phases other than refusal_scan.']
 lines+=['','Raw panics: '+str(len(d['raw_findings']['panics']['before']))+' parent, '+str(len(d['raw_findings']['panics']['after']))+' head. Distinct panics: '+str(len(v['panics']['before']))+' parent, '+str(len(v['panics']['after']))+' head. The paired JSON retains all new sites and reasons for audit.']
heading('## Complete selected-file full-mode finding ledgers')
lines+=['These tables include every recorded finding site and exact reason, including boundaries and dependencies. Counts preserve all attempting-unit occurrences; the paired JSON provides the raw and access-chain identities. "Absent" is absence at that site, not successful lowering of its caller.']
for case in ['debug-full','semver-full']:
 heading('### '+case)
 a=findings(full[case]['parent']);b=findings(full[case]['head']);rows=[]
 for where,kind,why in sorted({(f['where'],f['kind'],f['reason']) for f in a+b}):
  old=[f for f in a if (f['where'],f['kind'],f['reason'])==(where,kind,why)];new=[f for f in b if (f['where'],f['kind'],f['reason'])==(where,kind,why)]
  callers=sorted({short(f['unit']) for f in old+new})
  rows.append([short(where),kind,quote(why) if old else 'Absent',quote(why) if new else 'Absent',str(len(old))+' / '+str(len(new)),'<br>'.join(callers)])
 table(['Site','Kind','Parent reason','Head reason','Occurrences parent / head','Every attempting unit'],rows)
heading('## Coverage and checker-relation audit')
if (out/'audit.json').exists():
 audit=json.loads((out/'audit.json').read_text())
 rows=[]
 for case in cases:
  for side in ['parent','head']:
   a=audit[case+'-'+side];relation=a['relation_overflow']
   rows.append([case,side,a['files'],a['units'],quote(relation['after_load']),quote(relation['end'])])
 table(['Run','Side','Files','Unit records','Relation overflow after load','Relation overflow at end'],rows)
 lines+=['Empty quoted strings mean no checker relation overflow was observed at those endpoints. The meter used its default endpoint attribution, not per-unit relation probing. The source manifest was rehashed after measurement; all 81 files still match the one adapted input. Pairing checked identical headers, unit identities and eligibility. See audit.json for all-phase new-site checks.']
else:lines+=['Final coverage audit is pending the complete entry streams.']
heading('## Excluded entry diagnostics and measurement correction')
lines+=['The initial `tsc-first-*` invocation recorded only the root file and is excluded. The next `tsc-entry-first-*` expansion observed 81 files but put corePublic.ts first. Production `source_identity.go` seeds import-declaration routes from `program.Files()[0]`; this caused 2,406 parent measurement panics, including Version, with reason "lower: source module has no import-declaration route: /workspace/never-call-adapted/src/compiler/semver.ts" for Version. The matching head diagnostic was stopped with SIGTERM once the cause was identified. Neither stream is used in the compiler comparison.','',
'The final `tsc-entry-ordered-first-*` overlay preserves the single original entry both in LatentRoots and as Files()[0], appending each remaining resolved source once. It does not replace the root with 81 independent roots, suppress panics or change compiler proof logic. Both no-output guards remain active. The driver asserts entry-first order before measurement, and the audit checks the first file, complete closure, unit count and unchanged source hashes. The hook and driver are retained verbatim as `.go.txt` evidence.','',
'One ordered-head launch was attempted before its executable existed. No compiler child started; wrapper exit was 1, with no compiler exit code or memory measurement. It is a launch diagnostic, not a census panic. The valid head measurement started after both builds completed. The original ordered-head checkpoint retains this launch failure; the head-complete checkpoint contains the valid completed run.']
rows=[]
for case in ['tsc-first','tsc-entry-first']:
 for side in ['parent','head']:
  p=out/(case+'-'+side+'.timing.json')
  if p.exists():
   t=json.loads(p.read_text());rows.append([case,side,f"{t['seconds']:.3f}",t['peak_rss_kib'],t['child_exit_code'],'Excluded'])
table(['Diagnostic run','Side','Elapsed seconds','Peak RSS KiB','Compiler child exit','Use'],rows)
heading('## Evidence and checkpoints')
lines+=['`provenance.json`, setup/apply/build logs, binary hashes, overlay hashes, `.go.txt` entry drivers, `closure.json`, `inventory.json`, run wrappers, unit timing streams and exact pairing outputs are retained beside this report. The `*-paired.json` files are the complete comparisons. Streams over 20,000,000 bytes are gzip-compressed with a fixed gzip timestamp. `SHA256SUMS` records the SHA-256 of every retained artifact; `stream-hashes.json` additionally records compressed and decompressed stream identities.','',
'Checkpoints were pushed to separate `typescript/never-call-census-checkpoint-*` refs after apply and after each run. This keeps the final `typescript/never-call-census` branch as one commit directly on the parent without force pushing or rewriting history. The two `tsc-first-*` checkpoints contain the excluded root-only diagnostic observations; the `tsc-entry-ordered-first-*` checkpoints contain the corrected complete entry census.','',
'The preliminary root-only runs took '+ ' / '.join(f"{json.loads((out/('tsc-first-'+s+'.timing.json')).read_text())['seconds']:.3f}s ({json.loads((out/('tsc-first-'+s+'.timing.json')).read_text())['peak_rss_kib']:,} KiB peak RSS, exit 0)" for s in ['parent','head'])+' for parent / head. They are not used in any census delta.']
text='\n'.join(lines)+'\n'
assert '\u2013' not in text and '\u2014' not in text
(out/'REPORT.md').write_text(text)
print('render child exit code: 0; entry complete:',entry_ready)
