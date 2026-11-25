"use strict";
// Family 155 patch 1 round 5, read 2 (Performance): probes against the adapter's own -before and -after snapshots.
// Usage: node probes.cjs <applied-tree> <adapter-dir> <out-dir>. Each probe writes a drifted copy of the snapshot to a
// scratch tree, runs the adapter on it, and prints the exit, whether bytes changed, and the adapter's first stderr line.
// Drifted -after files are also kept under <out-dir> so a later step can type-check them with stock tsc.
const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const [tree, adapterDir, outDir] = process.argv.slice(2);
const ref = "refs/stage3/99-seats-scout24-syntactic-result-member";
const relative = "src/compiler/expressionToTypeNode.ts";
const adapter = path.join(adapterDir, "adapt.cjs");
const snapshot = state => execFileSync("git", ["-C", tree, "show", `${ref}-${state}:${relative}`]).toString("utf8");

function run(label, text) {
    const scratch = fs.mkdtempSync(path.join(outDir, "s-"));
    const file = path.join(scratch, relative);
    fs.mkdirSync(path.dirname(file), { recursive: true });
    fs.writeFileSync(file, text);
    const result = spawnSync(process.execPath, [adapter, scratch], { encoding: "utf8" });
    const changed = fs.readFileSync(file, "utf8") !== text;
    const first = (result.stderr.split("\n").find(line => /Error|assert|reviewed|enclosure|postcondition/.test(line)) || "").trim();
    console.log(`PROBE ${label}: exit=${result.status} wrote=${changed} ${first.slice(0, 220)}`);
    return result.status;
}

function edits(text) {
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    const all = [];
    (function visit(node) { all.push(node); ts.forEachChild(node, visit); })(source);
    const exact = (kind, expected) => {
        const found = all.filter(node => kind(node) && node.getText(source) === expected);
        assert.equal(found.length, 1, expected);
        return found[0];
    };
    const after = (node, added) => text.slice(0, node.end) + added + text.slice(node.end);
    const replace = (node, edited) => text.slice(0, node.getStart(source)) + edited + text.slice(node.end);
    const producer = exact(ts.isExpressionStatement, "resultType = syntacticResult(serializeTypeAnnotationOfDeclaration(declaredType, context, node, symbol));");
    const stringCall = "syntacticResult(factory.createKeywordTypeNode(SyntaxKind.StringKeyword))";
    const call = exact(ts.isCallExpression, stringCall);
    // Where is the first `const result = ...` holding a syntacticResult result, if any, relative to the first read?
    const firstRead = exact(ts.isConditionalExpression, "result.type !== undefined ? result.type : inferExpressionType(expr, context, result.reportFallback)");
    let holder = firstRead;
    while (holder && !ts.isFunctionLike(holder)) holder = holder.parent;
    const consts = all.filter(node => ts.isVariableDeclaration(node) && node.name.getText(source) === "result" && node.pos >= holder.pos && node.end <= holder.end);
    console.log(`INFO first read at line ${source.getLineAndCharacterOfPosition(firstRead.getStart(source)).line + 1}; its function's result declarations: ${consts.map(node => JSON.stringify(node.getText(source).slice(0, 160))).join(" | ") || "none"}`);
    return {
        // (1) a write to the type of a result made by a two-argument call with false: {TypeNode, false} under a first member claiming true.
        typeWrite: after(producer, "\r\n            alreadyReported.type = resultType.type;"),
        // (2) a read by destructuring, outside any guard.
        destructure: after(producer, "\r\n            const { reportFallback: observed } = resultType;\r\n            void observed;"),
        // (3) a pinned call's argument changed: only the statement pins see it.
        argument: replace(call, stringCall.replace("StringKeyword", "NumberKeyword")),
        // Seats' round 6 question: an unannotated binding from a result-returning call, whose type is written.
        extra: after(exact(ts.isVariableStatement, all.find(n => ts.isVariableStatement(n) && /^const result = typeFromExpression\(expr, context,/.test(n.getText(source))).getText(source)),
            "\r\n        const extra = typeFromExpression(expr, context);\r\n        extra.type = factory.createKeywordTypeNode(SyntaxKind.AnyKeyword);"),
    };
}

fs.mkdirSync(outDir, { recursive: true });
for (const state of ["before", "after"]) {
    const drifted = edits(snapshot(state));
    for (const [name, text] of Object.entries(drifted)) {
        run(`${state} ${name}`, text);
        if (state === "after") fs.writeFileSync(path.join(outDir, `after-${name}.ts`), text);
    }
}
