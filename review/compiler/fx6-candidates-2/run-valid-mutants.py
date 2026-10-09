from pathlib import Path
import difflib
import subprocess

root = Path.cwd()
evidence = root / 'review/compiler/fx6-candidates-2'
def mutant(name, changes, test):
    originals = {path: (root/path).read_text() for path in changes}
    patch = ''
    try:
        for path, mutate in changes.items():
            before = originals[path]
            after = mutate(before)
            assert after != before
            patch += ''.join(difflib.unified_diff(before.splitlines(True), after.splitlines(True), fromfile='a/'+path, tofile='b/'+path))
            (root/path).write_text(after)
        (evidence/(name+'.diff')).write_text(patch)
        with (evidence/(name+'.log')).open('w') as log:
            result = subprocess.run(['go','test','./internal/lower','-run',test,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=180)
        output = (evidence/(name+'.log')).read_text()
        print(name, 'exit', result.returncode, flush=True)
        print(output, flush=True)
        assert result.returncode != 0 and '[build failed]' not in output
        if name == 'revert-p70':
            assert 'got <nil>' in output
        else:
            assert 'JavaScript backend stdout' in output
    finally:
        for path, before in originals.items():
            (root/path).write_text(before)

mutant('revert-receiver', {
    'internal/lower/interface_cast.go': lambda s:s.replace('for typeID, contractID := range l.result.ViewContractTypes {','for _, contractID := range l.result.ViewContractTypes {').replace('l.result.CheckedFields[checkedViewFieldKey(typeID, field.Name)] = true','l.result.CheckedFields[field.Name] = true').replace('l.result.CheckedFields[checkedViewFieldKey(int(target.Id()), field)] = true','l.result.CheckedFields[field] = true'),
    'internal/lower/readiness.go': lambda s:s.replace('!(program.CheckedFields[checkedViewFieldKey(expression.ViewReceiverTypeID, expression.Name)] || program.CheckedFields[expression.Name])','!program.CheckedFields[expression.Name]'),
    'internal/lower/view_lazy.go':lambda s:s.replace('!(program.CheckedFields[checkedViewFieldKey(read.ViewReceiverTypeID, read.Name)] || program.CheckedFields[read.Name])','!program.CheckedFields[read.Name]'),
},'^TestFX6P37$')
mutant('revert-destructure-type-id', {'internal/lower/view_member_read.go':lambda s:s.replace('\t\tproperty.ViewTypeID = int(declared.Id())\n','')}, '^TestFX6P53$')
mutant('revert-p70', {'internal/lower/view_lazy.go':lambda s:s.replace('\t\t\tif err := tupleObjectViewRead(program, graph, index, read); err != nil {\n\t\t\t\trefused = err\n\t\t\t\treturn false\n\t\t\t}\n','')}, '^TestTupleObjectViewRefused$')
