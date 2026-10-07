#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const vm = require("node:vm");
const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { adapt, nodes } = require("./adapt.cjs");
const [before, after, output] = process.argv.slice(2);
assert(before && after && output, "usage: verify.cjs <before> <after> <report>");
const repository = path.resolve(__dirname, "../../..");
const fixtureDirectory = path.join(repository, "stage3/fixtures/host");
function parse(name, text) { return ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true); }
function emit(text) { return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText; }
function file(tree, name) { return parse(name, fs.readFileSync(path.join(tree, "src/compiler", name), "utf8")); }
function one(root, predicate) { const result = nodes(ts, root, predicate); assert.equal(result.length, 1); return result[0]; }
function func(source, name) { return one(source, node => ts.isFunctionDeclaration(node) && node.name?.text === name); }
const helperSource = file(after, "hostErrors.ts");
assert.equal(nodes(ts, helperSource, node => ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) || ts.isNonNullExpression(node) || node.kind === ts.SyntaxKind.AnyKeyword).length, 0);
const helpers = helperSource.statements.map(node => node.getText(helperSource).replace("export ", "")).join("\n");
const context = vm.createContext({});
vm.runInContext(emit(helpers), context);
for (const field of ["code", "message"]) {
    const helper = context[field === "code" ? "errorCode" : "errorMessage"];
    for (const value of [undefined, null, 1, false, "text", () => {}, {}, { [field]: 1 }, { [field]: null }]) assert.equal(helper(value), undefined);
    assert.equal(helper({ [field]: "" }), "");
    assert.equal(helper(Object.create({ [field]: "inherited" })), "inherited");
}
const originalSys = file(before, "sys.ts");
const adaptedSys = file(after, "sys.ts");
const originalRead = emit(func(originalSys, "readFile").getText(originalSys));
const adaptedRead = emit(func(adaptedSys, "readFile").getText(adaptedSys));
assert.equal(adaptedRead, originalRead, "byte assertions must erase identically");
const tracingOriginal = one(func(file(before, "tracing.ts"), "startTracing"), ts.isCatchClause);
const tracingAdapted = one(func(file(after, "tracing.ts"), "startTracing"), ts.isCatchClause);
function traceCatch(node, helperText) {
    return vm.runInContext(emit(`${helperText}\nfunction run(e: unknown) ${node.block.getText()}\nrun;`), vm.createContext({ Error }));
}
const beforeTrace = traceCatch(tracingOriginal, "");
const afterTrace = traceCatch(tracingAdapted, helpers);
for (const e of [new Error("builtin load failed"), new Error(""), { message: "load error" }, { message: "" }, "primitive", 42]) {
    let a, b;
    try { beforeTrace(e); } catch (error) { a = [error.name, error.message]; }
    try { afterTrace(e); } catch (error) { b = [error.name, error.message]; }
    assert.deepEqual(a, b);
}
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "host-errors-verify-"));
const fixtureResults = [];
let helperMutant;
try {
    const status = JSON.parse(fs.readFileSync(path.join(fixtureDirectory, "status.json"), "utf8"));
    for (const row of status) {
        const fixture = path.join(fixtureDirectory, row.file);
        let text = fs.readFileSync(fixture, "utf8");
        const source = parse(row.file, text);
        if (row.file.startsWith("13_")) {
            const method = one(adaptedSys, node => ts.isMethodDeclaration(node) && node.name.getText() === "createDirectory");
            const fixtureMethod = one(source, node => ts.isMethodDeclaration(node) && node.name.getText() === "createDirectory");
            text = helpers + "\n" + text.slice(0, fixtureMethod.getStart(source)) + method.getText(adaptedSys) + text.slice(fixtureMethod.end);
        } else if (/^0[1-4]_/.test(row.file)) {
            const fixtureRead = func(source, "readFile");
            text = text.slice(0, fixtureRead.getStart(source)) + func(adaptedSys, "readFile").getText(adaptedSys) + text.slice(fixtureRead.end);
        }
        const target = path.join(scratch, row.file + ".mjs");
        fs.writeFileSync(target, emit(text));
        function observe() {
            const result = spawnSync(process.execPath, [target], { encoding: "utf8" });
            if (result.error) throw result.error;
            return { stdout: result.stdout, stderr: result.stderr, exit: result.status };
        }
        const observed = observe();
        assert.deepEqual(observed, row.node, row.file);
        fixtureResults.push({ file: row.file, ...observed });
        if (row.file.startsWith("13_")) {
            const mutant = parse("helper-mutant.ts", helpers);
            const code = func(mutant, "errorCode");
            const reads = nodes(ts, code, node => ts.isPropertyAccessExpression(node) && node.name.text === "code");
            assert.equal(reads.length, 2);
            let mutated = helpers;
            // Wrong field in the validated expression; still produces valid JavaScript.
            for (const read of reads.sort((a, b) => b.name.getStart(mutant) - a.name.getStart(mutant))) mutated = mutated.slice(0, read.name.getStart(mutant)) + "message" + mutated.slice(read.name.end);
            const originalCode = emit(text);
            fs.writeFileSync(target, emit(mutated + text.slice(helpers.length)));
            helperMutant = observe();
            assert.notDeepEqual(helperMutant, row.node, "wrong-field mutant survived host fixture 13");
            fs.writeFileSync(target, originalCode);
        }
    }
} finally { fs.rmSync(scratch, { recursive: true }); }
// Demonstrate why injected producers cannot be replaced unconditionally.
const declineProofs = [];
for (const name of ["createGetSourceFile", "createWriteFileMeasuringIO"]) {
    const source = file(before, "program.ts");
    const declaration = func(source, name);
    const scope = vm.createContext({ performance: { mark() {}, measure() {} }, createSourceFile: () => "parsed", writeFileEnsuringDirectories: (...args) => args[3]() });
    const factory = vm.runInContext(emit(declaration.getText(source).replace("export ", "") + `\n${name};`), scope);
    for (const thrown of [null, undefined, { message: 17 }]) {
        let received = "not called";
        let failed = false;
        const thrower = () => { throw thrown; };
        const callback = value => { received = value; };
        try {
            if (name === "createGetSourceFile") factory(thrower, false)("a.ts", 99, callback);
            else factory(thrower, () => {}, () => true)("a.js", "", false, callback);
        } catch (error) { failed = error.name === "TypeError"; }
        assert.equal(failed, thrown === null || thrown === undefined);
        if (!failed) assert.equal(received, 17);
        declineProofs.push({ function: name, thrown: String(thrown), TypeError: failed, received: failed ? null : received });
    }
}
{
    const source = file(before, "commandLineParser.ts");
    const declaration = func(source, "tryReadFile");
    const scope = vm.createContext({ Diagnostics: { Cannot_read_file_0_Colon_1: 1 }, createCompilerDiagnostic: (...args) => args });
    const read = vm.runInContext(emit(declaration.getText(source).replace("export ", "") + "\ntryReadFile;"), scope);
    for (const thrown of [null, undefined, { message: 17 }]) {
        let received;
        let failed = false;
        try { received = read("config.json", () => { throw thrown; }); } catch (error) { failed = error.name === "TypeError"; }
        assert.equal(failed, thrown === null || thrown === undefined);
        if (!failed) assert.equal(received[2], 17);
        declineProofs.push({ function: "tryReadFile", thrown: String(thrown), TypeError: failed, received: failed ? null : received[2] });
    }
}
{
    const source = file(before, "sys.ts");
    const declaration = func(source, "watchPresentFileSystemEntry");
    const worker = () => { throw null; };
    const scope = vm.createContext({ hitSystemWatcherLimit: false, entryKind: 1, FileSystemEntryKind: { Directory: 1 }, fsWatchWithTimestamp: false,
        fsWatchWorker: worker, fileOrDirectory: "watched", recursive: false, inodeWatching: false, callback: () => {},
        sysLog: () => {}, watchPresentFileSystemEntryWithFsWatchFile: () => "polling" });
    const watch = vm.runInContext(emit(declaration.getText(source) + "\nwatchPresentFileSystemEntry;"), scope);
    assert.throws(watch, error => error.name === "TypeError");
    declineProofs.push({ function: "watchPresentFileSystemEntry", thrown: "null", TypeError: true, helperWouldReturnPolling: true });
}
const beforeBytes = new Map(["sys.ts", "tracing.ts", "hostErrors.ts"].map(name => [name, fs.readFileSync(path.join(after, "src/compiler", name))]));
const idempotence = adapt(path.resolve(after), ts);
assert.equal(idempotence.files, 0);
for (const [name, bytes] of beforeBytes) assert(bytes.equals(fs.readFileSync(path.join(after, "src/compiler", name))));
const report = { status: "pass", fixtures: fixtureResults, helperMutant: { caught: true, fixture: "13_createDirectory.a", ...helperMutant }, byteJavaScriptIdentical: true, tracingCases: 6, declineProofs, idempotence };
fs.writeFileSync(output, JSON.stringify(report, null, 2) + "\n");
console.log(JSON.stringify({ status: "pass", fixtures: fixtureResults.length, helperMutantCaught: true, idempotence }));
