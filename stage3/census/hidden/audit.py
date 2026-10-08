"""Independently check reported arithmetic using one byte per source byte."""
import gzip
import json
import re
from pathlib import Path
import sys

raw, stock_path, result_path = map(Path, sys.argv[1:])
with (gzip.open(raw, 'rt') if raw.suffix == '.gz' else raw.open()) as stream:
    rows = [json.loads(line) for line in stream]
with (gzip.open(stock_path, 'rt') if stock_path.suffix == '.gz' else stock_path.open()) as stream:
    stock = json.load(stream)
result = json.loads(result_path.read_text())
root = Path(rows[1]['file'].split('/src/compiler/')[0]) / 'src/compiler'


def location(where):
    name, line, column = where.rsplit(':', 2)
    try:
        name = Path(name).relative_to(root).as_posix()
    except ValueError:
        pass
    return name, f'{name}:{line}:{column}'


blocked = {name: bytearray(entry['bytes']) for name, entry in stock.items()}
covered = {name: bytearray(entry['bytes']) for name, entry in stock.items()}
failures, skips = {}, {}
for row in rows[1:]:
    for finding in row['findings']:
        if finding['kind'] not in ('Boundary', 'SkippedDependency'):
            continue
        name, where = location(finding['where'])
        if name not in stock:
            continue
        if finding['kind'] == 'Boundary':
            start, end = finding['start'], finding['end']
            failures.setdefault((name, finding['unit']), []).append((start, end))
        else:
            body = next(body for body in stock[name]['bodies'] if body['where'] == where)
            start, end = body['start'], body['end']
            skips.setdefault(name, []).append((start, end))
        blocked[name][start:end] = b'\x01' * (end - start)
    for unit in row['units']:
        if unit['status'] in ('split_checker_body', 'skipped_checker_body'):
            name, _ = location(unit['where'])
            if name in stock:
                start, end = unit['body_start'], unit['body_end']
                blocked[name][start:end] = b'\x01' * (end - start)
                skips.setdefault(name, []).append((start, end))
for row in rows[1:]:
    for unit in row['units']:
        if unit['status'] in ('split_checker_body', 'skipped_checker_body'):
            continue
        name, where = location(unit['where'])
        if name not in stock:
            continue
        external = next(candidate for candidate in stock[name]['units'] if candidate['where'] == where)
        start, end = external['start'], external['end']
        mask = bytearray(b'\x01' * (end - start))
        cuts = failures.get((name, unit['where']), []) + [span for span in skips.get(name, [])
            if start <= span[0] and span[1] <= end]
        for left, right in cuts:
            left, right = max(start, left), min(end, right)
            if left < right:
                mask[left-start:right-start] = b'\x00' * (right-left)
        for index, value in enumerate(mask, start):
            if value:
                covered[name][index] = 1
hidden_total = blocked_total = removed_total = 0
regions = []
assert set(result['files']) == set(stock), 'file coverage'
for name in stock:
    hidden_mask = bytearray(a and not b for a, b in zip(blocked[name], covered[name]))
    count = sum(hidden_mask)
    reported = result['files'][name]
    assert reported['hidden_bytes'] == count, f'file hidden bytes: {name}'
    assert reported['blocked_union_bytes'] == sum(blocked[name]), f'file union: {name}'
    assert reported['independently_examined_bytes'] == sum(blocked[name]) - count, f'file subtraction: {name}'
    assert reported['hidden_share'] == count / stock[name]['bytes'], f'file share: {name}'
    reconstructed = bytearray(stock[name]['bytes'])
    for start, end in reported['hidden_ranges']:
        reconstructed[start:end] = b'\x01' * (end-start)
    assert reconstructed == hidden_mask, f'file hidden ranges: {name}'
    assert sum(end-start for start, end in reported['hidden_ranges']) == count, f'overlapping reported ranges: {name}'
    for match in re.finditer(b'\x01+', hidden_mask):
        regions.append((match.end()-match.start(), name, match.start(), match.end()))
    hidden_total += count
    blocked_total += sum(blocked[name])
    removed_total += sum(blocked[name]) - count
assert result['total_bytes'] == sum(entry['bytes'] for entry in stock.values()), 'denominator'
assert result['hidden_bytes'] == hidden_total, 'headline hidden total'
assert result['blocked_union_bytes'] == blocked_total, 'headline union'
assert result['independently_examined_bytes'] == removed_total, 'headline subtraction'
assert result['hidden_share'] == hidden_total / result['total_bytes'], 'headline share'
expected_largest = sorted(regions, key=lambda entry: (-entry[0], entry[1], entry[2]))[:10]
actual_largest = [(entry['bytes'], entry['file'], entry['start'], entry['end']) for entry in result['largest_regions']]
assert actual_largest == expected_largest, 'top ten sizes and ranking'
for entry in result['largest_regions']:
    data = Path(result['provenance']['compiler_root']).joinpath(entry['file']).read_bytes()
    assert entry['start_line'] == data[:entry['start']].count(b'\n') + 1, 'top ten start line'
    assert entry['end_line'] == data[:entry['end']-1].count(b'\n') + 1, 'top ten end line'
    assert entry['causes'], 'top ten reason missing'
print(f'PASS: independent byte-mask oracle; {hidden_total} hidden = {blocked_total} union - {removed_total} independently examined; {len(stock)} files')
