#!/usr/bin/env python3
"""push-main's landings, end to end in a scratch repository (@system_adamic, Oct 9): each landing is one commit on
main, first parent the old main, second the landed sha, its tree exactly the landing's and its numbers as trailers,
and landings.py reads them back as the velocity table. Main moving by test-only commits doesn't spend a candidate's
gate unless they touch a package where the candidate changes code.

usage: python3 cloud/integration/push-main-landing-test.py
"""
import csv
import gzip
import io
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

directory = Path(__file__).parent
identity = {'GIT_AUTHOR_NAME': 't', 'GIT_AUTHOR_EMAIL': 't@t', 'GIT_COMMITTER_NAME': 't', 'GIT_COMMITTER_EMAIL': 't@t'}


def git(where, *arguments):
    return subprocess.run(['git', '-C', str(where)] + list(arguments), check=True, capture_output=True, text=True,
                          env=dict(os.environ, **identity)).stdout.strip()


class LandingTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        origin = root / 'origin.git'
        git(root, 'init', '-q', '--bare', str(origin))
        self.repository = root / 'repository'
        git(root, 'init', '-q', str(self.repository))
        git(self.repository, 'remote', 'add', 'origin', str(origin))
        self.write('code/a.go', 'package code\n')
        self.write('other/b.go', 'package other\n')
        self.main = self.commit('main')
        git(self.repository, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        # Developer tools' declared tools, which the lane's checks read.
        self.write('cloud/fast-gate/tools.txt', 'go\tall\tgo version\n')
        git(self.repository, 'push', '-q', 'origin', self.commit('tools') + ':refs/heads/devtools/fast-gate')
        git(self.repository, 'checkout', '-q', '--detach', self.main)
        # The scripts under test, beside a merge-back that does nothing.
        self.scripts = root / 'scripts'
        self.scripts.mkdir()
        for name in ('push-main.sh', 'landings.py', 'lane-checks.py', 'rerun_merge.py'):
            shutil.copy(directory / name, self.scripts / name)
        (self.scripts / 'merge-back.sh').write_text('#!/usr/bin/env bash\n')
        (self.scripts / 'merge-back.sh').chmod(0o755)

    def write(self, path, text):
        (self.repository / path).parent.mkdir(parents=True, exist_ok=True)
        (self.repository / path).write_text(text)

    def commit(self, message, on=None):
        if on:
            git(self.repository, 'checkout', '-q', '--detach', on)
        git(self.repository, 'add', '-A')
        git(self.repository, 'commit', '-qm', message)
        return git(self.repository, 'rev-parse', 'HEAD')

    def change(self, on, path, text, message):
        git(self.repository, 'checkout', '-q', '--detach', on)
        self.write(path, text)
        return self.commit(message)

    def push(self, *arguments):
        # The star's train and the fast gate's watcher keep their state in scratch directories here.
        root = Path(self.tmp.name)
        return subprocess.run(['bash', str(self.scripts / 'push-main.sh')] + list(arguments), cwd=self.repository,
                              capture_output=True, text=True, env=dict(os.environ, ADAMIC_STAR_FILE=str(root / 'star'),
                                                                       ADAMIC_FAST_GATE_STATE=str(root / 'watch'),
                                                                       ADAMIC_LANE_TREE=str(root / 'lane'), **identity))

    def main_now(self):
        git(self.repository, 'fetch', '-q', 'origin')
        return git(self.repository, 'rev-parse', 'origin/main')

    def assertLanded(self, landed, old, sha):
        self.assertEqual(landed.returncode, 0, landed.stdout + landed.stderr)
        new = self.main_now()
        self.assertEqual(git(self.repository, 'rev-list', '--parents', '-n', '1', new).split()[1:], [old, sha])
        return new

    def test_a_gated_landing_is_one_commit_with_its_numbers_as_trailers(self):
        sha = self.change(self.main, 'code/a.go', 'package code\n\n// new\n', 'code change')
        new = self.assertLanded(self.push(sha, '12', '40', '0', '3', 'code/branch, at 1234'), self.main, sha)
        self.assertEqual(git(self.repository, 'rev-parse', new + '^{tree}'), git(self.repository, 'rev-parse', sha + '^{tree}'))
        trailers = git(self.repository, 'log', '-1', '--format=%(trailers:only,unfold)', new)
        for line in ('Old-main: ' + self.main, 'Landed-commits: 2', 'Gate-minutes: 12', 'Pass: 40', 'Fail: 0', 'Skip: 3', 'Branches: code/branch, at 1234'):
            self.assertIn(line, trailers)
        # Main moved once: nothing on top of the landing commit.
        self.assertEqual(git(self.repository, 'rev-list', '--count', sha + '..' + new), '1')
        rows = list(csv.DictReader(io.StringIO(subprocess.run(['python3', str(self.scripts / 'landings.py')], cwd=self.repository,
                                                              capture_output=True, text=True, check=True).stdout)))
        self.assertEqual([(row['new_main'], row['old_main'], row['gate_minutes']) for row in rows], [(new, self.main, '12')])
        self.assertTrue(rows[0]['branches_landed'].startswith('code/branch; at 1234; lane checks '), rows[0]['branches_landed'])

    def test_a_test_only_landing_is_one_commit_too(self):
        sha = self.change(self.main, 'code/a_test.go', 'package code\n', 'a test')
        new = self.assertLanded(self.push('--test-only', sha, 'a split'), self.main, sha)
        self.assertIn('Gate-minutes: 0', git(self.repository, 'log', '-1', '--format=%B', new))
        refused = self.push('--test-only', self.change(new, 'code/a.go', 'package code\n\n// x\n', 'code'), 'not a split')
        self.assertIn('not test-only', refused.stderr)
        # The lane's checks refuse a test that shells out to a tool the gate doesn't declare, or isn't gofmt'd.
        tool = self.change(new, 'code/tool_test.go', 'package code\n\nimport "os/exec"\n\nvar _ = exec.Command("timeout", "1")\n', 'a tool')
        self.assertIn("runs timeout, which cloud/fast-gate/tools.txt doesn't declare", self.push('--test-only', tool, 'tool').stderr)
        messy = self.change(new, 'code/messy_test.go', 'package code\nvar  x = 1\n', 'messy')
        self.assertIn("code/messy_test.go isn't gofmt-formatted", self.push('--test-only', messy, 'messy').stderr)
        # A branch whose net diff is tests but whose history changes code is refused by that commit: landing it
        # would record the code commit as merged with its change dropped (cohere's estree split, Oct 9).
        code = self.change(new, 'code/a.go', 'package code\n\n// carried\n', 'a code change')
        undone = self.change(code, 'code/a.go', git(self.repository, 'show', new + ':code/a.go') + '\n', 'undo it')
        carried = self.change(undone, 'code/carried_test.go', 'package code\n', 'a test on top')
        refused = self.push('--test-only', carried, 'carried')
        self.assertIn('carries non-test history: %s' % code[:8], refused.stderr)

    def test_a_deletion_no_code_reads_lands_ungated(self):
        self.write('documentation/table.csv', 'a,b\n')
        self.write('documentation/read.csv', 'a,b\n')
        self.write('code/reader.go', 'package code\n\nconst table = "documentation/read.csv"\n')
        base = self.commit('records')
        git(self.repository, 'push', '-q', 'origin', base + ':refs/heads/main')
        git(self.repository, 'rm', '-q', 'documentation/table.csv')
        deleted = self.commit('delete the table')
        new = self.assertLanded(self.push('--deletion', deleted, 'table out'), base, deleted)
        message = git(self.repository, 'log', '-1', '--format=%B', new)
        self.assertIn('Land deletion %s over main %s' % (deleted[:8], base[:8]), message)
        self.assertIn('Gate-minutes: 0', message)
        self.assertIn('deletion no code reads, no gate', message)
        # A file code names, a change beside the deletion, or history that edits on the way all need a gate.
        git(self.repository, 'checkout', '-q', '--detach', new)
        git(self.repository, 'rm', '-q', 'documentation/read.csv')
        read = self.commit('delete a read file')
        self.assertIn('code reads documentation/read.csv: code/reader.go', self.push('--deletion', read, 'read').stderr)
        # A file checked by hand to name the path without reading it is passed by name, and the landing says so.
        named = self.assertLanded(self.push('--deletion', '--not-a-reader', 'code/reader.go', read, 'read'), new, read)
        self.assertIn('Named but not read, checked by hand: code/reader.go', git(self.repository, 'log', '-1', '--format=%B', named))
        new = named
        git(self.repository, 'checkout', '-q', '--detach', new)
        self.write('other/b.go', 'package other\n\n// beside\n')
        git(self.repository, 'rm', '-q', 'code/a.go')
        beside = self.commit('delete and edit')
        self.assertIn('not deletions only against main', self.push('--deletion', beside, 'beside').stderr)
        edited = self.change(new, 'other/c.txt', 'x\n', 'add a file')
        git(self.repository, 'rm', '-q', 'other/c.txt')
        undone = self.commit('delete it again')
        git(self.repository, 'rm', '-q', 'code/a.go')
        carried = self.commit('then a deletion')
        self.assertIn('carries more than deletions: %s' % edited[:8], self.push('--deletion', carried, 'carried').stderr)

    def test_a_doc_lands_on_its_markdown_corpus_verdict_only(self):
        # @system_adamic, Oct 9 02:16: a doc no code reads except the markdown corpora lands on those units alone,
        # through the test-only path, carrying the verdict. Without it, or with any non-Markdown file, it's refused.
        doc = self.change(self.main, 'docs/convention.md', '# Convention\n', 'a doc')
        self.assertIn('not test-only', self.push('--test-only', doc, 'doc').stderr)
        landed = self.assertLanded(self.push('--test-only', '--markdown-corpus', 'corpusfiles, markdowninline, markdownblocks census green on Home', doc, 'doc'), self.main, doc)
        self.assertIn('Markdown gated by its corpus alone', git(self.repository, 'log', '-1', '--format=%B', landed))
        mixed = self.change(landed, 'docs/more.md', '# More\n', 'a doc')
        mixed = self.change(mixed, 'code/a.go', 'package code\n\n// beside\n', 'and code')
        self.assertIn('not test-only', self.push('--test-only', '--markdown-corpus', 'green', mixed, 'mixed').stderr)

    def test_a_hyphenated_python_test_is_test_only_unless_a_non_test_file_runs_it(self):
        # @system_adamic, Oct 9 04:20: *-test.py takes the lane, but one a gate script or .sh names is gate logic.
        alone = self.change(self.main, 'cloud/check-test.py', 'print("ok")\n', 'a python test')
        landed = self.assertLanded(self.push('--test-only', alone, 'alone'), self.main, alone)
        runner = self.change(landed, 'cloud/run.sh', '#!/bin/sh\npython3 cloud/check-test.py\n', 'a gate script that runs it')
        self.assertLanded(self.push(runner, '1', '1', '0', '0', 'runner'), landed, runner)
        edited = self.change(self.main_now(), 'cloud/check-test.py', 'print("changed")\n', 'edit the test a script runs')
        refused = self.push('--test-only', edited, 'edited')
        self.assertIn('cloud/check-test.py is named by cloud/run.sh', refused.stderr)

    def test_a_candidate_built_ahead_lands_over_the_landing_commit_below_it(self):
        lower = self.change(self.main, 'code/a.go', 'package code\n\n// lower\n', 'lower')
        upper = self.change(lower, 'other/b.go', 'package other\n\n// upper\n', 'upper')
        landed = self.assertLanded(self.push(lower, '1', '1', '0', '0', 'lower'), self.main, lower)
        top = self.assertLanded(self.push(upper, '1', '1', '0', '0', 'upper'), landed, upper)
        self.assertEqual(git(self.repository, 'rev-parse', top + '^{tree}'), git(self.repository, 'rev-parse', upper + '^{tree}'))

    def test_main_moving_by_tests_keeps_a_gate_outside_the_candidates_packages(self):
        candidate = self.change(self.main, 'code/a.go', 'package code\n\n// candidate\n', 'candidate')
        elsewhere = self.change(self.main, 'other/b_test.go', 'package other\n', 'a test elsewhere')
        moved = self.assertLanded(self.push('--test-only', elsewhere, 'elsewhere'), self.main, elsewhere)
        landed = self.assertLanded(self.push(candidate, '1', '1', '0', '0', 'candidate'), moved, candidate)
        self.assertEqual(git(self.repository, 'show', landed + ':other/b_test.go'), 'package other')
        # A test in the candidate's own package meets its new code, so the gate is spent.
        second = self.change(landed, 'code/a.go', 'package code\n\n// second\n', 'second')
        same = self.change(landed, 'code/a_test.go', 'package code\n', 'a test beside it')
        self.assertLanded(self.push('--test-only', same, 'beside'), landed, same)
        refused = self.push(second, '1', '1', '0', '0', 'second')
        self.assertIn('beyond record and test-only commits (code/a_test.go)', refused.stderr)
        # It names B, the moved main merged with the gated sha, which descends from the gate (#mbexftz).
        moved = refused.stderr.split('on B ')[1].split()[0]
        self.assertEqual(git(self.repository, 'rev-list', '--parents', '-n', '1', moved).split()[1:], [self.main_now(), second])
        self.assertEqual(git(self.repository, 'rev-parse', moved + '^{tree}'),
                         git(self.repository, 'merge-tree', '--write-tree', self.main_now(), second).splitlines()[0])


    def test_a_moved_main_lands_on_a_rerun_of_the_units_it_changed(self):
        # #mbexftz: main moves by a test in the candidate's package, push-main names B (main merged with the gated
        # sha), and a rerun of only the moved unit at B lands B on the gate's kept verdicts, saying so in trailers.
        candidate = self.change(self.main, 'code/a.go', 'package code\n\n// candidate\n', 'candidate')
        units = lambda **verdicts: [{'id': name, 'input_sha256': digest, 'verdict': verdict} for name, (digest, verdict) in verdicts.items()]
        base = self.publish(candidate, 'base', {'sha': candidate, 'base': self.main, 'finished': True,
                                                'units': units(code=('c1', 'passed'), other=('o1', 'passed'))}, kind='full')
        beside = self.change(self.main, 'code/a_test.go', 'package code\n', 'a test beside it')
        moved = self.assertLanded(self.push('--test-only', beside, 'beside'), self.main, beside)
        refused = self.push(candidate, '1', '1', '0', '0', 'candidate')
        b = refused.stderr.split('on B ')[1].split()[0]
        git(self.repository, 'push', '-q', 'origin', b + ':refs/heads/moved-b')
        record = {'sha': b, 'base': moved, 'finished': True, 'skip': 0, 'packages': ['code'], 'fail': 0, 'pass': 3,
                  'build_ok': True, 'vet_ok': True, 'uncached_tests': True, 'wall_seconds': 60,
                  'steps_seconds': {stage: 5 for stage in ('build', 'vet', 'tests', 'smoke', 'census')},
                  'stages_exit': {stage: 0 for stage in ('build', 'vet', 'tests', 'smoke', 'census')},
                  'planned_stages': ['build', 'vet', 'tests', 'smoke', 'census'],
                  'rerun_of': base, 'units': units(code=('c2', 'passed'), other=('o1', 'kept'))}
        rerun = self.publish(b, 'rerun', record)
        landed = self.assertLanded(self.push('--fast-gate', rerun, b, 'moved'), moved, b)
        trailers = git(self.repository, 'log', '-1', '--format=%(trailers:only,unfold)', landed)
        self.assertIn('Rerun-units: code', trailers)
        self.assertIn('Moved-main: %s..%s' % (self.main, moved), trailers)

    def test_a_moved_main_whose_test_breaks_the_candidate_is_refused_on_the_rerun(self):
        # #mbexftz's proof: main takes a test in the candidate's package that fails against its code. The rerun on B
        # runs that unit and it's red, so B is refused. And a rerun that keeps every old verdict (the moved unit kept,
        # carrying the input hash it has at B) is refused too: a kept unit must have the gate's hash.
        candidate = self.change(self.main, 'code/a.go', 'package code\n\n// candidate\n', 'candidate')
        units = lambda **verdicts: [{'id': name, 'input_sha256': digest, 'verdict': verdict} for name, (digest, verdict) in verdicts.items()]
        base = self.publish(candidate, 'base', {'sha': candidate, 'base': self.main, 'finished': True,
                                                'units': units(code=('c1', 'passed'), other=('o1', 'passed'))}, kind='full')
        breaking = self.change(self.main, 'code/a_test.go', 'package code\n\n// asserts the old behavior\n', 'a test that breaks it')
        moved = self.assertLanded(self.push('--test-only', breaking, 'breaking'), self.main, breaking)
        b = self.push(candidate, '1', '1', '0', '0', 'candidate').stderr.split('on B ')[1].split()[0]
        git(self.repository, 'push', '-q', 'origin', b + ':refs/heads/moved-b')
        def rerun(name, failing, rerunUnits):
            record = {'sha': b, 'base': moved, 'finished': True, 'skip': 0, 'packages': ['code'], 'fail': len(failing), 'pass': 3,
                      'build_ok': True, 'vet_ok': True, 'uncached_tests': True, 'wall_seconds': 60,
                      'steps_seconds': {stage: 5 for stage in ('build', 'vet', 'tests', 'smoke', 'census')},
                      'stages_exit': dict({stage: 0 for stage in ('build', 'vet', 'smoke', 'census')}, tests=1 if failing else 0),
                      'planned_stages': ['build', 'vet', 'tests', 'smoke', 'census'], 'rerun_of': base, 'units': rerunUnits}
            return self.publish(b, name, record, failing=failing)
        red = rerun('red', [('code', 'TestOld')], units(code=('c2', 'failed'), other=('o1', 'kept')))
        refused = self.push('--fast-gate', red, b, 'red')
        self.assertNotEqual(refused.returncode, 0)
        self.assertIn('rerun unit code is failed', refused.stderr)
        kept = rerun('kept', [], units(code=('c2', 'kept'), other=('o1', 'kept')))
        refused = self.push('--fast-gate', kept, b, 'kept')
        self.assertNotEqual(refused.returncode, 0)
        self.assertIn('unit code is kept but its inputs changed (c1 to c2), so it must rerun', refused.stderr)
        self.assertEqual(self.main_now(), moved)

    def test_a_moved_main_under_a_whole_pool_record_starts_loom_s_rerun_on_b_and_holds(self):
        # #mbexftz: on a whole record, push-main pushes B (main merged with the gated sha) and starts Loom's rerun.sh on
        # it, which reruns only the units whose input hash moved; push-main holds (exit 3) until that record lands B.
        candidate = self.change(self.main, 'code/a.go', 'package code\n\n// candidate\n', 'candidate')
        git(self.repository, 'push', '-q', 'origin', candidate + ':refs/heads/candidate')
        stages = ('coverage', 'tools', 'build', 'vet', 'tests', 'wasi', 'stage3', 'catalog', 'determinism', 'census')
        record = {'sha': candidate, 'base': self.main, 'finished': True, 'skip': 0, 'packages': 'all', 'package_list': ['code', 'other'],
                  'fail': 0, 'pass': 3, 'build_ok': True, 'vet_ok': True, 'uncached_tests': True, 'wall_seconds': 60,
                  'steps_seconds': {stage: 5 for stage in stages}, 'stages_exit': {stage: 0 for stage in stages}, 'planned_stages': list(stages)}
        whole = self.publish(candidate, 'main', record, kind='full')
        beside = self.change(self.main, 'code/a_test.go', 'package code\n', 'a test beside it')
        moved = self.assertLanded(self.push('--test-only', beside, 'beside'), self.main, beside)
        root = Path(self.tmp.name)
        fake = root / 'rerun.sh'
        fake.write_text('#!/usr/bin/env bash\necho "$@" > %s\n' % (root / 'rerun-called'))
        fake.chmod(0o755)
        environment = dict(LOOM_RERUN=str(fake), ADAMIC_RERUN_LOGS=str(root / 'reruns'))
        held = subprocess.run(['bash', str(self.scripts / 'push-main.sh'), '--full-gate', whole, candidate, 'candidate'], cwd=self.repository,
                              capture_output=True, text=True, env=dict(os.environ, ADAMIC_STAR_FILE=str(root / 'star'),
                              ADAMIC_FAST_GATE_STATE=str(root / 'watch'), ADAMIC_LANE_TREE=str(root / 'lane'), **environment, **identity))
        self.assertEqual(held.returncode, 3, held.stdout + held.stderr)
        b = held.stderr.split('on B ')[1].split()[0]
        self.assertEqual(git(self.repository, 'rev-list', '--parents', '-n', '1', b).split()[1:], [moved, candidate])
        self.assertEqual(git(self.repository, 'ls-remote', 'origin', 'refs/gate-merges/' + b).split()[0], b)
        for _ in range(50):
            if (root / 'rerun-called').exists():
                break
            __import__('time').sleep(0.1)
        self.assertEqual((root / 'rerun-called').read_text().split(), [whole, b])
        self.assertEqual(self.main_now(), moved)

    def test_a_test_beside_the_stars_code_waits_while_its_gate_runs(self):
        root = Path(self.tmp.name)
        star = self.change(self.main, 'code/a.go', 'package code\n\n// the star\n', 'the star')
        git(self.repository, 'push', '-q', 'origin', star + ':refs/heads/cloud/land-train-9-slice-' + star[:8])
        (root / 'star').write_text('cloud/land-train-9-slice-%s\n' % star[:8])
        (root / 'watch' / 'running').mkdir(parents=True)
        (root / 'watch' / 'running' / '123').write_text('cloud/land-train-9-slice-%s %s S box B x:1 log\n' % (star[:8], star))
        beside = self.change(self.main, 'code/a_test.go', 'package code\n', 'a test beside the star')
        held = self.push('--test-only', beside, 'beside')
        self.assertEqual(held.returncode, 3, held.stdout + held.stderr)
        self.assertIn('held for the star: cloud/land-train-9-slice-%s %s has its fast gate running and changes code in code' % (star[:8], star[:8]), held.stderr)
        # Elsewhere keeps flowing.
        elsewhere = self.change(self.main, 'other/b_test.go', 'package other\n', 'elsewhere')
        moved = self.assertLanded(self.push('--test-only', elsewhere, 'elsewhere'), self.main, elsewhere)
        # The gate finishes: the held change lands.
        (root / 'watch' / 'running' / '123').unlink()
        self.assertLanded(self.push('--test-only', beside, 'beside'), moved, beside)


    def publish(self, sha, name, record, failing=(), kind='fast', jsonOnly=False):
        # A gate-logs record on origin: status.txt, its json and a test record, in a commit of their own.
        tree = Path(self.tmp.name) / ('record-' + name + sha[:6])
        tree.mkdir()
        (tree / (kind + '.json')).write_text(json.dumps(record))
        (tree / 'status.txt').write_text('%s: %s %s gate\n' % ('red' if failing else 'green', sha, kind))
        if jsonOnly:
            # A box's fast record: no test.jsonl.gz, the failing tests named in its json.
            record = dict(record, failed_tests=['github.com/system-inc/adamic/%s %s' % pair for pair in failing])
            (tree / (kind + '.json')).write_text(json.dumps(record))
        else:
            events = [{'Action': 'fail', 'Package': 'github.com/system-inc/adamic/' + package, 'Test': test} for package, test in failing]
            events.append({'Action': 'pass', 'Package': 'github.com/system-inc/adamic/other', 'Test': 'TestFine'})
            (tree / 'test.jsonl.gz').write_bytes(gzip.compress('\n'.join(json.dumps(event) for event in events).encode()))
        index = str(Path(self.tmp.name) / ('index-' + name))
        environment = dict(os.environ, GIT_INDEX_FILE=index, **identity)
        gitDirectory = git(self.repository, 'rev-parse', '--absolute-git-dir')
        subprocess.run(['git', '--git-dir', gitDirectory, '--work-tree', str(tree), 'add', '-A', '.'], env=environment, check=True)
        treeSha = subprocess.run(['git', '--git-dir', gitDirectory, 'write-tree'], env=environment, check=True, capture_output=True, text=True).stdout.strip()
        reference = 'gate-logs/%s/20261009T000000Z/%s' % (sha[:12], 'full-main' if name == 'main' else name)
        git(self.repository, 'push', '-q', 'origin', '%s:refs/heads/%s' % (git(self.repository, 'commit-tree', treeSha, '-m', 'record'), reference))
        return reference

    def test_a_go_tests_only_record_lands_only_beside_a_record_of_the_other_stages(self):
        sha = self.change(self.main, 'other/b.go', 'package other\n\n// pooled\n', 'pooled')
        common = {'sha': sha, 'base': self.main, 'finished': True, 'skip': 0, 'packages': ['other']}
        tests = self.publish(sha, 'fast', dict(common, runner='pool', covers='go-tests', fail=0, **{'pass': 10}, uncached_tests=True, wall_seconds=100,
                                               steps_seconds={'tests': 90}, stages_exit={'tests': 0}, planned_stages=['tests']))
        stages = self.publish(sha, 'stages', dict(common, fail=0, **{'pass': 0}, build_ok=True, vet_ok=True, wall_seconds=50,
                                                  steps_seconds={stage: 5 for stage in ('build', 'vet', 'smoke', 'census')},
                                                  stages_exit={stage: 0 for stage in ('build', 'vet', 'smoke', 'census')},
                                                  planned_stages=['build', 'vet', 'smoke', 'census']))
        alone = self.push('--fast-gate', tests, sha, 'pooled')
        self.assertNotEqual(alone.returncode, 0)
        self.assertIn('go build or go vet failed', alone.stderr)
        self.assertLanded(self.push('--fast-gate', tests, '--also-gate', stages, sha, 'pooled'), self.main, sha)


    def test_a_candidate_lands_with_only_the_reds_main_already_has(self):
        known = ('stage1/cohere/estree', 'TestScalarEdges')
        main = self.publish(self.main, 'main', {'sha': self.main, 'finished': True, 'fail': 1}, failing=[known], kind='full')
        def candidate(text, failing, jsonOnly=False):
            sha = self.change(self.main, 'other/b.go', 'package other\n\n// %s\n' % text, text)
            record = {'sha': sha, 'base': self.main, 'finished': True, 'skip': 0, 'packages': ['other'], 'fail': len(failing), 'pass': 10,
                      'build_ok': True, 'vet_ok': True, 'uncached_tests': True, 'wall_seconds': 60,
                      'steps_seconds': {stage: 5 for stage in ('build', 'vet', 'tests', 'smoke', 'census')},
                      'stages_exit': dict({stage: 0 for stage in ('build', 'vet', 'smoke', 'census')}, tests=1 if failing else 0),
                      'planned_stages': ['build', 'vet', 'tests', 'smoke', 'census']}
            return sha, self.publish(sha, 'fast', record, failing=failing, jsonOnly=jsonOnly)
        sha, record = candidate('known', [known])
        self.assertIn('1 failures', self.push('--fast-gate', record, sha, 'known').stderr)
        landed = self.push('--fast-gate', record, '--main-reds', main, sha, 'known')
        self.assertLanded(landed, self.main, sha)
        self.assertIn('lands with 1 reds main already has', git(self.repository, 'log', '-1', '--format=%B', 'origin/main'))
        # A box record names its reds in fast.json, not a test.jsonl.gz: the same diff applies.
        boxed, boxedRecord = candidate('boxed', [known, ('code', 'TestBoxed')], jsonOnly=True)
        self.assertIn('new reds against main: code TestBoxed', self.push('--fast-gate', boxedRecord, '--main-reds', main, boxed, 'boxed').stderr)
        fresh, freshRecord = candidate('fresh', [known, ('code', 'TestNew')])
        refused = self.push('--fast-gate', freshRecord, '--main-reds', main, fresh, 'fresh')
        self.assertIn('new reds against main: code TestNew', refused.stderr)
        # A red ruled main's while its fix is in flight counts as main's by that exact name, and the landing names it.
        now = self.main_now()
        ruled = self.change(now, 'other/c.go', 'package other\n\n// ruled\n', 'ruled')
        ruledRecord = self.publish(ruled, 'ruled', {'sha': ruled, 'base': now, 'finished': True, 'skip': 0, 'packages': ['other'], 'fail': 2, 'pass': 10,
                                   'build_ok': True, 'vet_ok': True, 'uncached_tests': True, 'wall_seconds': 60,
                                   'steps_seconds': {stage: 5 for stage in ('build', 'vet', 'tests', 'smoke', 'census')},
                                   'stages_exit': dict({stage: 0 for stage in ('build', 'vet', 'smoke', 'census')}, tests=1),
                                   'planned_stages': ['build', 'vet', 'tests', 'smoke', 'census']}, failing=[known, ('code', 'TestRuled')])
        self.assertIn('new reds against main: code TestRuled', self.push('--fast-gate', ruledRecord, '--main-reds', main, ruled, 'ruled').stderr)
        self.assertIn('new reds against main: code TestNew', self.push('--fast-gate', freshRecord, '--main-reds', main, '--infra-red', 'code TestRuled=#t4b9j71', fresh, 'fresh').stderr)
        landed = self.assertLanded(self.push('--fast-gate', ruledRecord, '--main-reds', main, '--infra-red', 'code TestRuled=#t4b9j71 timing kill', ruled, 'ruled'),
                                   now, ruled)
        self.assertIn("ruled infra, not the candidate's: code TestRuled=#t4b9j71 timing kill", git(self.repository, 'log', '-1', '--format=%B', landed))
        # Alone, without a main record: only the ruled name is excused.
        now = self.main_now()
        alone = self.change(now, 'other/d.go', 'package other\n\n// alone\n', 'alone')
        def aloneRecord(name, failing):
            return self.publish(alone, name, {'sha': alone, 'base': now, 'finished': True, 'skip': 0, 'packages': ['other'], 'fail': len(failing), 'pass': 10,
                                'build_ok': True, 'vet_ok': True, 'uncached_tests': True, 'wall_seconds': 60,
                                'steps_seconds': {stage: 5 for stage in ('build', 'vet', 'tests', 'smoke', 'census')},
                                'stages_exit': dict({stage: 0 for stage in ('build', 'vet', 'smoke', 'census')}, tests=1),
                                'planned_stages': ['build', 'vet', 'tests', 'smoke', 'census']}, failing=failing)
        twoReds = aloneRecord('two', [('code', 'TestRuled'), ('code', 'TestOther')])
        self.assertIn('reds not ruled infra: code TestOther', self.push('--fast-gate', twoReds, '--infra-red', 'code TestRuled=#t4b9j71', alone, 'two').stderr)
        oneRed = aloneRecord('one', [('code', 'TestRuled')])
        # Its one red is excused, so it reaches the pause rule, which main's red record here still holds.
        self.assertIn('landings are paused', self.push('--fast-gate', oneRed, '--infra-red', 'code TestRuled=#t4b9j71', alone, 'one').stderr)


if __name__ == '__main__':
    unittest.main()
