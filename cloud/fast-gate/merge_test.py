#!/usr/bin/env python3
"""cloud/fast-gate.sh and cloud/pool-job.sh gate a tip merged onto main's tip (#11ymb02), against real git repositories with ssh stubbed."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


def git(directory, *arguments):
    return subprocess.run(['git', '-C', str(directory), *arguments], check=True, capture_output=True, text=True).stdout.strip()


class GateMergeTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.origin = self.root / 'origin.git'
        subprocess.run(['git', 'init', '-q', '--bare', '-b', 'main', str(self.origin)], check=True)
        self.here = self.root / 'here'
        subprocess.run(['git', 'clone', '-q', str(self.origin), str(self.here)], check=True, capture_output=True)
        for key, value in (('user.name', 'Test'), ('user.email', 'test@example.invalid'), ('commit.gpgSign', 'false')):
            git(self.here, 'config', key, value)
        (self.here / 'cloud').mkdir()
        for name in ('fast-gate.sh', 'fast-gate-classify.sh'):
            shutil.copy(ROOT / 'cloud' / name, self.here / 'cloud' / name)
        (self.here / 'setup.go').write_text('package setup // budget 90 s\n')
        (self.here / 'feature.go').write_text('package feature\n')
        self.fork = self.commit('Base')
        # A main fix the candidate, forked before it, doesn't hold.
        (self.here / 'setup.go').write_text('package setup // budget 30 s, setup off the clock\n')
        self.main = self.commit('Main fixes the setup budget')
        git(self.here, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        self.bin = self.root / 'bin'
        self.bin.mkdir()
        # The box: records what it was asked to gate, and the tree's setup.go at that commit.
        self.script('ssh', '''printf '%s\\n' "$*" >> "$TEST_ROOT/ssh-calls"
gated=${5%% *}
echo "$gated" > "$TEST_ROOT/gated"
git -C "$TEST_HERE" show "$gated:setup.go" > "$TEST_ROOT/gated-setup.go"
''')
        self.script('scp', '''destination=${@: -1}
mkdir -p "$destination"
echo "green: $(cat "$TEST_ROOT/gated") fast gate in 1.0 s" > "$destination/status.txt"
echo '{"sha": "'"$(cat "$TEST_ROOT/gated")"'"}' > "$destination/fast.json"
if [ -f "$TEST_ROOT/big" ]; then
  python3 -c 'import json, sys; json.dump({"sha": sys.argv[2], "failed_tests": ["t%d" % n for n in range(700000)]}, open(sys.argv[1], "w"))' "$destination/fast.json" "$(cat "$TEST_ROOT/gated")"
  head -c 6000000 /dev/urandom > "$destination/determinism-adamic"
fi
''')

    def script(self, name, text):
        path = self.bin / name
        path.write_text('#!/usr/bin/env bash\n' + text)
        path.chmod(0o755)

    def commit(self, message):
        git(self.here, 'add', '-A')
        git(self.here, 'commit', '-q', '-m', message)
        return git(self.here, 'rev-parse', 'HEAD')

    def candidate(self, name, files):
        git(self.here, 'checkout', '-q', '--detach', self.fork)
        for path, text in files.items():
            (self.here / path).write_text(text)
        sha = self.commit(name)
        git(self.here, 'push', '-q', 'origin', '%s:refs/heads/%s' % (sha, name))
        git(self.here, 'checkout', '-q', '--detach', self.main)
        return sha

    def gate(self, sha, branch):
        env = dict(os.environ, PATH='%s:%s' % (self.bin, os.environ['PATH']), TEST_ROOT=str(self.root), TEST_HERE=str(self.here),
                   ADAMIC_FAST_GATE_TOOLS_ON_ORIGIN='1', ADAMIC_AI_DATABASE=str(self.root / 'none.db'), ADAMIC_FAST_GATE_BOX='box')
        return subprocess.run(['bash', str(self.here / 'cloud' / 'fast-gate.sh'), sha, '--branch', branch, '--class', 'S'],
                              env=env, capture_output=True, text=True, timeout=60)

    def record(self, sha):
        refs = git(self.origin, 'for-each-ref', '--format=%(refname)', 'refs/heads/gate-logs/%s/' % sha[:12]).split()
        self.assertEqual(len(refs), 1, refs)
        return git(self.origin, 'show', '%s:status.txt' % refs[0])

    def test_a_tip_forked_before_a_main_fix_gates_with_the_fix(self):
        sha = self.candidate('codex/feature', {'feature.go': 'package feature // more\n'})
        result = self.gate(sha, 'codex/feature')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        gated = (self.root / 'gated').read_text().strip()
        self.assertNotEqual(gated, sha)
        self.assertEqual(git(self.here, 'rev-list', '--parents', '-n', '1', gated).split()[1:], [self.main, sha])
        self.assertIn('setup off the clock', (self.root / 'gated-setup.go').read_text(), 'the box gated the fork point, not main')
        self.assertEqual(git(self.origin, 'rev-parse', 'refs/gate-merges/' + gated), gated, 'the box could not fetch the merge')
        status = self.record(sha)
        self.assertTrue(status.startswith('green: %s fast gate' % sha), status)
        self.assertIn('as %s' % gated, status)
        self.assertEqual(self.record(gated), status, 'push-main lands the merge by its own record')
        # One tip on one base is one merge commit on every attempt.
        self.gate(sha, 'codex/feature')
        self.assertEqual((self.root / 'gated').read_text().strip(), gated)

    def test_a_conflicting_tip_is_red_at_merge_without_the_box(self):
        sha = self.candidate('codex/conflict', {'setup.go': 'package setup // budget 120 s\n'})
        result = self.gate(sha, 'codex/conflict')
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertFalse((self.root / 'ssh-calls').exists(), 'a conflict spent gate time')
        status = self.record(sha)
        self.assertTrue(status.startswith('red: %s fast gate, first failure at merge' % sha), status)
        self.assertIn('setup.go', status)

    def test_a_tip_that_holds_main_gates_as_itself(self):
        git(self.here, 'checkout', '-q', '--detach', self.main)
        (self.here / 'feature.go').write_text('package feature // on main\n')
        sha = self.commit('On main')
        git(self.here, 'push', '-q', 'origin', '%s:refs/heads/codex/current' % sha)
        result = self.gate(sha, 'codex/current')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual((self.root / 'gated').read_text().strip(), sha)
        self.assertEqual(git(self.origin, 'for-each-ref', 'refs/gate-merges/'), '')
        self.assertNotIn('gated merged', self.record(sha))

    def test_a_complete_run_s_large_fast_json_stays_in_its_record_and_a_stray_binary_does_not(self):
        # Oct 9 11:07Z: the star's complete run wrote a 15.8 MB fast.json, and publishing moved it out of the record.
        (self.root / 'big').write_text('yes')
        sha = self.candidate('codex/feature', {'feature.go': 'package feature // big\n'})
        result = self.gate(sha, 'codex/feature')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        refs = git(self.origin, 'for-each-ref', '--format=%(refname)', 'refs/heads/gate-logs/%s/' % sha[:12]).split()
        files = git(self.origin, 'ls-tree', '--name-only', refs[0]).split()
        self.assertIn('fast.json', files)
        self.assertNotIn('determinism-adamic', files)
        self.assertEqual(len(json.loads(git(self.origin, 'show', '%s:fast.json' % refs[0]))['failed_tests']), 700000)

    def test_main_s_canary_gates_main_against_main_ten_landings_back(self):
        # Oct 9 10:42Z: main gated against itself selected no test (tests=0.0s), and that green promoted tools.
        git(self.here, 'checkout', '-q', '--detach', self.main)
        landings = []
        for number in range(11):
            (self.here / 'feature.go').write_text('package feature // landing %d\n' % number)
            landings.append(self.commit('Landing %d' % number))
        git(self.here, 'push', '-q', 'origin', 'HEAD:refs/heads/main')
        tip = landings[-1]
        result = self.gate(tip, 'canary/main')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn('against main~10 %s' % landings[0], result.stdout)
        self.assertEqual((self.root / 'gated').read_text().strip(), tip, 'main holds its own base: no merge')
        self.assertEqual(self.read_call().split()[5], landings[0])

    def read_call(self):
        return (self.root / 'ssh-calls').read_text().splitlines()[-1]

    def pool(self, sha, branch):
        shutil.copy(ROOT / 'cloud' / 'pool-job.sh', self.here / 'cloud' / 'pool-job.sh')
        jobs = self.root / 'jobs'
        # Not taken at once and no time to wait, so the waiter asks for its box race and returns.
        env = dict(os.environ, LOOM_FAST_JOBS=str(jobs), ADAMIC_POOL_TAKE_SECONDS='0', ADAMIC_POOL_JOB_SECONDS='0',
                   ADAMIC_FAST_GATE_WATCH_STATE=str(self.root / 'watch'))
        result = subprocess.run(['bash', str(self.here / 'cloud' / 'pool-job.sh'), sha, '--branch', branch, '--tools', 'tools', '--priority', '10'],
                                env=env, capture_output=True, text=True, timeout=60)
        job = jobs / (sha + '.json')
        return result, (json.loads(job.read_text()) if job.exists() else None)

    def test_the_pool_job_names_the_merge_and_answers_a_conflict_itself(self):
        sha = self.candidate('codex/feature', {'feature.go': 'package feature // more\n'})
        result, job = self.pool(sha, 'codex/feature')
        self.assertEqual(git(self.here, 'rev-list', '--parents', '-n', '1', job['gate']).split()[1:], [self.main, sha], result.stdout)
        self.assertEqual(job['sha'], sha)
        self.assertNotIn('complete', job, 'side work runs the fast plan')
        self.assertEqual(git(self.origin, 'rev-parse', 'refs/gate-merges/' + job['gate']), job['gate'])
        # Not started in time, it asks the watcher for a box race and leaves its job queued at Loom (#04gypqe), even once its
        # waiter gives up (no time to wait here): the job's record still lands the candidate.
        self.assertEqual((self.root / 'watch' / 'race-wanted' / sha).read_text().strip(), 'codex/feature')
        self.assertFalse((self.root / 'jobs' / (sha + '.cancel')).exists())
        # A new job of the same sha clears a cancel left by an earlier one, so the fresh job isn't stopped before it starts.
        (self.root / 'jobs' / (sha + '.cancel')).write_text('left by an earlier job\n')
        self.pool(sha, 'codex/feature')
        self.assertFalse((self.root / 'jobs' / (sha + '.cancel')).exists())
        # Every unit placed by the deadline: no race.
        placed = self.candidate('codex/placed', {'feature.go': 'package feature // placed\n'})
        (self.root / 'jobs').mkdir(exist_ok=True)
        (self.root / 'jobs' / (placed + '.placed')).write_text('13 13\n')
        self.pool(placed, 'codex/placed')
        self.assertFalse((self.root / 'watch' / 'race-wanted' / placed).exists())
        # Landings and areas run complete on the pool, stage 3 lane included (#vaf0xwc).
        landing = self.candidate('cloud/land-x', {'feature.go': 'package feature // landing\n'})
        self.assertIs(self.pool(landing, 'cloud/land-x')[1]['complete'], True)
        conflict = self.candidate('codex/conflict', {'setup.go': 'package setup // budget 120 s\n'})
        result, job = self.pool(conflict, 'codex/conflict')
        self.assertIsNone(job, 'a conflict spent pool time')
        self.assertEqual(result.returncode, 0)
        self.assertTrue(result.stdout.splitlines()[-1].startswith('red: %s fast gate on Loom\'s side pool, first failure at merge' % conflict), result.stdout)


if __name__ == '__main__':
    unittest.main()
