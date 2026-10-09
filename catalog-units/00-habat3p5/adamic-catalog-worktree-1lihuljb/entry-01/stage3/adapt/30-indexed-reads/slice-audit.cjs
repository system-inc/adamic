#!/usr/bin/env node
"use strict";
// Audit original source positions, not indices renumbered by declaration gathering.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { files } = require("./adapt.cjs");
const sites = require("./sites.json");
const [treeArgument, manifestArgument, outputArgument] = process.argv.slice(2);
if (!treeArgument || !manifestArgument || !outputArgument) throw new Error("usage: slice-audit.cjs <tree> <slice.json> <report.json>");
if (ts.version !== "6.0.3") throw new Error("stock TypeScript 6.0.3 required");
const tree = path.resolve(treeArgument), compiler = path.join(tree, "src/compiler");
const manifest = JSON.parse(fs.readFileSync(manifestArgument, "utf8"));
if (path.resolve(manifest.summary.tree) !== tree) throw new Error("manifest source tree differs");
const program = ts.createProgram(ts.sys.readDirectory(compiler, [".ts"]), {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler,
    target: ts.ScriptTarget.ES2024, noEmit: true, types: [], lib: ["lib.es2024.d.ts"],
});
// Ownership supplied by the user after the third-wave audit.
function ownerOf(file) {
    if (files.includes(file)) return "30";
    if (file === "checker.ts") return "31";
    if (["emitter.ts", "sourcemap.ts", "transformer.ts", "visitorPublic.ts"].includes(file) || file.startsWith("transformers/")) return "33";
    if (path.dirname(file) === "." && file.endsWith(".ts")) return "32";
    return "unassigned";
}
const checker = program.getTypeChecker(), rows = [];
for (const file of [...new Set(manifest.declarations.map(record => record.file))].sort()) {
    const source = program.getSourceFile(path.join(tree, file));
    if (!source) throw new Error("missing source: " + file);
    const relative = path.relative(compiler, source.fileName);
    const spans = manifest.declarations.filter(record => record.file === file);
    for (const span of spans) {
        const bytes = source.text.slice(span.start, span.end);
        if (crypto.createHash("sha256").update(bytes).digest("hex") !== span.sha256) throw new Error("source span drift: " + file + ":" + span.line);
    }
    const groups = new Map(), nodes = [];
    function visit(node) {
        if (ts.isElementAccessExpression(node)) {
            const expression = node.getText(source).replaceAll("!", "");
            const group = groups.get(expression) || [];
            group.push(node); groups.set(expression, group); nodes.push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    const reads = [];
    for (const node of nodes) {
        if (!spans.some(span => node.getStart(source) >= span.start && node.end <= span.end)) continue;
        const expression = node.getText(source).replaceAll("!", "");
        const occurrence = groups.get(expression).indexOf(node) + 1;
        const site = sites.find(site => site.file === relative && site.expression === expression && site.occurrence === occurrence);
        const valueType = checker.getTypeAtLocation(node);
        const type = checker.typeToString(valueType);
        const admitsUndefined = !!(valueType.flags & ts.TypeFlags.Undefined) ||
            valueType.isUnion() && valueType.types.some(member => member.flags & ts.TypeFlags.Undefined);
        const parent = node.parent;
        const location = source.getLineAndCharacterOfPosition(node.getStart(source));
        let status;
        if (site) {
            if (site.action === "assert" && !ts.isNonNullExpression(parent)) throw new Error("required assertion missing: " + file + ":" + (location.line + 1));
            status = "30:" + site.action;
        }
        else if (ts.isNonNullExpression(parent)) status = "upstream assertion";
        else if (ts.isBinaryExpression(parent) && parent.left === node && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken) status = "pure store";
        else if (!admitsUndefined) status = "checker does not require presence";
        else status = files.includes(relative) ? "unreviewed in 30" : "owner evidence required";
        reads.push({ line: location.line + 1, column: location.character + 1, expression, occurrence, type, status });
    }
    rows.push({ file, owner: ownerOf(relative), indexed_expressions: reads.length, reads });
}
const unresolved = rows.flatMap(row => row.reads.filter(read => ["unreviewed in 30", "owner evidence required"].includes(read.status)).map(read => ({file: row.file, ...read})));
// Every external indexed expression is handed to its owner, including stores
// and reads whose types already permit the original use.
const ownerSites = Object.fromEntries(["31", "32", "33", "unassigned"].map(owner => [owner,
    rows.filter(row => row.owner === owner).flatMap(row => row.reads.map(read => ({file: row.file, ...read}))),
]));
const report = { tool: "stock TypeScript 6.0.3", entries: manifest.summary.entries, files: rows, owner_sites: ownerSites, unresolved };
fs.writeFileSync(outputArgument, JSON.stringify(report, null, 2) + "\n");
console.log(JSON.stringify({files: rows.length, indexed_expressions: rows.reduce((sum,row) => sum + row.indexed_expressions, 0), unresolved: unresolved.length}));
if (unresolved.length) process.exitCode = 1;
