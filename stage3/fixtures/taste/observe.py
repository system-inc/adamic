#!/usr/bin/env python3
"""Record stock Node, main stage 0, and the taste branch without altering compilers."""
import argparse
import json
import os
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--main', type=Path, required=True)
parser.add_argument('--taste', type=Path, required=True)
parser.add_argument('--logs', type=Path, required=True)
args = parser.parse_args()
bucket = Path(__file__).resolve().parent
args.logs.mkdir(parents=True, exist_ok=True)
manifest = json.loads((bucket / 'manifest.json').read_text())

def run(command, cwd, label):
    result = subprocess.run(command, cwd=cwd, capture_output=True, timeout=180)
    observation = {'stdout': result.stdout.decode(), 'stderr': result.stderr.decode(), 'exit': result.returncode}
    (args.logs / (label + '.json')).write_text(json.dumps({'command': command, 'cwd': str(cwd), **observation}, indent=2) + '\n')
    return observation

def compile_fixture(repo, file, label, oracle):
    binary = str(args.logs / (label + '.bin'))
    build = run(['go', 'run', './cmd/adamic', 'build', str(file), '-o', binary], repo, label + '-build')
    diagnostic = build['stdout'] + build['stderr']
    if build['exit'] == 0:
        native = run([binary], repo, label + '-native')
        return {'outcome': 'Compiles', 'what': ''}, {'build': build, 'native': native, 'matches_node': native == oracle}
    if "stage 0 can't lower" in diagnostic:
        outcome = 'NotYet'
    elif 'Adamic 0.1 refuses' in diagnostic:
        outcome = 'Refused'
    else:
        outcome = 'Checker'
    return {'outcome': outcome, 'what': diagnostic}, {'build': build}

statuses = []
evidence = []
for item in manifest:
    file = bucket / item['file']
    node = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(file)], args.main, item['file'] + '-node')
    main, main_evidence = compile_fixture(args.main, file, item['file'] + '-main', node)
    taste, taste_evidence = compile_fixture(args.taste, file, item['file'] + '-taste', node)
    statuses.append({**item, 'node': node, 'stage0': main})
    evidence.append({'file': item['file'], 'main': main_evidence, 'taste': taste_evidence, 'taste_stage0': taste})
    print(item['file'], 'Node', node['exit'], 'main', main['outcome'], 'taste', taste['outcome'], flush=True)
(bucket / 'status.json').write_text(json.dumps(statuses, indent=2) + '\n')
(bucket / 'evidence.json').write_text(json.dumps(evidence, indent=2) + '\n')
