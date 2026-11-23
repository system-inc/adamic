"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const crypto = require("node:crypto");
const fs = require("node:fs");
const { createRequire, wrap } = require("node:module");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");

const packages = path.resolve(__dirname, "../../api/node_modules");
const bundle = path.join(packages, "typescript/lib/typescript.js");
const ts = require(bundle);
const snapshot = "refs/stage3/99-seats-build22-142-pragma-arg-counts-before";
const files = ["src/compiler/parser.ts", "src/compiler/types.ts"];
const pragmaNames = [
    "reference", "amd-dependency", "amd-module", "ts-check", "ts-nocheck",
    "jsx", "jsxfrag", "jsximportsource", "jsxruntime",
];
const texts = [
    undefined, "", "   ", "one", "  one  ", "one\ttwo", "one\ntwo",
    "one two", "one two three", "one two three four five six",
    "one two three four five six seven",
];

function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file);
    return source;
}

function find(source, kind, name) {
    const matches = [];
    function visit(node) {
        if (kind(node) && node.name?.text === name) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(matches.length, 1, name);
    return matches[0];
}

function emit(text) {
    const result = ts.transpileModule(text, {
        compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS },
        reportDiagnostics: true,
    });
    assert.deepEqual(result.diagnostics, []);
    return result.outputText;
}

function fragments(tree) {
    const [parser, types] = files.map(file => parse(file, fs.readFileSync(path.join(tree, file), "utf8")));
    return {
        helper: emit(find(parser, ts.isFunctionDeclaration, "getNamedPragmaArguments").getText()),
        table: emit("const " + find(types, ts.isVariableDeclaration, "commentPragmas").getText() + ";"),
        flags: emit(find(types, ts.isEnumDeclaration, "PragmaKindFlags").getText().replace(/^export /, "")),
    };
}

function evaluate({ helper, table, flags }) {
    const Debug = { fail(message) { throw new Error(message); } };
    return vm.runInNewContext(
        flags + table + helper + "\n({ helper: getNamedPragmaArguments, table: commentPragmas });",
        { Debug },
    );
}

function outcome(helper, pragma, text) {
    let result;
    try {
        result = helper(pragma, text);
    }
    catch (error) {
        return ["throw", error.message];
    }
    if (result === "fail") return ["fail"];
    assert(result && typeof result === "object" && !Array.isArray(result));
    return ["map", Object.entries(result)];
}

function checkHelper(original, adapted) {
    const before = evaluate(original);
    const after = evaluate(adapted);
    for (const side of [before, after]) assert.deepEqual(Object.keys(side.table), pragmaNames);
    for (const name of [...pragmaNames, "(absent)"]) {
        for (const text of texts) {
            const label = `helper ${name} text=${text === undefined ? "undefined" : JSON.stringify(text)}`;
            const expected = outcome(before.helper, before.table[name] ?? {}, text);
            const actual = outcome(after.helper, after.table[name] ?? {}, text);
            assert.deepEqual(actual, expected, label);
            console.log(`PASS ${label}`);
        }
    }
}

function replaceOnce(text, original, adapted) {
    const start = text.indexOf(original);
    assert(start >= 0 && text.indexOf(original, start + 1) === -1, "expected one bundle text match");
    return text.slice(0, start) + adapted.trimEnd() + text.slice(start + original.length);
}

function checkParser(adapted) {
    const text = fs.readFileSync(bundle, "utf8");
    const source = parse(bundle, text);
    const oldHelper = find(source, ts.isFunctionDeclaration, "getNamedPragmaArguments").getText();
    const oldTable = find(source, ts.isVariableDeclaration, "commentPragmas").parent.parent.getText();
    const patched = replaceOnce(replaceOnce(text, oldTable, adapted.table), oldHelper, adapted.helper);
    // In the bundle the table uses inlined flags. The adapted fragment can use
    // the bundle's own PragmaKindFlags enum, which is initialized before it.
    const patchedModule = { exports: {} };
    vm.runInThisContext(wrap(patched), { filename: bundle })(
        patchedModule.exports, createRequire(bundle), patchedModule, bundle, path.dirname(bundle),
    );
    const cases = [
        ["/** @jsx h */", "jsx"], ["/** @jsxFrag F */", "jsxfrag"],
        ["/** @jsxImportSource preact */", "jsximportsource"], ["/** @jsxRuntime classic */", "jsxruntime"],
        ["/** @jsx */", "jsx"], ["/** @jsx a b */", "jsx"],
        ["// @ts-check", "ts-check"], ["// @ts-nocheck", "ts-nocheck"],
        ['/// <amd-dependency path="p" name="n" />', "amd-dependency"],
        ['/// <reference path="a.ts" />', "reference"],
    ];
    const parsePragmas = (compiler, comment) => compiler.createSourceFile(
        "probe.tsx", "\n\n" + comment + "\nconst view = <div />;", ts.ScriptTarget.Latest,
    ).pragmas;
    for (const [comment, name] of cases) {
        const original = parsePragmas(ts, comment);
        const adapted = parsePragmas(patchedModule.exports, comment);
        assert.deepEqual([...original.keys()], [name], comment);
        assert.deepEqual([...adapted], [...original], `parser ${comment}`);
        console.log(`PASS parser ${comment}`);
    }
}

function runAdapter(adapter, tree) {
    return JSON.parse(execFileSync(process.execPath, [adapter, tree], {
        encoding: "utf8",
        env: { ...process.env, NODE_PATH: packages },
    }));
}

function hashes(tree) {
    return files.map(file => crypto.createHash("sha256").update(fs.readFileSync(path.join(tree, file))).digest("hex"));
}

function main() {
    const [, , input, adapter = path.join(__dirname, "adapt.cjs")] = process.argv;
    assert(input && process.argv.length <= 4, "usage: node test.cjs <adapted TypeScript tree> [adapter.cjs]");
    assert.equal(ts.version, "6.0.3");
    const tree = fs.mkdtempSync(path.join(os.tmpdir(), "pragma-arg-counts-"));
    try {
        fs.mkdirSync(path.join(tree, "src/compiler"), { recursive: true });
        for (const file of files) {
            fs.writeFileSync(path.join(tree, file), execFileSync("git", ["-C", input, "show", `${snapshot}:${file}`]));
        }
        const original = fragments(tree);
        const first = runAdapter(adapter, tree);
        assert.deepEqual(first, { files: 2, edits: 16 });
        console.log("PASS clean apply " + JSON.stringify(first));

        // Drive the first output before retrying. A planted runtime change can
        // also invalidate the adapter's post-apply fingerprints.
        const adapted = fragments(tree);
        checkHelper(original, adapted);

        const before = hashes(tree);
        const second = runAdapter(adapter, tree);
        assert.deepEqual(second, { files: 0, edits: 0 });
        assert.deepEqual(hashes(tree), before);
        console.log("PASS second apply " + JSON.stringify(second) + " and unchanged source hashes");
        checkParser(adapted);
    }
    finally {
        fs.rmSync(tree, { recursive: true, force: true });
    }
}

main();
