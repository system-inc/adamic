#!/usr/bin/env node
"use strict";

// A static import cannot retain a runtime load's conditional or lazy timing.
// This adapter inventories those calls and declines them instead of hoisting.
const fs = require("node:fs");
const path = require("node:path");
const { builtinModules } = require("node:module");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
if (ts.version !== "6.0.3") throw new Error(`want typescript@6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
const tree = path.resolve(process.argv[2]);
const lock = JSON.parse(fs.readFileSync(path.join(tree, "package-lock.json"), "utf8"));
if (lock.packages["node_modules/@types/node"].version !== "25.3.3") {
    throw new Error("want upstream @types/node@25.3.3");
}
const builtins = new Set(builtinModules.map(name => name.replace(/^node:/, "")));
const declined = [];
const retained = [];
for (const file of ts.sys.readDirectory(path.join(tree, "src"), [".ts"]).sort()) {
    const source = ts.createSourceFile(file, fs.readFileSync(file, "utf8"), ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);
    function visit(node) {
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === "require") {
            const literal = node.arguments.length === 1 && ts.isStringLiteral(node.arguments[0]) ? node.arguments[0] : undefined;
            const at = source.getLineAndCharacterOfPosition(node.getStart(source));
            const site = { file: path.relative(tree, file).replace(/\\/g, "/"), line: at.line + 1,
                column: at.character + 1, specifier: literal?.text ?? null };
            if (literal && builtins.has(literal.text.replace(/^node:/, ""))) {
                const barriers = [];
                for (let parent = node.parent; parent && !ts.isSourceFile(parent); parent = parent.parent) {
                    if (ts.isFunctionLike(parent)) barriers.push("function body: load is lazy");
                    if (ts.isIfStatement(parent) || ts.isConditionalExpression(parent) ||
                        ts.isSwitchStatement(parent) || ts.isIterationStatement(parent, false)) {
                        barriers.push("control flow: load is conditional");
                    }
                    if (ts.isTryStatement(parent)) barriers.push("try/catch: load failure is handled at runtime");
                }
                // There are no eager builtin require calls in the pinned compiler.
                // Refuse new eager sites too, pending a reviewed module-identity policy.
                declined.push({ ...site, replacement: "node:" + literal.text.replace(/^node:/, ""),
                    reason: [...new Set(barriers)].join("; ") || "eager site needs module identity and initialization-order review" });
            }
            else retained.push({ ...site, reason: literal ? "not a Node builtin; retain optional/package loading" : "nonliteral module loading" });
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}
console.log(JSON.stringify({ changed: 0, declined, retained }, null, 2));
