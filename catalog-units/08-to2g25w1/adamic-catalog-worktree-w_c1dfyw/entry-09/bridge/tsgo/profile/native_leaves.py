"""Resolve unknown native leaf PCs in go tool pprof -raw output using nm.

Usage: native_leaves.py executable raw-profile.log. This is leaf CPU sampling,
not an inclusive native stack profile; concurrent Go CPU is reported separately.
"""
from pathlib import Path
import bisect
from collections import Counter
import re
import subprocess
import sys

binary = str(Path(sys.argv[1]).resolve())
text = Path(sys.argv[2]).read_text()
match = re.search(r'(\d+): (0x\w+)/(0x\w+)/(0x\w+) ' + re.escape(binary) + r' ', text.split('Mappings\n')[1])
assert match, 'Executable mapping missing'
mapping = int(match[1])
base, stop, offset = [int(x, 16) for x in match.groups()[1:]]
symbols = []
for line in subprocess.check_output(['nm', '-n', binary], text=True).splitlines():
    match = re.match(r'([0-9a-f]+) [tT] (.+)', line)
    if match:
        symbols.append((int(match[1], 16), match[2]))
starts = [s[0] for s in symbols]
locations = {}
for line in text.split('Locations\n')[1].split('Mappings\n')[0].splitlines():
    match = re.match(r'\s*(\d+): (0x\w+) M=' + str(mapping) + r'  :', line)
    if match:
        pc = int(match[2], 16) - base + offset
        index = bisect.bisect_right(starts, pc) - 1
        assert index >= 0
        locations[int(match[1])] = symbols[index][1]
counts = Counter()
for line in text.split('Samples:\n')[1].split('Locations\n')[0].splitlines():
    match = re.match(r'\s*(\d+)\s+(\d+): (\d+)', line)
    if match and int(match[3]) in locations:
        counts[locations[int(match[3])]] += int(match[2])
for name, nanoseconds in counts.most_common():
    print(f'{nanoseconds / 1e9:.2f}s {name}')
