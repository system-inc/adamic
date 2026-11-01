#!/usr/bin/env node
"use strict";
// Row 3's statement, extracted from the tree before 98 and the tree after it, each
// transpiled by stock TypeScript 6.0.3 and run over the same thrown values with the
// flag both false and true, so ||='s short circuit is held too. Both must agree on
// the flag's new value, or on the thrown error's type and exact message.
// With --mutant, 94's null and undefined check is removed from the helper first.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");

const guard = `    if (value === null || value === undefined) {
        throw new TypeError(\`Cannot read properties of \${value === null ? "null" : "undefined"} (reading '\${key}')\`);
    }
`;

function statement(tree) {
    const fileName = path.join(tree, "src/compiler/sys.ts");
    const file = ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
    const found = [];
    (function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === "watchPresentFileSystemEntry") {
            (function inner(child) {
                if (ts.isCatchClause(child)) {
                    for (const statement of child.block.statements) {
                        if (ts.isExpressionStatement(statement) && statement.getText(file).startsWith("hitSystemWatcherLimit ||=")) found.push(statement.getText(file));
                    }
                }
                ts.forEachChild(child, inner);
            })(node);
        }
        ts.forEachChild(node, visit);
    })(file);
    assert.equal(found.length, 1, "row 3 statement moved");
    return found[0];
}

function compile(source, globals) {
    const output = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
    const module = { exports: {} };
    new Function("exports", "module", ...Object.keys(globals), output)(module.exports, module, ...Object.values(globals));
    return module.exports;
}

function getterOnly() {
    const object = {};
    Object.defineProperty(object, "code", { get() { return this === object ? "ENOSPC" : "wrong receiver"; } });
    return object;
}

const thrown = [
    ["thrown null", () => null],
    ["thrown undefined", () => undefined],
    ["a thrown string", () => "ENOSPC"],
    ["a thrown object with code ENOSPC", () => Object.assign(new Error("watch"), { code: "ENOSPC" })],
    ["a thrown object with code EMFILE", () => Object.assign(new Error("watch"), { code: "EMFILE" })],
    ["a thrown object without code", () => new Error("watch")],
    ["a getter-only code", getterOnly],
];

function observe(run) {
    try { return { value: run() }; }
    catch (error) { return { threw: error.constructor.name, message: error.message }; }
}

function main() {
    const args = process.argv.slice(2);
    const mutant = args[0] === "--mutant";
    if (mutant) args.shift();
    assert.equal(args.length, 3, "usage: node witnesses.cjs [--mutant] <tree before 98> <tree after 98> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [before, after, output] = args;
    let helper = fs.readFileSync(path.join(after, "src/compiler/hostErrors.ts"), "utf8");
    if (mutant) {
        assert.equal(helper.split(guard).length - 1, 1, "mutant site drift");
        helper = helper.replace(guard, "");
    }
    const { propertyOf } = compile(helper, {});
    const wrap = text => `export function run(e: unknown, hit: boolean): boolean { let hitSystemWatcherLimit = hit; ${text} return hitSystemWatcherLimit; }`;
    const statements = { before: statement(before), after: statement(after) };
    assert.notEqual(statements.before, statements.after, "row 3 is not adapted in the after tree");
    const original = compile(wrap(statements.before), {}).run;
    const adapted = compile(wrap(statements.after), { propertyOf }).run;
    const results = [];
    for (const [name, make] of thrown) {
        for (const hit of [false, true]) {
            const result = { name, hit, original: observe(() => original(make(), hit)), adapted: observe(() => adapted(make(), hit)) };
            result.agree = JSON.stringify(result.original) === JSON.stringify(result.adapted);
            results.push(result);
        }
    }
    const failed = results.filter(result => !result.agree);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, typescript: ts.version, mutant, statements, cases: results.length, failed: failed.map(result => `${result.name}, hit ${result.hit}`), results }, null, 2) + "\n");
    for (const result of results) console.log(`${result.agree ? "agree" : "DIFFER"}  ${result.name}, hit ${result.hit}: original ${JSON.stringify(result.original)}, adapted ${JSON.stringify(result.adapted)}`);
    console.log(`${results.length - failed.length} of ${results.length} agree`);
    if (failed.length) process.exitCode = 1;
}
main();
