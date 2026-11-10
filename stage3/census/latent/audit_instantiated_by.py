"""Prove that instantiated_by is the only byte change in a census stream."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path


def read(path):
    data = Path(path).read_bytes()
    return gzip.decompress(data) if str(path).endswith('.gz') else data


def strip_witnesses(data):
    text = data.decode('utf-8')
    token = ',"instantiated_by":'
    decoder = json.JSONDecoder()
    count = 0
    while (start := text.find(token)) >= 0:
        value_start = start + len(token)
        _, end = decoder.raw_decode(text, value_start)
        text = text[:start] + text[end:]
        count += 1
    return text.encode('utf-8'), count


def compare(base, head):
    before, after = read(base), read(head)
    stripped, count = strip_witnesses(after)
    assert stripped == before, 'bytes differ after removing instantiated_by'
    for line in after.splitlines():
        row = json.loads(line)
        for unit in row.get('units', []):
            witnesses = unit.get('instantiated_by', [])
            assert len(witnesses) <= 5
            # The entry driver re-encodes maps alphabetically. Restore the witness
            # struct field order used by Go's ordering key before comparing.
            keys = [json.dumps({k: w[k] for k in ('unit', 'call_site', 'type_arguments')}, ensure_ascii=False, separators=(',', ':')).replace('&', r'\u0026').replace('<', r'\u003c').replace('>', r'\u003e').replace('\u2028', r'\u2028').replace('\u2029', r'\u2029') for w in witnesses]
            assert keys == sorted(set(keys)), 'witness order or deduplication'
            if unit['status'] == 'deferred_generic_instantiation':
                assert bool(witnesses) == unit['instantiated'], 'witness/boolean mismatch'
            else:
                assert not witnesses, 'witnesses attached outside deferral inventory'
    return dict(base_sha256=hashlib.sha256(before).hexdigest(),
                stripped_sha256=hashlib.sha256(stripped).hexdigest(),
                head_sha256=hashlib.sha256(after).hexdigest(),
                units_with_witnesses=count, byte_identical_except_instantiated_by=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('base', type=Path)
    parser.add_argument('head', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    result = compare(args.base, args.head)
    args.output.write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps(result))
