#!/usr/bin/env node
"use strict";
// performanceCore.ts whole, from the tree before this adaptation and the tree after it, each run in a fresh Node
// process, with its `./_namespaces/ts.js` import bound to the same tree's own isNodeLikeSystem, extracted from
// core.ts with its ambient declarations. The ES-module rows load the source the way oracle/node.mjs does: Node's
// stripTypeScriptTypes in transform mode, format module, so the program has no require. The CommonJS rows load stock
// 6.0.3's CommonJS output, where require is defined as in upstream's bundled tsc. One row loads the whole of
// src/compiler as ES modules through `_namespaces/ts.ts`, and one imports node:perf_hooks alone, on the oracle's
// Node and, when the gates' switch names one, on internal/hostconfiguration's node-without-inspector.
// A recorded row runs both sides with isNodeLikeSystem's require test removed in memory, as Correctness's change
// to 99q will have it. With --mutant, the after tree's guard reads `isNodeLikeSystem() || perfHooks.performance`.
// tryGetPerformance's return type and PerformanceTime's readonly timeOrigin are types: both sides strip them, so the
// rows read the same JavaScript as without them, and the after tree must carry both.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const { pathToFileURL } = require("node:url");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { anchorText, importText, header, timeOrigin } = require("./adapt.cjs");
const { nodeFor } = require("../../../internal/hostconfiguration/configuration.cjs");
const registry = require("../../../internal/hostconfiguration/configurations.json");

const guard = "    if (isNodeLikeSystem()) {\r\n";
const mutation = "    if (isNodeLikeSystem() || perfHooks.performance) {\r\n";
const requireTest = "\r\n        && typeof require !== \"undefined\";";

function parse(fileName) {
    return ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
}

// isNodeLikeSystem with the ambient declarations it reads, as a module of its own: the function and every top-level
// `declare const` of core.ts, which strip to nothing. With nodeLike, the require test is removed.
function nodeLikeSource(tree, nodeLike) {
    const core = parse(path.join(tree, "src/compiler/core.ts"));
    const declarations = core.statements.filter(node => ts.isVariableStatement(node) && node.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.DeclareKeyword));
    const functions = core.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === "isNodeLikeSystem");
    assert.equal(functions.length, 1, "isNodeLikeSystem moved");
    let text = [...declarations, functions[0]].map(node => node.getText(core)).join("\r\n") + "\r\n";
    assert.equal(text.split(requireTest).length - 1, 1, "isNodeLikeSystem's require test moved: this witness is for today's 99q");
    if (nodeLike) text = text.replace(requireTest, ";");
    return text;
}

function performanceSource(tree, mutant) {
    let text = fs.readFileSync(path.join(tree, "src/compiler/performanceCore.ts"), "utf8");
    assert.equal(text.split(anchorText).length - 1, 1, "performanceCore.ts's import moved");
    if (mutant) {
        assert.equal(text.split(guard).length - 1, 1, "mutant site drift");
        text = text.replace(guard, mutation);
    }
    return { text, imports: text.includes(importText), typed: text.includes(header.adapted) && text.includes(timeOrigin.adapted) };
}

function strip(text) {
    const { stripTypeScriptTypes } = require("node:module");
    return stripTypeScriptTypes(text, { mode: "transform" });
}

function commonJS(text) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
}

// What a probe prints is one JSON line; a probe that throws prints its error instead.
function run(node, directory, file) {
    try {
        const stdout = execFileSync(node, ["--disable-warning=ExperimentalWarning", file], { cwd: directory, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] });
        return JSON.parse(stdout.trim().split("\n").at(-1));
    }
    catch (error) {
        return { exited: error.status, stderr: String(error.stderr).trim().split("\n").filter(line => /Error|at /.test(line)).slice(0, 3) };
    }
}

const performanceProbe = load => [
    load,
    "const hooks = core.tryGetNativePerformanceHooks();",
    "const first = core.timestamp();",
    "console.log(JSON.stringify({",
    "    hooks: hooks !== undefined,",
    "    shouldWriteNativeEvents: hooks?.shouldWriteNativeEvents,",
    "    performance: hooks?.performance === undefined ? 'none' : hooks.performance === perfHooks.performance ? 'node:perf_hooks performance' : hooks.performance === globalThis.performance ? 'globalThis.performance only' : 'another object',",
    "    performanceTime: hooks?.performanceTime === hooks?.performance,",
    "    timestamp: Number.isFinite(first) && core.timestamp() >= first,",
    "    path: hooks === undefined ? 'none' : hooks.shouldWriteNativeEvents ? 'the browser global' : 'node:perf_hooks',",
    "}));",
].join("\n");

// node:perf_hooks alone: whether it loads, and which of Node's own modules and bindings loading it adds.
const importProbe = [
    "const before = new Set(process.moduleLoadList);",
    "const perfHooks = await import('node:perf_hooks');",
    "const added = process.moduleLoadList.filter(name => !before.has(name));",
    "console.log(JSON.stringify({ loaded: typeof perfHooks.performance?.now === 'function', sameAsGlobal: perfHooks.performance === globalThis.performance, inspector: added.filter(name => /inspector/i.test(name)), added }));",
].join("\n");

// The whole of src/compiler as ES modules: oracle/node.mjs's load hook (stripTypeScriptTypes, transform, with its
// source map, format module), with tsc's `.js` specifiers resolved to their `.ts` sources as
// stage3/drivers/tsc-entry/node.mjs does. The mutant's guard is substituted in memory at load.
function compilerProbe(tree, mutant) {
    const entry = path.join(tree, "src/compiler/_namespaces/ts.ts");
    const target = pathToFileURL(path.join(tree, "src/compiler/performanceCore.ts")).href;
    const swap = mutant ? `if (url === ${JSON.stringify(target)}) source = source.replace(${JSON.stringify(guard)}, ${JSON.stringify(mutation)});` : "";
    return [
        "import { readFileSync, existsSync } from 'node:fs';",
        "import { registerHooks, stripTypeScriptTypes } from 'node:module';",
        "import { fileURLToPath } from 'node:url';",
        "registerHooks({",
        "    resolve(specifier, context, nextResolve) {",
        "        if (specifier.startsWith('.') && specifier.endsWith('.js')) {",
        "            const url = new URL(specifier.slice(0, -3) + '.ts', context.parentURL);",
        "            if (existsSync(fileURLToPath(url))) return { url: url.href, shortCircuit: true };",
        "        }",
        "        return nextResolve(specifier, context);",
        "    },",
        "    load(url, context, nextLoad) {",
        "        if (url.endsWith('.ts')) {",
        "            let source = readFileSync(fileURLToPath(url), 'utf8');",
        `            ${swap}`,
        "            return { format: 'module', source: stripTypeScriptTypes(source, { mode: 'transform', sourceMap: true, sourceUrl: url }), shortCircuit: true };",
        "        }",
        "        return nextLoad(url, context);",
        "    },",
        "});",
        "// As oracle/node.mjs does, so a stack names the source's own lines.",
        "process.setSourceMapsEnabled(true);",
        "let observed;",
        "try {",
        `    const ts = await import(${JSON.stringify(pathToFileURL(entry).href)});`,
        "    const hooks = ts.tryGetNativePerformanceHooks();",
        "    observed = { loaded: true, isNodeLikeSystem: ts.isNodeLikeSystem(), sys: ts.sys === undefined ? 'undefined' : 'defined', shouldWriteNativeEvents: hooks?.shouldWriteNativeEvents };",
        "}",
        "catch (error) {",
        "    observed = { loaded: false, threw: error.constructor.name, message: error.message };",
        "}",
        "console.log(JSON.stringify(observed));",
    ].join("\n");
}

// One side's probe files, written into its own directory, and each probe's observation.
function observe(directory, tree, mutant, noInspector) {
    fs.mkdirSync(directory);
    const performance = performanceSource(tree, mutant);
    const write = (name, text) => fs.writeFileSync(path.join(directory, name), text);
    write("core-node-like.mjs", strip(nodeLikeSource(tree, false)));
    write("core-node-like.cjs", commonJS(nodeLikeSource(tree, false)));
    write("core-node-like-stacked.mjs", strip(nodeLikeSource(tree, true)));
    write("performance-core.mjs", strip(performance.text.replace(anchorText, "import { isNodeLikeSystem } from \"./core-node-like.mjs\";")));
    write("performance-core.cjs", commonJS(performance.text.replace(anchorText, "import { isNodeLikeSystem } from \"./core-node-like.cjs\";")));
    write("performance-core-stacked.mjs", strip(performance.text.replace(anchorText, "import { isNodeLikeSystem } from \"./core-node-like-stacked.mjs\";")));
    const probes = {
        "performanceCore, ES module": [process.execPath, "performance-module.mjs", performanceProbe("import * as core from './performance-core.mjs';\nimport * as perfHooks from 'node:perf_hooks';")],
        "performanceCore, CommonJS": [process.execPath, "performance-commonjs.cjs", performanceProbe("const core = require('./performance-core.cjs');\nconst perfHooks = require('node:perf_hooks');")],
        "src/compiler, ES module": [process.execPath, "compiler-module.mjs", compilerProbe(tree, mutant)],
        "node:perf_hooks, ES module": [process.execPath, "perf-hooks-import.mjs", importProbe],
        "node:perf_hooks, node-without-inspector": [noInspector, "perf-hooks-import.mjs", importProbe],
        "performanceCore, ES module, without the require test": [process.execPath, "performance-module-stacked.mjs", performanceProbe("import * as core from './performance-core-stacked.mjs';\nimport * as perfHooks from 'node:perf_hooks';")],
    };
    const result = {};
    for (const [name, [node, file, text]] of Object.entries(probes)) {
        write(file, text + "\n");
        result[name] = node ? run(node, directory, file) : { skipped: `set ${registry["node-without-inspector"].environment}` };
    }
    return { imports: performance.imports, typed: performance.typed, result };
}

const browserGlobal = { hooks: true, shouldWriteNativeEvents: true, performance: "node:perf_hooks performance", performanceTime: true, timestamp: true, path: "the browser global" };
const perfHooksPath = { hooks: true, shouldWriteNativeEvents: false, performance: "node:perf_hooks performance", performanceTime: true, timestamp: true, path: "node:perf_hooks" };
// Which modules loading perf_hooks adds is recorded, not ruled: it is Node's own and may move between releases.
const loads = { loaded: true, sameAsGlobal: true, inspector: [] };
// The ruled rows: what both sides must observe, key by key.
const ruled = {
    "performanceCore, ES module": browserGlobal,
    "performanceCore, CommonJS": perfHooksPath,
    "src/compiler, ES module": { loaded: true, isNodeLikeSystem: false, sys: "undefined", shouldWriteNativeEvents: true },
    "node:perf_hooks, ES module": loads,
};

function main() {
    const args = process.argv.slice(2);
    const mutant = args[0] === "--mutant";
    if (mutant) args.shift();
    assert.equal(args.length, 3, "usage: node witnesses.cjs [--mutant] <tree before> <tree after> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [before, after, output] = args;
    // The gates' own reader: the binary the switch names, probed against configurations.json, or a skip.
    const spec = registry["node-without-inspector"];
    const noInspector = nodeFor({ configuration: { name: "node-without-inspector", node: spec.node, configure: spec.configure, ruling: spec.ruling } }, 1);
    if (noInspector) ruled["node:perf_hooks, node-without-inspector"] = loads;
    const scratch = `${output}.probes`;
    fs.mkdirSync(scratch);
    const sides = { original: observe(path.join(scratch, "original"), before, false, noInspector), adapted: observe(path.join(scratch, "adapted"), after, mutant, noInspector) };
    assert(!sides.original.imports, "the before tree is already adapted");
    assert(sides.adapted.imports, "the after tree is not adapted");
    assert(!sides.original.typed && sides.adapted.typed, "the return type and readonly timeOrigin are not exactly the after tree's");
    const results = Object.keys(sides.original.result).map(name => {
        const result = { row: name, original: sides.original.result[name], adapted: sides.adapted.result[name] };
        if (ruled[name]) result.asRuled = [result.original, result.adapted].every(side => Object.keys(ruled[name]).every(key => JSON.stringify(side[key]) === JSON.stringify(ruled[name][key])));
        else result.recorded = true;
        return result;
    });
    const failed = results.filter(result => result.asRuled === false);
    const moved = results.filter(result => JSON.stringify(result.original) !== JSON.stringify(result.adapted)).map(result => result.row);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, noInspector: noInspector ?? null, typescript: ts.version, mutant, cases: results.length, failed: failed.map(result => result.row), moved, results }, null, 2) + "\n");
    for (const result of results) {
        const verdict = result.recorded ? "recorded" : result.asRuled ? "as ruled" : "differ  ";
        console.log(`${verdict}  ${result.row}: original ${JSON.stringify(result.original)}; adapted ${JSON.stringify(result.adapted)}`);
    }
    for (const line of moved) console.log(`moves  ${line}`);
    const ruledCount = results.filter(result => !result.recorded).length;
    console.log(`${ruledCount - failed.length} of ${ruledCount} as ruled, ${moved.length} move`);
    if (failed.length) process.exitCode = 1;
}
main();
