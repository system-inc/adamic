#!/usr/bin/env python3
"""Prove daily-run guards fail independently, without modifying tracked sources."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[2]
logs = Path(__import__('sys').argv[1]).resolve()
logs.mkdir(parents=True, exist_ok=True)
script = (root / 'stage3/meter/twice-daily.sh').read_text()

def caught(name, args, env=None, witness='FAIL'):
    with (logs / (name + '.txt')).open('w') as log:
        result = subprocess.run(args, cwd=root, env=dict(os.environ, **(env or {})),
                                stdout=log, stderr=log, timeout=300)
    text = (logs / (name + '.txt')).read_text()
    if result.returncode == 0 or witness not in text:
        raise AssertionError(name + ': mutant survived or failed for the wrong reason\n' + text)
    print(name + ': caught')

with tempfile.TemporaryDirectory() as temporary:
    scratch = Path(temporary)
    changes = [
        ('ignore-pin', 'main_commit=$(git rev-parse --verify "${main_ref}^{commit}")',
         'main_commit=$(git rev-parse --verify origin/main)', 'test_unreachable_pin_fails_before_output'),
        ('unequal-pins', '[[ $main_commit == "$area_commit" ]] || { echo "main and area alias resolve to different shas" >&2; exit 1; }',
         ':', 'test_different_pins_fail_before_output'),
        ('checkout-compiler', 'build_census "$scratch/$label-source" "$binaries"',
         'build_census "$repository" "$binaries"', 'test_per_ref_builds_and_runs_each_pinned_compiler'),
        ('omit-origin-retry', 'git -C "$scratch/$label-source" -c protocol.file.allow=always submodule update --init --recursive --depth 1 >> "$run/$label/submodules.log" 2>&1',
         ':', 'test_missing_local_dependency_transport_retries_recorded_origins'),
    ]
    for name, old, new, test in changes:
        assert script.count(old) == 1
        path = scratch / (name + '.sh')
        path.write_text(script.replace(old, new))
        caught(name, ['python3', '-m', 'unittest', 'discover', '-s', 'stage3/meter', '-p', 'compiler_test.py', '-k', test],
               {'METER_SCRIPT_UNDER_TEST': str(path)})
    report = (root / 'stage3/meter/report.py').read_text()
    assert report.count('if previous:') == 1
    (scratch / 'report.py').write_text(report.replace('if previous:', 'if False:'))
    # Import the mutant ahead of the original test directory.
    caught('omit-previous', ['python3', '-c',
           'import sys,unittest;sys.path.insert(0,' + repr(str(scratch)) + ');sys.path.insert(1,' + repr(str(root / 'stage3/meter')) + ');'
           'import report_test;unittest.main(module=report_test,defaultTest="PreviousRunTests")'])
    statecopy = (root / 'stage3/census/latent/statecopy/main.go').read_text()
    start = statecopy.index('if name == "lowering" && id.Name == "impliedTargetQuery"')
    end = statecopy.index('} else {', start)
    permissive = statecopy[:start] + 'if _, ok := field.Type.(*ast.FuncType); ok { expr = "value." + id.Name; ' + statecopy[end:]
    hook = scratch / 'any-hook.go'
    hook.write_text(permissive)
    caught('any-hook', ['python3', '-m', 'unittest', 'discover', '-s', 'stage3/meter', '-p', 'generator_test.py'],
           {'STATECOPY_UNDER_TEST': str(hook)}, witness='populated hook accepted')
    (scratch / 'report.py').write_text(report.replace("def write_readme(run, result, revision='HEAD'):", "def write_readme(run, result, revision='HEAD'):\n    revision = 'HEAD'"))
    caught('ignore-baseline-pin', ['python3', '-c',
           'import sys,unittest;sys.path.insert(0,' + repr(str(scratch)) + ');sys.path.insert(1,' + repr(str(root / 'stage3/meter')) + ');'
           'import report_test;unittest.main(module=report_test,defaultTest="PreviousRunTests")'])
    path = root / 'stage3/census/latent/refusalrewrite/rewrite.go'
    source = path.read_text()
    old = 'if init != "err := l.referenceRefusal(module)" && init != "err := l.parallelPreflight(module)" {'
    assert source.count(old) == 1
    mutant = scratch / 'rewrite.go'
    mutant.write_text(source.replace(old, 'if init == "" {'))
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(path): str(mutant)}}))
    caught('any-preflight', ['go', 'test', '-overlay=' + str(overlay), '-count=1',
                            './stage3/census/latent/refusalrewrite', '-run', '^TestNamedPreflightsAndUnknownShapes$'],
           witness='unknown preflight published output')
print('All eight daily guard mutants caught; generator_test.py also kills the permissive-hook mutant.')
