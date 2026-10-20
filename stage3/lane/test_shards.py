"""Hold exact case selection, gate schema and executed union to failing witnesses."""
from collections import Counter
from contextlib import ExitStack
import copy
import gzip
import io
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
            if row['file'] != shard_plan.WATCHER and row['file'] not in shard_plan.CASE_PARTITIONS:
                ownership.setdefault(row['file'], set()).add(row['unit'])
        self.assertTrue(all(len(units) == 1 for units in ownership.values()))

    def test_generated_plan_preserves_every_execution(self):
        plan = shard_plan.load_plan()
        with gzip.open(shard_plan.LANE / 'evidence/shards/tests-and-shards.jsonl.gz', 'rt') as stream:
            original = [json.loads(line) for line in stream]
        with gzip.open(shard_plan.LANE / 'evidence/case-shards/tests-and-units.jsonl.gz', 'rt') as stream:
            assigned = [json.loads(line) for line in stream]
        self.assertEqual(shard_plan.names_digest(assigned), shard_plan.names_digest(original))
        self.assertEqual(Counter(row['id'] for row in assigned), Counter(row['id'] for row in original))
        self.assertEqual({row['unit'] for row in assigned}, {unit['name'] for unit in plan['units']})
        by_unit = {}
        for row in assigned:
            by_unit.setdefault(row['unit'], []).append(row)
        for unit in plan['units']:
            rows = by_unit[unit['name']]
            self.assertEqual(unit['names_digest'], shard_plan.names_digest(rows))
            self.assertEqual(unit['expect'], {
                'pass': sum(row['status'] == 'pass' for row in rows),
                'fail': sum(row['status'] == 'fail' for row in rows), 'skip': 0})
            tasks = {(task['runner'], task['file']) for task in unit['tasks']}
            self.assertTrue({(row['task']['runner'], row['task']['file']) for row in rows} <= tasks)
            if 'grep' in unit:
                import re
                pattern = re.compile(unit['grep'])
                selected = [row['id'] for row in assigned if row['task']['runner'] == 'unittest'
                            and pattern.search(shard_plan.title(row['name']))]
                self.assertEqual(sorted(selected), sorted(row['id'] for row in rows))
                self.assertIsNone(pattern.search(shard_plan.title(rows[0]['name']) + ' additional case'))
        with gzip.open(shard_plan.LANE / 'evidence/shards/candidate-plan.json.gz', 'rt') as stream:
            candidate = json.load(stream)
        expected = Counter((row['task']['runner'], row['task']['file']) for row in candidate['tasks']
                           if row['file'] != shard_plan.WATCHER and row['file'] not in shard_plan.CASE_PARTITIONS)
        for row in candidate['tasks']:
            if row['file'] in shard_plan.CASE_PARTITIONS:
                expected[(row['task']['runner'], row['task']['file'])] += shard_plan.CASE_GROUPS
        expected.update((row['task']['runner'], row['task']['file']) for row in original
                        if row['file'] == shard_plan.WATCHER)
        self.assertEqual(Counter((task['runner'], task['file']) for unit in plan['units']
                                 for task in unit['tasks']), expected)

    def test_case_partitions_reject_invalid_inventory(self):
        for duplicate_task, short_titles, message in [
                (True, False, 'exactly one upstream task'),
                (False, True, 'too few selectable titles')]:
            tasks, records = [], []
            for index, file in enumerate(shard_plan.CASE_PARTITIONS):
                task = dict(runner='unittest', file=f'root-{index}')
                tasks.append(dict(file=file, task=task, seconds=1))
                for case in range(1 if short_titles and index == 0 else shard_plan.CASE_GROUPS):
                    records.append(dict(id=f'{index}/{case}', file=file, task=task,
                                        name=[task['file'], f'family-{case}', f'case-{case}'], status='pass'))
            if duplicate_task:
                tasks.append(dict(file=next(iter(shard_plan.CASE_PARTITIONS)),
                                  task=dict(runner='unittest', file='empty additional root'), seconds=0))
            streams = [io.StringIO(json.dumps(dict(tasks=tasks))),
                       io.StringIO(''.join(json.dumps(row) + '\n' for row in records))]
            with self.subTest(message=message), patch.object(shard_plan.gzip, 'open', side_effect=streams):
                with self.assertRaisesRegex(ValueError, message):
                    shard_plan.build_plan('tree', 'api')

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

    def run_observed_unit(self, root, code, grep):
        unit, record = self.valid_observation()
        unit.update(tasks=[record['task']], inputs_hash='inputs')
        if grep:
            unit['grep'] = '^one$'
        plan = dict(artifact='tree', units=[unit])
        clock = [100.0]

        def execute(command, tree, log, env, seconds):
            self.assertEqual(seconds, 29.5)
            clock[0] = 131.0
            log.write_text('  1 passing\n  0 failing\n  0 pending\n')
            if code != 124:
                if grep:
                    (log.parent / 'events.jsonl').write_text(json.dumps(record) + '\n')
                else:
                    worker = log.parent / 'worker'; worker.mkdir()
                    (worker / 'tasks.jsonl').write_text(json.dumps(dict(
                        task=record['task'], passes=[record], errors=[], passing=1)) + '\n')
            return code

        with ExitStack() as patches:
            for probe in (patch.object(shard, 'check_unit_inputs'),
                          patch.object(shard, 'adapted_inputs', return_value=(root, root, None, None)),
                          patch.object(shard, 'make_view'),
                          patch.object(shard, 'baseline_diff', return_value=([], '')),
                          patch.object(shard, 'read_failures', return_value=[]),
                          patch.object(shard, 'shard_seconds', return_value=30.0),
                          patch.object(shard.time, 'monotonic', side_effect=lambda: clock[0]),
                          patch.object(shard, 'execute', side_effect=execute)):
                patches.enter_context(probe)
            result = shard.run_unit(unit['name'], root, 'inputs', plan=plan)
        report = json.loads((root / unit['name'] / 'report.json').read_text())
        return result, report

    def test_run_past_limit_passes_and_records_seconds(self):
        for grep in (False, True):
            with self.subTest(grep=grep), tempfile.TemporaryDirectory() as scratch:
                result, report = self.run_observed_unit(Path(scratch), 0, grep)
                self.assertEqual(report['errors'], [])
                self.assertEqual(result, 0)
                self.assertEqual(report['status'], 'pass')
                self.assertEqual(report['counts'], {'pass': 1, 'fail': 0, 'skip': 0})
                self.assertEqual(report['seconds'], 31.0)

    def test_killed_at_limit_fails_naming_unit(self):
        for grep in (False, True):
            with self.subTest(grep=grep), tempfile.TemporaryDirectory() as scratch:
                result, report = self.run_observed_unit(Path(scratch), 124, grep)
                self.assertEqual(result, 1)
                self.assertEqual(report['status'], 'fail')
                self.assertEqual(report['errors'], ['one timed out after 31.0 seconds, past its 30 second limit'])

    def test_inventory_detects_swapped_case_with_unchanged_counts(self):
        unit, record = self.valid_observation()
        mutant = copy.deepcopy(record); mutant['name'][-1] = 'another baseline'
        self.assertEqual(shard.validate_observation(unit, unit['expect'], [mutant], [], [], 0),
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
