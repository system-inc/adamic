"""Run one-line lowering mutants against original and converted acceptance rows."""
from pathlib import Path
import difflib
import subprocess

root = Path(__file__).resolve().parents[3]
evidence = root / 'review/compiler/agree-classes'
mutants = [
    ('static-call', 'internal/lower/class_static.go',
     'statements = append(statements, ir.Evaluate{Value: ir.Call{Function: index, Arguments: []ir.Expression{object}}})',
     'statements = append(statements, ir.Evaluate{Value: object})',
     'internal/lower/class_features_test.go', '^TestClassFeaturesStaticDeclarationsExecute$'),
    ('parameter-store', 'internal/lower/parameter_properties.go',
     'Name: parameter.Name().Text(), Value: fit(',
     'Name: parameter.Name().Text() + "_mutant", Value: fit(',
     'internal/lower/parameter_properties_test.go', '^TestParameterPropertyCheckerContracts$'),
]
for name, lowering, before, after, test_file, test in mutants:
    path = root / lowering
    original = path.read_text()
    assert original.count(before) == 1, name
    changed = original.replace(before, after)
    patch = ''.join(difflib.unified_diff(original.splitlines(True), changed.splitlines(True),
                                       fromfile='a/' + lowering, tofile='b/' + lowering))
    (evidence / (name + '.diff')).write_text(patch)
    test_patch = subprocess.check_output(['git', 'diff', '--', test_file], cwd=root)
    restored = True
    try:
        path.write_text(changed)
        subprocess.run(['git', 'apply', '-R'], input=test_patch, cwd=root, check=True)
        restored = False
        with (evidence / (name + '-old.log')).open('w') as log:
            old = subprocess.run(['go', 'test', './internal/lower', '-run', test, '-count=1', '-timeout', '90s'],
                                 cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=110)
        subprocess.run(['git', 'apply'], input=test_patch, cwd=root, check=True)
        restored = True
        with (evidence / (name + '-converted.log')).open('w') as log:
            new = subprocess.run(['go', 'test', './internal/lower', '-run', test, '-count=1', '-timeout', '90s'],
                                 cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=110)
        result = (evidence / (name + '-converted.log')).read_text()
        assert old.returncode == 0, name + ': original test failed'
        assert new.returncode != 0 and 'JavaScript backend stdout' in result, name + ': no stdout kill'
        print(name + ': original PASS; converted FAIL on JavaScript backend stdout', flush=True)
    finally:
        path.write_text(original)
        if not restored:
            subprocess.run(['git', 'apply'], input=test_patch, cwd=root, check=True)
