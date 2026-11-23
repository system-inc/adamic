#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 4, "usage: node test.cjs <tree before this adaptation> <output.json>");
const [, , tree, output] = process.argv;
const relative = "src/compiler/sourcemap.ts";
const ref = "refs/stage3/99-seats-build22-156-explicit-undefined-completion-before";
const lookup = spawnSync("git", ["-C", tree, "rev-parse", "--verify", "--quiet", ref], { encoding: "utf8" });
assert.ifError(lookup.error);
const before = lookup.status === 0
    ? execFileSync("git", ["-C", tree, "show", `${ref}:${relative}`], { encoding: "utf8" })
    : fs.readFileSync(path.join(tree, relative), "utf8");

function load(text) {
    const { outputText, diagnostics } = ts.transpileModule(text, {
        fileName: relative,
        compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS },
        reportDiagnostics: true,
    });
    assert.deepEqual(diagnostics, []);
    const CharacterCodes = Object.fromEntries(Object.entries({
        comma: ",", semicolon: ";", A: "A", Z: "Z", a: "a", z: "z",
        _0: "0", _9: "9", plus: "+", slash: "/",
    }).map(([name, char]) => [name, char.charCodeAt(0)]));
    const exports = {};
    vm.runInNewContext(outputText, {
        exports,
        require(name) {
            if (name === "./_namespaces/ts.js") return { CharacterCodes };
            assert.equal(name, "./_namespaces/ts.performance.js");
            return {};
        },
    }, { filename: "sourcemap.js" });
    return exports.decodeMappings;
}

const descriptor = { value: undefined, writable: true, enumerable: true, configurable: true };
const keys = ["value", "done"];
const expected = {
    "descriptor.value": descriptor.value,
    "descriptor.writable": descriptor.writable,
    "descriptor.enumerable": descriptor.enumerable,
    "descriptor.configurable": descriptor.configurable,
    "hasOwnProperty": true,
    "Reflect.ownKeys": keys,
    "Object.keys": keys,
    done: true,
};

function completion(result) {
    const property = Object.getOwnPropertyDescriptor(result, "value");
    return {
        "descriptor.value": property?.value,
        "descriptor.writable": property?.writable,
        "descriptor.enumerable": property?.enumerable,
        "descriptor.configurable": property?.configurable,
        hasOwnProperty: Object.prototype.hasOwnProperty.call(result, "value"),
        "Reflect.ownKeys": Reflect.ownKeys(result).map(key => typeof key === "symbol" ? String(key) : key),
        "Object.keys": Object.keys(result),
        done: result.done,
    };
}

function trace(decode, mappings) {
    const decoder = decode(mappings);
    for (let index = 0; index <= mappings.length; index++) {
        const result = decoder.next();
        if (result.done) return { done: completion(result), afterDone: completion(decoder.next()) };
    }
    throw new Error("decoder did not finish");
}

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-156-"));
try {
    const filename = path.join(scratch, relative);
    fs.mkdirSync(path.dirname(filename), { recursive: true });
    fs.writeFileSync(filename, before);
    const adapter = path.join(__dirname, "adapt.cjs");
    const run = () => {
        const message = execFileSync(process.execPath, [adapter, scratch], { encoding: "utf8" }).trim();
        console.log(message);
        return JSON.parse(message);
    };
    assert.equal(run().edits, 2, "expected pre-156 input");
    const after = fs.readFileSync(filename, "utf8");
    assert.notEqual(after, before);
    const originalDecoder = load(before);
    const adaptedDecoder = load(after);
    const results = [];
    const failures = [];
    for (const [name, mappings] of [
        ["empty", ""],
        ["multi-line", "AAAA,CAAC;AACA,EAAE"],
        ["malformed", "AAAA,?AAA"],
        ["trailing separator", "AAAA,;"],
    ]) {
        const base = trace(originalDecoder, mappings);
        const head = trace(adaptedDecoder, mappings);
        results.push({ name, mappings, base, head });
        const start = failures.length;
        for (const phase of ["done", "afterDone"]) {
            for (const [row, wanted] of Object.entries(expected)) {
                try {
                    assert.deepEqual(head[phase][row], base[phase][row]);
                    assert.deepEqual(head[phase][row], wanted);
                } catch {
                    const failure = { name, row: `${phase}.${row}`, base: base[phase][row], head: head[phase][row], expected: wanted };
                    failures.push(failure);
                    console.error("FAIL " + JSON.stringify(failure));
                }
            }
        }
        if (failures.length === start) console.log(`PASS ${name} (16 rows)`);
    }
    fs.writeFileSync(output, JSON.stringify({ results, failures }, (_, value) => value === undefined ? { undefined: true } : value, 2) + "\n");
    if (failures.length) process.exitCode = 1;
    else {
        assert.equal(run().edits, 0);
        assert.equal(fs.readFileSync(filename, "utf8"), after);
        console.log("PASS second adapter call preserved bytes");
    }
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
