"""Integration checks against the unfiltered full census, using its real overlay."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

repository = Path(__file__).resolve().parents[4]
products = {}


def built():
    """The census products, prepared once for the module: a gate's wrapper starts with an empty STAGE3_CACHE."""
    if not products:
        products.update(json.loads(subprocess.check_output(
            ['python3', 'stage3/meter/build-probes.py', '--prepare'], cwd=repository, text=True)))
    return products


class Scratch:
    @classmethod
    def make_scratch(cls, prefix):
        cls.temporary = tempfile.TemporaryDirectory(prefix=prefix)
        cls.scratch = Path(cls.temporary.name)

    @classmethod
    def tearDownClass(cls):
        cls.temporary.cleanup()

    @classmethod
    def run_logged(cls, command, name, extra=None, check=True):
        with (cls.scratch / (name + '.log')).open('w') as log:
            result = subprocess.run(command, cwd=repository,
                                    env=dict(os.environ, **(extra or {})),
                                    stdout=log, stderr=subprocess.STDOUT)
        if check and result.returncode:
            raise AssertionError((cls.scratch / (name + '.log')).read_text())
        return result


# On main these directory cases stop in setUpClass: the directory census routes every module from its first-sorted
# file, dependency.a, so probe.a's nested unit panics (#pn630c6).
class ReplayTests(Scratch, unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.make_scratch('census-replay-tests-')
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
        cls.census = Path(built()['latent'])
        cls.replay = Path(built()['replay'])
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
        # cdddfee346 made string and number conditions lower with JavaScript truthiness, so of the nested
        # body's three refusals only var's remains (full_audit.py's, #3dgsny7).
        reasons = {finding['reason'] for finding in cls.expected if finding['kind'] == 'Refused'}
        assert reasons == {'var'}, cls.expected
        assert not any(finding['kind'] == 'NotYet' for finding in cls.expected), cls.expected
        cls.target = next(finding for finding in cls.expected if finding['reason'] == 'var')

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


# A replay from an entry file rebuilds the program's files from the checker's resolved graph, whose order puts the
# entry's dependency before it. Lowering reads the first file as the entry and routes every module identity from it,
# so the entry must stay first (load.go.txt's LatentReplayEntryReach, as the entry census's LatentEntryReach does).
# With the dependency first, the entry, which the dependency doesn't import, panics "source module has no
# import-declaration route" (#4ft3dyf). The replayed unit lives in the entry, so its findings need the entry's route.
class EntryRootReplayTests(Scratch, unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.make_scratch('census-replay-entry-tests-')
        cls.source = cls.scratch / 'entry-root'
        cls.source.mkdir()
        cls.entry = cls.source / 'entry.a'
        cls.entry.write_text('''import { value } from './dependency.a';
export function parent(): void {
    const invalid: number = "checker refuses the parent";
    function nested(count: number): void {
        var old: number = count + value;
    }
}
''')
        (cls.source / 'dependency.a').write_text('export const value: number = 7;\n')
        cls.replay = Path(built()['replay'])
        output = cls.scratch / 'entry.jsonl'
        cls.run_logged([built()['entry'], str(cls.entry), str(output)], 'entry',
                       {'LATENT_FULL': '1', 'LATENT_ASSERT_NO_OUTPUT': '1'})
        rows = [json.loads(line) for line in output.read_text().splitlines()]
        header = rows[0]
        assert sorted(Path(source).name for source in header['sources']) == ['dependency.a', 'entry.a'], header
        cls.record = next(row for row in rows if row.get('file', '').endswith('entry.a'))
        cls.unit = next(unit for unit in cls.record['units'] if unit.get('name') == 'nested')
        assert cls.unit['depth'] == 1 and cls.unit['status'] == 'attempted', cls.unit
        cls.expected = [finding for finding in cls.record['findings']
                        if finding['unit'] == cls.unit['where'] and finding['phase'] == 'lowering']
        assert cls.expected, cls.record
        cls.target = next(finding for finding in cls.expected
                          if finding['kind'] == 'Refused' and finding['reason'] == 'var')

    def test_nested_unit_in_the_entry_replays_from_the_entry_file(self):
        result = self.run_logged([str(self.replay), '-project', str(self.entry),
                                  '-where', self.target['where'], '-kind', self.target['kind'],
                                  '-reason', self.target['reason']], 'entry-replay', check=False)
        log = (self.scratch / 'entry-replay.log').read_text()
        self.assertNotIn('no import-declaration route', log)
        self.assertEqual(result.returncode, 0, log)
        actual = json.loads(log.splitlines()[0])
        self.assertEqual([unit['where'] for unit in actual['units']], [self.unit['where']])
        self.assertEqual(actual['findings'], self.expected)
        self.assertIn('reproduced Refused: var', log)


if __name__ == '__main__':
    unittest.main()
