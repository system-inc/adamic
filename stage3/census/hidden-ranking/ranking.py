"""Partition recorded hidden bytes by outermost stopping cause."""
import hashlib
import json
from collections import defaultdict
from pathlib import Path
import re
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent.parent / 'hidden'))
import hidden

ROOT = Path(__file__).resolve().parent


def owners():
    result = {}
    # Read the complete pinned tables, including reasons now absent from the
    # final snapshot. Exact reason matches only; no guessed template ownership.
    for name in ('refusal-baseline.md', 'refusal-table.md'):
        for line in (ROOT / 'evidence' / name).read_text().splitlines():
            columns = [part.strip().replace('\\|', '|') for part in re.split(r'(?<!\\)\|', line)]
            if len(columns) in (10, 12) and columns[2].startswith('internal/lower/'):
                result[('Refused', columns[1])] = dict(owner=columns[2], disposition=columns[-3], owner_source=name)
    for row in json.loads((ROOT / 'evidence/notyet-summary.json').read_text())['rows']:
        result[('NotYet', row['reason'])] = dict(owner=None, disposition=row['disposition'],
            owner_source='notyet-summary.json: table contains no owner column')
    return result


def boundaries(rows, stock, original):
    diagnostics = defaultdict(set)
    for row in rows[1:]:
        for finding in row['findings']:
            if finding['kind'] != 'Boundary':
                kind = finding['kind']
                if kind in ('error', 'panic') and 'panic:' in finding['text']:
                    kind = 'panic'
                diagnostics[finding['text']].add((kind, finding['reason']))
    result = []
    for row in rows[1:]:
        for finding in row['findings']:
            if finding['kind'] != 'Boundary':
                continue
            where = hidden.local_where(finding['where'], original)
            name = where.rsplit(':', 2)[0]
            if name not in stock:
                continue
            causes = diagnostics[finding['text']]
            if len(causes) != 1:
                raise ValueError(f'boundary has no unique diagnostic cause: {finding}')
            kind, reason = next(iter(causes))
            result.append(dict(file=name, start=finding['start'], end=finding['end'], kind=kind,
                reason=reason, where=where, unit=hidden.local_where(finding['unit'], original),
                recovery=finding['reason'], diagnostic=finding['text']))
        name = hidden.local_file(row['file'], original)
        if name not in stock:
            continue
        for unit in row['units']:
            if unit['status'] in ('split_checker_body', 'skipped_checker_body'):
                result.append(dict(file=name, start=unit['body_start'], end=unit['body_end'],
                    kind='checker', reason='checker-rejected body', where=hidden.local_where(unit['where'], original),
                    unit=hidden.local_where(unit['where'], original), diagnostic=unit['checker_diagnostics']))
        for finding in row['findings']:
            if finding['kind'] == 'SkippedDependency':
                where = hidden.local_where(finding['where'], original)
                target = where.rsplit(':', 2)[0]
                if target not in stock:
                    continue
                body = next(body for body in stock[target]['bodies'] if body['where'] == where)
                result.append(dict(file=target, start=body['start'], end=body['end'], kind='checker',
                    reason='checker-rejected body', where=where, unit=where, diagnostic=finding['text']))
    return result


def partition(entries, files, owner_table, mutant=False):
    grouped = defaultdict(list)
    totals = defaultdict(int)
    contributing = defaultdict(set)
    segments = []
    for index, entry in enumerate(entries):
        entry['id'] = index
        entry['attributed_hidden_bytes'] = 0
        grouped[entry['file']].append(entry)
    for name, metadata in sorted(files.items()):
        spans = grouped[name]
        for left, right in metadata['hidden_ranges']:
            relevant = [e for e in spans if e['start'] < right and e['end'] > left]
            points = sorted({left, right} | {max(left, e['start']) for e in relevant} |
                            {min(right, e['end']) for e in relevant})
            for start, end in zip(points, points[1:]):
                active = [e for e in relevant if e['start'] <= start and end <= e['end']]
                if not active:
                    raise ValueError(f'unexplained hidden bytes: {name}:{start}-{end}')
                # True enclosing spans win. Crossing spans and equal-span causes
                # cannot identify one outermost reason and remain conflict buckets.
                outer = [e for e in active if not any(o['start'] <= e['start'] and e['end'] <= o['end']
                    and (o['start'], o['end']) != (e['start'], e['end']) for o in active)]
                causes = sorted({(e['kind'], e['reason']) for e in outer})
                key = causes[0] if len(causes) == 1 else ('conflict', json.dumps(causes))
                totals[key] += end - start
                contributing[key].update(e['id'] for e in outer)
                representative = min(outer, key=lambda e: e['id'])
                representative['attributed_hidden_bytes'] += end - start
                segments.append(dict(file=name, start=start, end=end, kind=key[0], reason=key[1],
                    boundary=representative['id'], competing_causes=causes if len(causes) > 1 else []))
                if mutant and len(active) > len(outer):
                    totals[key] += 1  # Mutant: count one nested byte twice.
    # Include even fully examined or shadowed reasons with zero byte credit.
    for entry in entries:
        totals[(entry['kind'], entry['reason'])] += 0
    ranking = []
    for (kind, reason), count in totals.items():
        matching = [e for e in entries if (e['kind'], e['reason']) == (kind, reason)]
        ids = contributing[(kind, reason)]
        examples = list(dict.fromkeys(entries[i]['where'].rsplit(':', 1)[0] for i in
            sorted(ids, key=lambda i: (-entries[i]['attributed_hidden_bytes'], entries[i]['where']))))[:3]
        unique = {(e['file'], e['start'], e['end']) for e in matching}
        info = owner_table.get((kind, reason), dict(owner=None, disposition=None,
            owner_source='exact reason absent from supplied tables'))
        ranking.append(dict(kind=kind, reason=reason, bytes_revealed_if_fixed_alone=count,
            boundary_count=len(unique), raw_boundary_records=len(matching),
            contributing_boundary_count=len({(entries[i]['file'], entries[i]['start'], entries[i]['end']) for i in ids}),
            examples=examples, **info))
    ranking.sort(key=lambda r: (-r['bytes_revealed_if_fixed_alone'], r['kind'], r['reason']))
    expected = sum(f['hidden_bytes'] for f in files.values())
    assert sum(totals.values()) == expected, 'each hidden byte must count exactly once'
    return dict(hidden_bytes=expected, ranked_reasons=[r for r in ranking if r['kind'] in ('NotYet', 'Refused')],
        other_causes=[r for r in ranking if r['kind'] not in ('NotYet', 'Refused')],
        attributed_segments=segments, boundaries=entries)


def main():
    source = ROOT.parent / 'hidden'
    rows = hidden.read_rows(source / 'evidence/full.jsonl.gz')
    stock = hidden.read_json(source / 'evidence/stock.json.gz')
    measured = hidden.read_json(source / 'RESULT.json')
    original = Path(rows[1]['file'].split('/src/compiler/')[0]) / 'src/compiler'
    recomputed = hidden.calculate(rows, stock, original)
    assert json.loads(json.dumps(recomputed['files'])) == measured['files'], 'frozen hidden ranges must reproduce'
    entries = boundaries(rows, stock, original)
    result = partition(entries, measured['files'], owners())
    result['provenance'] = dict(compiler_commit=measured['provenance']['compiler_commit'],
        refusal_table_commit='d35a81d36fdafccf827bad0f572d311b2a0d4deb',
        notyet_table_commit='dc6b1529ae9d2a2210672e105c8bb6374619a59d',
        hashes={str(p.relative_to(ROOT.parent)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in [source / 'RESULT.json', source / 'evidence/full.jsonl.gz', source / 'evidence/stock.json.gz',
                *sorted((ROOT / 'evidence').glob('*.md')), ROOT / 'evidence/notyet-summary.json']})
    (ROOT / 'RESULT.json').write_text(json.dumps(result, indent=2) + '\n')
    print(f"PASS: {len(entries)} boundary/skip records; {result['hidden_bytes']} bytes partitioned exactly once")
    print('reason bytes:', sum(r['bytes_revealed_if_fixed_alone'] for r in result['ranked_reasons']))
    print('other bytes:', sum(r['bytes_revealed_if_fixed_alone'] for r in result['other_causes']))


if __name__ == '__main__':
    main()
