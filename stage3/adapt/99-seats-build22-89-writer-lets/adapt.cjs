'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const single = {
    name: 'createSingleLineStringWriter',
    names: ['str'],
    before: ['var str = "";'],
    after: ['let str = "";'],
};
const multiline = {
    name: 'createTextWriter',
    names: ['output', 'indent', 'lineStart', 'lineCount', 'linePos', 'hasTrailingComment'],
    before: [
        'var output: string;',
        'var indent: number;',
        'var lineStart: boolean;',
        'var lineCount: number;',
        'var linePos: number;',
        'var hasTrailingComment = false;',
    ],
    after: [
        'let output: string = "";',
        'let indent: number = 0;',
        'let lineStart: boolean = true;',
        'let lineCount: number = 0;',
        'let linePos: number = 0;',
        'let hasTrailingComment = false;',
    ],
};
const reset = [
    'function reset(): void {',
    '        output = "";',
    '        indent = 0;',
    '        lineStart = true;',
    '        lineCount = 0;',
    '        linePos = 0;',
    '        hasTrailingComment = false;',
    '    }',
];

function owner(source, name) {
    const matches = source.statements.filter(statement =>
        ts.isFunctionDeclaration(statement) && statement.name?.text === name);
    assert.equal(matches.length, 1, name + ': expected one top-level function');
    assert(matches[0].body, name + ': expected a body');
    return matches[0];
}

function checkBindings(source, factory, names, declarations) {
    const pinned = new Map(names.map((name, index) => [
        name, declarations[index].declarationList.declarations[0]?.name,
    ]));

    function checkName(name) {
        if (ts.isIdentifier(name)) {
            if (pinned.has(name.text) && pinned.get(name.text) !== name) {
                const { line } = source.getLineAndCharacterOfPosition(name.getStart(source));
                assert.fail(`utilities.ts:${line + 1}: conflicting binding of ${name.text}`);
            }
        } else {
            for (const element of name.elements) {
                if (ts.isBindingElement(element)) checkName(element.name);
            }
        }
    }

    const { body } = factory;
    for (const parameter of factory.parameters) checkName(parameter.name);

    function visit(node) {
        if (ts.isVariableDeclarationList(node)) {
            const direct = ts.isVariableStatement(node.parent) && node.parent.parent === body;
            if (direct || !(node.flags & ts.NodeFlags.BlockScoped)) {
                for (const declaration of node.declarations) checkName(declaration.name);
            }
        }
        if (node.parent === body && (ts.isFunctionDeclaration(node) ||
            ts.isClassDeclaration(node) || ts.isEnumDeclaration(node) ||
            ts.isModuleDeclaration(node)) && node.name && ts.isIdentifier(node.name)) {
            checkName(node.name);
        }
        // Nested functions have their own parameters and locals. A shadow
        // there does not collide. Namespaces and static blocks also keep
        // their var declarations in a separate scope.
        if (ts.isFunctionLike(node) || ts.isModuleDeclaration(node) ||
            ts.isClassStaticBlockDeclaration(node)) return;
        ts.forEachChild(node, visit);
    }

    ts.forEachChild(body, visit);
}

function plan(source, pin, newline) {
    const factory = owner(source, pin.name);
    const statements = factory.body.statements;
    const declarations = statements.slice(0, pin.before.length);
    assert.equal(declarations.length, pin.before.length, pin.name + ': missing declarations');
    assert(declarations.every(ts.isVariableStatement), pin.name + ': expected variable statements');
    checkBindings(source, factory, pin.names, declarations);
    const start = declarations[0].getStart(source);
    const end = declarations.at(-1).end;
    const before = pin.before.join(newline + '    ');
    const after = pin.after.join(newline + '    ');
    const current = source.text.slice(start, end);
    assert(current === before || current === after, pin.name + ': declaration snapshot drift');

    if (pin === multiline) {
        const returnIndex = statements.length - 1;
        const helpers = statements.slice(declarations.length, returnIndex - 1);
        const resets = helpers.filter(statement =>
            ts.isFunctionDeclaration(statement) && statement.name?.text === 'reset');
        assert(helpers.every(ts.isFunctionDeclaration), 'multiline writer executes before reset');
        assert.equal(resets.length, 1, 'expected one multiline reset declaration');
        assert.equal(resets[0].getText(source), reset.join(newline), 'multiline reset snapshot drift');
        assert.equal(statements[returnIndex - 1]?.getText(source), 'reset();', 'initial reset moved');
        assert(ts.isReturnStatement(statements[returnIndex]), 'multiline writer return moved');
    }
    return { start, end, after, applied: current === after };
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const file = path.join(tree, 'src/compiler/utilities.ts');
    const text = fs.readFileSync(file, 'utf8');
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'utilities.ts parse diagnostics');
    const edits = [plan(source, single, newline), plan(source, multiline, newline)];
    assert.equal(edits[0].applied, edits[1].applied, 'writer declarations are partially adapted');
    if (edits[0].applied) return;
    let updated = text;
    for (const edit of edits.sort((left, right) => right.start - left.start)) {
        updated = updated.slice(0, edit.start) + edit.after + updated.slice(edit.end);
    }
    fs.writeFileSync(file, updated);
}

assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
adapt(path.resolve(process.argv[2]));
