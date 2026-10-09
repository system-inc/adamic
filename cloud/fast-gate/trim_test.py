#!/usr/bin/env python3
"""cloud/fast-gate/trim-go-cache.sh trims Go's build cache only over 80 percent of its disk (#93eddxh), with df and go
stood in for, against a real directory of old and fresh cache entries."""
import os
from pathlib import Path
import subprocess
import tempfile
import time
import unittest

script = Path(__file__).with_name('trim-go-cache.sh')


class TrimTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.cache = self.root / 'go-build'
        (self.cache / 'ab').mkdir(parents=True)
        self.old = self.cache / 'ab' / 'old-a'
        self.fresh = self.cache / 'ab' / 'fresh-a'
        self.old.write_text('old')
        self.fresh.write_text('fresh')
        stale = time.time() - 13 * 3600
        os.utime(self.old, (stale, stale))
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        self.stub('go', 'echo "$TEST_CACHE"\n')
        # Used percent before the trim, then after it (the first call and every later one).
        self.stub('df', 'n=$(cat "$TEST_ROOT/df-calls" 2>/dev/null || echo 0); echo $((n + 1)) > "$TEST_ROOT/df-calls"\n'
                        'if [ "$n" = 0 ]; then used=$TEST_BEFORE; else used=$TEST_AFTER; fi\n'
                        'printf "Filesystem 1024-blocks Used Available Capacity Mounted\\n/dev/sdd 100 %s 0 %s%% /\\n" "$used" "$used"\n')

    def stub(self, name, body):
        path = self.bin / name
        path.write_text('#!/bin/bash\n' + body)
        path.chmod(0o755)

    def trim(self, before, after):
        env = dict(os.environ, PATH='%s:%s' % (self.bin, os.environ['PATH']), TEST_ROOT=str(self.root), TEST_CACHE=str(self.cache),
                   TEST_BEFORE=str(before), TEST_AFTER=str(after))
        return subprocess.run(['bash', str(script)], env=env, capture_output=True, text=True, timeout=30)

    def test_over_80_percent_the_entries_unused_12_hours_go(self):
        result = self.trim(85, 60)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse(self.old.exists())
        self.assertTrue(self.fresh.exists())
        self.assertIn('disk 85% to 60% used', result.stdout)
        self.assertNotIn('needs a hand', result.stdout)

    def test_at_50_percent_nothing_is_touched(self):
        result = self.trim(50, 50)
        self.assertEqual((result.returncode, result.stdout), (0, ''))
        self.assertTrue(self.old.exists())

    def test_still_over_90_after_the_trim_says_so(self):
        self.assertIn('disk still 93% used', self.trim(97, 93).stdout)


if __name__ == '__main__':
    unittest.main()
