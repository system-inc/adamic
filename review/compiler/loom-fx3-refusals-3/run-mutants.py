from pathlib import Path
import subprocess
import difflib

root = Path.cwd()
evidence = root / 'review/compiler/loom-fx3-refusals-3'
source = root / 'internal/lower/object.go'
registry = root / 'internal/lower/agree_test.go'
for name, path, patch, test, expected in [
    ('every', source, 'review/compiler/fx5-every-refusal/revert-every.diff', 'TestArrayNarrowingEveryRefused', 'got <nil>'),
    ('findLast', source, 'review/compiler/fx5-every-refusal/revert-findLast.diff', 'TestArrayNarrowingFindLastRefused', 'got <nil>'),
    ('literal', source, 'review/compiler/fx5-literal-refusal/revert-refusal.diff', 'TestFX5NumberLiteralRefused', 'got <nil>'),
    ('registration', registry, None, 'TestNativeAgreementRowsNameRegisteredPrograms', 'must name a program in nativeAgreementPrograms by a literal'),
]:
    green = path.read_text()
    try:
        if patch:
            subprocess.run(['git', 'apply', patch], check=True, timeout=10)
        else:
            path.write_text(''.join(line for line in green.splitlines(True) if not line.lstrip().startswith('"FX5MixedLiteral":')))
        (evidence / (name + '.patch')).write_text(''.join(difflib.unified_diff(green.splitlines(True), path.read_text().splitlines(True), fromfile=str(path.relative_to(root)), tofile=str(path.relative_to(root)))))
        command = ['go', 'test', './internal/lower', '-run', '^' + test + '$', '-count=1', '-v', '-timeout', '90s']
        with (evidence / ('mutant-' + name + '.log')).open('w') as log:
            result = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=150)
        output = (evidence / ('mutant-' + name + '.log')).read_text()
        if result.returncode == 0 or expected not in output:
            raise RuntimeError(name + ' did not fail its intended assertion: ' + output)
        print(name + ': caught by ' + test, flush=True)
    finally:
        path.write_text(green)
