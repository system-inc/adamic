#!/usr/bin/env python3
"""Audit complete case coverage, cold shard budgets and planted dispatch failures."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path)
    out = parser.parse_args().output.resolve()
    out.mkdir(parents=True, exist_ok=False)
    root = Path(__file__).resolve().parents[2]
    env = dict(os.environ, GOMAXPROCS='4', PYTHONDONTWRITEBYTECODE='1',
               ADAMIC_GATE_UNCACHED='1')
    rows = []
    for group in ['meter', 'census']:
        count = 8 if group == 'meter' else 2
        mapping = json.loads(subprocess.check_output(
            ['python3', 'stage3/meter/shards.py', group, '--list'], cwd=root,
            env=dict(env, ADAMIC_TEST_SHARD=f'0/{count}')))
        (out / (group + '-cases.json')).write_text(json.dumps(mapping, indent=2) + '\n')
        target = (next(name for name in mapping if '.CompilerSelectionTests.' in name)
                  if group == 'meter' else next(iter(mapping)))
        seen, baseline = [], {}
        for mutant in [False, True]:
            failures = []
            for index in range(count):
                label = f'{group}-{index}-{int(mutant)}'
                result_path = out / (label + '.json')
                command = ['python3', 'stage3/meter/shards.py', group,
                           '--result', str(result_path)]
                if mutant:
                    command += ['--plant-failure', target]
                start = time.monotonic()
                with (out / (label + '.log')).open('w') as log:
                    process = subprocess.run(command, cwd=root,
                        env=dict(env, ADAMIC_TEST_SHARD=f'{index}/{count}'),
                        stdout=log, stderr=subprocess.STDOUT)
                seconds = time.monotonic() - start
                data = json.loads(result_path.read_text())
                expected = [name for name in mapping if mapping[name] == index]
                assert sorted(data['cases']) == sorted(expected), label
                if group == 'meter':
                    assert data['ran'] == len(expected) and not data['skipped'], label
                assert seconds < 30, label
                if not mutant:
                    seen += data['cases']
                    baseline[index] = data
                    if group == 'meter':
                        expected_failures = ([
                            'entry_test.FullEntryCensusTests.test_rejected_entry_measures_clean_nested_body_and_catches_first_error_mutant'
                        ] if index == 0 else [])
                        assert data['failures'] == expected_failures and not data['errors'], label
                        if expected_failures:
                            assert 'AssertionError: 0 != 3' in (out / (label + '.log')).read_text(), label
                    else:
                        assert not data['failures'], label
                        assert data['errors'] == (['setUpClass (replay_test.ReplayTests)'] if expected else []), label
                        if expected:
                            assert "'status': 'panic'" in (out / (label + '.log')).read_text(), label
                else:
                    assert data['errors'] == baseline[index]['errors'], label
                    added = [target] if mapping[target] == index else []
                    assert sorted(data['failures']) == sorted(baseline[index]['failures'] + added), label
                failures += data['failures']
                rows.append(dict(unit=group, shard=f'{index}/{count}', mutant=mutant,
                    target=target, seconds=round(seconds, 3), exit=process.returncode, **data))
                (out / 'results.json').write_text(json.dumps(rows, indent=2) + '\n')
                print(label, round(seconds, 3), process.returncode, flush=True)
            if mutant:
                assert failures.count(target) == 1, group
        assert sorted(seen) == sorted(mapping), group
    for selector in ['bad', '0/0', '-1/8', '8/8', '0/2/3']:
        with (out / ('invalid-' + selector.replace('/', '_') + '.log')).open('w') as log:
            process = subprocess.run(['python3', 'stage3/meter/shards.py', 'meter'],
                cwd=root, env=dict(env, ADAMIC_TEST_SHARD=selector),
                stdout=log, stderr=subprocess.STDOUT)
        assert process.returncode == 2, selector


if __name__ == '__main__':
    main()
