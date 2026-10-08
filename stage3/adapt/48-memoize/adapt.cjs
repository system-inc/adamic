#!/usr/bin/env node
"use strict";
// A adds the value the callback cell really holds; emitted JavaScript is unchanged.
// Native acceptance is recorded separately from this tsc source adaptation.
const fs = require("node:fs");
const path = require("node:path");
if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
const ts = require("typescript");
if (ts.version !== "6.0.3") throw new Error("requires TypeScript 6.0.3");
const file = path.join(process.argv[2], "src/compiler/core.ts");
const source = ts.createSourceFile(file, fs.readFileSync(file, "utf8"), ts.ScriptTarget.Latest, true);
const declarations = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === "memoize");
if (source.parseDiagnostics.length || declarations.length !== 1) throw new Error("memoize source drift");
const expected = "export function memoize<T>(callback: () => T): () => T { let value: T; return () => { if (callback) { value = callback(); callback = undefined!; } return value; }; }";
const declaration = declarations[0];
const current = declaration.getText(source);
const adapted = expected.replace("callback: () => T", "callback: (() => T) | undefined").replace("undefined!", "undefined");
const normalized = current.replace(/\s+/g, " ");
if (normalized !== expected && normalized !== adapted) throw new Error("unreviewed memoize implementation");
if (normalized === adapted) {
    console.log(JSON.stringify({files: 0, edits: 0, status: "applied", alreadyApplied: true}));
} else {
    const replacement = current.replace("callback: () => T", "callback: (() => T) | undefined").replace("callback = undefined!;", "callback = undefined;");
    const emit = code => ts.transpileModule(code, {compilerOptions: {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext}}).outputText;
    if (emit(current) !== emit(replacement)) throw new Error("memoize JavaScript changed");
    const next = source.text.slice(0, declaration.getStart(source)) + replacement + source.text.slice(declaration.end);
    if (ts.createSourceFile(file, next, ts.ScriptTarget.Latest, true).parseDiagnostics.length) throw new Error("adapted syntax error");
    fs.writeFileSync(file, next);
    console.log(JSON.stringify({files: 1, edits: 2, status: "applied", javascriptIdentical: true}));
}
