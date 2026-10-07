#!/usr/bin/env python3
"""Observe actual helper calls from each original consumer suite, without editing Go."""
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile

root = Path(__file__).resolve().parents[6]
cohere = root / 'cohere'
output = Path(sys.argv[1]).resolve()
output.mkdir(parents=True, exist_ok=True)
readiness = json.loads((root / 'stage1/cohere/lint/helpers/readiness.json').read_text())
inventory = json.loads((root / 'stage1/cohere/lint/inventory/inventory.json').read_text())
helpers = {
    'nodesFromStaticDeclarations': 'static_utility.go',
    'propertySort': 'walk.go',
    'ParseValue': 'value_parser.go',
}
consumer_names = {r['rule'] for r in readiness['remaining'] if any(
    'github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.' + name in r['remaining_helpers']
    for name in helpers
)}
assert len(consumer_names) == 6
wrapper = r'''
package tailwind
import (
    "encoding/json"
    "os"
    "sync"
)
var adamicCaptureLock sync.Mutex
func adamicCapture(helper string, input any) {
    adamicCaptureLock.Lock()
    defer adamicCaptureLock.Unlock()
    data, err := json.Marshal(struct { Helper string; Input any }{helper, input})
    if err != nil { panic(err) }
    f, err := os.OpenFile(os.Getenv("ADAMIC_HELPER_CAPTURE"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
    if err != nil { panic(err) }
    _, err = f.Write(append(data, '\n'))
    if err != nil { panic(err) }
    if err = f.Close(); err != nil { panic(err) }
}
func nodesFromStaticDeclarations(input []StaticDeclaration) []*Node {
    adamicCapture("nodesFromStaticDeclarations", input)
    return adamicOriginal_nodesFromStaticDeclarations(input)
}
func propertySort(input []*Node) (Sort, []string) {
    adamicCapture("propertySort", input)
    return adamicOriginal_propertySort(input)
}
func ParseValue(input string) []ValueNode {
    adamicCapture("ParseValue", input)
    return adamicOriginal_ParseValue(input)
}
'''
with tempfile.TemporaryDirectory(prefix='wave109-live-capture-') as scratch_text:
    scratch = Path(scratch_text)
    replacements = {}
    package = cohere / 'internal/lint/rules/tailwind/collapse'
    for name, filename in helpers.items():
        source = (package / filename).read_text()
        old = 'func ' + name + '('
        assert source.count(old) == 1
        copy = scratch / filename
        copy.write_text(source.replace(old, 'func adamicOriginal_' + name + '(', 1))
        replacements[str(package / filename)] = str(copy)
    copy = scratch / 'capture.go'
    copy.write_text(wrapper)
    replacements[str(package / 'adamic_wave109_capture.go')] = str(copy)
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': replacements}))
    results = {}
    for rule in inventory['rules']:
        name = rule['name']
        if name not in consumer_names:
            continue
        tests = set()
        for filename in rule['tests']['files']:
            tests.update(re.findall(r'^func (Test\w+)\(', (root / filename).read_text(), re.M))
        assert tests
        slug = name.split('/')[-1]
        capture = output / (slug + '.jsonl')
        log = output / (slug + '.log')
        environment = dict(os.environ, ADAMIC_HELPER_CAPTURE=str(capture))
        command = ['go', 'test', '-overlay=' + str(overlay), './internal/lint/rules/tailwind',
                   '-run', '^(' + '|'.join(sorted(tests)) + ')$', '-count=1', '-v', '-timeout=5m']
        with log.open('wb') as stream:
            result = subprocess.run(command, cwd=cohere, env=environment, stdout=stream, stderr=stream, timeout=330)
        counts = {helper: 0 for helper in helpers}
        if capture.exists():
            for line in capture.read_text().splitlines():
                record = json.loads(line)
                counts[record['Helper']] += 1
        text = log.read_text()
        results[name] = {'exit': result.returncode, 'testsSelected': len(tests),
                         'helperCalls': counts, 'skipLines': text.count('--- SKIP:'),
                         'command': command, 'log': log.name}
        print(name, json.dumps(results[name]))
    assert set(results) == consumer_names
    (output / 'coverage.json').write_text(json.dumps(results, indent=2) + '\n')
