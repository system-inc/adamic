from pathlib import Path
import subprocess
source = Path('internal/lower/object.go')
original = source.read_text()
guard = '\tif err := l.eepPropertyPresence(node); err != nil {\n\t\treturn nil, err\n\t}\n'
assert original.count(guard) == 1
try:
    source.write_text(original.replace(guard, ''))
    with Path('review/compiler/eep-presence/revert-mutant.log').open('w') as log:
        result = subprocess.run(['bash', '-lc', 'source /workspace/adamic-tools/env.sh; timeout 90 go test -buildvcs=false ./internal/lower -run TestEEPPresence -v -count=1 -timeout 90s'], stdout=log, stderr=subprocess.STDOUT, timeout=100)
    output = Path('review/compiler/eep-presence/revert-mutant.log').read_text()
    assert result.returncode == 1
    assert all('--- FAIL: TestEEPPresence' + str(i) in output for i in range(7))
    assert '--- PASS: TestEEPPresenceSupported' in output
    print('Revert mutant caught by all seven refusal tests; positive control passed')
finally:
    source.write_text(original)
