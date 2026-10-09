from pathlib import Path
import json
import subprocess
import time

root = Path(__file__).resolve().parents[3]
evidence = root / 'review/compiler/agree-enum-runner'
cases = [
    ('M01', '^TestNumericEnumsAreOpen$'),
    ('M02', '^TestNumericEnumsAreOpen$'),
    ('M04', '^TestEnumReverseMappingUsesSingleSlot$'),
    ('M14', '^TestEnumFlagProofsWithoutObservableLoweringEffect$'),
    ('M15', '^TestEnumFlagProofsWithoutObservableLoweringEffect$'),
    ('M16', '^TestEnumFlagProofsWithoutObservableLoweringEffect$'),
]
results = []
for label, selector in cases:
    patch = root / 'review/compiler/enum-guards' / (label + '.diff')
    subprocess.run(['git', 'apply', '--check', str(patch)], cwd=root, check=True, timeout=10)
    subprocess.run(['git', 'apply', str(patch)], cwd=root, check=True, timeout=10)
    try:
        command = ['go', 'test', './internal/lower', '-run', selector, '-count=1', '-timeout', '90s', '-v']
        started = time.monotonic()
        with (evidence / (label + '.log')).open('w') as log:
            result = subprocess.run(command, cwd=root, stdout=log, stderr=log, timeout=110)
        failures = [line for line in (evidence / (label + '.log')).read_text().splitlines() if '--- FAIL:' in line]
        results.append(dict(mutant=label, command=command, exit=result.returncode, seconds=round(time.monotonic()-started, 3), failures=failures))
        print(results[-1], flush=True)
        assert result.returncode == 1 and failures, 'mutant must fail by a test assertion'
    finally:
        subprocess.run(['git', 'apply', '-R', str(patch)], cwd=root, check=True, timeout=10)
    (evidence / 'mutants.json').write_text(json.dumps(results, indent=2) + '\n')
