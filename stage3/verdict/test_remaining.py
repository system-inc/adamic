"""Filesystem fixtures and exclusive mutants for the final exclusion expansion."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest.mock import patch
from cases import parse, configurations, layout
from library_summary import project
from project_scope import append_files
import resources


class RemainingChecks(unittest.TestCase):
    def test_project_scope_locations_and_comment_mutant(self):
        text = '{\n"compilerOptions": {"baseUrl": "/"},\n} // end\n/* shell: echo "{ wrong }" */\n'
        result = append_files(text, ['/test.ts'])
        self.assertEqual(result[:text.index('} // end')], text[:text.index('} // end')])
        self.assertLess(result.index('"files":'), result.index('} // end'))
        self.assertIn('/* shell: echo "{ wrong }" */', result)
        mutant = text[:text.rfind('}')] + '"files": []' + text[text.rfind('}'):]
        self.assertGreater(mutant.index('"files":'), mutant.index('} // end'))
        with self.assertRaises(RuntimeError):
            append_files('{"unfinished":', [])

    def test_absolute_source_is_preserved(self):
        raw = b'// @filename: /index.ts\n/// <reference path="/.lib/react.d.ts" />\nconst x=1;\n'
        units, settings, roots = parse(raw, 'input.ts')
        self.assertTrue(settings['__namespace'])
        self.assertIn('"/.lib/react.d.ts"', units[0]['content'])
        with tempfile.TemporaryDirectory() as scratch:
            cwd, physical, options = layout(units, {}, Path(scratch), settings)
            self.assertEqual(physical('/index.ts'), Path(scratch) / 'filesystem/index.ts')
            self.assertEqual(cwd, Path(scratch) / 'filesystem/.src')

    def test_drive_paths_and_canonical_alias_mutant(self):
        raw = b'// @filename: C:/dir/a.ts\nconst x=1;\n'
        units, settings, roots = parse(raw, 'input.ts')
        self.assertTrue(settings['__drives'])
        with tempfile.TemporaryDirectory() as scratch:
            _, physical, _ = layout(units, {}, Path(scratch), settings)
            self.assertEqual(physical(roots[0]), Path(scratch) / 'filesystem/.src/C:/dir/a.ts')
        collision = b'// @useCaseSensitiveFileNames: false\n// @filename: A:/foo/a.ts\nconst x=1;\n// @filename: a:/foo/b.ts\nconst y=1;'
        with self.assertRaisesRegex(ValueError, 'canonical path aliases'):
            parse(collision, 'input.ts')

    def test_output_path_is_temporary(self):
        raw = b'// @allowJs: true\n// @suppressOutputPathCheck: true\n// @filename: a.js\nconst x=1;'
        units, settings, _ = parse(raw, 'input.ts')
        options = configurations(settings)[0][1]
        with tempfile.TemporaryDirectory() as scratch:
            cwd, _, actual = layout(units, options, Path(scratch), settings)
            self.assertEqual(actual['outDir'], str(cwd / '.verdict-emit'))
            self.assertNotIn('outDir', options)

    def test_library_order_and_byte_mutants(self):
        source = b'a.ts(1,1): error TS2300: duplicate.\n'
        library = b'/compiler/lib.es5.d.ts(33,4): error TS2300: duplicate.\n'
        expected = source + b'lib.es5.d.ts(--,--): error TS2300: duplicate.\n'
        self.assertEqual(project(library + source, ['lib.es5.d.ts']), expected)
        for mutant in [library.replace(b'error', b'Error', 1) + source,
                       library.replace(b'2300', b'2301') + source,
                       library.replace(b'lib.es5', b'lib.es6') + source]:
            self.assertNotEqual(project(mutant, ['lib.es5.d.ts']), expected)
        self.assertEqual(project(b'prefix\n' + library, ['lib.es5.d.ts']), b'prefix\n' + library)

    def test_resource_byte_mutants(self):
        # Mutate copies of real pinned inputs, never a tree serving another run.
        tree = Path(os.environ.get('STAGE3_VERDICT_UPSTREAM', '/tmp/stage3-verdict-auto-upstream/upstream'))
        if not tree.exists():
            self.skipTest('set STAGE3_VERDICT_UPSTREAM to the pinned checkout for input-hash mutants')
        with tempfile.TemporaryDirectory() as scratch:
            base = Path(scratch)
            shutil.copytree(resources.ROOT, base / 'resources')
            shutil.copytree(tree / 'tests/lib', base / 'tree/tests/lib')
            manifest = json.loads((base / 'resources/manifest.json').read_text())
            with patch.object(resources, 'ROOT', base / 'resources'):
                good = base / 'good';good.mkdir()
                resources.prepare(base / 'tree', good)
                pairs = [(base / 'tree/tests/lib' / next(iter(manifest['test_libs'])), 'test library'),
                         (base / 'resources' / manifest['declarations'][0]['stored'], 'declaration')]
                for index, (file, catch) in enumerate(pairs):
                    raw = file.read_bytes();file.write_bytes(raw + b'\n')
                    output = base / str(index);output.mkdir()
                    with self.assertRaisesRegex(RuntimeError, catch + ' hash mismatch'):
                        resources.prepare(base / 'tree', output)
                    file.write_bytes(raw)


if __name__ == '__main__':
    unittest.main()
