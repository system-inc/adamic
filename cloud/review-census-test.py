#!/usr/bin/env python3
"""The review census against a real git origin: a program in the lane under another name counts, one outside pages
integration once per set, and a loose file beside the lane doesn't count. Run: python3 cloud/review-census-test.py"""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

script = Path(__file__).with_name('review-census.py')


def git(directory, *arguments):
    return subprocess.run(['git', '-C', str(directory), *arguments], check=True, capture_output=True, text=True).stdout.strip()


class ReviewCensusTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        origin = root / 'origin.git'
        git(root, 'init', '-q', '--bare', str(origin))
        self.work = root / 'work'
        git(root, 'clone', '-q', str(origin), str(self.work))
        (self.work / 'cloud').mkdir()
        shutil.copy(script, self.work / 'cloud' / 'review-census.py')
        (self.work / 'cloud' / 'review-tasks.txt').write_text('# review tasks\nfxspptb readings\n')
        lane = self.work / 'internal/oracle/testdata/review'
        (lane / 'agree').mkdir(parents=True)
        (lane / 'agree' / 'fxspptb_renamed.a').write_text('console.log("in the lane");\n')
        (lane / 'loose.a').write_text('console.log("loose");\n')
        git(self.work, 'add', '-A')
        git(self.work, '-c', 'user.name=t', '-c', 'user.email=t@t', 'commit', '-qm', 'lane')
        git(self.work, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        self.attachments = root / 'attachments'
        (self.attachments / 'fxspptb').mkdir(parents=True)
        (self.attachments / 'fxspptb' / 'oct8_p01.a').write_text('console.log("in the lane");\n')
        (self.attachments / 'fxspptb' / 'oct8_p02.a').write_text('console.log("loose");\n')
        (self.attachments / 'fxspptb' / 'notes.md').write_text('not a program\n')
        bin = root / 'bin'
        bin.mkdir()
        (bin / 'ahra').write_text('#!/bin/bash\nprintf "%%s|%%s\\n" "$3" "$4" >> %s\n' % (root / 'sends'))
        (bin / 'ahra').chmod(0o755)
        self.root = root
        self.environment = dict(os.environ, PATH=str(bin) + ':' + os.environ['PATH'], ADAMIC_REVIEW_ATTACHMENTS=str(self.attachments),
                                ADAMIC_REVIEW_STATE=str(root / 'state'), ADAMIC_REVIEW_AHRA_DIR=str(root))

    def once(self):
        result = subprocess.run(['python3', str(self.work / 'cloud' / 'review-census.py'), '--once'], env=self.environment,
                                capture_output=True, text=True, check=True)
        sends = (self.root / 'sends').read_text().splitlines() if (self.root / 'sends').exists() else []
        return result.stdout, sends

    def test_a_program_outside_the_lane_pages_integration_once(self):
        output, sends = self.once()
        self.assertIn('1 review programs outside the lane', output)
        self.assertEqual(len(sends), 1)
        self.assertTrue(sends[0].startswith('system_adamic_integration|'))
        # The renamed program counts by content; the loose file beside the lane doesn't.
        self.assertIn('fxspptb/oct8_p02.a', sends[0])
        self.assertNotIn('oct8_p01.a', sends[0])
        self.assertEqual(len(self.once()[1]), 1, 'the same set pages once')
        # Committing it to the lane clears the census, and a new program pages again.
        (self.work / 'internal/oracle/testdata/review/refused').mkdir()
        shutil.move(str(self.work / 'internal/oracle/testdata/review/loose.a'), str(self.work / 'internal/oracle/testdata/review/refused/fxspptb_p02.a'))
        git(self.work, 'add', '-A')
        git(self.work, '-c', 'user.name=t', '-c', 'user.email=t@t', 'commit', '-qm', 'lane')
        git(self.work, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        output, sends = self.once()
        self.assertIn('0 review programs outside the lane', output)
        self.assertEqual(len(sends), 1, 'a shrinking set is progress, not a page')
        (self.attachments / 'fxspptb' / 'oct8_p03.a').write_text('console.log("new");\n')
        self.assertEqual(len(self.once()[1]), 2)


if __name__ == '__main__':
    unittest.main()
