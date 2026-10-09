#!/usr/bin/env python3
"""Run the checkout's full adaptation and upstream oracle in fresh directories."""
import argparse
import os
import json
import shutil
from pathlib import Path
import subprocess
import time
import tempfile

from check import check_results, write_verdict

LANE = Path(__file__).resolve().parent
ROOT = LANE.parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('results', type=Path, nargs='?',
                        help='new results directory (default: a fresh temporary directory)')
    parser.add_argument('--artifact', type=Path, help='verified prepared artifact for the same-tree union measurement')
    args = parser.parse_args()
    if args.results is None:
        results = Path(tempfile.mkdtemp(prefix='stage3-lane-')).resolve()
    else:
        results = args.results.resolve()
        try:
            results.mkdir(parents=True, exist_ok=False)
        except FileExistsError:
            parser.error(f'refusing to replace existing results directory: {results}')
    print(results, flush=True)
    started = time.monotonic()
    execution = dict(commit=subprocess.check_output(['git', 'rev-parse', 'HEAD'],
                                                     cwd=ROOT, text=True).strip(),
                     apply_exit=None, oracle_exit=None)

    def run(name, command, env=None):
        with (results / f'{name}.log').open('w') as log:
            try:
                return subprocess.run(command, cwd=ROOT, stdout=log,
                                      stderr=subprocess.STDOUT, env=env).returncode
            except OSError as error:
                log.write(str(error) + '\n')
                return 127

    try:
        execution['platform'] = json.loads(subprocess.check_output(
            ['node', '-p', 'JSON.stringify({os:process.platform,arch:process.arch,node:process.version})'], text=True))
        key = '/'.join(execution['platform'][field] for field in ('os', 'arch', 'node'))
        expected = json.loads((LANE / 'expected.json').read_text())
        if key not in expected['platforms']:
            raise ValueError(f'unknown platform: {key}; add measured expectations deliberately')
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        (results / 'execution.json').write_text(json.dumps(execution, indent=2) + '\n')
        return write_verdict(results, dict(status='fail', errors=[str(error)], execution=execution))

    tree = results / 'adapted-tree'
    if args.artifact is None:
        execution['apply_exit'] = run('apply', ['bash', str(ROOT / 'stage3/apply.sh'), str(tree)])
    else:
        from shard import shared_artifact
        from shard_plan import load_plan
        from view import make_view
        os.environ['STAGE3_ARTIFACT'] = str(args.artifact.resolve())
        artifact, ready = shared_artifact(load_plan())
        make_view(artifact / 'tree', tree)
        execution['artifact'] = ready['key']
        execution['apply_exit'] = ready['phases']['apply']['exit']
        execution['producer_phases'] = ready['phases']
        shutil.copyfile(artifact / 'ready.json', results / 'artifact-ready.json')
    table_error = None
    if execution['apply_exit'] == 0:
        try:
            shutil.copyfile(tree / 'patch-set.md', results / 'patch-set.md')
        except OSError as error:
            table_error = f'cannot preserve apply patch table: {error}'
    if execution['apply_exit'] == 0 and table_error is None:
        command = ['bash', str(ROOT / 'stage3/oracle/run.sh'), str(tree), str(results / 'oracle')]
        env = dict(os.environ)
        if args.artifact is not None:
            command.append('--prepared')
            env['STAGE3_TASK_PROFILE'] = str(results / 'whole-tasks.jsonl')
            env['NODE_OPTIONS'] = env.get('NODE_OPTIONS', '') + ' --require=' + str(ROOT / 'stage3/oracle/profile-tasks.cjs')
            env['NODE_DISABLE_COMPILE_CACHE'] = '1'
        execution['oracle_exit'] = run('oracle', command, env)
    execution['wall_seconds'] = round(time.monotonic() - started, 3)
    (results / 'execution.json').write_text(json.dumps(execution, indent=2) + '\n')
    if execution['apply_exit'] != 0:
        report = dict(status='fail', errors=[f"apply exit: expected 0, observed {execution['apply_exit']}; see apply.log"])
    elif table_error is not None:
        report = dict(status='fail', errors=[table_error])
    else:
        if args.artifact is not None:
            os.environ['TSC_ADAPT_TYPESCRIPT'] = str(artifact / 'api')
        report = check_results(results)
    report['execution'] = execution
    report['results'] = str(results)
    return write_verdict(results, report)


if __name__ == '__main__':
    raise SystemExit(main())
