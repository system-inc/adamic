#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");

const helper = `/** @internal */
export function errorCode(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "code" in error && typeof error.code === "string" ? error.code : undefined;
}

/** @internal */
export function errorMessage(error: unknown): string | undefined {
    return typeof error === "object" && error !== null && "message" in error && typeof error.message === "string" ? error.message : undefined;
}
`;
function compiler(tree, ts) {
    if (ts.version !== "6.0.3") throw new Error(`want stock typescript@6.0.3, got ${ts.version}`);
    const configFile = path.join(tree, "src/compiler/tsconfig.json");
    const read = ts.readConfigFile(configFile, ts.sys.readFile);
    if (read.error) throw new Error(ts.flattenDiagnosticMessageText(read.error.messageText, "\n"));
    const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configFile), {
        strict: true, strictBindCallApply: true, useUnknownInCatchVariables: true,
        noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
        verbatimModuleSyntax: true, erasableSyntaxOnly: false,
        module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler,
        moduleDetection: ts.ModuleDetectionKind.Force, target: ts.ScriptTarget.ES2024,
        noEmit: true, emitDeclarationOnly: false,
    }, configFile);
    if (config.errors.length) throw new Error(JSON.stringify(config.errors));
    return ts.createProgram(config.fileNames, config.options);
}
function nodes(ts, root, predicate) {
    const found = [];
    function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); }
    visit(root);
    return found;
}
function adapt(tree, ts) {
    const program = compiler(tree, ts);
    const checker = program.getTypeChecker();
    const directory = path.join(tree, "src/compiler");
    const planned = new Map();
    const evidence = [];
    function source(name) {
        const file = program.getSourceFile(path.join(directory, name));
        if (!file || file.parseDiagnostics.length) throw new Error(`cannot parse ${name}`);
        return file;
    }
    function one(root, predicate, label) {
        const found = nodes(ts, root, predicate);
        if (found.length !== 1) throw new Error(`${label}: want one, got ${found.length}`);
        return found[0];
    }
    function edit(file, node, text) {
        if (node.getText(file) === text) return;
        if (!planned.has(file)) planned.set(file, []);
        planned.get(file).push({ start: node.getStart(file), end: node.end, text });
        const where = file.getLineAndCharacterOfPosition(node.getStart(file));
        evidence.push({ file: path.basename(file.fileName), line: where.line + 1, before: node.getText(file), after: text });
    }
    function catchRead(file, root, field, helperName) {
        const clause = one(root, ts.isCatchClause, "reviewed catch");
        const binding = clause.variableDeclaration;
        if (!binding || !ts.isIdentifier(binding.name)) throw new Error("unreviewed catch binding");
        const symbol = checker.getSymbolAtLocation(binding.name);
        if (!(checker.getTypeAtLocation(binding.name).flags & ts.TypeFlags.Unknown)) throw new Error("catch is not unknown");
        const reads = nodes(ts, clause.block, node => ts.isPropertyAccessExpression(node) && node.name.text === field && checker.getSymbolAtLocation(node.expression) === symbol);
        const calls = nodes(ts, clause.block, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === helperName && node.arguments.length === 1 && checker.getSymbolAtLocation(node.arguments[0]) === symbol);
        if (reads.length + calls.length !== 1) throw new Error(`unreviewed ${field} read`);
        if (reads.length) edit(file, reads[0], `${helperName}(${binding.name.text})`);
        const imports = file.statements.filter(ts.isImportDeclaration);
        const existing = imports.filter(node => node.moduleSpecifier.text === "./hostErrors.js");
        if (existing.length) {
            if (existing.length !== 1 || existing[0].importClause?.isTypeOnly || !existing[0].importClause?.namedBindings || !ts.isNamedImports(existing[0].importClause.namedBindings) || existing[0].importClause.namedBindings.elements.map(n => n.getText(file)).join() !== helperName) throw new Error("unreviewed helper import");
        } else {
            const at = imports.length ? imports[imports.length - 1].end : 0;
            const newline = file.text.includes("\r\n") ? "\r\n" : "\n";
            if (!planned.has(file)) planned.set(file, []);
            planned.get(file).push({ start: at, end: at, text: `${newline}import { ${helperName} } from "./hostErrors.js";` });
        }
    }
    const sys = source("sys.ts");
    const mkdir = one(sys, node => ts.isMethodDeclaration(node) && node.name.getText(sys) === "createDirectory", "Node mkdir method");
    const mkdirTry = one(mkdir, ts.isTryStatement, "Node mkdir try");
    const mkdirCall = one(mkdirTry.tryBlock, ts.isCallExpression, "Node mkdir call");
    if (mkdirCall.expression.getText(sys) !== "_fs.mkdirSync" || mkdirCall.arguments.length !== 1 || mkdirCall.arguments[0].getText(sys) !== "directoryName") throw new Error("unreviewed mkdir producer");
    catchRead(sys, mkdir, "code", "errorCode");
    const tracing = source("tracing.ts");
    const start = one(tracing, node => ts.isFunctionDeclaration(node) && node.name?.text === "startTracing", "startTracing");
    const loadTry = one(start, ts.isTryStatement, "builtin load try");
    const load = one(loadTry.tryBlock, ts.isCallExpression, "builtin require");
    if (load.expression.getText(tracing) !== "require" || load.arguments.length !== 1 || !ts.isStringLiteral(load.arguments[0]) || load.arguments[0].text !== "fs") throw new Error("unreviewed loader producer");
    catchRead(tracing, loadTry, "message", "errorMessage");
    const readFile = one(sys, node => ts.isFunctionDeclaration(node) && node.name?.text === "readFile", "Node readFile");
    const bufferDeclaration = one(readFile, node => ts.isVariableDeclaration(node) && node.name.getText(sys) === "buffer", "buffer declaration");
    const readTry = one(readFile, ts.isTryStatement, "buffer producer try");
    const readCall = one(readTry.tryBlock, ts.isCallExpression, "buffer producer call");
    const length = one(readFile, node => ts.isVariableDeclaration(node) && node.name.getText(sys) === "len", "buffer length");
    if (readCall.expression.getText(sys) !== "_fs.readFileSync" || readCall.arguments.length !== 1 || readCall.arguments[0].getText(sys) !== "fileName" || length.initializer.getText(sys) !== "buffer.length") throw new Error("unreviewed buffer producer");
    const bufferSymbol = checker.getSymbolAtLocation(bufferDeclaration.name);
    const accesses = nodes(ts, readFile, node => ts.isElementAccessExpression(node) && checker.getSymbolAtLocation(node.expression) === bufferSymbol);
    if (accesses.length !== 11) throw new Error(`unreviewed byte-read count: ${accesses.length}`);
    const expected = ["buffer[0]", "buffer[1]", "buffer[i]", "buffer[i]", "buffer[i + 1]", "buffer[i + 1]", "buffer[0]", "buffer[1]", "buffer[0]", "buffer[1]", "buffer[2]"];
    for (const [index, access] of accesses.entries()) {
        if (access.getText(sys) !== expected[index]) throw new Error("unreviewed byte read");
        const parent = access.parent;
        if (ts.isBinaryExpression(parent) && parent.left === access && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken) continue;
        if (!ts.isNonNullExpression(parent)) edit(sys, access, `${access.getText(sys)}!`);
    }
    // Pin the actual guards, not offsets. Earlier adaptations may add assertions.
    const conditions = nodes(ts, readFile, ts.isIfStatement).map(node => ts.createPrinter().printNode(ts.EmitHint.Unspecified, node.expression, sys));
    const normalized = conditions.map(text => text.split("!").join(""));
    const guards = ["len >= 2 && buffer[0] === 0xFE && buffer[1] === 0xFF", "len >= 2 && buffer[0] === 0xFF && buffer[1] === 0xFE", "len >= 3 && buffer[0] === 0xEF && buffer[1] === 0xBB && buffer[2] === 0xBF"];
    if (JSON.stringify(normalized) !== JSON.stringify(guards)) throw new Error("unreviewed BOM guards");
    const loop = one(readFile, ts.isForStatement, "byte-pair loop");
    if (loop.initializer.getText(sys) !== "let i = 0" || loop.condition.getText(sys) !== "i < len" || loop.incrementor.getText(sys) !== "i += 2" || !nodes(ts, readFile, node => ts.isBinaryExpression(node) && node.getText(sys) === "len &= ~1").length) throw new Error("unreviewed pair bounds");
    const helperFile = path.join(directory, "hostErrors.ts");
    if (fs.existsSync(helperFile) && fs.readFileSync(helperFile, "utf8") !== helper) throw new Error("hostErrors.ts already exists with different contents");
    const results = [];
    for (const [file, edits] of planned) {
        let text = file.text;
        let boundary = text.length;
        for (const change of edits.sort((a, b) => b.start - a.start)) {
            if (change.end > boundary) throw new Error("overlapping edits");
            text = text.slice(0, change.start) + change.text + text.slice(change.end);
            boundary = change.start;
        }
        if (ts.createSourceFile(file.fileName, text, ts.ScriptTarget.Latest, true).parseDiagnostics.length) throw new Error("adapted syntax error");
        if (fs.readFileSync(file.fileName, "utf8") !== file.text) throw new Error("source changed while planning");
        results.push([file.fileName, text]);
    }
    for (const [file, text] of results) fs.writeFileSync(file, text);
    const added = !fs.existsSync(helperFile);
    if (added) fs.writeFileSync(helperFile, helper);
    return { files: results.length + Number(added), propertyReads: evidence.filter(e => e.after.startsWith("error")).length, byteReads: evidence.filter(e => e.after.endsWith("!")).length, helperAdded: added, evidence };
}
module.exports = { adapt, compiler, nodes };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
