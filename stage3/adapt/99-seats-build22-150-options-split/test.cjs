#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const { createRequire } = require("node:module");
const { compileFunction } = require("node:vm");
const path = require("node:path");

const [tree, adapter = path.join(__dirname, "adapt.cjs")] = process.argv.slice(2);
assert(tree && process.argv.length <= 4, "usage: node test.cjs <applied-tree> [adapter]");
const relative = "src/compiler/commandLineParser.ts";
const snapshot = side => execFileSync("git", [
    "-C", tree, "show", `refs/stage3/${path.basename(__dirname)}-${side}:${relative}`,
]).toString("utf8");
const input = snapshot("before");
const expected = snapshot("after");
const digest = bytes => createHash("sha256").update(bytes).digest("hex");
const crlf = lines => lines.join("\r\n");
const env = { ...process.env, NODE_PATH: path.resolve(__dirname, "../../api/node_modules") };
const bundlePath = path.join(env.NODE_PATH, "typescript/lib/typescript.js");
const ts = require(bundlePath);
assert.equal(ts.version, "6.0.3");

function replaceOnce(text, before, after) {
    const start = text.indexOf(before);
    assert(start >= 0 && start === text.lastIndexOf(before), `plant must match once: ${before}`);
    return text.slice(0, start) + after + text.slice(start + before.length);
}

const slot = "    let watchOptions: WatchOptions | undefined;";
const slotContext = declaration => crlf([
    "    const options = {} as OptionsBase;", declaration, "    const fileNames: string[] = [];",
]);
const falsy = text => replaceOnce(text, slotContext(slot),
    slotContext("    let watchOptions: WatchOptions | undefined | false = false;"));
const parameter = text => replaceOnce(text, crlf([
    "    opt: CommandLineOption,", "    options: OptionsBase,", "    errors: Diagnostic[],", ") {",
]), crlf([
    "    opt: CommandLineOption,", "    options: OptionsBase | false,", "    errors: Diagnostic[],", ") {",
    "    if (options === false) return i + 1;",
]));
const typedDrift = text => parameter(replaceOnce(falsy(text), crlf([
    "    return {", "        options,", "        watchOptions,", "        fileNames,",
]), crlf([
    "    return {", "        options,", "        watchOptions: watchOptions === false ? undefined : watchOptions,",
    "        fileNames,",
])));
const watchOpt = "                    const watchOpt = getOptionDeclarationFromName(";
const beforeWatchOpt = (text, statement) => replaceOnce(text, watchOpt, `${statement}\r\n${watchOpt}`);
const ordinaryCall = "                    i = parseOptionValue(args, i, diagnostics, opt, options, errors);";
const plants = [
    ["a falsy declaration", "watchOptions declaration drift", falsy],
    ["b typed drift", "watchOptions declaration drift", typedDrift],
    ["c options parameter", "parseOptionValue options parameter drift", parameter],
    ["d callee destructuring shadow", "parseOptionValue binding drift", text =>
        beforeWatchOpt(text, "                    const { parseOptionValue } = replacement;")],
    ["d args block shadow", "args binding drift", text =>
        beforeWatchOpt(text, "                    const args = replacement;")],
    ["d i initializer", "i binding drift", text =>
        replaceOnce(text, "        let i = 0;\r\n        while (i < args.length)", "        let i = 1;\r\n        while (i < args.length)")],
    ["d diagnostics var shadow", "watchOptionsDidYouMeanDiagnostics binding drift", text =>
        replaceOnce(text, ordinaryCall, `                    var watchOptionsDidYouMeanDiagnostics = diagnostics;\r\n${ordinaryCall}`)],
    ["d watchOpt initializer", "watchOpt binding drift", text =>
        replaceOnce(text, watchOpt, "                    const watchOpt = replacement(")],
];

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "family150-"));
const file = path.join(scratch, relative);
fs.mkdirSync(path.dirname(file), { recursive: true });
let failures = 0;

function run(label, text, check) {
    fs.writeFileSync(file, text);
    const before = digest(fs.readFileSync(file));
    const result = spawnSync(process.execPath, [path.resolve(adapter), scratch], { env, encoding: "utf8" });
    if (result.error) throw result.error;
    assert.equal(result.signal, null, label);
    const bytes = fs.readFileSync(file);
    const after = digest(bytes);
    const message = result.stderr.split("\n").find(line => line.startsWith("AssertionError")) ?? "";
    const passed = check(result, before, after, message);
    console.log(`${passed ? "PASS" : "FAIL"} ${label}: adapter exit ${result.status}; sha256 ${before} -> ${after}${message ? `; ${message}` : ""}`);
    if (!passed) failures++;
    return bytes.toString("utf8");
}

function workerSpan(js, label) {
    const source = ts.createSourceFile(label, js, ts.ScriptTarget.ES2024, true, ts.ScriptKind.JS);
    assert.equal(source.parseDiagnostics.length, 0, label);
    const workers = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === "parseCommandLineWorker") {
            workers.push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(workers.length, 1, `${label}: parseCommandLineWorker declaration count`);
    // The ESNext emit exports the worker. Splice the declaration from its function keyword.
    const keyword = workers[0].getChildren(source).find(node => node.kind === ts.SyntaxKind.FunctionKeyword);
    assert(keyword, label);
    return [keyword.getStart(source), workers[0].end];
}

function checkBehavior(originalText, adaptedText) {
    const originalBundle = fs.readFileSync(bundlePath, "utf8");
    const [start, end] = workerSpan(originalBundle, bundlePath);
    const bundleRequire = createRequire(bundlePath);
    function loadParser(text, label) {
        const emitted = ts.transpileModule(text, {
            fileName: relative,
            compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext },
            reportDiagnostics: true,
        });
        assert.deepStrictEqual(emitted.diagnostics, [], label);
        const [workerStart, workerEnd] = workerSpan(emitted.outputText, label);
        const js = originalBundle.slice(0, start) + emitted.outputText.slice(workerStart, workerEnd)
            + originalBundle.slice(end);
        const parserModule = { exports: {} };
        compileFunction(js, ["exports", "require", "module", "__filename", "__dirname"], {
            filename: bundlePath,
        })(parserModule.exports, bundleRequire, parserModule, bundlePath, path.dirname(bundlePath));
        assert.equal(parserModule.exports.version, "6.0.3", label);
        return parserModule.exports;
    }

    const original = loadParser(originalText, "original parser");
    const head = loadParser(adaptedText, "adapted parser");
    const watchFile = ts.WatchFileKind.FixedPollingInterval;
    const cases = [
        { name: "no watch argument", args: ["a.ts"], watchOptions: undefined },
        { name: "valid watch argument", args: ["--watchFile", "fixedpollinginterval", "a.ts"],
            watchOptions: { watchFile } },
        { name: "second watch option via response file",
            args: ["--watchFile", "fixedpollinginterval", "@watch.args", "a.ts"],
            watchOptions: { watchFile, fallbackPolling: ts.PollingWatchKind.PriorityInterval } },
        { name: "invalid watch value", args: ["--watchFile", "bogus"],
            watchOptions: { watchFile: undefined }, fileNames: [], codes: [6046] },
    ];
    const parse = (parser, args) => {
        const result = parser.parseCommandLine(args, name =>
            name === "watch.args" ? "--fallbackPolling priorityinterval" : undefined);
        return {
            options: result.options,
            watchOptions: result.watchOptions,
            fileNames: result.fileNames,
            errors: result.errors.map(({ code, messageText }) => ({ code, messageText })),
        };
    };
    for (const { name, args, watchOptions, fileNames = ["a.ts"], codes = [] } of cases) {
        const before = parse(original, args);
        const after = parse(head, args);
        compare(`behavior ${name}: original matches stock`, before, parse(ts, args));
        compare(`behavior ${name}: head matches original`, after, before);
        compare(`behavior ${name}: head matches expected`, {
            ...after,
            errors: after.errors.map(error => error.code),
        }, { options: {}, watchOptions, fileNames, errors: codes });
    }
}

function compare(label, actual, expected) {
    try {
        assert.deepStrictEqual(actual, expected);
        console.log(`PASS ${label}`);
    } catch (error) {
        if (error.code !== "ERR_ASSERTION") throw error;
        console.log(`FAIL ${label}\n${error.message}`);
        failures++;
    }
}

try {
    const adapted = run("clean first", input, (result, before, after) =>
        result.status === 0 && before !== after);
    compare("clean output matches this adapter's after snapshot", adapted, expected);
    run("clean already adapted", adapted, (result, before, after) =>
        result.status === 0 && before === after);
    for (const [name, diagnostic, plant] of plants) {
        for (const [phase, text] of [["first", input], ["already adapted", adapted]]) {
            run(`${name}, ${phase}`, plant(text), (result, before, after, message) =>
                result.status !== 0 && before === after && message.includes(diagnostic));
        }
    }
    checkBehavior(input, adapted);
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
console.log(`${failures} failing checks`);
process.exitCode = failures ? 1 : 0;
