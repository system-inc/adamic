#!/usr/bin/env python3
"""Render the verified metrics and every exact stop reason into REPORT.md."""
import collections,csv,gzip,json,pathlib
R=pathlib.Path(__file__).resolve().parent;E=R/'evidence'
a=json.loads((E/'audit.json').read_text())
names=['checkOverloadResults','forwarderValues','finishClassCalls','finishAccessors','finishToPrimitive','checkIterableViews','exceptions','checkLibrarySpreads','checkAccessorSpreads','checkRecordGuardedReads','checkDynamicKeyedGetters','checkGeneratorRoads','NormalizeAsync','asyncEntrySource','checkBoundArguments','checkViewOptionalWrites','checkViewCopies','checkViewWidenedReads','checkHostMemberWrites','asyncPayloadFailures','checkPromiseConsumption','checkViewLogicalWrites','readiness','borrow','counters','voidDiscards']
def rows(p):
 if not p.exists():return []
 with gzip.open(p,'rt') as f:return [json.loads(x) for x in f if x.strip()]
calls=[];units={};findings=[];statuses=[]
for d in sorted((E/'main'/'prototype').iterdir()):
 if not d.is_dir():continue
 statuses.append(json.loads((d/'status.json').read_text()))
 calls+=rows(d/'passes.jsonl.gz')
 for f in rows(d/'census.jsonl.gz'):
  for u in f.get('units',[]):units[u['where']]=u
  findings+=f.get('findings',[])
byfinding={(f['unit'],f['phase']):f for f in findings}
metrics=[];stops=[];panics=[]
for name in names:
 ps=[p for p in calls if p['pass']==name]
 errors=[p for p in ps if 'error'in p];crashes=[p for p in ps if 'panic'in p]
 metrics.append(dict(pass_name=name,calls=len(ps),wall_seconds=sum(p['wall_seconds'] for p in ps),observed_process_hwm_kib=max((p['maxrss_kib'] for p in ps),default=None),sum_hwm_growth_kib=sum(p['hwm_growth_kib'] for p in ps),new_stopped_units=len(errors),panics=len(crashes),isolated_incremental_peak_memory=None))
 for p in errors:
  f=byfinding[(p['unit'],'whole_pass:'+name)];u=units[p['unit']]
  stops.append(dict(unit=p['unit'],name=u.get('name',''),unit_kind=u['kind'],pass_name=name,error_type=p['type'],diagnostic_kind=f['kind'],where=f['where'],reason=f['reason'],text=f['text'],raw_error=p['error']))
 for p in crashes:panics.append(p)
(E/'pass-metrics.json').write_text(json.dumps(metrics,indent=2)+'\n')
(E/'new-stops.json').write_text(json.dumps(stops,indent=2)+'\n')
(E/'panics.json').write_text(json.dumps(panics,indent=2)+'\n')
with (E/'pass-metrics.csv').open('w') as f:
 writer=csv.DictWriter(f,fieldnames=list(metrics[0]),lineterminator="\n");writer.writeheader();writer.writerows(metrics)
mi=lambda x:f'{x/1024:.3f}'
lines=['## Measurements at the audit tip','',f"All {a['units']:,} ordinal units in the 81-source checker-clean tsc entry were covered by disjoint ranges `[0,2000)`, `[2000,3500)`, `[3500,4566)`. There are {len(units):,} unit records per side. All final children exited 0 before the deadline; every stream has all 81 file rows and a relation trailer, with no after-load/end relation overflow or flagged cache pairs. Both sides omit the independent whole-file refusal scan as explained above.",'',f"Unit statuses: {dict(collections.Counter(u['status'] for u in units.values()))}. {a['tail_units']:,} units reached the first tail operation. Those include empty/import/type declaration products; the root-kind counts are in `evidence/audit.json`. Successful function-declaration roots reaching the tail: {a['tail_unit_kinds'].get('KindFunctionDeclaration',0)}. A large invocation count is not evidence that the corresponding graph contained a complete program.",'',f"The probe newly stops **{a['new_stopped_units']} units**, by the raw first post-body error; the paired stock body-only observation had no earlier lowering/deferred stop at those units. These are scoped diagnostic observations, **not certified production regressions**. Final-prototype pass panics: **{a['panics']}**.",'','| Pass | Calls | Added timed wall (s) | Observed process HWM (MiB) | HWM growth during calls (MiB) | Newly stopped units | Panics |','|---|---:|---:|---:|---:|---:|---:|']
for m in metrics:
 lines.append(f"| {m['pass_name']} | {m['calls']} | {m['wall_seconds']:.6f} | {mi(m['observed_process_hwm_kib']) if m['observed_process_hwm_kib'] is not None else 'N/M'} | {mi(m['sum_hwm_growth_kib'])} | {m['new_stopped_units']} | {m['panics']} |")
for name in ['findCycles and delegated checks','Program-region gate','plan application','ownership query']:
 lines.append(f'| {name} | not run | N/M | N/M | N/M | unknown | not exercised |')
lines += [f"| **Timed scoped total** | **{a['pass_calls']}** | **{a['pass_wall_seconds']:.6f}** | **{mi(a['prototype_peak_rss_kib'])} overall child peak** | non-additive across children | **{a['new_stopped_units']}** | **{a['panics']}** |",'',f"Stock total unit wall: **{a['stock_unit_wall_seconds']:.6f} s**; prototype: **{a['prototype_unit_wall_seconds']:.6f} s**. Stock overall child peak: **{mi(a['stock_peak_rss_kib'])} MiB**; prototype: **{mi(a['prototype_peak_rss_kib'])} MiB**. Neither their wall difference nor their peak difference is an isolated pass cost: paired children overlap, query caches warm differently, and there is only one observation per range. The direct pass-timer sum excludes logging and includes some semantically unready graph passes. Isolated incremental peak memory for every pass and in total remains **not measured**; do not read zero HWM growth as zero allocation.",'','| Range | Stock child wall (s) | Prototype child wall (s) |','|---|---:|---:|']
for d in sorted((E/'main'/'prototype').iterdir()):
 if not d.is_dir():continue
 p=json.loads((d/'status.json').read_text());s=json.loads((E/'main'/'stock'/d.name/'status.json').read_text())
 lines.append(f"| [{p['start']},{p['end']}) | {s['wall_seconds']:.6f} | {p['wall_seconds']:.6f} |")
lines+=['','### Exact added stop reasons','', 'Every unit, function name, precise failure location, original error type and full unmodified error text is retained in [`evidence/new-stops.json`](evidence/new-stops.json). The following groups use the exact diagnostic reason string; locations are not deduplicated out of the JSON ledger.','', '| Pass | Units | Exact reason |','|---|---:|---|']
counts=collections.Counter((p['pass_name'],p['diagnostic_kind'],p['reason']) for p in stops)
for (p,k,reason),n in sorted(counts.items(),key=lambda x:(names.index(x[0][0]),x[0][2])):
 reason=reason.replace('|','\\|').replace('\n',' ')
 lines.append(f'| {p} | {n} | {k}: {reason} |')
lines+=['','### Panics and per-unit misbehavior','', 'The final main experiment is summarized above; no initial pilot is silently included in it. The small fixture additionally exposes:', '', '* **checkPromiseConsumption:** the pre-entry-fix pilot panics at probe.ts:1:1 with `runtime error: invalid memory address or nil pointer dereference`. It unconditionally reads `l.entry.AsNode()`, even with no async work. Initializing entry removes it. Both raw versions are retained under `evidence/fixture/`.', '* **checkGeneratorRoads:** the generator-call fixture unit returns `lower: compiler bug: 1 generator values where only 0 are a consuming loop\'s or a dropped call\'s`. The unit has signature-only producer/consumer metadata. This is a census-scoping invariant error, not a demonstrated compiler bug in the complete source program.', '* **checkIterableViews:** the separately lowered consumer reports `a view over an iterable of unproven origin (local or result has no tracked producer; no reaching allocation established)` because its caller is absent. This is an explicit example of the incomplete-graph limitation.', '* **checkBoundArguments / later CFG checks:** calling them without findCycles\' resolveCountTypes leaves function-type targets and forwarded argument-count facts incomplete. A clean return is not evidence that the original pass question was answered. Their recommendations remain the inventory\'s omission/whole-program categories.', '', 'The scan-enabled main pilots were author-stopped during remaining refusal scans, with child SIGTERM exits (-15), not pass panics or deadline expirations. They have no relation trailer and are excluded from all totals above. Wrapper checks separately preserve child exit 7, SIGTERM -15 (shell 143), and forced SIGKILL -9 (shell 137); see the three `evidence/wrapper-*.json` records.', '', '### Reproduction and evidence checks', '', 'Build the stock guarded overlay with the pinned tree\'s `make_overlay.py`, extend it with `stage3/meter/entry_overlay.py`, then apply `prototype.py ENTRY_OVERLAY OUTPUT_OVERLAY` (or add `--stock`). Build only the guarded census tool using that overlay. Use each revision\'s own completed apply product and the lockfile-pinned API parser. Each preparation/build was separately bounded at 1,700 seconds here; the initial missing-checker-module and missing-upstream-input attempts failed and contributed no measurements.', '', 'Run each census chunk as its own invocation of `run-chunk.py BIN ENTRY DEST START END MODE`, with `LWP_SKIP_SCAN=1`, ranges shown above, and mode `prototype` or `stock`. The pin uses parent range `[3557,3608)`. The wrapper defaults to a 1,700-second deadline and can only be shortened through LWP_DEADLINE_SECONDS. Do not merge the paired scan-enabled pilots into the final ranges or sum repeated refusal scans.', '', '`collect.py SCRATCH` freezes completed streams with deterministic gzip and records input/binary hashes. `python3 audit.py` checks disjoint exhaustive coverage, successful child exits, clean relation trailers, matching baseline unit identities, exact first-error pass prefixes, no prior baseline body stop at a tail unit, and the negative pin. `measurements.py` renders this section from the audited raw data. Timings/costs are in `evidence/pass-metrics.json` and `.csv`; exact stops are in `evidence/new-stops.json`; immutable pins, input hashes, per-ref overlay equivalence and the box are alongside them.']
p=R/'REPORT.md';text=p.read_text();marker='\n## Measurements at the audit tip\n';text=text.split(marker)[0].rstrip()+'\n\n'+'\n'.join(lines)+'\n'
lead=f"<!-- measured-summary:start -->\n**Measured result:** {a['units']:,} top-level entry units; **{a['new_stopped_units']} newly stopped units** in the scoped tail probe; **{a['pass_wall_seconds']:.3f} seconds** of directly timed added pass work. This unit-only experiment omits independent whole-file refusal scans, and graph-dependent results are partial.\n\n**Pin gate failed:** the parent and 87028b08 both stop createTypeChecker on ParameterPropertyDeclaration before reaching the generator code or tail. The required five-site positive witness is not reproduced. The plan remains gated on a reachable witness and explicit pass coverage.\n<!-- measured-summary:end -->"
if '<!-- measured-summary:start -->' in text:
 before,rest=text.split('<!-- measured-summary:start -->',1);_,after=rest.split('<!-- measured-summary:end -->',1);text=before+lead+after
else:
 title,rest=text.split('\n',1);text=title+'\n\n'+lead+'\n'+rest
p.write_text(text)
print(json.dumps({'new_stops':len(stops),'pass_wall_seconds':a['pass_wall_seconds'],'metric_rows':len(metrics)},indent=2))
