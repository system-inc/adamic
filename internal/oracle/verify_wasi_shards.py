#!/usr/bin/env python3
"""Verify successful fixture completion, not merely the declared shard union.

Usage: python3 internal/oracle/verify_wasi_shards.py manifest.json shard-*.json
Generate manifest.json with:
  go test ./internal/oracle -run '^TestWASIShardUnion$' -count=1 -timeout 90s -json
Each shard log must come from an isolated gate invocation, for example:
  ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
    -run '^TestWASIAgreesWithNode$/^shard-000$' -count=1 -parallel 4 -timeout 90s -json
A skipped, absent, deadline-killed, or unsuccessful shard cannot complete the gate.
"""
import gzip
import json
import sys
from collections import Counter

PREFIX = 'TestWASIAgreesWithNode/'


def events(path):
    opener = gzip.open if str(path).endswith('.gz') else open
    with opener(path, 'rt') as stream:
        return [json.loads(line) for line in stream if line.strip()]


def manifest(log):
    if not any(e.get('Action') == 'pass' and 'Test' not in e for e in log):
        raise ValueError('missing successful manifest package completion')
    names = []
    passed = False
    for event in log:
        if event.get('Test') != 'TestWASIShardUnion':
            continue
        passed |= event.get('Action') == 'pass'
        output = event.get('Output', '')
        marker = 'wasi manifest: '
        if marker in output:
            shard, path = output.split(marker, 1)[1].strip().split(' ', 1)
            names.append(PREFIX + shard + '/' + path)
    if not passed or not names or len(names) != len(set(names)):
        raise ValueError('missing or invalid manifest union proof')
    return set(names)


def verify(expected, logs):
    completed, started, shards = Counter(), Counter(), Counter()
    expected_shards = {'/'.join(name.split('/')[:2]) for name in expected}
    for log in logs:
        if not any(e.get('Action') == 'pass' and 'Test' not in e for e in log):
            raise ValueError('missing successful package completion (failed/deadline-killed/truncated log)')
        for event in log:
            name, action = event.get('Test', ''), event.get('Action')
            if name in expected_shards and action == 'pass':
                shards[name] += 1
            if name.startswith(PREFIX) and name.count('/') >= 2:
                if name not in expected:
                    raise ValueError('unexpected fixture ' + name)
                if action == 'run':
                    started[name] += 1
                if action == 'pass':
                    completed[name] += 1
    for name in sorted(expected_shards):
        if not shards[name]:
            raise ValueError('missing shard ' + name)
        if shards[name] != 1:
            raise ValueError('repeated shard ' + name)
    for name in sorted(expected):
        if not completed[name]:
            raise ValueError('missing successful fixture ' + name)
        if completed[name] != 1 or started[name] != 1:
            raise ValueError('repeated or unstarted fixture ' + name)
    return f'union completed: {len(expected)} fixtures exactly once; {len(expected_shards)} shards'


def self_test():
    names = {PREFIX + 'shard-000/a', PREFIX + 'shard-001/b'}
    def log(shard, path):
        name = PREFIX + shard + '/' + path
        return [{'Test': name, 'Action': 'run'}, {'Test': name, 'Action': 'pass'},
                {'Test': PREFIX + shard, 'Action': 'pass'}, {'Action': 'pass'}]
    a, b = log('shard-000', 'a'), log('shard-001', 'b')
    verify(names, [a, b])
    for bad in ([a], [a, b, a], [a, b[:-1]], [a, [{**e, 'Action': 'skip'} if e.get('Test', '').endswith('/b') else e for e in b]]):
        try:
            verify(names, bad)
        except ValueError:
            continue
        raise AssertionError('missing/repeated/skipped/deadline-killed completion accepted')
    print('completion proof: missing, repeated, skipped, and deadline-killed logs rejected')


if __name__ == '__main__':
    try:
        if sys.argv[1:] == ['--self-test']:
            self_test()
        else:
            print(verify(manifest(events(sys.argv[1])), [events(p) for p in sys.argv[2:]]))
    except (ValueError, OSError, IndexError) as error:
        sys.exit(str(error))
