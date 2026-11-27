"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const ts = require("typescript");

const fileName = "src/compiler/types.ts";
const adapter = path.join(__dirname, "adapt.cjs");
const ref = "refs/stage3/99-seats-scout24-pragma-map-foreach";
assert.equal(process.argv.length, 3, "usage: node test.cjs <applied tree>");
const applied = process.argv[2];

function snapshot(side) {
    return execFileSync("git", ["-C", applied, "show", `${ref}-${side}:${fileName}`], { encoding: "utf8" });
}

function declaration(text, name) {
    const file = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    const found = file.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === name);
    assert.equal(found.length, 1, name);
    return found[0];
}

function method(owner) {
    const found = owner.members.filter(node => ts.isMethodSignature(node) && ts.isIdentifier(node.name) && node.name.text === "forEach");
    assert.equal(found.length, 1);
    return found[0];
}

const before = snapshot("before");
const after = snapshot("after");
const readonly = method(declaration(before, "ReadonlyPragmaMap"));
const trivia = before.slice(readonly.getFullStart(), readonly.name.getStart());
assert.equal(trivia, "\r\n    ");
const suffix = before.slice(readonly.name.end, readonly.end);
const insert = declaration(after, "ReadonlyPragmaMap").members.at(-1).end;
const mapName = method(declaration(before, "PragmaMap")).name;
const tree = fs.mkdtempSync(path.join(os.tmpdir(), "pragma-foreach-"));
const scratch = path.join(tree, fileName);
fs.mkdirSync(path.dirname(scratch), { recursive: true });

function attempt(input) {
    fs.writeFileSync(scratch, input);
    // An old timestamp makes a rewrite detectable even with coarse clock resolution.
    fs.utimesSync(scratch, new Date(1000), new Date(1000));
    const mtime = fs.statSync(scratch, { bigint: true }).mtimeNs;
    const result = spawnSync(process.execPath, [adapter, tree], { encoding: "utf8" });
    assert.ifError(result.error);
    assert.notEqual(result.status, null, result.signal);
    return { ...result, contents: fs.readFileSync(scratch), mtime, afterTime: fs.statSync(scratch, { bigint: true }).mtimeNs };
}

function rejects(label, input, reason) {
    const result = attempt(input);
    assert.notEqual(result.status, 0, label);
    assert.match(result.stderr, reason);
    assert.deepEqual(result.contents, Buffer.from(input), label);
    console.log(`PASS ${label} refused without changing bytes`);
}

try {
    const fresh = attempt(before);
    assert.equal(fresh.status, 0, fresh.stderr);
    assert.deepEqual(JSON.parse(fresh.stdout), { files: 1, edits: 2 });
    assert.deepEqual(fresh.contents, Buffer.from(after));
    console.log("PASS (a) before produces the exact after snapshot");

    const replay = attempt(after);
    assert.equal(replay.status, 0, replay.stderr);
    assert.deepEqual(JSON.parse(replay.stdout), { files: 0, edits: 0 });
    assert.deepEqual(replay.contents, Buffer.from(after));
    assert.equal(replay.afterTime, replay.mtime);
    console.log("PASS (b) after reports zero edits with identical bytes and mtime");

    for (const spelling of ["forEach", '"forEach"', "'forEach'", '["forEach"]']) {
        const input = after.slice(0, insert) + trivia + spelling + suffix + after.slice(insert);
        const reason = {
            forEach: /unexpected PragmaMap\.forEach/,
            '"forEach"': /unexpected ReadonlyPragmaMap\.forEach/,
            "'forEach'": /unexpected ReadonlyPragmaMap\.forEach/,
            '["forEach"]': /computed member name in ReadonlyPragmaMap/,
        }[spelling];
        rejects(`(c) ${spelling}`, input, reason);
    }

    const quoted = before.slice(0, mapName.getStart()) + '"forEach"' + before.slice(mapName.end);
    rejects('(d) PragmaMap "forEach"', quoted, /unexpected PragmaMap\.forEach/);
} finally {
    fs.rmSync(tree, { recursive: true, force: true });
}
