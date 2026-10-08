#!/usr/bin/env node
throw new Error("deliberate break: the fast gate kill test for stage 3 coverage");
"use strict";

// Use the checker's decisions, as cmd/adamic-meter does, but validate the
// specifiers with the AST and insert only the inline modifier into real text.
const fs = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

// The generator owns this import. Parse its output string as TypeScript, then
// classify its names against the actual exports of compiler/types.ts.
function generatorEdits(ts, tree, program) {
    const name = path.join(tree, "scripts/processDiagnosticMessages.mjs");
    const source = ts.createSourceFile(name, fs.readFileSync(name, "utf8"), ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse generator: ${name}`);
    const checker = program.getTypeChecker();
    const types = program.getSourceFile(path.join(tree, "src/compiler/types.ts"));
    const symbol = types && checker.getSymbolAtLocation(types);
    if (!symbol) throw new Error("cannot resolve compiler/types.ts exports");
    const symbols = new Map(checker.getExportsOfModule(symbol).map(item => [item.name, item]));
    const positions = new Set();
    let templates = 0;
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === "buildInfoFileOutput") {
            function findTemplate(child) {
                if (ts.isVariableDeclaration(child) && ts.isIdentifier(child.name) && child.name.text === "result" &&
                    child.initializer && ts.isArrayLiteralExpression(child.initializer)) {
                    for (const literal of child.initializer.elements) {
                        if (!ts.isStringLiteral(literal)) continue;
                        const output = ts.createSourceFile("template.ts", literal.text, ts.ScriptTarget.Latest, true);
                        const declaration = output.statements[0];
                        if (output.parseDiagnostics.length || output.statements.length !== 1 ||
                            !ts.isImportDeclaration(declaration) || !ts.isStringLiteral(declaration.moduleSpecifier) ||
                            declaration.moduleSpecifier.text !== "./types.js") continue;
                        const clause = declaration.importClause;
                        if (!clause || clause.isTypeOnly || clause.name || !clause.namedBindings ||
                            !ts.isNamedImports(clause.namedBindings)) throw new Error("unexpected generated import form");
                        // Pinned generator uses an unescaped string literal. Refuse an
                        // encoding change rather than mistake decoded offsets for source offsets.
                        const start = literal.getStart(source) + 1;
                        if (source.text.slice(start, literal.end - 1) !== literal.text) {
                            throw new Error("generated import template contains escaped source characters");
                        }
                        templates++;
                        for (const specifier of clause.namedBindings.elements) {
                            const exported = symbols.get((specifier.propertyName || specifier.name).text);
                            if (!exported) throw new Error(`unresolved generated import: ${specifier.name.text}`);
                            const target = exported.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(exported) : exported;
                            if (!(target.flags & ts.SymbolFlags.Value) && target.flags & ts.SymbolFlags.Type && !specifier.isTypeOnly) {
                                positions.add(start + specifier.getStart(output));
                            }
                        }
                    }
                }
                ts.forEachChild(child, findTemplate);
            }
            ts.forEachChild(node, findTemplate);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (templates !== 1) throw new Error(`want one generated types import template, got ${templates}`);
    return { source, positions };
}

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
        verbatimModuleSyntax: true,
        erasableSyntaxOnly: false,
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
    const generator = generatorEdits(ts, tree, program);
    const generated = path.join(directory, "diagnosticInformationMap.generated.ts");
    let regenerate = generator.positions.size > 0 || !program.getSourceFile(generated);
    const owned = new Set(roots.map(name => path.resolve(name)));
    const edits = new Map();
    if (generator.positions.size) edits.set(generator.source, generator.positions);
    const declined = [];
    let imports = 0;
    let exports = 0;
    for (const diagnostic of program.getSemanticDiagnostics()) {
        if (diagnostic.code !== 1484 && diagnostic.code !== 1205) continue;
        const source = diagnostic.file;
        // The generator template is adapted instead of its generated artifact.
        if (source && path.resolve(source.fileName) === generated) {
            regenerate = true;
            continue;
        }
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
    // Setup ran before this adaptation. Refresh its artifact through the same
    // upstream owner after fixing the template, including an already-stale tree.
    if (regenerate) {
        const result = spawnSync(process.execPath, ["scripts/processDiagnosticMessages.mjs", "src/compiler/diagnosticMessages.json"], {
            cwd: tree,
            stdio: "inherit",
        });
        if (result.error) throw result.error;
        if (result.status !== 0) throw new Error(`diagnostic generation failed: exit ${result.status}, signal ${result.signal}`);
    }
    console.log(JSON.stringify({ files: edits.size, imports, exports, generatorImports: generator.positions.size, regenerated: regenerate, declined }, null, 2));
    if (declined.length) process.exitCode = 1;
}

try { main(); }
catch (error) { console.error(error.message); process.exitCode = 1; }
