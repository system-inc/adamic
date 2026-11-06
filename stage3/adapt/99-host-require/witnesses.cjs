#!/usr/bin/env node
"use strict";
// core.ts's isNodeLikeSystem and performanceCore.ts whole, from the tree before this adaptation and the tree after
// it, each run in a fresh Node process. The ES-module rows load the source the way oracle/node.mjs does: Node's own
// stripTypeScriptTypes in transform mode, format module, so the program has no require, module or __filename. The
// CommonJS rows load stock 6.0.3's CommonJS output, where require is defined as it is in upstream's bundled tsc.
// performanceCore.ts's `./_namespaces/ts.js` import is bound to the same tree's extracted isNodeLikeSystem. The last
// rows load the whole of src/compiler as ES modules through `_namespaces/ts.ts`, as an oracle run of tsc's own
// source would, and record whether sys.ts's getNodeSystem ran and what it did.
// With --mutant, the after tree's isNodeLikeSystem gets `&& typeof require !== "undefined"` back, in memory.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { nodeLikeBody } = require("./adapt.cjs");

const anchor = "import { isNodeLikeSystem } from \"./_namespaces/ts.js\";";

function parse(fileName) {
    return ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
}

// isNodeLikeSystem with the ambient declarations it reads, as one module of its own: the function and every
// top-level `declare const` of core.ts, which strip to nothing.
function nodeLikeSource(tree, mutant) {
    const core = parse(path.join(tree, "src/compiler/core.ts"));
    const declarations = core.statements.filter(node => ts.isVariableStatement(node) && node.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.DeclareKeyword));
    const functions = core.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === "isNodeLikeSystem");
    assert.equal(functions.length, 1, "isNodeLikeSystem moved");
    let text = [...declarations, functions[0]].map(node => node.getText(core)).join("\r\n") + "\r\n";
    const body = functions[0].body.getText(core);
    if (mutant) {
        assert.equal(body, nodeLikeBody.adapted, "mutant site drift");
        text = text.replace(nodeLikeBody.adapted, nodeLikeBody.original);
    }
    return { text, tests: body.includes("typeof require") || mutant };
}

function performanceSource(tree) {
    const text = fs.readFileSync(path.join(tree, "src/compiler/performanceCore.ts"), "utf8");
    assert.equal(text.split(anchor).length - 1, 1, "performanceCore.ts's import moved");
    return text;
}

function strip(text) {
    const { stripTypeScriptTypes } = require("node:module");
    return stripTypeScriptTypes(text, { mode: "transform" });
}

function commonJS(text) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
}

// What a probe prints is one JSON line; a probe that throws prints its error instead.
function run(directory, file) {
    try {
        const stdout = execFileSync(process.execPath, ["--disable-warning=ExperimentalWarning", file], { cwd: directory, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] });
        return JSON.parse(stdout.trim().split("\n").at(-1));
    }
    catch (error) {
        return { exited: error.status, stderr: String(error.stderr).trim().split("\n").filter(line => /Error|at /.test(line)).slice(0, 3) };
    }
}

const performanceProbe = (load, from) => [
    load,
    `const hooks = core.tryGetNativePerformanceHooks();`,
    `const first = core.timestamp();`,
    `console.log(JSON.stringify({`,
    `    hooks: hooks !== undefined,`,
    `    shouldWriteNativeEvents: hooks?.shouldWriteNativeEvents,`,
    `    performance: hooks?.performance === undefined ? "none" : hooks.performance === perfHooks.performance ? "node:perf_hooks performance" : hooks.performance === globalThis.performance ? "globalThis.performance only" : "another object",`,
    `    performanceTime: hooks?.performanceTime === hooks?.performance,`,
    `    timestamp: Number.isFinite(first) && core.timestamp() >= first,`,
    `    path: hooks === undefined ? "none" : hooks.shouldWriteNativeEvents ? "the browser global" : "${from}",`,
    `}));`,
].join("\n");

// One side's probe files, written into its own directory, and each probe's observation.
function observe(directory, tree, mutant) {
    fs.mkdirSync(directory);
    const nodeLike = nodeLikeSource(tree, mutant);
    const performance = performanceSource(tree);
    fs.writeFileSync(path.join(directory, "core-node-like.mjs"), strip(nodeLike.text));
    fs.writeFileSync(path.join(directory, "core-node-like.cjs"), commonJS(nodeLike.text));
    fs.writeFileSync(path.join(directory, "performance-core.mjs"), strip(performance.replace(anchor, "import { isNodeLikeSystem } from \"./core-node-like.mjs\";")));
    fs.writeFileSync(path.join(directory, "performance-core.cjs"), commonJS(performance.replace(anchor, "import { isNodeLikeSystem } from \"./core-node-like.cjs\";")));
    const probes = {
        "ES module, Node": ["node-like-module.mjs", "import { isNodeLikeSystem } from './core-node-like.mjs';\nconsole.log(JSON.stringify({ isNodeLikeSystem: isNodeLikeSystem(), require: typeof require }));"],
        "ES module, process.browser set": ["node-like-browser.mjs", "import { isNodeLikeSystem } from './core-node-like.mjs';\nprocess.browser = true;\nconsole.log(JSON.stringify({ isNodeLikeSystem: isNodeLikeSystem(), require: typeof require }));"],
        "CommonJS, Node": ["node-like-commonjs.cjs", "const { isNodeLikeSystem } = require('./core-node-like.cjs');\nconsole.log(JSON.stringify({ isNodeLikeSystem: isNodeLikeSystem(), require: typeof require }));"],
        "performanceCore, ES module": ["performance-module.mjs", performanceProbe("import * as core from './performance-core.mjs';\nimport * as perfHooks from 'node:perf_hooks';", "node:perf_hooks")],
        "performanceCore, CommonJS": ["performance-commonjs.cjs", performanceProbe("const core = require('./performance-core.cjs');\nconst perfHooks = require('node:perf_hooks');", "node:perf_hooks")],
        "src/compiler, ES module": ["compiler-module.mjs", compilerProbe(tree, mutant)],
    };
    const result = {};
    for (const [name, [file, text]] of Object.entries(probes)) {
        fs.writeFileSync(path.join(directory, file), text + "\n");
        result[name] = run(directory, file);
    }
    return { tests: nodeLike.tests, result };
}

// The whole of src/compiler as ES modules: oracle/node.mjs's load hook (stripTypeScriptTypes, transform, with its
// source map, format module), with tsc's `.js` specifiers resolved to their `.ts` sources as stage3/drivers/tsc-entry/node.mjs does.
// The mutant's isNodeLikeSystem is substituted in memory at load.
function compilerProbe(tree, mutant) {
    const entry = path.join(tree, "src/compiler/_namespaces/ts.ts");
    const core = path.join(tree, "src/compiler/core.ts");
    const swap = mutant ? `if (url === ${JSON.stringify(require("node:url").pathToFileURL(core).href)}) source = source.replace(${JSON.stringify(nodeLikeBody.adapted)}, ${JSON.stringify(nodeLikeBody.original)});` : "";
    return [
        "import { readFileSync, existsSync } from 'node:fs';",
        "import { registerHooks, stripTypeScriptTypes } from 'node:module';",
        "import { fileURLToPath, pathToFileURL } from 'node:url';",
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
        `    const ts = await import(${JSON.stringify(require("node:url").pathToFileURL(entry).href)});`,
        "    observed = { loaded: true, isNodeLikeSystem: ts.isNodeLikeSystem(), sys: ts.sys === undefined ? 'undefined' : typeof ts.sys.readFile === 'function' ? 'the Node system' : 'another object' };",
        "    if (ts.sys !== undefined) observed.executingFilePath = ts.sys.getExecutingFilePath();",
        "}",
        "catch (error) {",
        "    observed = { loaded: false, threw: error.constructor.name, message: error.message, at: (error.stack.split('\\n').find(line => line.includes('.ts:')) ?? '').trim().replace(/^at /, '').replace(/\\(?(?:file:\\/\\/)?\\/[^ ]*?\\/src\\//, 'src/').replace(/\\)$/, '') };",
        "}",
        "console.log(JSON.stringify(observed));",
    ].join("\n");
}

// The ruled rows: what the original and adapted sides must each observe.
const ruled = {
    "ES module, Node": { original: { isNodeLikeSystem: false, require: "undefined" }, adapted: { isNodeLikeSystem: true, require: "undefined" } },
    "ES module, process.browser set": { original: { isNodeLikeSystem: false, require: "undefined" }, adapted: { isNodeLikeSystem: false, require: "undefined" } },
    "CommonJS, Node": { original: { isNodeLikeSystem: true, require: "function" }, adapted: { isNodeLikeSystem: true, require: "function" } },
    "performanceCore, ES module": {
        original: { hooks: true, shouldWriteNativeEvents: true, performance: "node:perf_hooks performance", performanceTime: true, timestamp: true, path: "the browser global" },
        adapted: { hooks: true, shouldWriteNativeEvents: false, performance: "node:perf_hooks performance", performanceTime: true, timestamp: true, path: "node:perf_hooks" },
    },
    "performanceCore, CommonJS": {
        original: { hooks: true, shouldWriteNativeEvents: false, performance: "node:perf_hooks performance", performanceTime: true, timestamp: true, path: "node:perf_hooks" },
        adapted: { hooks: true, shouldWriteNativeEvents: false, performance: "node:perf_hooks performance", performanceTime: true, timestamp: true, path: "node:perf_hooks" },
    },
};

function main() {
    const args = process.argv.slice(2);
    const mutant = args[0] === "--mutant";
    if (mutant) args.shift();
    assert.equal(args.length, 3, "usage: node witnesses.cjs [--mutant] <tree before> <tree after> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [before, after, output] = args;
    const scratch = `${output}.probes`;
    fs.mkdirSync(scratch);
    const sides = { original: observe(path.join(scratch, "original"), before, false), adapted: observe(path.join(scratch, "adapted"), after, mutant) };
    assert(sides.original.tests, "the before tree is already adapted");
    assert.equal(sides.adapted.tests, mutant, "the after tree is not adapted");
    const results = Object.keys(sides.original.result).map(name => {
        const result = { row: name, original: sides.original.result[name], adapted: sides.adapted.result[name] };
        if (ruled[name]) result.asRuled = JSON.stringify(result.original) === JSON.stringify(ruled[name].original) && JSON.stringify(result.adapted) === JSON.stringify(ruled[name].adapted);
        else result.recorded = true;
        return result;
    });
    const failed = results.filter(result => result.asRuled === false);
    const moved = results.filter(result => JSON.stringify(result.original) !== JSON.stringify(result.adapted)).map(result => result.row);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, typescript: ts.version, mutant, cases: results.length, failed: failed.map(result => result.row), moved, results }, null, 2) + "\n");
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
