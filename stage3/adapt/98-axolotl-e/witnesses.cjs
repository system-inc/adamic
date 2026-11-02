#!/usr/bin/env node
"use strict";
// project.ts's two plugin-error readers, extracted from the tree before 98-e and the tree after
// it, each transpiled by stock TypeScript 6.0.3 and run over the same failed results. Both must
// agree on the logged text, or on the thrown error's type and exact message.
// With --mutant, 94's null and undefined check is removed from the helper first.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");

const guard = `    if (value === null || value === undefined) {
        throw new TypeError(\`Cannot read properties of \${value === null ? "null" : "undefined"} (reading '\${key}')\`);
    }
`;
const readers = ["importServicePluginSync", "importServicePluginAsync"];

function initializers(tree) {
    const fileName = path.join(tree, "src/server/project.ts");
    const file = ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
    return readers.map(name => {
        const found = [];
        (function visit(node) {
            if (ts.isMethodDeclaration(node) && node.name.getText(file) === name) {
                (function inner(child) {
                    if (ts.isVariableDeclaration(child) && child.name.getText(file) === "err") found.push(child.initializer.getText(file));
                    ts.forEachChild(child, inner);
                })(node);
            }
            ts.forEachChild(node, visit);
        })(file);
        assert.equal(found.length, 1, `${name}'s reader moved`);
        return found[0];
    });
}

function compile(source, globals) {
    const output = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
    const module = { exports: {} };
    new Function("exports", "module", ...Object.keys(globals), output)(module.exports, module, ...Object.values(globals));
    return module.exports;
}

function withStack(message) {
    const error = new Error(message);
    error.stack = `Error: ${message}\n    at plugin (plugin.js:1:1)`;
    return error;
}

const failures = [
    ["error null", () => null],
    ["error undefined", () => undefined],
    ["a string", () => "plugin failed"],
    ["an Error with stack", () => withStack("plugin failed")],
    ["an object with only message", () => ({ message: "only a message" })],
    ["an object with neither", () => ({ name: "neither" })],
    ["an empty stack falls to message", () => ({ stack: "", message: "the message" })],
    ["a number", () => 42],
];

function observe(run) {
    try { return { value: run() }; }
    catch (error) { return { threw: error.constructor.name, message: error.message }; }
}

function main() {
    const args = process.argv.slice(2);
    const mutant = args[0] === "--mutant";
    if (mutant) args.shift();
    assert.equal(args.length, 3, "usage: node witnesses.cjs [--mutant] <tree before 98-e> <tree after 98-e> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [before, after, output] = args;
    let helper = fs.readFileSync(path.join(after, "src/compiler/hostErrors.ts"), "utf8");
    if (mutant) {
        assert.equal(helper.split(guard).length - 1, 1, "mutant site drift");
        helper = helper.replace(guard, "");
    }
    const { propertyOf } = compile(helper, {});
    const wrap = text => `export function run(result: any): unknown { return ${text}; }`;
    const texts = { before: initializers(before), after: initializers(after) };
    const results = [];
    for (const [index, name] of readers.entries()) {
        assert.notEqual(texts.before[index], texts.after[index], `${name} is not adapted in the after tree`);
        const original = compile(wrap(texts.before[index]), {}).run;
        const adapted = compile(wrap(texts.after[index]), { propertyOf }).run;
        for (const [label, make] of failures) {
            const result = { reader: name, name: label,
                original: observe(() => original({ module: undefined, error: make() })),
                adapted: observe(() => adapted({ module: undefined, error: make() })) };
            result.agree = JSON.stringify(result.original) === JSON.stringify(result.adapted);
            results.push(result);
        }
    }
    const failed = results.filter(result => !result.agree);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, typescript: ts.version, mutant, readers: texts, cases: results.length, failed: failed.map(result => `${result.reader}: ${result.name}`), results }, null, 2) + "\n");
    for (const result of results) console.log(`${result.agree ? "agree" : "DIFFER"}  ${result.reader}, ${result.name}: original ${JSON.stringify(result.original)}, adapted ${JSON.stringify(result.adapted)}`);
    console.log(`${results.length - failed.length} of ${results.length} agree`);
    if (failed.length) process.exitCode = 1;
}
main();
