#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { relative, before, after, plan } = require("./adapt.cjs");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 4, "usage: node test.cjs <tree> <output.json>");
const [, , tree, output] = process.argv;
const ref = "refs/stage3/99-z-seats-38m7dap-decoder-error-once";
const git = (...args) => execFileSync("git", ["-C", tree, ...args], { encoding: "utf8" });
const base = git("show", `${ref}-before:${relative}`);
const head = git("show", `${ref}-after:${relative}`);
const rows = [];
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-38m7dap-"));
const filename = path.join(scratch, relative);
const adapter = path.join(__dirname, "adapt.cjs");

function row(name, check) {
    try {
        rows.push({ name, status: "PASS", evidence: check() });
        console.log(`PASS ${name}`);
    } catch (error) {
        rows.push({ name, status: "FAIL", message: error.message });
        throw error;
    }
}

function replaceOnce(text, original, replacement) {
    assert.equal(text.split(original).length, 2, "expected one replacement anchor");
    return text.replace(original, () => replacement);
}

function inspect(text) {
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const owner = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "createDocumentPositionMapper");
    const mapper = owner.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "getDecodedMappings");
    const reads = [];
    const localUses = [];
    const log = [];
    function visit(node) {
        const text = node.getText(source);
        if (ts.isPropertyAccessExpression(node) && text === "decoder.error") reads.push(node);
        if (ts.isPropertyAccessExpression(node) && text === "host.log") {
            const parent = node.parent;
            log.push({ text, role: ts.isIfStatement(parent) ? "guard" : ts.isCallExpression(parent) ? "callee" : "other" });
        }
        if (ts.isIdentifier(node) && text === "decoderError" && !(ts.isVariableDeclaration(node.parent) && node.parent.name === node)) {
            localUses.push(ts.isBinaryExpression(node.parent) ? "test" : ts.isTemplateSpan(node.parent) ? "template" : "other");
        }
        ts.forEachChild(node, visit);
    }
    visit(mapper);
    const declaration = reads[0]?.parent;
    const initializer = reads.length === 1 && ts.isVariableDeclaration(declaration) && declaration.initializer === reads[0]
        && declaration.name.getText(source) === "decoderError" && (declaration.parent.flags & ts.NodeFlags.Const) !== 0;
    const block = mapper.body.statements[0].thenStatement;
    const index = initializer ? block.statements.indexOf(declaration.parent.parent) : -1;
    return {
        reads: reads.length,
        originalUses: reads.map(node => ts.isBinaryExpression(node.parent) ? "test" : ts.isTemplateSpan(node.parent) ? "template" : "other"),
        initializer,
        localUses,
        previous: index > 0 ? block.statements[index - 1].getText(source) : null,
        following: index >= 0 ? block.statements[index + 1]?.expression?.getText(source) : null,
        log,
    };
}

function spawnAdapter() {
    const run = spawnSync(process.execPath, [adapter, scratch], { encoding: "utf8" });
    assert.ifError(run.error);
    assert.equal(run.signal, null);
    return run;
}

function apply(edits, expected) {
    const run = spawnAdapter();
    assert.equal(run.status, 0, run.stderr);
    assert.deepEqual(JSON.parse(run.stdout), { adapter: "38m7dap", file: relative, edits });
    assert(fs.readFileSync(filename).equals(Buffer.from(expected)), "adapter changed unexpected bytes");
    return { edits };
}

function drift(name, text, message) {
    row(name, () => {
        assert.throws(() => plan(text), message);
        fs.writeFileSync(filename, text);
        const run = spawnAdapter();
        assert.equal(run.status, 1, "adapter must refuse drift");
        assert.match(run.stderr, message);
        assert(fs.readFileSync(filename).equals(Buffer.from(text)), "rejected input changed");
        return { exit: run.status, check: message.source };
    });
}

function load(text) {
    const { outputText, diagnostics } = ts.transpileModule(text, {
        fileName: relative,
        compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS },
        reportDiagnostics: true,
    });
    assert.equal(diagnostics.length, 0, ts.formatDiagnosticsWithColorAndContext(diagnostics, {
        getCurrentDirectory: () => "", getCanonicalFileName: name => name, getNewLine: () => "\n",
    }));
    const exports = {};
    vm.runInNewContext(outputText, {
        exports,
        require(name) {
            if (name === "./_namespaces/ts.js") return ts;
            assert.equal(name, "./_namespaces/ts.performance.js");
            return { nullTimer: { enter() {}, exit() {} } };
        },
    }, { filename: "sourcemap.js" });
    return exports;
}

const invalid = "Invalid character in VLQ";
function decoderRecord(decode) {
    const text = "A,?AAA";
    let inputReads = 0;
    const input = {
        get length() { inputReads++; return text.length; },
        charCodeAt(index) { inputReads++; return text.charCodeAt(index); },
    };
    const decoder = decode(input);
    const { get } = Object.getOwnPropertyDescriptor(decoder, "error");
    let getterCalls = 0;
    function read() {
        getterCalls++;
        return Reflect.apply(get, decoder, []);
    }
    const drained = [...decoder].length;
    const afterDrain = inputReads;
    const first = read();
    const afterFirst = inputReads;
    const { value, done } = decoder.next();
    const second = read();
    return { drained, afterDrain, first, afterFirst, next: { value, done }, second, afterSecond: inputReads, getterCalls };
}

class HostLogFetchError extends Error {}
class HostLogCallError extends Error {}
const prefix = "Encountered error while decoding sourcemap: ";
const pastEnd = "Error in decoding base64VLQFormatDecode, past the mapping string";

function mapperRecord(create, { valid, reenter, fetchThrows, callThrows, noLog }) {
    const events = [];
    const messages = [];
    let mappings = valid ? "AAAA" : "AAAA,?";
    let mapper;
    let fetches = 0;
    let calls = 0;
    const location = { fileName: "/out.js", pos: 0 };
    const map = {
        version: 3, file: "out.js", sources: ["input.ts"], names: [],
        get mappings() {
            events.push(`get map.mappings = ${JSON.stringify(mappings)}`);
            return mappings;
        },
    };
    const host = {
        get getSourceFileLike() {
            events.push("get host.getSourceFileLike");
            return file => {
                events.push(`call host.getSourceFileLike(${file})`);
                return valid ? ts.createSourceFile(file, "x", ts.ScriptTarget.Latest) : undefined;
            };
        },
        get getCanonicalFileName() {
            events.push("get host.getCanonicalFileName");
            return file => {
                events.push(`call host.getCanonicalFileName(${file})`);
                return file;
            };
        },
        get log() {
            const read = ++fetches;
            events.push(`get host.log #${read}`);
            if (read === fetchThrows) throw new HostLogFetchError();
            if (read === reenter) {
                mappings = "g";
                events.push("reenter with changed mappings");
                assert.equal(mapper.getSourcePosition(location), location);
                events.push("return from reentry");
            }
            if (noLog) return undefined;
            return function log(message) {
                assert.equal(this, host);
                calls++;
                events.push(`call host.log from read #${read}: ${message}`);
                messages.push(message);
                if (calls === callThrows) throw new HostLogCallError();
            };
        },
    };
    mapper = create(host, map, "/out.js.map");
    const throws = ["lookup", "lookup again"].map(lookup => {
        events.push(lookup);
        try {
            const { fileName, pos } = mapper.getSourcePosition(location);
            events.push({ returned: { fileName, pos } });
            return null;
        } catch (error) {
            return { class: error.constructor.name, at: events.length, after: events.at(-1) };
        }
    });
    events.push(`total host.log reads=${fetches} calls=${calls}`);
    return { events, messages, fetches, calls, throws };
}

const scout = [
    "get host.getSourceFileLike",
    "call host.getSourceFileLike(/out.js)",
    "get host.getCanonicalFileName",
    "call host.getCanonicalFileName(/input.ts)",
    "lookup",
    'get map.mappings = "AAAA,?"',
    "get host.getSourceFileLike",
    "call host.getSourceFileLike(/input.ts)",
    "get host.log #1",
    "reenter with changed mappings",
    'get map.mappings = "g"',
    "get host.log #2",
    "get host.log #3",
    `call host.log from read #3: ${prefix}${pastEnd}`,
    "return from reentry",
    "get host.log #4",
    `call host.log from read #4: ${prefix}${invalid}`,
    "lookup again",
    "total host.log reads=4 calls=2",
];

try {
    const input = inspect(base);
    const adapted = inspect(head);
    row("S3 decoder read order", () => {
        assert.equal(input.reads, 2);
        assert.deepEqual(input.originalUses, ["test", "template"]);
        assert.equal(adapted.reads, 1);
        assert.equal(adapted.initializer, true);
        assert.equal(adapted.previous, "const mappings = arrayFrom(decoder, processMapping);");
        assert.equal(adapted.following, "decoderError !== undefined");
        assert.deepEqual(adapted.localUses, ["test", "template"]);
        return { input: input.reads, adapted: adapted.reads, localUses: adapted.localUses };
    });
    row("S4 host.log guard and callee", () => {
        assert.deepEqual(input.log, [{ text: "host.log", role: "guard" }, { text: "host.log", role: "callee" }]);
        assert.deepEqual(adapted.log, input.log);
        return input.log;
    });
    row("S1 sourcemap path only", () => {
        const files = git("diff", "--name-only", ref + "-before", ref + "-after").trim().split("\n");
        assert.deepEqual(files, [relative]);
        return files;
    });
    row("S2 exact function replacement", () => {
        assert(replaceOnce(base, before, after) === head, "bytes outside the pinned replacement changed");
    });

    fs.mkdirSync(path.dirname(filename), { recursive: true });
    fs.writeFileSync(filename, base);
    row("A1 adapter writes the after snapshot", () => apply(1, head));
    row("A2 adapter replay", () => apply(0, head));

    const owner = "export function createDocumentPositionMapper";
    const moved = text => replaceOnce(text, owner, "// drift\r\n" + owner);
    drift("A3a original position drift", moved(base), /getDecodedMappings position drift/);
    drift("A3b replay position drift", moved(head), /getDecodedMappings position drift/);
    drift("A3c half edit function text drift", replaceOnce(base, '${decoder.error}', '${decoderError}'), /function text drift/);
    drift("A3d condition function text drift", replaceOnce(base, "if (decoder.error !== undefined)", "if (decoder.error !== void 0)"), /function text drift/);
    drift("A3e duplicate getDecodedMappings declaration", base + "\r\nfunction getDecodedMappings() {}\r\n", /expected one getDecodedMappings declaration/);

    fs.writeFileSync(filename, head);
    const original = load(base);
    const changed = load(head);
    row("decoder drained getter", () => {
        const left = decoderRecord(original.decodeMappings);
        const right = decoderRecord(changed.decodeMappings);
        assert.deepEqual(left, {
            drained: 1, afterDrain: 12, first: invalid, afterFirst: 12,
            next: { value: undefined, done: true }, second: invalid, afterSecond: 12, getterCalls: 2,
        });
        assert.deepEqual(right, left);
        return { before: left, after: right };
    });

    for (const [name, options, fetches, calls, last, thrown] of [
        ["re-entry", { reenter: 1 }, 4, 2, 4],
        ["one error", {}, 2, 1, 2],
        ["no log", { noLog: true }, 1, 0],
        ["valid", { valid: true }, 0, 0],
        ["guard fetch throws", { fetchThrows: 1 }, 3, 1, 3, ["HostLogFetchError", "get host.log #1"]],
        ["callee fetch throws", { fetchThrows: 2 }, 4, 1, 4, ["HostLogFetchError", "get host.log #2"]],
        ["log call throws", { callThrows: 1 }, 4, 2, 4, ["HostLogCallError", `call host.log from read #2: ${prefix}${invalid}`]],
        ["re-entry throws", { reenter: 1, fetchThrows: 2 }, 4, 1, 4, ["HostLogFetchError", "get host.log #2"]],
    ]) {
        row(`mapper ${name}`, () => {
            const left = mapperRecord(original.createDocumentPositionMapper, options);
            const right = mapperRecord(changed.createDocumentPositionMapper, options);
            if (name === "re-entry") assert.deepEqual(left.events.filter(event => typeof event === "string"), scout);
            assert.equal(left.fetches, fetches);
            assert.equal(left.calls, calls);
            if (!thrown) {
                assert.equal(left.events.slice(left.events.indexOf("lookup again") + 1, -1).length, 1,
                    "cached second lookup must only record its return");
            }
            const returns = left.events.filter(event => typeof event === "object");
            assert.equal(returns.length, thrown ? 1 : 2);
            if (options.valid) assert.deepEqual(returns.map(event => event.returned), [
                { fileName: "/input.ts", pos: 0 }, { fileName: "/input.ts", pos: 0 },
            ]);
            const lastError = name === "re-entry throws" ? pastEnd : invalid;
            if (last) assert.equal(left.events.filter(event => typeof event === "string" && event.startsWith("call host.log")).at(-1),
                `call host.log from read #${last}: ${prefix}${lastError}`);
            assert.deepEqual(left.throws, [
                thrown ? { class: thrown[0], at: left.events.indexOf(thrown[1]) + 1, after: thrown[1] } : null, null,
            ]);
            assert.deepEqual(right, left);
            return { before: left, after: right };
        });
    }
    row("second adapter call preserved bytes", () => apply(0, head));
} finally {
    fs.writeFileSync(output, JSON.stringify({ rows }, (_, value) => value === undefined ? { undefined: true } : value, 2) + "\n");
    fs.rmSync(scratch, { recursive: true, force: true });
}
