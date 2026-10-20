#!/usr/bin/env python3
"""Compare full latent streams, mapping annotation-shifted UTF-16 positions back."""
from array import array
from bisect import bisect_right
import gzip
import json
from pathlib import Path
import sys

before_tree, after_tree, plan_path, before_path, after_path, output = sys.argv[1:]
plan = json.loads(Path(plan_path).read_text())
cache = {}
modes = []

def source(tree, file):
    key = tree, file
    if key not in cache:
        text = (Path(tree) / file).read_bytes().decode()
        units = array('H', text.encode('utf-16-le'))
        cache[key] = (units, [0] + [i + 1 for i, c in enumerate(units) if c == 10])
    return cache[key]

def mapped(file, line, column, after):
    if not after:
        return line, column
    units, starts = source(after_tree, file)
    position = starts[line - 1] + column - 1
    delta = 0
    original = None
    for e in sorted((e for e in plan['edits'] if e['file'] == file), key=lambda e:e['start']):
        start = e['start'] + delta
        length = len(e['after'].encode('utf-16-le')) // 2
        if position < start:
            break
        if position <= start + length:
            original = e['start'] + min(position - start, e['end'] - e['start'])
            break
        delta += length - (e['end'] - e['start'])
    position = position - delta if original is None else original
    _, starts = source(before_tree, file)
    row = bisect_right(starts, position) - 1
    return row + 1, position - starts[row] + 1

def read(path, after=False):
    opener = gzip.open if path.endswith('.gz') else open
    found = {}
    with opener(path, 'rt') as stream:
        header = json.loads(next(stream))
        assert header['latent_mode'] in ('full', 'void-widening-only-source-scan')
        modes.append(header['latent_mode'])
        for raw in stream:
            record = json.loads(raw)
            for f in (record.get('findings') or []):
                if f['kind'] not in ('Refused', 'NotYet'):
                    continue
                filename, line, column = f['where'].rsplit(':', 2)
                if '/src/' not in filename:
                    continue
                file = 'src/' + filename.split('/src/', 1)[1]
                line, column = mapped(file, int(line), int(column), after)
                row = {'file': file, 'line': line, 'column': column, 'kind': f['kind'], 'reason': f['reason']}
                found[json.dumps(row, sort_keys=True)] = row
    return found

before, after = read(before_path), read(after_path, True)
assert len(set(modes)) == 1, modes
removed = [before[k] for k in before.keys() - after.keys()]
added = [after[k] for k in after.keys() - before.keys()]
void = lambda r:r['reason'].startswith(('a void value at a return site whose function returns ', 'a void-returning function at the '))
result = {'measurementMode':modes[0], 'before':len(before), 'after':len(after), 'new':added, 'removed':removed,
          'voidBefore':len({(r['file'],r['line'],r['column']) for r in before.values() if void(r)}),
          'voidAfter':len({(r['file'],r['line'],r['column']) for r in after.values() if void(r)}),
          'voidRemoved':len({(r['file'],r['line'],r['column']) for r in removed if void(r)}),
          'newRefused':[r for r in added if r['kind']=='Refused']}
Path(output).write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in ('new','removed','newRefused')} | {'new':len(added), 'newRefused':len(result['newRefused'])}))
