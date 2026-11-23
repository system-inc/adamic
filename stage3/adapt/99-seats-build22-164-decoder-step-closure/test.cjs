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
const ref = "refs/stage3/99-seats-build22-164-decoder-step-closure-before";
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
    const exports = {};
    vm.runInNewContext(outputText, {
        exports,
        require(name) {
            if (name === "./_namespaces/ts.js") return ts;
            assert.equal(name, "./_namespaces/ts.performance.js");
            return { nullTimer: { enter() {}, exit() {} } };
        },
    }, { filename: "sourcemap.js" });
    return exports;
}

function trace(decode, mappings) {
    const decoder = decode(mappings);
    const rows = [];
    const record = (row, value) => rows.push({ row, value: structuredClone(value) });
    const getters = label => {
        for (const property of ["pos", "error", "state"]) record(`${label}.${property}`, decoder[property]);
    };
    record("members", Object.getOwnPropertyNames(decoder));
    record("Reflect.ownKeys", Reflect.ownKeys(decoder).map(String));
    record("selfIterator", decoder[Symbol.iterator]() === decoder);
    assert.equal(rows.at(-1).value, true);
    getters("initial");
    for (let index = 0; index <= mappings.length; index++) {
        const result = decoder.next();
        record(`next[${index}]`, result);
        if (result.done) {
            getters("done");
            record("nextAfterDone", decoder.next());
            getters("afterDone");
            return rows;
        }
    }
    throw new Error("decoder did not finish");
}

function consumers(api) {
    const host = { getCurrentDirectory: () => "/", getCanonicalFileName: file => file, getSourceFileLike: () => undefined };
    const rows = [];
    const record = (row, read) => {
        try {
            rows.push({ row, value: structuredClone(read()) });
        } catch (error) {
            rows.push({ row, value: error.message });
        }
    };
    for (const [name, mappings, start, end] of [
        ["bounded", "A,C;A,E", { line: 0, character: 1 }, { line: 1, character: 0 }],
        ["unbounded", "A,C;A,E", undefined, undefined],
        ["early break", "A,C,E,?AAA", { line: 0, character: 1 }, { line: 0, character: 1 }],
    ]) {
        record(`appendSourceMap.${name}`, () => {
            const generator = api.createSourceMapGenerator(host, "x.js", "", "/", {});
            const reads = [];
            const input = {
                length: mappings.length,
                charCodeAt(index) {
                    reads.push(index);
                    if (mappings[index] === "?") throw new Error("read after early break");
                    return mappings.charCodeAt(index);
                },
            };
            generator.appendSourceMap(0, 0, { version: 3, file: "x.js", sources: [], names: [], mappings: input }, "/x.js.map", start, end);
            return { mappings: generator.toJSON().mappings, reads };
        });
    }
    record("createDocumentPositionMapper.errors", () => {
        const errors = [];
        const mapper = api.createDocumentPositionMapper({ ...host, log: message => errors.push(message) }, {
            version: 3, file: "x.js", sources: [], names: [], mappings: "A,?AAA",
        }, "/x.js.map");
        const position = { fileName: "/x.js", pos: 0 };
        assert.equal(mapper.getSourcePosition(position), position);
        assert.equal(mapper.getSourcePosition(position), position);
        return errors;
    });
    return rows;
}

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-164-"));
try {
    const filename = path.join(scratch, relative);
    fs.mkdirSync(path.dirname(filename), { recursive: true });
    fs.writeFileSync(filename, before);
    const adapter = path.join(__dirname, "adapt.cjs");
    const run = () => execFileSync(process.execPath, [adapter, scratch], { encoding: "utf8" }).trim();
    const applied = run();
    console.log(applied);
    assert.equal(JSON.parse(applied).edits, 2);
    const after = fs.readFileSync(filename, "utf8");
    assert.notEqual(after, before);
    const original = load(before);
    const adapted = load(after);
    const results = [];
    const failures = [];
    function compare(name, base, head) {
        const start = failures.length;
        for (let index = 0; index < Math.max(base.length, head.length); index++) {
            try {
                assert.deepEqual(head[index], base[index]);
            } catch {
                const failure = { name, row: base[index]?.row ?? head[index]?.row, base: base[index]?.value, head: head[index]?.value };
                failures.push(failure);
                console.error("FAIL " + JSON.stringify(failure));
            }
        }
        if (failures.length === start) console.log(`PASS ${name} (${base.length} rows)`);
    }

    for (const [name, mappings, yields, error] of [
        ["multi-line", "AAAA,CAAC;AACA,EAAE", 4, undefined],
        ["empty", "", 0, undefined],
        ["malformed", "AAAA,?AAA", 1, "Invalid character in VLQ"],
        ["trailing separator", "AAAA,;", 1, undefined],
    ]) {
        const base = trace(original.decodeMappings, mappings);
        const head = trace(adapted.decodeMappings, mappings);
        assert.equal(base.filter(row => row.row.startsWith("next[") && !row.value.done).length, yields);
        assert.equal(base.find(row => row.row === "done.error").value, error);
        results.push({ name, mappings, base, head });
        compare(name, base, head);
    }
    const baseConsumers = consumers(original);
    const headConsumers = consumers(adapted);
    for (const row of baseConsumers.slice(0, 3)) {
        assert.equal(typeof row.value.mappings, "string", `base consumer failed: ${JSON.stringify(row)}`);
        assert.ok(Array.isArray(row.value.reads));
    }
    assert.deepEqual(baseConsumers[3].value, ["Encountered error while decoding sourcemap: Invalid character in VLQ"]);
    results.push({ name: "real consumers", base: baseConsumers, head: headConsumers });
    compare("real consumers", baseConsumers, headConsumers);

    fs.writeFileSync(output, JSON.stringify({ results, failures }, (_, value) => value === undefined ? { undefined: true } : value, 2) + "\n");
    if (failures.length) process.exitCode = 1;
    else {
        const repeated = JSON.parse(run());
        assert.equal(repeated.edits, 0);
        assert.equal(fs.readFileSync(filename, "utf8"), after);
        console.log("PASS second adapter call preserved bytes");
    }
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
