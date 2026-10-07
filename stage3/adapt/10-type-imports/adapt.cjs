#!/usr/bin/env node
"use strict";

// Use the checker's decisions, as cmd/adamic-meter does, but validate the
// specifiers with the AST and insert only the inline modifier into real text.
const fs = require("node:fs");
const path = require("node:path");

function main() {
    if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
    const tree = path.resolve(process.argv[2]);
    const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
    if (ts.version !== "6.0.3") throw new Error(`want stock typescript@6.0.3, got ${ts.version}`);
    const directory = path.join(tree, "src/compiler");
    const roots = ts.sys.readDirectory(directory, [".ts"], undefined, ["**/*"]).sort();
    if (!roots.length) throw new Error(`no compiler sources in ${directory}`);
    const options = {
        strict: true,
        exactOptionalPropertyTypes: true,
        noUncheckedIndexedAccess: true,
        noImplicitReturns: true,
        noFallthroughCasesInSwitch: true,
        verbatimModuleSyntax: true,
        erasableSyntaxOnly: true,
        allowImportingTsExtensions: true,
        noEmit: true,
        module: ts.ModuleKind.ESNext,
        moduleDetection: ts.ModuleDetectionKind.Force,
        moduleResolution: ts.ModuleResolutionKind.Bundler,
        target: ts.ScriptTarget.ES2024,
        lib: ["lib.es2024.d.ts"],
        types: [],
    };
    const program = ts.createProgram(roots, options);
    const syntax = program.getSyntacticDiagnostics();
    if (syntax.length) throw new Error(ts.formatDiagnosticsWithColorAndContext(syntax, {
        getCanonicalFileName: name => name,
        getCurrentDirectory: () => tree,
        getNewLine: () => "\n",
    }));
    const owned = new Set(roots.map(name => path.resolve(name)));
    const edits = new Map();
    const declined = [];
    let imports = 0;
    let exports = 0;
    for (const diagnostic of program.getSemanticDiagnostics()) {
        if (diagnostic.code !== 1484 && diagnostic.code !== 1205) continue;
        const source = diagnostic.file;
        if (!source || !owned.has(path.resolve(source.fileName))) {
            declined.push({ code: diagnostic.code, file: source?.fileName, reason: "outside src/compiler" });
            continue;
        }
        let specifier;
        function visit(node) {
            if (diagnostic.start < node.getFullStart() || diagnostic.start >= node.end) return;
            if (ts.isImportSpecifier(node) || ts.isExportSpecifier(node)) specifier = node;
            ts.forEachChild(node, visit);
        }
        visit(source);
        const isImport = specifier && ts.isImportSpecifier(specifier);
        const declaration = specifier?.parent.parent;
        const valid = specifier && !specifier.isTypeOnly && (
            diagnostic.code === 1484 && isImport && ts.isImportClause(declaration) && !declaration.isTypeOnly ||
            diagnostic.code === 1205 && !isImport && ts.isExportDeclaration(declaration) &&
                !declaration.isTypeOnly && declaration.moduleSpecifier
        );
        if (!valid) {
            const where = source.getLineAndCharacterOfPosition(diagnostic.start);
            declined.push({ code: diagnostic.code, file: source.fileName, line: where.line + 1,
                reason: "requires a named inline specifier on an import or re-export" });
            continue;
        }
        const at = specifier.getStart(source);
        let positions = edits.get(source);
        if (!positions) edits.set(source, positions = new Set());
        if (!positions.has(at)) {
            positions.add(at);
            if (isImport) imports++;
            else exports++;
        }
    }
    // Plan every edit before writing. No printer, sorting, line-ending conversion,
    // declaration deletion, or whole-declaration import type conversion occurs.
    for (const [source, positions] of edits) {
        let text = source.text;
        for (const at of [...positions].sort((a, b) => b - a)) {
            text = text.slice(0, at) + "type " + text.slice(at);
        }
        if (fs.readFileSync(source.fileName, "utf8") !== source.text) {
            throw new Error(`source changed while planning: ${source.fileName}`);
        }
        fs.writeFileSync(source.fileName, text);
    }
    console.log(JSON.stringify({ files: edits.size, imports, exports, declined }, null, 2));
    if (declined.length) process.exitCode = 1;
}

try { main(); }
catch (error) { console.error(error.message); process.exitCode = 1; }
