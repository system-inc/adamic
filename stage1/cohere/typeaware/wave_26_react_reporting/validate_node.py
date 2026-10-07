"""Node runs original .a over raw checker facts; records the current JS-emission refusal."""
import argparse
import json
import pathlib
import subprocess
import time
ROOT = pathlib.Path(__file__).resolve().parents[4]
OWN = pathlib.Path(__file__).resolve().parent
p = argparse.ArgumentParser()
p.add_argument('--tools', required=True)
p.add_argument('--scratch', required=True)
p.add_argument('--static', required=True)
p.add_argument('--render', required=True)
p.add_argument('--effect', required=True)
a = p.parse_args()
scratch = pathlib.Path(a.scratch).resolve()
scratch.mkdir(parents=True, exist_ok=True)
records = []
def run(name, command, expected=0):
    started = time.monotonic()
    with (scratch / (name + '.stdout')).open('wb') as out, (scratch / (name + '.stderr')).open('wb') as err:
        result = subprocess.run([str(x) for x in command], cwd=ROOT, stdout=out, stderr=err)
    records.append(dict(name=name, command=[str(x) for x in command], exit=result.returncode, seconds=time.monotonic() - started))
    (scratch / 'runs.json').write_text(json.dumps(records, indent=2) + '\n')
    assert result.returncode == expected, (name, (scratch / (name + '.stderr')).read_text())
    return (scratch / (name + '.stdout')).read_bytes()
collector = scratch / 'fixtures'
run('collector-build', ['go', 'build', '-o', collector, OWN / 'testdata/checker_fixtures.go'])
for name in ['static', 'render', 'effect']:
    directory = pathlib.Path(getattr(a, name)).resolve()
    config = directory / 'tsconfig.json'
    manifest = directory / 'controls.manifest'
    facts = scratch / (name + '-facts.json')
    facts.write_bytes(run(name + '-facts', [collector, config, manifest]))
    entry = OWN / (name + '_suite.a')
    expected = (directory / 'controls-go.stdout').read_bytes()
    actual = run(name + '-source', ['node', '--disable-warning=ExperimentalWarning', OWN / 'testdata/node_oracle.mjs', facts, entry, config, manifest])
    assert actual == expected, name + '-source'
    assert (scratch / (name + '-source.stderr')).read_bytes() == b''
    run(name + '-emit-refusal', [pathlib.Path(a.tools) / 'adamic', 'js', entry], 1)
    assert b'unlinked typescript-go library call' in (scratch / (name + '-emit-refusal.stderr')).read_bytes()
    print(name + ': source Node matches production Go bytes; checker-linked JavaScript emission explicitly refuses', flush=True)
