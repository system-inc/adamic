"""Arithmetic witnesses and deliberately broken implementations of each operation."""
import copy
import os
from pathlib import Path
import random
import types
import unittest

import hidden

# Mutants run the same tests and must exit nonzero at the intended assertion.
mutant = os.environ.get('HIDDEN_MUTANT', '')
if mutant:
    source = Path(hidden.__file__).read_text()
    replacements = {
        'ignore-dependency': ("if finding['kind'] == 'SkippedDependency':", "if False and finding['kind'] == 'SkippedDependency':"),
        'ignore-skipped': ("blocked[name].append(dict(start=start, end=end, category='checker',", "continue\n            blocked[name].append(dict(start=start, end=end, category='checker',"),
        'double-count': ('def union(spans):\n    result = []', 'def union(spans):\n    return sorted(spans)\n    result = []'),
        'forget-checker-child': ("if external['start'] <= span[0] and span[1] <= external['end']", "if False and external['start'] <= span[0] and span[1] <= external['end']"),
        'forget-independent': ("examined[name].extend(subtract([(external['start'], external['end'])], cuts))", 'pass  # mutant drops independent coverage'),
    }
    old, new = replacements[mutant]
    if source.count(old) != 1:
        raise RuntimeError('mutant insertion anchor changed')
    module = types.ModuleType('mutated_hidden')
    exec(compile(source.replace(old, new), hidden.__file__, 'exec'), module.__dict__)
    hidden = module

ROOT = Path('/source/compiler')
WHERE = '/source/compiler/probe.a:1:1'


def ledger():
    return [dict(latent_mode='full'), dict(file=str(ROOT / 'probe.a'), units=[], findings=[])], {
        'probe.a': dict(bytes=100, sha256='synthetic', units=[])}


def add_skip(rows, stock, start, end, line=1):
    where = f'/source/compiler/probe.a:{line}:1'
    rows[1]['units'].append(dict(where=where, status='split_checker_body', body_start=start,
        body_end=end, kind='KindFunctionDeclaration', checker_diagnostics=['deliberate checker diagnostic']))
    stock['probe.a']['units'].append(dict(where=f'probe.a:{line}:1', start=start, end=end,
        function=True, body_start=start, body_end=end))


def total(rows, stock):
    return hidden.calculate(rows, stock, ROOT)['hidden_bytes']


class Arithmetic(unittest.TestCase):
    def test_fake_skipped_span_grows_exactly(self):
        rows, stock = ledger()
        baseline = total(rows, stock)
        add_skip(rows, stock, 10, 27)
        self.assertEqual(total(rows, stock) - baseline, 17, 'fake skipped body adds exactly 17 bytes')

    def test_overlapping_skipped_span_does_not_double_count(self):
        rows, stock = ledger()
        add_skip(rows, stock, 10, 30)
        add_skip(rows, stock, 20, 40, 2)
        self.assertEqual(total(rows, stock), 30, 'overlapping skips counted once')

    def test_nested_attempt_exposes_only_its_examined_part(self):
        rows, stock = ledger()
        add_skip(rows, stock, 10, 90)
        child = dict(where='/source/compiler/probe.a:2:1', status='attempted',
            body_start=30, body_end=70, kind='KindFunctionDeclaration', checker_diagnostics=[])
        rows[1]['units'].append(child)
        stock['probe.a']['units'].append(dict(where='probe.a:2:1', start=25, end=70,
            function=True, body_start=30, body_end=70))
        rows[1]['findings'].append(dict(kind='Boundary', where='/source/compiler/probe.a:3:1',
            unit=child['where'], start=40, end=50, reason='failed statement', text='deliberate refusal'))
        self.assertEqual(total(rows, stock), 45, '80 blocked - (45 child bytes - 10 failed bytes)')
        rows[1]['findings'].append(copy.deepcopy(rows[1]['findings'][0]))
        self.assertEqual(total(rows, stock), 45, 'duplicate boundary cannot grow total')

    def test_dependency_skip_exposes_no_expression_body(self):
        rows, stock = ledger()
        stock['probe.a']['bodies'] = [dict(where='probe.a:1:1', start=10, end=27)]
        finding = dict(kind='SkippedDependency', where=WHERE, unit=WHERE,
            reason='checker rejects expression body', text='deliberate diagnostic')
        rows[1]['findings'].extend([finding, copy.deepcopy(finding)])
        self.assertEqual(total(rows, stock), 17, 'dependency body counted once even without a declaration unit')

    def test_attempted_parent_cannot_expose_skipped_child(self):
        rows, stock = ledger()
        rows[1]['units'].append(dict(where=WHERE, status='attempted', body_start=10, body_end=90))
        stock['probe.a']['units'].append(dict(where='probe.a:1:1', start=0, end=90,
            function=True, body_start=10, body_end=90))
        add_skip(rows, stock, 40, 60, 2)
        self.assertEqual(total(rows, stock), 20, 'attempted parent retains checker-skipped child')

    def test_signature_failure_keeps_whole_body_hidden(self):
        rows, stock = ledger()
        rows[1]['units'].append(dict(where=WHERE, status='attempted', body_start=20, body_end=80))
        stock['probe.a']['units'].append(dict(where='probe.a:1:1', start=10, end=80,
            function=True, body_start=20, body_end=80))
        rows[1]['findings'].append(dict(kind='Boundary', where=WHERE, unit=WHERE,
            start=20, end=80, reason='signature failed', text='unsupported parameter'))
        self.assertEqual(total(rows, stock), 60)

    def test_ranges_against_independent_byte_sets(self):
        generator = random.Random(603)
        for _ in range(500):
            spans = [tuple(sorted(generator.sample(range(101), 2))) for _ in range(12)]
            cuts = [tuple(sorted(generator.sample(range(101), 2))) for _ in range(12)]
            expected = set().union(*(set(range(a, b)) for a, b in spans))
            expected -= set().union(*(set(range(a, b)) for a, b in cuts))
            actual = hidden.subtract(spans, cuts)
            self.assertEqual(hidden.size(actual), len(expected))
            self.assertEqual(set().union(*(set(range(a, b)) for a, b in actual)), expected)

    def test_legacy_and_invalid_ranges_fail(self):
        rows, stock = ledger()
        rows[0]['latent_mode'] = 'first-error'
        with self.assertRaisesRegex(ValueError, 'full latent'):
            total(rows, stock)
        rows[0]['latent_mode'] = 'full'
        add_skip(rows, stock, 10, 101)
        with self.assertRaisesRegex(ValueError, 'invalid skipped body'):
            total(rows, stock)
        with self.assertRaisesRegex(ValueError, 'invalid half-open'):
            hidden.union([(-1, 3)])


if __name__ == '__main__':
    unittest.main(verbosity=2)
