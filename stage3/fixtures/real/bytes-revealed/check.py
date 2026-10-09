#!/usr/bin/env python3
"""Measure this frozen unit with ed6e2975's unchanged fixture harness."""
import argparse
import copy
import json
from pathlib import Path
import re
import shutil
import subprocess
import tempfile

PIN = 'ed6e29751ee47d86fad450cd1674139883bc0f70'
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('compiler_tree', type=Path, help='checkout at the pinned compiler, with cohere initialized')
parser.add_argument('logs', type=Path)
args = parser.parse_args()
tree = args.compiler_tree.resolve()
root = Path(__file__).resolve().parent
logs = args.logs.resolve()
logs.mkdir(parents=True, exist_ok=True)
head = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=tree, text=True).strip()
assert head == PIN, f'expected compiler {PIN}, found {head}'
assert not subprocess.check_output(['git', 'diff', 'HEAD', '--', 'cmd', 'internal', 'stage3/fixtures', 'oracle'], cwd=tree), 'compiler or harness has local edits'
entries = json.loads((root / 'status.json').read_text())
assert len(entries) == 10 and len({entry['reason'] for entry in entries}) == 10
for entry in entries:
    path = root / entry['file']
    header = path.read_text().splitlines()[0]
    expected = entry['a_check']
    assert header == '// a-check: ' + expected, f'wrong a-check header: {path}'
    assert entry['stage0']['outcome'] in ('NotYet', 'Refused')

observations = []


def run(name, mutation=None):
    # The unchanged harness discovers one-level buckets. A scratch real bucket
    # nests this new unit at its actual relative path, without including f9's
    # separate ten-entry manifest. Its diagnostic path normalization is unchanged.
    with tempfile.TemporaryDirectory(prefix='adamic-real-bytes-') as scratch:
        fixture_root = Path(scratch)
        bucket = fixture_root / 'real'
        copied = bucket / 'bytes-revealed'
        working = copy.deepcopy(entries)
        for entry in working:
            relative = Path(entry['file'])
            assert not relative.is_absolute() and '..' not in relative.parts
            shutil.copytree(root / relative.parent, copied / relative.parent)
        first = working[0]
        if mutation == 'output-byte':
            golden = copied / Path(first['file']).parent / 'expected.stdout'
            before = golden.read_bytes()
            after = bytes([ord('X') if before[0] != ord('X') else ord('Y')]) + before[1:]
            assert len(before) == len(after) and sum(a != b for a, b in zip(before, after)) == 1
            golden.write_bytes(after)
        for entry in working:
            # expected.stdout is the authoritative Node golden; the JSON records
            # the same observed bytes for convenient inspection.
            entry['node']['stdout'] = (copied / Path(entry['file']).parent / 'expected.stdout').read_text()
            entry['file'] = 'bytes-revealed/' + entry['file']
        if mutation == 'diagnostic-byte':
            before = first['stage0']['what']
            index = before.index('__String')
            first['stage0']['what'] = before[:index] + 'X' + before[index + 1:]
            assert sum(a != b for a, b in zip(before.encode(), first['stage0']['what'].encode())) == 1
        bucket.mkdir(parents=True, exist_ok=True)
        (bucket / 'status.json').write_text(json.dumps(working, indent=2) + '\n')
        selector = '^TestFixtures$/^real$'
        if mutation:
            selector += '/^bytes-revealed$/^escaped-string$/^main.a$'
        command = ['go', 'test', './stage3/fixtures', '-run', selector, '-count=1', '-timeout', '10m', '-v', '-args', '-fixtures', str(fixture_root)]
        with (logs / (name + '.log')).open('w') as output:
            result = subprocess.run(command, cwd=tree, stdout=output, stderr=subprocess.STDOUT)
        output = (logs / (name + '.log')).read_text()
        return result.returncode, output, command


code, output, command = run('fixtures')
assert code == 0, f'fixture measurement failed: {logs / "fixtures.log"}'
leaves = re.findall(r'--- PASS: .*?/(node|stage0) \(', output)
assert leaves.count('node') == 10 and leaves.count('stage0') == 10, 'all twenty leaf checks must run'
observations.append({'check': 'fixtures', 'exit': code, 'node_passes': 10, 'diagnostic_passes': 10, 'command': command})
for name, failure, independent in [('output-byte', 'recorded Node byte comparison failed', 'stage0'), ('diagnostic-byte', 'gap changed:', 'node')]:
    code, output, command = run(name, name)
    assert code != 0 and failure in output, f'{name} survived or failed elsewhere'
    leaves = re.findall(r'--- (PASS|FAIL): .*?/(node|stage0) \(', output)
    assert leaves.count(('PASS', independent)) == 1 and sum(state == 'FAIL' for state, _ in leaves) == 1, 'mutant must fail only its intended leaf'
    observations.append({'check': name, 'exit': code, 'bytes_changed': 1, 'caught_by': failure, 'independent_check_passed': independent, 'command': command})
(logs / 'results.json').write_text(json.dumps({'compiler_commit': PIN, 'checks': observations}, indent=2) + '\n')
print('PASS: 10 fixtures, 20 Node/full-diagnostic comparisons; output-byte and diagnostic-byte mutants caught independently')
