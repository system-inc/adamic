#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, "src/compiler/builder.ts");
const program = compiler(tree, ts);
const source = program.getSourceFile(fileName);
assert(source && !source.parseDiagnostics.length, "cannot parse builder.ts");

const edits = [];
const resolutions = [];
const owners = nodes(ts, source, node => ts.isFunctionDeclaration(node) && node.name?.text === "toFileIdListId");
assert.equal(owners.length, 1, "expected one toFileIdListId");
const statements = owners[0].body.statements;
const first = statements[0];
const start = first.getStart(source);
const before = "const fileIds = arrayFrom(set.keys(), toFileId).sort(compareValues);";
const after = [
    "const fileIds: IncrementalBuildInfoFileId[] = [];",
    "for (const path of set.keys()) {",
    "    fileIds.push(toFileId(path));",
    "}",
    "fileIds.sort(compareValues);",
].join("\r\n        ");

assert.equal(source.getLineAndCharacterOfPosition(start).character, 8, "unexpected indentation");
if (source.text.slice(start, statements[2]?.end) === after) {
    console.log("105: toFileIdListId already adapted");
} else {
    assert.equal(first.getText(source), before, "unexpected snapshot in toFileIdListId");
    const calls = nodes(ts, first, node => ts.isCallExpression(node) && node.getText(source) === "arrayFrom(set.keys(), toFileId)");
    assert.equal(calls.length, 1, "expected one pinned arrayFrom call");
    const declaration = program.getTypeChecker().getResolvedSignature(calls[0])?.declaration;
    const core = declaration?.getSourceFile();
    const overload = "export function arrayFrom<T, U>(iterator: Iterable<T>, map: (t: T) => U): U[];";
    assert.equal(core?.fileName, path.join(tree, "src/compiler/core.ts"), "arrayFrom must resolve to core.ts");
    assert.equal(declaration.getText(core), overload, "arrayFrom must resolve to the mapped overload");

    edits.push({ start, end: first.end, after });
    const where = source.getLineAndCharacterOfPosition(calls[0].getStart(source));
    resolutions.push(`105: builder.ts:${where.line + 1}:${where.character + 1} resolved ${overload}`);
}

{
    const owners = nodes(ts, source, node => ts.isFunctionDeclaration(node) && node.name?.text === "toPerFileEmitDiagnostics");
    assert.equal(owners.length, 1, "expected one toPerFileEmitDiagnostics");
    const statements = owners[0].body.statements;
    const first = statements[0];
    const start = first.getStart(source);
    const before = "return diagnostics && arrayToMap(diagnostics, value => toFilePath(value[0]), value => value[1]);";
    const after = [
        "if (!diagnostics) return diagnostics;",
        "const result = new Map<Path, readonly ReusableDiagnostic[]>();",
        "for (let i = 0; i < diagnostics.length; i++) {",
        "    const value = diagnostics[i]!;",
        "    const path: Path | undefined = toFilePath(value[0]);",
        "    if (path !== undefined) result.set(path, value[1]);",
        "}",
        "return result;",
    ].join("\r\n        ");

    assert.equal(source.getLineAndCharacterOfPosition(start).character, 8, "unexpected indentation");
    if (statements.length === 4 && source.text.slice(start, statements[3].end) === after) {
        console.log("106: toPerFileEmitDiagnostics already adapted");
    } else {
        assert.equal(statements.length, 1, "unexpected statements in toPerFileEmitDiagnostics");
        assert.equal(first.getText(source), before, "unexpected snapshot in toPerFileEmitDiagnostics");
        const call = first.expression.right;
        assert(ts.isCallExpression(call), "expected the pinned arrayToMap call");
        const checker = program.getTypeChecker();
        const signature = checker.getResolvedSignature(call);
        const declaration = signature?.declaration;
        const core = declaration?.getSourceFile();
        const overload = "export function arrayToMap<K, V1, V2>(array: readonly V1[], makeKey: (value: V1) => K | undefined, makeValue: (value: V1) => V2): Map<K, V2>;";
        assert.equal(core?.fileName, path.join(tree, "src/compiler/core.ts"), "arrayToMap must resolve to core.ts");
        assert.equal(declaration.getText(core), overload, "arrayToMap must resolve to the mapped-value overload");
        assert.equal(checker.typeToString(signature.getReturnType()), "Map<Path, readonly ReusableDiagnostic[]>");

        edits.push({ start, end: first.end, after });
        const where = source.getLineAndCharacterOfPosition(call.getStart(source));
        resolutions.push(`106: builder.ts:${where.line + 1}:${where.character + 1} resolved ${overload}`);
    }
}

if (edits.length) {
    let output = source.text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) {
        output = output.slice(0, edit.start) + edit.after + output.slice(edit.end);
    }
    assert.equal(ts.createSourceFile(fileName, output, ts.ScriptTarget.Latest).parseDiagnostics.length, 0);
    assert.equal(fs.readFileSync(fileName, "utf8"), source.text, "builder.ts changed during adaptation");
    fs.writeFileSync(fileName, output);
    for (const line of resolutions) console.log(line);
}
