#!/usr/bin/env python3
"""Compare actual checked inputs with the base, then kill six isolated mutants using Go overlays.
Usage: python3 review/split-fuzz/prove.py OUTPUT_DIRECTORY
Source the cloud setup environment first. No repository sources are modified.
"""
import collections
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
out = Path(sys.argv[1]).resolve()
out.mkdir(parents=True, exist_ok=True)
base = '54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8'
targets = {
    'TestGeneratedProgramsCheckAndLower': ('fuzz_test.go', 1, 60, 13, 'seeds-011-015'),
    'TestRegexProgramsPassTheChecker': ('fuzz_test.go', 1, 30, 18, 'seeds-016-020'),
    'TestUndefinedNumbersShapes': ('undefined_numbers_test.go', 1, 40, 28, 'seeds-026-030'),
    'TestOverridesShapesAndLower': ('overrides_test.go', 0, 49, 43, 'seeds-040-044'),
}
paths = list(dict.fromkeys('internal/fuzz/' + row[0] for row in targets.values()))
current = {path: (root / path).read_text() for path in paths}
original = {path: subprocess.check_output(['git', 'show', base + ':' + path], cwd=root).decode() for path in paths}

def run(label, sources, parent=None):
    directory = out / label
    directory.mkdir(exist_ok=True)
    replacements = {}
    for path, source in sources.items():
        saved = directory / Path(path).name
        saved.write_text(source)
        replacements[str(root / path)] = str(saved)
    overlay = directory / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': replacements}, indent=2) + '\n')
    pattern = '^(' + '|'.join(targets) + ')$' if parent is None else '^' + parent + '$'
    log = out / (label + '.jsonl')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), '-json', '-count=1',
                                 '-parallel=4', '-timeout=10m', '-run=' + pattern, './internal/fuzz'],
                                cwd=root, env=dict(os.environ, GOMAXPROCS='4', ADAMIC_GATE_UNCACHED='1'),
                                stdout=output, stderr=subprocess.STDOUT)
    events = [json.loads(line) for line in log.read_text().splitlines() if line.startswith('{')]
    return result.returncode, events

def instrument(sources):
    answer = {}
    for path, source in sources.items():
        source = source.replace('import (', 'import (\n\t"crypto/sha256"', 1)
        needle = 'if err := os.WriteFile(path, []byte(source), 0o644); err != nil {'
        source = source.replace(needle, 't.Logf("split-case seed=%d sha256=%x", seed, sha256.Sum256([]byte(source)))\n' + needle)
        answer[path] = source
    return answer

inventories = {}
for label, sources in [('union-before', original), ('union-after', current)]:
    code, events = run(label, instrument(sources))
    assert code == 0, (label, code)
    inventory = collections.defaultdict(list)
    for event in events:
        match = re.search(r'split-case seed=(\d+) sha256=([a-f0-9]{64})', event.get('Output', ''))
        if match:
            parent = event['Test'].split('/')[0]
            inventory[parent].append((int(match[1]), match[2]))
    inventories[label] = {parent: sorted(rows) for parent, rows in inventory.items()}
assert inventories['union-before'] == inventories['union-after'], 'checked input union changed'
for parent, (_, first, last, _, _) in targets.items():
    rows = inventories['union-after'][parent]
    assert [seed for seed, _ in rows] == list(range(first, last + 1)), (parent, rows)
(out / 'inputs.json').write_text(json.dumps(inventories, indent=2) + '\n')
print('Input union: 60 + 30 + 40 + 50 = 180 checked cases, every seed exactly once, all SHA-256 equal', flush=True)

mutants = []
for parent, (file, _, _, seed, leaf) in targets.items():
    path = 'internal/fuzz/' + file
    sources = dict(current)
    source = sources[path]
    begin = source.index('func ' + parent + '(')
    end = source.find('\nfunc ', begin + 1)
    if end < 0:
        end = len(source)
    block = source[begin:end]
    needle = 'if err := os.WriteFile(path, []byte(source), 0o644); err != nil {'
    assert block.count(needle) == 1
    block = block.replace(needle, 'if seed == ' + str(seed) + ' { source += "\\nconst splitFailure: number = \'planted\';\\n" }\n' + needle)
    sources[path] = source[:begin] + block + source[end:]
    mutants.append((parent + '-seed-' + str(seed), parent, leaf, sources, 'splitFailure'))
for parent, old, new, diagnostic in [
    ('TestUndefinedNumbersShapes', '[]string{"omitted",', '[]string{"split-planted-missing-source",', 'never started from split-planted-missing-source'),
    ('TestOverridesShapesAndLower', '"items[0] ="', '"split-planted-missing-marker"', 'never generated split-planted-missing-marker'),
]:
    path = 'internal/fuzz/' + targets[parent][0]
    sources = dict(current)
    assert sources[path].count(old) == 1
    sources[path] = sources[path].replace(old, new)
    mutants.append((parent + '-vocabulary', parent, 'vocabulary', sources, diagnostic))

summary = []
for label, parent, leaf, sources, diagnostic in mutants:
    code, events = run(label, sources, parent)
    failed = [event['Test'] for event in events if event['Action'] == 'fail' and '/' in event.get('Test', '')]
    passed = [event['Test'] for event in events if event['Action'] == 'pass' and '/' in event.get('Test', '')]
    output = ''.join(event.get('Output', '') for event in events)
    expected = parent + '/' + leaf
    assert code == 1 and failed == [expected], (label, code, failed)
    if diagnostic == 'splitFailure':
        # The checker reports TS2322 at the planted declaration, rather than quoting its name.
        assert 'TS2322' in output and ('seed ' + str(targets[parent][3]) + ':') in output, (label, output)
    else:
        assert diagnostic in output, (label, output)
    expected_leaves = (targets[parent][2] - targets[parent][1] + 1) // 5 + (parent in ['TestUndefinedNumbersShapes', 'TestOverridesShapesAndLower'])
    assert len(passed) == expected_leaves - 1, (label, passed)
    summary.append(dict(mutant=label, failed=expected, passing_siblings=len(passed), exit=code))
    print(summary[-1], flush=True)
(out / 'mutants.json').write_text(json.dumps(summary, indent=2) + '\n')

# The checkout tests each retain one preparation, one fixture, and the original operations.
checkout_paths = ['internal/fuzz/reduce_test.go', 'internal/fuzz/runtime_test.go']
checkout_current = {path: (root / path).read_text() for path in checkout_paths}
checkout_original = {path: subprocess.check_output(['git', 'show', base + ':' + path], cwd=root).decode() for path in checkout_paths}

def checkout_instrument(sources):
    answer = {}
    for path, source in sources.items():
        source = source.replace('checkout, err := Prepare(', 't.Log("split-operation prepare")\ncheckout, err := Prepare(', 1)
        if path.endswith('reduce_test.go'):
            source = source.replace('import (', 'import (\n\t"crypto/sha256"', 1)
            source = source.replace('original := checkout.Observe(', 't.Logf("split-operation original sha256=%x", sha256.Sum256([]byte(planted)))\noriginal := checkout.Observe(', 1)
            source = source.replace('signature, err := Derive(', 't.Log("split-operation derive")\nsignature, err := Derive(', 1)
            source = source.replace('return checkout.Observe(source,', 't.Logf("split-operation candidate sha256=%x", sha256.Sum256([]byte(source)))\nreturn checkout.Observe(source,', 1)
            needle = 'reduced := Reduce("planted.a", planted, signature, original, observe, 4, 500)'
            source = source.replace(needle, needle + '\nt.Logf("split-operation reduced tried=%d sha256=%x", reduced.Tried, sha256.Sum256([]byte(reduced.Source)))', 1)
            source = source.replace('if again := checkout.Observe(', 't.Logf("split-operation again sha256=%x", sha256.Sum256([]byte(reduced.Source)))\nif again := checkout.Observe(', 1)
        else:
            source = source.replace('library, err := native.RuntimeLibrary(', 't.Log("split-operation library")\nlibrary, err := native.RuntimeLibrary(', 1)
            source = source.replace('outcome := checkout.Try(', 't.Log("split-operation program")\noutcome := checkout.Try(', 1)
        answer[path] = source
    return answer

operations = {}
for label, sources in [('checkout-before', checkout_original), ('checkout-after', checkout_current)]:
    inventory = {}
    for parent in ['TestReduceKeepsTheSignature', 'TestFuzzerSharesRuntimeLibrary']:
        code, events = run(label + '-' + parent, checkout_instrument(sources), parent)
        assert code == 0, (label, parent, code)
        rows = []
        for event in events:
            match = re.search(r'split-operation (.*)', event.get('Output', ''))
            if match:
                rows.append(match[1])
        inventory[parent] = sorted(rows)
    operations[label] = inventory
assert operations['checkout-before'] == operations['checkout-after'], 'checkout operation/input union changed'
(out / 'checkout-inputs.json').write_text(json.dumps(operations, indent=2) + '\n')
print('Checkout union: identical preparations, fixture observations, derivation, runtime comparison, program execution, and reduction candidate hashes/counts', flush=True)

# Fail each checkout assertion without corrupting its sibling's shared preparation.
for label, parent, leaf, path, old, new, diagnostic in [
    ('runtime-program', 'TestFuzzerSharesRuntimeLibrary', 'program', 'internal/fuzz/runtime_test.go',
     "console.log('shared runtime');", "console.log('planted runtime');", 'native:'),
    ('runtime-library', 'TestFuzzerSharesRuntimeLibrary', 'library', 'internal/fuzz/runtime_test.go',
     'native.Options{Sanitize: true}', 'native.Options{Sanitize: false}', 'fuzzer uses'),
    ('reduce-signature', 'TestReduceKeepsTheSignature', 'signature', 'internal/fuzz/reduce_test.go',
     'want := "Adamic 0.1 refuses the non-null assertion !"', 'want := "planted wrong signature"', 'want the non-null refusal'),
    ('reduce-minimum', 'TestReduceKeepsTheSignature', 'reduction', 'internal/fuzz/reduce_test.go',
     'if reduced.Source != plantedMinimum {', 'if reduced.Source != plantedMinimum + "// planted extra statement\\n" {', 'reduced to'),
]:
    sources = dict(checkout_current)
    assert sources[path].count(old) == 1
    sources[path] = sources[path].replace(old, new)
    code, events = run(label, sources, parent)
    failed = [event['Test'] for event in events if event['Action'] == 'fail' and '/' in event.get('Test', '')]
    passed = [event['Test'] for event in events if event['Action'] == 'pass' and '/' in event.get('Test', '')]
    output = ''.join(event.get('Output', '') for event in events)
    assert code == 1 and failed == [parent + '/' + leaf] and len(passed) == 1 and diagnostic in output, (label, code, failed, passed)
    summary.append(dict(mutant=label, failed=failed[0], passing_siblings=1, exit=code))
    print(summary[-1], flush=True)
(out / 'mutants.json').write_text(json.dumps(summary, indent=2) + '\n')
