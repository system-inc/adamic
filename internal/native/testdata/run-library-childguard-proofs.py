#!/usr/bin/env python3
"""Exercise the library deadline sites with silent mutants and CPU saturation.

Run with cloud/setup.sh's environment sourced. Mutants are Go overlays; repository
sources are never replaced. Test output goes directly to the named log files.
"""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parents[3]
SITES = [
    ('native-limited', 'internal/native/regexp_test.go', './internal/native', 'TestRegExpBytecodePatternUnits',
     'adamic_start(argc,argv);adamic_regex_set_step_limit(argc>1?0:10000000);',
     'if(argc==1) for(;;) {}\n adamic_start(argc,argv);adamic_regex_set_step_limit(argc>1?0:10000000);'),
    ('native-unlimited', 'internal/native/regexp_test.go', './internal/native', 'TestRegExpBytecodePatternUnits',
     'adamic_start(argc,argv);adamic_regex_set_step_limit(argc>1?0:10000000);',
     'if(argc>1) for(;;) {}\n adamic_start(argc,argv);adamic_regex_set_step_limit(argc>1?0:10000000);'),
    ('native-node', 'internal/native/regexp_test.go', './internal/native', 'TestRegExpBytecodeRandomNode',
     'command := exec.Command("node", "-e", `\n', 'command := exec.Command("node", "-e", `\nfor(;;) {}\n'),
    ('matcher-node', 'internal/regexp/matcher_oracle_test.go', './internal/regexp', 'TestMatcherNodeControls',
     'command := exec.Command("node", "-e", `\n', 'command := exec.Command("node", "-e", `\nfor(;;) {}\n'),
    ('unicode-scanner', 'internal/unicodeproperties/node_test.go', './internal/unicodeproperties', 'TestNodeAgrees',
     'const nodeScanner = `\n', 'const nodeScanner = `\nfor(;;) {}\n'),
    ('legacy-values', 'internal/unicodeproperties/canonicalize_test.go', './internal/unicodeproperties', 'TestCanonicalizeLegacyNode',
     'const legacyValueScript = `\n', 'const legacyValueScript = `\nfor(;;) {}\n'),
    ('legacy-scan', 'internal/unicodeproperties/canonicalize_test.go', './internal/unicodeproperties', 'TestCanonicalizeLegacyNode',
     'const legacyScanScript = `\n', 'const legacyScanScript = `\nfor(;;) {}\n'),
    ('unicode-scan', 'internal/unicodeproperties/canonicalize_test.go', './internal/unicodeproperties', 'TestCanonicalizeUnicodeNode',
     'const unicodeScanScript = `\n', 'const unicodeScanScript = `\nfor(;;) {}\n'),
]


def run(command, log, timeout):
    with log.open('w') as output:
        process = subprocess.Popen(command, cwd=ROOT, stdout=output, stderr=subprocess.STDOUT,
                                   start_new_session=True)
        try:
            return process.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            # A proof's outer timeout is a failed proof, never a successful kill.
            os.killpg(process.pid, signal.SIGKILL)
            process.wait()
            raise RuntimeError('proof supervisor timeout: ' + str(log))


def kills(logs):
    for name, filename, package, test, before, after in SITES:
        original = ROOT / filename
        base = original.read_text()
        if base.count(before) != 1 or 'childguard.Options{}' not in base:
            raise RuntimeError('site changed: ' + name)
        for first_output in (True, False):
            proof_name = name + ('-silent' if first_output else '-started')
            mutation = after
            if not first_output:
                if name in ('native-limited', 'native-unlimited'):
                    condition = 'argc==1' if name == 'native-limited' else 'argc>1'
                    mutation = mutation.replace('if(' + condition + ') for(;;) {}',
                        'if(' + condition + ') {puts("guard progress");fflush(stdout);for(;;) {}}')
                else:
                    mutation = mutation.replace('for(;;) {}', "console.error('guard progress');for(;;) {}")
            source = base.replace(before, mutation)
            # Only the proof changes windows. Production uses shared defaults.
            source = source.replace('childguard.Options{}',
                'childguard.Options{FirstOutput: 2_000_000_000, Stall: 2_000_000_000, Ceiling: 30_000_000_000}')
            mutant = logs / (proof_name + '.go')
            mutant.write_text(source)
            overlay = logs / (proof_name + '.json')
            overlay.write_text(json.dumps({'Replace': {str(original): str(mutant)}}))
            log = logs / (proof_name + '.log')
            status = run(['go', 'test', '-overlay=' + str(overlay), package, '-run=^' + test + '$',
                          '-count=1', '-v', '-timeout=3m'], log, 210)
            output = log.read_text()
            marker = 'stalled: no first output for 2s' if first_output else 'stalled: no output for 2s'
            if status != 1 or marker not in output or '--- FAIL: ' + test not in output:
                raise RuntimeError('hang was not caught by childguard: ' + str(log))
            print(proof_name + ': hang caught by ' + marker + ' (' + str(log) + ')', flush=True)


def loaded(logs):
    cpus = int(subprocess.check_output(['nproc'], text=True))
    workers = 2 * cpus
    # Each spinner is bounded even if its supervisor is interrupted.
    spinner = 'import time\nend=time.monotonic()+10800\nwhile time.monotonic()<end: pass\n'
    children = []
    try:
        for _ in range(workers):
            children.append(subprocess.Popen([sys.executable, '-c', spinner], stdout=subprocess.DEVNULL,
                                             stderr=subprocess.DEVNULL))
        started = time.monotonic()
        metadata = {'nproc': cpus, 'workers': workers, 'load_before': Path('/proc/loadavg').read_text().strip()
                    if Path('/proc/loadavg').exists() else 'unavailable', 'cpu.max': Path('/sys/fs/cgroup/cpu.max').read_text().strip()
                    if Path('/sys/fs/cgroup/cpu.max').exists() else 'unavailable'}
        (logs / 'load.json').write_text(json.dumps(metadata, indent=2) + '\n')
        print('loaded proof: ' + str(workers) + ' CPU spinners for nproc=' + str(cpus), flush=True)
        loaded_kills = logs / 'loaded-kills'
        loaded_kills.mkdir(exist_ok=True)
        kills(loaded_kills)
        # The outer test timer accommodates deliberate saturation. Every child
        # retains childguard's unchanged first-output, stall and ceiling windows.
        pattern = ('^(TestRegExpBytecode.*|TestRegExpNativeStepLimit|TestMatcher.*|'
                   'TestNodeAgrees|TestNodeStringProperties|'
                   'TestCanonicalizeLegacyNode|TestCanonicalizeUnicodeNode|TestProgress|'
                   'TestStalled|TestCeiling|TestExitIsNotGuardError|TestKillsProcessGroup|TestNoFirstOutput)$')
        log = logs / 'loaded.log'
        status = run(['go', 'test', './internal/childguard', './internal/native', './internal/regexp',
                      './internal/unicodeproperties', '-run=' + pattern, '-count=1', '-v', '-timeout=2h'], log, 7500)
        if status:
            raise RuntimeError('loaded comparisons failed: ' + str(log))
        if any(child.poll() is not None for child in children):
            raise RuntimeError('a load worker ended before the comparison completed')
        metadata['elapsed_seconds'] = time.monotonic() - started
        metadata['load_after'] = Path('/proc/loadavg').read_text().strip() if Path('/proc/loadavg').exists() else 'unavailable'
        if Path('/proc').exists():
            ticks = os.sysconf('SC_CLK_TCK')
            metadata['worker_cpu_seconds'] = []
            for child in children:
                fields = Path('/proc/' + str(child.pid) + '/stat').read_text().rsplit(') ', 1)[1].split()
                metadata['worker_cpu_seconds'].append((int(fields[11]) + int(fields[12])) / ticks)
            if any(seconds <= 0 for seconds in metadata['worker_cpu_seconds']):
                raise RuntimeError('a load worker consumed no CPU')
        (logs / 'load.json').write_text(json.dumps(metadata, indent=2) + '\n')
        print('loaded comparisons passed: ' + str(log), flush=True)
    finally:
        for child in children:
            if child.poll() is None:
                child.terminate()
        for child in children:
            child.wait()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--phase', choices=('kills', 'loaded', 'all'), default='all')
    parser.add_argument('--logs', type=Path, default=Path('/tmp/library-childguard-proof'))
    arguments = parser.parse_args()
    logs = arguments.logs.resolve()
    logs.mkdir(parents=True, exist_ok=True)
    status = run(['go', 'test', './internal/native', './internal/regexp', './internal/unicodeproperties',
                  '-run=^$', '-count=1', '-timeout=30m'], logs / 'compile.log', 1200)
    if status:
        raise RuntimeError('proof compilation failed: ' + str(logs / 'compile.log'))
    if arguments.phase in ('kills', 'all'):
        kills(logs)
    if arguments.phase in ('loaded', 'all'):
        loaded(logs)


if __name__ == '__main__':
    signal.signal(signal.SIGTERM, lambda signum, frame: sys.exit(128 + signum))
    main()
