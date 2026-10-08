#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require("typescript");
const tree = path.resolve(process.argv[2]);
const output = process.argv[3];
if (!output || ts.version !== "6.0.3") throw new Error("usage: check-callers.cjs <installed-tree> <output.json>");
const config = path.join(tree, "src/compiler/tsconfig.json");
const loaded = ts.readConfigFile(config, ts.sys.readFile);
if (loaded.error) throw new Error("cannot read compiler config");
const parsed = ts.parseJsonConfigFileContent(loaded.config, ts.sys, path.dirname(config), {}, config);
if (parsed.errors.length) throw new Error("cannot parse compiler config");
const options = {...parsed.options, noEmit: true, composite: false, isolatedDeclarations: false};
const file = path.join(tree, "src/compiler/core.ts");
const original = fs.readFileSync(file, "utf8");
const sf = ts.createSourceFile(file, original, ts.ScriptTarget.Latest, true);
const fn = sf.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === "memoize");
const start = fn.getStart(sf), end = fn.end;
let a = fn.getText(sf).replace("callback: () => T", "callback: (() => T) | undefined").replace("undefined!", "undefined");
if (process.argv[4] === "--mutant") a = a.replace("callback: (() => T) | undefined", "callback: string");
function inspect(text) {
    const host = ts.createCompilerHost(options);
    const read = host.readFile;
    host.readFile = name => path.resolve(name) === file ? text : read(name);
    const program = ts.createProgram(parsed.fileNames, options, host);
    const checker = program.getTypeChecker();
    const diagnostics = program.getSemanticDiagnostics();
    const callers = [];
    for (const source of program.getSourceFiles()) {
        if (!source.fileName.startsWith(path.join(tree, "src/compiler"))) continue;
        function visit(node) {
            if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === "memoize") {
                const position = source.getLineAndCharacterOfPosition(node.getStart(source));
                callers.push({file: path.relative(tree, source.fileName), line: position.line + 1,
                    source: node.getText(source), type: checker.typeToString(checker.getTypeAtLocation(node)),
                    diagnostics: diagnostics.filter(d => d.file === source && d.start >= node.getStart(source) && d.start < node.end)
                        .map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, "\n")}))});
            }
            ts.forEachChild(node, visit);
        }
        visit(source);
    }
    return {callers, diagnostics: diagnostics.map(d => ({file: d.file && path.relative(tree, d.file.fileName), code: d.code,
        message: ts.flattenDiagnosticMessageText(d.messageText, "\n")}))};
}
const before = inspect(original);
const after = inspect(original.slice(0, start) + a + original.slice(end));
assert.equal(before.callers.length, 22);
assert.deepEqual(after.callers, before.callers);
assert.ok(before.callers.every(c => c.diagnostics.length === 0));
assert.deepEqual(after.diagnostics, before.diagnostics);
const report = {node: process.version, typescript: ts.version, candidate: "A", source_changed_on_disk: false,
    unchanged_callers: before.callers, before_diagnostics: before.diagnostics, after_diagnostics: after.diagnostics};
fs.writeFileSync(output, JSON.stringify(report, null, 2) + "\n");
console.log(JSON.stringify({unchanged_callers: before.callers.length, caller_diagnostics: before.callers.reduce((n, c) => n + c.diagnostics.length, 0), total_diagnostics: before.diagnostics.length, unchanged: true}));
