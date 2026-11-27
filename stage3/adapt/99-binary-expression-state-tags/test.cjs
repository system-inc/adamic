#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const { createRequire } = require("node:module");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

assert.equal(process.argv.length, 3, "usage: node test.cjs <tree stage3/apply.sh wrote>");
assert.equal(ts.version, "6.0.3");
const tree = path.resolve(process.argv[2]);
const relative = "src/compiler/factory/utilities.ts";
const ref = "refs/stage3/99-binary-expression-state-tags";
const adapter = path.join(__dirname, "adapt.cjs");

function git(...args) {
    return execFileSync("git", ["-C", tree, ...args], { encoding: "utf8", maxBuffer: 4 * 1024 * 1024 });
}

const before = git("show", `${ref}-before:${relative}`);
const after = git("show", `${ref}-after:${relative}`);

// The adapter is a CLI. Evaluate its declarations with its own require, without invoking the CLI.
const adapterText = fs.readFileSync(adapter, "utf8");
const cli = adapterText.indexOf("\nassert.equal(process.argv.length, 3,");
assert.notEqual(cli, -1, "adapter CLI boundary");
const owners = vm.runInNewContext(adapterText.slice(0, cli) + "\nowners;", {
    require: createRequire(adapter),
}, { filename: adapter });
const original = owners(before);
const adapted = owners(after);

let passed = 0;
function pass(label) {
    passed++;
    console.log("PASS " + label);
}

function scratchRow(label, input, output, expectedError) {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "binary-expression-state-test-"));
    try {
        const file = path.join(scratch, relative);
        fs.mkdirSync(path.dirname(file), { recursive: true });
        fs.writeFileSync(file, input);
        const fixedTime = new Date("2000-01-01T00:00:00Z");
        fs.utimesSync(file, fixedTime, fixedTime);
        const beforeTime = fs.statSync(file).mtimeMs;
        const run = spawnSync(process.execPath, [adapter, scratch], { encoding: "utf8" });
        if (run.error) throw run.error;
        assert.equal(run.signal, null, label);
        if (expectedError) {
            assert.notEqual(run.status, 0, label + " must refuse");
            assert.equal(run.stderr.includes(expectedError), true, run.stderr.slice(0, 1800));
        } else {
            assert.equal(run.status, 0, run.stderr.slice(0, 1800));
        }
        assert.equal(Buffer.compare(fs.readFileSync(file), Buffer.from(output)), 0, label + " output bytes");
        if (input === output) assert.equal(fs.statSync(file).mtimeMs, beforeTime, label + " wrote the file");
        pass(label);
    } finally {
        fs.rmSync(scratch, { recursive: true, force: true });
    }
}

assert.equal(git("diff", "--name-only", ref + "-before", ref + "-after").trim(), relative);
pass("control: only factory utilities changed");
scratchRow("control: replay before to after", before, after);
scratchRow("control: after writes nothing", after, after);

function replaceOwner(source, node, text) {
    return source.slice(0, node.getStart()) + text + source.slice(node.end);
}

const changedMachine = replaceOwner(before, original.machine, original.machine.getText().replace("onEnter:", "onEntex:"));
assert.equal(changedMachine.length, before.length);
assert.notEqual(changedMachine === before, true);
scratchRow("refusal: changed class byte", changedMachine, changedMachine, "state machine callbacks changed");

const mixed = replaceOwner(before, original.state, adapted.state.getText());
scratchRow("refusal: mixed snapshot", mixed, mixed, "partially adapted walker");

const unknown = replaceOwner(before, original.state, original.state.getText().replace("outerState:", "otherState:"));
assert.notEqual(unknown === before, true);
scratchRow("refusal: unknown owner text", unknown, unknown, "state: unrecognized snapshot");

const duplicate = before + "\r\n" + original.state.getText() + "\r\n";
scratchRow("refusal: duplicate owner", duplicate, duplicate, "BinaryExpressionState: expected one declaration");

function compile(declarations) {
    const source = Object.values(declarations).map(node => node.getText()).join("\n");
    const js = ts.transpileModule(source, {
        compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS },
    }).outputText;
    return new vm.Script(js + "\nexports[" + JSON.stringify(declarations.walker.name.text) + "](...callbacks)(root, outer);");
}

function stubs(trace, label, originalSide) {
    function failure(message) {
        return message ? `Debug Failure. ${message}` : "Debug Failure.";
    }
    function record(name, args, ok, message) {
        const values = args.map((value, index) =>
            originalSide && name === "assertEqual" && index < 2 && typeof value === "function"
                ? value.name
                : label(value));
        trace.push([name, values, ok, message ?? null]);
        if (!ok) throw new Error(message);
    }
    return {
        AssertionLevel: { Aggressive: 2 },
        Debug: {
            shouldAssert() {
                record("shouldAssert", Array.from(arguments), true, null);
                return true;
            },
            assertEqual(a, b, msg, msg2) {
                const ok = a === b;
                const detail = msg ? (msg2 ? `${msg} ${msg2}` : msg) : "";
                record("assertEqual", Array.from(arguments), ok,
                    ok ? null : failure(`Expected ${a} === ${b}. ${detail}`));
            },
            assertIsDefined(value, message) {
                const ok = value !== undefined && value !== null;
                record("assertIsDefined", Array.from(arguments), ok, ok ? null : failure(message));
            },
            assert(expression, message, verbose) {
                const ok = Boolean(expression);
                let detail = message ? `False expression: ${message}` : "False expression.";
                if (!ok && verbose) detail += "\r\nVerbose Debug Information: " + (typeof verbose === "string" ? verbose : verbose());
                record("assert", Array.from(arguments), ok, ok ? null : failure(detail));
            },
            fail(message) {
                record("fail", Array.from(arguments), false, failure(message));
            },
        },
    };
}

function run(script, { left, operator, right, fold, circular, throwing, nested, objects }, originalSide) {
    const trace = [];
    const root = { name: "root", left: { name: "L", operatorToken: "*" }, right: { name: "R", operatorToken: "-" }, operatorToken: "+" };
    const inputs = [root, root.left, root.right];
    if (nested) {
        root.left.left = { name: "LL", operatorToken: "a" };
        root.left.right = { name: "LR", operatorToken: "b" };
        root.right.left = { name: "RL", operatorToken: "c" };
        root.right.right = { name: "RR", operatorToken: "d" };
        inputs.push(root.left.left, root.left.right, root.right.left, root.right.right);
    }
    const identities = new Map(inputs.map(node => [node, { node: node.name }]));
    const outer = { name: "outer" };
    identities.set(outer, { role: "outer" });
    const creations = new Map();
    function registerState(callbackName, state) {
        const count = (creations.get(callbackName) ?? 0) + 1;
        creations.set(callbackName, count);
        identities.set(state, { role: `${callbackName}#${count}` });
        return state;
    }
    const foreign = new Map();
    if (circular) root.left = root;
    function callback(name, args) {
        trace.push([name, Array.from(args, observe)]);
    }
    function onEnter(node) {
        callback("onEnter", arguments);
        return objects ? registerState("onEnter", { name: node.name }) : node.name;
    }
    function onLeft(child) {
        callback("onLeft", arguments);
        return child;
    }
    function onOperator() {
        callback("onOperator", arguments);
    }
    function onRight(child) {
        callback("onRight", arguments);
        return child;
    }
    function onExit(node, state) {
        callback("onExit", arguments);
        if (throwing && node === root) throw new Error("callback failure");
        return state;
    }
    function foldState(state, result, side) {
        callback("foldState", arguments);
        if (objects) return registerState("foldState", { state, result, side });
        return `${state}/${side}=${result}`;
    }
    const callbacks = [onEnter, left ? onLeft : undefined, operator ? onOperator : undefined,
        right ? onRight : undefined, onExit, fold ? foldState : undefined];
    function label(value) {
        if (value === undefined) return { undefined: true };
        if (identities.has(value)) return identities.get(value);
        if (typeof value === "function") {
            const index = callbacks.indexOf(value);
            if (index >= 0) return { callback: index };
        }
        if (typeof value === "function" || (value !== null && typeof value === "object")) {
            // Give foreign references their own labels, even when two have the same name.
            if (!foreign.has(value)) foreign.set(value, foreign.size);
            return { foreign: typeof value, id: foreign.get(value) };
        }
        return value;
    }
    function observe(value) {
        const identity = label(value);
        if (!identities.has(value)) return identity;
        const properties = Reflect.ownKeys(value).map(key => {
            const descriptor = Object.getOwnPropertyDescriptor(value, key);
            const { enumerable, configurable } = descriptor;
            const flags = { enumerable, configurable };
            const property = Object.hasOwn(descriptor, "value")
                ? { value: label(descriptor.value), writable: descriptor.writable, ...flags }
                : { get: label(descriptor.get), set: label(descriptor.set), ...flags };
            return [String(key), property];
        });
        const prototype = Object.getPrototypeOf(value);
        return {
            ...identity,
            properties,
            extensible: Object.isExtensible(value),
            prototype: prototype === Object.prototype ? "Object.prototype" : label(prototype),
        };
    }
    const context = vm.createContext({ ...stubs(trace, label, originalSide), exports: {}, callbacks, root, outer });
    try {
        trace.push(["result", observe(script.runInContext(context, { timeout: 5000 }))]);
    } catch (error) {
        if (error.code === "ERR_SCRIPT_EXECUTION_TIMEOUT") throw error;
        trace.push(["throw", error.message]);
    }
    trace.push(["afterWalk", Array.from(identities.keys(), observe)]);
    return trace;
}

const cases = [];
for (let mask = 0; mask < 8; mask++) {
    for (const fold of [false, true]) {
        cases.push({
            name: `handlers=${mask} fold=${fold}`, left: Boolean(mask & 1),
            operator: Boolean(mask & 2), right: Boolean(mask & 4), fold,
        });
    }
}
cases.push(
    { name: "nested both sides", left: true, operator: true, right: true, fold: true, nested: true },
    { name: "object states", left: true, operator: true, right: true, fold: true, nested: true, objects: true },
    { name: "circular child", left: true, operator: true, right: true, fold: true, circular: true },
    { name: "throwing callback", left: true, operator: true, right: true, fold: true, throwing: true },
);

const originalWalker = compile(original);
const adaptedWalker = compile(adapted);
for (const row of cases) {
    const expected = run(originalWalker, row, true);
    const actual = run(adaptedWalker, row, false);
    assert.equal(expected.at(-2)[0], row.circular || row.throwing ? "throw" : "result", row.name);
    assert.equal(JSON.stringify(actual), JSON.stringify(expected), row.name);
    pass(`${row.name}: before=${expected.length} after=${actual.length}`);
    console.log("TRACE before " + JSON.stringify(expected));
    console.log("TRACE after  " + JSON.stringify(actual));
}
console.log(`${passed} rows passed, including ${cases.length} paired walker traces`);
