"use strict";

const { spawnSync } = require("node:child_process");
const { existsSync, mkdtempSync, readFileSync, rmSync } = require("node:fs");
const { tmpdir } = require("node:os");
const path = require("node:path");

const snapshotRef = "refs/stage3/99-family34-local-stubs-after";
const coordinate = /((?:[\w.@-]+\/)+[\w.@-]+\.[cm]?[jt]sx?):\d+(?::\d+)?/g;

function run(command, args, label) {
    const result = spawnSync(command, args);
    if (result.error) throw new Error(label + ": " + result.error.message);
    if (result.signal || result.status !== 0) {
        process.stderr.write(result.stderr);
        throw new Error(label + " failed with " + (result.signal ?? result.status));
    }
    return result.stdout;
}

function checkTrace(name, stdout) {
    const expected = readFileSync(path.join(__dirname, name + ".expected.txt"));
    // Preserve all bytes outside the ASCII coordinates.
    const actual = Buffer.from(stdout.toString("latin1").replace(coordinate, "$1:<coordinate>"), "latin1");
    if (actual.equals(expected)) return;

    const oldLines = expected.toString("latin1").split("\n");
    const newLines = actual.toString("latin1").split("\n");
    let line = 0;
    while (oldLines[line] === newLines[line]) line++;
    const format = value => value === undefined ? "<end of transcript>" : Buffer.from(value, "latin1").toString("utf8");
    throw new Error(name + ".cjs differs at line " + (line + 1)
        + "\n  expected: " + format(oldLines[line])
        + "\n  actual:   " + format(newLines[line]));
}

function checkExposedStops(lines) {
    const name = "exposed-stops.jsonl";
    const raw = readFileSync(path.join(__dirname, name));
    const text = raw.toString("utf8").trimEnd();
    const findings = text === "" ? [] : text.split("\n").map(line => JSON.parse(line));
    const headings = lines.filter(line => line === "## Exposed stops").length;
    if (headings !== 1) throw new Error("README needs exactly one Exposed stops section, found " + headings);
    const section = lines.indexOf("## Exposed stops");
    const following = lines.findIndex((line, index) => index > section && line.startsWith("## "));
    const rows = lines.slice(section + 1, following === -1 ? undefined : following)
        .filter(line => /^\| \d+ \|/.test(line));
    for (const { text, unit, where } of findings) {
        const sourceStart = where.lastIndexOf("/src/");
        const prefix = where.slice(0, sourceStart + 1);
        if (sourceStart < 0 || !unit.startsWith(prefix + "src/") || !text.startsWith(where + ": ")) {
            throw new Error("invalid raw finding location " + where);
        }
        const row = "| 34 | `" + unit.slice(prefix.length) + "` | `" + text.slice(prefix.length)
            + "` | Outside the ruling. Left alone. |";
        const index = rows.indexOf(row);
        if (index === -1) throw new Error("missing README exposed stop " + where.slice(prefix.length));
        rows.splice(index, 1);
    }
    if (rows.length) throw new Error("unexpected README exposed stop row");
    console.log("PASS " + name + " (" + findings.length + " README row(s))");
}

function replay(tree) {
    const lines = readFileSync(path.join(__dirname, "README.md"), "utf8").split(/\r?\n/);
    checkExposedStops(lines);
    const scratch = mkdtempSync(path.join(tmpdir(), "family34-replay-"));
    try {
        const archive = path.join(scratch, "src.tar");
        run("git", ["-C", tree, "archive", "--format=tar", "--output=" + archive, snapshotRef, "src"],
            "git archive " + snapshotRef);
        run("tar", ["-xf", archive, "-C", scratch], "tar extract " + snapshotRef);
        for (const name of ["stub-readers", "slot-to-baseline", "remaining-stubs"]) {
            const script = path.join(__dirname, name + ".cjs");
            if (!existsSync(script)) throw new Error("missing trace script " + name + ".cjs");
            checkTrace(name, run(process.execPath, [script, scratch], name + ".cjs"));
            console.log("PASS " + name + ".cjs");
        }
    } finally {
        rmSync(scratch, { recursive: true, force: true });
    }
}

try {
    if (process.argv.length !== 3) throw new Error("usage: node replay.cjs <tree stage3/apply.sh wrote>");
    replay(path.resolve(process.argv[2]));
} catch (error) {
    console.error("FAIL " + error.message);
    process.exitCode = 1;
}
