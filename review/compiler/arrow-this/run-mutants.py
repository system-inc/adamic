import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
evidence = pathlib.Path(__file__).resolve().parent
mutants = [
    ('constructor-context', 'internal/lower/class_inheritance.go',
     '\touterClosures := l.closures\n\tl.closures = nil\n\tdefer func() { l.closures = outerClosures }()\n',
     '', 'TestArrowThisField'),
    ('dropped-capture', 'internal/lower/locals.go',
     '\t\t\tfunction.Environment = append(function.Environment, local)\n',
     '\t\t\t// Mutant: omit the captured cell from the environment.\n', 'TestArrowThisSuper'),
]
for name, relative, before, after, test in mutants:
    path = root / relative
    original = path.read_text()
    assert original.count(before) == 1, name
    mutated = original.replace(before, after)
    import difflib
    (evidence / (name + '.patch')).write_text(''.join(difflib.unified_diff(
        original.splitlines(True), mutated.splitlines(True),
        fromfile='a/' + relative, tofile='b/' + relative)))
    try:
        path.write_text(mutated)
        with (evidence / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^' + test + '$',
                                     '-count=1', '-timeout', '90s'], cwd=root,
                                    stdout=log, stderr=subprocess.STDOUT, timeout=110)
        output = (evidence / (name + '.log')).read_text()
        assert result.returncode != 0 and 'use of undeclared identifier' in output, output
        print(name + ': caught by clang undeclared identifier, exit=' + str(result.returncode), flush=True)
    finally:
        path.write_text(original)
