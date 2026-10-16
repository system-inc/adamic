#!/usr/bin/env python3
"""Run one manifest unit cold against the adapted tree: stage3/lane's product, or a verified shared artifact."""
import argparse
from collections import Counter
import difflib
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import time
from artifact import input_key, payload_digest
from check import api_check, test_counts
from shard_plan import LANE, digest, load_plan, names_digest, runtime_hash, title
from view import make_view
sys.path.append(str(LANE.parent / 'oracle'))
from failure_details import read_failures, markdown
sys.path.append(str(LANE.parent))
from caches import cache_root


def shared_artifact(plan):
    supplied = os.environ.get('STAGE3_ARTIFACT')
    artifact = Path(supplied) if supplied else cache_root() / 'adamic-stage3-artifacts' / plan['artifact']
    if not (artifact / 'ready.json').is_file():
        raise ValueError('shared artifact missing; run prepare_shards.py before the gate and set STAGE3_ARTIFACT')
    ready = json.loads((artifact / 'ready.json').read_text())
    key, inputs = input_key()
    if ready['key'] != plan['artifact'] or key != ready['key'] or ready['inputs'] != inputs:
        raise ValueError('shared artifact inputs hash does not match this checkout and toolchain')
    if any(ready['phases'][phase]['exit'] != 0 for phase in ('apply', 'install', 'build', 'harness')):
        raise ValueError('shared artifact producer phase failed')
    if ready['payload'] != payload_digest(artifact / 'tree'):
        raise ValueError('shared artifact payload corruption')
    if payload_digest(artifact / 'api') != plan['api_digest']:
        raise ValueError('shared stock API parser corruption')
    if runtime_hash() != plan['runtime']:
        raise ValueError('shard harness inputs changed; regenerate the manifest')
    return artifact.resolve(), ready


def adapted_inputs(plan):
    """The adapted tree a unit runs against, the stock API parser, the producer's phases, and the packages the view links
    in as node_modules (None when the tree holds its own). With STAGE3_ADAPTED_TREE, the tree is stage3/lane's
    TestProduct_Stage3AdaptedTree product, whose key is its identity, so none of the shared artifact's identity checks
    apply; the parser is stage3/api's install, checked against the manifest's digest, and the packages are
    stage3/typescript's install, the tree's devDependencies, which the product leaves out."""
    supplied = os.environ.get('STAGE3_ADAPTED_TREE')
    if not supplied:
        artifact, ready = shared_artifact(plan)
        return artifact / 'tree', artifact / 'api', ready['phases'], None
    tree = Path(supplied).resolve()
    if not (tree / 'built/local/run.js').is_file():
        raise ValueError('the adapted tree product holds no built harness: ' + str(tree))
    api = LANE.parent / 'api/node_modules/typescript'
    if payload_digest(api) != plan['api_digest']:
        raise ValueError('stage3/api\'s stock API parser differs from the manifest\'s')
    if runtime_hash() != plan['runtime']:
        raise ValueError('shard harness inputs changed; regenerate the manifest')
    return tree, api, None, LANE.parent / 'typescript/node_modules'


def baseline_diff(tree):
    differences, parts = [], []
    local, reference = tree / 'tests/baselines/local', tree / 'tests/baselines/reference'
    for file in sorted(local.rglob('*')):
        if not file.is_file():
            continue
        relative = file.relative_to(local)
        deleted = file.name.endswith('.delete')
        refname = Path(str(relative)[:-7]) if deleted else relative
        ref = reference / refname
        old = ref.read_bytes() if ref.exists() else b''
        new = b'' if deleted else file.read_bytes()
        if old != new or deleted or not ref.exists():
            differences.append(str(relative))
            parts.extend(difflib.unified_diff(old.decode('utf-8', errors='backslashreplace').splitlines(True),
                new.decode('utf-8', errors='backslashreplace').splitlines(True),
                fromfile='reference/' + str(refname), tofile='local/' + str(relative)))
    return differences, ''.join(parts)


def execute(command, tree, log, env, seconds):
    with log.open('w') as stream:
        child = subprocess.Popen(command, cwd=tree, env=env, stdout=stream,
                                 stderr=subprocess.STDOUT, start_new_session=True)
        try:
            code = child.wait(timeout=max(.01, seconds))
        except subprocess.TimeoutExpired:
            code = 124
        finally:
            try:
                os.killpg(child.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            child.wait()
    return code


def shard_seconds():
    """A unit's wall limit: SHARD_CONTRACT's 30 s for a unit run on its own, or, when a TestShard_ test runs it, the limit
    that test sets in STAGE3_SHARD_SECONDS, its own deadline less a margin for teardown, so the child is always killed
    before the test is. Inside a Go test the time bound is the test's and its unit's, where Planner and the census hold
    it, and a shard measured at 24.7 s alone on 4 CPUs runs past 30 s with seven others at an 8-CPU share."""
    supplied = os.environ.get('STAGE3_SHARD_SECONDS')
    return float(supplied) if supplied else 30.0


def validate_observation(unit, counts, records, paths, failures, seconds, code, limit=30.0):
    errors = []
    if counts != unit['expect']:
        errors.append(f"counts: expected {unit['expect']}, observed {counts}")
    if names_digest(records) != unit['names_digest']:
        errors.append('selected test identities do not equal this unit inventory')
    if paths != unit['baseline_diffs']:
        errors.append(f"baseline paths: expected {unit['baseline_diffs']}, observed {paths}")
    expected = json.loads((LANE / 'expected.json').read_text())
    wanted = expected['failed_tests'] if unit['expect']['fail'] else []
    if failures != wanted:
        errors.append(f'failure titles: expected {wanted}, observed {failures}')
    if code != (1 if unit['expect']['fail'] else 0):
        errors.append(f'upstream exit: {code}')
    if seconds >= limit:
        errors.append(f'invalid test unit: cold wall {seconds:.6f} seconds exceeds {limit:g} seconds')
    return errors


def timed_out(name, code, started, limit):
    """Upstream killed at the limit is a timeout naming the unit, before anything reads what it left half written."""
    if code == 124:
        raise ValueError(f'{name} timed out after {time.monotonic() - started:.1f} seconds, past its {limit:g} second limit')


def check_unit_inputs(unit, plan, wanted_hash):
    definition = {key: value for key, value in unit.items() if key != 'inputs_hash'}
    actual_hash = digest(dict(artifact=plan['artifact'], api=plan['api_digest'], runtime=runtime_hash(), unit=definition))
    if actual_hash != wanted_hash or wanted_hash != unit['inputs_hash']:
        raise ValueError('unit inputs hash differs from the manifest')


def run_unit(name, results, wanted_hash, plan=None, mutant_case=None, mutant_content=None, mutant_runner=None):
    started = time.monotonic()
    limit = shard_seconds()
    plan = plan or load_plan()
    unit = next(row for row in plan['units'] if row['name'] == name)
    output = results / name
    output.mkdir(parents=True, exist_ok=False)
    report = dict(name=name, inputs_hash=wanted_hash, artifact=plan['artifact'], counts={}, errors=[], status='fail')
    try:
        check_unit_inputs(unit, plan, wanted_hash)
        source, api, phases, modules = adapted_inputs(plan)
        tree = output / 'adapted-tree'
        if mutant_case is not None:
            if mutant_case.is_absolute() or '..' in mutant_case.parts or not mutant_case.is_relative_to('tests/cases'):
                raise ValueError('mutant case must be a relative tests/cases path')
            if mutant_content is None:
                raise ValueError('mutant case requires content')
        make_view(source, tree, mutant_case)
        if modules is not None:
            (tree / 'node_modules').symlink_to(modules, target_is_directory=True)
            # node's ESM loader resolves a module by its real path, so scripts/build/options.mjs reached through the
            # view's link would look for minimist above the product, never in the view's node_modules: the watcher
            # cases' run-prepared.mjs imports from scripts/, so it's the view's own copy.
            (tree / 'scripts').unlink()
            shutil.copytree(source / 'scripts', tree / 'scripts', symlinks=True)
        if mutant_case is not None:
            import hashlib
            content = mutant_content.read_bytes()
            (tree / mutant_case).write_bytes(content)
            report['mutant'] = dict(case=str(mutant_case), sha256=hashlib.sha256(content).hexdigest())
        if mutant_runner is not None:
            import hashlib
            content = mutant_runner.read_bytes()
            (tree / 'built/local/run.js').write_bytes(content)
            report['mutant'] = dict(case='built/local/run.js', sha256=hashlib.sha256(content).hexdigest())
        oracle = output / 'oracle'
        oracle.mkdir()
        observations = oracle / 'mocha-errors'
        observations.mkdir()
        env = dict(os.environ, NODE_DISABLE_COMPILE_CACHE='1', STAGE3_ERROR_DIR=str(observations))
        env['NODE_OPTIONS'] = env.get('NODE_OPTIONS', '') + ' --require=' + str(LANE.parent / 'oracle/observe-errors.cjs')
        if 'grep' in unit:
            env['STAGE3_TEST_EVENTS'] = str(oracle / 'events.jsonl')
            env['NODE_OPTIONS'] += ' --require=' + str(LANE.parent / 'oracle/observe-tests.cjs')
            command = ['node', str(LANE.parent / 'oracle/run-prepared.mjs'), str(tree),
                       '--workers=1', '--light=false', '--no-colors',
                       '--runners=unittest', '--tests=' + unit['grep']]
            code = execute(command, tree, oracle / 'tests.log', env, limit - 0.5 - (time.monotonic() - started))
            timed_out(name, code, started, limit)
            events = [json.loads(line) for line in (oracle / 'events.jsonl').read_text().splitlines()]
            records = [dict(row, task=unit['tasks'][0]) for row in events]
            counts = {'pass': sum(row['status'] == 'pass' for row in records),
                      'fail': sum(row['status'] == 'fail' for row in records),
                      'skip': sum(row['status'] == 'pending' for row in records)}
            reporter = test_counts((oracle / 'tests.log').read_text())
            if reporter != dict(passing=counts['pass'], failing=counts['fail'], pending=counts['skip']):
                raise ValueError('upstream reporter and case observation counts disagree')
        else:
            tasks = output / 'tasks.json'
            tasks.write_text(json.dumps(unit['tasks']))
            command = ['node', str(LANE.parent / 'oracle/run-tasks.cjs'), str(tree), str(tasks), str(oracle / 'worker'), '1']
            code = execute(command, tree, oracle / 'tests.log', env, limit - 0.5 - (time.monotonic() - started))
            timed_out(name, code, started, limit)
            rows = [json.loads(line) for line in (oracle / 'worker/tasks.jsonl').read_text().splitlines()]
            if Counter((row['task']['runner'], row['task']['file']) for row in rows) != Counter(
                    (task['runner'], task['file']) for task in unit['tasks']):
                raise ValueError('upstream task union differs from assigned tasks')
            records = [dict(record, task=row['task'], status=status) for row in rows
                       for status, key in [('pass', 'passes'), ('fail', 'errors')] for record in row[key]]
            counts = {'pass': sum(row['passing'] for row in rows), 'fail': sum(len(row['errors']) for row in rows), 'skip': 0}
        (output / 'observations.json').write_text(json.dumps(records) + '\n')
        paths, diff = baseline_diff(tree)
        (oracle / 'baseline.diff').write_text(diff)
        failures = [title(row['name']) for row in records if row['status'] == 'fail']
        # Native IPC does not print a reporter summary. Keep original errors and
        # a conventional summary for the unchanged diagnostic parser.
        if 'grep' not in unit:
            text = '\n' + f"  {counts['pass']} passing\n  {counts['fail']} failing\n  {counts['skip']} pending\n"
            for index, row in enumerate(record for record in records if record['status'] == 'fail'):
                text += f"\n  {index + 1}) {title(row['name'])}:\n\n{row['error']}\n{row['stack']}\n"
            with (oracle / 'tests.log').open('a') as stream:
                stream.write(text)
        errors = []
        if name == 'stage3-lane-byte-comparisons':
            os.environ['TSC_ADAPT_TYPESCRIPT'] = str(api)
            errors, report['api'] = api_check(output, json.loads((LANE / 'sanctioned-api.json').read_text()))
        seconds = time.monotonic() - started
        errors += validate_observation(unit, counts, records, paths, failures, seconds, code, limit)
        report.update(counts=counts, failed_tests=failures, baseline_diffs=paths,
                      upstream_exit=code, errors=errors, status='fail' if errors else 'pass',
                      producer_phases=phases, command=command)
        details = read_failures((oracle / 'tests.log').read_text(), diff, observations)
        report['failures'] = details
        (output / 'failures.md').write_text(markdown(details))
    except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as error:
        report['errors'].append(str(error))
    report['seconds'] = time.monotonic() - started
    if report['seconds'] >= limit and report['status'] == 'pass':
        report['status'] = 'fail'
        report['errors'].append(f'invalid test unit: total wall exceeds {limit:g} seconds')
    (output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
    (output / 'verdict.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps({key: report[key] for key in ['name', 'status', 'counts', 'seconds', 'errors']}), flush=True)
    return 0 if report['status'] == 'pass' else 1


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    parser.add_argument('results')
    parser.add_argument('--inputs-hash', required=True)
    parser.add_argument('--mutant-case', type=Path, help='real private case input for failure proofs only')
    parser.add_argument('--mutant-content', type=Path)
    parser.add_argument('--mutant-runner', type=Path, help='private built test input for case-level failure proofs only')
    args = parser.parse_args()
    results = Path(os.path.expandvars(args.results)).resolve()
    if '$' in str(results):
        parser.error('STAGE3_RESULTS is unset')
    raise SystemExit(run_unit(args.name, results, args.inputs_hash, mutant_case=args.mutant_case, mutant_content=args.mutant_content, mutant_runner=args.mutant_runner))
