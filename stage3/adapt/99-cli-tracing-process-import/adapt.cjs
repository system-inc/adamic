#!/usr/bin/env node
"use strict";
// tracing.ts's read of an ambient process: startTracing names its trace files with `process.pid`, through 65's
// `declare const process: { pid: number };`. Adamic refuses an ambient value declaration, so process arrives through a
// static top-level `import process from "node:process";` and the declaration goes. The default import is the form the
// checker, Lower and Node agree on: it binds to the export= value of @types/node's `node:process`, the global process
// Lower lowers `process.pid` from, and under Node it is that process object itself, where a namespace import is a
// module namespace holding a copy of its properties. 65's `declare const require: (specifier: "fs") => any;` is not
// this adaptation's: tracing's require("fs") moves to a node:fs import in its own change, so the require line is left
// as it is, and this adapter also accepts the file with it already gone.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");

const tracingName = "src/compiler/tracing.ts";
// tracing.ts's second import, upstream's, and its third, which 47 adds; the node:process import follows them.
const namespaceText = "import * as performance from \"./_namespaces/ts.performance.js\";";
const anchorText = "import { errorMessage } from \"./hostErrors.js\";";
const importText = "import process from \"node:process\";";
const local = "process";
// 65's declarations, appended with LF line breaks to the CRLF file: require is the node:fs change's, process is this one's.
const requireText = "declare const require: (specifier: \"fs\") => any;";
const declarationText = "declare const process: { pid: number };";
// startTracing's countPart, upstream's text in both states: the two process.pid reads this adaptation rebinds.
const countPartText = "const countPart = mode === \"build\" ? `.${process.pid}-${++traceCount}`\r\n"
    + "            : mode === \"server\" ? `.${process.pid}`\r\n"
    + "            : ``;";

function one(found, label) {
    assert.equal(found.length, 1, `${label}: want one, got ${found.length}`);
    return found[0];
}

function transpile(ts, text, removeComments) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
}

// tracing.ts is CRLF throughout, apart from 65's appended LF lines: the blank line before them, the require line while
// it stands, and this adaptation's declaration until it goes.
function crlf(text, label) {
    let own = text;
    for (const line of [declarationText, requireText]) {
        if (own.endsWith(`\n${line}\n`)) own = own.slice(0, -(line.length + 1));
    }
    if (own.endsWith("\r\n\n")) own = own.slice(0, -1);
    assert.equal(own.split("\r\n").length, own.split("\n").length, `${label} has a line break that is not CRLF`);
}

function identifiers(ts, root, name) {
    return nodes(ts, root, node => ts.isIdentifier(node) && node.text === name);
}

// The top-level `declare const` statements naming one identifier, by that name.
function declared(ts, file, name) {
    return file.statements.filter(node => ts.isVariableStatement(node)
        && node.declarationList.declarations.some(declaration => ts.isIdentifier(declaration.name) && declaration.name.text === name));
}

// The default import right after the ./hostErrors.js import, and 65's process declaration at the end of the file,
// present or absent together. startTracing's countPart is upstream's text in both states, and holds the file's only
// process reads.
function fileState(ts, file) {
    const imports = file.statements.filter(ts.isImportDeclaration);
    assert(imports.length >= 3 && imports.slice(0, 3).every((node, index) => file.statements[index] === node), "tracing.ts's imports moved");
    assert.equal(imports[0].moduleSpecifier.text, "./_namespaces/ts.js", "unreviewed first import of tracing.ts");
    assert.equal(imports[1].getText(file), namespaceText, "unreviewed second import of tracing.ts");
    const anchor = imports[2];
    assert.equal(anchor.getText(file), anchorText, "unreviewed third import of tracing.ts");
    let importState = "original";
    if (imports.length === 4) {
        assert.equal(imports[3].getText(file), importText, "unreviewed node:process import");
        assert.equal(file.statements[3], imports[3], "the node:process import does not follow the ./hostErrors.js import");
        assert.equal(file.text.slice(anchor.end, imports[3].getStart(file)), "\r\n", "unreviewed text between the imports");
        importState = "adapted";
    }
    else assert.equal(imports.length, 3, "tracing.ts imports moved");
    let declaration = { state: "adapted" };
    const declarations = declared(ts, file, local);
    if (declarations.length) {
        const statement = one(declarations, "process declaration");
        assert.equal(statement.getText(file), declarationText, "unreviewed process declaration");
        assert.equal(file.statements.at(-1), statement, "the process declaration is not the last statement");
        assert.equal(file.text.slice(statement.end), "\n", "unreviewed text after the process declaration");
        // After 65's require line, or, once the node:fs change has taken that line, after the file's own last statement.
        const previous = file.statements.at(-2);
        const requires = declared(ts, file, "require");
        const leading = file.text.slice(statement.getFullStart(), statement.getStart(file));
        if (previous.getText(file) === requireText) {
            assert.equal(leading, "\n", "unreviewed text before the process declaration");
            declaration = { state: "original", statement, replacement: "\n" };
        }
        else {
            assert.equal(requires.length, 0, "65's require declaration does not directly precede the process declaration");
            assert.equal(leading, "\r\n\n", "unreviewed text before the process declaration");
            declaration = { state: "original", statement, replacement: "\r\n" };
        }
    }
    const owner = one(nodes(ts, file, node => ts.isFunctionDeclaration(node) && node.name?.text === "startTracing" && node.body), "startTracing");
    const countPart = one(nodes(ts, owner.body, node => ts.isVariableStatement(node)
        && node.declarationList.declarations.some(item => ts.isIdentifier(item.name) && item.name.text === "countPart")), "startTracing's countPart");
    assert.equal(countPart.getText(file), countPartText, "unreviewed countPart");
    const pieces = { import: importState, "process declaration": declaration.state };
    const states = Object.values(pieces);
    assert(states.every(state => state === states[0]), `half-adapted tracing.ts: ${Object.entries(pieces).map(([name, state]) => `${name} ${state}`).join(", ")}`);
    // process is named three times in either state: the declaration or the import binding, and the two pid reads,
    // both in countPart, each the receiver of `.pid`.
    const named = identifiers(ts, file, local);
    assert.equal(named.length, 3, "tracing.ts names process outside the reviewed sites");
    const binding = states[0] === "original" ? declaration.statement.declarationList.declarations[0].name : imports[3].importClause.name;
    const reads = named.filter(node => node !== binding);
    assert.equal(reads.length, 2, "tracing.ts's process binding moved");
    for (const read of reads) {
        assert(read.pos >= countPart.pos && read.end <= countPart.end, "process is read outside countPart");
        assert(ts.isPropertyAccessExpression(read.parent) && read.parent.expression === read && read.parent.name.text === "pid", "process is read other than as process.pid");
    }
    return { state: states[0], anchor, declaration, countPart, binding, reads };
}

// Before the edit, by symbol: the process declaration is read only at countPart's two pid reads, and each resolves to
// the file's own declaration, never to a global.
function onlyAtSite(checker, found) {
    const symbol = checker.getSymbolAtLocation(found.binding);
    assert(symbol, "the process declaration has no symbol");
    for (const read of found.reads) assert.equal(checker.getSymbolAtLocation(read), symbol, "a process read is not the file's own declaration");
}

// The planned process reads resolve to the import. When the program has Node's types, the import resolves to
// @types/node's `node:process`, and binds to its export= value, the global process, whose pid is a number. When it has
// none, as 47's program over an adapted tree has none (stage3 never runs npm), the specifier draws exactly the
// diagnostic sys.ts's `node:fs` import draws in the same program, with the name changed. Returns that one diagnostic,
// or none.
function plannedBinding(ts, program, fileName, sysName) {
    const file = program.getSourceFile(fileName);
    const checker = program.getTypeChecker();
    const found = identifiers(ts, file, local);
    const declaration = one(found.filter(node => ts.isImportClause(node.parent)), `${local} import binding`);
    const alias = checker.getSymbolAtLocation(declaration);
    for (const read of found.filter(node => node !== declaration)) assert.equal(checker.getSymbolAtLocation(read), alias, `a ${local} read does not resolve to the import`);
    const specifier = declaration.parent.parent.moduleSpecifier;
    const module = checker.getSymbolAtLocation(specifier);
    if (module) {
        // @types/node declares `node:process` as a module of its own whose export= is the global process.
        assert(module.declarations?.some(node => ts.isModuleDeclaration(node) && ts.isStringLiteral(node.name) && node.name.text === "node:process"
            && /[\\/]@types[\\/]node[\\/]/.test(node.getSourceFile().fileName)), "node:process does not resolve to @types/node's module");
        const target = checker.getAliasedSymbol(alias);
        assert.equal(target, checker.resolveName(local, undefined, ts.SymbolFlags.Value, false), "the node:process default import is not the global process");
        assert(target.declarations?.every(node => /[\\/]@types[\\/]node[\\/]/.test(node.getSourceFile().fileName)), "the global process is not @types/node's");
        const pid = checker.getTypeOfSymbol(target).getProperty("pid");
        assert(pid && checker.getTypeOfSymbol(pid).flags & ts.TypeFlags.Number, "the global process has no number pid");
        return [];
    }
    const sys = program.getSourceFile(sysName);
    const fsImport = one(sys.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === "node:fs"), "sys.ts's node:fs import");
    assert.equal(checker.getSymbolAtLocation(fsImport.moduleSpecifier), undefined, "node:fs resolves where node:process does not");
    const at = node => diagnostic => diagnostic.start === node.getStart(node.getSourceFile()) && diagnostic.length === node.getWidth(node.getSourceFile());
    const fsDiagnostic = one(program.getSemanticDiagnostics(sys).filter(at(fsImport.moduleSpecifier)), "the node:fs import's diagnostic");
    const processDiagnostic = one(program.getSemanticDiagnostics(file).filter(at(specifier)), "the node:process import's diagnostic");
    const message = diagnostic => `TS${diagnostic.code} ${ts.flattenDiagnosticMessageText(diagnostic.messageText, "\n")}`;
    assert.equal(message(processDiagnostic), message(fsDiagnostic).replace("'node:fs'", "'node:process'"), "the node:process import's diagnostic is not node:fs's");
    return [message(processDiagnostic)];
}

// A program over the tree with tracing.ts's text replaced, under the given program's options.
function plannedProgram(ts, program, fileName, text) {
    const host = ts.createCompilerHost(program.getCompilerOptions(), true);
    const original = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, fresh) => path.resolve(name) === path.resolve(fileName)
        ? ts.createSourceFile(name, text, language, true)
        : original(name, language, onError, fresh);
    return ts.createProgram(program.getRootFileNames(), program.getCompilerOptions(), host);
}

// The file's diagnostics, by code and message, with the line numbers left out: the import moves every line below it.
function diagnosticsOf(ts, program, fileName) {
    const file = program.getSourceFile(fileName);
    return [...program.getSyntacticDiagnostics(file), ...program.getSemanticDiagnostics(file)]
        .map(diagnostic => `TS${diagnostic.code} ${ts.flattenDiagnosticMessageText(diagnostic.messageText, "\n")}`).sort();
}

function apply(text, edits) {
    let result = text;
    let boundary = text.length;
    for (const edit of [...edits].sort((a, b) => b.start - a.start)) {
        assert(edit.end <= boundary, "overlapping edits");
        result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
        boundary = edit.start;
    }
    return result;
}

// Stock 6.0.3's JavaScript of the planned file is upstream's with the import added, and nothing else: 65's declaration
// is types only. The expected text is built from the before output, never from source.
function provesJavaScript(ts, before, after) {
    for (const removeComments of [false, true]) {
        const expected = transpile(ts, before, removeComments);
        assert.equal(expected.split(`${anchorText}\n`).length - 1, 1, "upstream JavaScript lacks the ./hostErrors.js import");
        assert.equal(transpile(ts, after, removeComments), expected.replace(`${anchorText}\n`, `${anchorText}\n${importText}\n`), "the adapted JavaScript differs by more than the import");
    }
}

// The declarations are identical, with stripInternal as typescript.d.ts is built and without it: tracingEnabled is
// @internal, and neither the process declaration nor the import reaches a declaration.
function provesDeclarations(ts, fileName, before, after) {
    for (const stripInternal of [true, false]) {
        const declare = text => ts.transpileDeclaration(text, { compilerOptions: { stripInternal }, fileName: path.basename(fileName) }).outputText;
        assert.equal(declare(after), declare(before), `the declarations changed (stripInternal ${stripInternal})`);
    }
}

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const fileName = path.join(tree, tracingName);
    const text = fs.readFileSync(fileName, "utf8");
    crlf(text, "tracing.ts");
    const parsed = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(parsed.parseDiagnostics.length, 0, "cannot parse tracing.ts");
    if (fileState(ts, parsed).state === "adapted") return { fileName, before: text, after: text, evidence: [] };
    // Only an original file needs the checker, to prove the removed declaration is read nowhere else.
    const program = compiler(tree, ts);
    const checker = program.getTypeChecker();
    const file = program.getSourceFile(fileName);
    assert(file && file.text === text, "the program's tracing.ts is not the file on disk");
    const found = fileState(ts, file);
    onlyAtSite(checker, found);
    const where = node => file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1;
    const statement = found.declaration.statement;
    const edits = [
        { start: found.anchor.end, end: found.anchor.end, text: `\r\n${importText}` },
        // From the end of the statement before the declaration to the end of the file: that statement's own ending again.
        { start: statement.getFullStart(), end: text.length, text: found.declaration.replacement },
    ];
    const evidence = [
        { site: "node:process import", file: tracingName, line: where(found.anchor) + 1, before: "", after: importText },
        { site: "process declaration", file: tracingName, line: where(statement), before: declarationText, after: "" },
        ...found.reads.map(read => ({ site: "process.pid read, rebound", file: tracingName, line: where(read), text: read.parent.getText(file) })),
    ];
    const after = apply(text, edits);
    // Re-read the planned text: it must parse, keep the file's line breaks, and be exactly this adaptation's state.
    const replanned = ts.createSourceFile(fileName, after, ts.ScriptTarget.Latest, true);
    assert.equal(replanned.parseDiagnostics.length, 0, "adapted syntax error");
    crlf(after, "the adapted tracing.ts");
    assert.equal(declared(ts, replanned, "require").length, declared(ts, file, "require").length, "the require declaration moved");
    assert.equal(fileState(ts, replanned).state, "adapted", "the planned text is not this adaptation's");
    provesJavaScript(ts, text, after);
    provesDeclarations(ts, fileName, text, after);
    // The planned tree type-checks as before under 47's program: the same diagnostics plus at most the unresolved
    // specifier's, which sys.ts's node: imports share, and the process reads bound to the import.
    const planned = plannedProgram(ts, program, fileName, after);
    const unresolved = plannedBinding(ts, planned, fileName, path.join(tree, "src/compiler/sys.ts"));
    assert.deepEqual(diagnosticsOf(ts, planned, fileName), [...diagnosticsOf(ts, program, fileName), ...unresolved].sort(), "tracing.ts's diagnostics changed");
    evidence.push({ site: "47's program", file: tracingName, before: diagnosticsOf(ts, program, fileName), after: diagnosticsOf(ts, planned, fileName) });
    return { fileName, before: text, after, evidence };
}

function adapt(tree, ts) {
    const change = plan(tree, ts);
    assert.equal(fs.readFileSync(change.fileName, "utf8"), change.before, "source changed while planning");
    if (change.after !== change.before) fs.writeFileSync(change.fileName, change.after);
    return { files: change.after === change.before ? 0 : 1, javascriptChanging: true, evidence: change.evidence };
}

module.exports = { adapt, plan, plannedBinding, plannedProgram, diagnosticsOf, importText, anchorText, declarationText, requireText };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
