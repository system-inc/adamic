#!/usr/bin/env python3
"""The three independent byte checks must each notice its own mutation."""
from pathlib import Path
import os
import subprocess
import tempfile
import unittest
import copy
import re
from unittest.mock import patch
from census import diagnostic_sources
from run import PIN, compare, first_difference, validate_manifest, baseline_diagnostics, upstream_tree


class Comparisons(unittest.TestCase):
    def test_independent_stream_mutants(self):
        expected = {'stdout': b'f.ts(1,1): error TS2322: bad\n', 'stderr': b'', 'exit': b'2\n'}
        with tempfile.TemporaryDirectory() as scratch:
            folder = Path(scratch)
            for mutant in ('stdout', 'stderr', 'exit'):
                for stream, value in expected.items():
                    actual = value
                    if stream == mutant:
                        actual = value.replace(b'error', b'Error', 1) if stream == 'stdout' else (b'x' if stream == 'stderr' else b'1\n')
                    (folder / ('actual.' + stream)).write_bytes(actual)
                self.assertEqual(set(compare(folder, expected)), {mutant})

    def test_baseline_root_mapping_and_filename_mutant(self):
        folder = Path('/scratch/case')
        expected = b"case.ts(1,1): error TS2694: Namespace '\"case\".N' has no exported member 'X'.\n"
        actual = expected.replace(b'"case"', b'"/scratch/case/case"')
        self.assertEqual(baseline_diagnostics(actual, folder), expected)
        mutant = actual.replace(b'/scratch/case/case"', b'/scratch/case/casE"')
        self.assertNotEqual(baseline_diagnostics(mutant, folder), expected)
        unrelated = actual.replace(b'/scratch/case/', b'/scratch/other/')
        self.assertNotEqual(baseline_diagnostics(unrelated, folder), expected)

    def test_library_placeholder_census_mutant(self):
        summary = b"f.ts(1,1): error TS2374: duplicate.\nlib.es5.d.ts(--,--): error TS2374: duplicate.\n"
        wanted = ['f.ts', 'lib.es5.d.ts']
        self.assertEqual(diagnostic_sources(summary), wanted)
        old_guard = re.compile(rb'^([^\n]+?)\(\d+,\d+\):', re.M)
        with patch('census.DIAGNOSTIC_SOURCES', old_guard):
            with self.assertRaises(AssertionError):
                self.assertEqual(diagnostic_sources(summary), wanted)

    def test_census_mutants(self):
        manifest = {'upstream_commit': PIN, 'total': 2, 'selected': 1, 'excluded': 1,
                    'cases': [{'source': 'a.ts'}], 'exclusions': [{'source': 'b.ts'}]}
        validate_manifest(manifest)
        for mutation in ('pin', 'selected', 'excluded', 'duplicate'):
            changed = copy.deepcopy(manifest)
            if mutation == 'pin':
                changed['upstream_commit'] = 'wrong'
            elif mutation == 'duplicate':
                changed['exclusions'][0]['source'] = 'a.ts'
            else:
                changed[mutation] += 1
            with self.assertRaisesRegex(RuntimeError, 'invalid baseline census'):
                validate_manifest(changed)

    def test_eof_difference(self):
        self.assertEqual(first_difference(b'abc', b'ab')['byte_offset'], 2)
        self.assertEqual(first_difference(b'', b'x')['byte_offset'], 0)


class UpstreamSource(unittest.TestCase):
    """The baselines read a provided checkout at the pin, never one cloned from a mirror or the network."""

    def head(self, stdout):
        return patch('run.subprocess.run', return_value=subprocess.CompletedProcess([], 0, stdout=stdout))

    def test_neither_source_refuses_and_runs_nothing(self):
        # Mutant: the clone restored, which runs git against the network instead of refusing.
        environment = {key: value for key, value in os.environ.items()
                       if key not in ('STAGE3_VERDICT_UPSTREAM', 'ADAMIC_TYPESCRIPT_SOURCE')}
        with patch.dict(os.environ, environment, clear=True), self.head(PIN + '\n') as ran:
            with self.assertRaisesRegex(RuntimeError, 'ADAMIC_TYPESCRIPT_SOURCE.*never clones'):
                upstream_tree()
            ran.assert_not_called()

    def test_the_gate_inputs_checkout_is_read_and_the_verdicts_own_wins(self):
        with tempfile.TemporaryDirectory() as gate, tempfile.TemporaryDirectory() as own:
            with patch.dict(os.environ, {'ADAMIC_TYPESCRIPT_SOURCE': gate}), self.head(PIN + '\n') as ran:
                os.environ.pop('STAGE3_VERDICT_UPSTREAM', None)
                self.assertEqual(upstream_tree(), Path(gate).resolve())
                self.assertEqual(ran.call_args.args[0], ['git', '-C', str(Path(gate).resolve()), 'rev-parse', 'HEAD'])
            with patch.dict(os.environ, {'ADAMIC_TYPESCRIPT_SOURCE': gate, 'STAGE3_VERDICT_UPSTREAM': own}), \
                    self.head(PIN + '\n'):
                self.assertEqual(upstream_tree(), Path(own).resolve())

    def test_a_checkout_off_the_pin_refuses(self):
        with tempfile.TemporaryDirectory() as gate, patch.dict(os.environ, {'ADAMIC_TYPESCRIPT_SOURCE': gate}):
            os.environ.pop('STAGE3_VERDICT_UPSTREAM', None)
            with self.head('0' * 40 + '\n'), self.assertRaisesRegex(RuntimeError, 'not the pinned ' + PIN):
                upstream_tree()
            # Not a checkout at all: git answers nothing, and the refusal says so. The ceiling keeps git from finding
            # a checkout the temporary directory happens to sit inside.
            with patch.dict(os.environ, {'GIT_CEILING_DIRECTORIES': str(Path(gate).resolve().parent)}), \
                    self.assertRaisesRegex(RuntimeError, 'is at no commit'):
                upstream_tree()


if __name__ == '__main__':
    unittest.main()
