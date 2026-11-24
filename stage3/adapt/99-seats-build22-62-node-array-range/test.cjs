"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

const [tree] = process.argv.slice(2);
assert(process.argv.length === 3 && ts.version === "6.0.3", "usage: node test.cjs <applied tree>");
const adapter = path.join(__dirname, "adapt.cjs");
const refs = "refs/stage3/99-seats-build22-62-node-array-range";
const visitor = "src/compiler/visitorPublic.ts";
const utility = "src/compiler/utilities.ts";
const publicUtility = "src/compiler/factory/utilitiesPublic.ts";

function git(...args) {
    return execFileSync("git", ["-C", tree, ...args]);
}

function declaration(file, text, name) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const matches = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.body && node.name?.text === name);
    assert.equal(matches.length, 1, name);
    return matches[0].getText(source).replace(/^export /, "");
}

function run(worker, setters, change) {
    const trace = [];
    const parameter = {};
    const updated = {};
    const parameters = [parameter];
    parameters.hasTrailingComma = true;
    Object.defineProperties(parameters, {
        pos: { get() { trace.push("read-pos"); return 2; } },
        end: { get() { trace.push("read-end"); return 9; } },
    });
    let created, pos, end;
    const context = {
        factory: {
            createNodeArray(nodes, trailing) {
                trace.push("factory");
                assert.deepEqual(Array.from(nodes), [updated]);
                assert.equal(trailing, true);
                created = Array.from(nodes);
                Object.defineProperties(created, {
                    pos: { get() { return pos; }, set(value) { trace.push("write-pos"); pos = value; } },
                    end: { get() { return end; }, set(value) { trace.push("write-end"); end = value; } },
                });
                return created;
            },
        },
    };
    const js = ts.transpileModule([...setters, worker, "addDefaultValueAssignmentsIfNeeded;"].join("\n"), {
        compilerOptions: { target: ts.ScriptTarget.ES2020 },
    }).outputText;
    const fn = vm.runInNewContext(js, {
        addDefaultValueAssignmentIfNeeded(value) { return change ? updated : value; },
    });
    const returned = fn(parameters, context);
    if (!change) {
        assert.equal(returned, parameters, "no-result branch must return parameters");
        assert.deepEqual(trace, [], "no-result branch must skip factory and range operations");
        return;
    }
    assert.ok(created, "worker must call the factory");
    assert.equal(returned, created, "worker must return the factory array");
    assert.deepEqual(trace, ["factory", "read-pos", "read-end", "write-pos", "write-end"]);
    const endpoints = [returned.pos, returned.end];
    assert.deepEqual(endpoints, [2, 9], "range endpoints");
    return { trace, endpoints };
}

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-range-output-"));
try {
    const inputs = new Map();
    for (const file of [visitor, utility, publicUtility]) {
        const bytes = git("show", refs + "-before:" + file);
        const target = path.join(scratch, file);
        fs.mkdirSync(path.dirname(target), { recursive: true });
        fs.writeFileSync(target, bytes);
        inputs.set(file, bytes.toString("utf8"));
    }
    const apply = () => JSON.parse(execFileSync(process.execPath, [adapter, scratch], { encoding: "utf8" }));
    assert.deepEqual(apply(), { edits: 1 });
    const output = fs.readFileSync(path.join(scratch, visitor));
    const setters = [declaration(publicUtility, inputs.get(publicUtility), "setTextRange")];
    for (const name of ["setTextRangePosEnd", "setTextRangePos", "setTextRangeEnd"]) {
        setters.push(declaration(utility, inputs.get(utility), name));
    }
    const name = "addDefaultValueAssignmentsIfNeeded";
    const before = declaration(visitor, inputs.get(visitor), name);
    const after = declaration(visitor, output.toString("utf8"), name);
    assert.deepEqual(run(after, setters, true), run(before, setters, true));
    run(before, setters, false);
    run(after, setters, false);
    const changed = git("diff", "--name-only", refs + "-before", refs + "-after").toString().trim();
    assert.equal(changed, visitor);
    assert.deepEqual(output, git("show", refs + "-after:" + visitor), "visitor differs from after snapshot");
    assert.deepEqual(apply(), { edits: 0 });
    assert.deepEqual(fs.readFileSync(path.join(scratch, visitor)), output);
    for (const file of [utility, publicUtility]) {
        assert.equal(fs.readFileSync(path.join(scratch, file), "utf8"), inputs.get(file));
    }
    console.log("PASS: 1 edit matches the after snapshot, then 0 with identical bytes; factory identity, range trace, endpoints, and no-result identity");
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
