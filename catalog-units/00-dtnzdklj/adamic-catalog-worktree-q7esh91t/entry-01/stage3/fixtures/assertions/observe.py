"""Record outside Node and current stage0 observations, including a single-token mutant."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

bucket = Path(__file__).resolve().parent
repository = bucket.parents[2]
manifest = json.loads((bucket / 'manifest.json').read_text())
logdir = bucket / 'logs'
logdir.mkdir(exist_ok=True)

def run(command):
    result = subprocess.run(command, cwd=repository, capture_output=True, timeout=120)
    return {'stdout': result.stdout.decode(), 'stderr': result.stderr.decode(), 'exit': result.returncode}

def outcome(observation):
    text = observation['stdout'] + observation['stderr']
    if observation['exit'] == 0:
        return {'outcome': 'Compiles', 'what': ''}
    if 'Adamic 0.1 refuses' in text:
        name = 'Refused'
    elif "stage 0 can't lower" in text:
        name = 'NotYet'
    elif 'TS' in text:
        name = 'Checker'
    else:
        raise AssertionError('Unclassified failure: ' + text)
    return {'outcome': name, 'what': text.removesuffix('exit status 1\n')}

records = []
with tempfile.TemporaryDirectory(prefix='assertions-observe-') as scratch:
    for row in manifest:
        filename = 'stage3/fixtures/assertions/' + row['file']
        node = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', filename])
        assert node['exit'] == 0 and node['stderr'] == '', (filename, node)
        command = ['go', 'run', './cmd/adamic', 'build', filename, '-o', scratch + '/program']
        build = run(command)
        stage0 = outcome(build)
        native = None
        if stage0['outcome'] == 'Compiles':
            native = run([scratch + '/program'])
            assert native == node, ('SILENT MISCOMPILE', filename, native, node)
        (logdir / (row['file'] + '.json')).write_text(json.dumps({'command': command, 'build': build, 'node': node, 'native': native}, indent=2) + '\n')
        records.append({k: row[k] for k in ['file', 'tsc', 'reason']} | {'node': node, 'stage0': stage0})
        print(row['file'], stage0['outcome'], flush=True)
    original = (bucket / '01_map_call.a').read_text()
    assert original.count('map.get(key)!') == 1
    mutant_path = Path(scratch) / 'missing-bang.a'
    mutant_path.write_text(original.replace('map.get(key)!', 'map.get(key)'))
    mutant = run(['go', 'run', './cmd/adamic', 'build', str(mutant_path), '-o', scratch + '/mutant'])
    assert outcome(mutant)['outcome'] == 'Checker', mutant
    assert records[0]['stage0']['outcome'] != 'Checker'
    # Normalize only the scratch path so the committed evidence has a stable location.
    mutant['stderr'] = mutant['stderr'].replace(scratch, '<scratch>')
    (logdir / 'mutant.json').write_text(json.dumps({'edit': '01_map_call.a: map.get(key)! -> map.get(key)', 'baseline': records[0]['stage0'], 'mutant': mutant, 'outcome': outcome(mutant)}, indent=2) + '\n')
    print('missing ! mutant: Checker', flush=True)
(bucket / 'status.json').write_text(json.dumps(records, indent=2) + '\n')
