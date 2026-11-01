#!/usr/bin/env node
"use strict";
// Runs the adapted tree's actual propertyOf, transpiled by stock TypeScript 6.0.3,
// against a plain property read of the same key on the same value. Each case must
// agree on the value read, or on the thrown error's type and exact message.
// With --mutant, the null and undefined check is removed first, and the run must fail.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");

const guard = `    if (value === null || value === undefined) {
        throw new TypeError(\`Cannot read properties of \${String(value)} (reading '\${key}')\`);
    }
`;

function load(source) {
    const output = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
    const module = { exports: {} };
    new Function("exports", "module", output)(module.exports, module);
    assert.equal(typeof module.exports.propertyOf, "function", "propertyOf is not exported");
    return module.exports.propertyOf;
}

// A plain read written as the call sites write it: `value.key`.
const plain = {
    code: value => value.code,
    stack: value => value.stack,
    message: value => value.message,
    length: value => value.length,
};

function getterOnly() {
    const object = {};
    Object.defineProperty(object, "code", { get() { return this === object ? "ENOSPC" : "wrong receiver"; }, enumerable: true });
    return object;
}

// Each value is built fresh per read, so a getter or a proxy is observed on its own.
const cases = [
    ["null", () => null, "code"],
    ["undefined", () => undefined, "code"],
    ["a string, code", () => "ENOSPC", "code"],
    ["a string, length", () => "ENOSPC", "length"],
    ["a number", () => 28, "code"],
    ["an object with code ENOSPC", () => ({ code: "ENOSPC" }), "code"],
    ["an object without code", () => ({ errno: -28 }), "code"],
    ["an object with a numeric code", () => ({ code: 28 }), "code"],
    ["a getter-only code", getterOnly, "code"],
    ["an inherited code", () => Object.create({ code: "EMFILE" }), "code"],
    ["an Error's message", () => new Error("plugin failed"), "message"],
    ["an Error's stack is a string", () => new Error("plugin failed"), "stack"],
    ["an object with only message", () => ({ message: "only a message" }), "message"],
    ["an object with only message, its stack", () => ({ message: "only a message" }), "stack"],
    ["an object with a numeric message", () => ({ message: 42 }), "message"],
    ["an object with neither, its stack", () => ({ name: "neither" }), "stack"],
    ["an object with neither, its message", () => ({ name: "neither" }), "message"],
    ["a function with code", () => Object.assign(function thrown() {}, { code: "ENOSPC" }), "code"],
    ["a proxy", () => new Proxy({}, { get: (_, key, receiver) => `${String(key)} read through ${typeof receiver}` }), "code"],
    ["a getter that throws", () => ({ get code() { throw new RangeError("getter threw"); } }), "code"],
];

function observe(read) {
    try {
        const value = read();
        // Stack strings name the frames that built them, which differ by caller; compare their first line.
        if (typeof value === "string" && value.includes("\n    at ")) return { type: "string", stackFirstLine: value.split("\n")[0] };
        return { type: typeof value, value };
    }
    catch (error) {
        return { threw: error.constructor.name, message: error.message };
    }
}

function run(propertyOf) {
    // A primitive's prototype getter must see the primitive itself as its receiver.
    const results = cases.map(([name, make, key]) => ({ name, key, plain: observe(() => plain[key](make())), helper: observe(() => propertyOf(make(), key)) }));
    Object.defineProperty(String.prototype, "adamicWitness", { configurable: true, get() { return `${typeof this}:${this}`; } });
    try {
        results.push({ name: "a string's prototype getter sees the primitive", key: "adamicWitness",
            plain: observe(() => "ENOSPC".adamicWitness), helper: observe(() => propertyOf("ENOSPC", "adamicWitness")) });
    }
    finally {
        delete String.prototype.adamicWitness;
    }
    for (const result of results) result.agree = JSON.stringify(result.plain) === JSON.stringify(result.helper);
    return results;
}

function main() {
    const args = process.argv.slice(2);
    const mutant = args[0] === "--mutant";
    if (mutant) args.shift();
    assert.equal(args.length, 2, "usage: node witnesses.cjs [--mutant] <tree> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [tree, output] = args;
    let source = fs.readFileSync(path.join(tree, "src/compiler/hostErrors.ts"), "utf8");
    if (mutant) {
        assert.equal(source.split(guard).length - 1, 1, "mutant site drift");
        source = source.replace(guard, "");
    }
    const results = run(load(source));
    const failed = results.filter(result => !result.agree);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, typescript: ts.version, mutant, cases: results.length, failed: failed.map(result => result.name), results }, null, 2) + "\n");
    for (const result of results) console.log(`${result.agree ? "agree" : "DIFFER"}  ${result.name} (.${result.key}): plain ${JSON.stringify(result.plain)}, helper ${JSON.stringify(result.helper)}`);
    console.log(`${results.length - failed.length} of ${results.length} agree`);
    if (failed.length) process.exitCode = 1;
}
main();
