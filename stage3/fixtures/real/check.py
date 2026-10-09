#!/usr/bin/env python3
"""Run the shared fixture checks and isolated byte/diagnostic mutants without editing the checkout."""
import argparse
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('logs', type=Path, help='directory for complete test logs')
args = parser.parse_args()
repository = Path(__file__).resolve().parents[3]
root = Path(__file__).resolve().parent
logs = args.logs.resolve()
logs.mkdir(parents=True, exist_ok=True)


def run(name, selector, fixtures=None):
    command = ['go', 'test', './stage3/fixtures', '-run', selector, '-count=1', '-timeout', '10m', '-v']
    if fixtures is not None:
        command += ['-args', '-fixtures', str(fixtures)]
    path = logs / (name + '.log')
    with path.open('w') as output:
        result = subprocess.run(command, cwd=repository, stdout=output, stderr=subprocess.STDOUT)
    return result.returncode, path.read_text(), command


observations = []
code, output, command = run('fixtures', '^TestFixtures$/^real$')
assert code == 0, f'fixture checks failed; see {logs / "fixtures.log"}'
observations.append({'check': 'fixtures', 'exit': code, 'command': command})
entries = json.loads((root / 'status.json').read_text())
assert len(entries) == 10, f'expected ten fixtures, found {len(entries)}'
for name, field, expected_failure, expected_pass in [
    ('output-byte', 'node', 'recorded Node byte comparison failed', '/stage0'),
    ('diagnostic-byte', 'stage0', 'gap changed:', '/node'),
]:
    with tempfile.TemporaryDirectory(prefix='adamic-real-' + name + '-') as scratch:
        fixture_root = Path(scratch)
        copied = fixture_root / 'real'
        shutil.copytree(root, copied)
        entries = json.loads((copied / 'status.json').read_text())
        # Use the first actual fixture. Mutate precisely one ASCII byte of its golden,
        # preserving its source, the other golden, and every other entry.
        entry = entries[0]
        key = 'stdout' if field == 'node' else 'what'
        original = entry[field][key]
        index = 0 if field == 'node' else original.index('predicate', original.index('Adamic 0.1 refuses'))
        replacement = 'X' if original[index] != 'X' else 'Y'
        entry[field][key] = original[:index] + replacement + original[index + 1:]
        before, after = original.encode(), entry[field][key].encode()
        assert len(before) == len(after) and sum(a != b for a, b in zip(before, after)) == 1
        (copied / 'status.json').write_text(json.dumps(entries, indent=2) + '\n')
        code, output, command = run(name, '^TestFixtures$/^real$/^type-predicate$/^main.a$', fixture_root)
        assert code != 0 and expected_failure in output, f'{name} survived or failed elsewhere; see {logs / (name + ".log")}'
        assert any('--- PASS:' in line and expected_pass in line for line in output.splitlines()), f'{name} also broke the independent check'
        assert sum('--- FAIL:' in line and (line.rstrip().endswith('/node (0.00s)') or line.rstrip().endswith('/stage0 (0.00s)')) for line in output.splitlines()) == 1, 'mutant must fail only its intended leaf check'
        observations.append({'check': name, 'exit': code, 'bytes_changed': 1, 'caught_by': expected_failure, 'independent_check_passed': expected_pass, 'command': command})
(logs / 'results.json').write_text(json.dumps(observations, indent=2) + '\n')
print('PASS: 10 fixtures, 20 Node/diagnostic checks; output-byte and diagnostic-byte mutants caught independently')
