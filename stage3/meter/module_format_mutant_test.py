"""Keep the measured format visible; kill a report that erases its declaration."""
from pathlib import Path
import unittest
import json
import importlib.util
import os
import subprocess
import tempfile
from unittest.mock import patch
import report


class ModuleFormatTests(unittest.TestCase):
    def test_explicit_formats_and_historical_default(self):
        self.assertEqual(report.measured_format([{}]), 'EcmaScript')
        for value in ('CommonJSBundle', 'EcmaScript'):
            self.assertEqual(report.measured_format([{'module_format': value}]), value)
        for rows in ([{'module_format': 'typo'}], [{}, {'module_format': 'CommonJSBundle'}]):
            with self.assertRaises(ValueError):
                report.measured_format(rows)

    @unittest.skipUnless(os.environ.get('ENTRY_CENSUS_BINARY') and os.environ.get('CENSUS_BINARY'),
                         'requires prepared census probes')
    def test_entry_and_ordinary_record_declared_format(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            entry = root / 'entry.ts'
            entry.write_text('export function value(): number { return 1; }\n')
            for requested, expected in (('', 'CommonJSBundle'), ('EcmaScript', 'EcmaScript')):
                for variable in ('CENSUS_BINARY', 'ENTRY_CENSUS_BINARY'):
                    output = root / 'out.jsonl'
                    child = subprocess.run([os.environ[variable], str(entry), str(output)],
                                           env=dict(os.environ, CENSUS_MODULE_FORMAT=requested,
                                                    LATENT_ASSERT_NO_OUTPUT='1'),
                                           capture_output=True, timeout=30)
                    self.assertEqual(child.returncode, 0, child.stderr.decode())
                    header = json.loads(output.read_text().splitlines()[0])
                    self.assertEqual(header['module_format'], expected)

    @unittest.skipUnless(os.environ.get('ENTRY_CENSUS_BINARY') and os.environ.get('CENSUS_BINARY'),
                         'requires prepared census probes')
    def test_declared_format_reaches_real_lowering(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'core.ts').write_text('declare const require: unknown;\nexport function isNodeLikeSystem(): boolean { return typeof require !== "undefined"; }\n')
            entry = root / 'entry.ts'
            entry.write_text("import './core';\n")
            for requested, expected in (('', 'NotYet'), ('EcmaScript', 'accepted')):
                output = root / 'ordinary.jsonl'
                child = subprocess.run([os.environ['CENSUS_BINARY'], str(entry), str(output)],
                                       env=dict(os.environ, CENSUS_MODULE_FORMAT=requested),
                                       capture_output=True, timeout=30)
                self.assertEqual(child.returncode, 0, child.stderr.decode())
                record = json.loads(output.read_text().splitlines()[0])
                self.assertEqual(record['kind'], expected, record)
                if not requested:
                    self.assertIn('in a build declared one CommonJS bundle', record['reason'])
                output = root / 'entry.jsonl'
                child = subprocess.run([os.environ['ENTRY_CENSUS_BINARY'], str(entry), str(output)],
                                       env=dict(os.environ, CENSUS_MODULE_FORMAT=requested,
                                                LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1'),
                                       capture_output=True, timeout=30)
                self.assertEqual(child.returncode, 0, child.stderr.decode())
                records = [json.loads(line) for line in output.read_text().splitlines()]
                stops = [finding for row in records for finding in row.get('findings', [])
                         if 'in a build declared one CommonJS bundle' in finding.get('reason', '')]
                self.assertEqual(bool(stops), not bool(requested), records)

    def test_entry_template_changes_probe_key_and_kills_omission_mutant(self):
        path = Path(__file__).with_name('build-probes.py')
        spec = importlib.util.spec_from_file_location('format_build_probes', path)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        source = path.read_text()
        needle = ", ROOT / 'stage3/meter/entry.go.txt'"
        self.assertIn(needle, source)
        namespace = {'__file__': str(path), '__name__': 'probe_key_mutant'}
        exec(compile(source.replace(needle, ''), str(path), 'exec'), namespace)
        import types
        mutant = types.SimpleNamespace(**namespace)
        for candidate in (module, mutant):
            with tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                names = ('stage3/meter/build-probes.py', 'stage3/meter/entry_overlay.py',
                         'stage3/census/latent/make_overlay.py', 'stage3/meter/entry.go.txt')
                for name in names:
                    target = root / name
                    target.parent.mkdir(parents=True, exist_ok=True)
                    target.write_text('initial')
                globals_ = candidate.key.__globals__
                with patch.dict(globals_, ROOT=root, __file__=str(root / names[0])), patch.object(candidate.hook, 'action_key', return_value='compiler'):
                    before = candidate.key()
                    (root / names[-1]).write_text('changed module format metadata')
                    after = candidate.key()
                if candidate is module:
                    self.assertNotEqual(before, after)
                else:
                    with self.assertRaises(AssertionError):
                        self.assertNotEqual(before, after, 'template omission mutant survived')

    @staticmethod
    def assert_format_visible(module):
        summary = dict(module_format='CommonJSBundle', totals=dict(NotYet=2, Refused=3, error=0, panic=0),
                       boundaries=0, excluded_nested_functions=[], reason_rows=[], top_reasons=[],
                       panic_sites=[], error_sites=[], relation_report_complete=True)
        with patch.object(module, 'coverage_line', return_value=''), patch.object(module, 'tail_line', return_value=''):
            text = '\n'.join(module.latent_table('main', summary))
        assert 'Module format: CommonJSBundle' in text, text
        assert 'NotYet: 2; Refused: 3' in text, text

    def test_format_beside_numbers(self):
        self.assert_format_visible(report)

    def test_erased_report_format_mutant(self):
        path = Path(report.__file__)
        source = path.read_text()
        needle = " Module format: {summary.get('module_format', 'EcmaScript')}."
        self.assertIn(needle, source)
        mutant = source.replace(needle, '')
        namespace = {'__file__': str(path), '__name__': 'format_report_mutant'}
        exec(compile(mutant, str(path), 'exec'), namespace)
        import types
        namespace['coverage_line'] = lambda summary: ''
        namespace['tail_line'] = lambda summary: ''
        with self.assertRaises(AssertionError):
            self.assert_format_visible(types.SimpleNamespace(**namespace))
