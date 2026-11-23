"""fsdiff.py sorts every unit into exactly one kind, and families.py holds the callable-slot rewording as one family.

Each case builds two small first-stop logs in firststops.py's format, runs fsdiff.py on them as a process, and reads
its counts line. Mutants: drop the reworded split in fsdiff.py and test_reworded_is_not_a_move goes red; drop the
shift split and test_a_shifted_line_is_not_a_move goes red; drop the callable-slot rule from families.py and
test_callable_slot_reasons_are_one_family goes red.
"""

import re
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import families

# fsdiff.py refuses a table under 100 first stops, so every log carries this many units that never change.
FILLER = [('src/compiler/filler.ts', f'f{index}', 'KindFunctionDeclaration', '1', 'NotYet',
           f'src/compiler/filler.ts:{index + 1}:1', 'a case whose type differs from the switch\'s') for index in range(120)]
CALLABLE = ("a () => true seen as a () => Info, whose result is true, held unboxed, in one and Info, held as a boxed "
            "union, in the other (a callable's boxed and unboxed representations differ, and the adapter between them "
            "isn't built)")
WEAK = 'a () => true seen as a () => Info (one keeps something weakly that the other keeps strongly)'


def log(rows):
    return ''.join('\t'.join(('FS', 'entry') + row) + '\n' for row in FILLER + rows)


def unit(name, where, reason, kind='Refused'):
    return ('src/compiler/x.ts', name, 'KindFunctionDeclaration', '1', kind, where, reason)


class FirstStopsDiffTests(unittest.TestCase):
    def counts(self, before, after):
        with tempfile.TemporaryDirectory() as scratch:
            previous, current = Path(scratch) / 'before.log', Path(scratch) / 'after.log'
            previous.write_text(log(before))
            current.write_text(log(after))
            result = subprocess.run([sys.executable, '-I', '-B', str(HERE / 'fsdiff.py'), str(previous), str(current), 'entry'],
                                    check=True, capture_output=True, text=True)
        line = result.stdout.splitlines()[1]
        return {kind: int(count) for count, kind in re.findall(r'(\d+) ([a-z][a-z ]+?)(?:,|$)', line)}

    def test_lowers_now_and_stops_newly(self):
        counts = self.counts([unit('a', 'src/compiler/x.ts:1:1', 'a cast the runtime can\'t check')],
                             [unit('b', 'src/compiler/x.ts:2:1', 'a cast the runtime can\'t check')])
        self.assertEqual(counts['lower now'], 1)
        self.assertEqual(counts['stop newly'], 1)

    def test_moved_family_and_site(self):
        counts = self.counts([unit('a', 'src/compiler/x.ts:1:1', 'a cast the runtime can\'t check'),
                              unit('b', 'src/compiler/x.ts:5:1', 'a cast the runtime can\'t check')],
                             [unit('a', 'src/compiler/y.ts:9:1', 'sort without a comparator'),
                              unit('b', 'src/compiler/y.ts:7:1', 'a cast the runtime can\'t check')])
        self.assertEqual(counts['moved family'], 1)
        self.assertEqual(counts['moved site in one family'], 1)

    def test_reworded_is_not_a_move(self):
        counts = self.counts([unit('a', 'src/compiler/x.ts:3:4', WEAK)], [unit('a', 'src/compiler/x.ts:3:4', CALLABLE)])
        self.assertEqual(counts['reworded at the same site'], 1)
        self.assertEqual(counts['moved family'], 0)

    def test_a_shifted_line_is_not_a_move(self):
        counts = self.counts([unit('a', 'src/compiler/x.ts:3:4', 'new a ParenthesizedExpression', 'NotYet')],
                             [unit('a', 'src/compiler/x.ts:30:4', 'new a ParenthesizedExpression', 'NotYet')])
        self.assertEqual(counts['same stop at a shifted line'], 1)
        self.assertEqual(counts['moved site in one family'], 0)

    def test_callable_slot_reasons_are_one_family(self):
        other = ("a (n: Node) => boolean seen as a (element: Node) => \"quit\" | boolean, whose result is boolean, held "
                 "unboxed, in one and \"quit\" | boolean, held as a boxed union, in the other (a callable's boxed and "
                 "unboxed representations differ, and the adapter between them isn't built)")
        self.assertEqual(families.familyOf(CALLABLE), families.familyOf(other))
        self.assertNotEqual(families.familyOf(CALLABLE), CALLABLE)
        self.assertEqual(families.familyOf(WEAK), 'a <A> seen as a <B> (one keeps something weakly that the other keeps strongly)')

    def test_a_short_table_is_refused(self):
        with tempfile.TemporaryDirectory() as scratch:
            short = Path(scratch) / 'short.log'
            short.write_text('FS\tentry\tsrc/compiler/x.ts\ta\tKindFunctionDeclaration\t1\tRefused\tsrc/compiler/x.ts:1:1\tr\n')
            result = subprocess.run([sys.executable, '-I', '-B', str(HERE / 'fsdiff.py'), str(short), str(short)],
                                    capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('under 100', result.stderr)


if __name__ == '__main__':
    unittest.main()
