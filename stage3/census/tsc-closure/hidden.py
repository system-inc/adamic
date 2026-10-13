"""UTF-8 body accounting with an exclusive deferred-generic bucket."""
import gzip
import json
from pathlib import Path
import re
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / 'latent'))
from unit_status import DEFERRED, SKIPPED, attempted, deferred_summary


def read_json(path):
    with (gzip.open(path, 'rt') if Path(path).suffix == '.gz' else Path(path).open()) as stream:
        return json.load(stream)


def read_rows(path):
    with (gzip.open(path, 'rt') if Path(path).suffix == '.gz' else Path(path).open()) as stream:
        return [json.loads(line) for line in stream]


def calculate(rows, stock, root):
    def location(where):
        name, line, column = where.rsplit(':', 2)
        path = Path(name)
        name = path.relative_to(root).as_posix() if path.is_relative_to(root) else name
        return name, f'{name}:{line}:{column}'
    blocked = {n: bytearray(e['bytes']) for n, e in stock.items()}
    covered = {n: bytearray(e['bytes']) for n, e in stock.items()}
    deferred = {n: bytearray(e['bytes']) for n, e in stock.items()}
    failures, skips = {}, {}
    for row in rows[1:]:
        for unit in row['units']:
            name, _ = location(unit['where'])
            if name not in stock:
                continue
            status = unit['status']
            attempted(unit)  # Reject unknown statuses rather than crediting them.
            if status in SKIPPED | {DEFERRED}:
                start, end = unit.get('body_start', 0), unit.get('body_end', 0)
                assert 0 <= start <= end <= stock[name]['bytes'], 'unit body span'
                target = deferred if status == DEFERRED else blocked
                target[name][start:end] = b'\x01' * (end-start)
                skips.setdefault(name, []).append((start, end))
        for finding in row['findings']:
            if finding['kind'] not in ('Boundary', 'SkippedDependency'):
                continue
            name, where = location(finding['where'])
            if name not in stock:
                continue
            if finding['kind'] == 'Boundary':
                start, end = finding.get('start', 0), finding['end']
                failures.setdefault((name, finding['unit']), []).append((start, end))
            else:
                body = next(b for b in stock[name]['bodies'] if b['where'] == where)
                start, end = body['start'], body['end']
                skips.setdefault(name, []).append((start, end))
            blocked[name][start:end] = b'\x01' * (end-start)
    for row in rows[1:]:
        for unit in row['units']:
            if not attempted(unit):
                continue
            name, where = location(unit['where'])
            if name not in stock:
                continue
            external = next(u for u in stock[name]['units'] if u['where'] == where)
            start, end = external['start'], external['end']
            mask = bytearray(b'\x01' * (end-start))
            for left, right in failures.get((name, unit['where']), []) + skips.get(name, []):
                left, right = max(start, left), min(end, right)
                if left < right:
                    mask[left-start:right-start] = b'\x00' * (right-left)
            for index, value in enumerate(mask, start):
                covered[name][index] |= value
    files, regions = {}, []
    for name, entry in stock.items():
        for index, value in enumerate(deferred[name]):
            if value:
                blocked[name][index] = covered[name][index] = 0
        hidden = bytes(a and not b for a, b in zip(blocked[name], covered[name]))
        ranges = [[m.start(), m.end()] for m in re.finditer(b'\x01+', hidden)]
        files[name] = dict(bytes=entry['bytes'], hidden_bytes=sum(hidden),
                           hidden_ranges=ranges, hidden_share=sum(hidden)/entry['bytes'] if entry['bytes'] else 0,
                           blocked_union_bytes=sum(blocked[name]),
                           independently_examined_bytes=sum(blocked[name])-sum(hidden),
                           deferred_body_bytes=sum(deferred[name]))
        for start, end in ranges:
            regions.append(dict(file=name, start=start, end=end, bytes=end-start,
                                causes=['checker body or lowering recovery boundary']))
    result = dict(files=files, total_bytes=sum(e['bytes'] for e in stock.values()),
                  largest_regions=sorted(regions, key=lambda r: (-r['bytes'], r['file'], r['start']))[:10],
                  **deferred_summary([u for row in rows[1:] for u in row['units']]))
    for key in ('hidden_bytes', 'blocked_union_bytes', 'independently_examined_bytes', 'deferred_body_bytes'):
        result[key] = sum(f[key] for f in files.values())
    result['provenance'] = dict(compiler_root=str(root))
    for region in result['largest_regions']:
        data = (root/region['file']).read_bytes()
        region['start_line'] = data[:region['start']].count(b'\n') + 1
        region['end_line'] = data[:region['end']-1].count(b'\n') + 1
    result['hidden_share'] = result['hidden_bytes']/result['total_bytes'] if result['total_bytes'] else 0
    return result
