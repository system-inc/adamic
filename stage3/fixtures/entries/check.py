#!/usr/bin/env python3
"""Node goldens and current observations by default; ruled backend contracts with --acceptance."""
import argparse
import json
import re
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
p = argparse.ArgumentParser()
p.add_argument('--scratch', type=Path, required=True)
p.add_argument('--record', action='store_true')
p.add_argument('--acceptance', action='store_true')
args = p.parse_args()
assert not (args.record and args.acceptance)
args.scratch.mkdir(parents=True, exist_ok=True)
contracts = json.loads((HERE / 'expectations.json').read_text())
recorded = json.loads((HERE / 'status.json').read_text()) if (HERE / 'status.json').exists() else None
recorded_by_file = {r['file']: r for r in recorded or []}
assert len(contracts) == 6 and {r['file'] for r in contracts} == {f.name for f in HERE.glob('*.a')}
rows, mutations, acceptance = [], [], []


def run(command, label):
    with (args.scratch / (label + '.stdout')).open('wb') as out, (args.scratch / (label + '.stderr')).open('wb') as err:
        code = subprocess.run(command, cwd=ROOT, stdout=out, stderr=err, timeout=180).returncode
    result = dict(stdout=(args.scratch / (label + '.stdout')).read_text(), stderr=(args.scratch / (label + '.stderr')).read_text(), exit=code)
    (args.scratch / (label + '.exit')).write_text(str(code) + '\n')
    return result


def normalize(text):
    text = text.removeprefix('adamic: ').replace(str(ROOT) + '/', '').replace(str(args.scratch.resolve()) + '/', '<scratch>/')
    return text.removesuffix('exit status 1\n').removesuffix('\n')


def classify(build):
    message = normalize(build['stderr'] + build['stdout'])
    if build['exit'] == 0:
        return dict(outcome='Compiles', what='')
    outcome = 'NotYet' if "stage 0 can't lower" in message else 'Refused' if 'Adamic 0.1 refuses' in message else 'Checker' if ' error TS' in message else None
    assert outcome, ('unexpected compiler failure', build)
    return dict(outcome=outcome, what=message)


def runtime_matches(actual, expected):
    if actual['exit'] != expected['exit'] or actual['stdout'] != expected['stdout']:
        return False
    if 'stderr' in expected:
        return actual['stderr'] == expected['stderr']
    prefix = expected['stderr_prefix']
    return actual['stderr'].startswith(prefix) and actual['stderr'].endswith('\n') and len(actual['stderr'].strip()) > len(prefix.strip())


def contracts_match(stage, expected, native, javascript):
    if stage['outcome'] != expected['outcome']:
        return False
    if expected['outcome'] == 'Refused':
        return any(reason.lower() in stage['what'].lower() for reason in expected['diagnostic_any_of'])
    return native is not None and javascript is not None and runtime_matches(native, expected['runtime']) and runtime_matches(javascript, expected['runtime']) and native == javascript


def observe(file, label, expected):
    node = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(file)], label + '.node')
    assert node['exit'] == 0 and node['stderr'] == '', ('source Node unexpectedly failed', label, node)
    binary = args.scratch / (label + '.native')
    build = run(['go', 'run', './cmd/adamic', 'build', str(file), '-o', str(binary)], label + '.build')
    stage = classify(build)
    native = javascript = None
    if stage['outcome'] == 'Compiles':
        native = run([str(binary.resolve())], label + '.native-run')
        js = run(['go', 'run', './cmd/adamic', 'js', str(file)], label + '.javascript-build')
        assert js['exit'] == 0, ('native-only support does not satisfy both backends', label, js)
        js_file = args.scratch / (label + '.javascript.mjs')
        js_file.write_text(js['stdout'])
        javascript = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(js_file.resolve())], label + '.javascript-run')
        if expected['outcome'] == 'Compiles' and expected['runtime']['exit'] == 70:
            assert contracts_match(stage, expected, native, javascript), ('MISSING OR WRONG REFLECTION CHECK', label, native, javascript)
        else:
            assert native == node and javascript == node, ('SILENT MISCOMPILE', label, node, native, javascript)
    return node, stage, native, javascript


for row in contracts:
    file = HERE / row['file']
    original = file.read_text()
    assert original.splitlines()[0].startswith('// a-check: ')
    node, stage, native, javascript = observe(file, file.stem, row['expected'])
    if not args.record:
        assert node == recorded_by_file[row['file']]['node'], ('Node golden changed', row['file'])
    if not args.record and not args.acceptance:
        assert stage == recorded_by_file[row['file']]['stage0'], ('gap changed; review and refresh observations', row['file'], stage)
    if row['expected']['outcome'] == 'Compiles' and row['expected']['runtime']['exit'] == 0:
        assert node == row['expected']['runtime'], ('ruled success disagrees with Node', row['file'])
    rows.append(dict(file=row['file'], tsc=row['tsc'], reason=row['reason'], node=node, stage0=stage))
    matched = contracts_match(stage, row['expected'], native, javascript)
    acceptance.append(dict(file=row['file'], variant='baseline', expected=row['expected'], actual=stage, native=native, javascript=javascript, matched=matched))
    edit = row['mutant']
    assert original.count(edit['find']) == 1
    mutant = args.scratch / row['file']
    mutant.write_text(original.replace(edit['find'], edit['replace']))
    mutated_node, mutated_stage, mutated_native, mutated_js = observe(mutant, file.stem + '.mutant', edit['expected'])
    assert mutated_node != node, ('source mutant survived Node golden', row['file'])
    if edit['expected']['outcome'] == 'Compiles' and edit['expected']['runtime']['exit'] == 0:
        assert mutated_node == edit['expected']['runtime'], ('mutant success contract disagrees with Node', row['file'])
    mutations.append(dict(file=row['file'], edit=dict(find=edit['find'], replace=edit['replace']), node=mutated_node, stage0=mutated_stage, caught_by='source Node stdout golden comparison'))
    acceptance.append(dict(file=row['file'], variant='mutant', expected=edit['expected'], actual=mutated_stage, native=mutated_native, javascript=mutated_js, matched=contracts_match(mutated_stage, edit['expected'], mutated_native, mutated_js)))
    print('PASS', row['file'], 'Node golden and source mutant;', stage['outcome'], flush=True)

scanner = json.loads((HERE / 'scanner-source.json').read_text())
source = (HERE / '01_scanner_keywords.a').read_text()
assert scanner['entries_count'] == 84
for span in scanner['spans']:
    assert span['text'].replace('\r\n', '\n') in source, ('exact scanner source changed', span['file'], span['line'])
assert rows[0]['node']['stdout'].count('\n') == 85
print('PASS exact scanner spans and all 84 keyword rows')
if args.record:
    (HERE / 'status.json').write_text(json.dumps(rows, indent=2) + '\n')
    (HERE / 'mutants.json').write_text(json.dumps(mutations, indent=2) + '\n')
elif not args.acceptance:
    assert mutations == json.loads((HERE / 'mutants.json').read_text()), 'recorded mutant observation changed'
(args.scratch / 'acceptance.json').write_text(json.dumps(acceptance, indent=2) + '\n')
if args.acceptance:
    failed = [r['file'] + ':' + r['variant'] for r in acceptance if not r['matched']]
    print('Ruled acceptance:', len(acceptance) - len(failed), '/', len(acceptance), 'matched; gaps:', ', '.join(failed))
    raise SystemExit(1 if failed else 0)
print('PASS six fixtures and six source mutants; current observations held independently of future contracts')
