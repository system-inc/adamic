#!/usr/bin/env node
"use strict";
// Neither candidate passes native compilation on the pinned area compiler.
// Keep discovery reproducible without applying an unproved source adaptation.
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
if (declarations[0].getText(source).replace(/\s+/g, " ") !== expected) throw new Error("unreviewed memoize implementation");
console.log(JSON.stringify({files: 0, edits: 0, status: "declined", reason: "A and B fail native compilation; explicit presence probes both reach adamic/cycle-capable"}));
