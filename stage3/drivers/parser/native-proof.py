#!/usr/bin/env python3
"""Build and compare the real parser; retain failures, timings, and mutants."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import resource
import shutil
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('compiler', type=Path)
parser.add_argument('scratch', type=Path)
parser.add_argument('slice', type=Path)
parser.add_argument('corpus', type=Path)
parser.add_argument('node_run', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('--case-checkout', type=Path)
parser.add_argument('--full-tree', type=Path)
args = parser.parse_args()
for name in vars(args):
    if getattr(args, name) is not None:
        setattr(args, name, getattr(args, name).resolve())
args.output.mkdir(exist_ok=False)
here = Path(__file__).resolve().parent
binary = args.output / 'parser-native'
report = {'compiler': str(args.compiler), 'scratch': str(args.scratch),
          'entry': str(args.slice / 'parser-proof-main.a'), 'corpus': str(args.corpus)}

def run(command, stem, env=None):
    before = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime
    with (args.output / (stem + '.stdout')).open('wb') as stdout, (args.output / (stem + '.stderr')).open('wb') as stderr:
        result = subprocess.run(command, cwd=args.scratch, env=env, stdout=stdout, stderr=stderr)
    return {'exit': result.returncode,
            'user_seconds': resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime - before}

def digest(file):
    data = file.read_bytes()
    return {'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest()}

oracle = args.node_run / 'node.dump'
report['oracle'] = digest(oracle)
reference = json.loads((here / 'reference.json').read_text())
assert report['oracle'] == {key: reference[key] for key in ['bytes', 'sha256']}
report['dump_schema'] = reference['schema']
report['build'] = run([str(args.compiler), 'build', report['entry'], '-o', str(binary)], 'build')
node_env = dict(os.environ, PARSER_RUNTIME=str(here.parents[2] / 'oracle/adamic.mjs'))
assert node_env.get('PARSER_TYPESCRIPT'), 'set PARSER_TYPESCRIPT to stock 6.0.3'
node_command = ['node', '--disable-warning=ExperimentalWarning', str(here / 'node.mjs'),
                report['entry'], str(args.corpus), str(args.node_run / 'manifest')]
report['node_runs'] = []
for index in range(3):
    stem = 'node-' + str(index)
    result = run(node_command, stem, node_env)
    result['comparison'] = run(['cmp', str(oracle), str(args.output / (stem + '.stdout'))], stem + '-cmp')['exit']
    report['node_runs'].append(result)
    assert result['exit'] == 0 and result['comparison'] == 0
report['node_best_user_seconds'] = min(r['user_seconds'] for r in report['node_runs'])
report['native_runs'] = []
if report['build']['exit'] == 0:
    report['native_binary_bytes'] = binary.stat().st_size
    native_command = [str(binary), str(args.corpus), str(args.node_run / 'manifest')]
    for index in range(3):
        stem = 'native-' + str(index)
        result = run(native_command, stem)
        result['comparison'] = run(['cmp', str(oracle), str(args.output / (stem + '.stdout'))], stem + '-cmp')['exit']
        result['stderr_comparison'] = run(['cmp', str(args.node_run / 'node.stderr'), str(args.output / (stem + '.stderr'))], stem + '-stderr-cmp')['exit']
        report['native_runs'].append(result)
        if index == 0:
            output = args.output / (stem + '.stdout')
            report['native_output'] = digest(output)
            mutant = args.output / 'native-byte-mutant.stdout'
            shutil.copyfile(output, mutant)
            if mutant.stat().st_size:
                with mutant.open('r+b') as changed:
                    byte = changed.read(1)
                    changed.seek(0)
                    changed.write(bytes([byte[0] ^ 1]))
                report['native_byte_mutant_cmp'] = run(['cmp', str(output), str(mutant)], 'native-byte-mutant-cmp')['exit']
                assert report['native_byte_mutant_cmp'] == 1
        if result['exit'] != 0:
            break
    report['native_best_user_seconds'] = min(r['user_seconds'] for r in report['native_runs'])
    if shutil.which('perf'):
        report['perf_native'] = run(['perf', 'stat', '-e', 'instructions', *native_command], 'perf-native')
        report['perf_node'] = run(['perf', 'stat', '-e', 'instructions', *node_command], 'perf-node', node_env)
    else:
        report['instructions'] = 'unavailable: perf executable not installed'
else:
    report['native_binary_bytes'] = None
    report['native_best_user_seconds'] = None
    report['native_byte_mutant_cmp'] = None
    report['instructions'] = 'not measured: build failed; perf executable ' + ('present' if shutil.which('perf') else 'not installed')
report['case_acceptance'] = None
if report['build']['exit'] == 0 and len(report['native_runs']) == 3 and all(
        r['exit'] == 0 and r['comparison'] == 0 for r in report['native_runs']):
    report['case_acceptance'] = run([
        'python3', str(here / 'case-reference.py'),
        str(args.case_checkout or args.corpus), str(args.full_tree or args.corpus),
        str(args.slice), str(args.output / 'cases'), '--native', str(binary),
        '--reference', str(here / 'cases-reference.json')], 'case-acceptance', node_env)
report['status'] = 'green' if len(report['native_runs']) == 3 and all(
    r['exit'] == 0 and r['comparison'] == 0 and r['stderr_comparison'] == 0 for r in report['native_runs']) and report['case_acceptance'] is not None and report['case_acceptance']['exit'] == 0 else 'red'
(args.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report, indent=2))
