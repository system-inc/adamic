"""Reproduce the find census using main's access-chain pairing, then render the brief.

Usage: python3 compare.py SCRATCH_DIRECTORY
Expects find-{base,head}-{latent,times}.jsonl, find-adapted and find-core-walls.json.
"""
import collections
import gzip
import hashlib
import importlib.util
import json
from pathlib import Path
import re
import sys
import subprocess

sys.dont_write_bytecode = True
OUT = Path(__file__).resolve().parent
R = Path(sys.argv[1]).resolve()
A = R / 'find-adapted'
spec = importlib.util.spec_from_file_location('census_pair', OUT.parent / 'pair/pair.py')
pair = importlib.util.module_from_spec(spec)
spec.loader.exec_module(pair)

def norm(value):
    if isinstance(value, str):
        return value.replace(str(A) + '/', '')
    if isinstance(value, list):
        return [norm(v) for v in value]
    if isinstance(value, dict):
        return {k: norm(v) for k, v in value.items()}
    return value

streams = {s: norm(pair.load(R / f'find-{s}-latent.jsonl')) for s in ('base', 'head')}
assert streams['base'][0]['latent_mode'] == 'first-error'
delta = pair.compare(streams['base'], streams['head'], A,
                     R / 'find-base/stage3/api/node_modules/typescript')
ids = pair.identities(list(streams.values()), A, R / 'find-base/stage3/api/node_modules/typescript')

context_run = subprocess.run(['node', str(OUT / 'find-context.cjs'), str(A),
    str(R / 'find-base/stage3/api/node_modules/typescript'), str(R / 'find-base-latent.jsonl')],
    check=True, capture_output=True, text=True)
find_context = norm(json.loads(context_run.stdout))
direct_find = {x['where'] for x in find_context if x['direct_find']}

def match(reason, where):
    return bool(re.search(r'overload|predicate', reason) and
                (re.search(r'\bfind\b', reason) or where in direct_find))

def identity(f, distinct=False):
    i = ids[f['where']]
    return (i['file'], i['sha256'], i['read_offset' if distinct else 'byte_offset'])

def key(f, distinct=False):
    return (identity(f, distinct), f['reason']) if distinct else (
        identity(f), f['kind'], f['reason'], f['text'])

def compact(f):
    return {k: f[k] for k in ('where', 'kind', 'reason')} if f else None

finds = {}
all_findings = {s: [f for row in rows[1:] for f in row['findings']]
                for s, rows in streams.items()}
for name, distinct in [('raw_findings', False), ('distinct_reads', True)]:
    b = {key(f, distinct): f for f in all_findings['base']}
    h = {key(f, distinct): f for f in all_findings['head']}
    hs = collections.defaultdict(list)
    for f in h.values():
        hs[identity(f, distinct)].append(f)
    transitions = []
    for k, f in sorted(b.items(), key=lambda v: str(v[0])):
        if match(f['reason'], f['where']):
            after = hs[identity(f, distinct)]
            transitions.append(dict(base=f, head=after, identity=identity(f, distinct),
                                    status='unchanged' if k in h else 'moved' if after else 'gone'))
    summary = {status: sum(t['status'] == status for t in transitions)
               for status in ('gone', 'moved', 'unchanged')}
    summary['new_reasons'] = dict(collections.Counter(
        g['reason'] for t in transitions if t['status'] == 'moved' for g in t['head']))
    finds[name] = dict(summary=summary, transitions=transitions)

# Keep repeated unit observations separate from meter-unique findings.
units = {s: {u['where']: u for row in rows[1:] for u in row['units']}
         for s, rows in streams.items()}
first = {s: {} for s in streams}
for s in streams:
    for f in all_findings[s]:
        if f['phase'] == 'lowering':
            first[s].setdefault(f['unit'], f)
occurrences = []
for f in all_findings['base']:
    if not match(f['reason'], f['where']):
        continue
    if f['phase'] == 'lowering':
        after = first['head'].get(f['unit'])
    else:
        after = next((g for g in all_findings['head']
                      if g['phase'] == f['phase'] and g['unit'] == f['unit']
                      and identity(g, True) == identity(f, True)), None)
    occurrences.append(dict(base=f, head=after,
        status='gone' if after is None else 'unchanged' if compact(f) == compact(after) else 'moved'))

walls = json.loads((R / 'find-core-walls.json').read_text())
wall = next(r for r in walls['rows'] if r['id'] == 'CW003')
assert wall['instance_count'] == 25 and wall['unit_count'] == 25
historical = []
for f in wall['instances']:
    old = first['base'].get(f['unit'])
    new = first['head'].get(f['unit'])
    historical.append(dict(historical=f, base=old, head=new,
                           name=units['base'].get(f['unit'], {}).get('name')))
assert all(r['base'] and compact(r['base']) == compact(r['historical']) for r in historical)
historical_occurrences = [t for t in occurrences if t['base']['phase'] == 'lowering'
                          and t['base']['reason'] == wall['reason']]
assert len(historical_occurrences) == 25
historical_summary = dict(collections.Counter(t['status'] for t in historical_occurrences))
for status in ('gone', 'moved', 'unchanged'): historical_summary.setdefault(status, 0)
instance_reasons = dict(collections.Counter(t['head']['reason'] for t in historical_occurrences if t['status'] == 'moved'))

regressions = list(delta['distinct_reads']['regressions'])
regressions += [dict(type='previously_finding_free_lowering_site', **r)
                for r in delta['distinct_reads']['new_sites']
                if r['finding'].get('phase') != 'refusal_scan' and r['finding']['kind'] != 'panic']
# The task's bar explicitly includes any changed first stop with a new head stop.
for row in delta['raw_findings']['unit_first_stop_delta']:
    if row['base'] and row['head']:
        regressions.append(dict(type='changed_first_lowering_stop', **row))
timing = {}
for s in streams:
    rows = norm(pair.load(R / f'find-{s}-times.jsonl'))
    attempted = {w for w,u in units[s].items() if u['status'] in ('attempted', 'panic')}
    assert len(rows) == len(attempted) and {r['where'] for r in rows} == attempted
    timing[s] = dict(attempts=len(rows), wall_seconds=sum(r['wall_seconds'] for r in rows),
                     cpu_seconds=sum(r['cpu_seconds'] for r in rows))
resources = {s: json.loads((R / f'find-{s}-resource.json').read_text()) for s in streams}
assert all(v['exit_code'] == 0 and v['signal'] is None for v in resources.values())
assert set(units['base']) == set(units['head'])
for s in streams:
    assert not (R / f'find-{s}-status.txt').read_text()
    assert (R / f'find-{s}-memory-before.txt').read_text() == (R / f'find-{s}-memory-after.txt').read_text()
base_sha = '529b6fdb427a7fa7ccd9bdf58d4347e72e49f16c'
head_sha = '82c1dc7ca68b74166657a332017e35a471c33061'
delta.update(base_commit=base_sha, head_commit=head_sha,
    typescript_commit='050880ce59e30b356b686bd3144efe24f875ebc8',
    latent_mode='first-error', find_context=find_context, matched_reasons=sorted({f['reason'] for f in all_findings['base'] if match(f['reason'], f['where'])}),
    find_sites=finds, find_occurrences=occurrences, sort_cross_check=historical,
    regression_bar=regressions, regression_bar_counts=dict(collections.Counter(r['type'] for r in regressions)),
    historical_find_summary=historical_summary, historical_find_new_reasons=instance_reasons,
    timing_sums=timing, resources=resources,
    box=json.loads((R / 'find-box.json').read_text()))
(OUT / 'delta.json').write_text(json.dumps(delta, indent=2) + '\n')

def esc(s):
    return str(s).replace('|', '\\|').replace('\n', ' ')

def stop(f):
    return 'no lowering finding' if not f else f"{f['kind']} at `{f['where']}`: {esc(f['reason'])}"

md = ['# Paired find-predicate census', '',
      f'Base compiler: `{base_sha}`. Head compiler: `{head_sha}`.', '',
      'Report only, not for landing. One adapted tree from the base apply.sh; two sequential first-error compiler-directory censuses.', '',
      f"The sort's 25 observations: **{historical_summary['gone']} gone, {historical_summary['moved']} moved, {historical_summary['unchanged']} unchanged**. All 25 reproduce on the base. The additional predicate-forwarding scan finding is unchanged.", '',
      '| Find observations | Gone | Moved | Unchanged |', '| --- | ---: | ---: | ---: |']
for name in finds:
    x = finds[name]['summary']
    md.append(f"| {name.replace('_', ' ')} | {x['gone']} | {x['moved']} | {x['unchanged']} |")
c = collections.Counter(t['status'] for t in occurrences)
md.append(f"| Recorded instances, paired by attempting unit | {c['gone']} | {c['moved']} | {c['unchanged']} |")
md += ['', 'New reasons for the 23 moved historical observations:', '']
md += [f'- {n}: {esc(reason)}' for reason, n in sorted(instance_reasons.items())]
md += ['', f"Regression bar: **{len(regressions)}** overlapping category entries. **The requested zero bar is not met.** Changed first lowering stops: **{len(delta['raw_findings']['unit_first_stop_delta'])}**.", '',
       'Exact matched base reason texts (overload or predicate, with find named in the reason or a stock-AST-verified direct find call at the diagnostic):', '']
md += ['- ' + esc(r) for r in delta['matched_reasons']]
md += ['', '## Every find finding', '',
       'Stock TypeScript classified every base predicate/overload diagnostic by its call ancestors. Generic predicate reasons are included only at direct find or ts.find calls; unrelated checked predicate overload target diagnostics are excluded. The context ledger is retained in delta.json. Raw findings use the meter key (kind, location, reason, text). Distinct reads use main\'s pair tool and the outermost contiguous access chain. These counts deduplicate repeated attempts; the instance table below preserves their progress to a later location.', '',
       '| Base location | Base reason | Status | Head reasons at this identity |', '| --- | --- | --- | --- |']
for t in finds['raw_findings']['transitions']:
    md.append(f"| `{t['base']['where']}` | {esc(t['base']['reason'])} | {t['status']} | {'; '.join(stop(f) for f in t['head']) or 'none'} |")
md += ['', '### Recorded find instances and their next stop', '',
       '| Attempting unit | Name | Base finding | Status | Head lowering stop or scan finding |', '| --- | --- | --- | --- | --- |']
for t in occurrences:
    f = t['base']
    md.append(f"| `{f['unit']}` | {esc(units['base'].get(f['unit'], {}).get('name', ''))} | {stop(f)} | {t['status']} | {stop(t['head'])} |")
md += ['', '## Cross-check against the core-walls sort', '',
       'Read CW003 from typescript/core-walls-sort at ad29102e9c145c5b480629227574d2ffa235a4b1. Its 25 instances are 25 reaching units for one overload-result refusal at core.ts:162:1, not 25 unique source locations. Twenty utilitiesPublic units reach it through getFirstJSDocTag. All 25 historical unit locations, finding locations and reasons exactly match this base; their outcomes are 2 gone and 23 moved. That entry-root run and this compiler-directory run have different eligibility and isolated contexts; the following table checks each historical reaching unit without treating the old count as a required current total.', '',
       '| Historical unit | Name | Current base stop | Current head stop |', '| --- | --- | --- | --- |']
for t in historical:
    md.append(f"| `{t['historical']['unit']}` | {esc(t['name'])} | {stop(t['base'])} | {stop(t['head'])} |")
md += ['', '## Regressions, the zero bar', '',
       'Includes previously finding-free attempted units or source identities in either phase, new panics, and every already-blocked unit whose first lowering stop changed to a new head stop. The last category can be exposure after progress, but is explicitly included in this task\'s bar. Categories can overlap at the same underlying exposure. Raw-only relocation differences are retained in delta.json.', '']
if not regressions:
    md.append('None.')
md += ['| Bar category | Count |', '| --- | ---: |']
bar_counts = collections.Counter(r['type'] for r in regressions)
for category in ('previously_finding_free_unit', 'previously_finding_free_scan_site', 'new_panic',
                 'previously_finding_free_lowering_site', 'changed_first_lowering_stop'):
    md.append(f'| {category} | {bar_counts[category]} |')
md += ['', 'The 20 new source identities are the Map/Set exposures in already-blocked find callers. Three certificate stops were already observed elsewhere on the base. Of 34 changed stops with a remaining head finding, 23 are find progress and 11 keep the findAncestor overload refusal at the same source location with a new failed-return suffix. These 11 are diagnostic-text changes; they remain in the requested exhaustive first-stop ledger. Two more units lose their lowering stop entirely.', '']
for r in regressions:
    if r['type'] == 'changed_first_lowering_stop':
        md.append(f"- Changed first stop, unit `{r['unit']['where']}` ({esc(r['unit'].get('name', ''))}): {stop(r['base'])} -> {stop(r['head'])}.")
    elif 'finding' in r:
        md.append(f"- {r['type']}: {stop(r['finding'])}; unit `{r['finding']['unit']}`.")
    else:
        md.append(f"- {r['type']}, unit `{r['unit']['where']}`: {stop(r['head'])}.")
md += ['', '## Unique finding kinds', '', '| Kind | Base raw | Head raw | Base distinct | Head distinct |', '| --- | ---: | ---: | ---: | ---: |']
for kind in ('NotYet', 'Refused', 'panic', 'error', 'SkippedDependency'):
    values = [delta[v]['totals'][s].get(kind, 0) for v,s in
              [('raw_findings','before'),('raw_findings','after'),('distinct_reads','before'),('distinct_reads','after')]]
    md.append(f"| {kind} | " + ' | '.join(map(str, values)) + ' |')
md += ['', '## Units whose first stop moved', '',
       'Only lowering-phase events define first stops. Refusal-scan owners are not treated as attempted units.', '',
       '| Unit | Name | Base first stop | Head first stop |', '| --- | --- | --- | --- |']
for r in delta['raw_findings']['unit_first_stop_delta']:
    md.append(f"| `{r['unit']['where']}` | {esc(r['unit'].get('name', ''))} | {stop(r['base'])} | {stop(r['head'])} |")
if not delta['raw_findings']['unit_first_stop_delta']:
    md.append('| None | | | |')
md += ['', '## Unit lowering times', '',
       'One sample on this box per compiler, sequentially, base before head. Sums of unit timers include synchronous dependency work and exclude checker/shared setup. They are not a speedup claim.', '',
       '| Compiler | Attempts | Wall sum seconds | CPU sum seconds | Process wall seconds | Peak RSS KiB |', '| --- | ---: | ---: | ---: | ---: | ---: |']
for s in streams:
    t, r = timing[s], resources[s]
    md.append(f"| {s} | {t['attempts']} | {t['wall_seconds']:.6f} | {t['cpu_seconds']:.6f} | {r['wall_seconds']:.6f} | {r['peak_rss_kib']} |")
md += ['', 'Box metadata:', '', '```json', json.dumps(delta['box'], indent=2), '```', '',
       '## Method, provenance and limits', '',
       'Copied the method after reading typescript/kind-predicate-census REPORT.md and compare.py at 33dca393d8b57dc942a25ce9262d61f724afc031, and the base meter README.md and twice-daily.sh. Stock TypeScript is v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. API packages were installed with the base lockfile and checked by apply. No adaptation or compiler code changed. The compiler diff is only the requested two commits b34735bd and 82c1dc7c.', '',
       'Each exact compiler worktree supplied make_overlay.py and its pinned submodules. Overlays were formatted and built with go build -buildvcs=false -overlay=... ./stage3/census/latent/tool. The base build and census finished before the head build and census. See run.sh and evidence for commands, pins, compiler cleanliness, hashes, resource records and logs. Submodules were initialized using command-local URL configuration, so submodule status prints an inactive marker; the resolved-pins JSON independently verifies each actual clean checkout against its gitlink. No dependency pin differs.', '',
       'Invocation for each compiler:', '', '```sh',
       'LATENT_TIMES=times.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 latent-census SHARED_ADAPTED/src/compiler latent.jsonl', '```', '',
       'Main\'s stage3/census/pair at 529b6fdb supplied the comparison (unchanged on the report parent). Headers, unit identities and attempted eligibility must agree. Both streams use the same adapted bytes; delta.json records their hashes. All recorded base find instances, all transitions and changed first stops are retained. Raw normalized streams and timers are archived under evidence/.', '',
       'A gone finding means no reason remains at that paired identity in this first-error observation. It does not prove full unit lowering or native compilation. Generic bodies can remain deferred, diagnosed bodies are skipped, isolated context can differ from whole-program lowering, and no backend runs.', '',
       'The complete report accounting was independently checked with validate.py, including a deliberately wrong headline count that must be rejected. validation.txt records the result. artifacts.sha256.json covers the archived streams, logs, scripts and delta.', '',
       'Preparation initially checked the TypeScript pin before its clone had completed and stopped without starting a census. The successful retry reused the completed clone. This has no measured observations in either stream.', '']
(OUT / 'REPORT.md').write_text('\n'.join(md).replace('\u2014', '-').replace('\u2013', '-'))
print(json.dumps(dict(find_sites={k:v['summary'] for k,v in finds.items()},
    instances=dict(c), regressions=len(regressions), first_stops=len(delta['raw_findings']['unit_first_stop_delta']), timing=timing), indent=2))
