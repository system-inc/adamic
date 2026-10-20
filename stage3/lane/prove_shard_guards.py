#!/usr/bin/env python3
"""Run isolated shard-plan and generated-declaration guard mutants without editing source files."""
import copy
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

import shard_plan
import test_shards


def caught(name, method):
    print('MUTANT ' + name, flush=True)
    result = unittest.TextTestRunner(verbosity=2).run(
        unittest.TestSuite([test_shards.ShardTests(method)]))
    if len(result.failures) != 1 or result.errors:
        raise AssertionError(name + ' was not caught by its intended assertion')
    print('CAUGHT ' + name, flush=True)


def main():
    source = Path(shard_plan.__file__).read_text()
    for name, predicate in [('multiple-task-guard', 'if len(tasks) != 1:'),
                            ('too-few-titles-guard', 'if len(titles) < CASE_GROUPS:')]:
        try:
            assert source.count(predicate) == 1
            exec(compile(source.replace(predicate, 'if False:', 1), shard_plan.__file__, 'exec'),
                 shard_plan.__dict__)
            caught(name, 'test_case_partitions_reject_invalid_inventory')
        finally:
            exec(compile(source, shard_plan.__file__, 'exec'), shard_plan.__dict__)

    plan = shard_plan.load_plan()
    with patch.object(shard_plan, 'exact_grep', side_effect=lambda name: re.escape(shard_plan.title(name))):
        unanchored, _ = shard_plan.build_plan(plan['artifact'], plan['api_digest'])
    with patch.object(shard_plan, 'load_plan', return_value=unanchored):
        caught('unanchored-case-groups', 'test_generated_plan_preserves_every_execution')

    missing = copy.deepcopy(plan)
    unit = next(unit for unit in missing['units']
                if unit['name'].startswith('stage3-lane-file-') and len(unit['tasks']) > 1)
    removed = unit['tasks'].pop()
    print('Removed planned task: ' + str(removed), flush=True)
    with patch.object(shard_plan, 'load_plan', return_value=missing):
        caught('missing-planned-task', 'test_generated_plan_preserves_every_execution')

    with tempfile.TemporaryDirectory(prefix='lane-generated-mutant-') as scratch:
        output = Path(scratch) / 'shards_generated_test.go'
        output.write_bytes((shard_plan.LANE / 'shards_generated_test.go').read_bytes().replace(
            b't.Parallel()', b'// missing parallel declaration', 1))
        code = subprocess.run([sys.executable, '-B', str(shard_plan.LANE / 'generate_shards.py'),
                               '--check', '--output', str(output)], timeout=20).returncode
        if code != 1:
            raise AssertionError('stale generated declaration survived')
    print('CAUGHT stale-generated-declaration', flush=True)
    assert shard_plan.runtime_hash() == plan['runtime']
    print('Five guard mutants caught; source and production plan unchanged', flush=True)
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
