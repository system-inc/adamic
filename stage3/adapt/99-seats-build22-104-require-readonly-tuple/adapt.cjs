'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const relative = 'src/compiler/types.ts';
const mutableTuple = '[StringLiteralLike]';
const readonlyTuple = 'readonly [StringLiteralLike]';

function argumentType(text) {
    assert.equal(ts.version, '6.0.3');
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const aliases = source.statements.filter(node =>
        ts.isTypeAliasDeclaration(node) && node.name.text === 'RequireOrImportCall');
    assert.equal(aliases.length, 1, 'expected one RequireOrImportCall');
    const type = aliases[0].type;
    assert(ts.isIntersectionTypeNode(type) && type.types.length === 2);
    const [call, fields] = type.types;
    assert.equal(call.getText(), 'CallExpression');
    assert(ts.isTypeLiteralNode(fields) && fields.members.length === 2);
    const [callee, args] = fields.members;
    assert(ts.isPropertySignature(callee) && ts.isPropertySignature(args));
    assert.equal(callee.getText(), 'expression: Identifier;');
    assert.equal(args.name.getText(), 'arguments');
    assert(args.type && !args.questionToken);
    return args.type;
}

function plan(text) {
    const type = argumentType(text);
    const actual = type.getText();
    if (actual === readonlyTuple) return text;
    assert.equal(actual, mutableTuple, 'unreviewed RequireOrImportCall.arguments');
    return text.slice(0, type.getStart()) + readonlyTuple + text.slice(type.end);
}

function adapt(tree) {
    const file = path.join(tree, relative);
    const text = fs.readFileSync(file, 'utf8');
    const output = plan(text);
    if (output === text) return { files: 0, edits: 0 };
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'source changed during adaptation');
    fs.writeFileSync(file, output);
    return { files: 1, edits: 1 };
}

module.exports = { argumentType, plan, relative };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
