#!/usr/bin/env python3
"""stage3's caches follow location.BuildCache's rule, never the home directory when the build cache is named."""
import os
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent))
import caches


class CacheLocationTests(unittest.TestCase):
    def test_the_build_cache_holds_stage3s_cache(self):
        with tempfile.TemporaryDirectory() as build, patch.dict(os.environ, {'ADAMIC_BUILD_CACHE_DIR': build}, clear=True):
            self.assertEqual(caches.stage3_cache(), Path(build).resolve() / 'adamic-stage3')
            self.assertEqual(caches.cache_root(), Path(build).resolve())

    def test_stage3_cache_overrides_everything(self):
        environment = {'STAGE3_CACHE': '/somewhere/stage3', 'ADAMIC_BUILD_CACHE_DIR': '/elsewhere', 'ADAMIC_SANDBOX': '1'}
        with patch.dict(os.environ, environment, clear=True):
            self.assertEqual(caches.stage3_cache(), Path('/somewhere/stage3'))

    def test_a_sandbox_refuses_an_unnamed_cache(self):
        with patch.dict(os.environ, {'ADAMIC_SANDBOX': '1', 'HOME': '/home/runner'}, clear=True):
            with self.assertRaises(RuntimeError):
                caches.stage3_cache()

    def test_otherwise_the_user_cache_directory(self):
        with patch.dict(os.environ, {'HOME': '/home/runner', 'XDG_CACHE_HOME': '/xdg'}, clear=True), \
                patch.object(caches.sys, 'platform', 'linux'):
            self.assertEqual(caches.stage3_cache(), Path('/xdg/adamic-stage3'))
        with patch.dict(os.environ, {'HOME': '/home/runner', 'XDG_CACHE_HOME': 'relative'}, clear=True), \
                patch.object(caches.sys, 'platform', 'linux'):
            self.assertEqual(caches.stage3_cache(), Path('/home/runner/.cache/adamic-stage3'))
        with patch.dict(os.environ, {'HOME': '/Users/runner'}, clear=True), patch.object(caches.sys, 'platform', 'darwin'):
            self.assertEqual(caches.stage3_cache(), Path('/Users/runner/Library/Caches/adamic-stage3'))


if __name__ == '__main__':
    unittest.main(verbosity=2)
