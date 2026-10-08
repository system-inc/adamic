#!/usr/bin/env python3
"""Exercise ordinary apply on a clean checkout with a deliberately stale table."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parent.parent


class ApplyCheckoutTests(unittest.TestCase):
    def test_ordinary_apply_leaves_git_clean(self):
        # Not parallel: apply installs into its shared stock-API cache with npm ci.
        with tempfile.TemporaryDirectory(prefix='stage3-apply-checkout-') as scratch:
            source = Path(scratch) / 'source'
            output = Path(scratch) / 'adapted'
            subprocess.run(['git', '-C', str(ROOT), 'worktree', 'add', '--detach', str(source), 'HEAD'],
                           check=True)
            try:
                table = source / 'stage3/patch-set.md'
                table.write_text(table.read_text() + '\n<!-- Deliberately stale test table. -->\n')
                if os.environ.get('STAGE3_APPLY_MUTANT') == '1':
                    script = source / 'stage3/apply.py'
                    text = script.read_text()
                    self.assertEqual(text.count('if args.write_table:'), 1)
                    script.write_text(text.replace('if args.write_table:', 'if True:'))
                subprocess.run(['git', '-C', str(source), 'add', 'stage3/patch-set.md', 'stage3/apply.py'], check=True)
                subprocess.run(['git', '-C', str(source), '-c', 'user.name=Apply test',
                                '-c', 'user.email=apply-test@example.invalid', 'commit', '-m',
                                'Seed the apply checkout test'], check=True)
                before = table.read_bytes()
                self.assertEqual(subprocess.check_output(['git', '-C', str(source), 'status', '--porcelain']), b'')
                subprocess.run(['bash', str(source / 'stage3/apply.sh'), str(output)], check=True)
                status = subprocess.check_output(['git', '-C', str(source), 'status', '--porcelain'], text=True)
                self.assertEqual(status, '', 'ordinary apply dirtied the source checkout')
                self.assertEqual(table.read_bytes(), before)
                generated = (output / 'patch-set.md').read_text()
                self.assertIn('| 75-optional-widening |', generated)
                self.assertNotIn('Deliberately stale test table', generated)
            finally:
                subprocess.run(['git', '-C', str(ROOT), 'worktree', 'remove', '--force', str(source)], check=True)


if __name__ == '__main__':
    unittest.main(verbosity=2)
