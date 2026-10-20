#!/usr/bin/env python3
"""Check the cold ledger, execute union and timing mutants, and retain matched load experiments."""
import argparse
from collections import Counter
import copy
import json
import math
import os
from pathlib import Path
import statistics
import subprocess
import sys

from shard_plan import LANE, names_digest

BEFORE = [f'stage3-lane-file-{index:03d}' for index in (7, 18, 109, 10, 36, 49, 57, 63)]
AFTER = [f'stage3-lane-file-{index:03d}' for index in range(20, 41)]


def stats(rows):
    values = sorted(row['seconds'] for row in rows)
    return dict(n=len(values), minimum=values[0], median=statistics.median(values),
                p95=values[math.ceil(.95 * len(values)) - 1], maximum=values[-1])


def require_budget(rows, seconds):
    assert all(row['seconds'] < seconds for row in rows), f'alone budget {seconds:g} s'


def task_key(task):
    return task['runner'], task['file']


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('cold', type=Path)
    parser.add_argument('whole', type=Path)
    parser.add_argument('before_plan', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('--concurrency', type=int, default=8, help='active shard processes in each load cohort')
    args = parser.parse_args()
    root = args.output.resolve(); root.mkdir(parents=True, exist_ok=False)
    cold, whole = args.cold.resolve(), args.whole.resolve()
    plan_path = LANE / 'shard-plan.json'
    plan = json.loads(plan_path.read_text())
    ledger = json.loads((cold / 'ledger.json').read_text())
    rows = ledger['units']
    definitions = {unit['name']: unit for unit in plan['units']}
    cold_times = {row['name']: row['seconds'] for row in rows}
    assert len(rows) == len(definitions) == len({row['name'] for row in rows}), 'incomplete cold ledger'
    assert {row['name'] for row in rows} == set(definitions), 'cold unit union'
    assert all(row['inputs_hash'] == definitions[row['name']]['inputs_hash'] for row in rows), 'cold inputs'
    assert all(row['exit'] == 0 and row['status'] == 'pass' for row in rows), 'cold verdict'
    require_budget(rows, 12)
    print(json.dumps(dict(event='cold proof', distribution=stats(rows))), flush=True)

    def run(label, command, env=None, timeout=180, allow_fail=False):
        with (root / 'commands.jsonl').open('a') as stream:
            stream.write(json.dumps(dict(label=label, command=command, timeout=timeout)) + '\n')
        with (root / (label + '.log')).open('w') as log:
            code = subprocess.run(command, cwd=LANE.parents[1], env=env,
                                  stdout=log, stderr=subprocess.STDOUT, timeout=timeout).returncode
        print(json.dumps(dict(event=label, exit=code)), flush=True)
        assert allow_fail or code == 0, label + ' failed'
        return code

    def measure(label, path, names, concurrency=1, load=0, limit=50, allow_fail=False):
        directory = root / label
        command = [sys.executable, '-B', str(LANE / 'measure_budget.py'), str(path), str(directory),
                   '--names', *names, '--concurrency', str(concurrency), '--load', str(load)]
        env = dict(os.environ, STAGE3_SHARD_SECONDS=str(limit), PYTHONDONTWRITEBYTECODE='1')
        run(label, command, env, math.ceil(len(names) / concurrency) * 65 + 30, allow_fail)
        result = json.loads((directory / 'ledger.json').read_text())
        return directory, result

    merge_command = [sys.executable, '-B', str(LANE / 'merge_shards.py'), str(cold), '--whole', str(whole),
                     '--output', str(root / 'merged.json')]
    run('merge', merge_command)
    merged = json.loads((root / 'merged.json').read_text())
    assert merged['tests'] == 106367 and merged['counts'] == {'pass': 106366, 'fail': 1, 'skip': 0}

    # Remove a real one-execution task, then rebind its local expectations. Only
    # the complete union and total checks can reject this internally green shard.
    dropped = copy.deepcopy(plan)
    selected = None
    for unit in dropped['units']:
        if 'grep' in unit or len(unit['tasks']) < 2 or unit['expect']['fail']:
            continue
        records = json.loads((cold / unit['name'] / 'observations.json').read_text())
        counts = Counter(task_key(row['task']) for row in records)
        task = next((task for task in unit['tasks'] if counts[task_key(task)] == 1), None)
        if task is not None:
            selected = unit, records, task
            break
    assert selected is not None
    unit, records, task = selected
    kept = [row for row in records if task_key(row['task']) != task_key(task)]
    unit['tasks'] = [item for item in unit['tasks'] if task_key(item) != task_key(task)]
    unit['expect']['pass'] -= 1
    unit['names_digest'] = names_digest(kept)
    drop_plan = root / 'dropped-case-plan.json'
    drop_plan.write_text(json.dumps(dropped, indent=2) + '\n')
    drop_results, drop_ledger = measure('dropped-case', drop_plan, [unit['name']], limit=30)
    for original in plan['units']:
        if original['name'] != unit['name']:
            (drop_results / original['name']).symlink_to(cold / original['name'], target_is_directory=True)
    rebound = json.loads((drop_results / 'plan.json').read_text())
    from merge_shards import merge
    previous_parser = os.environ.get('TSC_ADAPT_TYPESCRIPT')
    try:
        os.environ['TSC_ADAPT_TYPESCRIPT'] = str(Path(os.environ['STAGE3_ARTIFACT']) / 'api')
        rejected = merge(drop_results, whole, rebound)
    finally:
        if previous_parser is None:
            os.environ.pop('TSC_ADAPT_TYPESCRIPT', None)
        else:
            os.environ['TSC_ADAPT_TYPESCRIPT'] = previous_parser
    (root / 'dropped-case-merge.json').write_text(json.dumps(rejected, indent=2) + '\n')
    assert rejected['status'] == 'fail'
    assert any('executed test union differs' in error for error in rejected['errors'])
    assert not any('unit verdict rejected' in error or 'unit observation identity mismatch' in error
                   or 'unit input identity mismatch' in error for error in rejected['errors'])
    print(json.dumps(dict(event='CAUGHT dropped one case', unit=unit['name'], task=task,
                          errors=rejected['errors'])), flush=True)

    _, unloaded = measure('new-unloaded', plan_path, AFTER, concurrency=args.concurrency)
    by_name = {row['name']: row for row in unloaded['units']}
    load = 12
    attempts = []
    for attempt in range(6):
        _, loaded = measure(f'new-loaded-{attempt}-cpu-{load}', plan_path, AFTER,
                            concurrency=args.concurrency, load=load, allow_fail=True)
        ratio = statistics.median(row['seconds'] / by_name[row['name']]['seconds'] for row in loaded['units'])
        good = all(row['exit'] == 0 and row['seconds'] < 40 for row in loaded['units'])
        attempts.append(dict(load=load, median_slowdown_vs_unloaded_cohort=ratio,
                             median_slowdown_vs_alone=statistics.median(
                                 row['seconds'] / cold_times[row['name']] for row in loaded['units']),
                             distribution=stats(loaded['units']), good=good))
        print(json.dumps(dict(event='load calibration', **attempts[-1])), flush=True)
        if 2.2 <= ratio <= 2.8 and good:
            break
        if not good and ratio <= 2.8:
            load = max(1, round(load * .8))
        else:
            load = min(64, max(1, round(load * 1.5 / max(.1, ratio - 1))))
        tested = {item['load'] for item in attempts}
        if load in tested:
            load = min((value for value in range(1, 65) if value not in tested),
                       key=lambda value: (abs(value - load), value))
    else:
        raise AssertionError('could not calibrate the available CPU allocation to the requested load')
    _, before_unloaded = measure('before-unloaded', args.before_plan.resolve(), BEFORE,
                                 concurrency=args.concurrency, allow_fail=True)
    _, before_loaded = measure('before-loaded', args.before_plan.resolve(), BEFORE,
                               concurrency=args.concurrency, load=load, allow_fail=True)
    assert any(row['seconds'] > 50 and row['exit'] != 0 for row in before_loaded['units']), 'before did not exceed 50 s'
    assert any('timed out' in error for row in before_loaded['units'] for error in (row['errors'] or [])), 'before timeout'

    # A real empty upstream task measures process and harness startup. It is a
    # private calibration control, never added to the production manifest.
    empty_task = next(task for item in plan['units'] for task in item['tasks']
                      if task['runner'] == 'fourslash' and task['file'] == 'tests/cases/fourslash/fourslash.ts')
    empty = dict(name='stage3-lane-overhead-control', tasks=[empty_task],
                 expect={'pass': 0, 'fail': 0, 'skip': 0}, names_digest=names_digest([]), baseline_diffs=[])
    control = dict(plan, units=[empty], counts=empty['expect'])
    control_path = root / 'overhead-control-plan.json'; control_path.write_text(json.dumps(control, indent=2) + '\n')
    _, control_ledger = measure('overhead-control', control_path, [empty['name']], limit=30)
    overhead = control_ledger['units'][0]['seconds']
    times = {row['name']: row['seconds'] for row in rows}
    candidates = [definitions[f'stage3-lane-file-{index:03d}'] for index in range(20, 80)]
    number = 2
    while number < len(candidates) and sum(max(.01, times[item['name']] - overhead)
                                         for item in candidates[:number]) + overhead < 25:
        number += 1
    for attempt in range(5):
        chosen = candidates[:number]
        names = {item['name'] for item in chosen}
        observations = [row for item in chosen
                        for row in json.loads((cold / item['name'] / 'observations.json').read_text())]
        oversized = dict(name='stage3-lane-budget-mutant', tasks=[task for item in chosen for task in item['tasks']],
                         expect={'pass': len(observations), 'fail': 0, 'skip': 0},
                         names_digest=names_digest(observations), baseline_diffs=[])
        mutation = dict(plan, units=[item for item in plan['units'] if item['name'] not in names] + [oversized])
        assert {key: sum(item['expect'][key] for item in mutation['units']) for key in ('pass', 'fail', 'skip')} == plan['counts']
        mutation_path = root / f'budget-mutant-{attempt}-plan.json'
        mutation_path.write_text(json.dumps(mutation, indent=2) + '\n')
        _, solo = measure(f'budget-mutant-alone-{attempt}', mutation_path, [oversized['name']], limit=30, allow_fail=True)
        observed = solo['units'][0]
        if observed['exit'] == 0 and 23 <= observed['seconds'] <= 27:
            break
        if observed['exit']:
            number = max(2, number - max(1, number // 5))
        else:
            number = min(50, max(2, round(number * (25 - overhead) / max(.1, observed['seconds'] - overhead))))
    else:
        raise AssertionError('could not plant a roughly 25 s real-work shard')
    try:
        require_budget([observed], 12)
    except AssertionError as error:
        print(json.dumps(dict(event='CAUGHT alone budget', error=str(error))), flush=True)
    else:
        raise AssertionError('planted shard survived the alone budget')
    companions = [f'stage3-lane-file-{index:03d}' for index in range(100, 107)]
    _, mutant_loaded = measure('budget-mutant-loaded', mutation_path, [oversized['name'], *companions],
                              concurrency=args.concurrency, load=load, allow_fail=True)
    red = next(row for row in mutant_loaded['units'] if row['name'] == oversized['name'])
    assert red['exit'] != 0 and any('timed out' in error for error in (red['errors'] or [])), 'budget mutant survived'
    assert all(row['exit'] == 0 and row['seconds'] < 40 for row in mutant_loaded['units'] if row['name'] != oversized['name'])
    print(json.dumps(dict(event='CAUGHT budget mutant', alone=observed, loaded=red)), flush=True)
    affinity = ledger.get('affinity') or sorted(os.sched_getaffinity(0))
    quota, period = ledger['cpu_max'].split()
    capacity = min(len(affinity), float(quota) / float(period)) if quota != 'max' else len(affinity)
    limitation = None if capacity >= 8 and args.concurrency == 8 else (
        f'Available CPU capacity {capacity:g}, active shards {args.concurrency}; not the requested eight-CPU proof.')
    report = dict(status='pass', shard_count_before=len(json.loads(args.before_plan.read_text())['units']),
                  shard_count_after=len(plan['units']), cold=stats(rows), union=merged,
                  dropped_case=dict(unit=unit['name'], task=task, verdict=rejected),
                  hardware=dict(affinity=affinity, cpu_max=ledger['cpu_max'], active_shards=args.concurrency),
                  after_cohort=AFTER, before_cohort=BEFORE, calibration=attempts, load_processes=load,
                  slowdown_reference=f'Same {args.concurrency}-concurrent cohort without artificial load; serial ratios are also reported.',
                  after_unloaded=stats(unloaded['units']), after_loaded=stats(loaded['units']),
                  before_unloaded=stats(before_unloaded['units']), before_loaded=stats(before_loaded['units']),
                  mutant_alone=observed, mutant_loaded=red,
                  limitation=limitation)
    (root / 'proof.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(dict(event='proof complete', output=str(root / 'proof.json'))), flush=True)
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
