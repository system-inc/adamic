"""Union census boundaries and checker skips, then subtract independent coverage."""
import argparse
from bisect import bisect_right
from collections import Counter, defaultdict
import gzip
import hashlib
import json
from pathlib import Path


def union(spans):
    result = []
    for start, end in sorted(spans):
        if start == end:
            continue
        if not 0 <= start < end:
            raise ValueError(f'invalid half-open span: {start}, {end}')
        if result and start <= result[-1][1]:
            result[-1] = (result[-1][0], max(end, result[-1][1]))
        else:
            result.append((start, end))
    return result


def subtract(spans, examined):
    cuts = union(examined)
    result = []
    index = 0
    for start, end in union(spans):
        while index < len(cuts) and cuts[index][1] <= start:
            index += 1
        cursor, current = start, index
        while current < len(cuts) and cuts[current][0] < end:
            left, right = cuts[current]
            if cursor < left:
                result.append((cursor, min(left, end)))
            cursor = max(cursor, right)
            if cursor >= end:
                break
            current += 1
        if cursor < end:
            result.append((cursor, end))
    return result


def size(spans):
    return sum(end - start for start, end in spans)


def read_rows(path):
    opener = gzip.open if path.suffix == '.gz' else open
    with opener(path, 'rt') as stream:
        return [json.loads(line) for line in stream]


def read_json(path):
    opener = gzip.open if path.suffix == '.gz' else open
    with opener(path, 'rt') as stream:
        return json.load(stream)


def local_file(name, original):
    try:
        return Path(name).relative_to(original).as_posix()
    except ValueError:
        return str(Path(name))


def local_where(where, original):
    name, line, column = where.rsplit(':', 2)
    return f'{local_file(name, original)}:{line}:{column}'


def calculate(rows, stock, original):
    if rows[0].get('latent_mode') != 'full':
        raise ValueError('requires a full latent census')
    blocked, examined = defaultdict(list), defaultdict(list)
    boundaries = defaultdict(list)
    skipped = defaultdict(list)
    counts = Counter()
    observed = set()
    for row in rows[1:]:
        name = local_file(row['file'], original)
        if name not in stock:
            continue  # Imported files outside src/compiler are not this denominator.
        if name in observed:
            raise ValueError(f'duplicate file record: {name}')
        observed.add(name)
        for finding in row['findings']:
            if finding['kind'] == 'SkippedDependency':
                where = local_where(finding['where'], original)
                target = where.rsplit(':', 2)[0]
                if target not in stock:
                    continue
                body = next((body for body in stock[target].get('bodies', []) if body['where'] == where), None)
                if body is None:
                    raise ValueError(f'skipped dependency absent from stock bodies: {where}')
                start, end = body['start'], body['end']
                blocked[target].append(dict(start=start, end=end, category='checker_dependency',
                    reason=finding['reason'], failure=finding['text'], unit=where))
                skipped[target].append((start, end))
                counts['skipped_dependency_records'] += 1
            if finding['kind'] != 'Boundary':
                continue
            target = local_where(finding['where'], original).rsplit(':', 2)[0]
            if target not in stock:
                continue
            start, end = finding.get('start', 0), finding.get('end', 0)
            if not 0 <= start < end <= stock[target]['bytes']:
                raise ValueError(f'invalid boundary: {finding}')
            entry = dict(start=start, end=end, category='Boundary', reason=finding['reason'],
                         failure=finding['text'], unit=local_where(finding['unit'], original))
            blocked[target].append(entry)
            boundaries[(target, finding['unit'])].append((start, end))
            counts['boundary_records'] += 1
        for unit in row['units']:
            if unit['status'] not in ('split_checker_body', 'skipped_checker_body'):
                continue
            start, end = unit['body_start'], unit['body_end']
            if not 0 <= start < end <= stock[name]['bytes']:
                raise ValueError(f'invalid skipped body: {unit}')
            blocked[name].append(dict(start=start, end=end, category='checker',
                reason='checker diagnostics in own body; body skipped',
                failure='\n'.join(unit['checker_diagnostics']),
                unit=local_where(unit['where'], original)))
            skipped[name].append((start, end))
            counts['checker_skipped_bodies'] += 1
    expected = {name for name in stock if name.endswith(('.ts', '.a'))}
    if observed != expected:
        raise ValueError(f'incomplete compiler files: {sorted(expected - observed)}')
    # An independent attempt exposes its declaration except for its own failures
    # and diagnosed children. A parent's failed body does not veto a child's attempt.
    for row in rows[1:]:
        name = local_file(row['file'], original)
        if name not in stock:
            continue
        catalog = {unit['where']: unit for unit in stock[name]['units']}
        for unit in row['units']:
            where = local_where(unit['where'], original)
            external = catalog.get(where)
            if external is None:
                raise ValueError(f'unit absent from stock AST: {where}')
            if unit.get('body_end', 0) != external['body_end'] or unit.get('body_start', 0) != external['body_start']:
                # Body fields exist only for function declarations in the census.
                if external['function']:
                    raise ValueError(f'body span disagrees with stock AST: {where}')
            if unit['status'] in ('split_checker_body', 'skipped_checker_body'):
                continue
            if unit['status'] not in ('attempted', 'panic'):
                raise ValueError(f'unknown unit status: {unit["status"]}')
            cuts = boundaries[(name, unit['where'])] + [span for span in skipped[name]
                if external['start'] <= span[0] and span[1] <= external['end']]
            examined[name].extend(subtract([(external['start'], external['end'])], cuts))
            counts['independent_attempts'] += 1
    files, regions = {}, []
    for name, metadata in sorted(stock.items()):
        blocked_union = union((entry['start'], entry['end']) for entry in blocked[name])
        hidden = subtract(blocked_union, examined[name])
        removed = size(blocked_union) - size(hidden)
        files[name] = dict(bytes=metadata['bytes'], sha256=metadata['sha256'],
            blocked_union_bytes=size(blocked_union), independently_examined_bytes=removed,
            hidden_bytes=size(hidden), hidden_share=size(hidden) / metadata['bytes'] if metadata['bytes'] else 0,
            hidden_ranges=hidden)
        for start, end in hidden:
            causes = [entry for entry in blocked[name] if entry['start'] < end and entry['end'] > start]
            regions.append(dict(file=name, start=start, end=end, bytes=end-start,
                causes=sorted({(entry['category'], entry['reason'], entry['failure'], entry['unit']) for entry in causes})))
    total = sum(entry['bytes'] for entry in files.values())
    hidden_total = sum(entry['hidden_bytes'] for entry in files.values())
    return dict(total_bytes=total, hidden_bytes=hidden_total, hidden_share=hidden_total / total,
        typescript_source_bytes=sum(entry['bytes'] for name, entry in files.items() if name.endswith(('.ts', '.a'))),
        hidden_share_of_typescript_source=hidden_total / sum(entry['bytes'] for name, entry in files.items() if name.endswith(('.ts', '.a'))),
        blocked_union_bytes=sum(entry['blocked_union_bytes'] for entry in files.values()),
        independently_examined_bytes=sum(entry['independently_examined_bytes'] for entry in files.values()),
        counts=dict(counts), files=files,
        largest_regions=sorted(regions, key=lambda entry: (-entry['bytes'], entry['file'], entry['start']))[:10])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('compiler', type=Path)
    parser.add_argument('census', type=Path)
    parser.add_argument('stock', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('--commit', required=True)
    args = parser.parse_args()
    stock = read_json(args.stock)
    actual = {path.relative_to(args.compiler).as_posix() for path in args.compiler.rglob('*') if path.is_file()}
    if actual != set(stock):
        raise ValueError('stock manifest does not cover all compiler files')
    for name, metadata in stock.items():
        data = (args.compiler / name).read_bytes()
        if len(data) != metadata['bytes'] or hashlib.sha256(data).hexdigest() != metadata['sha256']:
            raise ValueError(f'source hash disagrees: {name}')
    rows = read_rows(args.census)
    original = Path(rows[1]['file'].split('/src/compiler/')[0]) / 'src/compiler'
    result = calculate(rows, stock, original)
    for region in result['largest_regions']:
        data = (args.compiler / region['file']).read_bytes()
        lines = [0] + [index + 1 for index, value in enumerate(data) if value == 10]
        region['start_line'] = bisect_right(lines, region['start'])
        region['end_line'] = bisect_right(lines, region['end'] - 1)
    result['provenance'] = dict(compiler_commit=args.commit, compiler_root=str(args.compiler),
        census_sha256=hashlib.sha256(args.census.read_bytes()).hexdigest(),
        stock_sha256=hashlib.sha256(args.stock.read_bytes()).hexdigest(),
        measurement=rows[0]['measurement'], checker_rejected=rows[0]['checker_rejected'])
    result['definition'] = 'UTF-8 half-open byte ranges; union(Boundary, checker-skipped unit and dependency bodies) minus union(independent attempted declaration spans minus their own boundaries and checker-skipped bodies). All regular src/compiler files form the denominator, including comments, whitespace and non-code files.'
    args.output.write_text(json.dumps(result, indent=2) + '\n')
    print(f"hidden {result['hidden_bytes']:,} / {result['total_bytes']:,} bytes ({result['hidden_share']:.6%}); subtracted {result['independently_examined_bytes']:,}")


if __name__ == '__main__':
    main()
