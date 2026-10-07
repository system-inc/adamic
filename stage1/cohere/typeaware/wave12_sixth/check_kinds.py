"""Validate owned named-kind declarations against pinned Go ast.Kind.String()."""
import ast
import json
from pathlib import Path
import re
import subprocess
import sys

repository = Path(__file__).resolve().parents[4]
owned = Path(__file__).resolve().parent
legacy = {
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
active = {
    'jsx_fragments': ['JsxFragment', 'JsxElement', 'JsxSelfClosingElement'],
    'jsx_no_constructed_context_values': ['JsxOpeningElement', 'JsxSelfClosingElement'],
    'jsx_no_undef': ['JsxOpeningElement', 'JsxSelfClosingElement'],
}

def valid(actual, expected, allowed):
    return (isinstance(actual, list) and all(isinstance(kind, str) for kind in actual)
            and len(set(actual)) == len(actual) and all(kind in allowed for kind in actual)
            and actual == expected)

def validate_kinds(scratch, run):
    truth, stderr = run('ast-kind-names', ['go', 'run', owned / 'kind_names.go'], repository / 'cohere')
    assert not stderr
    allowed = set(json.loads(truth))
    assert 'JsxFragment' in allowed and 'CallExpression' in allowed
    declarations = {}
    for relative, expected in legacy.items():
        text = (owned.parent / relative).read_text()
        match = re.search(r'export const listenerKinds: readonly string\[\] = (\[[^\]]*\]);', text)
        assert match is not None, relative
        declarations[relative] = (ast.literal_eval(match.group(1)), expected)
    node_mutants = []
    for folder, expected in active.items():
        descriptor = json.loads((owned / folder / 'rule.json').read_text())
        assert descriptor.get('node') is True, folder
        changed = dict(descriptor); changed['node'] = False
        assert changed.get('node') is not True, folder
        node_mutants.append({'declaration': folder, 'mutation': 'refetch-instead-of-supplied-node'})
        declarations[folder] = (descriptor['kinds'], expected)
    mutations = node_mutants
    for label, (actual, expected) in declarations.items():
        assert valid(actual, expected, allowed), label
        for name, replacement in [('unknown-name', 'NoSuchAstKind'), ('numeric-kind', 80),
                                  ('wrong-listener', 'Identifier' if actual[0] != 'Identifier' else 'SourceFile')]:
            changed = list(actual); changed[0] = replacement
            assert not valid(changed, expected, allowed), (label, name)
            mutations.append({'declaration': label, 'mutation': name})
        print(label, actual, 'PASS; unknown-name, numeric-kind and wrong-listener mutants caught')
    (scratch / 'kind-summary.json').write_text(json.dumps({'declarations': len(declarations), 'mutants': mutations, 'ast_names': len(allowed)}, indent=2) + '\n')
    return mutations

if __name__ == '__main__':
    scratch = Path(sys.argv[1]); scratch.mkdir(parents=True, exist_ok=True)
    def run(name, args, directory):
        with (scratch / (name + '.stdout')).open('wb') as stdout, (scratch / (name + '.stderr')).open('wb') as stderr:
            result = subprocess.run([str(arg) for arg in args], cwd=directory, stdout=stdout, stderr=stderr)
        assert result.returncode == 0, name
        return (scratch / (name + '.stdout')).read_bytes(), (scratch / (name + '.stderr')).read_bytes()
    validate_kinds(scratch, run)
    print('PASS: 15 named declarations; 48 metadata mutants. No numeric source-adapter requirement.')
