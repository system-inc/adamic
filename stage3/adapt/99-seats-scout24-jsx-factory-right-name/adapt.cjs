'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const relative = 'src/compiler/factory/utilities.ts';
const name = 'createJsxFactoryExpressionFromEntityName';
const before = [
    'function createJsxFactoryExpressionFromEntityName(factory: NodeFactory, jsxFactory: EntityName, parent: JsxOpeningLikeElement | JsxOpeningFragment): Expression {',
    '    if (isQualifiedName(jsxFactory)) {',
    '        const left = createJsxFactoryExpressionFromEntityName(factory, jsxFactory.left, parent);',
    '        const right = factory.createIdentifier(idText(jsxFactory.right)) as Mutable<Identifier>;',
    '        right.escapedText = jsxFactory.right.escapedText;',
    '        return factory.createPropertyAccessExpression(left, right);',
    '    }',
    '    else {',
    '        return createReactNamespace(idText(jsxFactory), parent);',
    '    }',
    '}',
].join('\r\n');
const after = before
    .replace(' as Mutable<Identifier>;', ';')
    .replace('\r\n        right.escapedText = jsxFactory.right.escapedText;', '');

function findOwner(text, ownerName = name) {
    assert.equal(ts.version, '6.0.3');
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const owners = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === ownerName);
    assert.equal(owners.length, 1, ownerName);
    return owners[0];
}

function plan(text) {
    const owner = findOwner(text);
    const snapshot = owner.getText();
    if (snapshot === after) return text;
    assert.equal(snapshot, before, 'unreviewed JSX factory snapshot');

    const [branch] = owner.body.statements;
    assert(ts.isIfStatement(branch) && ts.isBlock(branch.thenStatement));
    const [, right, write] = branch.thenStatement.statements;
    assert(ts.isVariableStatement(right));
    const [binding] = right.declarationList.declarations;
    assert.equal(binding.name.getText(), 'right');
    assert(ts.isAsExpression(binding.initializer));
    assert.equal(binding.initializer.type.getText(), 'Mutable<Identifier>');
    assert(ts.isExpressionStatement(write));
    assert.equal(write.getText(), 'right.escapedText = jsxFactory.right.escapedText;');
    return text.slice(0, owner.getStart()) + after + text.slice(owner.end);
}

module.exports = { relative, findOwner, plan };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <applied tree>');
    const file = path.join(process.argv[2], relative);
    const text = fs.readFileSync(file, 'utf8');
    const adapted = plan(text);
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'source changed during adaptation');
    if (adapted !== text) fs.writeFileSync(file, adapted);
    console.log(name + ': ' + (adapted === text ? '0 edits' : '1 edit'));
}
