#!/usr/bin/env python3
"""Read the submodule corpus in place. This is evidence discovery, not a verifier."""
import hashlib
import json
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CORPUS = ROOT / 'cohere/TypeScript/tsc/testdata/fixtures/compiler'


def function(text, name):
    match = re.search(r'\bfunction ' + re.escape(name) + r'(?:<[^\n]*?>)?\(', text)
    if not match:
        raise ValueError(name)
    begin = text.index('{', match.start())
    depth = 1
    end = begin + 1
    # The selected functions contain no brace-bearing strings or templates.
    while depth:
        depth += (text[end] == '{') - (text[end] == '}')
        end += 1
    return text[begin:end], text.count('\n', 0, match.start()) + 1


def site(path, line):
    return {'path': str(path.relative_to(CORPUS)), 'line': line}


def main():
    utilities = CORPUS / 'utilities.ts'
    factory = CORPUS / 'factory/nodeFactory.ts'
    types = CORPUS / 'types.ts'
    constructors = {}
    wrappers = {}
    for name in ('Node', 'Token', 'Identifier'):
        body, line = function(utilities.read_text(), name)
        constructors[name] = {**site(utilities, line), 'assigned_fields': sorted(set(re.findall(r'\bthis\.(\w+)\s*=', body)))}
    for name in ('createBaseNode', 'createBaseIdentifier', 'createIdentifier', 'createNumericLiteral', 'cloneNode'):
        body, line = function(factory.read_text(), name)
        wrappers[name] = {**site(factory, line), 'assigned_fields': sorted(set(re.findall(r'\bnode\.(\w+)\s*=', body))), 'contains_unchecked_initialization': 'undefined!' in body, 'copies_dynamic_fields': 'clone[key] = node[key]' in body}
    calls = []
    for path in sorted(CORPUS.rglob('*.ts')):
        text = path.read_text()
        for match in re.finditer(r'\bcreateBaseNode<([^<>\n]+)>\(SyntaxKind\.(\w+)\)', text):
            calls.append({**site(path, text.count('\n', 0, match.start()) + 1), 'interface_argument': match[1], 'kind': match[2]})
    tag_declarations = []
    text = types.read_text()
    for match in re.finditer(r'export interface (\w+)[^{]*\{([^}]*?)\n\}', text):
        kind = re.search(r'\breadonly kind: SyntaxKind\.(\w+);', match[2])
        if kind:
            tag_declarations.append({**site(types, text.count('\n', 0, match.start()) + 1), 'interface': match[1], 'kind': kind[1]})
    missing = 'escapedText' not in constructors['Identifier']['assigned_fields']
    completed = 'escapedText' in wrappers['createBaseIdentifier']['assigned_fields']
    assert missing and completed, 'selected staged-construction witness changed; reread source'
    version = lambda path: subprocess.check_output(['git', '-C', str(path), 'rev-parse', 'HEAD'], text=True).strip()
    evidence = {
        'corpus': str(CORPUS.relative_to(ROOT)),
        'cohere_commit': version(ROOT / 'cohere'),
        'typescript_go_commit': version(ROOT / 'cohere/TypeScript'),
        'source_sha256': {str(path.relative_to(CORPUS)): hashlib.sha256(path.read_bytes()).hexdigest() for path in (utilities, factory, types, CORPUS / 'factory/baseNodeFactory.ts', CORPUS / 'parser.ts')},
        'cast_counts': 'Use stage3/fixtures/assertions/ledger-summary.json at 5b173f3920ab2c5b7058f0a9fe4b8e4a91f52523; not remeasured',
        'predicate_obligations': 'Use stage3/fixtures/predicates/ledger.json at e42eaf9854563617734ff13987cc27e4e09f78b1; descriptions are not proof verdicts',
        'constructors': constructors,
        'wrappers': wrappers,
        'direct_generic_base_allocation_sites': calls,
        'direct_interface_kind_declarations': tag_declarations,
        'observed_identifier_constructor_lacks_escapedText': missing,
        'observed_identifier_wrapper_sets_escapedText': completed,
        'limits': 'Lexical discovery for explicitly named direct generic allocations and kind declarations only. No full initialization, escape, inheritance, constructor replacement or publication proof. Interfaces without their own kind declaration are not counted. No assertion annotations are trusted.'
    }
    print(json.dumps(evidence, indent=2))


if __name__ == '__main__':
    main()
