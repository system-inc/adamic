from pathlib import Path
import difflib
import subprocess

root = Path.cwd()
evidence = root / 'review/compiler/fx6-candidates'
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
        if name == 'revert-element-refusal':
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
mutant('revert-destructure',{'internal/lower/collections.go':lambda s:s.replace('\t\tvalue.ViewTypeID = int(fieldType.Id())\n','')},'^TestFX6P53$')
mutant('revert-element-refusal',{'internal/lower/element_access_fields.go':lambda s:s.replace('func (l *lowering) refuseViewElementReads(fields map[string]bool) error {','func (l *lowering) refuseViewElementReads(fields map[string]bool) error {\n fields = nil')},'^TestViewStringElementReadRefused$')
