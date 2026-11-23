"""Check that each adapter supplies one ledger file and the ledger is generated."""

from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

STAGE = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(STAGE))
from apply import adaptation_directories

UPSTREAM = STAGE / 'upstream'


class UpstreamLedgerTests(unittest.TestCase):
    def test_layout_and_format(self):
        self.assertFalse((UPSTREAM / 'LEDGER.md').exists(),
                         'stage3/upstream/LEDGER.md must be absent')
        for directory in adaptation_directories():
            with self.subTest(adaptation=directory.name):
                path = directory / 'LEDGER.md'
                self.assertTrue(path.is_file(), f'missing ledger: {path.relative_to(STAGE)}')
                text = path.read_text(encoding='utf-8')
                self.assertTrue(text.strip(), 'ledger is empty')
                if text.startswith('none:'):
                    self.assertRegex(text, r'\Anone: [^\n]+\n?\Z')
                else:
                    # Markdown entries can be plain paragraphs. Reject control
                    # characters and content with no textual words.
                    self.assertTrue(all(c.isprintable() or c in '\n\t' for c in text),
                                    'ledger must contain Markdown text')
                    self.assertTrue(any(c.isalnum() for c in text),
                                    'ledger must contain Markdown text')

    def test_generator_preserves_entries_in_apply_order(self):
        blocks = [(UPSTREAM / 'header.md').read_bytes()]
        for directory in adaptation_directories():
            path = directory / 'LEDGER.md'
            if not path.is_file():
                self.fail(f'missing ledger: {path.relative_to(STAGE)}')
            entry = path.read_bytes()
            if not entry.startswith(b'none:'):
                blocks.append(entry)
        expected = b''.join(block.rstrip(b'\n') + b'\n\n' for block in blocks)[:-1]
        command = [sys.executable, '-B', str(UPSTREAM / 'ledger.py')]
        actual = subprocess.run(command, check=True, capture_output=True)
        self.assertTrue(actual.stdout == expected,
                        'generated ledger differs from entries in apply order')
        with tempfile.TemporaryDirectory() as scratch:
            output = Path(scratch) / 'ledger.md'
            subprocess.run([*command, str(output)], check=True, capture_output=True)
            self.assertTrue(output.read_bytes() == expected,
                            'named ledger differs from entries in apply order')
