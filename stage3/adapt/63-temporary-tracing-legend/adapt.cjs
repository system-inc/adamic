'use strict';
const fs = require('node:fs'), path = require('node:path');
const assert = require('node:assert/strict'), crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript'), pins = require('./pins.json');
const sha = value => crypto.createHash('sha256').update(value).digest('hex');
const normalize = value => value.replace(/\r\n/g, '\n');
function inspect(text) {
    const source = ts.createSourceFile('tracing.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'tracing.ts parse drift');
    const owners = [], bindings = [];
    function walk(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === pins.owner && node.body) owners.push(node);
        if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === pins.binding) bindings.push(node);
        if (ts.isBinaryExpression(node) && node.operatorToken.kind >= ts.SyntaxKind.FirstAssignment &&
            node.operatorToken.kind <= ts.SyntaxKind.LastAssignment) {
            function assigned(target) {
                if (ts.isIdentifier(target)) assert.notEqual(target.text, pins.binding, 'binding reassigned');
                else if (ts.isArrayLiteralExpression(target)) target.elements.forEach(assigned);
                else if (ts.isObjectLiteralExpression(target)) target.properties.forEach(p => assigned(p.initializer || p.name));
            }
            assigned(node.left);
        }
        if ((ts.isPrefixUnaryExpression(node) || ts.isPostfixUnaryExpression(node)) &&
            [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(node.operator) && ts.isIdentifier(node.operand)) {
            assert.notEqual(node.operand.text, pins.binding, 'binding updated');
        }
        // Also covers nested literals and the local object literals spread into descriptor.
        if (ts.isObjectLiteralExpression(node)) for (const property of node.properties) {
            if (property.name) {
                assert(!ts.isComputedPropertyName(property.name), 'computed literal member drift');
                assert.notEqual(property.name.text, 'toJSON', 'literal declares toJSON');
            }
        }
        ts.forEachChild(node, walk);
    }
    walk(source);
    assert.equal(owners.length, 1, 'owner drift');
    const owner = owners[0];
    assert(ts.isModuleBlock(owner.parent) && ts.isModuleDeclaration(owner.parent.parent) &&
        owner.parent.parent.name.text === 'tracingEnabled', 'owner namespace drift');
    assert.equal(bindings.length, 1, 'binding drift');
    const binding = bindings[0];
    assert(binding.parent.flags & ts.NodeFlags.Const, 'binding must be const');
    assert(pins.binding === 'legend' ? ts.isArrayLiteralExpression(binding.initializer) : ts.isObjectLiteralExpression(binding.initializer),
        'binding must be an object or array literal');
    assert.equal(normalize(binding.getText(source)), normalize(pins.literal), 'literal drift');
    if (pins.binding === 'legend') {
        assert.equal(binding.parent.parent.parent, owner.parent, 'legend must remain namespace-level');
    } else {
        assert(binding.pos > owner.pos && binding.end < owner.end, 'descriptor must remain local');
    }
    return {source, owner, body: owner.getText(source)};
}
function plan(text) {
    assert.equal(ts.version, pins.version);
    assert.equal(pins.commit, '050880ce59e30b356b686bd3144efe24f875ebc8');
    for (const state of ['before', 'after']) assert.equal(sha(pins[state]), pins[state + 'Sha256'], 'pin corruption');
    assert.equal(pins.before.split(pins.callBefore).length, 2, 'pinned call count');
    assert.equal(pins.before.replace(pins.callBefore, pins.callAfter), pins.after, 'only reviewed call may change');
    const original = inspect(text), body = normalize(original.body);
    const state = body === normalize(pins.before) ? 'before' : body === normalize(pins.after) ? 'after' : undefined;
    assert(state, 'unreviewed or partially adapted ' + pins.owner);
    if (state === 'after') return text;
    const replacement = pins.after.replace(/\r\n/g, original.body.includes('\r\n') ? '\r\n' : '\n');
    const changed = text.slice(0, original.owner.getStart(original.source)) + replacement + text.slice(original.owner.end);
    assert.equal(inspect(changed).body, replacement, 'after owner drift');
    for (const removeComments of [false, true]) {
        const emit = value => ts.transpileModule(value, {compilerOptions: {
            target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments,
        }}).outputText;
        assert.equal(emit(changed), emit(text), 'runtime JavaScript changed');
    }
    assert.equal(plan(changed), changed, 'after validation');
    return changed;
}
function adapt(tree) {
    const file = path.join(tree, 'src/compiler/tracing.ts');
    const before = fs.readFileSync(file, 'utf8'), after = plan(before);
    assert.equal(fs.readFileSync(file, 'utf8'), before, 'concurrent source change');
    if (after !== before) fs.writeFileSync(file, after);
    return {files: Number(after !== before), edits: Number(after !== before), javascriptChanging: false};
}
module.exports = {inspect, plan, adapt};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(process.argv[2])));
}
