#!/usr/bin/env python3
"""Exercise ordinary apply on a clean checkout with a deliberately stale table."""
import io
import os
import importlib.util
import json
import shutil
import sys
from unittest import mock
from pathlib import Path
import subprocess
import tempfile
import unittest

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
SPEC = importlib.util.spec_from_file_location('stage3_apply', ROOT / 'stage3/apply.py')
APPLY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(APPLY)


def shard(case, count):
    return CASES.index(case) % count




class ApplyCheckoutTests(unittest.TestCase):
    def test_ordinary_apply_leaves_git_clean(self):
        # Not parallel: this proof writes its own detached git worktree.
        with tempfile.TemporaryDirectory(prefix='stage3-apply-checkout-') as scratch:
            source = Path(scratch) / 'source'
            output = Path(scratch) / 'adapted'
            subprocess.run(['git', '-C', str(ROOT), 'worktree', 'add', '--detach', str(source), 'HEAD'],
                           check=True)
            try:
                table = source / 'stage3/patch-set.md'
                table.write_text(table.read_text() + '\n<!-- Deliberately stale test table. -->\n')
                # Exercise working changes too, before the delivery commit.
                for name in ['apply.py', 'apply.sh']:
                    shutil.copyfile(ROOT / 'stage3' / name, source / 'stage3' / name)
                helpers = subprocess.check_output(['git', '-C', str(ROOT), 'diff', '--name-only',
                                                   'HEAD', '--', 'stage3/adapt'], text=True).splitlines()
                helpers += subprocess.check_output(['git', '-C', str(ROOT), 'ls-files', '--others',
                                                    '--exclude-standard', 'stage3/adapt'], text=True).splitlines()
                for name in helpers:
                    target = source / name
                    target.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(ROOT / name, target)
                subprocess.run(['git', '-C', str(source), 'add', 'stage3/patch-set.md',
                                'stage3/apply.py', 'stage3/apply.sh', 'stage3/adapt'], check=True)
                subprocess.run(['git', '-C', str(source), '-c', 'user.name=Apply test',
                                '-c', 'user.email=apply-test@example.invalid', 'commit', '-m',
                                'Seed the apply checkout test'], check=True)
                before = table.read_bytes()
                self.assertEqual(subprocess.check_output(['git', '-C', str(source), 'status', '--porcelain']), b'')
                command = ['bash', str(source / 'stage3/apply.sh'), str(output)]
                if os.environ.get('STAGE3_APPLY_MUTANT') == '1':
                    # Force the real write-table path without altering build inputs:
                    # changing apply.py's bytes must invalidate the product key.
                    command.append('--write-table')
                subprocess.run(command, check=True)
                status = subprocess.check_output(['git', '-C', str(source), 'status', '--porcelain'], text=True)
                self.assertEqual(status, '', 'ordinary apply dirtied the source checkout')
                self.assertEqual(table.read_bytes(), before)
                generated = (output / 'patch-set.md').read_text()
                self.assertIn('| 75-optional-widening |', generated)
                self.assertNotIn('Deliberately stale test table', generated)
            finally:
                subprocess.run(['git', '-C', str(ROOT), 'worktree', 'remove', '--force', str(source)], check=True)


class ProductTests(unittest.TestCase):
    def test_input_key_covers_all_inputs(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            inputs = ['source.json', 'apply.py', 'api/package.json',
                      'api/package-lock.json', 'adapt/00-setup/adapt.cjs',
                      'adapt/99-probe/proof.a', 'adapt/99-probe/helper.json']
            for name in inputs:
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text(name)
            with mock.patch.object(APPLY, 'stage', root):
                baseline = APPLY.product_key()
                for name in inputs:
                    path = root / name
                    before = path.read_bytes()
                    path.write_bytes(before + b'mutant')
                    self.assertNotEqual(APPLY.product_key(), baseline, name)
                    path.write_bytes(before)
                with mock.patch.object(APPLY.subprocess, 'check_output', return_value=b'v0.0.0'):
                    self.assertNotEqual(APPLY.product_key(), baseline, 'Node version')
                extra = root / 'adapt/99-probe/new.a'
                extra.write_text('new input')
                self.assertNotEqual(APPLY.product_key(), baseline)
                extra.unlink()
                empty = root / 'adapt/98-empty'
                empty.mkdir()
                self.assertNotEqual(APPLY.product_key(), baseline)
                empty.rmdir()
                file_link = root / 'adapt/99-probe/evidence-link'
                file_link.symlink_to('helper.json')
                linked_key = APPLY.product_key()
                self.assertNotEqual(linked_key, baseline)
                target = root / 'adapt/99-probe/helper.json'
                original = target.read_bytes()
                target.write_bytes(original + b'changed linked bytes')
                self.assertNotEqual(APPLY.product_key(), linked_key)
                target.write_bytes(original)
                file_link.unlink()
                link = root / 'adapt/99-probe/external'
                link.symlink_to(root / 'api', target_is_directory=True)
                with self.assertRaisesRegex(RuntimeError, 'unkeyed symlink'):
                    APPLY.product_key()
                link.unlink()
                (root / 'patch-set.md').write_text('stale generated output')
                self.assertEqual(APPLY.product_key(), baseline)

    def test_unkeyed_parser_override_is_refused(self):
        with tempfile.TemporaryDirectory() as scratch:
            with mock.patch.dict(os.environ, CENSUS_TYPESCRIPT='unkeyed-parser'), mock.patch.object(
                    APPLY, 'run', side_effect=AssertionError('unkeyed build ran')):
                with self.assertRaisesRegex(RuntimeError, 'without overrides'):
                    APPLY.build_product(Path(scratch) / 'output', Path(scratch) / 'cache')
            self.assertFalse((Path(scratch) / 'output').exists())

    def product(self, scratch):
        source = scratch / 'source'
        source.mkdir()
        (source / 'patch-set.md').write_text('generated table')
        (source / 'input.a').write_text('const value = 1;')
        store = scratch / 'products'
        APPLY.save_product(str(store), 'test-key', source)
        return store

    def test_restore_fetches_verified_product(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            store = self.product(scratch)
            output = scratch / 'output'
            # file:// exercises the URL reader in a fresh destination.
            APPLY.restore_product(store.as_uri(), 'test-key', output)
            self.assertEqual((output / 'input.a').read_text(), 'const value = 1;')
            self.assertEqual((output / 'patch-set.md').read_text(), 'generated table')

    def test_payload_mutant_is_rejected(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            store = self.product(scratch)
            archive = store / 'test-key.p000'
            data = bytearray(archive.read_bytes())
            data[len(data) // 2] ^= 1
            archive.write_bytes(data)
            output = scratch / 'output'
            with self.assertRaisesRegex(RuntimeError, 'part hash mismatch'):
                APPLY.restore_product(str(store), 'test-key', output)
            self.assertFalse(output.exists())

    def test_payload_manifest_mutant_is_rejected(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            store = self.product(scratch)
            path = store / 'test-key.manifest'
            manifest = json.loads(path.read_text())
            manifest['sha256'] = '0' * 64
            path.write_text(json.dumps(manifest))
            with self.assertRaisesRegex(RuntimeError, 'payload hash mismatch'):
                APPLY.restore_product(str(store), 'test-key', scratch / 'output')
            self.assertFalse((scratch / 'output').exists())

    def test_manifest_mutant_is_rejected(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            store = self.product(scratch)
            path = store / 'test-key.manifest'
            manifest = json.loads(path.read_text())
            manifest['key'] = 'other-inputs'
            path.write_text(json.dumps(manifest))
            with self.assertRaisesRegex(RuntimeError, 'input key mismatch'):
                APPLY.restore_product(str(store), 'test-key', scratch / 'output')
            self.assertFalse((scratch / 'output').exists())

    def test_missing_product_builds_and_reports(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            output = scratch / 'output'
            def build(out, cache):
                out.mkdir()
                (out / 'patch-set.md').write_text('fallback table')
            message = io.StringIO()
            with (mock.patch.object(APPLY, 'build_product', side_effect=build) as builder, mock.patch.object(
                    APPLY, 'product_key', return_value='missing'), mock.patch.dict(
                    os.environ, STAGE3_PRODUCT_STORE=str(scratch / 'products'), STAGE3_CACHE=str(scratch / 'cache')),
                    mock.patch.object(sys, 'argv', ['apply.py', str(output)]), mock.patch.object(sys, 'stderr', message)):
                APPLY.main()
            builder.assert_called_once_with(output, scratch / 'cache')
            self.assertIn('cache miss: missing; building adapted tree locally', message.getvalue())
            self.assertEqual((output / 'patch-set.md').read_text(), 'fallback table')
            APPLY.restore_product(str(scratch / 'products'), 'missing', scratch / 'fetched')
            self.assertEqual((scratch / 'fetched/patch-set.md').read_text(), 'fallback table')

    def test_only_manifest_absence_is_a_cache_miss(self):
        with tempfile.TemporaryDirectory() as scratch:
            output = Path(scratch) / 'output'
            for error in [FileNotFoundError('missing'),
                          APPLY.urllib.error.HTTPError('https://cache/key.manifest', 404, 'missing', {}, None)]:
                with mock.patch.object(APPLY, 'fetch', side_effect=error):
                    with self.assertRaises(RuntimeError) as caught:
                        APPLY.restore_product('https://cache', 'key', output)
                    self.assertIsInstance(caught.exception, APPLY.ProductMissing)
            for error in [PermissionError('denied'),
                          APPLY.urllib.error.HTTPError('https://cache/key.manifest', 403, 'denied', {}, None)]:
                with mock.patch.object(APPLY, 'fetch', side_effect=error):
                    with self.assertRaises(RuntimeError) as caught:
                        APPLY.restore_product('https://cache', 'key', output)
                    self.assertNotIsInstance(caught.exception, APPLY.ProductMissing)
            self.assertFalse(output.exists())

    def test_corrupt_product_never_builds(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            store = self.product(scratch)
            (store / 'test-key.p000').write_bytes(b'corrupt')
            with (mock.patch.object(APPLY, 'build_product', side_effect=AssertionError('rebuilt corruption')),
                    mock.patch.object(APPLY, 'product_key', return_value='test-key'), mock.patch.dict(
                    os.environ, STAGE3_PRODUCT_STORE=str(store)), mock.patch.object(
                    sys, 'argv', ['apply.py', str(scratch / 'output')])):
                with self.assertRaisesRegex(RuntimeError, 'part hash mismatch'):
                    APPLY.main()

    def test_publish_hook_describes_local_payloads(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            store = self.product(scratch)
            hook = APPLY.publish_hook(str(store), 'test-key')
            self.assertEqual(hook['key'], 'test-key')
            self.assertEqual(hook['pinned_source'], json.loads((ROOT / 'stage3/source.json').read_text()))
            self.assertIn('stage3/apply.py', hook['inputs'])
            self.assertEqual([Path(p['path']).name for p in hook['payloads']],
                             ['test-key.p000', 'test-key.manifest'])
            self.assertTrue(all(p['asset'] == 'adamic/build-cache/' + Path(p['path']).name
                                for p in hook['payloads']))
            self.assertEqual(hook['publish_order'], 'parts first, manifest last')
            path = store / 'test-key.manifest'
            original = json.loads(path.read_text())
            for field, value, message in [('key', 'other', 'input key mismatch'),
                                           ('parts', [], 'part manifest mismatch'),
                                           ('parts', [dict(original['parts'][0], suffix='../../outside')], 'part manifest mismatch')]:
                path.write_text(json.dumps(dict(original, **{field: value})))
                with self.assertRaises(Exception) as caught:
                    APPLY.publish_hook(str(store), 'test-key')
                self.assertIsInstance(caught.exception, RuntimeError)
                self.assertRegex(str(caught.exception), message)
            path.write_text(json.dumps(original))
            (store / 'test-key.p000').write_bytes(b'corrupt')
            with self.assertRaisesRegex(RuntimeError, 'part hash mismatch'):
                APPLY.publish_hook(str(store), 'test-key')

    def test_write_table_is_explicit(self):
        with tempfile.TemporaryDirectory() as scratch:
            scratch = Path(scratch)
            store = self.product(scratch)
            stage = scratch / 'stage'
            stage.mkdir()
            table = stage / 'patch-set.md'
            table.write_text('stale table')
            with mock.patch.object(APPLY, 'stage', stage), mock.patch.object(
                    APPLY, 'product_key', return_value='test-key'), mock.patch.dict(
                    os.environ, STAGE3_PRODUCT_STORE=str(store)), mock.patch.object(
                    sys, 'argv', ['apply.py', str(scratch / 'output'), '--write-table']):
                APPLY.main()
            self.assertEqual(table.read_text(), 'generated table')

    def test_existing_output_is_refused(self):
        with tempfile.TemporaryDirectory() as scratch:
            with mock.patch.object(APPLY, 'product_key', return_value='test-key'), mock.patch.object(
                    sys, 'argv', ['apply.py', scratch]), mock.patch.object(
                    APPLY, 'restore_product', side_effect=AssertionError('attempted overwrite')):
                with self.assertRaisesRegex(SystemExit, 'refusing to replace'):
                    APPLY.main()

    def test_shards_partition_every_case(self):
        expected = sorted(CASES)
        for count in [1, 2, 3, len(CASES), len(CASES) + 1]:
            groups = [[case for case in CASES if shard(case, count) == index]
                      for index in range(count)]
            self.assertEqual(sorted(case for group in groups for case in group), expected)
            for case in CASES:
                self.assertEqual(sum(case in group for group in groups), 1)
                print(f'{case} shard={shard(case, count)}/{count}')


class ProofShardTests(unittest.TestCase):
    def test_every_indexed_read_file_has_one_shard(self):
        script = """
const assert = require('node:assert/strict');
const {select} = require('./stage3/adapt/00-setup/test-shards.cjs');
for (const name of ['30-indexed-reads', '31-indexed-reads-checker',
                    '32-indexed-reads-program', '33-indexed-reads-emit']) {
    const {files} = require('./stage3/adapt/' + name + '/adapt.cjs');
    for (const count of [1, 2, 3, files.length, files.length + 1]) {
        const groups = [];
        for (let index = 0; index < count; index++) {
            process.env.ADAMIC_TEST_SHARD = `${index}/${count}`;
            groups.push(select(files));
        }
        assert.deepEqual(groups.flat().sort(), [...files].sort());
        for (const file of files) {
            assert.equal(groups.filter(group => group.includes(file)).length, 1);
            console.log(`${name}/${file} shard=${groups.findIndex(group => group.includes(file))}/${count}`);
        }
    }
}
"""
        env = dict(os.environ)
        env.pop('ADAMIC_TEST_LIST_SHARDS', None)
        subprocess.run(['node', '-e', script], cwd=ROOT, env=env, check=True)


CASES = sorted(test.id().removeprefix(__name__ + '.') for cls in [ApplyCheckoutTests, ProductTests, ProofShardTests]
               for test in unittest.defaultTestLoader.loadTestsFromTestCase(cls))


if __name__ == '__main__':
    if len(sys.argv) == 3 and sys.argv[1] == '--list-shards':
        count = int(sys.argv[2])
        if count <= 0:
            raise SystemExit('shard count must be positive')
        print(json.dumps({case: f'{shard(case, count)}/{count}' for case in CASES}, indent=2))
    elif 'ADAMIC_TEST_SHARD' in os.environ:
        try:
            index, count = map(int, os.environ['ADAMIC_TEST_SHARD'].split('/'))
            if not 0 <= index < count:
                raise ValueError()
        except ValueError:
            raise SystemExit('ADAMIC_TEST_SHARD must be i/n with 0 <= i < n')
        suite = unittest.TestSuite(unittest.defaultTestLoader.loadTestsFromName(case, sys.modules[__name__])
                                   for case in CASES if shard(case, count) == index)
        result = unittest.TextTestRunner(verbosity=2).run(suite)
        raise SystemExit(not result.wasSuccessful())
    else:
        unittest.main(verbosity=2)
