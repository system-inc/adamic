import collections,json
from pathlib import Path
out=Path(__file__).resolve().parent
read=lambda n:json.loads((out/n).read_text())
p=read('provenance.json');groups=read('grouped-first-stops.json');art=read('root-order.json')
checks={r['root']:r for r in art['checks']};excerpts=read('source-excerpts.json');emitters=read('emitters.json')
rows=[json.loads(s) for s in (out/'census.jsonl').read_text().splitlines()]
entry=[json.loads(s) for s in (out/'entry-latent.jsonl').read_text().splitlines()]
counts=collections.Counter(f['kind'] for r in entry for f in r.get('findings') or [])
ordinary=collections.Counter(r['kind'] for r in rows)
totals={'guarded_entry':{'NotYet':counts['NotYet'],'Refused':counts['Refused'],'panics':counts['panic'],'complete':art['status']=='complete','completed_sources':art['completed_sources'],'expected_sources':art['expected_sources'],'child_exit':art['entry_status']['exit_code'],'seconds':art['entry_status']['seconds']},'ordinary_counted_first_stops':dict(ordinary),'ordinary_latent_totals':None}
(out/'latent-totals.json').write_text(json.dumps(totals,indent=2)+'\n')
lines=['# Next stops on main b322b6ba','','Measured for #baobab and #b5w3ycg. Report only.','',f"Compiler and apply SHA: `{p['sha']}`. TypeScript 6.0.3 pin: `{p['typescript_pin']}`.",'',f"Box: `{p['box']}`; {p['cpus']} visible CPUs; cpu.max `{p['cpu_max']}`; {p['meminfo']}. Started {p['date']}.",'','## Method','','Reuses the method and scripts from `f7fccf67f87b6f9039b714aa46c359e36b231c50`, `typescript/next-stops-c3b7baf0`. The prior report and grouped table are adjacent. Setup sourced `/workspace/adamic-tools/env.sh`, ran `bash cloud/setup.sh` and `npm ci --prefix stage3/api`, and set `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-source` at the pin above. See command/status ledgers for child exit codes.','','One adapted tree from this SHA using `bash stage3/apply.sh /workspace/scratch/next-stops-adapted`. Ordinary binary: `go build -o /workspace/scratch/next-stops-census ./stage3/census/tool`, no overlay. Sorted roots, sequential groups of 20, 20, 20 and 19; each single-root child emits two observations, only the first counts. All duplicate non-timing fields agree. Raw streams retain 158 observations and `census.jsonl` retains 79 counted observations.','','The separate entry binary uses the measured SHA meter overlay, LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, with a 1800 second bound and process-group cleanup. Its stream supplies root-order evidence and observed latent findings. It cannot emit usable IR. `binaries.json`, build info and source manifests identify measured bytes.','','## Grouped first stops','','Root names and locations are relative to `src/compiler/`. Largest groups first.','','| Roots | Location | Kind | Exact reason | Root names |','| ---: | --- | --- | --- | --- |']
for g in groups:
 names=', '.join('`'+r+'`'+(' ('+checks[r]['classification']+')' if r in checks else '') for r in g['roots'])
 reason=g['reason'].replace('|','\\|')
 lines.append(f"| {g['count']} | `{g['where']}` | {g['kind']} | {reason} | {names} |")
accepted=[r for g in groups if g['kind']=='accepted' for r in g['roots']]
lines+=['','## Roots lowering whole','',f"{len(accepted)}/79: "+', '.join('`'+r+'`' for r in accepted)+'.','','This is lowering acceptance of each entire resolved root program under the census loader options; no native execution is measured.','','## Comparison with c3b7baf0','','Unchanged means the root retains the exact (location, kind, reason) tuple. Moved means a changed tuple already present in the old table. New means a tuple absent from the old table. `comparison.json` records every root before and after.','','| Current stop | Roots | Unchanged | Moved to known group | New | Root names |','| --- | ---: | ---: | ---: | ---: | --- |']
old=read('previous-grouped-first-stops.json');baseline={r:(g['where'],g['kind'],g['reason']) for g in old for r in g['roots']}
comparison=read('comparison.json')
for r in comparison:
 r['classification']='unchanged' if r['before']==r['after'] else 'moved' if tuple(r['after']) in baseline.values() else 'new'
(out/'comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
for g in groups:
 sub=[r for r in comparison if r['root'] in g['roots']];cs=collections.Counter(r['classification'] for r in sub)
 lines.append(f"| `{g['where']}` {g['kind']} | {g['count']} | {cs['unchanged']} | {cs['moved']} | {cs['new']} | "+', '.join('`'+r['root']+'` ('+r['classification']+')' for r in sub)+' |')
lines+=['','## Source and lowering emitters','','Source excerpts are included for every stop, including every new group. Emitter excerpts are in `lowering-emitter-excerpts.txt`.']
for e in excerpts:
 lines+=['',"### "+e['where'],'',emitters[e['where']],'',f"Adapted source lines {e['start_line']} through {e['end_line']}:",'','```ts',e['text'].rstrip(),'```']
lines+=['','## Root-order artifact evidence','',f"Entry child exit {art['entry_status']['exit_code']} after {art['entry_status']['seconds']:.3f} seconds; completed source records {art['completed_sources']}/{art['expected_sources']}; stream records {art['entry_stream_records']}.",'']
if art['status']=='incomplete':lines+=['Stopping reason: '+art['stopping_reason'],'','The incomplete ledger cannot establish the attempted-source and readiness criteria in `stage3/meter/report.py`. Readiness roots retain the same explicit incomplete label used by the prior report.']
lines+=['','| Root | Independent stop | Node exit | First error | Label |','| --- | --- | ---: | --- | --- |']
for r in art['checks']:lines.append(f"| `{r['root']}` | `{r['where']}` | {r['node']['exit']} | {r['node']['first_error']} | {r['classification']} |")
lines+=['','## Latent totals','',f"Guarded entry stream observed findings: NotYet **{counts['NotYet']}**, Refused **{counts['Refused']}**, panics **{counts['panic']}**. These count emitted findings, including any relation trailer findings exactly once.",'']
if art['status']=='incomplete':lines+=['These are partial observed stream totals, not totals for the entry program. The run did not finish inside its 1800 second bound. No findings from an unfinished source record are available in this stream. `entry-times.jsonl`, if present, records progress within unfinished work.']
else:lines+=['The entry run completed inside its bound.']
lines+=['',f"Ordinary census, 79 counted first observations: NotYet **{ordinary['NotYet']}**, Refused **{ordinary['Refused']}**, accepted **{ordinary['accepted']}**; no child panics observed. The ordinary stream has no latent finding counters, so ordinary latent totals are unavailable. These first-stop counts are not latent totals.",'','## Checkpoints and cleanup','','Apply and each root group were pushed plainly to `typescript/next-stops-b322b6ba-checkpoint`. The final branch `typescript/next-stops-b322b6ba` has one report-only commit directly on b322b6ba. Main was never pushed. Every census child exit and process-group cleanup is recorded. `validation.json` checks 79 roots, 158 observations, duplicate agreement, source cleanliness, checkpoint pushes and no surviving census processes. Evidence files over 1 MB are gzip-compressed.']
text='\n'.join(lines)+'\n';assert '\u2014' not in text and '\u2013' not in text
(out/'REPORT.md').write_text(text)
