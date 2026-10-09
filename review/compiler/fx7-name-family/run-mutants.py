from pathlib import Path
import difflib
import subprocess

root = Path.cwd()
evidence = root / 'review/compiler/fx7-name-family'
def mutant(name, changes, test):
    originals = {path: (root/path).read_text() for path in changes}
    try:
        patch = ''
        for path, change in changes.items():
            before = originals[path]
            after = change(before)
            assert before != after
            patch += ''.join(difflib.unified_diff(before.splitlines(True), after.splitlines(True), fromfile='a/'+path, tofile='b/'+path))
            (root/path).write_text(after)
        (evidence/(name+'.diff')).write_text(patch)
        with (evidence/(name+'.log')).open('w') as log:
            result = subprocess.run(['go','test','./internal/lower','-run',test,'-v','-count=1','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=180)
        output = (evidence/(name+'.log')).read_text()
        print(name, result.returncode, output, flush=True)
        assert result.returncode != 0 and '[build failed]' not in output and 'backend' in output
    finally:
        for path, before in originals.items():
            (root/path).write_text(before)

helper = '\nfunc (e *emitter) fx7NameChecked(name string) bool { for key, checked := range e.program.CheckedFields { if checked && strings.HasSuffix(key, ":"+name) { return true } }; return false }\n'
def writes(source, types):
    return source.replace('if e.program.CheckedFields[statement.Name] {', 'if e.program.CheckedFields[statement.Name] || (('+types+') && e.fx7NameChecked(statement.Name)) {') + helper
mutant('147-name-wide-undefined-write', {'internal/native/emit_statements.go':lambda s:writes(s, 'statement.Value.Type() == ir.String')}, '^TestFX7P(04|06)$')
mutant('148-name-wide-read', {
 'internal/lower/interface_cast.go':lambda s:s.replace('for typeID, contractID := range l.result.ViewContractTypes {','for _, contractID := range l.result.ViewContractTypes {').replace('l.result.CheckedFields[checkedViewFieldKey(typeID, field.Name)] = true','l.result.CheckedFields[field.Name] = true').replace('l.result.CheckedFields[checkedViewFieldKey(int(target.Id()), field)] = true','l.result.CheckedFields[field] = true'),
 'internal/lower/readiness.go':lambda s:s.replace('!(program.CheckedFields[checkedViewFieldKey(expression.ViewReceiverTypeID, expression.Name)] || program.CheckedFields[expression.Name])','!program.CheckedFields[expression.Name]'),
 'internal/lower/view_lazy.go':lambda s:s.replace('!(program.CheckedFields[checkedViewFieldKey(read.ViewReceiverTypeID, read.Name)] || program.CheckedFields[read.Name])','!program.CheckedFields[read.Name]'),
}, '^TestFX7P(75|77)$')
mutant('149-name-wide-null-write', {'internal/native/emit_statements.go':lambda s:writes(s, 'statement.Value.Type() == ir.Array || statement.Value.Type() == ir.Map')}, '^TestFX7P(08|59)$')
