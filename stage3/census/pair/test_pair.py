import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

import pair

HERE = Path(__file__).resolve().parent
TS = HERE.parents[1] / 'api/node_modules/typescript'
REASON = 'an object refinement using an open numeric enum as a literal tag'
HEADER = dict(status='measurement', latent_mode='first-error')


class PairTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.witnesses = {}
        for label in ('qeanycw', 'kind-predicates'):
            root = HERE / 'testdata' / label
            base, head = (pair.load(root / f'{side}.jsonl.gz') for side in ('base', 'head'))
            cls.witnesses[label] = pair.compare(base, head, root / 'tree', TS)

    def test_qeanycw_distinct_reads_cleared_with_no_new(self):
        d = self.witnesses['qeanycw']['distinct_reads']
        self.assertEqual(len(d['new_sites']), 0)
        self.assertEqual(d['per_reason'][REASON]['lowered'], 70)
        self.assertEqual(sum(d['per_reason'][REASON]['moved_by_reason'].values()), 2)
        self.assertEqual(len(d['new_findings']), 0)
        self.assertEqual(len(d['regressions']), 0)

    def test_kind_predicates_239_lowered_zero_regressions(self):
        witness = self.witnesses['kind-predicates']
        self.assertEqual(witness['raw_findings']['per_reason'][REASON]['lowered'], 239)
        d = witness['distinct_reads']
        self.assertEqual(d['per_reason'][REASON]['lowered'], 181)
        self.assertEqual(len(d['regressions']), 0)

    def test_normalization_off_mutant_turns_qeanycw_zero_new_red_with_54(self):
        root = HERE / 'testdata/qeanycw'
        mutant = pair.compare(pair.load(root / 'base.jsonl.gz'), pair.load(root / 'head.jsonl.gz'),
                              root / 'tree', TS, normalize=False)['distinct_reads']
        self.assertEqual(len(mutant['new_sites']), 54)
        # Run the unchanged positive test under the mutant, and require its
        # named zero-new check to fail with the historical 54-site miscount.
        probe = subprocess.run(['python3', '-E', '-S', '-B', '-c', '''
import sys, unittest, pair, test_pair
real_compare = pair.compare
pair.compare = lambda *a, **kw: real_compare(*a, **dict(kw, normalize=False))
suite = unittest.defaultTestLoader.loadTestsFromName(
    'test_pair.PairTests.test_qeanycw_distinct_reads_cleared_with_no_new')
result = unittest.TextTestRunner().run(suite)
sys.exit(not result.wasSuccessful())
'''], cwd=HERE, capture_output=True, text=True)
        self.assertEqual(probe.returncode, 1, probe.stderr)
        self.assertIn('test_qeanycw_distinct_reads_cleared_with_no_new', probe.stderr)
        self.assertIn('54 != 0', probe.stderr)

    def synthetic(self, text, before, after, units=None):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'input.ts').write_text(text)
            def rows(findings):
                fs = [dict(kind='Refused', where=f'input.ts:1:{column}', reason=reason,
                           text=reason, phase=phase, unit='input.ts:1:1')
                      for column, reason, phase in findings]
                return [HEADER, dict(file='input.ts', units=units or [], findings=fs)]
            return pair.compare(rows(before), rows(after), root, TS)

    def test_access_names_chains_elements_utf16_and_separate_reads(self):
        text = '/* 😀 */ a.b["x"].c; other.c;'
        col = lambda token: len(text[:text.index(token)].encode('utf-16-le')) // 2 + 1
        a, b, c, other = (col(token) for token in ('a.b', 'b[', 'c;', 'other'))
        findings = [(a, 'r', 'refusal_scan'), (b, 'r', 'refusal_scan'),
                    (c, 'r', 'refusal_scan'), (other, 'r', 'refusal_scan'),
                    (b, 'different reason', 'refusal_scan')]
        d = self.synthetic(text, findings, findings)
        self.assertEqual(d['raw_findings']['totals']['before'], {'Refused': 5})
        self.assertEqual(d['distinct_reads']['totals']['before'], {'Refused': 3})
        self.assertEqual(d['distinct_reads']['per_reason']['r']['before'], 2)

    def test_raw_dedup_across_attempts_and_moves_by_reason(self):
        d = self.synthetic('a.b;', [(1, 'old', 'refusal_scan')] * 2,
                           [(3, 'new', 'refusal_scan')])
        self.assertEqual(d['raw_findings']['totals']['before'], {'Refused': 1})
        self.assertEqual(d['distinct_reads']['per_reason']['old']['moved_by_reason'], {'new': 1})
        self.assertEqual(len(d['distinct_reads']['lowered']), 0)
        self.assertEqual(len(d['distinct_reads']['new_sites']), 0)

    def test_only_lowering_phase_changes_unit_first_stop(self):
        unit = dict(where='input.ts:1:1', kind='function', status='attempted')
        d = self.synthetic('a.b; c.d;', [(1, 'scan', 'refusal_scan'), (6, 'old', 'lowering')],
                           [(1, 'changed scan', 'refusal_scan'), (8, 'new', 'lowering')], [unit])
        stops = d['distinct_reads']['unit_first_stop_delta']
        self.assertEqual(len(stops), 1)
        self.assertEqual(stops[0]['base']['reason'], 'old')
        self.assertEqual(stops[0]['head']['reason'], 'new')
        self.assertEqual(len(d['distinct_reads']['regressions']), 0)

    def test_panic_is_retained_when_read_reason_is_unchanged(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'input.ts').write_text('a.b;')
            before = dict(where='input.ts:1:1', kind='Refused', reason='r', text='r',
                          phase='refusal_scan', unit='input.ts:1:1')
            after = dict(before, where='input.ts:1:3', kind='panic')
            rows = lambda f: [HEADER, dict(file='input.ts', units=[], findings=[f])]
            d = pair.compare(rows(before), rows(after), root, TS)['distinct_reads']
            self.assertEqual(d['totals']['after'], {'panic': 1})
            self.assertEqual(len(d['panics']['after']), 1)
            self.assertEqual([r['type'] for r in d['regressions']], ['new_panic'])

    def test_fixture_stream_hashes_match_provenance(self):
        for name in self.witnesses:
            root = HERE / 'testdata' / name
            manifest = json.loads((root / 'provenance.json').read_text())
            for side in ('base', 'head'):
                payload = gzip.decompress((root / f'{side}.jsonl.gz').read_bytes())
                self.assertEqual(hashlib.sha256(payload).hexdigest(),
                                 manifest['streams'][side]['fixture_sha256'])

    def test_call_arguments_are_separate_access_chains(self):
        d = self.synthetic('fn(a.b).c;', [(4, 'r', 'refusal_scan'), (1, 'r', 'refusal_scan')],
                           [(6, 'r', 'refusal_scan'), (9, 'r', 'refusal_scan')])
        self.assertEqual(d['distinct_reads']['per_reason']['r']['before'], 2)
        self.assertEqual(d['distinct_reads']['per_reason']['r']['after'], 2)
        self.assertEqual(len(d['distinct_reads']['new_findings']), 0)

    def test_cli_writes_json_and_side_by_side_report(self):
        root = HERE / 'testdata/qeanycw'
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            subprocess.run(['python3', '-B', str(HERE / 'pair.py'), str(root / 'base.jsonl.gz'),
                            str(root / 'head.jsonl.gz'), str(root / 'tree'), '--typescript', str(TS),
                            '--json', str(output / 'delta.json'), '--report', str(output / 'report.md')],
                           check=True, capture_output=True, text=True)
            self.assertIn('| Raw findings | Distinct reads |', (output / 'report.md').read_text())
            delta = json.loads((output / 'delta.json').read_text())
            self.assertEqual(delta['distinct_reads']['per_reason'][REASON]['lowered'], 70)


if __name__ == '__main__':
    unittest.main()
