"""Pair the two saved first-error streams with main's census/pair, then report."""
import collections
import datetime
import gzip
import hashlib
import importlib.util
import json
from pathlib import Path
import sys

OUT = Path(__file__).resolve().parent
TREE = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else OUT / 'scratch/adapted'
PAIR = OUT.parent / 'pair/pair.py'
spec = importlib.util.spec_from_file_location('census_pair', PAIR)
pair = importlib.util.module_from_spec(spec)
spec.loader.exec_module(pair)
E = OUT / 'evidence'
BASE = 'b0efce6dce1cd7445ed8a570aecac387a4a7d684'
HEAD = '66fff4ecc05c3bfb2258cf4bb2a7b0a2b762c59b'

def norm(value):
    if isinstance(value, str):
        return value.replace(str(TREE) + '/', '')
    if isinstance(value, list):
        return [norm(v) for v in value]
    if isinstance(value, dict):
        return {k: norm(v) for k, v in value.items()}
    return value

def load(label, kind):
    compressed = E / f'{label}-{kind}.jsonl.gz'
    plain = E / f'{label}-{kind}.jsonl'
    rows = pair.load(compressed if compressed.exists() else plain)
    rows = norm(rows)
    if not compressed.exists():
        with gzip.open(compressed, 'wt') as stream:
            for row in rows:
                stream.write(json.dumps(row) + '\n')
    return rows

D = {label: load(label, 'latent') for label in ('base', 'head')}
if any(not (E / f'{label}-exit.txt').exists() for label in D):
    import incomplete
    incomplete.write(OUT, TREE, pair, D, load, BASE, HEAD)
    raise SystemExit(0)
assert all(rows[0]['latent_mode'] == 'first-error' for rows in D.values())
assert all((E / f'{label}-exit.txt').read_text().strip() == '0' for label in D)
assert D['base'][0] == D['head'][0]
assert {row['file'] for row in D['base'][1:]} == {str(p.relative_to(TREE)) for p in (TREE / 'src/compiler').rglob('*.ts')}
assert {row['file'] for row in D['base'][1:]} == {row['file'] for row in D['head'][1:]}
result = pair.compare(D['base'], D['head'], TREE, OUT / 'scratch/base/stage3/api/node_modules/typescript')
(OUT / 'pair-report.md').write_text(pair.report(result))
result.update(base_commit=BASE, head_commit=HEAD, latent_mode='first-error', typescript_pin=json.loads((OUT / 'scratch/base/stage3/source.json').read_text()))
units = {label: {u['where']: u for row in rows[1:] for u in row['units']} for label, rows in D.items()}
first = {}
for row in D['head'][1:]:
    for f in row['findings']:
        if f['phase'] == 'lowering':
            first.setdefault(f['unit'], f)
any_members = []
seen = set()
for row in D['head'][1:]:
    for f in row['findings']:
        if not f['reason'].startswith('a member typed any'):
            continue
        key = tuple(f[k] for k in ('kind', 'where', 'reason', 'text', 'unit', 'phase'))
        if key in seen:
            continue
        seen.add(key)
        any_members.append(dict(finding=f, owner=units['head'].get(f['unit']), is_unit_first_stop=f['phase'] == 'lowering' and first.get(f['unit']) == f))
result['any_member_findings'] = any_members
T = {label: load(label, 'times') for label in D}
assert {x['where'] for x in T['base']} == {x['where'] for x in T['head']}
for label in D:
    assert len(T[label]) == len({x['where'] for x in T[label]})
    assert len(T[label]) == sum(u['status'] in ('attempted', 'panic') for u in units[label].values())
    assert (E / f'{label}-memory-before.txt').read_text() == (E / f'{label}-memory-after.txt').read_text()
    assert not (E / f'{label}-status.txt').read_text()
result['unit_time_sums'] = {label: {k: sum(x[k] for x in rows) for k in ('wall_seconds', 'cpu_seconds')} for label, rows in T.items()}
result['unit_time_sample_counts'] = {label: len(rows) for label, rows in T.items()}
times = {label: {r['where']: r for r in rows} for label, rows in T.items()}
result['unit_timing_delta'] = [dict(where=w, base=times['base'][w], head=times['head'][w], wall_delta_seconds=times['head'][w]['wall_seconds']-times['base'][w]['wall_seconds'], cpu_delta_seconds=times['head'][w]['cpu_seconds']-times['base'][w]['cpu_seconds']) for w in sorted(times['base'])]
result['process_elapsed_seconds_from_utc_timestamps'] = {label: (datetime.datetime.fromisoformat((E / f'{label}-end.txt').read_text().strip()) - datetime.datetime.fromisoformat((E / f'{label}-start.txt').read_text().strip())).total_seconds() for label in D}
manifest = {str(p.relative_to(TREE)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((TREE / 'src').rglob('*')) if p.is_file()}
manifest_path = E / 'source-manifest.json'
if manifest_path.exists():
    assert json.loads(manifest_path.read_text()) == manifest
else:
    manifest_path.write_text(json.dumps(manifest, indent=2) + '\n')
result['box'] = json.loads((E / 'box.json').read_text())
result['cgroup'] = json.loads((E / 'cgroup.json').read_text())
(OUT / 'delta.json').write_text(json.dumps(result, indent=2) + '\n')

def esc(s):
    return str(s).replace('|', '\\|').replace('\n', ' ')

lines = ['# Paired any-members census (tmdhz1d)', '', f'**Any-member findings: {len(any_members)}. Unit first stops among them: {sum(r["is_unit_first_stop"] for r in any_members)}.**', '']
for r in any_members:
    f = r['finding']
    owner = r['owner'] or {}
    lines += [f'- `{f["where"]}`; owner `{f["unit"]}` ({owner.get("name", "file scan")}); phase `{f["phase"]}`; first stop: {"yes" if r["is_unit_first_stop"] else "no"}. Reason: {f["reason"]}']
if not any_members:
    lines += ['No head finding has a reason starting `a member typed any`.']
lines += ['', f'Base: `{BASE}`. Head: `{HEAD}`.', '', 'TypeScript v6.0.3, upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8`.', '', f'Box: `{result["box"]["hostname"]}`; {result["box"]["nproc"]} available processors; CPU quota `{result["box"]["cpu_quota"]}`; host memory `{result["box"]["total_memory"]}`; cgroup memory limit `{result["cgroup"]["memory.max"]}` bytes (16 GiB).', '', '## Regressions and changed first stops', '']
for view in ('distinct_reads', 'raw_findings'):
    v = result[view]
    lines += [f'### {view.replace("_", " ").title()}', '', f'Regressions: {len(v["regressions"])}. Changed first stops: {len(v["unit_first_stop_delta"])}.', '']
    for r in v['regressions']:
        f = r.get('finding') or r.get('head')
        lines += [f'- {r["type"]}: `{f["where"]}`; owner `{f["unit"]}`; {f["kind"]}: {f["reason"]}']
    lines += [f'New source sites: {len(v["new_sites"])}.', '']
    for r in v['new_sites']:
        f = r['finding']
        lines += [f'- `{f["where"]}`; owner `{f["unit"]}`; phase `{f["phase"]}`; {f["kind"]}: {f["reason"]}']
    for r in v['unit_first_stop_delta']:
        lines += [f'- Unit `{r["unit"]["where"]}` ({r["unit"].get("name", r["unit"]["kind"])}):']
        for label in ('base', 'head'):
            f = r[label]
            lines += [f'  - {label}: `{f["where"]}`; {f["kind"]}: {f["reason"]}' if f else f'  - {label}: finding-free.']
lines += ['', '## Unique findings', '', 'Meter counts deduplicate `(kind, where, reason, text)` across attempts.', '', '| Kind | Base | Head |', '| --- | ---: | ---: |']
for kind in ('NotYet', 'Refused', 'panic', 'error'):
    totals = result['raw_findings']['totals']
    lines += [f'| {kind} | {totals["before"].get(kind, 0)} | {totals["after"].get(kind, 0)} |']
lines += ['', '## Unit lowering time', '', 'One sequential sample per compiler on this box. Sums of unit timers include registration and synchronous dependency work, and exclude checker work, file scans and shared setup.', '', '| Sum | Base seconds | Head seconds |', '| --- | ---: | ---: |']
for k in ('wall_seconds', 'cpu_seconds'):
    lines += [f'| {k} | {result["unit_time_sums"]["base"][k]:.6f} | {result["unit_time_sums"]["head"][k]:.6f} |']
lines += ['', '## Method and validation', '', 'Copied the method in `typescript/kind-predicate-census` (`33dca393`), after reading its REPORT.md and compare.py and the meter launcher and README. One shared adapted tree was made with base stage3/apply.sh, its lockfile-pinned API packages and the upstream TypeScript pin above. Both exact compiler worktrees use their recorded submodule pins. The head has no stage3 changes.', '', f'Process elapsed times from UTC start/end records: base {result["process_elapsed_seconds_from_utc_timestamps"]["base"]:.0f} seconds; head {result["process_elapsed_seconds_from_utc_timestamps"]["head"]:.0f} seconds. The base build and run completed before the head build and run. Each guarded overlay was generated by main stage3/census/latent/make_overlay.py, formatted with gofmt and built with `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`. Each ran sequentially over the same `src/compiler` with `LATENT_TIMES=... LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`. No backend was invoked.', '', 'The findings file stayed at its header while unit timers advanced. This main revision buffers every file record inside Latent until all callers finish so deferred-generic instantiation witnesses can be finalized. A header-only stream therefore did not establish a stall; no sustained timer stall was observed and neither mode was changed. Main stage3/census/pair supplies both raw and access-chain-normalized pairing. Only lowering-phase events define unit first stops; refusal-scan owners are retained separately. delta.json preserves all changed sites, first stops, regressions, panics, any-member owners and all unit timers. pair-report.md retains the standard pair report.', '', f'Validated matching first-error headers, all {len(D["base"])-1} compiler file rows, unit identities and eligibility, and {len(T["base"])} timed attempts per run. Both runs exited 0, both compiler status files are empty, and cgroup memory event counters were unchanged. Source hashes and compressed raw streams are retained in evidence/.', '', f'The header reports checker_rejected={D["base"][0]["checker_rejected"]} with {len(D["base"][0]["diagnostics"])} checker diagnostics. The tool retains its fixed measurement label verbatim in the evidence. Finding-free means no finding in this first-error census; later blockers and deferred or checker-skipped bodies remain outside that claim. Any zero result is limited to these eligible attempts and scans.', '']
(OUT / 'REPORT.md').write_text('\n'.join(lines))
print(json.dumps(dict(any_member_findings=any_members, regressions=result['distinct_reads']['regressions'], changed_first_stops=result['distinct_reads']['unit_first_stop_delta'], counts=result['raw_findings']['totals'], time_sums=result['unit_time_sums']), indent=2))
