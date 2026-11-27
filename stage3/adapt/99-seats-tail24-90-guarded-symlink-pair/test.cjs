"use strict";
const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { adapt, locate, relative } = require("./adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node test.cjs APPLIED_TREE");
assert.equal(ts.version, "6.0.3");
const tree = path.resolve(process.argv[2]);
const ref = "refs/stage3/99-seats-tail24-90-guarded-symlink-pair";
const read = side => execFileSync("git", ["-C", tree, "show", `${ref}-${side}:${relative}`], {encoding: "utf8"});
const before = read("before");
const after = read("after");
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "guarded-symlink-test-"));
try {
    const file = path.join(scratch, relative);
    fs.mkdirSync(path.dirname(file), {recursive: true});
    fs.writeFileSync(file, before);
    assert.deepEqual(adapt(scratch), {files: 1, edits: 1});
    assert.equal(fs.readFileSync(file, "utf8"), after, "Adapter output differs from the after snapshot");
    console.log("PASS adapter replay matches its after snapshot");
    const mtime = fs.statSync(file, {bigint: true}).mtimeNs;
    assert.deepEqual(adapt(scratch), {files: 0, edits: 0});
    assert.equal(fs.statSync(file, {bigint: true}).mtimeNs, mtime);
    assert.equal(fs.readFileSync(file, "utf8"), after);
    console.log("PASS second application leaves bytes and mtime unchanged");
    const oldOwner = locate(before), newOwner = locate(after);
    assert.equal(after.slice(0, newOwner.start) + before.slice(oldOwner.start, oldOwner.end) + after.slice(newOwner.end), before);
    console.log("PASS restoring the original owner reproduces the before snapshot");
    const examples = [undefined, ["a/", "b/"], ["", "b/"], ["a/", ""], ["", ""]];
    for (const pair of examples) {
        const original = evaluate(before, oldOwner, pair);
        const current = evaluate(after, newOwner, pair);
        assert.deepEqual(current, original, `calls or return differ for ${JSON.stringify(pair)}`);
        const cacheCalls = current.calls.filter(call => call[0].startsWith("setSymlinked"));
        assert.deepEqual(cacheCalls.map(call => call[0]), pair?.[0] && pair?.[1]
            ? ["setSymlinkedFile", "setSymlinkedDirectory"] : ["setSymlinkedFile"]);
        assert(current.calls.findIndex(call => call[0] === "setSymlinkedFile") <
            current.calls.findIndex(call => call[0] === "guessDirectorySymlink"));
        console.log(`PASS ${JSON.stringify(pair)}: ${JSON.stringify(current.calls)}; return ${current.result}`);
    }
} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}

function evaluate(text, owner, pair) {
    const body = ts.transpileModule(text.slice(owner.start, owner.end), {
        compilerOptions: {target: ts.ScriptTarget.ES2020},
    }).outputText;
    const calls = [];
    const record = (name, result) => (...args) => { calls.push([name, ...args]); return result; };
    const getCanonicalFileName = value => value;
    const cwd = "/root";
    const context = vm.createContext({
        cwd, getCanonicalFileName, emptyArray: [],
        guessDirectorySymlink: record("guessDirectorySymlink", pair),
        toPath: (value, ...args) => record("toPath", `path:${value}`)(value, ...args),
        ensureTrailingDirectorySeparator: value => record("ensureTrailingDirectorySeparator", value.endsWith("/") ? value : value + "/")(value),
        cache: {
            setSymlinkedFile: record("setSymlinkedFile"),
            setSymlinkedDirectory: record("setSymlinkedDirectory"),
        },
        resolution: {resolvedFileName: "/real/a.ts", originalPath: "/link/a.ts"},
    });
    const result = vm.runInContext(body + "\nprocessResolution(cache, resolution);", context);
    return {calls: JSON.parse(JSON.stringify(calls, (_, value) => value === getCanonicalFileName ? "<canonicalizer>" : value)), result};
}
