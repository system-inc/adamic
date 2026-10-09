import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[3]
review = Path(__file__).resolve().parent
path = root / 'internal/lower/regexp.go'
original = path.read_bytes()
selection = '^TestRegExp(UnicodeClass|SlashClass|EscapedSlash)SourceAgreesWithNode$'

def test(name):
    with (review / (name + '.log')).open('w') as log:
        result = subprocess.run(['go', 'test', './internal/lower', '-run', selection, '-count=1', '-timeout', '90s', '-v'], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=120)
    print(name, 'exit', result.returncode, flush=True)
    assert result.returncode == 1
    text = (review / (name + '.log')).read_text()
    assert 'native regex stdout' in text, 'mutant must fail behavioral comparison, not compilation'
    return text

try:
    subprocess.run(['git', 'apply', str(review / 'regex-M20.patch')], cwd=root, check=True)
    text = test('regex-M20')
    assert '--- FAIL: TestRegExpUnicodeClassSourceAgreesWithNode' in text
finally:
    path.write_bytes(original)

try:
    old = b"if c == '/' && !escaped && depth == 0 {"
    new = b"if c == '/' && depth <= 1 {"
    assert original.count(old) == 1
    path.write_bytes(original.replace(old, new))
    patch = subprocess.run(['git', 'diff', '--', 'internal/lower/regexp.go'], cwd=root, capture_output=True, check=True).stdout
    (review / 'regex-overescape.patch').write_bytes(patch)
    text = test('regex-overescape')
    for name in ('UnicodeClass', 'SlashClass', 'EscapedSlash'):
        assert '--- FAIL: TestRegExp' + name + 'SourceAgreesWithNode' in text
finally:
    path.write_bytes(original)
