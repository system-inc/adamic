#!/usr/bin/env python3
"""The three independent byte checks must each notice its own mutation."""
from pathlib import Path
import tempfile
import unittest
import copy
import re
from unittest.mock import patch
from census import diagnostic_sources
from run import PIN, compare, first_difference, validate_manifest, baseline_diagnostics


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


if __name__ == '__main__':
    unittest.main()
