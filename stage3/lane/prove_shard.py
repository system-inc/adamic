#!/usr/bin/env python3
"""Break shard 001's owned test source in a private tree and prove its isolation."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]


def run(command, log, *, cwd=ROOT, env=None, expected=0):
    started = time.monotonic()
    with log.open('w') as output:
        result = subprocess.run(command, cwd=cwd, env=env, stdout=output,
                                stderr=subprocess.STDOUT, timeout=90)
    print(f'{command!r}: exit={result.returncode}, seconds={time.monotonic()-started:.6f}; {log}', flush=True)
    if result.returncode != expected:
        raise RuntimeError(f'wanted exit {expected}, got {result.returncode}')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('artifact', type=Path)
    args = parser.parse_args()
    scratch = Path(tempfile.mkdtemp(prefix='hefhych-shard-proof-', dir='/workspace/scratch'))
    manifest = ROOT / 'stage3/lane/shards.json'
    original = manifest.read_bytes()
    units = json.loads(original)['units']
    before = json.loads((args.artifact / 'ready.json').read_text())['payload']
    tree = scratch / 'tree'
    shutil.copytree(args.artifact / 'tree', tree, symlinks=True)
    relative = Path('src/testRunner/unittests/tsserver/projectReferencesSourcemap.ts')
    owned = tree / relative
    source = owned.read_bytes()
    needle = b'it(options.scenario, () => {'
    assert source.count(needle) == 1
    owned.write_bytes(source.replace(needle, needle + b'\n            if (process.pid > 0) throw new Error("hefhych owned-file mutant");', 1))
    (scratch / 'provenance.json').write_text(json.dumps(dict(source=str(relative),
        original=hashlib.sha256(source).hexdigest(), mutant=hashlib.sha256(owned.read_bytes()).hexdigest()), indent=2)+'\n')
    (scratch / 'empty.rc').write_text('')
    env = dict(os.environ, npm_config_cache=str(scratch / 'npm-cache'), npm_config_userconfig='/dev/null',
               npm_config_globalconfig=str(scratch / 'empty.rc'), npm_config_offline='true', npm_config_update_notifier='false')
    run(['npm', 'exec', '--offline', 'hereby', '--', 'tests'], Path('/tmp/hefhych-D-mutant-build.log'), cwd=tree, env=env)
    for name, wanted in [('stage3-lane-file-001', 1), ('stage3-lane-file-002', 0)]:
        row = next(unit for unit in units if unit['name'] == name)
        results = scratch / name
        results.mkdir()
        env = dict(os.environ, STAGE3_ARTIFACT=str(args.artifact), STAGE3_RESULTS=str(results), PYTHONDONTWRITEBYTECODE='1')
        run(row['command'] + ['--mutant-runner', str(tree / 'built/local/run.js')],
            Path('/tmp/hefhych-D-mutant-' + name + '.log'), env=env, expected=wanted)
        report = json.loads((results / name / 'report.json').read_text())
        if wanted:
            assert report['status'] == 'fail' and report['counts'].get('fail', 0) > 0
            assert 'hefhych owned-file mutant' in (results / name / 'oracle/tests.log').read_text()
        else:
            assert report['status'] == 'pass' and report['counts'] == row['expect']
    try:
        manifest.unlink()
        run(['go', 'test', './stage3/lane', '-run', '^TestShard_Stage3LaneFile001$', '-count=1', '-timeout', '90s', '-v'],
            Path('/tmp/hefhych-D-missing-manifest.log'), expected=1)
        assert 'no such file or directory' in Path('/tmp/hefhych-D-missing-manifest.log').read_text()
        manifest.write_text('{"version":1,"units":[]}\n')
        run(['go', 'test', './stage3/lane', '-run', '^TestShard_Stage3LaneFile001$', '-count=1', '-timeout', '90s', '-v'],
            Path('/tmp/hefhych-D-empty-manifest.log'), expected=1)
        assert 'missing version 1 units' in Path('/tmp/hefhych-D-empty-manifest.log').read_text()
        for label, passing, code, catcher in [
                ('counts', 272, 0, 'differs from manifest'),
                ('exit', 273, 1, 'manifest command')]:
            plan = json.loads(original)
            row = next(unit for unit in plan['units'] if unit['name'] == 'stage3-lane-file-001')
            script = ("import json,os,pathlib,sys; "
                      "p=pathlib.Path(os.environ['STAGE3_RESULTS'])/'stage3-lane-file-001'; "
                      "p.mkdir(); "
                      f"(p/'report.json').write_text(json.dumps(dict(status='pass',counts={{'pass':{passing},'fail':0,'skip':0}}))); "
                      f"sys.exit({code})")
            row['command'] = ['python3', '-c', script]
            manifest.write_text(json.dumps(plan))
            log = Path('/tmp/hefhych-D-' + label + '-predicate.log')
            run(['go', 'test', './stage3/lane', '-run', '^TestShard_Stage3LaneFile001$', '-count=1', '-timeout', '90s', '-v'],
                log, expected=1)
            assert catcher in log.read_text()
    finally:
        manifest.write_bytes(original)
    import sys
    sys.path.insert(0, str(ROOT / 'stage3/lane'))
    from artifact import payload_digest
    assert payload_digest(args.artifact / 'tree') == before, 'mutant changed the product'
    print(f'product unchanged; source-copy provenance and result reports: {scratch}', flush=True)


if __name__ == '__main__':
    main()
