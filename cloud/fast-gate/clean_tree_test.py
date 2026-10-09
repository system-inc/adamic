#!/usr/bin/env python3
"""A gate tree is exactly its commit before a gate runs: clean-tree.sh, against a planted stale file in a tree and its
submodule."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

script = Path(__file__).with_name('clean-tree.sh')


def git(directory, *arguments):
    return subprocess.run(['git', '-C', str(directory), '-c', 'user.name=t', '-c', 'user.email=t@t', '-c', 'protocol.file.allow=always'] + list(arguments),
                          check=True, capture_output=True, text=True).stdout


class CleanTreeTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        module = root / 'module'
        module.mkdir()
        git(module, 'init', '-q')
        (module / '.gitignore').write_text('/.generated/\n')
        (module / 'source.ts').write_text('export const a = 1;\n')
        git(module, 'add', '.')
        git(module, 'commit', '-qm', 'module')
        self.tree = root / 'tree'
        self.tree.mkdir()
        git(self.tree, 'init', '-q')
        (self.tree / '.gitignore').write_text('node_modules/\n')
        (self.tree / 'main.go').write_text('package main\n')
        git(self.tree, 'add', '.')
        git(self.tree, 'submodule', 'add', '-q', str(module), 'cohere')
        git(self.tree, 'commit', '-qm', 'tree')

    def test_stale_files_in_the_tree_and_its_submodule_are_gone(self):
        # As Oct 9's whole-gate tree held them: a generated registry and a node_modules from earlier candidates.
        (self.tree / 'cohere' / '.generated').mkdir()
        (self.tree / 'cohere' / '.generated' / 'registry.ts').write_text('stale\n')
        (self.tree / 'node_modules').mkdir()
        (self.tree / 'node_modules' / 'x.js').write_text('stale\n')
        (self.tree / 'leftover.txt').write_text('stale\n')
        result = subprocess.run(['bash', str(script), str(self.tree)], capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        for stale in ('cohere/.generated/registry.ts', 'node_modules/x.js', 'leftover.txt'):
            self.assertFalse((self.tree / stale).exists(), stale)
        self.assertIn('removed from the gate tree: cohere/.generated/', result.stdout)
        # What the commit holds stays.
        self.assertEqual((self.tree / 'cohere' / 'source.ts').read_text(), 'export const a = 1;\n')
        self.assertEqual(git(self.tree, 'status', '--porcelain', '--ignored'), '')


if __name__ == '__main__':
    unittest.main()
