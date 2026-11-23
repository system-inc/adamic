#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");

const relative = "src/compiler/sourcemap.ts";
const filename = path.join(process.argv[2], relative);
const text = fs.readFileSync(filename, "utf8");
const source = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
const newline = "\r\n";

function one(nodes, description) {
    assert.equal(nodes.length, 1, `expected one ${description}`);
    return nodes[0];
}

function namedFunction(statements, name) {
    return one(statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body), name);
}

function snapshot(name) {
    return fs.readFileSync(path.join(__dirname, name + ".before.txt"), "utf8")
        .replace(/\r?\n$/, "").replace(/\r?\n/g, newline);
}

function replaceOnce(original, before, after) {
    assert.equal(original.split(before).length, 2, `expected one snapshot fragment: ${before}`);
    return original.replace(before, () => after);
}

const decoderBefore = snapshot("decoder");
const pinnedSource = ts.createSourceFile(relative, decoderBefore, ts.ScriptTarget.Latest, true);
const pinnedDecoder = namedFunction(pinnedSource.statements, "decodeMappings");
const pinnedReturn = one(pinnedDecoder.body.statements.filter(ts.isReturnStatement), "decoder object return");
assert.ok(ts.isObjectLiteralExpression(pinnedReturn.expression));
const pinnedNext = one(pinnedReturn.expression.properties.filter(node => ts.isMethodDeclaration(node) && node.name.getText(pinnedSource) === "next"), "decoder next method");
const stepBody = replaceOnce(pinnedNext.body.getText(pinnedSource), "value: captureMapping(hasSource, hasName), done", "value: captureMapping(hasSource, hasName), done: false")
    .replaceAll(newline + "    ", newline);
let iterator = replaceOnce(pinnedReturn.expression.getText(pinnedSource), pinnedNext.getText(pinnedSource),
    ["next() {", "            return step();", "        }"].join(newline));
iterator = replaceOnce(iterator, "return pos;", "return readPos();");
iterator = replaceOnce(iterator, "return error;", "return readError();");
iterator = replaceOnce(iterator, "return captureMapping(/*hasSource*/ true, /*hasName*/ true);", "return readState();");

const comment = "// TODO(jakebailey): can we implement this without writing next ourselves?";
let factory = replaceOnce(decoderBefore, "export function decodeMappings(mappings: string): MappingsDecoder", "function createMappingDecoder(mappings: string)");
factory = replaceOnce(factory, "    " + comment + newline, "");
factory = replaceOnce(factory, pinnedReturn.getText(pinnedSource), [
    "return { step, readPos, readError, readState };",
    "",
    "    function step(): { value: Mapping; done: false } | { value: undefined; done: true } " + stepBody,
    "",
    "    function readPos(): number {",
    "        return pos;",
    "    }",
    "",
    "    function readError(): string | undefined {",
    "        return error;",
    "    }",
    "",
    "    function readState(): Required<Mapping> {",
    "        return captureMapping(/*hasSource*/ true, /*hasName*/ true);",
    "    }",
].join(newline));

const decoderAfter = [
    "export function decodeMappings(mappings: string): MappingsDecoder {",
    "    const { step, readPos, readError, readState } = createMappingDecoder(mappings);",
    "",
    "    " + comment,
    "    return " + iterator + ";",
    "}",
    "",
    factory,
].join(newline);

const appendBefore = snapshot("append");
const appendAfter = replaceOnce(appendBefore, [
    "const mappingIterator = decodeMappings(map.mappings);",
    "        for (const raw of mappingIterator) {",
].join(newline), [
    "const { step } = createMappingDecoder(map.mappings);",
    "        for (;;) {",
    "            const result = step();",
    "            if (result.done) break;",
    "            const raw = result.value;",
].join(newline));

const decoder = namedFunction(source.statements, "decodeMappings");
const generator = namedFunction(source.statements, "createSourceMapGenerator");
const append = namedFunction(generator.body.statements, "appendSourceMap");
const factories = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === "createMappingDecoder");
assert.ok(factories.length <= 1, "multiple private decoder factories");
assert.ok(append.end < decoder.getStart(source), "expected appendSourceMap before decodeMappings");
const decoderEnd = factories[0]?.end ?? decoder.end;
const decoderText = text.slice(decoder.getStart(source), decoderEnd);
const appendText = append.getText(source);

if (decoderText === decoderAfter && appendText === appendAfter) {
    console.log(JSON.stringify({ family: 164, file: relative, edits: 0 }));
} else {
    assert.equal(decoderText, decoderBefore, "unexpected decodeMappings snapshot");
    assert.equal(appendText, appendBefore, "unexpected appendSourceMap snapshot");
    const withDecoder = text.slice(0, decoder.getStart(source)) + decoderAfter + text.slice(decoderEnd);
    fs.writeFileSync(filename, withDecoder.slice(0, append.getStart(source)) + appendAfter + withDecoder.slice(append.end));
    console.log(JSON.stringify({ family: 164, file: relative, edits: 2 }));
}
