#!/usr/bin/env python3
"""Real 60 second launchd intervals, with separate scratch repositories per case.

usage: python3 cloud/integration/gate-lane-timing-test.py
"""
import concurrent.futures
import importlib.util
import os
from pathlib import Path
import subprocess
import sys
import time
import unittest

sys.dont_write_bytecode = True

spec = importlib.util.spec_from_file_location('lane_tests', Path(__file__).with_name('gate-lane-test.py'))
fixtures = importlib.util.module_from_spec(spec)
spec.loader.exec_module(fixtures)


def check_landing(landed, pushed):
    assert landed.exists(), 'no landing timestamp: push-main never landed the planted record'
    delay = float(landed.read_text().splitlines()[0]) - pushed
    assert 0 <= delay <= 120, 'landing delay %.3f s exceeds 120 s' % delay
    return delay


def scenario(kind):
    fixture = fixtures.GateLaneTests()
    fixture.setUp()
    try:
        root = fixture.root
        landed = root / 'landed'
        fixture.candidates.write_text('cloud/land-timing\t%s\ttask1\t\n' % fixture.sha)
        helper = root / 'push.py'
        helper.write_text("import subprocess, sys, time\nfrom pathlib import Path\n"
            "ref = sys.argv[2]\n"
            "status = subprocess.run(['git', 'show', 'origin/' + ref + ':status.txt'], "
            "check=True, capture_output=True, text=True, timeout=25).stdout\n"
            "if status.startswith('green'):\n"
            "    with open(%r, 'a') as log: log.write(str(time.time()) + '\\n')\n"
            "    print('Pushed main scratch'); sys.exit(0)\n"
            "print('refused: planted red record', file=sys.stderr); sys.exit(1)\n" % str(landed))
        push = root / 'push-main.sh'
        push.write_text('exec python3 "%s" "$@"\n' % helper)
        poster = root / 'poster.py'
        poster.write_text("#!/usr/bin/env python3\nimport sys\nfrom pathlib import Path\n"
            "with open(%r, 'a') as log: log.write(Path(sys.argv[sys.argv.index('--text-file') + 1]).read_text() + '\\n')\n"
            "print('Comment added')\n" % str(fixture.posts))
        poster.chmod(0o755)
        lane = fixtures.script
        if kind == 'mutant':
            lane = root / 'gate-lane-mutant.py'
            source = fixtures.script.read_text()
            original = '["bash", pushMain, *arguments, gated, "%s (gate lane)" % branch]'
            assert original in source, 'push-main mutant site missing'
            lane.write_text(source.replace(original, '["true"]'))
        environment = dict(os.environ, GATE_LANE_CANDIDATES=str(fixture.candidates),
            GATE_LANE_STATE=str(root / 'state'), GATE_LANE_PUSH_MAIN=str(push), GATE_LANE_AHRA=str(poster))
        start = time.monotonic()
        pushed = None
        for iteration in range(3):
            # Runs are synchronous: a late run cannot overlap the next interval.
            remaining = start + iteration * 60 - time.monotonic()
            while remaining > 0:
                time.sleep(min(remaining, 1))
                remaining = start + iteration * 60 - time.monotonic()
            ran = subprocess.run(['python3', str(lane)], cwd=fixture.repository,
                env=environment, capture_output=True, text=True, timeout=30)
            assert ran.returncode == 0, ran.stderr
            print('%s iteration %d: %s' % (kind, iteration, ran.stdout.strip()), flush=True)
            if iteration == 0:
                fixture.record('20261009T110000Z', 'fast',
                    'red: planted' if kind == 'red' else 'green: planted')
                pushed = time.time()
            elif kind == 'green' and landed.exists():
                break
        if kind == 'red':
            assert fixture.posts.read_text().count("doesn't land: refused: planted red record") == 1
            assert fixture.sha in fixture.candidates.read_text()
            assert not landed.exists()
            return 'red: refusal posted exactly once; candidate retained across three iterations'
        if kind == 'mutant':
            try:
                check_landing(landed, pushed)
            except AssertionError as error:
                return 'true mutant caught by timing assertion: ' + str(error)
            raise AssertionError('true mutant survived the timing assertion')
        delay = check_landing(landed, pushed)
        assert fixture.sha not in fixture.candidates.read_text()
        return 'green: measured landing delay %.3f s (limit 120 s)' % delay
    finally:
        fixture.doCleanups()


class TimingTests(unittest.TestCase):
    def test_launchd_and_push_main_mutant(self):
        with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
            futures = [pool.submit(scenario, kind) for kind in ('green', 'red', 'mutant')]
            for future in futures:
                print(future.result(timeout=180), flush=True)


if __name__ == '__main__':
    unittest.main()
