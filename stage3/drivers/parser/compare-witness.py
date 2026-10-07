#!/usr/bin/env python3
"""Compare the available projection, preserving every unequal node record."""
import difflib
import gzip
import hashlib
import json
from pathlib import Path
import sys

node, witness, directory = map(Path, sys.argv[1:])
directory.mkdir(parents=True, exist_ok=True)

def sections(path, full):
    result = {}
    name = None
    selected = False
    for line in path.read_text().splitlines():
        if line.startswith('file\t'):
            name = line.split('\t', 1)[1]
            selected = name.endswith('.ts')
            if selected:
                if name in result: raise RuntimeError('duplicate file')
                result[name] = []
        elif selected and not line.startswith(('diagnostic ', 'diagnostics ')):
            if full:
                metadata, text = line.split('\t', 1)
                kind, pos, end, flags = metadata.split()
                line = f'{kind} {pos} {end}\t{text if kind in ["Identifier", "PrivateIdentifier"] else ""}'
            result[name].append(line)
    return result

expected, actual = sections(node, True), sections(witness, False)
assert list(expected) == list(actual), 'corpus file population/order mismatch'
projection = ''.join('file\t' + name + '\n' + ''.join(line + '\n' for line in rows) for name, rows in expected.items()).encode()
(directory / 'node-projection.dump').write_bytes(projection)
changes = []
files = []
for name in expected:
    a, b = expected[name], actual[name]
    hunks = []
    for tag, first, last, other_first, other_last in difflib.SequenceMatcher(None, a, b, autojunk=False).get_opcodes():
        if tag == 'equal': continue
        hunk = {'file': name, 'operation': tag, 'nodeOrdinals': [first + 1, last], 'witnessOrdinals': [other_first + 1, other_last], 'createSourceFile': a[first:last], 'stage1': b[other_first:other_last]}
        hunks.append(hunk)
    if hunks:
        changes.extend(hunks)
        files.append({'file': name, 'hunks': len(hunks), 'createSourceFileNodes': len(a), 'stage1Nodes': len(b), 'expectedUnequalRecords': sum(len(h['createSourceFile']) for h in hunks), 'witnessUnequalRecords': sum(len(h['stage1']) for h in hunks)})
raw = (json.dumps(changes, indent=2) + '\n').encode()
(directory / 'every-difference.json.gz').write_bytes(gzip.compress(raw, mtime=0))
summary = {'fields': ['kind', 'pos', 'end', 'identifier/private-identifier text'], 'positions': 'UTF-16 code units, directly from both node tables', 'kindNormalization': {'stage1 EndOfFile': 'EndOfFileToken'}, 'excluded': ['node flags', 'parse diagnostics', 'literal text', 'three JSON corpus files'], 'filesCompared': len(expected), 'filesDifferent': len(files), 'nodeCount': sum(map(len, expected.values())), 'witnessNodeCount': sum(map(len, actual.values())), 'differenceHunks': len(changes), 'files': files, 'nodeProjectionBytes': len(projection), 'nodeProjectionSha256': hashlib.sha256(projection).hexdigest(), 'witnessBytes': witness.stat().st_size, 'witnessSha256': hashlib.sha256(witness.read_bytes()).hexdigest(), 'differencesUncompressedBytes': len(raw), 'differencesSha256': hashlib.sha256(raw).hexdigest()}
(directory / 'report.json').write_text(json.dumps(summary, indent=2)+'\n')
print(json.dumps(summary, indent=2))
