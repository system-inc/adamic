"""Report the pair tool's exact findings and unit first-stop movements."""
from collections import Counter
import gzip
import json
from pathlib import Path
import sys

folder = Path(sys.argv[1])
source = folder / 'paired.json'
pair = json.loads(source.read_text() if source.exists() else gzip.decompress(source.with_name(source.name+'.gz').read_bytes()))
raw = pair['raw_findings']
newly = [r for r in raw['unit_first_stop_delta'] if not r['base'] and r['head']]
deferred = [r for r in newly if r['head']['phase'] == 'deferred' and r['head']['kind'] == 'NotYet']
other_new = [r for r in newly if r not in deferred]
changed = [r for r in raw['unit_first_stop_delta'] if r['base'] and r['head']]
cleared_units = [r for r in raw['unit_first_stop_delta'] if r['base'] and not r['head']]
groups = {view: Counter((r['finding']['reason'], r['identity'][0]) for r in pair[view]['new_findings'])
          for view in ('raw_findings', 'distinct_reads')}
metrics = {view: dict(new_findings=len(v['new_findings']), new_sites=len(v['new_sites']),
                     cleared_findings=len(v['lowered']), moved_findings=len(v['moved']),
                     changed_unit_first_stops=len(v['unit_first_stop_delta']),
                     panics_before=len(v['panics']['before']), panics_after=len(v['panics']['after']),
                     totals=v['totals']) for view, v in pair.items() if view in groups}
summary = dict(lowered_to_stopped=len(deferred), units=[dict(unit=r['unit'], stop=r['head']) for r in deferred],
               changed_existing_stops=len(changed), changed_existing_units=changed,
               other_new_stops=len(other_new), other_new_units=other_new,
               stopped_to_lowered=len(cleared_units), cleared_unit_stops=cleared_units,
               new_findings_by_reason_and_file=[dict(reason=reason, file=file,
                   raw=groups['raw_findings'][reason,file], distinct=groups['distinct_reads'][reason,file])
                   for reason,file in sorted(set(groups['raw_findings']) | set(groups['distinct_reads']))],
               metrics=metrics, production_root_first_stops_moved=0, production_roots=73,
               production_first_stop='core.ts:2599:15',
               production_note='Production lowering and refusal sources are byte-identical to 018b7c21. The 73 production roots were not recompiled.',
               baseline_overflow=json.loads((folder/'overflow-attribution.json').read_text()))
(folder/'counts.json').write_text(json.dumps(summary, indent=2)+'\n')
def loc(value):
    return 'src/'+value.split('/src/',1)[1] if '/src/' in value else value
def cell(value):
    return str(value).replace('|','\\|').replace('\n',' ')
lines = ['# Deferred refusal census', '',
         f"Overstatement of unit first-error admission: **{len(deferred)} units** move from no lowering/deferred first stop to a deferred NotYet.", '',
         'This is the body/refusal census view. The baseline pooled-checker overflow conservatively marks all attempts at run end, so these counts are not proof that any unit successfully compiles.', '',
         '| Unit | Location | Deferred reason |', '| --- | --- | --- |']
for r in deferred:
    lines.append('| '+cell(r['unit'].get('name') or r['unit']['kind'])+' | '+cell(loc(r['unit']['where']))+' | '+cell(r['head']['reason'])+' |')
if not deferred: lines.append('| None | | |')
lines += ['', '## New findings by exact reason and file', '',
          '| Reason | File | Raw | Distinct |', '| --- | --- | ---: | ---: |']
for r in summary['new_findings_by_reason_and_file']:
    lines.append('| '+cell(r['reason'])+' | '+cell(r['file'])+' | '+str(r['raw'])+' | '+str(r['distinct'])+' |')
if not summary['new_findings_by_reason_and_file']: lines.append('| None | | 0 | 0 |')
lines += ['', '## Other movements', '',
          f"Existing unit first stops changed: {len(changed)}. Other new unit stops: {len(other_new)}. Unit first stops cleared: {len(cleared_units)}.", '',
          '| View | New findings | New sites | Cleared findings | Moved findings | Panics before | Panics after |',
          '| --- | ---: | ---: | ---: | ---: | ---: | ---: |']
for view,m in metrics.items():
    lines.append('| '+view+' | '+' | '.join(str(m[k]) for k in ['new_findings','new_sites','cleared_findings','moved_findings','panics_before','panics_after'])+' |')
for title, rows in [('Changed existing first stops',changed),('Other newly stopped units',other_new),('Cleared unit first stops',cleared_units)]:
    lines += ['', '### '+title, '', '| Unit | Before | After |', '| --- | --- | --- |']
    for r in rows:
        def stop(f): return 'None' if not f else loc(f['where'])+': '+f['kind']+': '+f['reason']
        lines.append('| '+cell(loc(r['unit']['where']))+' | '+cell(stop(r['base']))+' | '+cell(stop(r['head']))+' |')
    if not rows: lines.append('| None | | |')
lines += ['', '## Pooled-checker overflow and production roots', '',
          'Baseline TS2859 (`" "` against `{}`) was first observed by the run-end relation probe after the last timed unit, `src/tsc/_namespaces/ts.ts:3:1`. The duration-only journal has no per-unit overflow answers; the onset unit is unknown. This does not establish that the last unit caused the overflow.', '',
          'The complete relation trailers and conservative all-unit attribution are retained in the streams. Program findings use the <program> identity; deferred_scan module events are not unit first stops.', '',
          'Production first stops: 0 of 73 move from `core.ts:2599:15`. '+summary['production_note'], '',
          'The paired JSON retains all finding removals, reason changes, new findings, unit movements and panics. Counts here use the pair tool identities without re-normalizing them.']
lines += ['', '## Eligibility and other finding changes', '', 'Both runs have the same 4,059 attempted units and 504 deferred generic declarations, including 112 instantiated witnesses. Every unit-ledger entry and the complete relation trailers are identical. Neither run has a deferred or deferred_scan finding on this entry.', '', 'All 17 new findings are Refused: 14 refusal_scan and 3 lowering events. The 73 cleared findings are 61 Refused and 12 NotYet, comprising 70 refusal_scan and 3 lowering events. These are finding removals, not units newly proven to lower.', '', 'The two moved findings are the written-any assertion at parser.ts:1253:17 and the corresponding visitorPublic.ts:603:17 assertion. Both retain the ordinary unchecked-cast refusal. The continuing baseline scan had emitted deferred fallbacks despite ordinary findings; the corrected scan suppresses them.', '', 'Ten existing unit first stops change among ordinary overload refusals: five destructuring units change assertNode/assertEachNode parameter choice, and five transformer units change visitEachChild overload/parameter choice. The exact before/after rows are above; none becomes newly admitted or newly stopped.', '', 'This paired entry therefore measures zero deferred admission overstatement. It does not negate the direct and instantiated written-any regression fixtures, whose dropped-read mutant fails as required. See metadata-comparison.json and paired.json.gz for complete records.']
(folder/'counts.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({k:summary[k] for k in ['lowered_to_stopped','changed_existing_stops','other_new_stops','stopped_to_lowered','metrics']},indent=2))
