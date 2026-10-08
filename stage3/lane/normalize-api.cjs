#!/usr/bin/env node
'use strict';
// Formatting-independent declarations, with every type-bearing AST child retained.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const cache = process.env.STAGE3_CACHE || path.join(require('node:os').homedir(), '.cache/adamic-stage3');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || path.join(cache, 'api/node_modules/typescript'));
assert.equal(ts.version, '6.0.3');
function canonical(node) {
    if (ts.isParenthesizedTypeNode(node)) return canonical(node.type);
    const result = [ts.SyntaxKind[node.kind]];
    if (ts.isIdentifier(node) || ts.isPrivateIdentifier(node) || ts.isLiteralExpression(node)) result.push(node.text);
    if (node.isTypeOnly !== undefined) result.push(['isTypeOnly', node.isTypeOnly]);
    if (node.isExportEquals !== undefined) result.push(['isExportEquals', node.isExportEquals]);
    ts.forEachChild(node, child => { result.push(canonical(child)); });
    return result;
}
function declarations(text) {
    const source = ts.createSourceFile('api.d.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'invalid API snapshot');
    const entries = new Map();
    function record(key, value) {
        let occurrence = 0;
        while (entries.has(`${key}#${occurrence}`)) occurrence++;
        entries.set(`${key}#${occurrence}`, JSON.stringify(value));
    }
    function name(node) {
        if (!node.name) return '';
        return ts.isIdentifier(node.name) || ts.isStringLiteral(node.name)
            ? node.name.text : JSON.stringify(canonical(node.name));
    }
    function visit(node, scope) {
        const owner = [...scope, name(node)].filter(Boolean);
        const key = ts.SyntaxKind[node.kind] + ':' + JSON.stringify(owner);
        if (ts.isModuleDeclaration(node)) {
            record(key, [node.modifiers?.map(canonical) || [], canonical(node.name),
                         !!(node.flags & ts.NodeFlags.GlobalAugmentation)]);
            if (node.body && ts.isModuleBlock(node.body)) node.body.statements.forEach(child => visit(child, owner));
            else if (node.body) visit(node.body, owner);
        } else if (ts.isInterfaceDeclaration(node) || ts.isClassDeclaration(node) || ts.isEnumDeclaration(node)) {
            record(key, [node.modifiers?.map(canonical) || [], canonical(node.name),
                         node.typeParameters?.map(canonical) || [], node.heritageClauses?.map(canonical) || []]);
            node.members.forEach(member => record('member:' + JSON.stringify([...owner, name(member)]), canonical(member)));
        } else {
            record(key, canonical(node));
        }
    }
    source.statements.forEach(node => visit(node, []));
    // Comment wording is also part of the snapshot; permit only whitespace reflow.
    const scanner = ts.createScanner(ts.ScriptTarget.Latest, false, ts.LanguageVariant.Standard, text);
    const comments = [];
    for (let kind = scanner.scan(); kind !== ts.SyntaxKind.EndOfFileToken; kind = scanner.scan()) {
        if (kind === ts.SyntaxKind.SingleLineCommentTrivia || kind === ts.SyntaxKind.MultiLineCommentTrivia) {
            comments.push([ts.SyntaxKind[kind], scanner.getTokenText().replace(/\s+/g, ' ')]);
        }
    }
    record('comments', comments);
    return entries;
}
function changes(before, after) {
    const old = declarations(before), current = declarations(after);
    return [...new Set([...old.keys(), ...current.keys()])].sort().flatMap(declaration =>
        old.get(declaration) === current.get(declaration) ? [] : [{ declaration,
            before: old.get(declaration) ?? null, after: current.get(declaration) ?? null }]);
}
module.exports = { canonical, declarations, changes };
if (require.main === module) {
    assert([4, 5].includes(process.argv.length), 'usage: normalize-api.cjs <original> <reference> [<local>]');
    const texts = process.argv.slice(2).map(file => fs.readFileSync(file, 'utf8'));
    const result = texts.length === 2 ? changes(...texts)
        : { reference_changes: changes(texts[0], texts[1]), composed_changes: changes(texts[0], texts[2]) };
    console.log(JSON.stringify(result));
}
