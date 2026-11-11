#!/usr/bin/env node
"use strict";
// tracing.ts from the tree before this adaptation and the tree after it. Each runtime probe loads both sides into one
// fresh Node process, so the two sides read one process.pid and the trace file names they make compare byte for byte.
// The ES-module rows load the whole of src/compiler through `_namespaces/ts.ts` the way oracle/node.mjs loads a
// source (Node's stripTypeScriptTypes in transform mode, format module), with tsc's `.js` specifiers resolved to their
// `.ts` sources as stage3/drivers/tsc-entry/node.mjs does, and call tracingEnabled.startTracing. As the oracle loads
// it there is no require, so startTracing stops at its require("fs") before it reads process.pid; that row is
// recorded. The ruled rows give the probe a require after the compiler has loaded (isNodeLikeSystem has already run),
// standing in for the node:fs import, which is that change's and not this one's. The CommonJS row loads stock 6.0.3's
// CommonJS output of tracing.ts alone, with its two namespace imports bound to small stand-ins. One row reads the
// checker with @types/node 25.3.3, the declarations Adamic's own checker reads, from stage3/api.
// With --mutant namespace, the after tree's import is `import * as process from "node:process";`, the namespace form.
// With --mutant undo, the after tree's change is undone in memory: the import goes and 65's declaration comes back.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const { pathToFileURL } = require("node:url");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { compiler } = require("../47-host-errors/adapt.cjs");
const { plannedBinding, plannedProgram, diagnosticsOf, importText, declarationText, requireText } = require("./adapt.cjs");

const tracingName = "src/compiler/tracing.ts";
const namespaceText = "import * as process from \"node:process\";";
// A pid no process here has, defined on process before startTracing (pid is read only and configurable): a read of the
// live process object sees it, a copy taken at import doesn't.
const writtenPid = 2147480001;

function tracingSource(tree, mutant) {
    let text = fs.readFileSync(path.join(tree, tracingName), "utf8");
    if (mutant === "namespace") {
        assert.equal(text.split(importText).length - 1, 1, "mutant site drift");
        text = text.replace(importText, namespaceText);
    }
    if (mutant === "undo") {
        assert.equal(text.split(`\r\n${importText}`).length - 1, 1, "mutant site drift");
        text = text.replace(`\r\n${importText}`, "");
        text += text.endsWith(`${requireText}\n`) ? `${declarationText}\n` : `\n${declarationText}\n`;
    }
    return text;
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

// Both trees' src/compiler as ES modules in one process, the after tree's tracing.ts read from `adapted`, which carries
// any mutant. Then, for each side, startTracing(mode) into its own directory, and the names it made there.
function compilerProbe(sides, mode, options) {
    const hook = [
        "import { readFileSync, existsSync, mkdirSync, readdirSync } from 'node:fs';",
        "import { registerHooks, stripTypeScriptTypes, createRequire } from 'node:module';",
        "import { fileURLToPath, pathToFileURL } from 'node:url';",
        `const replaced = new Map(${JSON.stringify(sides.map(side => [pathToFileURL(path.join(side.tree, tracingName)).href, side.tracing]))});`,
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
        "            const source = replaced.get(url) ?? readFileSync(fileURLToPath(url), 'utf8');",
        "            return { format: 'module', source: stripTypeScriptTypes(source, { mode: 'transform', sourceMap: true, sourceUrl: url }), shortCircuit: true };",
        "        }",
        "        return nextLoad(url, context);",
        "    },",
        "});",
        "process.setSourceMapsEnabled(true);",
        `const sides = ${JSON.stringify(sides.map(side => ({ name: side.name, entry: pathToFileURL(path.join(side.tree, "src/compiler/_namespaces/ts.ts")).href })))};`,
        "const loaded = [];",
        "for (const side of sides) loaded.push([side.name, await import(side.entry)]);",
        options.require ? "globalThis.require = createRequire(import.meta.url);" : "",
        options.writePid ? `Object.defineProperty(process, 'pid', { value: ${writtenPid} });` : "",
        "const observed = { pid: process.pid, names: {} };",
        "for (const [name, ts] of loaded) {",
        `    const directory = fileURLToPath(new URL('./trace-${mode}-${options.label}-' + name, import.meta.url));`,
        "    mkdirSync(directory);",
        "    try {",
        `        ts.tracingEnabled.startTracing(${JSON.stringify(mode)}, directory);`,
        "        observed.names[name] = readdirSync(directory).sort();",
        "    }",
        "    catch (error) {",
        "        observed.names[name] = { threw: error.constructor.name, message: error.message.split('\\n')[0] };",
        "    }",
        "}",
        "console.log(JSON.stringify(observed));",
    ];
    return hook.filter(line => line !== "").join("\n");
}

// tracing.ts alone as stock 6.0.3's CommonJS, as upstream's bundled tsc has require, its imports bound to stand-ins with
// the four values startTracing reads.
function commonJS(text) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
}

const hostErrorsStandIn = "exports.errorMessage = error => error && error.message;";
const standIn = [
    "exports.Debug = { assert(condition, message) { if (!condition) throw new Error(message); } };",
    "exports.combinePaths = (directory, name) => directory.replace(/\\/$/, '') + '/' + name;",
    "exports.timestamp = () => performance.now();",
].join("\n");

function commonJSProbe(sides, mode) {
    return [
        "const fs = require('node:fs');",
        "const path = require('node:path');",
        "const observed = { pid: process.pid, names: {} };",
        `for (const name of ${JSON.stringify(sides.map(side => side.name))}) {`,
        "    const tracing = require(`./commonjs-${name}/tracing.js`);",
        `    const directory = path.join(__dirname, 'trace-${mode}-commonjs-' + name);`,
        "    fs.mkdirSync(directory);",
        "    try {",
        `        tracing.tracingEnabled.startTracing(${JSON.stringify(mode)}, directory);`,
        "        observed.names[name] = fs.readdirSync(directory).sort();",
        "    }",
        "    catch (error) {",
        "        observed.names[name] = { threw: error.constructor.name, message: error.message.split('\\n')[0] };",
        "    }",
        "}",
        "console.log(JSON.stringify(observed));",
    ].join("\n");
}

// The default import against Node's own process, and what loading node:process adds to the process's modules once
// sys.ts's and performanceCore.ts's node: imports have loaded, as they have when tracing.ts's import is linked. Node
// builds a builtin's ES module facade from the object's own properties, so process's getters (stdout, stdin, stderr,
// report and the rest) run when node:process is first imported, in either form.
const importProbe = [
    "for (const name of ['node:fs', 'node:path', 'node:os', 'node:crypto', 'node:perf_hooks']) await import(name);",
    "const before = new Set(process.moduleLoadList);",
    "const { default: defaultProcess } = await import('node:process');",
    "const namespaceProcess = await import('node:process');",
    "const added = process.moduleLoadList.filter(name => !before.has(name));",
    "console.log(JSON.stringify({ defaultIsProcess: defaultProcess === globalThis.process, namespaceIsProcess: namespaceProcess === globalThis.process, pid: defaultProcess.pid === globalThis.process.pid, added }));",
].join("\n");

// Whether node:process's stdin, made by that facade, holds the event loop: stdin is a pipe a shell holds open for two
// seconds, and the probe reports its uptime when it exits.
const stdinProbe = [
    "import process from 'node:process';",
    "process.on('exit', () => console.log(JSON.stringify({ exitedWithinOneSecond: process.uptime() < 1 })));",
].join("\n");

// The trace file names startTracing makes in each mode, for a pid.
function expectedNames(mode, pid) {
    const countPart = mode === "build" ? `.${pid}-1` : mode === "server" ? `.${pid}` : "";
    return [`trace${countPart}.json`];
}

// Both sides made the same names, and they are the expected ones.
function namesHold(observed, mode, pid) {
    const want = JSON.stringify(expectedNames(mode, pid ?? observed.pid));
    return observed.names && Object.values(observed.names).every(names => JSON.stringify(names) === want);
}

// The checker with @types/node: tracing.ts's diagnostics on each side, and what its process reads bind to.
function checkerRow(sides) {
    const typeRoot = path.join(path.dirname(require.resolve("@types/node/package.json")), "..");
    const result = {};
    for (const side of sides) {
        const base = compiler(side.tree, ts);
        const options = { ...base.getCompilerOptions(), typeRoots: [typeRoot], types: ["node"] };
        const typed = ts.createProgram(base.getRootFileNames(), options);
        const fileName = path.join(side.tree, tracingName);
        const program = plannedProgram(ts, typed, fileName, side.tracing);
        let binding;
        try {
            assert.deepEqual(plannedBinding(ts, program, fileName, path.join(side.tree, "src/compiler/sys.ts")), []);
            binding = "the node:process default import, the global process of @types/node";
        }
        catch (error) {
            const file = program.getSourceFile(fileName);
            const checker = program.getTypeChecker();
            const reads = [];
            ts.forEachChild(file, function visit(node) {
                if (ts.isPropertyAccessExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === "process") reads.push(node.expression);
                ts.forEachChild(node, visit);
            });
            const declarations = reads.map(read => checker.getSymbolAtLocation(read)?.declarations?.[0]);
            binding = declarations.every(node => node && node.getSourceFile() === file && ts.isVariableDeclaration(node)) ? "the file's own declaration"
                : declarations.every(node => node && ts.isNamespaceImport(node)) ? "a namespace import" : `unreviewed: ${error.message}`;
        }
        result[side.name] = { diagnostics: diagnosticsOf(ts, program, fileName), binding };
    }
    return result;
}

function main() {
    const args = process.argv.slice(2);
    let mutant;
    if (args[0] === "--mutant") {
        mutant = args[1];
        assert(["namespace", "undo"].includes(mutant), "usage: --mutant namespace|undo");
        args.splice(0, 2);
    }
    assert.equal(args.length, 3, "usage: node witnesses.cjs [--mutant namespace|undo] <tree before> <tree after> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [before, after, output] = args.map(arg => path.resolve(arg));
    const sides = [
        { name: "original", tree: before, tracing: tracingSource(before) },
        { name: "adapted", tree: after, tracing: tracingSource(after, mutant) },
    ];
    assert(!sides[0].tracing.includes(importText) && sides[0].tracing.includes(declarationText), "the before tree is already adapted");
    const onDisk = tracingSource(after);
    assert(onDisk.includes(importText) && !onDisk.includes(declarationText), "the after tree is not adapted");
    const scratch = `${output}.probes`;
    fs.mkdirSync(scratch);
    const write = (name, text) => fs.writeFileSync(path.join(scratch, name), text + "\n");
    for (const side of sides) {
        fs.mkdirSync(path.join(scratch, `commonjs-${side.name}`, "_namespaces"), { recursive: true });
        fs.writeFileSync(path.join(scratch, `commonjs-${side.name}`, "tracing.js"), commonJS(side.tracing));
        fs.writeFileSync(path.join(scratch, `commonjs-${side.name}`, "_namespaces", "ts.js"), standIn + "\n");
        fs.writeFileSync(path.join(scratch, `commonjs-${side.name}`, "_namespaces", "ts.performance.js"), "\n");
        fs.writeFileSync(path.join(scratch, `commonjs-${side.name}`, "hostErrors.js"), hostErrorsStandIn + "\n");
    }
    const rows = [];
    const probe = (row, file, text, holds) => {
        write(file, text);
        const observed = run(scratch, file);
        rows.push({ row, observed, ...(holds ? { asRuled: holds(observed) } : { recorded: true }) });
    };
    probe("tracing, ES module, as the oracle loads it", "oracle-build.mjs", compilerProbe(sides, "build", { label: "oracle" }), undefined);
    for (const mode of ["build", "server", "project"]) {
        probe(`tracing, ES module, ${mode}`, `module-${mode}.mjs`, compilerProbe(sides, mode, { label: "module", require: true }), observed => namesHold(observed, mode));
    }
    probe("tracing, ES module, server, process.pid written first", "module-written.mjs", compilerProbe(sides, "server", { label: "written", require: true, writePid: true }), observed => namesHold(observed, "server", writtenPid));
    for (const mode of ["build", "server"]) {
        probe(`tracing, CommonJS, ${mode}`, `commonjs-${mode}.cjs`, commonJSProbe(sides, mode), observed => namesHold(observed, mode));
    }
    probe("node:process, ES module", "process-import.mjs", importProbe, observed => observed.defaultIsProcess === true && observed.pid === true);
    write("process-stdin.mjs", stdinProbe);
    rows.push({ row: "node:process, ES module, stdin a pipe held open", observed: JSON.parse(execFileSync("/bin/sh", ["-c", `sleep 2 | "${process.execPath}" process-stdin.mjs`], { cwd: scratch, encoding: "utf8" }).trim()), recorded: true });
    const checker = checkerRow(sides);
    rows.push({
        row: "checker with @types/node 25.3.3",
        observed: checker,
        asRuled: JSON.stringify(checker.original.diagnostics) === JSON.stringify(checker.adapted.diagnostics)
            && checker.original.binding === "the file's own declaration"
            && checker.adapted.binding === "the node:process default import, the global process of @types/node",
    });
    const failed = rows.filter(row => row.asRuled === false);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, typescript: ts.version, mutant: mutant ?? null, cases: rows.length, failed: failed.map(row => row.row), rows }, null, 2) + "\n");
    for (const row of rows) {
        const verdict = row.recorded ? "recorded" : row.asRuled ? "as ruled" : "differ  ";
        console.log(`${verdict}  ${row.row}: ${JSON.stringify(row.observed)}`);
    }
    const ruled = rows.filter(row => !row.recorded).length;
    console.log(`${ruled - failed.length} of ${ruled} as ruled`);
    if (failed.length) process.exitCode = 1;
}
main();
