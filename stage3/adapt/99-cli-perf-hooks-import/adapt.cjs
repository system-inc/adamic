#!/usr/bin/env node
"use strict";
// performanceCore.ts's read of an ambient require: tryGetPerformance calls `require("perf_hooks")` under
// `if (isNodeLikeSystem())`, through 65's `declare const require: (specifier: "perf_hooks") => any;`. Adamic refuses
// an ambient value declaration, and ruled (Nov 6, #00qn5az) that perf_hooks arrives through a static top-level
// `node:perf_hooks` namespace import, in 93's idiom, and the declaration goes. isNodeLikeSystem and its require test
// are 99q's, and Correctness changes them in her own adaptation; this one leaves core.ts alone.
// 65's cast gave Node's perf_hooks object the file's own Performance type. Without it, Node's Performance reaches
// tryGetPerformance's inferred return, and Lower refuses its views as the file's types. So tryGetPerformance declares
// its return with the file's own Performance, and PerformanceTime's timeOrigin is readonly, which makes that view
// true of Node's object (tsc never writes timeOrigin).
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");

const performanceName = "src/compiler/performanceCore.ts";
const anchorText = "import { isNodeLikeSystem } from \"./_namespaces/ts.js\";";
const importText = "import * as perfHooks from \"node:perf_hooks\";";
const local = "perfHooks";
// The perf_hooks binding in tryGetPerformance as 65 leaves it and as this adaptation writes it.
const binding = {
    original: "const { performance } = require(\"perf_hooks\") as { performance?: Performance | undefined };",
    adapted: "const { performance } = perfHooks;",
};
// tryGetPerformance's header, from `function` to its body, without a return type and with the file's own view.
const header = {
    original: "function tryGetPerformance() ",
    adapted: "function tryGetPerformance(): { shouldWriteNativeEvents: boolean; performance: Performance; } | undefined ",
};
// PerformanceTime's timeOrigin member.
const timeOrigin = {
    original: "timeOrigin: number;",
    adapted: "readonly timeOrigin: number;",
};
// 65 appends its declaration with LF line breaks to the CRLF file; removing it gives back the file's own ending.
const requireText = "declare const require: (specifier: \"perf_hooks\") => any;";
const requireTrivia = "\r\n\n";

function one(found, label) {
    assert.equal(found.length, 1, `${label}: want one, got ${found.length}`);
    return found[0];
}

function transpile(ts, text, removeComments) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
}

// performanceCore.ts is CRLF throughout, apart from 65's appended LF lines, which this adaptation removes.
function crlf(text, label) {
    const own = text.endsWith(`${requireTrivia}${requireText}\n`) ? text.slice(0, -(requireText.length + 2)) : text;
    assert.equal(own.split("\r\n").length, own.split("\n").length, `${label} has a line break that is not CRLF`);
}

function textState(text, states, label) {
    const found = Object.keys(states).filter(state => states[state] === text);
    assert.equal(found.length, 1, `uninventoried ${label}: ${text}`);
    return found[0];
}

function identifiers(ts, file, name) {
    return nodes(ts, file, node => ts.isIdentifier(node) && node.text === name);
}

// The namespace import right after the one ./_namespaces/ts.js import, the binding in tryGetPerformance, 65's
// require declaration at the end of the file, tryGetPerformance's return type and PerformanceTime's readonly
// timeOrigin, each present or absent together.
function fileState(ts, file) {
    const imports = file.statements.filter(ts.isImportDeclaration);
    const anchor = imports[0];
    assert(anchor && anchor.getText(file) === anchorText && file.statements[0] === anchor, "unreviewed first import of performanceCore.ts");
    let importState = "original";
    if (imports.length === 2) {
        assert.equal(imports[1].getText(file), importText, "unreviewed perf_hooks import");
        assert.equal(file.statements[1], imports[1], "the perf_hooks import does not follow the ./_namespaces/ts.js import");
        assert.equal(file.text.slice(anchor.end, imports[1].getStart(file)), "\r\n", "unreviewed text between the imports");
        importState = "adapted";
    }
    else assert.equal(imports.length, 1, "performanceCore.ts imports moved");
    const owner = one(file.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === "tryGetPerformance" && node.body), "tryGetPerformance");
    const statement = one(nodes(ts, owner, node => ts.isVariableStatement(node)
        && node.declarationList.declarations.some(declaration => ts.isObjectBindingPattern(declaration.name))), "tryGetPerformance's destructuring");
    const bindingState = textState(statement.getText(file), binding, "perf_hooks binding");
    // The binding sits in the try block under `if (isNodeLikeSystem())`, as upstream has it.
    const block = statement.parent;
    assert(ts.isBlock(block) && ts.isTryStatement(block.parent) && block.parent.tryBlock === block, "the perf_hooks binding left its try block");
    const guard = block.parent.parent.parent;
    assert(ts.isIfStatement(guard) && guard.expression.getText(file) === "isNodeLikeSystem()", "the perf_hooks binding left its isNodeLikeSystem guard");
    const requires = file.statements.filter(node => ts.isVariableStatement(node)
        && node.declarationList.declarations.some(declaration => ts.isIdentifier(declaration.name) && declaration.name.text === "require"));
    let requireDeclaration = { state: "adapted" };
    if (requires.length) {
        const declaration = one(requires, "require declaration");
        assert.equal(declaration.getText(file), requireText, "unreviewed require declaration");
        assert.equal(file.statements.at(-1), declaration, "the require declaration is not the last statement");
        assert.equal(file.text.slice(declaration.getFullStart(), declaration.getStart(file)), requireTrivia, "unreviewed text before the require declaration");
        assert.equal(file.text.slice(declaration.end), "\n", "unreviewed text after the require declaration");
        requireDeclaration = { state: "original", declaration };
    }
    const headerState = textState(file.text.slice(owner.getStart(file), owner.body.getStart(file)), header, "tryGetPerformance header");
    const performanceTime = one(file.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === "PerformanceTime"), "PerformanceTime");
    const member = one(performanceTime.members.filter(node => ts.isPropertySignature(node) && node.name.getText(file) === "timeOrigin"), "PerformanceTime's timeOrigin");
    const timeOriginState = textState(member.getText(file), timeOrigin, "timeOrigin member");
    const pieces = { import: importState, binding: bindingState, "require declaration": requireDeclaration.state, "return type": headerState, timeOrigin: timeOriginState };
    const states = Object.values(pieces);
    assert(states.every(state => state === states[0]), `half-adapted performanceCore.ts: ${Object.entries(pieces).map(([name, state]) => `${name} ${state}`).join(", ")}`);
    // Before, require is named twice (the call and the declaration) and the local nowhere; after, require nowhere
    // and the local twice (the import binding and the destructured read).
    const original = states[0] === "original";
    assert.equal(identifiers(ts, file, "require").length, original ? 2 : 0, "performanceCore.ts names require outside the reviewed sites");
    assert.equal(identifiers(ts, file, local).length, original ? 0 : 2, `performanceCore.ts names ${local} outside the reviewed sites`);
    // timeOrigin is named at PerformanceTime's member and tryGetPerformanceHooks's typeof read, both sides, so no other
    // member redeclares it writable.
    assert.equal(identifiers(ts, file, "timeOrigin").length, 2, "performanceCore.ts names timeOrigin outside the reviewed sites");
    return { state: states[0], anchor, statement, requireDeclaration, owner, member };
}

// Before the edit, by symbol: the require declaration is read only at the binding being replaced, and that read
// resolves to the file's own declaration, never to a global.
function onlyAtSite(ts, checker, file, found) {
    const name = found.requireDeclaration.declaration.declarationList.declarations[0].name;
    const symbol = checker.getSymbolAtLocation(name);
    assert(symbol, "the require declaration has no symbol");
    const uses = identifiers(ts, file, "require").filter(node => node !== name);
    assert.equal(uses.length, 1, "require: want one read");
    assert.equal(checker.getSymbolAtLocation(uses[0]), symbol, "the require read is not the file's own declaration");
    assert(uses[0].pos >= found.statement.pos && uses[0].end <= found.statement.end, "require is read outside the replaced binding");
}

// The planned perfHooks read resolves to the import. When the program has Node's types, the import resolves to
// @types/node's `node:perf_hooks`, which exports performance; when it has none, as 47's program over an adapted
// tree has none (stage3 never runs npm), the specifier draws exactly the diagnostic sys.ts's `node:fs` import
// draws in the same program, with the name changed. Returns that one diagnostic, or none.
function plannedBinding(ts, program, fileName, sysName) {
    const file = program.getSourceFile(fileName);
    const checker = program.getTypeChecker();
    const found = identifiers(ts, file, local);
    const declaration = one(found.filter(node => ts.isNamespaceImport(node.parent)), `${local} import binding`);
    const read = one(found.filter(node => node !== declaration), `${local} read`);
    assert.equal(checker.getSymbolAtLocation(read), checker.getSymbolAtLocation(declaration), `${local} read does not resolve to the import`);
    const specifier = declaration.parent.parent.parent.moduleSpecifier;
    const module = checker.getSymbolAtLocation(specifier);
    if (module) {
        // @types/node declares `node:perf_hooks` as a module of its own that re-exports `perf_hooks`.
        assert(module.declarations?.some(node => ts.isModuleDeclaration(node) && ts.isStringLiteral(node.name) && node.name.text === "node:perf_hooks"
            && /[\\/]@types[\\/]node[\\/]/.test(node.getSourceFile().fileName)), "node:perf_hooks does not resolve to @types/node's module");
        assert(checker.getExportsOfModule(module).some(symbol => symbol.name === "performance"), "node:perf_hooks exports no performance");
        return [];
    }
    const sys = program.getSourceFile(sysName);
    const fsImport = one(sys.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === "node:fs"), "sys.ts's node:fs import");
    assert.equal(checker.getSymbolAtLocation(fsImport.moduleSpecifier), undefined, "node:fs resolves where node:perf_hooks does not");
    const at = node => diagnostic => diagnostic.start === node.getStart(node.getSourceFile()) && diagnostic.length === node.getWidth(node.getSourceFile());
    const fsDiagnostic = one(program.getSemanticDiagnostics(sys).filter(at(fsImport.moduleSpecifier)), "the node:fs import's diagnostic");
    const perfDiagnostic = one(program.getSemanticDiagnostics(file).filter(at(specifier)), "the node:perf_hooks import's diagnostic");
    const message = diagnostic => `TS${diagnostic.code} ${ts.flattenDiagnosticMessageText(diagnostic.messageText, "\n")}`;
    assert.equal(message(perfDiagnostic), message(fsDiagnostic).replace("'node:fs'", "'node:perf_hooks'"), "the node:perf_hooks import's diagnostic is not node:fs's");
    return [message(perfDiagnostic)];
}

// The planned return type's Performance is the file's own interface, never a global one, and the timeOrigin it
// inherits from PerformanceTime is readonly.
function localView(ts, program, fileName) {
    const file = program.getSourceFile(fileName);
    const checker = program.getTypeChecker();
    const owner = one(file.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === "tryGetPerformance"), "planned tryGetPerformance");
    const reference = one(nodes(ts, owner.type, node => ts.isTypeReferenceNode(node) && node.typeName.getText(file) === "Performance"), "the return type's Performance");
    const declarations = checker.getSymbolAtLocation(reference.typeName)?.declarations ?? [];
    assert(declarations.length === 1 && ts.isInterfaceDeclaration(declarations[0]) && declarations[0].getSourceFile() === file, "the return type's Performance is not the file's own interface");
    const member = checker.getTypeAtLocation(reference).getProperty("timeOrigin");
    assert(member?.declarations?.length === 1 && ts.getCombinedModifierFlags(member.declarations[0]) & ts.ModifierFlags.Readonly, "the return type's timeOrigin is not readonly");
}

function plannedProgram(ts, program, fileName, text) {
    const host = ts.createCompilerHost(program.getCompilerOptions(), true);
    const original = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, fresh) => path.resolve(name) === path.resolve(fileName)
        ? ts.createSourceFile(name, text, language, true)
        : original(name, language, onError, fresh);
    return ts.createProgram(program.getRootFileNames(), program.getCompilerOptions(), host);
}

// The file's diagnostics under 47's program, by code and message, with the line numbers left out: the import moves
// every line below it.
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

function replaceOnce(text, from, to, label) {
    assert.equal(text.split(from).length - 1, 1, `upstream JavaScript lacks ${label}`);
    return text.replace(from, to);
}

// Stock 6.0.3's JavaScript of the planned file is upstream's with the import added and the binding replaced. The
// expected text is built from the before output, never from source. The return type and readonly are types only:
// the planned text and the planned text without them give the same JavaScript.
function provesJavaScript(ts, before, untyped, after) {
    for (const removeComments of [false, true]) {
        let expected = transpile(ts, before, removeComments);
        expected = replaceOnce(expected, `${anchorText}\n`, `${anchorText}\n${importText}\n`, "the ./_namespaces/ts.js import");
        expected = replaceOnce(expected, "const { performance } = require(\"perf_hooks\");\n", `${binding.adapted}\n`, "the perf_hooks require");
        assert.equal(transpile(ts, untyped, removeComments), transpile(ts, after, removeComments), "the return type or readonly changes JavaScript");
        assert.equal(transpile(ts, after, removeComments), expected, "the adapted JavaScript differs by more than the import and the binding");
    }
}

// With stripInternal, as typescript.d.ts is built, the declarations are identical: tryGetPerformance is not exported,
// PerformanceTime is @internal, and neither the require declaration nor the import reaches a declaration. Without
// stripInternal, the one difference is readonly on PerformanceTime's timeOrigin.
function provesDeclarations(ts, fileName, before, after) {
    for (const stripInternal of [true, false]) {
        const declare = text => ts.transpileDeclaration(text, { compilerOptions: { stripInternal }, fileName: path.basename(fileName) }).outputText;
        let expected = declare(before);
        if (!stripInternal) {
            assert.equal(expected.split(`    ${timeOrigin.original}`).length - 1, 1, "the declarations lack PerformanceTime's timeOrigin");
            expected = expected.replace(`    ${timeOrigin.original}`, `    ${timeOrigin.adapted}`);
        }
        assert.equal(declare(after), expected, `the declarations changed (stripInternal ${stripInternal})`);
    }
}

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const fileName = path.join(tree, performanceName);
    const text = fs.readFileSync(fileName, "utf8");
    crlf(text, "performanceCore.ts");
    const parsed = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(parsed.parseDiagnostics.length, 0, "cannot parse performanceCore.ts");
    if (fileState(ts, parsed).state === "adapted") return { fileName, before: text, after: text, evidence: [] };
    // Only an original file needs the checker, to prove the removed declaration is read nowhere else.
    const program = compiler(tree, ts);
    const checker = program.getTypeChecker();
    const file = program.getSourceFile(fileName);
    assert(file && file.text === text, "the program's performanceCore.ts is not the file on disk");
    const found = fileState(ts, file);
    onlyAtSite(ts, checker, file, found);
    const where = node => file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1;
    const declaration = found.requireDeclaration.declaration;
    const edits = [
        { start: found.anchor.end, end: found.anchor.end, text: `\r\n${importText}` },
        { start: found.statement.getStart(file), end: found.statement.end, text: binding.adapted },
        // From the end of the last statement before the declaration to the end of the file: `Date.now;\r\n` again.
        { start: declaration.getFullStart(), end: text.length, text: "\r\n" },
    ];
    const types = [
        { start: found.owner.getStart(file), end: found.owner.body.getStart(file), text: header.adapted },
        { start: found.member.getStart(file), end: found.member.end, text: timeOrigin.adapted },
    ];
    const evidence = [
        { site: "perf_hooks import", file: performanceName, line: where(found.anchor) + 1, before: "", after: importText },
        { site: "tryGetPerformance binding", file: performanceName, line: where(found.statement), before: binding.original, after: binding.adapted },
        { site: "require declaration", file: performanceName, line: where(declaration), before: requireText, after: "" },
        { site: "tryGetPerformance return type", file: performanceName, line: where(found.owner), before: header.original, after: header.adapted },
        { site: "PerformanceTime timeOrigin", file: performanceName, line: where(found.member), before: timeOrigin.original, after: timeOrigin.adapted },
    ];
    const untyped = apply(text, edits);
    const after = apply(text, [...edits, ...types]);
    // Re-read the planned text: it must parse, be CRLF throughout, and be exactly this adaptation's state.
    const replanned = ts.createSourceFile(fileName, after, ts.ScriptTarget.Latest, true);
    assert.equal(replanned.parseDiagnostics.length, 0, "adapted syntax error");
    assert.equal(after.split("\r\n").length, after.split("\n").length, "the adapted performanceCore.ts has a line break that is not CRLF");
    assert.equal(fileState(ts, replanned).state, "adapted", "the planned text is not this adaptation's");
    provesJavaScript(ts, text, untyped, after);
    provesDeclarations(ts, fileName, text, after);
    // The planned tree type-checks as before under 47's program: the same diagnostics plus at most the unresolved
    // specifier's, which sys.ts's node: imports share, the perfHooks read bound to the import, and the return type's
    // Performance the file's own, with timeOrigin readonly.
    const planned = plannedProgram(ts, program, fileName, after);
    const unresolved = plannedBinding(ts, planned, fileName, path.join(tree, "src/compiler/sys.ts"));
    localView(ts, planned, fileName);
    assert.deepEqual(diagnosticsOf(ts, planned, fileName), [...diagnosticsOf(ts, program, fileName), ...unresolved].sort(), "performanceCore.ts's diagnostics changed");
    evidence.push({ site: "47's program", file: performanceName, before: diagnosticsOf(ts, program, fileName), after: diagnosticsOf(ts, planned, fileName) });
    return { fileName, before: text, after, evidence };
}

function adapt(tree, ts) {
    const change = plan(tree, ts);
    assert.equal(fs.readFileSync(change.fileName, "utf8"), change.before, "source changed while planning");
    if (change.after !== change.before) fs.writeFileSync(change.fileName, change.after);
    return { files: change.after === change.before ? 0 : 1, javascriptChanging: true, evidence: change.evidence };
}

module.exports = { adapt, plan, binding, header, timeOrigin, importText, anchorText };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
