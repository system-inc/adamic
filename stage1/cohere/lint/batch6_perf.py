#!/usr/bin/env python3
"""Record native Linux perf CPU-clock stacks and summarize sampled costs."""
import argparse
import collections
import hashlib
import json
from pathlib import Path
import re
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--perf', default='perf')
parser.add_argument('--binary', required=True, type=Path)
parser.add_argument('--manifest', required=True, type=Path)
parser.add_argument('--output', required=True, type=Path)
arguments = parser.parse_args()
arguments.output.mkdir(parents=True, exist_ok=True)
output = arguments.output

def run(command, destination, errors):
    with destination.open('wb') as stdout, errors.open('wb') as stderr:
        subprocess.run(command, stdout=stdout, stderr=stderr, check=True)

run([arguments.perf, 'record', '-e', 'cpu-clock:u', '-F', '997', '--call-graph', 'fp',
     '-o', str(output / 'perf.data'), '--', str(arguments.binary), '--manifest',
     str(arguments.manifest), '--count'], output / 'findings.txt', output / 'record.log')
for name, mode in [('self', '--no-children'), ('inclusive', '--children')]:
    run([arguments.perf, 'report', '--stdio', mode, '--call-graph', 'none', '--sort',
         'symbol', '--percent-limit', '0.5', '-i', str(output / 'perf.data')],
        output / (name + '.txt'), output / (name + '.log'))
run([arguments.perf, 'script', '-i', str(output / 'perf.data')],
    output / 'stacks.txt', output / 'script.log')
self_costs = collections.Counter()
dispatch_costs = collections.Counter()
dispatch = 0
samples = 0
for record in (output / 'stacks.txt').read_text().strip().split('\n\n'):
    lines = record.splitlines()
    match = re.search(r':\s+(\d+) cpu-clock', lines[0])
    if not match:
        continue
    symbols = [re.split(r'\+0x', line.split()[1])[0]
               for line in lines[1:] if len(line.split()) > 1]
    if not symbols:
        continue
    weight = int(match.group(1))
    samples += 1
    self_costs[symbols[0]] += weight
    if any('visitRules' in symbol for symbol in symbols):
        dispatch += weight
        dispatch_costs[symbols[0]] += weight
period = sum(self_costs.values())
summary = {
    'event': 'cpu-clock:u', 'frequency': 997,
    'binary_sha256': hashlib.sha256(arguments.binary.read_bytes()).hexdigest(),
    'manifest_sha256': hashlib.sha256(arguments.manifest.read_bytes()).hexdigest(),
    'samples': samples, 'sampled_cpu_seconds_approximate': period / 1e9,
    'dispatch_frame_inclusive_percent': 100 * dispatch / period if dispatch else None,
    'dispatch_frame_observed': bool(dispatch),
    'dispatch_note': 'Missing frames can be inlined; a missing frame is not zero cost.',
    'top_self': [{'symbol': symbol, 'self_percent': 100 * weight / period,
                  'cpu_seconds_approximate': weight / 1e9,
                  'self_under_dispatch_frame_percent': 100 * dispatch_costs[symbol] / period if dispatch else None}
                 for symbol, weight in self_costs.most_common(20)],
}
(output / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary, indent=2))
