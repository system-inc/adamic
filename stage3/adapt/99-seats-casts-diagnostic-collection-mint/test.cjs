"use strict";

const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

assert.equal(process.argv.length, 3, "usage: node test.cjs <applied TypeScript tree>");
const applied = path.resolve(process.argv[2]);
const snapshot = "refs/stage3/99-seats-casts-diagnostic-collection-mint";
const relative = "src/compiler/utilities.ts";
const adapter = path.join(__dirname, "adapt.cjs");
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-diagnostic-mint-"));

function run(command, args) {
    const result = spawnSync(command, args, { encoding: "utf8" });
    if (result.error) throw result.error;
    assert.equal(result.signal, null, command + " terminated by " + result.signal);
    return result;
}

function checked(command, args) {
    const result = run(command, args);
    assert.equal(result.status, 0, result.stdout + result.stderr);
}

function readSnapshot(side) {
    const directory = path.join(scratch, "snapshot-" + side);
    fs.mkdirSync(directory);
    const archive = path.join(directory, "source.tar");
    checked("git", ["-C", applied, "archive", "--format=tar", "--output=" + archive, snapshot + "-" + side, relative]);
    checked("tar", ["-xf", archive, "-C", directory]);
    fs.unlinkSync(archive);
    return fs.readFileSync(path.join(directory, relative));
}

function replaceOnce(text, old, replacement, label) {
    assert.equal(text.split(old).length, 2, label + " expected one replacement anchor");
    return text.replace(old, replacement);
}

function check(label, input, expected, output) {
    try {
        const directory = path.join(scratch, label);
        const file = path.join(directory, relative);
        fs.mkdirSync(path.dirname(file), { recursive: true });
        fs.writeFileSync(file, input);
        const result = run(process.execPath, [adapter, directory]);
        if (output === undefined) {
            assert.notEqual(result.status, 0, "accepted");
            assert(result.stderr.includes(expected), "wrong refusal: " + result.stderr.slice(0, 700));
            assert(fs.readFileSync(file).equals(input), "refused input was written");
        } else {
            assert.equal(result.status, 0, result.stderr.slice(0, 700));
            assert.equal(result.stdout.trim(), output, "adapter counts");
            assert(fs.readFileSync(file).equals(expected), "emitted bytes differ from the after snapshot");
        }
        console.log("ok " + label);
    } catch (error) {
        throw new Error(label + ": " + error.message, { cause: error });
    }
}

try {
    const before = readSnapshot("before");
    const after = readSnapshot("after");
    const text = before.toString("utf8");
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    const cast = "[] as Diagnostic[] as SortedArray<DiagnosticWithLocation>";
    const assignment = "diagnostics = " + cast + ";";
    const comment = " // See GH#19873";
    const line = assignment + comment;
    const exact = "unreviewed createDiagnosticCollection.add cast:";
    const count = "expected one array literal in createDiagnosticCollection.add";
    const owner = "unreviewed createDiagnosticCollection.add cast owner";
    const shadow = "let diagnostics: SortedArray<DiagnosticWithLocation>; ";
    const nextAssignment = expression => line + newline + "                diagnostics = " + expression + ";";
    const rows = [
        ["wrong-target", cast, cast.replace("DiagnosticWithLocation", "Diagnostic"), exact],
        ["wrong-operand", cast, cast.replace("[] as", "[diagnostic] as"), exact],
        ["wrong-inner-cast", cast, cast.replace("Diagnostic[]", "Diagnostic []"), exact],
        ["nonliteral-operand", cast, cast.replace("[] as Diagnostic[]", "Array<Diagnostic>()"), count],
        ["wrong-assignment", assignment, "nonFileDiagnostics = " + cast + ";", "unreviewed createDiagnosticCollection.add cast assignment"],
        ["wrong-owner", "function add(diagnostic: Diagnostic): void", "function add(actual: Diagnostic): void", "unreviewed add header"],
        ["duplicate", line, nextAssignment(cast), count],
        ["parenthesized-duplicate", line, nextAssignment("([] as Diagnostic[]) as SortedArray<DiagnosticWithLocation>"), count],
        ["relocated-owner", line, "function makeLocatedDiagnostics() { " + shadow + assignment
            + " return diagnostics; }" + newline + "                diagnostics = makeLocatedDiagnostics();" + comment, owner],
        ["angle-duplicate", line, nextAssignment("<SortedArray<DiagnosticWithLocation>><Diagnostic[]>[]"), count],
        ["satisfies-duplicate", line, nextAssignment("([] satisfies Diagnostic[]) as SortedArray<DiagnosticWithLocation>"), count],
        ["spaced-duplicate", line, nextAssignment("[] as Diagnostic[] as SortedArray< DiagnosticWithLocation >"), count],
        ["arrow-owner", line, "const make = () => { " + shadow + assignment + " return diagnostics; };"
            + newline + "                diagnostics = make();" + comment, owner],
        ["block-shadow", line, "{ " + shadow + assignment + " }" + comment, "expected one diagnostics declaration"],
        ["class-instance-owner", line, "class C { x = (diagnostics = " + cast + "); }"
            + newline + "                diagnostics = new C().x;" + comment, owner],
        ["class-static-owner", line, "class C { static { " + assignment + " } }" + comment, owner],
        ["class-static-field-owner", line, "class C { static x = (diagnostics = " + cast + "); }"
            + newline + "                diagnostics = C.x;" + comment, owner],
        ["class-expression-owner", line, "const C = class { x = (diagnostics = " + cast + "); };"
            + newline + "                diagnostics = new C().x;" + comment, owner],
    ];
    check("before", before, after, '{"files":1,"edits":1}');
    check("after", after, after, '{"files":0,"edits":0}');
    for (const [label, old, replacement, message] of rows) {
        check(label, Buffer.from(replaceOnce(text, old, replacement, label)), message);
    }
    const total = rows.length + 2;
    console.log(`ok ${total} of ${total}`);
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
