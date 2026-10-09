#!/usr/bin/env python3
"""The landing candidate loop against a real origin repository, with a stand-in ahra for the task tree."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]


def git(directory, *args):
    return subprocess.run(['git', '-C', str(directory), '-c', 'user.name=t', '-c', 'user.email=t@t', *args],
                          check=True, capture_output=True, text=True).stdout.strip()


class LandingCandidateTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = self.root = Path(self.tmp.name)
        git(root, 'init', '-q', '--bare', '-b', 'main', 'origin.git')
        work = self.work = root / 'work'
        git(root, 'clone', '-q', str(root / 'origin.git'), 'work')
        (work / 'a.txt').write_text('one\n')
        git(work, 'add', '.'); git(work, 'commit', '-q', '-m', 'base'); git(work, 'push', '-q', 'origin', 'HEAD:main')
        # The tools checkout the loop runs from, its own clone of origin.
        tools = self.tools = root / 'tools'
        git(root, 'clone', '-q', str(root / 'origin.git'), 'tools')
        (tools / 'cloud').mkdir()
        for name in ('landing-candidate.sh', 'first-step-branches.sh'):
            shutil.copy(ROOT / 'cloud' / name, tools / 'cloud' / name)
        bin = root / 'bin'
        bin.mkdir()
        (bin / 'ahra').write_text('''#!/bin/bash
if [ "$1" = tasks ] && [ "$2" = ready ]; then echo " 48  #nsnehdz   Running  04 Wall"; exit 0; fi
if [ "$1" = tasks ] && [ "$2" = show ]; then cat "$TEST_ROOT/body"; exit 0; fi
printf '%s|%s\\n' "$3" "$4" >> "$TEST_ROOT/sends"
''')
        (bin / 'ahra').chmod(0o755)
        (root / 'body').write_text('Branches: cloud/land-area-next*\nSource: compiler/area-next-fixtures\n')
        self.env = dict(os.environ, PATH=str(bin) + ':' + os.environ['PATH'], TEST_ROOT=str(root),
                        ADAMIC_FAST_GATE_AHRA_DIR=str(root), ADAMIC_LANDING_CANDIDATE_STATE=str(root / 'state'))

    def source(self, text, base='main'):
        git(self.work, 'fetch', '-q', 'origin')
        git(self.work, 'checkout', '-q', '-B', 'fix', 'origin/' + base)
        (self.work / 'b.txt').write_text(text)
        git(self.work, 'add', '.'); git(self.work, 'commit', '-q', '-m', 'fix ' + text.strip())
        git(self.work, 'push', '-q', '-f', 'origin', 'HEAD:compiler/area-next-fixtures')
        return git(self.work, 'rev-parse', 'HEAD')

    def run_once(self):
        return subprocess.run(['/bin/bash', str(self.tools / 'cloud/landing-candidate.sh'), '--once'],
                              env=self.env, capture_output=True, text=True)

    def candidates(self):
        out = git(self.root / 'origin.git', 'for-each-ref', '--format=%(refname:short) %(objectname)', 'refs/heads/cloud/')
        return dict(line.split() for line in out.splitlines())

    def test_a_new_source_tip_becomes_a_reserved_candidate_once(self):
        main = git(self.root / 'origin.git', 'rev-parse', 'main')
        tip = self.source('two\n')
        self.run_once()
        name = 'cloud/land-area-next-auto-' + tip[:8]
        self.assertEqual(list(self.candidates()), [name])
        commit = self.candidates()[name]
        self.assertEqual(git(self.root / 'origin.git', 'log', '-1', '--format=%P', commit).split(), [main, tip])
        self.assertIn('Merge compiler/area-next-fixtures at %s onto main %s' % (tip[:8], main[:8]),
                      git(self.root / 'origin.git', 'log', '-1', '--format=%s', commit))
        self.run_once()
        self.assertEqual(len(self.candidates()), 1)
        tip2 = self.source('three\n')
        self.run_once()
        self.assertIn('cloud/land-area-next-auto-' + tip2[:8], self.candidates())

    def test_a_conflict_builds_nothing_and_tells_integration_once(self):
        self.source('mine\n')
        git(self.work, 'checkout', '-q', '-B', 'other', 'origin/main')
        (self.work / 'b.txt').write_text('theirs\n')
        git(self.work, 'add', '.'); git(self.work, 'commit', '-q', '-m', 'main moves'); git(self.work, 'push', '-q', 'origin', 'HEAD:main')
        self.run_once(); self.run_once()
        self.assertEqual(self.candidates(), {})
        sends = (self.root / 'sends').read_text().splitlines()
        self.assertEqual(len(sends), 1)
        self.assertTrue(sends[0].startswith('system_adamic_integration|Landing candidate for #nsnehdz not built'))

    def test_a_tip_integration_already_merged_by_hand_builds_nothing(self):
        main = git(self.root / 'origin.git', 'rev-parse', 'main')
        tip = self.source('two\n')
        git(self.work, 'checkout', '-q', '-B', 'hand', 'origin/main')
        git(self.work, 'merge', '-q', '--no-ff', '-m', 'by hand', tip)
        git(self.work, 'push', '-q', 'origin', 'HEAD:cloud/land-area-next-12')
        self.run_once()
        self.assertEqual(list(self.candidates()), ['cloud/land-area-next-12'])

    def test_no_source_line_or_a_landed_source_builds_nothing(self):
        (self.root / 'body').write_text('Branches: cloud/land-area-next*\n')
        self.source('two\n')
        self.run_once()
        self.assertEqual(self.candidates(), {})
        (self.root / 'body').write_text('Branches: cloud/land-area-next*\nSource: compiler/area-next-fixtures\n')
        git(self.work, 'push', '-q', 'origin', 'fix:main')
        self.run_once()
        self.assertEqual(self.candidates(), {})


if __name__ == '__main__':
    unittest.main()
