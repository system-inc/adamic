#!/usr/bin/env python3
"""Run retained hash mutants in memory, requiring each named regression test to fail."""
import contextlib
import difflib
import io
import json
from pathlib import Path
import unittest

import shard
import shard_plan
import merge_shards
import test_shards

LANE = Path(__file__).resolve().parent
MUTANTS = [
    ('stale-runner-hash-mutant', shard,
     'return key, unit_inputs_hash(unit, plan, key, runtime)',
     "return key, 'cc6ef6522fd63209b4d826a8c7139554b07e40b429a52b0a7d80a7b2160951b4'",
     'test_tree_hash_reports_change_after_adaptation_edit'),
    ('accept-other-tree-mutant', merge_shards,
     " or report['artifact'] != key", '',
     'test_merge_rejects_report_key_from_another_tree'),
    ('artifact-in-plan-mutant', shard_plan,
     'return dict(version=1, api_digest=api_digest, runtime=runtime,',
     "return dict(version=1, artifact=__import__('artifact').input_key()[0], api_digest=api_digest, runtime=runtime,",
     'test_plan_bytes_ignore_adaptation_edit'),
]


def main():
    results = []
    for name, module, before, after, method in MUTANTS:
        source = Path(module.__file__).read_text()
        assert source.count(before) == 1, name
        mutated = source.replace(before, after, 1)
        expected = ''.join(difflib.unified_diff(source.splitlines(True), mutated.splitlines(True),
                           fromfile='a/stage3/lane/' + Path(module.__file__).name,
                           tofile='b/stage3/lane/' + Path(module.__file__).name))
        assert (LANE / (name + '.patch')).read_text() == expected, name + ' patch is stale'
        try:
            exec(compile(mutated, module.__file__, 'exec'), module.__dict__)
            stream = io.StringIO()
            with contextlib.redirect_stdout(stream):
                result = unittest.TextTestRunner(stream=stream, verbosity=2).run(
                    unittest.TestSuite([test_shards.ShardTests(method)]))
            print(stream.getvalue(), end='')
            assert len(result.failures) == 1 and not result.errors, name + ' survived or errored'
            results.append(dict(mutant=name + '.patch', test='test_shards.ShardTests.' + method, caught=True))
        finally:
            exec(compile(source, module.__file__, 'exec'), module.__dict__)
    print(json.dumps(results, indent=2))


if __name__ == '__main__':
    main()
