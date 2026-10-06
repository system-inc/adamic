"""Use the existing oracle on the note, then remove its temporary registration."""
import os
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[2]
registration = root / 'internal/oracle/regex_coverage_difference_probe_test.go'
assert not registration.exists()
log = pathlib.Path('/tmp/regex-coverage/difference-oracle.log')
try:
    registration.write_text('package oracle\nfunc init() { fixtures = append(fixtures, struct { path string; lowers bool; checked bool }{"notes/regex/split_surrogate_assertions.a", true, false}) }\n')
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/notes/regex/split_surrogate_assertions', '-count=1', '-timeout', '15m', '-v'], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode != 0 and 'stdout differs' in log.read_text(), log.read_text()
    print('Split disagreement reproduced by the oracle; exit', result.returncode)
finally:
    registration.unlink()
