from pathlib import Path
import os
import subprocess

root = Path('/tmp/string-views-mutant')
source = root / 'internal/native/runtime/string_share.c'
original = source.read_text()
old = 'shared->owner = owner->heap.references == 0 ? NULL : adamic_retain((adamic_string *)owner);'
new = 'shared->owner = owner->heap.references == 0 ? NULL : (adamic_string *)owner;'
assert original.count(old) == 1
command = ['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/string_views_lifetime.a', '-count=1', '-v', '-timeout', '30m']
environment = {**os.environ, 'ADAMIC_GATE_UNCACHED': '1'}
try:
    source.write_text(original.replace(old, new))
    with open('/tmp/string-views-mutant.log', 'w') as log:
        result = subprocess.run(command, cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT)
    report = Path('/tmp/string-views-mutant.log').read_text()
    assert result.returncode != 0, 'mutant survived'
    assert 'heap-use-after-free' in report, report
    assert '[-Werror' not in report and 'stage 0 can' not in report, report
    print('Missing owner retain caught by AddressSanitizer heap-use-after-free; exit', result.returncode, flush=True)
finally:
    source.write_text(original)
assert source.read_text() == original
with open('/tmp/string-views-mutant-restored.log', 'w') as log:
    result = subprocess.run(command, cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT)
assert result.returncode == 0, Path('/tmp/string-views-mutant-restored.log').read_text()
print('Restored code passes the same uncached oracle; exit 0', flush=True)
