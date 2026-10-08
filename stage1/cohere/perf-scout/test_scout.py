import importlib.util
import os
from pathlib import Path
import tempfile
import unittest
import scout

class Guards(unittest.TestCase):
    def test_one_byte_mutant(self):
        with self.assertRaises(RuntimeError): scout.checked(b'ok\ta\n',b'ok\tb\n')
    def test_drop_row_mutant(self):
        with self.assertRaises(RuntimeError): scout.checked(b'case 0\n',b'case 0\ncase 1\n')
    def test_reorder_mutant(self):
        with self.assertRaises(RuntimeError): scout.checked(b'case 1\ncase 0\n',b'case 0\ncase 1\n')
    def test_empty_input(self):
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'empty';p.write_text('\n')
            with self.assertRaises(ValueError):scout.rows(p)
    def test_accounting_mutant(self):
        spec=importlib.util.spec_from_file_location('accounting',scout.REPO/'stage1/typescript/scanner/profile.py')
        mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'callgrind';p.write_text('events: Ir\nfn=work\n1 7\nsummary: 7\n')
            self.assertEqual(mod.summarize(p)['total'],7)
            p.write_text('events: Ir\nfn=work\n1 7\nsummary: 8\n')
            with self.assertRaises(RuntimeError):mod.summarize(p)
    def test_skipped_mutant(self):
        scout.no_skips(b'fixed\t// skipped work and refused requests\n')
        for marker in (b'skipped rule\n', b'refused checker file 0 0 reason\n'):
            with self.assertRaises(ValueError): scout.no_skips(marker)
    def test_allocation_mutant(self):
        text='adamic: counts: allocations 7 frees 7 retains 2 releases 9 peak 3 regions 0\n'
        self.assertEqual(scout.balanced_counts(text)['frees'],7)
        with self.assertRaises(ValueError):scout.balanced_counts(text.replace('frees 7','frees 6'))
    def test_saved_counters(self):
        root=Path(os.environ['ADAMIC_SCOUT_RESULTS'])
        import json
        for name in ('measurements','encoding-measurements'):
            for workload in json.loads((root/(name+'.json')).read_text())['workloads'].values():
                scout.balanced_counts(workload['allocation_counts'])
        report=json.loads((root/'fused.json').read_text())
        for side in ('before','after'):scout.balanced_counts(report[side+'_counts'])
    def test_saved_answers(self):
        root=Path(os.environ['ADAMIC_SCOUT_RESULTS'])
        for name in ('owned','upstream','compiler','public','repository','printer'):
            p=root/'measure'/name
            want=(p/'Go.answer').read_bytes()
            scout.no_skips(want)
            scout.checked((p/'native.answer').read_bytes(),want)
            scout.checked((p/'Node.answer').read_bytes(),want)
            with self.assertRaises(RuntimeError):scout.checked(want[:-1],want)
