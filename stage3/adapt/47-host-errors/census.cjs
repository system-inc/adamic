#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { compiler, nodes } = require("./adapt.cjs");
const tree = path.resolve(process.argv[2]);
const program = compiler(tree, ts);
const root = path.join(tree, "src/compiler/");
const files = program.getSourceFiles().filter(file => file.fileName.startsWith(root));
const diagnostics = ts.getPreEmitDiagnostics(program).filter(d => d.file?.fileName.startsWith(root));
function location(file, position) {
    const point = file.getLineAndCharacterOfPosition(position);
    return { file: path.relative(tree, file.fileName), line: point.line + 1, column: point.character + 1 };
}
const unknown = diagnostics.filter(d => d.code === 18046).map(d => ({ ...location(d.file, d.start), code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, "\n") }));
const catches = files.flatMap(file => nodes(ts, file, ts.isCatchClause).map(node => ({ ...location(file, node.getStart(file)), binding: node.variableDeclaration?.name.getText(file) ?? null, propertyReads: nodes(ts, node.block, ts.isPropertyAccessExpression).map(read => read.getText(file)) })));
const sys = files.find(file => path.basename(file.fileName) === "sys.ts");
const readFile = nodes(ts, sys, node => ts.isFunctionDeclaration(node) && node.name?.text === "readFile")[0];
const reads = nodes(ts, readFile, node => ts.isElementAccessExpression(node) && node.expression.getText(sys) === "buffer").filter(node => !(ts.isBinaryExpression(node.parent) && node.parent.left === node && node.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken));
const bytes = reads.map(node => ({ ...location(sys, node.getStart(sys)), expression: node.getText(sys), asserted: ts.isNonNullExpression(node.parent), diagnostics: diagnostics.filter(d => d.file === sys && d.start <= node.end && d.start + (d.length || 0) >= node.getStart(sys)).map(d => ({ code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, "\n") })) }));
const bufferDiagnostics = diagnostics.filter(d => d.file === sys && d.start >= readFile.getStart(sys) && d.start < readFile.end).map(d => ({ ...location(d.file, d.start), code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, "\n") }));
const report = { bufferDiagnostics, typescript: ts.version, options: program.getCompilerOptions(), roots: program.getRootFileNames().length, diagnosticCount: diagnostics.length, catchCount: catches.length, boundCatchCount: catches.filter(c => c.binding !== null).length, TS18046: unknown, byteReads: bytes, catches };
fs.writeFileSync(process.argv[3], JSON.stringify(report, null, 2) + "\n");
console.log(JSON.stringify({ catchCount: report.catchCount, boundCatchCount: report.boundCatchCount, TS18046: unknown.length, byteReadsUnasserted: bytes.filter(b => !b.asserted).length, byteDiagnostics: bufferDiagnostics.length }));

if (process.argv[4] === "--check" && (bytes.some(b => !b.asserted) || bufferDiagnostics.length || unknown.length !== 4)) { console.error("census failed: required byte presence or documented catch remainder changed"); process.exitCode = 1; }
