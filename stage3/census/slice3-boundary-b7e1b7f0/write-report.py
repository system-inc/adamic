import json
from pathlib import Path
D=Path(__file__).resolve().parent
s=json.loads((D/'summary.json').read_text());new=json.loads((D/'new-stopped-units.json').read_text());pro=json.loads((D/'provenance.json').read_text());positions=json.loads((D/'position-verification.json').read_text())
lines=['# Every newly view-stopped unit','']
for e in new:
 h=e['head'];lines += [f"## {e['id']}. {e['unit']}",'',f"Native view site: `{h['where']}`.",'',f"Base finding-free regression: **{'yes' if e['base_no_other_finding'] else 'no'}**. Cause: {e['cause']}.",'','Exact head reason:','',h['reason'],'','Source expression or native predicate site:','','```typescript',h['text'],'```','']
 base=e['base_ordinary_findings'];lines += ['Base ordinary findings: '+ ('none.' if not base else ''),'']
 for f in base:lines += [f"- `{f['where']}`: {f['reason']}"]
 b=e['base_boundary'];lines += ['', 'Base modeled boundary: '+(b.get('reason','clear')),'','Counterfactuals:','']
 for mode,r in h['counterfactuals'].items():lines += [f"- {mode}: "+(f"`{r.get('where','')}`: {r.get('reason','stop')}" if r['stop'] else 'clear')]
 lines+=['']
if not new:lines+=['None. The head has no newly view-stopped units in this census. There are no new expressions or reasons to quote.','']
(D/'NEW-STOPS.md').write_text('\n'.join(lines))
runrows=[];longest=0;bad=[]
for p in sorted((D/'raw').rglob('*.json')):
 try:r=json.loads(p.read_text())
 except json.JSONDecodeError:continue
 if not isinstance(r,dict) or 'child_exit' not in r:continue
 t=r['elapsed_seconds'];longest=max(longest,t);assert t<1800,(p,t)
 runrows.append(f"| {p.relative_to(D/'raw')} | {r['child_exit']} | {t:.3f} | {r['peak_rss_kib']} | {'yes' if r.get('timed_out') else 'no'} |")
(D/'RUNS.md').write_text('# Child exits, time and peak memory\n\nRows marked attempt-1 are excluded from the final counts. The initial worktree command outside the repository failed and was retried successfully. Controls and setup remain separate from census children.\n\n| Metadata under raw/ | Child exit | Elapsed seconds | Peak RSS KiB | Timed out |\n| --- | ---: | ---: | ---: | --- |\n'+'\n'.join(runrows)+'\n')
a=s['base'];b=s['head'];n=s['new_stops'];reg=s['base_finding_free_regressions'];cause_rows='\n'.join(f"| {cause} | {v['unit_observations']} | {v['raw_findings']} | {v['distinct_reads']} |" for cause,v in s['causes'].items())
def count(cat,side,field):return s['view_pair_counts'][cat][field]
text=f'''# Slice 3 graph-adoption boundary census

Task ef727kr. Base: `{pro['base']}`. Head: scratch merge of `{pro['branch_tip']}` onto that base, tree `{pro['scratch_merge_tree']}`. The merge has no conflicts and touches no adaptation. Both compilers use one tree produced by the base apply.sh from pinned TypeScript 6.0.3. This report covers **conditional unit-local boundary stops**, not whole-program graph membership or acceptance.

## 1. Newly stopped units and regressions

**{n['unit_observations']} units newly stop at a boundary view refusal**, representing {n['raw_findings']} raw findings and {n['distinct_reads']} distinct reads. **{reg['unit_observations']} had no other finding at the base**, representing {reg['raw_findings']} raw findings and {reg['distinct_reads']} distinct reads. Those are the unit-local regressions.

[NEW-STOPS.md](NEW-STOPS.md) quotes every new unit, native location, exact reason and source expression or predicate site, and marks whether the base had another finding. new-stopped-units.json retains ordinary and boundary baseline findings and all counterfactual reasons. No incomplete or discarded attempt contributes a count.

## 2. Causes

Each head view refusal is rerun on fresh snapshots with the deeper fields/index/Map/Set walk disabled, unconstrained-parameter refusal disabled, and both disabled. No head view refusal occurred, so no tsc counterfactual rerun was needed. The counterfactual harness tests the real boundary, including another view that may become its first stop. Unit observations and deduplicated raw/distinct findings are separate.

| Cause of new unit view stop | Unit observations | Raw findings | Distinct reads |
| --- | ---: | ---: | ---: |
{cause_rows}

The joint categories are explicit rather than silently charged to one widening. A stop that survives both disabled is reported separately. The [method](METHOD.md) defines these counterfactuals and the normalization.

## 3. Cleared stops, moved stops and panics

At the unit level, {s['cleared_view_stopped_units']} base view-stopped units have no head view refusal; {s['moved_view_first_stop_units']} remain view-stopped with a changed first location or reason. A cleared view refusal may expose a different boundary category. Across all modeled boundary first stops, {s['all_boundary_unit_changes']['cleared']} clear, {s['all_boundary_unit_changes']['moved']} move, and {s['all_boundary_unit_changes']['new']} are new. Every row is retained in cleared-units.json, moved-units.json, and all-boundary-*-units.json.

The pair tool reports raw/distinct cleared view findings **{count('raw_findings','', 'lowered')}/{count('distinct_reads','', 'lowered')}**, moved view findings **{count('raw_findings','', 'moved')}/{count('distinct_reads','', 'moved')}**, new view findings **{count('raw_findings','', 'new_findings')}/{count('distinct_reads','', 'new_findings')}**, and units whose view first stop changes **{count('raw_findings','', 'unit_first_stop_delta')}/{count('distinct_reads','', 'unit_first_stop_delta')}**. New findings at an already blocked unit are not counted as a newly stopped unit in section 1. [view-delta.md](view-delta.md) and view-delta.json preserve the site-level pair.

New ordinary panics: **{len(s['new_panics'])}**. Probe errors in included selections: **0**. ordinary-delta.json retains ordinary first-stop changes and panic details. Every serialized ordinary header/file/unit/finding byte matches across every paired selection, excluding relation-report telemetry (hashes in ordinary-byte-verification.json). All combined ordinary file/unit/finding records equal across the paired runs: **{str(s['all_ordinary_file_unit_findings_equal']).lower()}**.

## 4. Totals

| Measure | Base | Merged head |
| --- | ---: | ---: |
| Observed unit results | {a['observed_units']} | {b['observed_units']} |
| Recorded view sources | {a['views_recorded']} | {b['views_recorded']} |
| Units with a boundary view stop | {a['view_stopped_units']} | {b['view_stopped_units']} |
| Raw view-stop findings at reported positions | {a['raw_view_findings']} | {b['raw_view_findings']} |
| Distinct view-stop reads at reported positions | {a['distinct_view_reads']} | {b['distinct_view_reads']} |
| Units with another modeled boundary stop | {a['other_boundary_stopped_units']} | {b['other_boundary_stopped_units']} |
| Units with any modeled boundary stop | {a['boundary_stopped_units']} | {b['boundary_stopped_units']} |
| Ordinary finding observations | {a['ordinary_finding_observations']} | {b['ordinary_finding_observations']} |
| Ordinary panics | {a['ordinary_panics']} | {b['ordinary_panics']} |
| Native viewHosted calls | {a['native_host_walk_calls']} | {b['native_host_walk_calls']} |

The base has {positions['base']['view_stops_with_unit_fallback']} view-stopped unit observations with an explicitly marked unit fallback because its legacy diagnostic site is absent from the unit snapshot. Base raw/distinct totals count reported native or fallback positions; they are not exact expression identities for those observations. The head has no view-stop fallback. position-verification.json retains the precision counts. The new-stop counts in sections 1 and 2 are zero and need no fallback attribution.

Ordinary unit inventory by status: base `{json.dumps(a['ordinary_unit_status'],sort_keys=True)}`; head `{json.dumps(b['ordinary_unit_status'],sort_keys=True)}`. Every selected file and disjoint unit range completed on both sides. The entry graph has 81 files; source-manifest.json hashes all 747 adapted source files.

## Reachability proof and limits

The stock meter never calls findCycles and therefore skips graphAdoptionBoundary. The scratch-only unit-end observer calls the real boundary on a fresh finder and the unit's snapshot, without selecting graph members. Native entry counters record **{b['native_host_walk_calls']} calls to the changed viewHosted** in the included tsc runs. Before measuring tsc, the native host-walk test passed and the identity control made 14 native walk calls. Forced-refusal controls produced a stop on both sides; real/mutant fixture ordinary output matches runs with observation disabled. final-control-results.json and raw/host-walk-tests.* retain the proof.

The pinned base has the blanket checked-view adapter boundary. The merged head replaces it with host-sensitive checks and introduces viewHosted. This pair is broader than a shallow/deep-only comparison, which is why already blocked units and moved first reasons are kept separate.

The first attempted full-boundary seam encountered a base diagnostic-only Where(nil) panic on a dynamic-shape stop with no graphDynamicSite in the unit snapshot. That attempt was stopped and excluded. A stack-traced isolated reproducer identifies the exact call. The final scratch formatter handles only missing diagnostic locations, preserves the real refusal and reason, and marks a unit fallback. The repaired unit reports its dynamic-shape refusal without a probe panic. Native head view-stop locations and source text are unchanged. See [METHOD.md](METHOD.md) for the exact scope and interpretation.

The census uses first-error top-level independent units. Callee bodies reached by them are included; standalone generics without a call witness are deferred. Views before an ordinary stop remain latent observations in already blocked units. This does not model whole-program membership, earlier whole-program cycle proofs, or unselected graph-pair membership. It does not establish native tsc acceptance.

[RUNS.md](RUNS.md) lists each child's own exit code, elapsed time and peak RSS. Every completed process is below 30 minutes; the longest recorded child took {longest:.3f} seconds. Census children have a 1490-second cap. Raw streams over 1 MB are gzipped. Production Go evidence uses .go.txt names. The measurement seams remain outside the commit; provenance.json records their hashes. The commit changes only the census directory; it contains no production or adaptation edits.
'''
assert '\u2013' not in text and '\u2014' not in text
for compressed in D.glob('*.gz'):
 text=text.replace(compressed.name.removesuffix('.gz'),compressed.name)
(D/'REPORT.md').write_text(text)
print('Rendered report with',len(new),'newly view-stopped units')
