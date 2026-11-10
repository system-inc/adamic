'use strict';
// A never-typed phantom owner, separate from the any-only brands class.
const fs = require('node:fs'), path = require('node:path');
const assert = require('node:assert/strict'), crypto = require('node:crypto');
const ts = require('typescript'), rule = require('./resolved-config-brand.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function plan(text, file = rule.file, reverse = false) {
    assert.equal(ts.version, '6.0.3');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const owners = [];
    function visit(node) {
        if (ts.isTypeAliasDeclaration(node) && node.name.text === rule.owner) owners.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(owners.length, 1, 'exactly one resolved-config brand owner');
    const node = owners[0], spelling = node.getText(source).replace(/\r\n/g, '\n');
    const states = ['before', 'after'].filter(state => spelling === rule[state] && digest(spelling) === rule[state + 'Sha256']);
    assert.equal(states.length, 1, 'resolved-config brand owner text drift');
    const member = node.type.types?.[1]?.members?.[0];
    assert(ts.isIntersectionTypeNode(node.type) && node.type.types.length === 2 &&
        node.type.types[0].kind === ts.SyntaxKind.StringKeyword &&
        ts.isTypeLiteralNode(node.type.types[1]) && node.type.types[1].members.length === 1 &&
        ts.isPropertySignature(member) && member.name.text === rule.key && !member.questionToken);
    const target = reverse ? 'never' : 'undefined';
    const result = text.slice(0, member.type.getStart(source)) + target + text.slice(member.type.end);
    return {text: result, files: Number(result !== text), state: states[0]};
}
function apply(tree) {
    const file = path.join(tree, rule.file), before = fs.readFileSync(file, 'utf8'), result = plan(before);
    assert.equal(fs.readFileSync(file, 'utf8'), before, 'concurrent source change');
    if (result.files) fs.writeFileSync(file, result.text);
    console.log(JSON.stringify({class: rule.class, files: result.files}));
    return result;
}
module.exports = {plan, apply, rule};
