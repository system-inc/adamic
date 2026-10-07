"""Check owned listener metadata against the external TypeScript enum."""
import re
import sys
from pathlib import Path

repository = Path(__file__).resolve().parents[5]
reference = Path(sys.argv[1]).read_text()
body = reference.split('export const enum SyntaxKind {', 1)[1].split('}', 1)[0]
values = {}
number = 0
for line in body.splitlines():
    match = re.fullmatch(r'(\w+)(?:\s*=\s*([^,]+))?,?', line.split('//')[0].strip())
    if match is None:
        continue
    name, expression = match.groups()
    if expression:
        if expression.isdigit():
            number = int(expression)
        elif expression in values:
            number = values[expression]
        else:
            break
    values[name] = number
    number += 1

files = {
    'no_redeclare.a': ['ClassDeclaration', 'InterfaceDeclaration', 'TypeAliasDeclaration', 'EnumDeclaration', 'ModuleDeclaration', 'FunctionDeclaration', 'VariableDeclaration', 'BindingElement'],
    'no_test_on_global_regex.a': ['CallExpression'],
    'no_write_only_collection.a': ['VariableDeclaration'],
    'wave12_next/no_process_exit_after_output.a': ['CallExpression'],
    'wave12_next/no_uncleared_race_timeout.a': ['CallExpression'],
    'wave12_next/require_blocking_standard_streams.a': ['SourceFile'],
    'wave12_third/no_new_func.a': ['CallExpression', 'NewExpression'],
    'wave12_third/no_new_native_nonconstructor.a': ['NewExpression'],
    'wave12_third/no_new_wrappers.a': ['NewExpression'],
    'wave12_fourth/prefer_regex_literals.a': ['Identifier'],
    'wave12_fourth/prefer_rest_params.a': ['Identifier'],
    'wave12_fourth/exhaustive_deps.a': ['CallExpression'],
}

def declared(text):
    match = re.search(r'export const listenerKinds: readonly number\[\] = \[([^\]]*)\];', text)
    assert match is not None, 'missing numeric listener declaration'
    return [int(value.strip()) for value in match[1].split(',')]

for relative, names in files.items():
    text = (repository / 'stage1/cohere/typeaware' / relative).read_text()
    expected = [values[name] for name in names]
    assert declared(text) == expected, relative
    original = declared(text)
    mutated = list(original)
    mutated[0] += 1
    changed = text.replace('[' + ', '.join(map(str, original)) + '];', '[' + ', '.join(map(str, mutated)) + '];', 1)
    assert declared(changed) != expected, 'listener mutant survived: ' + relative
    print(relative, expected, 'PASS; numeric +1 mutant caught')
print('PASS: 12 declarations; 12 metadata mutants. Callback migration remains blocked on shared numeric ParseNode API.')
