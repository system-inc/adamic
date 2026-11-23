#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const crypto = require("node:crypto");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const snapshots = {
    "before": {
        "PragmaDefinition": "f12f877a548bce6118e95c17a0a1c7614f0ba3797a9f4e5a5257749f159aec02",
        "commentPragmas": "10dcb58989aed204d33c28597385c57a771a26c67a7018aca0173a2bb4b2b04e",
        "ConcretePragmaSpecs": "5000576841d9bcc2a9bb1eedc8cb93b353758408717b29dd4ed3fc69a2e8e8ed",
        "getNamedPragmaArguments": "1596c77a8d428a6c0004163422f7c1e746bc11d091876e9be31d6b1657d95ac7"
    },
    "after": {
        "PragmaDefinition": "2e2b298a255224673bf11c66cbecc2d64090ad0e133c1ee8a9ea4ead986da84b",
        "commentPragmas": "54b1956380661091877a5bb12f7cfc9de8a94e26440249d402df8cde2829cf13",
        "ConcretePragmaSpecs": "a2bfec6ebc831e1b58175395d0875513a154bd131d2451b44e6abd253673cec1",
        "getNamedPragmaArguments": "36161e2c60e8cbd8769dedc91c73ca5fd2b220c4995be5c3011ec2af97a52392"
    }
};

function only(nodes, predicate, name) {
    const matches = nodes.filter(predicate);
    assert.equal(matches.length, 1, name);
    return matches[0];
}

function parse(tree, basename) {
    const file = path.join(tree, "src/compiler", basename);
    const text = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, basename);
    return { file, text, source, edits: [] };
}

function named(node, name) {
    return node.name?.text === name;
}

function owners(types, parser) {
    const declaration = name => only(types.source.statements,
        node => ts.isInterfaceDeclaration(node) && named(node, name), name);
    const table = only(types.source.statements.filter(ts.isVariableStatement)
        .flatMap(node => [...node.declarationList.declarations]),
        node => named(node, "commentPragmas"), "commentPragmas");
    const consumer = only(parser.source.statements,
        node => ts.isFunctionDeclaration(node) && named(node, "getNamedPragmaArguments"),
        "getNamedPragmaArguments");
    const definition = declaration("PragmaDefinition");
    const specs = declaration("ConcretePragmaSpecs");
    for (const node of [definition, table.parent.parent, specs]) {
        assert(ts.getJSDocTags(node).some(tag => tag.tagName.text === "internal"), "public pragma declaration");
    }
    return [definition, table, specs, consumer];
}

function fingerprints(nodes) {
    return Object.fromEntries(nodes.map(node => [node.name.text,
        crypto.createHash("sha256").update(node.getFullText()).digest("hex")]));
}

function insertBefore(file, node, text) {
    const start = node.getStart(file.source);
    const lineStart = file.text.lastIndexOf("\n", start) + 1;
    const indent = file.text.slice(lineStart, start);
    assert.match(indent, /^ +$/);
    const eol = file.text.includes("\r\n") ? "\r\n" : "\n";
    file.edits.push({ start, end: start, text: text + eol + indent });
}

function plan(types, parser, [definition, table, specs, consumer]) {
    const args = node => node.name?.text === "args";
    insertBefore(types, only(definition.members, args, "PragmaDefinition.args"), "readonly argCount?: number;");
    assert(ts.isAsExpression(table.initializer) && ts.isObjectLiteralExpression(table.initializer.expression));
    const counted = [];
    for (const entry of table.initializer.expression.properties) {
        assert(ts.isPropertyAssignment(entry) && ts.isObjectLiteralExpression(entry.initializer));
        const slot = entry.initializer.properties.find(args);
        if (!slot) continue;
        assert(ts.isPropertyAssignment(slot) && ts.isArrayLiteralExpression(slot.initializer));
        assert(slot.initializer.elements.every(ts.isObjectLiteralExpression));
        insertBefore(types, slot, "argCount: " + slot.initializer.elements.length + ",");
        counted.push(entry.name.text);
    }
    const countSpecs = [];
    for (const spec of specs.members) {
        assert(ts.isPropertySignature(spec) && ts.isTypeLiteralNode(spec.type) && ts.isStringLiteral(spec.name));
        const slot = spec.type.members.find(args);
        if (!slot) continue;
        const ref = "ConcretePragmaSpecs[" + spec.name.getText() + ']["args"]["length"]';
        insertBefore(types, slot, "readonly argCount: " + ref + ";");
        countSpecs.push(spec.name.text);
    }
    assert.deepEqual(counted, countSpecs);

    const loop = only(consumer.body.statements, ts.isForStatement, "pragma loop");
    const condition = loop.condition;
    assert(ts.isBinaryExpression(condition) && condition.operatorToken.kind === ts.SyntaxKind.LessThanToken);
    const length = condition.right;
    assert(ts.isPropertyAccessExpression(length) && length.getText() === "pragma.args.length");
    parser.edits.push({ start: length.getStart(parser.source), end: length.end, text: "pragma.argCount!" });
}

function applyEdits(file) {
    let result = file.text;
    for (const edit of file.edits.sort((a, b) => b.start - a.start)) {
        result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
    }
    return result;
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const types = parse(tree, "types.ts");
    const parser = parse(tree, "parser.ts");
    const declarations = owners(types, parser);
    const current = fingerprints(declarations);
    if (JSON.stringify(current) === JSON.stringify(snapshots.after)) return { files: 0, edits: 0 };
    assert.deepEqual(current, snapshots.before, "unreviewed pragma snapshots");
    plan(types, parser, declarations);
    for (const file of [types, parser]) {
        assert.equal(fs.readFileSync(file.file, "utf8"), file.text, file.file);
    }
    for (const file of [types, parser]) fs.writeFileSync(file.file, applyEdits(file));
    return { files: 2, edits: types.edits.length + parser.edits.length };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
