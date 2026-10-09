"""Generate condition-only reductions retaining all 78 original operands and forms."""
import json
import pathlib
import re

root = pathlib.Path(__file__).resolve().parent
rows = json.loads((root / 'ledger-types.json').read_text())['sites']
assert len(rows) == len({row['where'] for row in rows}) == 78
functions, calls = [], []
for index, row in enumerate(rows):
    expression = row['expression']
    families = set(row['families'])
    present = families - {'undefined', 'null'}
    setup = []
    name = 'input'
    if expression == 'data?.diagnostics?.length':
        name = 'data'
        kind = '{ readonly diagnostics: readonly number[] | undefined } | undefined'
        values = ['undefined', '{ diagnostics: undefined }', '{ diagnostics: emptyNumbers }', '{ diagnostics: [1] }']
    elif expression == 'decl.importClause?.isTypeOnly':
        kind = '{ readonly isTypeOnly: boolean | undefined } | undefined'
        values = ['undefined', '{ isTypeOnly: undefined }', '{ isTypeOnly: false }', '{ isTypeOnly: true }']
        setup = ['const decl = { importClause: input };']
    elif expression == 'text.match(sourceMapCommentRegExpDontCareLineStart)':
        kind, values, name = 'string', ["''", "'a'"], 'text'
        setup = ['const sourceMapCommentRegExpDontCareLineStart = /a/;']
    elif 'null' in families:
        kind = row['type']
        method = 'exec' if 'Exec' in kind else 'match'
        values = ["/a/.exec('')", "/a/.exec('a')"] if method == 'exec' else ["''.match(/a/)", "'a'.match(/a/)"]
    elif present == {'string', 'object'}:
        kind, values = 'string | { readonly value: number } | undefined', ['undefined', "''", "'0'", '{ value: 0 }']
    elif present == {'string', 'array'}:
        kind, values = 'string | readonly string[] | undefined', ['undefined', "''", "'0'", 'emptyStrings', "['a']"]
    elif present == {'boolean', 'number'}:
        kind, values = 'boolean | number | undefined', ['undefined', 'false', 'true', '0', '-0', 'NaN', '1']
    else:
        kind, values = {
            'string': ('string', ["''", "'a'", "'0'"]),
            'number': ('number', ['0', '-0', 'NaN', '1', '-1']),
            'boolean': ('boolean', ['false', 'true']),
            'object': ('{ readonly value: number }', ['{ value: 0 }']),
            'array': ('readonly number[]', ['emptyNumbers', '[0]']),
            'function': ('(() => number)', ['() => 0']),
        }[next(iter(present))]
        if 'undefined' in families:
            kind += ' | undefined'
            values = ['undefined'] + values
        if row['type'] == 'true | undefined':
            kind, values = 'true | undefined', ['undefined', 'true']
    if not setup and name == 'input':
        if ' & ' in expression:
            left, right = expression.split(' & ')
            mask = row['enum_constant']
            assert isinstance(mask, int)
            values = ['0', str(mask), '-0', 'NaN']
            owner, member = right.split('.')
            setup.append(f'const {owner} = {{ {member}: {mask} }};')
            if left == 'getCombinedNodeFlags(node.declarationList)':
                setup += ['const node = { declarationList: input };', 'const getCombinedNodeFlags = (value: number): number => value;']
            elif '.' in left:
                owner, member = left.split('.')
                setup.append(f'const {owner} = {{ {member}: input }};')
            else:
                setup.append(f'const {left} = input;')
        elif expression == 'value.length':
            kind, values, name = 'readonly string[]', ['emptyStrings', "['a']"], 'value'
        elif expression == 'canReportDiagnostics(system, compilerOptions)':
            setup += ['const system = 0;', 'const compilerOptions = { value: input };']
        elif re.fullmatch(r'[A-Za-z_$][\w$]*', expression):
            name = expression
        else:
            parts = expression.split('.')
            assert all(re.fullmatch(r'[A-Za-z_$][\w$]*', part) for part in parts), expression
            value = 'input'
            for part in reversed(parts[1:]):
                value = '{ ' + part + ': ' + value + ' }'
            setup.append(f'const {parts[0]} = {value};')
    if row['context'] == 'IfStatement':
        body = f'if ({expression}) {{ return true; }}\n    return false;'
    elif row['context'] == 'WhileStatement':
        body = f'while ({expression}) {{ return true; }}\n    return false;'
    elif row['context'] == 'ConditionalExpression':
        body = f'return {expression} ? true : false;'
    else:
        raise AssertionError(row['context'])
    function = f'conditionSite{index:02d}'
    prefix = '// ' + row['where'] + ' (' + row['type'] + ')\n'
    declarations = ''.join('    ' + statement + '\n' for statement in setup)
    functions.append(prefix + f'function {function}({name}: {kind}): boolean {{\n' + declarations + '    ' + body + '\n}\n')
    for sample, value in enumerate(values):
        calls.append(f"console.log(`{row['where']} sample {sample}: ${{{function}({value})}}`);\n")
    row['reduction_function'] = function
    row['reduction_type'] = kind
    row['samples'] = values
    row['condition_preserved'] = True
output = root.parents[2] / 'internal/oracle/testdata/conditions_ledger78.a'
output.write_text('// Condition-only reductions; full upstream function bodies are not reproduced.\n' + 'const emptyNumbers: readonly number[] = [];\nconst emptyStrings: readonly string[] = [];\nfunction canReportDiagnostics(_system: number, options: { readonly value: boolean | undefined }): boolean | undefined { return options.value; }\n' + '\n'.join(functions) + '\n' + ''.join(calls))
(root / 'ledger-coverage.json').write_text(json.dumps({'sites': rows}, indent=2) + '\n')
print(f'{len(rows)} original conditions, {len(calls)} samples, {output}')
