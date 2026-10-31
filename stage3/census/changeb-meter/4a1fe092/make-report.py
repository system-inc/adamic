#!/usr/bin/env python3
import gzip, hashlib, json
from pathlib import Path
p=Path(__file__).resolve().parent
def get(name):return json.loads((p/name).read_text())
def records(name):
 f=p/name
 data=f.read_text() if f.exists() else gzip.decompress(Path(str(f)+'.gz').read_bytes()).decode()
 return [json.loads(l) for l in data.splitlines()]
v=get('validation.json'); t=get('timing-summary.json'); rows=get('timing-table.json'); reasons=get('reason-changes.json'); regressions=get('lowering-regressions.json')
a=get('A.resources.json');b=get('B.resources.json');box=get('box.json')
left,right=records('A.jsonl'),records('B.jsonl')
def table_top():
 lines=['# Twenty largest B−A unit times','','Unit wall seconds include registration, lowering and panic recovery. Ratio = B/A. Sorted by signed B−A.','','| Unit | Name | A seconds | B seconds | B−A seconds | B/A |','|---|---|---:|---:|---:|---:|']
 for r in rows[:20]:lines.append(f"| {r['where']} | {r['name']} | {r['A_seconds']:.6f} | {r['B_seconds']:.6f} | {r['delta_seconds']:+.6f} | {r['ratio']:.4f} |")
 (p/'timing-top20.md').write_text('\n'.join(lines)+'\n')
table_top()
def examples(rs,kind,reason):return sorted({f['where'].replace('/workspace/scratch/tree/','') for r in rs[1:] for f in r.get('findings') or [] if f['kind']==kind and f['reason']==reason})[:5]
lines=['# View, branded-supertype and predicate reason rows','','Every matching reason is named here, including unchanged rows. Examples are up to five distinct locations on each side.','','| Kind and exact reason | A | B | A examples | B examples |','|---|---:|---:|---|---|']
for k in sorted(v['A']['per_reason'].keys()|v['B']['per_reason'].keys()):
 if not any(w in k.lower() for w in ('view','predicate','branded supertype')):continue
 kind,reason=k.split(': ',1)
 lines.append('| '+k.replace('|','&#124;')+' | '+str(v['A']['per_reason'].get(k,0))+' | '+str(v['B']['per_reason'].get(k,0))+' | '+'<br>'.join(examples(left,kind,reason))+' | '+'<br>'.join(examples(right,kind,reason))+' |')
(p/'view-predicate-rows.md').write_text('\n'.join(lines)+'\n')
lines=['# A units with no lowering stop that are absent or stopped under B','','| Unit | Name | B status | B lowering stops |','|---|---|---|---|']
for r in regressions:
 u=r['B']; stops='; '.join(f["kind"]+': '+f['reason']+' at '+f['where'].replace('/workspace/scratch/tree/','') for f in u['lowering_stops']) if u else 'absent'
 lines.append('| '+r['where']+' | '+r['A'].get('name','')+' | '+(u['status'] if u else 'absent')+' | '+stops.replace('|','&#124;')+' |')
if not regressions:lines+=['','None.']
lines+=['','An A success here means status attempted with no lowering-phase NotYet, Refused, error, panic or SkippedDependency. File refusal-scan findings remain separate; this does not claim native execution or every expression was reached. Complete transition records are in lowering-regressions.json.']
(p/'lowering-regressions.md').write_text('\n'.join(lines)+'\n')
registration={'distinct_branded_supertypes_with_registered_views':None,'total_subtype_views':None,'available':False,'reason':'The first-error ledger records declarations, units and findings, not successful view registrations. Logs contain per-file unit/finding counts, not registration events. A type named by a refusal does not establish successful registration. No extra internal instrumentation was added.'}
(p/'registration-counts.json').write_text(json.dumps(registration,indent=2)+'\n')
# Retain the context of every new distinct stop, including already-stopped A units.
ua={u['where']:u for r in left[1:] for u in r.get('units',[])}
old={tuple(f.get(k,'') for k in ('kind','where','reason','text')) for r in left[1:] for f in r.get('findings') or []}
contexts=[]
for r in right[1:]:
 for f in r.get('findings') or []:
  if f.get('phase')!='lowering' or f['kind'] not in ('NotYet','Refused','error','panic'):continue
  if tuple(f.get(k,'') for k in ('kind','where','reason','text')) in old:continue
  contexts.append(dict(B=f,A_unit=ua.get(f['unit']),A_lowering_findings=[x for ar in left[1:] for x in ar.get('findings') or [] if x.get('unit')==f['unit'] and x.get('phase')=='lowering']))
(p/'new-stop-context.json').write_text(json.dumps(contexts,indent=2)+'\n')
countlines='\n'.join(f"| {k} | {v['A']['totals'][k]:,} | {v['B']['totals'][k]:,} |" for k in ('NotYet','Refused','SkippedDependency','error','panic'))
controls=t['controls']
callouts='\n'.join(f"- `{r['kind']}: {r['reason']}`: {r['A']} → {r['B']}." for r in reasons if r['callout']) or 'None.'
(p/'REPORT.md').write_text(f'''# New B 4a1fe092 against A: paired latent census

Both censuses completed **{v['A']['source_files']} / {v['B']['source_files']} files** (A / B), exit **{a['exit_code']} / {b['exit_code']}**. Both passed the existing meter’s `report.py.lowering_summary` coverage and distinct-site validator. This is the checker-rejected, first-error measurement; no usable IR or native output is claimed.

| Distinct (kind, where, reason, text) sites | A | B |
|---|---:|---:|
{countlines}

There are **{len(reasons)} changed kind/reason rows** and **{len(regressions)} A-successful units absent or stopped under B**. [Every changed reason, counts and up to five removed/added examples](reason-changes.md), the requested [census delta](census-delta.json), [every unit regression](lowering-regressions.md), and [new-stop context](new-stop-context.json) are retained. The context also covers B stops in units that already stopped under A.

Changed rows mentioning a view, branded supertype or predicate:

{callouts}

The new stop is in `getAllAccessorDeclarations` at `utilities.ts:6838:9`. A already stopped in that unit at `sourcemap.ts:474:9`, `a getter in a checked field contract`, so this is a changed first stop, not a regression from a successfully lowered A unit. No new read-through-view or write-through-view reason was recorded. The rerun A ledger is byte-for-byte identical to the parent run A ledger (SHA-256 `2c2bd81803ebe5e57a81eba37b39f1784850af0fe29eb987da25f6c79d4fcee8`).

[All named view/predicate rows](view-predicate-rows.md), including unchanged rows, carry A/B counts and examples. **Distinct branded supertypes registered and total subtype views are unavailable from these ledgers/logs**, not zero: successful registrations are not recorded. See [registration-counts.json](registration-counts.json). No compiler instrumentation was added to obtain them.

| Timing/resources | A | B | B−A / B:A |
|---|---:|---:|---|
| Sum of unit wall seconds | {t['A_total_seconds']:.6f} | {t['B_total_seconds']:.6f} | {t['B_total_seconds']-t['A_total_seconds']:+.6f} s / {t['B_total_seconds']/t['A_total_seconds']:.6f}× |
| Sum of unit CPU seconds | {t['A_cpu_seconds']:.6f} | {t['B_cpu_seconds']:.6f} | {t['B_cpu_seconds']-t['A_cpu_seconds']:+.6f} s |
| Whole-run wall seconds | {a['wall_seconds']:.6f} | {b['wall_seconds']:.6f} | {b['wall_seconds']-a['wall_seconds']:+.6f} s |
| Peak RSS, KiB | {a['peak_rss_kib']:,} | {b['peak_rss_kib']:,} | {b['peak_rss_kib']-a['peak_rss_kib']:+,} KiB |

[Top twenty by B−A](timing-top20.md) and [all {t['units']:,} paired timed units](timing-table.md) give seconds and ratios; JSON preserves CPU time too. Empirical noise controls: **{controls['units']:,}** imports, interfaces and type aliases with identical findings, no predicate body in their own statements: median absolute delta **{controls['median_absolute_delta']*1000:.3f} ms**, p95 **{controls['p95_absolute_delta']*1000:.3f} ms**, median B/A **{controls['median_ratio']:.4f}×**. These include common registration overhead; scheduling and GC create outliers. One paired run does not isolate per-view cost or guarantee that a later expression was reached after a first stop.

Same box `{box['hostname']}`, nproc {box['nproc']}, CPU quota {box['cpu_quota']}, memory {box['total_memory']}. A was rerun beside new B, with simultaneous starts and separate `LATENT_TIMES`, `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, and 5,400-second limits. The shared tree hashes matched the previous pair before this run and matched again afterward.

A: `e0391d48aab3227d3aa5f4c93da39e3fbb2513be`; B: `4a1fe09255a6442aeb7befde57a6f14c5899e423`; unchanged adaptation main: `4e45d80b59bdd03c054d7611c50e6b89cace1a0b`; upstream: `050880ce59e30b356b686bd3144efe24f875ebc8`. The same meter-only recorder changes from `1bab347531e8c238d97940ea843ad3c166509e18` were applied to new B. Meter sources match A byte-for-byte; no `internal/` changes. [Exact commands](COMMANDS.md), [binary SHA-256s](binaries.sha256), gzip ledgers/logs, resource records and validated summaries are in this directory. Parent evidence is preserved unchanged.
''')
