"""Skip one UTF-16 candidate in the branch code and restore it in all cases."""
import os
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[2]
target = root / 'internal/native/runtime/regexp.c'
original = target.read_text()
site = '\t\tstart = (size_t)at + 1;'
assert original.count(site) == 1
log = pathlib.Path('/tmp/regex-coverage/mutant.log')
command = ['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/regex_coverage_methods', '-count=1', '-timeout', '15m', '-v']
try:
    target.write_text(original.replace(site, '\t\tstart = (size_t)at + 2;'))
    with log.open('w') as output:
        result = subprocess.run(command, cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode != 0 and 'stdout differs' in log.read_text(), log.read_text()
    assert 'error:' not in log.read_text(), 'Mutant must fail at runtime, not clang'
    print('One-line UTF-16 search mutant caught by stdout; exit', result.returncode)
finally:
    target.write_text(original)
with pathlib.Path('/tmp/regex-coverage/restored.log').open('w') as output:
    result = subprocess.run(command, cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
assert result.returncode == 0
print('Restored oracle passed')
