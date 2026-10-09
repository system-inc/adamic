#!/usr/bin/env node
"use strict";
// Run only on an isolated copy of the adapted tree.
const fs = require("node:fs");
const path = require("node:path");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { nodes } = require("./adapt.cjs");
if (process.argv.length !== 3) throw new Error("usage: mutate-buffer.cjs <scratch-copy>");
const file = path.join(path.resolve(process.argv[2]), "src/compiler/sys.ts");
const text = fs.readFileSync(file, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const selected = nodes(ts, source, node => ts.isNonNullExpression(node) && node.expression.getText(source) === "buffer[i + 1]");
if (selected.length !== 1) throw new Error("ambiguous byte-read mutant");
const node = selected[0];
fs.writeFileSync(file, text.slice(0, node.end - 1) + text.slice(node.end));
console.log("removed byte-pair RHS assertion");
