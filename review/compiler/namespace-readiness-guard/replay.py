from pathlib import Path
import subprocess
import json
root = Path(__file__).resolve().parents[3]
out = root / 'review/compiler/namespace-readiness-guard'
patch = out / 'M14.diff'
subprocess.run(['git', 'apply', '--check', str(patch)], cwd=root, check=True, timeout=10)
subprocess.run(['git', 'apply', str(patch)], cwd=root, check=True, timeout=10)
try:
    command = ['go', 'test', './internal/lower', '-run', '^TestNamespaceForwarderReadinessBeforeInitialization$', '-count=1', '-timeout', '90s', '-v']
    with (out / 'M14.log').open('w') as log:
        result = subprocess.run(command, cwd=root, stdout=log, stderr=log, timeout=110)
    (out / 'result.json').write_text(json.dumps(dict(command=command, exit=result.returncode), indent=2)+'\n')
    text = (out / 'M14.log').read_text()
    print(text)
    assert result.returncode == 1 and 'JavaScript backend stdout' in text, 'M14 must fail by observable stdout disagreement'
finally:
    subprocess.run(['git', 'apply', '-R', str(patch)], cwd=root, check=True, timeout=10)
