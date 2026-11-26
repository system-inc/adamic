"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { adapt, file, locate, before: statementBefore, after: statementAfter } = require("./adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node test.cjs APPLIED_TREE");
const tree = path.resolve(process.argv[2]);
const ref = "refs/stage3/99-seats-tail24-128-set-visit-discard";
const snapshot = side => execFileSync("git", ["-C", tree, "show", `${ref}-${side}:${file}`], { encoding: "utf8" });
const before = snapshot("before");
const after = snapshot("after");
const root = fs.mkdtempSync(path.join(os.tmpdir(), "set-visit-test-"));
try {
    const archive = execFileSync("git", ["-C", tree, "archive", `${ref}-before`, "src"], { maxBuffer: 128 * 1024 * 1024 });
    execFileSync("tar", ["-x", "-C", root], { input: archive });
    const target = path.join(root, file);
    assert.equal(fs.readFileSync(target, "utf8"), before);
    assert.deepEqual(adapt(root), { files: 1, edits: 1 });
    assert.equal(fs.readFileSync(target, "utf8"), after);
    console.log("PASS clean apply matches after snapshot");

    function evaluate(text) {
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
        const statement = locate(source).statement.getText(source);
        const fn = `(function(state, filePath, invalidateJsFiles, seenFileAndExportsOfFile,
            cancellationToken, host, handleDtsMayChangeOfFileAndExportsOfFile) { ${statement} })`;
        const compiled = ts.transpileModule(fn, { compilerOptions: { target: ts.ScriptTarget.ES2020 } });
        return vm.runInNewContext(compiled.outputText);
    }
    const original = evaluate(before);
    const adapted = evaluate(after);
    function assertCalls(actual, expected) {
        assert.equal(actual.length, expected.length, "worker call count changed");
        for (const [call, args] of expected.entries()) {
            assert.equal(actual[call].length, args.length, `worker argument count changed at call ${call}`);
            for (const [argument, value] of args.entries()) {
                assert.equal(actual[call][argument], value,
                    `worker argument identity changed at call ${call}, argument ${argument}`);
            }
        }
    }

    const paths = ["first", "second", "third"];
    const thrown = new Error("second path");
    for (const scenario of ["ordinary", "throw", "undefined"]) {
        const receiver = scenario === "undefined" ? undefined : new Set(paths);
        const state = { referencedMap: { getKeys(filePath) {
            assert.equal(filePath, "root");
            lookups++;
            return receiver;
        } } };
        const seen = new Set();
        const cancellationToken = {};
        const host = {};
        let lookups;
        function observe(visit) {
            lookups = 0;
            const calls = [];
            let error;
            try {
                visit(state, "root", true, seen, cancellationToken, host, (...args) => {
                    calls.push(args);
                    if (scenario === "throw" && calls.length === 2) throw thrown;
                    return [true, false, undefined][calls.length - 1];
                });
            } catch (failure) {
                error = failure;
            }
            return { calls, error, lookups };
        }
        const expected = observe(original);
        const actual = observe(adapted);
        assert.equal(actual.error, expected.error, "different error or missing optional guard");
        assertCalls(actual.calls, expected.calls);
        assert.equal(actual.lookups, 1);
        assert.equal(expected.lookups, 1);
        const visited = receiver === undefined ? [] : paths.slice(0, scenario === "throw" ? 2 : paths.length);
        assertCalls(expected.calls, visited.map(name => [state, name, true, seen, cancellationToken, host]));
        assert.equal(expected.error, scenario === "throw" ? thrown : undefined);
        console.log(`PASS ${scenario}: identical calls, arguments, and error identity`);
    }

    const mtime = fs.statSync(target).mtimeMs;
    assert.deepEqual(adapt(root), { files: 0, edits: 0 });
    assert.equal(fs.statSync(target).mtimeMs, mtime);
    assert.equal(fs.readFileSync(target, "utf8"), after);
    console.log("PASS second apply does not write");

    fs.writeFileSync(target, after.replace(statementAfter, () => statementBefore));
    assert.equal(fs.readFileSync(target, "utf8"), before);
    console.log("PASS restoration matches before snapshot");

    for (const [label, input] of [
        ["duplicate", before.replace(statementBefore, () => statementBefore + "\r\n    " + statementBefore)],
        ["missing", before.replace(statementBefore, "")],
        ["drift", before.replace(statementBefore, () => statementBefore.replace("getKeys(filePath)", "getKeys(otherPath)"))],
    ]) {
        fs.writeFileSync(target, input);
        assert.throws(() => adapt(root), /missing or duplicate|drift/);
        assert.equal(fs.readFileSync(target, "utf8"), input);
        console.log(`PASS ${label} rejected before any write`);
    }
} finally {
    fs.rmSync(root, { recursive: true, force: true });
}
