#!/usr/bin/env python3
"""Mutate real generator, wrapper and union inputs, restoring each after its proof."""
import copy
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
LOGS = Path(os.environ.get('STAGE3_PROOF_LOGS', '/workspace/scratch/hefhych-wave2/mutants'))


def failure(name, command, catcher):
    LOGS.mkdir(parents=True, exist_ok=True)
    log = LOGS / (name + '.log')
    with log.open('w') as output:
        result = subprocess.run(command, cwd=ROOT, stdout=output, stderr=subprocess.STDOUT, timeout=90)
    print(json.dumps(dict(name=name, command=command, exit=result.returncode, catcher=catcher)), flush=True)
    if result.returncode != 1 or catcher not in log.read_text():
        raise RuntimeError('mutant survived or failed for a different reason: ' + name)


def changed(name, path, edit, test, catcher):
    original = path.read_text()
    try:
        mutant = edit(original)
        assert mutant != original
        path.write_text(mutant)
        failure(name, ['go', 'test', './stage3/lane', '-run', '^' + test + '$', '-count=1', '-timeout', '90s', '-v'], catcher)
    finally:
        path.write_text(original)


def main():
    lane = ROOT / 'stage3/lane'
    generated = lane / 'shards_generated_test.go'
    changed('delete-generated-test', generated,
            lambda s: re.sub(r'func TestShard_\w+\(t \*testing.T\) \{[^}]*\}\s*', '', s, count=1),
            'TestShardsMatchManifest', 'generated shard declarations differ')
    changed('name-drift', generated, lambda s: s.replace('TestShard_Stage3LaneByteComparisons', 'TestShard_Drifted', 1),
            'TestShardsMatchManifest', 'generated shard declarations differ')
    with tempfile.TemporaryDirectory(prefix='hefhych-manifest-', dir='/workspace/scratch') as scratch:
        manifest = json.loads((lane / 'shards.json').read_text())
        row = copy.deepcopy(manifest['units'][-1]); row['name'] = 'stage3-lane-proof-added'
        manifest['units'].append(row)
        path = Path(scratch) / 'shards.json'; path.write_text(json.dumps(manifest))
        changed('added-manifest-unit', lane / 'shards_manifest_test.go',
                lambda s: s.replace('"--check")', '"--check", "--manifest", ' + json.dumps(str(path)) + ')', 1),
                'TestShardsMatchManifest', 'generated shard declarations differ')
    extra = lane / 'wave2_extra_mutant_test.go'
    assert not extra.exists()
    try:
        extra.write_text('package lane\nimport "testing"\nfunc TestShard_Extra(t *testing.T) {t.Parallel()}\n')
        failure('extra-Go-shard', ['go','test','./stage3/lane','-run','^TestShardsMatchManifest$','-count=1','-timeout','90s'], 'manifest/test names differ')
    finally:
        extra.unlink()
    changed('delete-Python-wrapper', lane / 'python_units_test.go',
            lambda s: re.sub(r'func TestStage3ApplyShard16\(t \*testing.T\) \{[^}]*\}\s*', '', s, count=1),
            'TestPythonUnitsMatchCommands', 'Python Go unit names differ')
    changed('added-old-command-case', lane / 'test_table.py',
            lambda s: s.replace("\n\nif __name__", '\n    def test_wave2_added_case(self):\n        pass\n\n\nif __name__', 1),
            'TestPythonUnitsMatchCommands', 'Python units differ from old command case discovery')
    changed('dropped-executed-Python-case', lane / 'python_units.py',
            lambda s: s.replace('    records = []', '    suite = unittest.TestSuite(list(suite)[:-1])\n    records = []', 1),
            'TestStage3LaneLaneTableTests', 'Python cases differ')
    changed('ignore-outcomes', lane / 'merge_shards.py', lambda s: s.replace("title(row['name']), row['status'])", "title(row['name']), 'pass')", 1),
            'TestUnionIdentityOutcomes', 'AssertionError')
    changed('ignore-multiplicity', lane / 'merge_shards.py',
            lambda s: s.replace('    def listing(counter):', '    actual = Counter({key: 1 for key in actual})\n    wanted = Counter({key: 1 for key in wanted})\n    def listing(counter):', 1),
            'TestUnionIdentityOutcomes', 'AssertionError')


if __name__ == '__main__':
    main()
