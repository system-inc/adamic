#!/usr/bin/env python3
"""Prove stress observations detect extra evaluation and missing/retargeted cleanup."""
import argparse
import json
from pathlib import Path
import subprocess

p = argparse.ArgumentParser()
p.add_argument('--taste', type=Path, required=True)
p.add_argument('--output', type=Path, required=True)
a = p.parse_args()
a.output.mkdir(parents=True, exist_ok=True)
r = Path(__file__).resolve().parent
status = {x['file']: x['node'] for x in json.loads((r / 'status.json').read_text())}
cases = [
    ('receiver_twice', '22_assignment_once.a', 'receiver().value &&= right();', 'receiver().value &&= (receiver(), right());'),
    ('index_twice', '22_assignment_once.a', 'arrayReceiver()[index()] &&= right();', 'arrayReceiver()[index()] &&= (index(), right());'),
    ('finally_missing', '23_labels_finally.a', 'console.log("finally:" + String(offsetB) + ":" + String(offsetA));', ''),
    ('outer_continue_retargeted', '23_labels_finally.a', 'continue loopB;', 'continue loopA;'),
]
results = []
for name, file, before, after in cases:
    text = (r / file).read_text()
    assert text.count(before) == 1
    mutant = a.output / (name + '.a')
    mutant.write_text(text.replace(before, after))
    binary = a.output / (name + '.bin')
    command = ['go', 'run', './cmd/adamic', 'build', str(mutant), '-o', str(binary)]
    build = subprocess.run(command, cwd=a.taste, capture_output=True, timeout=180)
    result = {'name': name, 'file': file, 'command': command, 'build': {'stdout': build.stdout.decode(), 'stderr': build.stderr.decode(), 'exit': build.returncode}}
    assert build.returncode == 0, result
    native = subprocess.run([str(binary)], capture_output=True, timeout=60)
    result['native'] = {'stdout': native.stdout.decode(), 'stderr': native.stderr.decode(), 'exit': native.returncode}
    result['caught_by_reference_output'] = result['native'] != status[file]
    assert native.returncode == 0 and result['caught_by_reference_output'], result
    results.append(result)
    print(name, 'compiled; original Node reference caught changed output', flush=True)
(r / 'stress-mutants.json').write_text(json.dumps(results, indent=2) + '\n')
