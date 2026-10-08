#!/usr/bin/env python3
"""Check only this unit's four fixtures, census provenance, runtime control, and mutants."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
EXPECTED = {
    '01_factory_clone.a': '2:10:20:true:3\n2:10:20:false:3\nsmall-copy:true same:true clone:true element:true\nlarge-reused:true 5:-1:-1:false:3\n0:-1:-1:false:0\n',
    '02_parser_range.a': '2:7:42:false:3\n1:9:11:true:1\nslice-pos:false slice-end:false\nslice-comma:false slice-flags:false\n1:12:15:false:1\n',
    '03_repair_flags.a': 'before-own:true before-undefined:true\nsame:true after-own:true flags:3\n',
    '04_missing_presence.a': 'ordinary-own:false missing:false\nmissing-own:true missing:true\nundefined-own:true missing:false\n',
}
parser = argparse.ArgumentParser()
parser.add_argument('upstream', type=Path)
parser.add_argument('scratch', type=Path)
parser.add_argument('--compiler', type=Path)
parser.add_argument('--record', action='store_true')
parser.add_argument('--runtime-baseline', type=Path)
parser.add_argument('--runtime-instrumented', type=Path)
parser.add_argument('--runtime-mutant', type=Path)
args = parser.parse_args()
args.scratch.mkdir(parents=True, exist_ok=True)
manifest = json.loads((HERE / 'manifest.json').read_text())
census = json.loads((HERE / 'census.json').read_text())
assert census['stock_diagnostics'] == 0
for row in census['files']:
    assert hashlib.sha256((args.upstream / row['file']).read_bytes()).hexdigest() == row['sha256'], row['file']
assert set(EXPECTED) == {row['file'] for row in manifest}
keys = [(x['file'], x['start'], x['field']) for x in census['accesses']]
assert len(keys) == len(set(keys)), 'duplicate field access'
assert census['summary']['accesses'] == len(census['accesses'])
assert {f['name'] for l in census['layouts'] for f in l['fields']} == {'pos', 'end', 'hasTrailingComma', 'transformFlags'}
status = []
mutants = []

def run(command, stem):
    with (args.scratch / (stem + '.stdout')).open('wb') as out, (args.scratch / (stem + '.stderr')).open('wb') as err:
        code = subprocess.run(command, cwd=ROOT, stdout=out, stderr=err, timeout=300).returncode
    return {'stdout': (args.scratch / (stem + '.stdout')).read_text(), 'stderr': (args.scratch / (stem + '.stderr')).read_text(), 'exit': code}

for row in manifest:
    file = HERE / row['file']
    source = file.read_text()
    for span in row['spans']:
        assert span['text'].replace('\r\n', '\n') in source, 'upstream body changed: ' + row['file']
    node = run(['node', '--disable-warning=ExperimentalWarning', str(ROOT / 'oracle/node.mjs'), str(file)], row['file'] + '-node')
    assert node == {'stdout': EXPECTED[row['file']], 'stderr': '', 'exit': 0}, (row['file'], node)
    mutation = row['mutant']
    assert source.count(mutation['find']) == 1, (row['file'], 'mutation does not select exactly one span')
    mutated = args.scratch / row['file']
    mutated.write_text(source.replace(mutation['find'], mutation['replace']))
    result = run(['node', '--disable-warning=ExperimentalWarning', str(ROOT / 'oracle/node.mjs'), str(mutated)], row['file'] + '-mutant')
    assert result['exit'] == 0 and result['stderr'] == '', ('mutant failed for the wrong reason', row['file'], result)
    assert result['stdout'] != EXPECTED[row['file']], ('surviving mutant', row['file'])
    mutants.append({'file': row['file'], 'edit': mutation, 'node': result, 'caught_by': 'independent Node stdout comparison'})
    command = [str(args.compiler.resolve())] if args.compiler else ['go', 'run', './cmd/adamic']
    binary = args.scratch / (file.stem + '-native')
    build = run([*command, 'build', str(file), '-o', str(binary)], row['file'] + '-build')
    if build['exit']:
        message = build['stderr'] or build['stdout']
        outcome = 'Refused' if 'refused:' in message or 'Adamic 0.1 refuses' in message else 'NotYet' if 'not yet:' in message or 'notyet:' in message or "stage 0 can't lower" in message else 'Checker'
        stage0 = {'outcome': outcome, 'what': message.rstrip('\n')}
    else:
        native = run([str(binary.resolve())], row['file'] + '-native')
        assert native == node, ('SILENT MISCOMPILE', row['file'], node, native)
        stage0 = {'outcome': 'Compiles', 'what': ''}
    status.append({'file': row['file'], 'tsc': [s['file'] + ':' + str(s['line']) for s in row['spans']], 'reason': row['description'], 'node': node, 'stage0': stage0})
    print('PASS', row['file'], 'Node bytes, unchanged source, killed mutant;', stage0['outcome'])

if args.runtime_baseline or args.runtime_instrumented:
    assert args.runtime_baseline and args.runtime_instrumented
    baseline = json.loads(args.runtime_baseline.read_text())
    observed = json.loads(args.runtime_instrumented.read_text())
    assert baseline['corpus'] == observed['corpus'], 'instrumentation changed inputs'
    assert baseline['lexical'] == observed['lexical'], 'instrumentation changed lexical results'
    assert baseline['semantic_digest'] == observed['semantic_digest'], 'instrumentation changed AST or NodeArray values/presence'
    assert baseline['passes'] == observed['passes'], 'instrumentation changed parser traversal'
    print('PASS pristine/instrumented corpus, lexical counts, all three AST passes and semantic digest')
    if args.runtime_mutant:
        mutant = json.loads(args.runtime_mutant.read_text())
        assert mutant['corpus'] == baseline['corpus']
        assert mutant['lexical'] == baseline['lexical']
        assert mutant['passes'] == baseline['passes'], 'mutation changed traversal rather than only field state'
        assert mutant['semantic_digest'] != baseline['semantic_digest'], 'runtime presence mutant survived'
        assert observed['phases']['factory result']['fields']['hasTrailingComma']['present_undefined'] == 0
        assert mutant['phases']['factory result']['fields']['hasTrailingComma']['present_undefined'] > 0
        print('PASS real compiler initializer mutant: present-undefined comma detected and digest differs')

if args.record:
    (HERE / 'status.json').write_text(json.dumps(status, indent=2) + '\n')
    (HERE / 'mutants.json').write_text(json.dumps(mutants, indent=2) + '\n')
else:
    assert json.loads((HERE / 'status.json').read_text()) == status, 'recorded status changed'
    assert json.loads((HERE / 'mutants.json').read_text()) == mutants, 'mutant observations changed'
print('PASS four NodeArray fixtures and four real-source mutants')
