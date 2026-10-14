"""The supplied upstream install must match its committed byte pin before copying."""
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import products


class ProductPackageTests(unittest.TestCase):
    def test_changed_installed_bytes_are_refused_before_copy(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            source = root / 'source'
            module = source / 'node_modules/probe'
            module.mkdir(parents=True)
            package = dict(version='1.0.0', integrity='sha512-pinned')
            (source / 'package-lock.json').write_text(json.dumps(dict(packages={
                '': dict(name='upstream'), 'node_modules/probe': package})))
            (source / 'node_modules/.package-lock.json').write_text(json.dumps(dict(
                packages={'node_modules/probe': package})))
            (module / 'package.json').write_text(json.dumps(dict(version='1.0.0')))
            payload = module / 'index.js'
            payload.write_text('original installed bytes')
            (root / 'stage3/lane').mkdir(parents=True)
            (root / 'stage3/source.json').write_text('{}')
            (root / 'stage3/lane/node_modules.sha256').write_text(products.apply.packages_digest(source / 'node_modules'))
            with patch.object(products, 'ROOT', root), patch.object(products.apply, 'typescript_source', return_value=source):
                self.assertEqual(products.packages(), (source, source / 'node_modules'))
                payload.write_text('changed bytes, unchanged version and hidden lockfile')
                with self.assertRaisesRegex(ValueError, 'supplied upstream packages hash to .*not pinned'):
                    products.packages()


if __name__ == '__main__':
    unittest.main()
