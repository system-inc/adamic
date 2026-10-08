"""Integration checks against the unfiltered full census, using its real overlay."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


class ReplayTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.temporary = tempfile.TemporaryDirectory(prefix='census-replay-tests-')
        cls.scratch = Path(cls.temporary.name)
        cls.repository = Path(__file__).resolve().parents[4]
        cls.source = cls.scratch / 'source'
        cls.source.mkdir()
        (cls.source / 'probe.a').write_text('''import { value } from './dependency.a';
export function parent(text: string): void {
    const captured: number = value;
    const invalid: number = "checker refuses the parent";
    function nested(count: number): void {
        const binding: number = captured;
        if (text) { }
        if (count) { }
        const marker: string = "🌍"; var old = binding;
    }
}
export function sibling(): void { var other: number = value; }
''')
        (cls.source / 'dependency.a').write_text('export const value: number = 7;\n')
        cls.overlay = cls.scratch / 'overlay'
        cls.run_logged(['python3', 'stage3/census/latent/make_overlay.py',
                        str(cls.repository), str(cls.overlay)], 'overlay')
        cls.census = cls.scratch / 'census'
        cls.replay = cls.scratch / 'replay'
        for name, package in [('census', 'tool'), ('replay', 'replay/worker')]:
            cls.run_logged(['go', 'build', '-buildvcs=false',
                            '-overlay=' + str(cls.overlay / 'overlay.json'),
                            '-o', str(cls.scratch / name),
                            './stage3/census/latent/' + package], 'build-' + name)
        output = cls.scratch / 'full.jsonl'
        cls.run_logged([str(cls.census), str(cls.source), str(output)], 'full',
                       {'LATENT_FULL': '1', 'LATENT_ASSERT_NO_OUTPUT': '1'})
        rows = [json.loads(line) for line in output.read_text().splitlines()]
        cls.record = next(row for row in rows if row.get('file', '').endswith('probe.a'))
        cls.unit = next(unit for unit in cls.record['units'] if unit.get('name') == 'nested')
        assert cls.unit['depth'] == 1 and cls.unit['status'] == 'attempted', cls.unit
        parent = next(unit for unit in cls.record['units'] if unit.get('name') == 'parent')
        assert parent['status'] == 'split_checker_body', parent
        cls.expected = [finding for finding in cls.record['findings']
                        if finding['unit'] == cls.unit['where'] and finding['phase'] == 'lowering']
        reasons = {finding['reason'] for finding in cls.expected if finding['kind'] == 'Refused'}
        assert reasons == {'a string as a condition', 'a number as a condition', 'var'}, cls.expected
        assert not any(finding['kind'] == 'NotYet' for finding in cls.expected), cls.expected
        cls.target = next(finding for finding in cls.expected if finding['reason'] == 'var')

    @classmethod
    def tearDownClass(cls):
        cls.temporary.cleanup()

    @classmethod
    def run_logged(cls, command, name, extra=None, check=True):
        with (cls.scratch / (name + '.log')).open('w') as log:
            result = subprocess.run(command, cwd=cls.repository,
                                    env=dict(os.environ, **(extra or {})),
                                    stdout=log, stderr=subprocess.STDOUT)
        if check and result.returncode:
            raise AssertionError((cls.scratch / (name + '.log')).read_text())
        return result

    def replay_arguments(self):
        return [str(self.replay), '-project', str(self.source),
                '-where', self.target['where'], '-kind', self.target['kind'],
                '-reason', self.target['reason']]

    def test_nested_findings_match_full_census_in_order(self):
        result = self.run_logged(self.replay_arguments(), 'nested', check=False)
        log = (self.scratch / 'nested.log').read_text()
        self.assertEqual(result.returncode, 0, log)
        actual = json.loads(log.splitlines()[0])
        self.assertEqual([unit['where'] for unit in actual['units']], [self.unit['where']])
        self.assertEqual(actual['findings'], self.expected)
        self.assertIn('reproduced Refused: var', log)

    def test_parent_sibling_boundary_fails_signature_assertion(self):
        result = self.run_logged(self.replay_arguments(), 'parent-sibling',
                                 {'LATENT_REPLAY_MUTANT_PARENT_SIBLING': '1'}, check=False)
        log = (self.scratch / 'parent-sibling.log').read_text()
        self.assertEqual(result.returncode, 1, log)
        self.assertIn('replay signature did not reproduce', log)
        actual = json.loads(log.splitlines()[0])
        self.assertEqual([unit['name'] for unit in actual['units']], ['sibling'])
        self.assertNotIn(self.target, actual['findings'])


if __name__ == '__main__':
    unittest.main()
