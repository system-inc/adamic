"""Check the two reviewed selection files in separate gate units."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import unittest

from census import ROOT, select
from corpus import PIN, ROOT as DRIVER, choose, selection_manifest
from run import first_difference


class Regeneration(unittest.TestCase):
    def setUp(self):
        upstream = os.environ.get('ADAMIC_TYPESCRIPT_SOURCE')
        self.assertTrue(upstream, 'set ADAMIC_TYPESCRIPT_SOURCE to regenerate the pinned selections')
        self.tree = Path(upstream).resolve()
        self.api = self.tree / 'lib/typescript.js'
        head = subprocess.check_output(['git', '-C', str(self.tree), 'rev-parse', 'HEAD'], text=True).strip()
        self.assertEqual(head, PIN, 'upstream pin mismatch')

    def assert_generator_bytes(self, path, manifest):
        actual = (json.dumps(manifest, indent=2) + '\n').encode()
        expected = path.read_bytes()
        if actual != expected:
            difference = first_difference(expected, actual)
            self.fail(f'{path}: selection bytes differ: {difference}')
        print(f'{path.parent.name}: {len(manifest["cases"])} cases, '
              f'SHA-256 {hashlib.sha256(actual).hexdigest()}, byte-identical')

    def test_acceptance_regenerates(self):
        self.assert_generator_bytes(DRIVER / 'selection.json',
                                    selection_manifest(self.tree, *choose(self.tree, self.api)))

    def test_upstream_regenerates(self):
        self.assert_generator_bytes(ROOT / 'selection.json', select(self.tree, self.api))


if __name__ == '__main__':
    result = unittest.main(exit=False).result
    if result.skipped:
        print(f'Regeneration must run without skips: {result.skipped}', file=sys.stderr)
    raise SystemExit(not result.wasSuccessful() or bool(result.skipped))
