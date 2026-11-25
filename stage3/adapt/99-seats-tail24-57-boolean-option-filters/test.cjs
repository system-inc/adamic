"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { adapt, file, locate } = require("./adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node test.cjs APPLIED_TREE");
const tree = path.resolve(process.argv[2]);
const ref = "refs/stage3/99-seats-tail24-57-boolean-option-filters";
const snapshot = side => execFileSync("git", ["-C", tree, "show", `${ref}-${side}:${file}`], { encoding: "utf8" });
const before = snapshot("before");
const after = snapshot("after");
const root = fs.mkdtempSync(path.join(os.tmpdir(), "option-filter-test-"));
try {
    const target = path.join(root, file);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, before);
    assert.deepEqual(adapt(root), { files: 1, edits: 2 });
    assert.equal(fs.readFileSync(target, "utf8"), after);
    console.log("PASS clean apply matches after snapshot");

    const oldSites = locate(before);
    const newSites = locate(after);
    function evaluate(site) {
        const text = ts.transpileModule(`const predicate = ${site.arrow.getText(site.source)};`, {
            compilerOptions: { target: ts.ScriptTarget.ES2020 },
        }).outputText;
        return vm.runInNewContext(text + "\npredicate");
    }

    const tables = [ts.optionDeclarations, ts.optionsForWatch];
    assert(tables.every(Array.isArray), "stock TypeScript must export the real option tables");
    const keys = ["allowConfigDirTemplateSubstitution", "isCommandLineOnly", "isFilePath"];
    const absent = Symbol("absent");
    const choices = [absent, undefined, false, true];
    for (let index = 0; index < tables.length; index++) {
        const original = evaluate(oldSites[index]);
        const adapted = evaluate(newSites[index]);
        const expected = tables[index].filter(original);
        const kept = tables[index].filter(adapted);
        assert(kept.length === expected.length && kept.every((entry, i) => entry === expected[i]),
            "kept entries differ by identity or order");
        console.log(`PASS table ${index + 1}: ${kept.length}/${tables[index].length} entries, identical objects in order`);

        function observe(predicate, values) {
            const reads = [];
            const option = {};
            for (const [i, value] of values.entries()) {
                if (value !== absent) Object.defineProperty(option, keys[i], { value });
            }
            const result = predicate(new Proxy(option, {
                get(target, key, receiver) {
                    reads.push(key);
                    return Reflect.get(target, key, receiver);
                },
            }));
            return { result, reads };
        }
        for (const first of choices) {
            for (const second of choices) {
                for (const third of choices) {
                    const values = [first, second, third];
                    const old = observe(original, values);
                    const current = observe(adapted, values);
                    assert.equal(typeof current.result, "boolean");
                    assert.equal(current.result, Boolean(old.result));
                    assert.deepEqual(current.reads, old.reads, "property read order changed");
                }
            }
        }
        console.log(`PASS table ${index + 1}: boolean result and ordered reads across 64 inputs`);
    }

    const mtime = fs.statSync(target).mtimeMs;
    assert.deepEqual(adapt(root), { files: 0, edits: 0 });
    assert.equal(fs.statSync(target).mtimeMs, mtime);
    assert.equal(fs.readFileSync(target, "utf8"), after);
    console.log("PASS second apply does not write");

    let restored = after;
    for (const site of newSites) restored = restored.replace(site.after, () => site.before);
    fs.writeFileSync(target, restored);
    assert.equal(fs.readFileSync(target, "utf8"), before);
    console.log("PASS restoration matches before snapshot");

    for (const [label, input] of [
        ["duplicate", before + "\r\n" + oldSites[0].before],
        ["missing", before.replace(oldSites[1].before, "")],
        ["drift", before.replace(oldSites[1].before, oldSites[1].before.replace("option =>", "item =>"))],
    ]) {
        fs.writeFileSync(target, input);
        assert.throws(() => adapt(root), /missing or duplicate|drift/);
        assert.equal(fs.readFileSync(target, "utf8"), input);
        console.log(`PASS ${label} rejected before any write`);
    }
} finally {
    fs.rmSync(root, { recursive: true, force: true });
}
