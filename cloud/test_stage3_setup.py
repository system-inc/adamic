#!/usr/bin/env python3
"""Hold the API npm stamp to all its inputs and its actual installed contents."""
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

sys.dont_write_bytecode = True
SOURCE = Path(__file__).resolve().parent
MODULE = Path(os.environ.get('ADAMIC_STAGE3_SETUP_MODULE', SOURCE / 'setup-stage3-api.py'))
spec = importlib.util.spec_from_file_location('stage3_setup', MODULE)
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)


class Stage3Key(unittest.TestCase):
    def test_each_component_invalidates(self):
        inputs = dict(lock=b'lock', manifest=b'manifest', bootstrap=b'bootstrap', helper=b'helper', node='v24')
        before = helper.stamp_key(**inputs)
        for name in inputs:
            with self.subTest(component=name):
                changed = dict(inputs, **{name: 'changed' if name == 'node' else b'changed'})
                self.assertNotEqual(before, helper.stamp_key(**changed), name)

    def test_tree_validates_bytes_modes_links(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            target = root / 'index.js'
            target.write_text('before')
            target.chmod(0o644)
            link = root / 'tsc'
            link.symlink_to('index.js')
            previous = helper.tree_digest(root)
            target.write_text('after')
            changed = helper.tree_digest(root)
            self.assertNotEqual(previous, changed)
            target.chmod(0o600)
            permissions = helper.tree_digest(root)
            self.assertNotEqual(changed, permissions)
            other = root / 'other.js'
            other.write_text('other executable')
            linked = helper.tree_digest(root)
            link.unlink()
            link.symlink_to('other.js')
            self.assertNotEqual(linked, helper.tree_digest(root))
            link.unlink()
            link.symlink_to('missing.js')
            with self.assertRaisesRegex(ValueError, 'loses its target'):
                helper.tree_digest(root)
            link.unlink()
            link.symlink_to('/outside/install')
            with self.assertRaisesRegex(ValueError, 'escapes'):
                helper.tree_digest(root)

    def test_absent_lock_needs_no_tools(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.assertIn('package-lock.json absent', helper.prepare(root, root / 'tools', '/missing/node'))
            self.assertFalse((root / 'tools').exists())


@unittest.skipUnless(os.environ.get('ADAMIC_SETUP_INTEGRATION') == '1', 'opt-in registry')
class Stage3Integration(unittest.TestCase):
    def test_lock_mutation_and_uncached_bytes(self):
        scratch = Path(tempfile.mkdtemp(prefix='stage3-setup-proof-', dir='/tmp/adamic-gate'))
        print('integration logs:', scratch, flush=True)
        repository = scratch / 'repository'
        api = repository / 'stage3/api'
        shutil.copytree(SOURCE / 'testdata/stage3-api', api)
        tools = scratch / 'tools'
        node = str(Path(os.environ.get('ADAMIC_TOOLS', '/opt/adamic-tools')) / 'bin/node')
        number = 0

        def run(uncached=False, failure=False):
            nonlocal number
            number += 1
            log = scratch / f'{number:02d}.log'
            environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1' if uncached else '0')
            with log.open('wb') as output:
                result = subprocess.run(['python3', str(MODULE), str(repository), str(tools), node],
                                        env=environment, stdout=output, stderr=subprocess.STDOUT)
            answer = log.read_text()
            if failure:self.assertNotEqual(result.returncode, 0, answer)
            else:self.assertEqual(result.returncode, 0, answer)
            return answer

        self.assertIn('installed (npm ci', run())
        self.assertEqual(json.loads((api / 'node_modules/@types/node/package.json').read_text())['version'], '25.3.3')
        self.assertTrue((api / 'node_modules/.bin/tsc').is_symlink())
        self.assertIn('skipped (validated API lock', run())
        lock = api / 'package-lock.json'
        lock.write_bytes(lock.read_bytes() + b'\n')
        self.assertIn('installed (npm ci', run())
        self.assertIn('skipped (validated API lock', run())
        before = helper.tree_digest(api / 'node_modules')
        self.assertIn('installed (npm ci', run(uncached=True))
        self.assertEqual(before, helper.tree_digest(api / 'node_modules'))
        declarations = api / 'node_modules/@types/node/index.d.ts'
        declarations.write_text('corrupted declarations\n')
        self.assertIn('installed (npm ci', run())
        self.assertEqual(before, helper.tree_digest(api / 'node_modules'))
        # A bad lock integrity must fail rather than stamping the failed npm run.
        content = json.loads(lock.read_text())
        content['packages']['node_modules/@types/node']['integrity'] = 'sha512-' + 'A' * 86 + '=='
        lock.write_text(json.dumps(content))
        self.assertIn('EINTEGRITY', run(failure=True))


if __name__ == '__main__':
    unittest.main()
