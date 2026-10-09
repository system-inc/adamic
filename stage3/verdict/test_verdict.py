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
from cases import parse, configurations, arguments, safe_name, layout, diagnostics, pretty_diagnostics
from run import baseline_suite, ROOT


class Comparisons(unittest.TestCase):
    def test_virtual_directory_mapping_mutants(self):
        with tempfile.TemporaryDirectory() as scratch:
            folder = Path(scratch)
            units = [{'name': '/a.ts', 'content': ''}, {'name': '../c.ts', 'content': ''}]
            settings = {'currentdirectory': '/project/nested', '__links': []}
            cwd, physical, options = layout(units, {}, folder, settings)
            self.assertEqual(physical('../c.ts'), folder / 'filesystem/project/c.ts')
            expected = b'/a.ts(1,1): error TS2322: bad\n'
            actual = b'../../a.ts(1,1): error TS2322: bad\n'
            self.assertEqual(diagnostics(actual, folder, cwd, units, settings), expected)
            for mutant in (actual.replace(b'a.ts', b'A.ts'), actual.replace(b'TS2322', b'TS2323')):
                self.assertNotEqual(diagnostics(mutant, folder, cwd, units, settings), expected)

    def test_pretty_wrapper_spacing_mutants(self):
        block = b'\x1b[96ma.ts\x1b[0m:\x1b[93m1\x1b[0m:\x1b[93m1\x1b[0m - \x1b[91merror\x1b[0m\x1b[90m TS2322: \x1b[0mbad\n\ncontext\n'
        actual = block + b'\n' + block + b'\n\nFound 2 errors.\n'
        self.assertEqual(pretty_diagnostics(actual), block + block)
        for mutant in (actual.replace(b'error', b'Error', 1), actual.replace(b'\n\ncontext', b'\ncontext', 1)):
            self.assertNotEqual(pretty_diagnostics(mutant), block + block)

    def test_configuration_identity_mutants(self):
        manifest = {'upstream_commit': PIN, 'total': 1, 'selected': 1, 'excluded': 0,
                    'configurations': 2, 'cases': [{'source': 'a.ts', 'configuration': 'target=es2015'},
                    {'source': 'a.ts', 'configuration': 'target=es2020'}], 'exclusions': [],
                    'configuration_exclusions': [{'source': 'a.ts', 'configuration': 'target=es5'}]}
        validate_manifest(manifest)
        for mutant in ('duplicate', 'overlap', 'unknown'):
            changed = copy.deepcopy(manifest)
            if mutant == 'duplicate':
                changed['cases'][1]['configuration'] = 'target=es2015'
            elif mutant == 'overlap':
                changed['configuration_exclusions'][0]['configuration'] = 'target=es2015'
            else:
                changed['configuration_exclusions'][0]['source'] = 'unknown.ts'
            with self.assertRaisesRegex(RuntimeError, 'invalid baseline census'):
                validate_manifest(changed)

    def test_external_package_guard(self):
        with tempfile.TemporaryDirectory() as scratch:
            parent = Path(scratch)
            (parent / 'package.json').write_text('{"type":"module"}')
            with self.assertRaisesRegex(RuntimeError, 'outside package metadata'):
                baseline_suite(ROOT / 'standins/empty.sh', parent, parent / 'baselines', None)

    def test_units_roots_and_directives(self):
        raw = b'// @target: ES2015\n// @lib: es2015,dom\n// @filename: a.ts\nexport const x=1;\n// @filename: folder/b.ts\nimport a=require("../a");\n'
        units, settings, roots = parse(raw, 'original.ts')
        self.assertEqual([u['name'] for u in units], ['a.ts', 'folder/b.ts'])
        self.assertEqual(units[0]['content'], 'export const x=1;')
        self.assertEqual(roots, ['folder/b.ts'])
        self.assertEqual(configurations(settings)[0][1]['lib'], ['es2015', 'dom'])
        # Mutant: passing all units instead of the upstream last-unit rule.
        with self.assertRaises(AssertionError):
            self.assertEqual([u['name'] for u in units], roots)
        no_reference = raw.replace(b'import a=require("../a");', b'const b=1;')
        self.assertEqual(parse(no_reference, 'original.ts')[2], ['a.ts', 'folder/b.ts'])

    def test_variants_and_alias_mutants(self):
        settings = {'target': 'es6,es2015,es2020', 'strict': 'true,false', 'lib': 'es2015,dom'}
        configs = configurations(settings)
        self.assertEqual(len(configs), 4)
        self.assertEqual(configs[0][0], 'strict=true,target=es6')
        self.assertEqual(len(configurations({'target': '*,-es3,!es5'})), 12)
        # Mutant: expansion counts the es6/es2015 alias twice.
        with self.assertRaises(AssertionError):
            self.assertEqual(6, len(configs))
        for value in ('/outside.ts', '../outside.ts', 'C:\\outside.ts'):
            with self.assertRaisesRegex(ValueError, 'rooted or escaping'):
                safe_name(value)

    def test_emit_options_are_preserved(self):
        argv = arguments({'declaration': True, 'noEmitOnError': True, 'lib': ['es2015']}, ['a.ts'])
        self.assertIn('--declaration', argv)
        self.assertNotIn('--noEmit', argv)
        self.assertNotIn('--ignoreDeprecations', argv)
        self.assertEqual(argv[-1], 'a.ts')
        # Mutants erase declaration or force noEmit, hiding emit diagnostics.
        with self.assertRaises(AssertionError):
            self.assertIn('--declaration', [a for a in argv if a != '--declaration'])
        with self.assertRaises(AssertionError):
            self.assertNotIn('--noEmit', argv + ['--noEmit', 'true'])

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
