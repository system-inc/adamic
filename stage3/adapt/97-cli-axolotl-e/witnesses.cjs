#!/usr/bin/env node
"use strict";
// createGetSourceFile and createWriteFileMeasuringIO, extracted whole from program.ts in the tree before
// 97e and the tree after it, each transpiled by stock TypeScript 6.0.3, and run with a callback that throws
// each value below. The returned function's onError records what it receives. Each row must match its
// ruled before and after: null and undefined throw the same TypeError, with propertyOf's frame on top
// after; an Error's string message is unchanged; every other message is converted to text.
// Then what tsc's own callbacks do with what onError received: each site's consumer formats it into its
// diagnostic with stock 6.0.3's formatStringFromArgs and Diagnostics. Before, an undefined message fails
// Debug.checkDefined; after, every received message formats.
// With --mutant, the helper first returns undefined on null.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");

const guard = "    if (value === null || value === undefined) {\n";
const mutation = "    if (value === null) return undefined;\n";
const functions = ["createGetSourceFile", "createWriteFileMeasuringIO"];

// Each site's consumer in tsc: program.ts's findSourceFile hands getSourceFile's message to
// Cannot_read_file_0_Colon_1, and utilities.ts's writeFile hands writeFile's to Could_not_write_file_0_Colon_1.
const sites = {
    createGetSourceFile: { fileName: "a.ts", diagnostic: ts.Diagnostics.Cannot_read_file_0_Colon_1 },
    createWriteFileMeasuringIO: { fileName: "a.js", diagnostic: ts.Diagnostics.Could_not_write_file_0_Colon_1 },
};

function declarations(tree) {
    const fileName = path.join(tree, "src/compiler/program.ts");
    const file = ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
    return functions.map(name => {
        const found = file.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
        assert.equal(found.length, 1, `${name} moved`);
        return found[0].getText(file);
    });
}

function compile(source, globals) {
    const output = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
    const module = { exports: {} };
    new Function("exports", "module", ...Object.keys(globals), output)(module.exports, module, ...Object.values(globals));
    return module.exports;
}

// What the two functions read from the ts namespace: performance's marks are no-ops, createSourceFile
// returns its text, and writeFileEnsuringDirectories calls the supplied writer once, as its try does.
const namespace = {
    performance: { mark() {}, measure() {} },
    createSourceFile: (fileName, text) => ({ fileName, text }),
    writeFileEnsuringDirectories: (fileName, data, writeByteOrderMark, actualWriteFile) => actualWriteFile(fileName, data, writeByteOrderMark),
};

// The ruled rows: a TypeError on both sides, or what onError receives before and after.
const received = value => [{ type: typeof value, value }];
const thrown = [
    ["null", () => null, { threw: "TypeError" }],
    ["undefined", () => undefined, { threw: "TypeError" }],
    ["a string", () => "disk failed", { before: received(undefined), after: received("undefined") }],
    ["a number", () => 5, { before: received(undefined), after: received("undefined") }],
    ["an Error", () => new Error("disk failed"), { before: received("disk failed"), after: received("disk failed") }],
    ["{ message: 42 }", () => ({ message: 42 }), { before: received(42), after: received("42") }],
    ["{}", () => ({}), { before: received(undefined), after: received("undefined") }],
];

// The frames above this harness's own, by function name: the compiled module's frames, which are where
// the two sides can differ. The returned arrow runs in new Function's code, so V8 names it eval.
function frames(stack) {
    const names = stack.split("\n").slice(1).map(line => line.trim());
    // A compiled frame names this file too, in its eval origin, so only a frame outside eval is the harness's.
    const harness = names.findIndex(frame => frame.includes(__filename) && !frame.includes("eval at "));
    assert(harness > 0, "the stack does not reach this harness");
    return names.slice(0, harness).map(frame => frame.replace(/ \(.*\)$/, ""));
}

function observe(call) {
    const messages = [];
    const onError = message => messages.push({ type: typeof message, value: message });
    try {
        const returned = call(onError);
        return { received: messages, returned };
    }
    catch (error) {
        return { received: messages, threw: error.constructor.name, message: error.message, frames: frames(error.stack) };
    }
}

// The consumer's diagnostic text for a received message, or how stock 6.0.3 fails to build it. Debug.fail
// hides checkDefined's own frame (captureStackTrace with checkDefined as the crawl mark), so the failure is
// named by the stock line its top frame points at: formatStringFromArgs's `Debug.checkDefined(args[+index])`,
// utilities.ts:8584 on 050880ce.
function format(site, observation) {
    if (observation.received.length !== 1) return undefined;
    try {
        return { text: ts.formatStringFromArgs(sites[site].diagnostic.message, [sites[site].fileName, observation.received[0].value]) };
    }
    catch (error) {
        const top = /\(?([^()\s]+):(\d+):\d+\)?$/.exec(error.stack.split("\n")[1].trim());
        const line = top ? fs.readFileSync(top[1], "utf8").split("\n")[Number(top[2]) - 1].trim() : undefined;
        return { threw: error.constructor.name, message: error.message.split("\n")[0], at: top ? `${path.basename(top[1])}:${top[2]}` : undefined, line };
    }
}

function calls(exports, make) {
    const getSourceFile = exports.createGetSourceFile(() => { throw make(); }, false);
    const writeFile = exports.createWriteFileMeasuringIO(() => { throw make(); }, () => {}, () => true);
    return [
        ["createGetSourceFile", onError => getSourceFile("a.ts", 99, onError)],
        ["createWriteFileMeasuringIO", onError => writeFile("a.js", "data", false, onError)],
    ];
}

// A row holds when both sides are the ruled ones. A TypeError must be the same on both sides, with only
// propertyOf's frame added on top. Otherwise each side's onError gets its ruled message, the calls return
// the same, and the adapted message always formats into the consumer's diagnostic.
function asRuled(result, expected) {
    const { original, adapted } = result;
    if (expected.threw) {
        return original.threw === expected.threw && adapted.threw === expected.threw && original.message === adapted.message
            && !original.received.length && !adapted.received.length && adapted.frames[0] === "at propertyOf"
            && JSON.stringify(adapted.frames.slice(1)) === JSON.stringify(original.frames);
    }
    const wanted = sites[result.site].diagnostic.message.replace("{0}", sites[result.site].fileName).replace("{1}", String(expected.after[0].value));
    const failsBefore = expected.before[0].value === undefined;
    return !original.threw && !adapted.threw
        && JSON.stringify(original.received) === JSON.stringify(expected.before) && JSON.stringify(adapted.received) === JSON.stringify(expected.after)
        && JSON.stringify(original.returned) === JSON.stringify(adapted.returned)
        && result.formatted.adapted.text === wanted
        && (failsBefore ? result.formatted.original.threw === "Error" && result.formatted.original.message === "Debug Failure." && result.formatted.original.line?.includes("Debug.checkDefined(args[+index])") : result.formatted.original.text !== undefined);
}

// Why a row moves: what onError receives changes, and whether the consumer's answer changes with it.
function movement(result) {
    if (JSON.stringify(result.original.received) === JSON.stringify(result.adapted.received)) return "unchanged";
    const before = result.formatted.original;
    const after = result.formatted.adapted;
    // One side never reached onError: it threw (the mutant's null row).
    if (!before || !after) return `before, ${before ? `onError receives ${JSON.stringify(result.original.received)}` : `${result.original.threw} thrown`}; after, ${after ? `onError receives ${JSON.stringify(result.adapted.received)}` : `${result.adapted.threw} thrown`}`;
    if (before.threw) return `onError's message becomes text, and the consumer's ${before.message} becomes "${after.text}"`;
    return before.text === after.text ? `onError's message becomes text; the consumer's diagnostic text is unchanged, "${after.text}"` : "the consumer's diagnostic text changes";
}

function main() {
    const args = process.argv.slice(2);
    const mutant = args[0] === "--mutant";
    if (mutant) args.shift();
    assert.equal(args.length, 3, "usage: node witnesses.cjs [--mutant] <tree before 97e> <tree after 97e> <output.json>");
    assert.equal(ts.version, "6.0.3");
    const [before, after, output] = args;
    let helper = fs.readFileSync(path.join(after, "src/compiler/hostErrors.ts"), "utf8");
    if (mutant) {
        assert.equal(helper.split(guard).length - 1, 1, "mutant site drift");
        helper = helper.replace(guard, mutation + guard);
    }
    const { propertyOf } = compile(helper, {});
    const sources = { before: declarations(before), after: declarations(after) };
    functions.forEach((name, index) => assert.notEqual(sources.before[index], sources.after[index], `${name} is not adapted in the after tree`));
    assert(sources.before.every(source => !source.includes("propertyOf")), "the before tree is already adapted");
    const original = compile(sources.before.join("\n"), namespace);
    const adapted = compile(sources.after.join("\n"), { ...namespace, propertyOf });
    const results = [];
    for (const [value, make, expected] of thrown) {
        const originals = calls(original, make);
        const adapteds = calls(adapted, make);
        originals.forEach(([site, call], index) => {
            const result = { site, value, original: observe(call), adapted: observe(adapteds[index][1]) };
            result.formatted = { original: format(site, result.original), adapted: format(site, result.adapted) };
            result.asRuled = asRuled(result, expected);
            result.moves = movement(result);
            results.push(result);
        });
    }
    const failed = results.filter(result => !result.asRuled);
    const moved = results.filter(result => result.moves !== "unchanged").map(result => `${result.site}, ${result.value}: ${result.moves}`);
    fs.writeFileSync(output, JSON.stringify({ node: process.version, typescript: ts.version, mutant, cases: results.length, failed: failed.map(result => `${result.site}, ${result.value}`), moved, results }, null, 2) + "\n");
    for (const result of results) {
        console.log(`${result.asRuled ? "as ruled" : "DIFFER  "}  ${result.site}, thrown ${result.value}: original ${JSON.stringify(result.original)} formats ${JSON.stringify(result.formatted.original)}; adapted ${JSON.stringify(result.adapted)} formats ${JSON.stringify(result.formatted.adapted)}`);
    }
    for (const line of moved) console.log(`moves  ${line}`);
    console.log(`${results.length - failed.length} of ${results.length} as ruled, ${moved.length} move`);
    if (failed.length) process.exitCode = 1;
}
main();
