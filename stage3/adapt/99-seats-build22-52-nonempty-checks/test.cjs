"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require(process.env.CENSUS_TYPESCRIPT || "../../api/node_modules/typescript");
const indexedReads = require("../30-indexed-reads/adapt.cjs");
const { adapt } = require("./adapt.cjs");

function implementation(tree, file, name) {
    const source = ts.createSourceFile(file, fs.readFileSync(path.join(tree, "src/compiler", file), "utf8"),
        ts.ScriptTarget.ES2024, true);
    assert.deepEqual(source.parseDiagnostics, []);
    const matches = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
    assert.equal(matches.length, 1, name);
    return matches[0].getText(source).replace(/^export /, "");
}

function run(tree, name, state, variant) {
    const source = implementation(tree, "core.ts", "some") + "\n"
        + implementation(tree, "visitorPublic.ts", name);
    const emitted = ts.transpileModule(source + "\nglobalThis.visit = " + name + ";", {
        compilerOptions: { target: ts.ScriptTarget.ES2024 },
        reportDiagnostics: true,
    });
    assert.deepEqual(emitted.diagnostics, []);
    const events = [];
    const declarations = state === "undefined" ? undefined : state === "empty" ? [] : ["hoisted"];
    const updated = variant === "missing" ? undefined : variant === "block"
        ? { block: true, statements: ["body"] } : "body";
    const globals = {
        Debug: { assert: assert.ok },
        visitNode: () => assert.fail("visitor argument was replaced"),
        isConciseBody: () => true,
        isStatement: () => true,
        isBlock: node => node.block,
        factory: {
            mergeLexicalEnvironment(statements, hoisted) {
                events.push("merge");
                return hoisted.concat(statements);
            },
        },
    };
    const context = {
        resumeLexicalEnvironment() { events.push("resume"); },
        startBlockScope() { events.push("start"); },
        endLexicalEnvironment() { events.push("end"); return declarations; },
        endBlockScope() { events.push("end"); return declarations; },
        factory: {
            liftToBlock() { assert.fail("visitor called liftToBlock"); },
            converters: {
                convertToFunctionBlock(node) { events.push("convert"); return { statements: [node] }; },
            },
            createBlock(statements) { events.push("create"); return { statements }; },
            updateBlock(_block, statements) { events.push("update"); return { statements }; },
        },
    };
    vm.runInNewContext(emitted.outputText, globals, { timeout: 1000 });
    const result = globals.visit("input", () => {}, context, () => {
        events.push("visit");
        return updated;
    });
    return { events, result, declarations, sameResult: result === updated,
        usesDeclarations: declarations !== undefined && result?.statements === declarations };
}

const pristine = process.argv[2];
assert(pristine, "usage: node test.cjs <pristine TypeScript v6.0.3 checkout>");
const tree = fs.mkdtempSync(path.join(os.tmpdir(), "nonempty-checks-"));
try {
    fs.mkdirSync(path.join(tree, "src"));
    fs.cpSync(path.join(pristine, "src/compiler"), path.join(tree, "src/compiler"), { recursive: true });
    fs.copyFileSync(path.join(pristine, "src/tsconfig-base.json"), path.join(tree, "src/tsconfig-base.json"));
    // The production adapter pins some's body after adapter 30 adds its indexed-read assertion.
    const core = path.join(tree, "src/compiler/core.ts");
    const indexed = indexedReads.plan(ts, "core.ts", fs.readFileSync(core, "utf8"));
    fs.writeFileSync(core, indexed.text);
    adapt(tree, ts);
    for (const [name, variants] of [["visitFunctionBody", ["missing", "statement"]],
        ["visitIterationBody", ["block", "statement"]]]) {
        for (const state of ["undefined", "empty", "nonempty"]) {
            for (const variant of variants) {
                const label = `${name} ${state} ${variant}`;
                assert.deepEqual(run(tree, name, state, variant), run(pristine, name, state, variant), label);
                console.log("PASS " + label);
            }
        }
    }
} finally {
    fs.rmSync(tree, { recursive: true, force: true });
}
