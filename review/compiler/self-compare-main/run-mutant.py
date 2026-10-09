from pathlib import Path
import subprocess

source = Path('internal/native/units.go')
original = source.read_text()
condition = 'if d.function && d.tokens[0].text == "static" && d.tokens[1].text == "inline" {'
assert original.count(condition) == 1
mutated = original.replace(condition, 'if false {', 1)
import difflib
Path('review/compiler/self-compare-main/static-inline-mutant.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile='a/internal/native/units.go', tofile='b/internal/native/units.go')))
try:
    source.write_text(mutated)
    with open('review/compiler/self-compare-main/mutant.log', 'w') as log:
        result = subprocess.run(['timeout', '60', 'go', 'test', './internal/native', '-run', '^TestSplitSelfCompareAgreesWithNode$', '-count=1', '-v', '-timeout', '90s'], stdout=log, stderr=subprocess.STDOUT)
    output = Path('review/compiler/self-compare-main/mutant.log').read_text()
    assert result.returncode == 1, (result.returncode, output)
    assert "inline function 'adamic_unit_adamic_same_number' is not defined" in output, output
    assert 'note: used here' in output, output
    print('static-inline mutant caught by TestSplitSelfCompareAgreesWithNode: -Werror,-Wundefined-inline; exit 1')
finally:
    source.write_text(original)
