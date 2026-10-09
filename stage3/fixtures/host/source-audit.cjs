#!/usr/bin/env node
"use strict";
// Hold extracted declarations to the adapted source, independently of their goldens.
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const tree = path.resolve(process.argv[2]);
const bucket = __dirname;
const printer = ts.createPrinter({ newLine: ts.NewLineKind.LineFeed, removeComments: true });
function tokens(node) {
    return printer.printNode(ts.EmitHint.Unspecified, node, node.getSourceFile());
}
const wanted = new Set();
for (const file of fs.readdirSync(bucket).filter(name => /^\d\d_.*\.a$/.test(name))) {
    const source = ts.createSourceFile(file, fs.readFileSync(path.join(bucket, file), "utf8"), ts.ScriptTarget.Latest, true);
    function visit(node) {
        if ((ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node)) && node.name) wanted.add(node.name.getText(source));
        ts.forEachChild(node, visit);
    }
    visit(source);
}

const originals = new Map();
for (const file of ts.sys.readDirectory(path.join(tree, "src/compiler"), [".ts"])) {
    const source = ts.createSourceFile(file, fs.readFileSync(file, "utf8"), ts.ScriptTarget.Latest, true);
    function visit(node) {
        if ((ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node)) && node.body && node.name) {
            const name = node.name.getText(source);
            if (!wanted.has(name)) { ts.forEachChild(node, visit); return; }
            const list = originals.get(name) || [];
            list.push({ tokens: tokens(node), node });
            originals.set(name, list);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}
function audit(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file);
    let count = 0;
    function visit(node) {
        if ((ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node)) && node.body && node.name) {
            const name = node.name.getText(source);
            if (name === "disableCPUProfiler") {
                // These three exit fixtures intentionally keep only the inactive arm.
                const original = originals.get(name)[0].node;
                const last = original.body.statements[0];
                assert.ok(ts.isIfStatement(last));
                assert.equal(tokens(node.body.statements[0].elseStatement), tokens(last.elseStatement));
            }
            else {
                assert.ok(originals.get(name)?.some(original => original.tokens === tokens(node)), file + ': changed upstream function ' + name);
            }
            count++;
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    return count;
}
const manifest = JSON.parse(fs.readFileSync(path.join(bucket, "source-spans.json"), "utf8"));
function canonical(node) { return tokens(node).replace(/^export /, ""); }
function auditDeclarations(file, text) {
    const row = manifest.fixtures.find(row => row.file === file);
    assert.ok(row, file + ": missing source manifest");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const ranges = [];
    const recorded = row.spans.filter(span => span.tokens && !span.partial);
    const found = new Set();
    function collect(node) {
        if (ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node) ||
            ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node) || ts.isVariableStatement(node)) {
            const value = canonical(node);
            for (const span of recorded) {
                if (span.tokens !== value) continue;
                const original = fs.readFileSync(path.join(tree, span.file), "utf8");
                const upstream = ts.createSourceFile(span.file, original, ts.ScriptTarget.Latest, true);
                let agrees = false;
                function seek(candidate) {
                    if (candidate.kind === node.kind && canonical(candidate) === value) agrees = true;
                    ts.forEachChild(candidate, seek);
                }
                seek(upstream);
                assert.ok(agrees, file + ": declaration differs from adapted source: " + span.name);
                ranges.push([node.getStart(source), node.end]);
                found.add(span);
            }
        }
        ts.forEachChild(node, collect);
    }
    collect(source);
    assert.equal(found.size, recorded.length, file + ": missing or changed copied declaration");
    function guard(node) {
        if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) || ts.isNonNullExpression(node) || node.kind === ts.SyntaxKind.AnyKeyword) {
            assert.ok(ranges.some(([start, end]) => start <= node.getStart(source) && node.end <= end),
                file + ": reduction introduced a cast, non-null assertion, or any: " + node.getText(source));
        }
        ts.forEachChild(node, guard);
    }
    guard(source);
    return found.size;
}
let declarationCount = 0;
for (const file of fs.readdirSync(bucket).filter(name => /^\d\d_.*\.a$/.test(name)).sort()) {
    declarationCount += auditDeclarations(file, fs.readFileSync(path.join(bucket, file), "utf8"));
}
console.log(`pass: ${declarationCount} copied declarations match adapted source; no reduction-added casts, non-null assertions or any`);
const driverFile = "05_writeFile.a";
const driverSource = fs.readFileSync(path.join(bucket, driverFile), "utf8");
assert.throws(() => auditDeclarations(driverFile, driverSource.replace("errorCode(e)", "(e as NodeJS.ErrnoException).code")), { code: "ERR_ASSERTION" });
console.log("caught mutant: reduction-only ErrnoException cast rejected");
let count = 0;
for (const file of fs.readdirSync(bucket).filter(name => /^\d\d_.*\.a$/.test(name)).sort()) {
    count += audit(file, fs.readFileSync(path.join(bucket, file), "utf8"));
}
console.log(`pass: ${count} upstream function and method spans retain their tokens`);
const file = "21_createHash.a";
const original = fs.readFileSync(path.join(bucket, file), "utf8");
assert.throws(() => audit(file, original.replace('createHash("sha256")', 'createHash("sha1")')), { code: "ERR_ASSERTION" });
console.log("caught mutant: SHA256 call changed to SHA1, rejected by upstream function tokens");
