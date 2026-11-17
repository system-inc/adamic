#!/usr/bin/env python3
"""#rv75m9k: pairs one file's latent census before and after road A, the way the receiver adaptations' proofs did.

Each record is keyed by (kind, phase, reason, text) with source positions stripped, as a multiset. Pairs drop out; what
is left is printed as removed (before only) or added (after only), with its position from its own run. A census that
didn't finish (no trailer) or reported checker diagnostics is said so, never read as clean.
Usage: compare.py <file> <before.jsonl> <after.jsonl>
"""
import collections
import json
import re
import sys

position = re.compile(r'(\.ts)?:\d+:\d+')
# A census names its tree's absolute path in every message; two trees' runs pair only once that's cut to src/.
tree_path = re.compile(r'/\S*?/(?=src/)')


def findings(value, out):
    """Every finding nested anywhere in a census row: a dict carrying a phase or a reason."""
    if isinstance(value, dict):
        if 'phase' in value or 'reason' in value:
            out.append(value)
        for child in value.values():
            findings(child, out)
    elif isinstance(value, list):
        for child in value:
            findings(child, out)


def load(path):
    rows, records = [], []
    with open(path) as stream:
        for line in stream:
            line = line.strip()
            if line:
                rows.append(json.loads(line))
    if not rows or rows[-1].get('status') != 'relation_report':
        print(f'note {path}: no relation_report trailer, so this census did not finish')
    for row in rows[1:-1]:
        findings(row, records)
    return records


def key(record):
    fields = []
    for name in ('kind', 'phase', 'reason', 'text', 'what', 'message'):
        value = record.get(name)
        if isinstance(value, str):
            fields.append(position.sub('', tree_path.sub('', value)))
    return tuple(fields)


def where(record):
    for name in ('where', 'site', 'position'):
        if isinstance(record.get(name), str):
            return record[name]
    return '?'


file, before_path, after_path = sys.argv[1:]
before, after = load(before_path), load(after_path)
for name, records in (('before', before), ('after', after)):
    kinds = collections.Counter(record.get('kind', '?') for record in records)
    print(f'census {file} {name}: {len(records)} records {dict(kinds)}')
before_keys = collections.Counter(key(record) for record in before)
after_keys = collections.Counter(key(record) for record in after)
removed = before_keys - after_keys
added = after_keys - before_keys
print(f'census {file}: paired {sum((before_keys & after_keys).values())}, removed {sum(removed.values())}, added {sum(added.values())}')
shown = collections.Counter()
for record in before:
    k = key(record)
    if shown[k] < removed[k]:
        shown[k] += 1
        print(f'removed {file} {where(record)} {" | ".join(k)[:300]}')
shown = collections.Counter()
for record in after:
    k = key(record)
    if shown[k] < added[k]:
        shown[k] += 1
        print(f'added {file} {where(record)} {" | ".join(k)[:300]}')
if before:
    print(f'sample {file}: {json.dumps(before[0])[:400]}')
