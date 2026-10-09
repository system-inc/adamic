"""Hold exact case selection, gate schema and executed union to failing witnesses."""
from collections import Counter
import copy
import gzip
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import shard
import shard_plan
import merge_shards


class ShardTests(unittest.TestCase):
    def test_manifest_exact_schema_and_counts(self):
        plan = shard_plan.load_plan()
        manifest = json.loads((shard_plan.LANE / 'shards.json').read_text())
        self.assertEqual(manifest, shard_plan.gate_manifest(plan))
        self.assertEqual(set(manifest), {'version', 'units'})
        self.assertEqual(len({unit['name'] for unit in manifest['units']}), len(manifest['units']))
        for unit in manifest['units']:
            self.assertEqual(set(unit), {'name', 'command', 'expect'})
            self.assertEqual(set(unit['expect']), {'pass', 'fail', 'skip'})
        self.assertEqual({key: sum(unit['expect'][key] for unit in manifest['units'])
                          for key in ('pass', 'fail', 'skip')}, plan['counts'])
        self.assertEqual([unit['name'] for unit in manifest['units'] if unit['expect']['fail']],
                         ['stage3-lane-byte-comparisons'])

    def test_watcher_grep_selects_exactly_one_case(self):
        plan = shard_plan.load_plan()
        with gzip.open(shard_plan.LANE / 'evidence/case-shards/tests-and-units.jsonl.gz', 'rt') as stream:
            rows = [json.loads(line) for line in stream]
        watcher = [row for row in rows if row['file'] == shard_plan.WATCHER]
        self.assertEqual(len(watcher), 16)
        import re
        for row in watcher:
            grep = shard_plan.exact_grep(row['name'])
            matches = [case['id'] for case in watcher if re.search(grep, shard_plan.title(case['name']))]
            self.assertEqual(matches, [row['id']])
            self.assertIsNone(re.search(grep, shard_plan.title(row['name']) + ' additional case'))
        self.assertEqual(len({row['unit'] for row in watcher}), 16)
        self.assertEqual(len({row['id'] for row in rows}), len(rows))
        self.assertEqual(Counter(row['status'] for row in rows), Counter({'pass': 106366, 'fail': 1}))
        ownership = {}
        for row in rows:
            if row['file'] != shard_plan.WATCHER:
                ownership.setdefault(row['file'], set()).add(row['unit'])
        self.assertTrue(all(len(units) == 1 for units in ownership.values()))

    def test_view_keeps_mutable_outputs_and_configuration_private(self):
        from view import make_view
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch); tree = root / 'tree'; tree.mkdir()
            for directory in ('tests/baselines/reference', 'tests/cases', 'built/local'):
                (tree / directory).mkdir(parents=True)
            (tree / 'Herebyfile.mjs').write_text('original build definition')
            (tree / 'test.config').write_text('stale artifact config')
            for name in ('run.js', 'run.js.map'):
                (tree / 'built/local' / name).write_text('immutable build')
            view = root / 'view'; make_view(tree, view)
            self.assertFalse((view / 'Herebyfile.mjs').is_symlink())
            self.assertFalse((view / 'test.config').exists())
            (view / 'test.config').write_text('private config')
            (view / 'built/local/run.js').write_text('private mutant')
            self.assertEqual((tree / 'test.config').read_text(), 'stale artifact config')
            self.assertEqual((tree / 'built/local/run.js').read_text(), 'immutable build')

    def valid_observation(self):
        record = dict(task=dict(runner='compiler', file='one.ts'), name=['', 'compiler', 'one.ts', 'baseline'], status='pass')
        unit = dict(name='one', expect={'pass': 1, 'fail': 0, 'skip': 0},
                    names_digest=shard_plan.names_digest([record]), baseline_diffs=[])
        return unit, record

    def test_budget_can_fail_with_all_other_checks_green(self):
        unit, record = self.valid_observation()
        self.assertEqual(shard.validate_observation(unit, unit['expect'], [record], [], [], 29, 0), [])
        errors = shard.validate_observation(unit, unit['expect'], [record], [], [], 30, 0)
        self.assertEqual(len(errors), 1)
        self.assertIn('invalid test unit', errors[0])

    def test_inventory_detects_swapped_case_with_unchanged_counts(self):
        unit, record = self.valid_observation()
        mutant = copy.deepcopy(record); mutant['name'][-1] = 'another baseline'
        self.assertEqual(shard.validate_observation(unit, unit['expect'], [mutant], [], [], 1, 0),
                         ['selected test identities do not equal this unit inventory'])

    def test_input_hash_guard_detects_changed_harness(self):
        unit, _ = self.valid_observation()
        plan = dict(artifact='tree', api_digest='api')
        with patch.object(shard, 'runtime_hash', return_value='original'):
            unit['inputs_hash'] = shard_plan.digest(dict(artifact='tree', api='api', runtime='original', unit=unit))
            shard.check_unit_inputs(unit, plan, unit['inputs_hash'])
        with patch.object(shard, 'runtime_hash', return_value='changed'):
            with self.assertRaisesRegex(ValueError, 'inputs hash'):
                shard.check_unit_inputs(unit, plan, unit['inputs_hash'])

    def test_merge_rejects_changed_union_despite_equal_counts(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch); whole = root / 'whole'; whole.mkdir()
            (whole / 'oracle').mkdir()
            unit = root / 'one'; unit.mkdir(); (unit / 'oracle').mkdir()
            original = dict(task=dict(runner='compiler', file='one.ts'), name=['baseline'], status='pass')
            mutant = copy.deepcopy(original); mutant['name'] = ['replacement baseline']
            plan_unit = dict(name='one', inputs_hash='inputs', expect={'pass': 1, 'fail': 0, 'skip': 0},
                             names_digest=shard_plan.names_digest([mutant]))
            plan = dict(artifact='tree', units=[plan_unit], counts=plan_unit['expect'])
            (unit / 'observations.json').write_text(json.dumps([mutant]))
            (unit / 'report.json').write_text(json.dumps(dict(name='one', inputs_hash='inputs', artifact='tree',
                status='pass', counts=plan_unit['expect'], seconds=1, failed_tests=[], baseline_diffs=[])))
            (unit / 'oracle/baseline.diff').write_text('')
            (whole / 'execution.json').write_text(json.dumps(dict(artifact='tree')))
            (whole / 'oracle/report.json').write_text(json.dumps(dict(counts=dict(passing=1, failing=0, pending=0))))
            (whole / 'oracle/baseline.diff').write_text('')
            (whole / 'whole-tasks.jsonl').write_text(json.dumps(dict(task=original['task'], passes=[original], errors=[])) + '\n')
            expected = root / 'expected.json'; expected.write_text(json.dumps(dict(failed_tests=[], baseline_diffs=[])))
            with patch.object(merge_shards, 'LANE', root), patch.object(merge_shards, 'check_results', return_value=dict(status='pass')):
                report = merge_shards.merge(root, whole, plan)
            self.assertEqual(report['errors'], ['executed test union differs from the unsharded lane'])
            self.assertEqual(report['status'], 'fail')

    def test_merge_rejects_newline_tampering_with_all_other_checks_green(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch); whole = root / 'whole'; whole.mkdir()
            (whole / 'oracle').mkdir()
            unit = root / 'one'; unit.mkdir(); (unit / 'oracle').mkdir()
            original = dict(task=dict(runner='compiler', file='one.ts'), name=['baseline'], status='pass')
            mutant = copy.deepcopy(original)
            plan_unit = dict(name='one', inputs_hash='inputs', expect={'pass': 1, 'fail': 0, 'skip': 0},
                             names_digest=shard_plan.names_digest([mutant]))
            plan = dict(artifact='tree', units=[plan_unit], counts=plan_unit['expect'])
            (unit / 'observations.json').write_text(json.dumps([mutant]))
            (unit / 'report.json').write_text(json.dumps(dict(name='one', inputs_hash='inputs', artifact='tree',
                status='pass', counts=plan_unit['expect'], seconds=1, failed_tests=[], baseline_diffs=[])))
            (unit / 'oracle/baseline.diff').write_bytes(b'--- reference/api/typescript.d.ts\r\n')
            (whole / 'execution.json').write_text(json.dumps(dict(artifact='tree')))
            (whole / 'oracle/report.json').write_text(json.dumps(dict(counts=dict(passing=1, failing=0, pending=0))))
            (whole / 'oracle/baseline.diff').write_bytes(b'--- reference/api/typescript.d.ts\n')
            (whole / 'whole-tasks.jsonl').write_text(json.dumps(dict(task=original['task'], passes=[original], errors=[])) + '\n')
            expected = root / 'expected.json'; expected.write_text(json.dumps(dict(failed_tests=[], baseline_diffs=[])))
            with patch.object(merge_shards, 'LANE', root), patch.object(merge_shards, 'check_results', return_value=dict(status='pass')):
                report = merge_shards.merge(root, whole, plan)
            self.assertEqual(report['errors'], ['merged baseline bytes differ from the unsharded lane'])
            self.assertEqual(report['status'], 'fail')


class SharedArtifactTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory(); self.addCleanup(self.scratch.cleanup)
        self.artifact = Path(self.scratch.name)
        (self.artifact / 'tree').mkdir(); (self.artifact / 'api').mkdir()
        (self.artifact / 'tree/compiler.js').write_text('original compiler')
        (self.artifact / 'api/typescript.js').write_text('original stock parser')
        self.inputs = dict(proof='original')
        self.ready = dict(key='tree', inputs=self.inputs,
            payload=shard.payload_digest(self.artifact / 'tree'),
            phases={phase: dict(exit=0) for phase in ('apply', 'install', 'build', 'harness')})
        self.plan = dict(artifact='tree', runtime='runtime', api_digest=shard.payload_digest(self.artifact / 'api'))
        self.write_ready()
        for probe in (patch.dict(shard.os.environ, STAGE3_ARTIFACT=str(self.artifact)),
                      patch.object(shard, 'input_key', return_value=('tree', self.inputs)),
                      patch.object(shard, 'runtime_hash', return_value='runtime')):
            probe.start(); self.addCleanup(probe.stop)
        shard.shared_artifact(self.plan)

    def write_ready(self):
        (self.artifact / 'ready.json').write_text(json.dumps(self.ready))

    def test_stock_parser_corruption_is_rejected(self):
        (self.artifact / 'api/typescript.js').write_text('wrong parser')
        with self.assertRaisesRegex(ValueError, 'stock API parser corruption'):
            shard.shared_artifact(self.plan)

    def test_failed_producer_phase_is_rejected(self):
        self.ready['phases']['build']['exit'] = 7; self.write_ready()
        with self.assertRaisesRegex(ValueError, 'producer phase failed'):
            shard.shared_artifact(self.plan)

    def test_changed_ready_input_identity_is_rejected(self):
        self.ready['inputs'] = dict(proof='changed'); self.write_ready()
        with self.assertRaisesRegex(ValueError, 'inputs hash'):
            shard.shared_artifact(self.plan)

    def test_payload_corruption_is_rejected_before_execution(self):
        (self.artifact / 'tree/compiler.js').write_text('wrong compiler')
        with self.assertRaisesRegex(ValueError, 'payload corruption'):
            shard.shared_artifact(self.plan)


if __name__ == '__main__':
    unittest.main()
