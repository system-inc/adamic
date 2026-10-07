"""Record the external Node oracle and current-main stage0 observations."""
import json
from pathlib import Path
import subprocess

BUCKET = Path(__file__).resolve().parent
ROOT = BUCKET.parents[2]
rows = json.loads((BUCKET / 'sources.json').read_text())
logs = BUCKET / 'logs'
logs.mkdir(exist_ok=True)
for row in rows:
    file = (BUCKET / row['file']).relative_to(ROOT).as_posix()
    name = Path(row['file']).parts[0]
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', file],
                          cwd=ROOT, capture_output=True, text=True)
    row['node'] = dict(stdout=node.stdout, stderr=node.stderr, exit=node.returncode)
    binary = '/tmp/cycles-' + name
    build = subprocess.run(['go', 'run', './cmd/adamic', 'build', file, '-o', binary],
                           cwd=ROOT, capture_output=True, text=True)
    (logs / (name + '.build.log')).write_text(build.stdout + build.stderr)
    diagnostic = build.stdout + build.stderr.removesuffix('exit status 1\n')
    if build.returncode == 0:
        outcome = 'Compiles'
        native = subprocess.run([binary], cwd=ROOT, capture_output=True, text=True)
        observation = dict(stdout=native.stdout, stderr=native.stderr, exit=native.returncode)
        (logs / (name + '.native.json')).write_text(json.dumps(observation, indent=2) + '\n')
        if observation != row['node']:
            raise RuntimeError('SILENT MISCOMPILE: ' + file)
        diagnostic = ''
    elif 'refused:' in diagnostic or 'refuses' in diagnostic:
        outcome = 'Refused'
    elif 'not yet:' in diagnostic or 'NotYet' in diagnostic or 'not yet' in diagnostic:
        outcome = 'NotYet'
    elif 'TS' in diagnostic:
        outcome = 'Checker'
    else:
        raise RuntimeError('unclassified stage0 failure: ' + diagnostic)
    row['stage0'] = dict(outcome=outcome, what=diagnostic)
    print(name + ': Node exit ' + str(node.returncode) + ', stage0 ' + outcome)
(BUCKET / 'status.json').write_text(json.dumps(rows, indent=2) + '\n')
safe, mutant = rows[4], rows[5]
assert safe['node']['exit'] == 0
assert mutant['node']['exit'] == 70
assert "ReferenceError: Cannot access 'emptyArray' before initialization" in mutant['node']['stderr']
assert safe['node']['stdout'] != mutant['node']['stdout']
# Exactly the same files except for the order of the two main imports.
for relative in ['core.a', 'utilities.a', '_namespaces/ts.a']:
    assert (BUCKET / Path(safe['file']).parent / relative).read_bytes() == (BUCKET / Path(mutant['file']).parent / relative).read_bytes()
a = (BUCKET / safe['file']).read_text().splitlines()
b = (BUCKET / mutant['file']).read_text().splitlines()
assert a[:3] == b[:3] and a[3:5] == b[3:5][::-1] and a[5:] == b[5:]
raw_runner = Path('/tmp/cycles-raw-node.mjs')
raw_runner.write_text((ROOT / 'oracle/node.mjs').read_text().replace("await import(runtimeUrl);", "// Raw Node: no Adamic panic normalization."))
raw = subprocess.run(['node', '--disable-warning=ExperimentalWarning', str(raw_runner), str(BUCKET / mutant['file'])], cwd=ROOT, capture_output=True, text=True)
(logs / '06_import_order_mutant.raw-node.json').write_text(json.dumps(dict(stdout=raw.stdout, stderr=raw.stderr, exit=raw.returncode), indent=2) + '\n')
assert raw.returncode == 1
assert "ReferenceError: Cannot access 'emptyArray' before initialization" in raw.stderr
print('Import-order mutant caught by Node stdout, stderr and exit comparison; raw Node exits 1 with ReferenceError.')
