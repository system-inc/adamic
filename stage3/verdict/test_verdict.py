#!/usr/bin/env python3
"""The three independent byte checks must each notice its own mutation."""
from pathlib import Path
import tempfile
import unittest
from run import compare, first_difference


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

    def test_eof_difference(self):
        self.assertEqual(first_difference(b'abc', b'ab')['byte_offset'], 2)
        self.assertEqual(first_difference(b'', b'x')['byte_offset'], 0)


if __name__ == '__main__':
    unittest.main()
