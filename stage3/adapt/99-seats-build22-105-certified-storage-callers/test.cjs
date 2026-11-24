"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const vm = require("node:vm");
const { createHash } = require("node:crypto");
const { execFileSync, spawnSync } = require("node:child_process");
const { createRequire } = require("node:module");
const path = require("node:path");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");

assert.equal(process.argv.length, 3, "usage: node test.cjs <tree>");
assert.equal(ts.version, "6.0.3");
const tree = process.argv[2];

const snapshot = "refs/stage3/99-seats-build22-105-certified-storage-callers";

function snapshotFile(side, filename) {
    return execFileSync("git", ["-C", tree, "show", `${snapshot}-${side}:${filename}`], { encoding: "utf8" });
}

function sourceFile(file) {
    const filename = `src/compiler/${file}`;
    const source = ts.createSourceFile(filename, snapshotFile("after", filename), ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    return source;
}

function bodyFunction(source, name) {
    const found = nodes(ts, source, node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
    assert.equal(found.length, 1, name);
    return found[0];
}

function implementation(file, name) {
    const source = sourceFile(file);
    const owner = bodyFunction(source, name);
    // Drop the public export on the helper when extracting it as a standalone function.
    const start = owner.getChildren(source).find(child => child.kind === ts.SyntaxKind.FunctionKeyword);
    assert(start, name);
    return source.text.slice(start.getStart(source), owner.end);
}

function transpile(text) {
    const { outputText, diagnostics } = ts.transpileModule(text, {
        compilerOptions: { target: ts.ScriptTarget.ES2024 },
        reportDiagnostics: true,
    });
    assert.equal(diagnostics.length, 0);
    return outputText;
}

function load(text, name, bindings) {
    return Function(...Object.keys(bindings), transpile(text) + "\nreturn " + name)(...Object.values(bindings));
}

const adapter = path.join(__dirname, "adapt.cjs");
const sortBefore = "const fileIds = arrayFrom(set.keys(), toFileId).sort(compareValues);";
assert.equal(fs.readFileSync(adapter, "utf8").split(sortBefore).length, 2, "adapter's pinned stock sort");
const builder = sourceFile("builder.ts");
const sortOwner = bodyFunction(builder, "toFileIdListId");
const keyStatement = sortOwner.body.statements.find(stmt => ts.isVariableStatement(stmt)
    && stmt.declarationList.declarations.some(decl => decl.name.getText(builder) === "key"
        && decl.initializer?.getText(builder) === "fileIds.join()"));
assert(keyStatement, "fileIds.join boundary");
const sortAfter = builder.text.slice(sortOwner.body.statements[0].getStart(builder), keyStatement.getStart(builder));
const corePublic = sourceFile("corePublic.ts");
const comparison = corePublic.statements.find(node => ts.isEnumDeclaration(node) && node.name.text === "Comparison");
assert(comparison);
const helpers = [
    comparison.getText(corePublic).replace(/^export /, ""),
    implementation("core.ts", "compareComparableValues"),
    implementation("core.ts", "compareValues"),
    implementation("core.ts", "arrayFrom"),
].join("\n");
const debug = { fail: message => assert.fail(message) };
const wrap = statements => helpers + "\nfunction sortIds(set, toFileId) {\n" + statements + "\nreturn fileIds;\n}";
const stockSort = load(wrap(sortBefore), "sortIds", { Debug: debug });
const headSort = load(wrap(sortAfter), "sortIds", { Debug: debug });

function sortFixture(sort, keys) {
    const reserved = new Map([["a", 2], ["b", 11]]);
    let visited = "";
    let next = 20;
    const ids = sort(new Set(keys), key => {
        visited += key;
        if (!reserved.has(key)) reserved.set(key, next++);
        return reserved.get(key);
    }).join();
    return { ids, visited, next };
}

for (const [keys, expected] of [
    [["b", "c", "a", "b"], { ids: "2,11,20", visited: "bca", next: 21 }],
    [[], { ids: "", visited: "", next: 20 }],
]) {
    const before = sortFixture(stockSort, keys);
    const after = sortFixture(headSort, keys);
    assert.equal(after.ids, expected.ids, `105 sort expected ${expected.ids} got ${after.ids}`);
    assert.deepEqual(before, expected, "stock sort");
    assert.deepEqual(after, before, "105 sort and mapper order");
}
console.log("105 sort: 2,11,20; mapper bca; next 21; empty set agrees");

function outcome(run) {
    try {
        const value = run();
        return { value: value instanceof Map ? [...value] : value };
    } catch (error) {
        return { throws: { name: error.name, message: error.message } };
    }
}

function compareOutcome(before, after, label) {
    assert.deepEqual(after, before, `${label}: stock ${JSON.stringify(before)} | head ${JSON.stringify(after)}`);
}

const toFilePath = id => ["A", "B"][id - 1];
const arrayToMap = load(implementation("core.ts", "arrayToMap"), "arrayToMap", { identity: value => value });
const adapted = load(implementation("builder.ts", "toPerFileEmitDiagnostics"), "toPerFileEmitDiagnostics", { toFilePath });
const old = [{ message: "old" }];
const middle = [{ message: "middle" }];
const latest = [{ message: "new" }];
const cases = [
    ["undefined", undefined, undefined],
    ["null", null, null],
    ["empty", [], []],
    ["single", [[1, old]], [["A", old]]],
    ["duplicate", [[1, old], [2, middle], [1, latest]], [["A", latest], ["B", middle]]],
    ["missing path", [[1, old], [3, latest]], [["A", old]]],
    ["empty object", {}, []],
    ["array-like object", { 0: [1, old], length: 1 }, [["A", old]]],
    ["true", true, []],
    ["one", 1, []],
];
function checkTable() {
    for (const [name, input, expected] of cases) {
        const before = outcome(() => input && arrayToMap(input, row => toFilePath(row[0]), row => row[1]));
        const after = outcome(() => adapted(input));
        assert.deepEqual(before, { value: expected }, "stock " + name);
        compareOutcome(before, after, "106 table " + name);
        console.log(name + ": agrees");
    }
}

function checkReader() {
    const bundle = require.resolve(process.env.CENSUS_TYPESCRIPT || "typescript");
    const text = fs.readFileSync(bundle, "utf8");
    const source = ts.createSourceFile(bundle, text, ts.ScriptTarget.Latest, true);
    const original = bodyFunction(source, "toPerFileEmitDiagnostics");
    const replacement = transpile(implementation("builder.ts", "toPerFileEmitDiagnostics"));
    const mod = { exports: {} };
    vm.compileFunction(text.slice(0, original.getStart(source)) + replacement + text.slice(original.end),
        ["exports", "require", "module", "__filename", "__dirname"], { filename: bundle })(
        mod.exports, createRequire(bundle), mod, bundle, path.dirname(bundle));

    const options = { incremental: true, tsBuildInfoFile: path.join(os.tmpdir(), "fixture.tsbuildinfo") };
    const diagnostic = messageText => [{ category: 1, code: 2322, file: false, messageText }];
    const first = diagnostic("first");
    const second = diagnostic("second");
    function read(api, input) {
        const info = { version: ts.version, fileNames: ["./one.ts", "./two.ts"], fileInfos: ["v1", "v2"],
            options, emitDiagnosticsPerFile: input };
        const host = { getCurrentDirectory: () => os.tmpdir(), useCaseSensitiveFileNames: () => true,
            readFile: () => JSON.stringify(info) };
        const builder = api.readBuilderProgram(options, host);
        assert(builder, "reader rejected probe build info");
        return builder.state.emitDiagnosticsPerFile;
    }
    for (const [name, input] of [
        ["absent", undefined], ["null", null], ["false", false], ["zero", 0], ["empty string", ""],
        ["missing path", [[3, first]]],
        ["two files", [[1, first], [2, second]]],
        ["duplicate", [[1, first], [2, second], [1, diagnostic("last")]]],
        ["empty object", {}], ["array-like object", { 0: [1, first], length: 1 }],
    ]) {
        const before = outcome(() => read(ts, input));
        const after = outcome(() => read(mod.exports, input));
        assert(!before.throws, "stock reader " + name);
        compareOutcome(before, after, "106 reader " + name);
        console.log("reader " + name + ": agrees");
    }
}

function replaceOnce(text, before, after) {
    assert.equal(text.split(before).length, 2, before);
    return text.replace(before, after);
}

function checkAdapter() {
    const stock = snapshotFile("before", "src/compiler/builder.ts");
    const archive = execFileSync("git", ["-C", tree, "archive", `${snapshot}-after`, "src"], {
        maxBuffer: 128 * 1024 * 1024,
    });
    const original106 = "return diagnostics && arrayToMap(diagnostics, value => toFilePath(value[0]), value => value[1]);";
    const rewritten = replaceOnce(stock, original106, original106.replace("value => value[1]", "value => value[1].slice()"));
    const failures = [];
    for (const [label, text, error] of [
        ["clean reapply", builder.text, null],
        ["105 indent", replaceOnce(builder.text, "\r\n        const fileIds: IncrementalBuildInfoFileId[] = [];",
            "\r\n      const fileIds: IncrementalBuildInfoFileId[] = [];"), "unexpected indentation"],
        ["106 indent", replaceOnce(builder.text, "\r\n        if (!diagnostics) return diagnostics;",
            "\r\n      if (!diagnostics) return diagnostics;"), "unexpected indentation"],
        ["unrecognized 106 call", rewritten, "unexpected snapshot in toPerFileEmitDiagnostics"],
    ]) {
        const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "certified-storage-caller-"));
        try {
            execFileSync("tar", ["-x", "-C", scratch], { input: archive });
            const file = path.join(scratch, "src/compiler/builder.ts");
            fs.writeFileSync(file, text);
            const digest = () => createHash("sha256").update(fs.readFileSync(file)).digest("hex");
            const before = digest();
            const result = spawnSync(process.execPath, [adapter, scratch], { encoding: "utf8", env: process.env });
            assert.ifError(result.error);
            assert.equal(result.signal, null, label);
            assert.equal(digest(), before, label + ": adapter wrote builder.ts");
            const details = `${label}: exit ${result.status}\n${result.stdout}${result.stderr}`;
            if (error) {
                assert.notEqual(result.status, 0, details);
                assert(result.stderr.includes(error), details);
            } else {
                assert.equal(result.status, 0, details);
                assert.match(result.stdout, /^105: toFileIdListId already adapted$/m);
                assert.match(result.stdout, /^106: toPerFileEmitDiagnostics already adapted$/m);
            }
            console.log(label + ": " + (error ? "rejected" : "passed") + ", bytes identical");
        } catch (failure) {
            failures.push(failure);
            console.error(failure.stack);
        } finally {
            fs.rmSync(scratch, { recursive: true, force: true });
        }
    }
    assert.equal(failures.length, 0, "adapter cases");
}

// Run both independently so a bad walk reports its table and reader failure.
const parityErrors = [];
for (const check of [checkTable, checkReader]) {
    try {
        check();
    } catch (error) {
        parityErrors.push(error);
        console.error(error.stack);
    }
}
assert.equal(parityErrors.length, 0, "106 parity");
checkAdapter();
