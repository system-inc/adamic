"""Compare same-input latent streams, preserving raw findings and distinct reads."""
import argparse
from collections import Counter, defaultdict
import gzip
import json
from pathlib import Path
import subprocess


def load(path):
    opener = gzip.open if str(path).endswith('.gz') else open
    with opener(path, 'rt', encoding='utf-8') as stream:
        return [json.loads(line) for line in stream if line.strip()]


def relative(file, tree):
    p = Path(file)
    if not p.is_absolute():
        result = p
    elif p.is_relative_to(tree):
        result = p.relative_to(tree)
    elif '/src/' in file:
        # Archived streams retain the original worker's absolute tree root.
        result = Path('src/' + file.split('/src/', 1)[1])
    else:
        raise ValueError(f'cannot locate {file} in {tree}')
    if '..' in result.parts:
        raise ValueError(f'source escapes tree: {file}')
    return result.as_posix()


def identities(streams, tree, typescript):
    files = defaultdict(list)
    seen = set()
    for rows in streams:
        for row in rows[1:]:
            for item in (row.get('findings') or []) + (row.get('units') or []):
                for field in ('where', 'unit'):
                    where = item.get(field)
                    if not where or where in seen:
                        continue
                    seen.add(where)
                    file, line, column = where.rsplit(':', 2)
                    files[relative(file, tree)].append([where, int(line), int(column)])
    request = dict(tree=str(tree), typescript=str(typescript), files=files)
    run = subprocess.run(['node', str(Path(__file__).with_name('normalize.cjs'))],
                         input=json.dumps(request), text=True, capture_output=True)
    if run.returncode:
        raise ValueError(f'stock TypeScript normalization failed:\n{run.stderr}')
    return json.loads(run.stdout)


def compare(base, head, tree, typescript, normalize=True):
    tree = Path(tree).resolve()
    ids = identities([base, head], tree, Path(typescript).resolve())
    if base[0] != head[0]:
        raise ValueError('measurement headers differ')

    def site(where, distinct):
        i = ids[where]
        return (i['file'], i['sha256'], i['read_offset' if distinct and normalize else 'byte_offset'])

    def view(rows, distinct):
        findings, units, first = {}, {}, {}
        for row in rows[1:]:
            for u in row.get('units') or []:
                units[site(u['where'], False)] = u
            for f in row.get('findings') or []:
                key = (site(f['where'], distinct), f['reason']) if distinct else (
                    site(f['where'], False), f['kind'], f['reason'], f['text'])
                findings.setdefault(key, f)
                if f.get('phase') == 'lowering':
                    first.setdefault(site(f['unit'], False), f)
        bysite = defaultdict(list)
        for f in findings.values():
            bysite[site(f['where'], distinct)].append(f)
        return findings, bysite, units, first

    result = {'schema_version': 1, 'normalization': normalize, 'source_hashes': {
        i['file']: i['sha256'] for i in ids.values()}}
    for name, distinct in [('raw_findings', False), ('distinct_reads', True)]:
        bf, bs, bu, bfirst = view(base, distinct)
        hf, hs, hu, hfirst = view(head, distinct)
        if bu.keys() != hu.keys():
            raise ValueError('unit identities differ')
        for u in bu:
            stable = lambda v: {k: x for k, x in v.items() if k not in ('where', 'status', 'instantiated')}
            if stable(bu[u]) != stable(hu[u]) or ((bu[u]['status'] in ('attempted', 'panic')) !=
                                               (hu[u]['status'] in ('attempted', 'panic'))):
                raise ValueError(f'unit eligibility differs: {u}')
        reasons = sorted({f['reason'] for f in list(bf.values()) + list(hf.values())})
        per_reason = {}
        lowered, moved = [], []
        for key in sorted(bf.keys() - hf.keys(), key=str):
            f = bf[key]
            after = hs.get(site(f['where'], distinct), [])
            row = dict(identity=site(f['where'], distinct), base=f, head=after)
            (moved if after else lowered).append(row)
        new = [dict(identity=site(f['where'], distinct), finding=f)
               for key, f in hf.items() if key not in bf]
        new_sites = [r for r in new if tuple(r['identity']) not in bs]
        for reason in reasons:
            per_reason[reason] = dict(before=sum(f['reason'] == reason for f in bf.values()),
                                     after=sum(f['reason'] == reason for f in hf.values()),
                                     lowered=sum(r['base']['reason'] == reason for r in lowered),
                                     moved_by_reason=dict(Counter(g['reason'] for r in moved
                                         if r['base']['reason'] == reason for g in r['head'])))
        def stop(f):
            return (site(f['where'], distinct), f['kind'], f['reason']) if f else None
        units = [dict(identity=u, unit=bu[u], base=bfirst.get(u), head=hfirst.get(u))
                 for u in sorted(bu) if stop(bfirst.get(u)) != stop(hfirst.get(u))]
        regressions = [dict(type='previously_finding_free_unit', **r) for r in units
                       if not r['base'] and r['head']]
        regressions += [dict(type='new_panic', identity=site(f['where'], distinct), finding=f)
                        for f in hf.values() if f['kind'] == 'panic' and not any(
                            old['kind'] == 'panic' and old['reason'] == f['reason']
                            for old in bs.get(site(f['where'], distinct), []))]
        regressions += [dict(type='previously_finding_free_scan_site', **r) for r in new_sites
                        if r['finding'].get('phase') == 'refusal_scan']
        result[name] = dict(per_reason=per_reason, lowered=lowered, moved=moved,
                            new_findings=new, new_sites=new_sites, unit_first_stop_delta=units,
                            regressions=regressions,
                            totals={s: dict(Counter(f['kind'] for f in fs.values()))
                                    for s, fs in [('before', bf), ('after', hf)]},
                            panics={s: [f for f in fs.values() if f['kind'] == 'panic']
                                    for s, fs in [('before', bf), ('after', hf)]})
    return result


def report(delta):
    lines = ['# Paired latent census', '',
             '| Measure | Raw findings | Distinct reads |', '| --- | ---: | ---: |']
    for label, field in [('Lowered (no remaining finding)', 'lowered'), ('Moved by reason', 'moved'),
                         ('New findings', 'new_findings'), ('New sites', 'new_sites'),
                         ('Units whose first stop moved', 'unit_first_stop_delta'), ('Regressions', 'regressions')]:
        lines.append(f"| {label} | {len(delta['raw_findings'][field])} | {len(delta['distinct_reads'][field])} |")
    for side in ('before', 'after'):
        for kind in sorted(set(delta['raw_findings']['totals'][side]) | set(delta['distinct_reads']['totals'][side])):
            lines.append(f"| {side} {kind} | {delta['raw_findings']['totals'][side].get(kind, 0)} | {delta['distinct_reads']['totals'][side].get(kind, 0)} |")
    lines += ['', '| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |',
              '| --- | ---: | ---: | ---: | ---: | ---: |']
    for reason, r in delta['distinct_reads']['per_reason'].items():
        raw = delta['raw_findings']['per_reason'][reason]
        escaped = reason.replace('|', '\\|').replace('\n', ' ')
        lines.append(f"| {escaped} | {raw['before']} | {raw['after']} | {r['before']} | {r['after']} | {r['lowered']} |")
    for name in ('raw_findings', 'distinct_reads'):
        v = delta[name]
        lines += ['', f'## {name.replace("_", " ").title()}', '', 'Moved by reason:', '']
        for reason, counts in v['per_reason'].items():
            for after, count in counts['moved_by_reason'].items():
                lines.append(f'- {count}: {reason} -> {after}')
        for title, field in [('New findings', 'new_findings'), ('Units whose first stop moved', 'unit_first_stop_delta')]:
            lines += ['', title + ':', '']
            for row in v[field]:
                f = row.get('finding') or row.get('head') or row.get('base')
                lines.append(f'- `{f["where"]}`: {f["kind"]}: {f["reason"]}')
            if not v[field]:
                lines.append('None.')
        lines += ['', f'Panics: {len(v["panics"]["before"])} before, {len(v["panics"]["after"])} after.']
    lines += ['', 'The JSON retains every lowered site, move, new finding, unit stop and panic.', '']
    return '\n'.join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('base', type=Path)
    parser.add_argument('head', type=Path)
    parser.add_argument('tree', type=Path)
    parser.add_argument('--typescript', type=Path, help='stock 6.0.3 package or typescript.js from this tree')
    parser.add_argument('--json', required=True, type=Path)
    parser.add_argument('--report', required=True, type=Path)
    args = parser.parse_args()
    ts = args.typescript or args.tree / 'node_modules/typescript'
    delta = compare(load(args.base), load(args.head), args.tree, ts)
    args.json.write_text(json.dumps(delta, indent=2) + '\n')
    args.report.write_text(report(delta))


if __name__ == '__main__':
    main()
