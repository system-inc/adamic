import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[3]
review = Path(__file__).resolve().parent
path = root / 'internal/lower/switch.go'
original = path.read_bytes()
selection = '^TestSwitch(ConditionalBreak|EmptyElse|EmptyCase|EmptyBlock)FallsThrough$'
try:
    subprocess.run(['git', 'apply', str(review / 'M02.patch')], cwd=root, check=True)
    with (review / 'M02.log').open('w') as log:
        result = subprocess.run(['go', 'test', './internal/lower', '-run', selection, '-count=1', '-timeout', '90s', '-v'], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=120)
    print('M02 exit', result.returncode, flush=True)
    text = (review / 'M02.log').read_text()
    assert result.returncode == 1
    assert 'JavaScript backend stdout' in text, 'require behavioral mismatch rather than a build failure'
    for name in ('ConditionalBreak', 'EmptyElse', 'EmptyCase', 'EmptyBlock'):
        assert '--- FAIL: TestSwitch' + name + 'FallsThrough' in text
finally:
    path.write_bytes(original)
