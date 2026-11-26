#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
const relative = "src/compiler/sourcemap.ts";
const newline = "\r\n";
const before = [
    "function getDecodedMappings() {",
    "        if (decodedMappings === undefined) {",
    "            const decoder = decodeMappings(map.mappings);",
    "            const mappings = arrayFrom(decoder, processMapping);",
    "            if (decoder.error !== undefined) {",
    "                if (host.log) {",
    "                    host.log(`Encountered error while decoding sourcemap: ${decoder.error}`);",
    "                }",
    "                decodedMappings = emptyArray;",
    "            }",
    "            else {",
    "                decodedMappings = mappings;",
    "            }",
    "        }",
    "        return decodedMappings;",
    "    }",
].join(newline);
const after = [
    "function getDecodedMappings() {",
    "        if (decodedMappings === undefined) {",
    "            const decoder = decodeMappings(map.mappings);",
    "            const mappings = arrayFrom(decoder, processMapping);",
    "            const decoderError = decoder.error;",
    "            if (decoderError !== undefined) {",
    "                if (host.log) {",
    "                    host.log(`Encountered error while decoding sourcemap: ${decoderError}`);",
    "                }",
    "                decodedMappings = emptyArray;",
    "            }",
    "            else {",
    "                decodedMappings = mappings;",
    "            }",
    "        }",
    "        return decodedMappings;",
    "    }",
].join(newline);

function one(nodes, label) {
    assert.equal(nodes.length, 1, `expected one ${label}`);
    return nodes[0];
}

function plan(text) {
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, "sourcemap.ts parse diagnostics");

    function find(root, predicate) {
        const matches = [];
        function visit(node) {
            if (predicate(node)) matches.push(node);
            ts.forEachChild(node, visit);
        }
        visit(root);
        return matches;
    }

    const owner = one(source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === "createDocumentPositionMapper"),
    "top-level createDocumentPositionMapper");
    assert(owner.body, "expected createDocumentPositionMapper body");
    const mapper = one(find(source, node =>
        ts.isFunctionDeclaration(node) && node.name?.text === "getDecodedMappings"), "getDecodedMappings declaration");
    assert(mapper.parent === owner.body && mapper.body, "expected getDecodedMappings directly inside createDocumentPositionMapper");

    const snapshot = mapper.getText(source);
    const original = snapshot === before;
    assert(original || snapshot === after, "getDecodedMappings function text drift");
    function at(node, line, column, label) {
        const position = source.getLineAndCharacterOfPosition(node.getStart(source));
        const actual = `${position.line + 1}:${position.character + 1}`;
        assert.equal(actual, `${line}:${column}`, `${label} position drift`);
    }

    const operand = original ? "decoder.error" : "decoderError";
    const condition = one(find(mapper, node => ts.isIfStatement(node) &&
        ts.isBinaryExpression(node.expression) && node.expression.operatorToken.kind === ts.SyntaxKind.ExclamationEqualsEqualsToken &&
        node.expression.left.getText(source) === operand && node.expression.right.getText(source) === "undefined"),
    "decoder error test");
    const span = one(find(mapper, node => ts.isTemplateSpan(node) && node.expression.getText(source) === operand),
        "decoder error template span");
    const reads = find(mapper, node => ts.isPropertyAccessExpression(node) && node.getText(source) === "decoder.error");

    at(mapper, 765, 5, "getDecodedMappings");
    if (original) {
        assert(reads.length === 2 && reads[0] === condition.expression.left && reads[1] === span.expression,
            "expected decoder.error in the test and template");
        at(condition, 769, 13, "original if");
        at(reads[0].expression, 769, 17, "original test decoder");
        at(reads[1].expression, 771, 77, "original template decoder");
    } else {
        const read = one(reads, "replay decoder.error read");
        const declaration = read.parent;
        assert(ts.isVariableDeclaration(declaration) && declaration.initializer === read &&
            declaration.name.getText(source) === "decoderError" && (declaration.parent.flags & ts.NodeFlags.Const) !== 0 &&
            ts.isVariableStatement(declaration.parent.parent), "expected const decoderError initializer");
        at(declaration.parent.parent, 769, 13, "replay decoderError statement");
        at(read.expression, 769, 34, "replay initializer decoder");
        at(condition, 770, 13, "replay if");
        at(condition.expression.left, 770, 17, "replay test decoderError");
        at(span.expression, 772, 77, "replay template decoderError");
    }

    return snapshot === after ? text : text.slice(0, mapper.getStart(source)) + after + text.slice(mapper.end);
}

module.exports = { relative, before, after, plan };

if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
    const filename = path.join(process.argv[2], relative);
    const input = fs.readFileSync(filename);
    const text = input.toString("utf8");
    const updated = plan(text);
    const edits = updated === text ? 0 : 1;
    if (edits) {
        assert(fs.readFileSync(filename).equals(input), "sourcemap.ts changed before write");
        fs.writeFileSync(filename, updated);
    }
    console.log(JSON.stringify({ adapter: "38m7dap", file: relative, edits }));
}
