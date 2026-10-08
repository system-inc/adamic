#!/usr/bin/env python3
"""Account perf self samples by innermost DWARF source, including ThinLTO inlines."""
import argparse
import collections
import json
from pathlib import Path
import re
import subprocess

p = argparse.ArgumentParser()
p.add_argument('data')
p.add_argument('binary')
p.add_argument('output')
p.add_argument('--perf', required=True)
p.add_argument('--llvm', required=True)
a = p.parse_args()
binary = str(Path(a.binary).resolve())
nm = subprocess.check_output([a.llvm + '/llvm-nm', '--defined-only', binary], text=True)
symbols = {}
for line in nm.splitlines():
    fields = line.split()
    if len(fields) == 3:
        symbols[fields[2]] = int(fields[0], 16)
raw = subprocess.check_output([a.perf, 'script', '-i', a.data, '-G', '-F',
                               'ip,sym,symoff,dso,period'], text=True)
rows = []
biases = collections.Counter()
for line in raw.splitlines():
    m = re.fullmatch(r'\s*(\d+)\s+([0-9a-f]+)\s+(.+?)\s+\((.+)\)', line)
    if not m:
        raise RuntimeError('unparsed perf sample: ' + line)
    period, ip, symbol, dso = m.groups()
    ip = int(ip, 16)
    rows.append((int(period), ip, symbol, dso))
    s = re.fullmatch(r'(.+)\+0x([0-9a-f]+)', symbol)
    if dso == binary and s and s[1] in symbols:
        biases[ip - symbols[s[1]] - int(s[2], 16)] += 1
if len(biases) != 1:
    raise RuntimeError(f'inconsistent executable relocation: {biases}')
bias = next(iter(biases))
addresses = sorted({ip - bias for _, ip, _, dso in rows if dso == binary})
text = subprocess.check_output([a.llvm + '/llvm-symbolizer', '--obj=' + binary,
                                '--output-style=JSON'],
                               input=''.join(f'0x{ip:x}\n' for ip in addresses), text=True)
frames = {int(row['Address'], 16): row['Symbol'] for row in map(json.loads, text.splitlines())}
shares = collections.Counter()
functions = collections.Counter()
locations = collections.Counter()
samples = collections.Counter()
for period, ip, symbol, dso in rows:
    if dso != binary:
        group, function, file, line = 'external', symbol.split('+0x')[0], dso, 0
    else:
        stack = frames[ip - bias]
        leaf = stack[0] if stack else {}
        file, function, line = leaf.get('FileName', ''), leaf.get('FunctionName', symbol), leaf.get('Line', 0)
        if '/adamic/runtime/' in file:
            group = 'runtime'
        elif Path(file).name == 'main.c':
            group = 'generated'
        else:
            group = 'unresolved'
    shares[group] += period
    functions[group, function] += period
    locations[group, function, Path(file).name, line] += period
    samples[group] += 1
assert sum(shares.values()) == sum(row[0] for row in rows)
total = sum(shares.values())
result = {
    'samples': len(rows), 'period_total': total, 'relocation_bias': hex(bias),
    'split': {group: {'percent': 100 * value / total, 'samples': samples[group], 'period': value}
              for group, value in shares.most_common()},
    'functions': [{'group': group, 'function': fn, 'percent': 100 * value / total,
                   'period': value} for (group, fn), value in functions.most_common()],
    'locations': [{'group': group, 'function': fn, 'file': file, 'line': line,
                   'percent': 100 * value / total}
                  for (group, fn, file, line), value in locations.most_common()],
}
Path(a.output).write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result['split'], indent=2))
for row in result['functions'][:25]:
    print(f"{row['percent']:6.2f}% {row['group']:10} {row['function']}")
