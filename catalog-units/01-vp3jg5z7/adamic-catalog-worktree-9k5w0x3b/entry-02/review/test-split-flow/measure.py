#!/usr/bin/env python3
"""Measure tests separately on four CPUs with uncached observations."""
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time

root = Path(__file__).resolve().parents[2]
mode = sys.argv[1]
logs = Path('/tmp/test-split-flow') / mode
logs.mkdir(parents=True, exist_ok=True)
packages = ['flow', 'fresh', 'lower', 'ir']
env = {**os.environ, 'ADAMIC_GATE_UNCACHED': '1', 'GOMAXPROCS': '4'}
manifest = {}
for relative in subprocess.check_output(['git', 'ls-files', '--cached', '--others', '--exclude-standard'], cwd=root, text=True).splitlines():
    path = root / relative
    if path.is_file() and (relative.startswith(('internal/', 'oracle/', 'dedication/', 'go.', 'cohere/'))):
        manifest[relative] = hashlib.sha256(path.read_bytes()).hexdigest()
results = []
for package in packages:
    binary = logs / (package + '.test')
    if mode == 'after':
        with (logs / (package + '-build.log')).open('w') as log:
            built = subprocess.run(['go', 'test', '-c', '-o', str(binary), './internal/' + package], cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
        assert built.returncode == 0, (package, 'build failed')
    command = ['go', 'test', './internal/' + package, '-json', '-list', '^Test', '-count=1']
    with (logs / (package + '-list.jsonl')).open('w') as log:
        listed = subprocess.run(command, cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
    assert listed.returncode == 0, (package, 'list failed')
    names = []
    for line in (logs / (package + '-list.jsonl')).read_text().splitlines():
        try:
            row = json.loads(line)
        except json.JSONDecodeError:
            continue
        output = row.get('Output', '').strip()
        if re.fullmatch(r'Test\w+', output):
            names.append(output)
    assert names, package
    for name in names:
        command = ['go', 'test', './internal/' + package, '-json', '-run', '^' + name + '$', '-count=1', '-parallel=4', '-timeout=' + ('30s' if mode == 'after' else '30m')]
        working_directory = root
        if mode == 'after':
            command = ['go', 'tool', 'test2json', '-t', '-p', 'github.com/system-inc/adamic/internal/' + package, str(binary), '-test.v=test2json', '-test.run=^' + name + '$', '-test.count=1', '-test.parallel=4', '-test.timeout=30s']
            working_directory = root / 'internal' / package
        log_path = logs / (package + '-' + name + '.jsonl')
        start = time.monotonic()
        with log_path.open('w') as log:
            run = subprocess.run(command, cwd=working_directory, env=env, stdout=log, stderr=subprocess.STDOUT)
        events = []
        for line in log_path.read_text().splitlines():
            try:
                events.append(json.loads(line))
            except json.JSONDecodeError:
                pass
        terminal = [r for r in events if r.get('Test') == name and r['Action'] in ('pass', 'fail', 'skip')]
        assert len(terminal) == 1, (package, name, 'missing test result')
        result = {'package': package, 'test': name, 'elapsed': terminal[0]['Elapsed'], 'wall': time.monotonic()-start, 'exit': run.returncode, 'action':terminal[0]['Action'], 'command': command, 'log':str(log_path)}
        results.append(result)
        (root / 'review/test-split-flow' / (mode + '.json')).write_text(json.dumps({'environment':{'GOMAXPROCS':'4','ADAMIC_GATE_UNCACHED':'1'},'inputs':manifest,'tests':results},indent=2)+'\n')
        print(f"{package}/{name}: {result['elapsed']:.3f}s {result['action']}", flush=True)
        if mode == 'after':
            assert result['exit'] == 0 and result['elapsed'] < 30, result
