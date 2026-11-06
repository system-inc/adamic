#!/usr/bin/env node
"use strict";
// tsc's two reads of an ambient require: isNodeLikeSystem tests `typeof require !== "undefined"` (core.ts), and
// tryGetPerformance calls `require("perf_hooks")` under it (performanceCore.ts). Adamic refuses an ambient value
// declaration, and ruled (Nov 6, #00qn5az) one JavaScript-changing adaptation: isNodeLikeSystem drops its require
// test and core.ts its `declare const require: unknown;`, and performanceCore.ts reads perf_hooks through a static
// `node:perf_hooks` namespace import, in 93's idiom, and drops its `declare const require`. 99q already reads
// `!process.browser` through the declared `browser?: unknown`, so this adaptation runs after it and leaves it.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");

const crlfLines = lines => lines.join("\r\n");
const coreName = "src/compiler/core.ts";
const performanceName = "src/compiler/performanceCore.ts";

// core.ts: isNodeLikeSystem as 99q leaves it and as this adaptation writes it, and the declarations above it.
const nodeLikeHead = "export function isNodeLikeSystem(): boolean";
const nodeLikeBody = {
    original: crlfLines([
        "{",
        "    // This is defined here rather than in sys.ts to prevent a cycle from its",
        "    // use in performanceCore.ts.",
        "    return typeof process !== \"undefined\"",
        "        && !!process.nextTick",
        "        && !process.browser",
        "        && typeof require !== \"undefined\";",
        "}",
    ]),
    adapted: crlfLines([
        "{",
        "    // This is defined here rather than in sys.ts to prevent a cycle from its",
        "    // use in performanceCore.ts.",
        "    return typeof process !== \"undefined\"",
        "        && !!process.nextTick",
        "        && !process.browser;",
        "}",
    ]),
};
// 30's process view, which isNodeLikeSystem reads in both states, and 30's require declaration on the same line.
const processText = "declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;";
const coreRequireText = "declare const require: unknown;";

// performanceCore.ts: the perf_hooks binding in tryGetPerformance as 65 leaves it and as this adaptation writes it.
const anchorText = "import { isNodeLikeSystem } from \"./_namespaces/ts.js\";";
const importText = "import * as perfHooks from \"node:perf_hooks\";";
const local = "perfHooks";
const binding = {
    original: "const { performance } = require(\"perf_hooks\") as { performance?: Performance | undefined };",
    adapted: "const { performance } = perfHooks;",
};
// 65 appends its declaration with LF line breaks to the CRLF file; removing it gives back the file's own ending.
const performanceRequireText = "declare const require: (specifier: \"perf_hooks\") => any;";
const performanceRequireTrivia = "\r\n\n";

function one(found, label) {
    assert.equal(found.length, 1, `${label}: want one, got ${found.length}`);
    return found[0];
}

function transpile(ts, text, removeComments) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
}

// Both files are CRLF throughout, except 65's appended declaration in performanceCore.ts, which is LF.
function crlf(text, label) {
    assert.equal(text.split("\r\n").length, text.split("\n").length, `${label} has a line break that is not CRLF`);
}

function textState(text, states, label) {
    const found = Object.keys(states).filter(state => states[state] === text);
    assert.equal(found.length, 1, `uninventoried ${label}: ${text}`);
    return found[0];
}

// A top-level `declare const <name>` statement, or none.
function ambient(ts, file, name) {
    return file.statements.filter(node => ts.isVariableStatement(node)
        && node.declarationList.declarations.some(declaration => ts.isIdentifier(declaration.name) && declaration.name.text === name));
}

function identifiers(ts, file, name) {
    return nodes(ts, file, node => ts.isIdentifier(node) && node.text === name);
}

// core.ts: isNodeLikeSystem's body, and the require declaration, present exactly as 30 left it beside the process
// view, or absent. process is read in both states through the same declaration.
function coreState(ts, file) {
    const declaration = one(nodes(ts, file, node => ts.isFunctionDeclaration(node) && node.name?.text === "isNodeLikeSystem"), "isNodeLikeSystem");
    assert(file.statements.includes(declaration), "isNodeLikeSystem is not at module scope");
    assert(declaration.body, "isNodeLikeSystem has no body");
    assert.equal(file.text.slice(declaration.getStart(file), declaration.body.getStart(file)).trimEnd(), nodeLikeHead, "isNodeLikeSystem: unreviewed signature");
    const body = textState(declaration.body.getText(file), nodeLikeBody, "isNodeLikeSystem body");
    const processStatement = one(ambient(ts, file, "process"), "process declaration");
    assert.equal(processStatement.getText(file), processText, "unreviewed process declaration");
    const requires = ambient(ts, file, "require");
    let requireDeclaration = { state: "adapted" };
    if (requires.length) {
        const statement = one(requires, "core.ts require declaration");
        assert.equal(statement.getText(file), coreRequireText, "unreviewed core.ts require declaration");
        assert.equal(file.statements[file.statements.indexOf(statement) - 1], processStatement, "the require declaration no longer follows the process view");
        assert.equal(file.text.slice(statement.getFullStart(), statement.getStart(file)), " ", "unreviewed text before the core.ts require declaration");
        requireDeclaration = { state: "original", statement };
    }
    const pieces = { "isNodeLikeSystem body": body, "require declaration": requireDeclaration.state };
    const states = Object.values(pieces);
    assert(states.every(state => state === states[0]), `half-adapted core.ts: ${Object.entries(pieces).map(([name, state]) => `${name} ${state}`).join(", ")}`);
    // Before, require is named twice (the declaration and the test); after, nowhere in core.ts.
    assert.equal(identifiers(ts, file, "require").length, states[0] === "original" ? 2 : 0, "core.ts names require outside the reviewed sites");
    return { state: states[0], declaration, requireDeclaration };
}

// performanceCore.ts: the namespace import right after the one ./_namespaces/ts.js import, the binding in
// tryGetPerformance, and 65's require declaration at the end of the file, each present or absent together.
function performanceState(ts, file) {
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
    const requires = ambient(ts, file, "require");
    let requireDeclaration = { state: "adapted" };
    if (requires.length) {
        const declaration = one(requires, "performanceCore.ts require declaration");
        assert.equal(declaration.getText(file), performanceRequireText, "unreviewed performanceCore.ts require declaration");
        assert.equal(file.statements.at(-1), declaration, "the performanceCore.ts require declaration is not the last statement");
        assert.equal(file.text.slice(declaration.getFullStart(), declaration.getStart(file)), performanceRequireTrivia, "unreviewed text before the performanceCore.ts require declaration");
        assert.equal(file.text.slice(declaration.end), "\n", "unreviewed text after the performanceCore.ts require declaration");
        requireDeclaration = { state: "original", declaration };
    }
    const pieces = { import: importState, binding: bindingState, "require declaration": requireDeclaration.state };
    const states = Object.values(pieces);
    assert(states.every(state => state === states[0]), `half-adapted performanceCore.ts: ${Object.entries(pieces).map(([name, state]) => `${name} ${state}`).join(", ")}`);
    // Before, require is named twice (the call and the declaration) and the local nowhere; after, require nowhere
    // and the local twice (the import binding and the destructured read).
    const original = states[0] === "original";
    assert.equal(identifiers(ts, file, "require").length, original ? 2 : 0, "performanceCore.ts names require outside the reviewed sites");
    assert.equal(identifiers(ts, file, local).length, original ? 0 : 2, `performanceCore.ts names ${local} outside the reviewed sites`);
    return { state: states[0], anchor, statement, requireDeclaration };
}

// Before the edit, by symbol: each file's require declaration is read only at the site being replaced, and that
// read resolves to the file's own declaration, never to a global.
function onlyAtSites(ts, checker, core, performance) {
    const pairs = [
        { label: "core.ts require", file: core.file, declaration: core.found.requireDeclaration.statement, site: core.found.declaration.body },
        { label: "performanceCore.ts require", file: performance.file, declaration: performance.found.requireDeclaration.declaration, site: performance.found.statement },
    ];
    for (const pair of pairs) {
        const name = pair.declaration.declarationList.declarations[0].name;
        const symbol = checker.getSymbolAtLocation(name);
        assert(symbol, `${pair.label} has no symbol`);
        const uses = identifiers(ts, pair.file, "require").filter(node => node !== name);
        assert.equal(uses.length, 1, `${pair.label}: want one read`);
        assert.equal(checker.getSymbolAtLocation(uses[0]), symbol, `${pair.label}: the read is not the file's own declaration`);
        assert(uses[0].pos >= pair.site.pos && uses[0].end <= pair.site.end, `${pair.label} is read outside the replaced site`);
    }
}

// The planned perfHooks read resolves to the import. When the program has Node's types, the import resolves to
// @types/node's `node:perf_hooks`, which exports performance; when it has none, as 47's program over an adapted
// tree has none (stage3 never runs npm), the specifier draws exactly the diagnostic sys.ts's `node:fs` import
// draws in the same program, with the name changed, and nothing else changes. Returns that one diagnostic, or none.
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

function plannedProgram(ts, program, texts) {
    const host = ts.createCompilerHost(program.getCompilerOptions(), true);
    const original = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, fresh) => texts.has(path.resolve(name))
        ? ts.createSourceFile(name, texts.get(path.resolve(name)), language, true)
        : original(name, language, onError, fresh);
    return ts.createProgram(program.getRootFileNames(), program.getCompilerOptions(), host);
}

// Each file's diagnostics under 47's program, by code and message, with the line numbers left out: an edit moves
// lines, and a removed line moves those below it.
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

// Stock 6.0.3's JavaScript of each planned file is upstream's with exactly the ruled pieces replaced. The expected
// text is built from the before output, never from source.
function provesJavaScript(ts, core, performance) {
    for (const removeComments of [false, true]) {
        let expected = transpile(ts, core.before, removeComments);
        expected = replaceOnce(expected, "        && !process.browser\n        && typeof require !== \"undefined\";\n", "        && !process.browser;\n", "the require test");
        assert.equal(transpile(ts, core.after, removeComments), expected, "core.ts's JavaScript differs by more than the require test");
        expected = transpile(ts, performance.before, removeComments);
        expected = replaceOnce(expected, `${anchorText}\n`, `${anchorText}\n${importText}\n`, "the ./_namespaces/ts.js import");
        expected = replaceOnce(expected, "const { performance } = require(\"perf_hooks\");\n", `${binding.adapted}\n`, "the perf_hooks require");
        assert.equal(transpile(ts, performance.after, removeComments), expected, "performanceCore.ts's JavaScript differs by more than the import and the binding");
    }
}

// No declaration changes: isNodeLikeSystem keeps its signature, tryGetPerformance is not exported, and neither the
// require declarations nor the perf_hooks import reach a declaration. With stripInternal, as typescript.d.ts is
// built, and without it.
function provesDeclarations(ts, change) {
    for (const stripInternal of [true, false]) {
        const declare = text => ts.transpileDeclaration(text, { compilerOptions: { stripInternal }, fileName: path.basename(change.fileName) }).outputText;
        assert.equal(declare(change.after), declare(change.before), `${path.basename(change.fileName)}'s declarations changed (stripInternal ${stripInternal})`);
    }
}

function read(tree, name) {
    const fileName = path.join(tree, name);
    return { fileName, text: fs.readFileSync(fileName, "utf8") };
}

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const files = [read(tree, coreName), read(tree, performanceName)];
    for (const file of files) {
        // 65's appended declaration is the one LF line; the rest of each file is CRLF.
        const text = file.text.endsWith(`${performanceRequireTrivia}${performanceRequireText}\n`) ? file.text.slice(0, -(performanceRequireText.length + 2)) : file.text;
        crlf(text, path.basename(file.fileName));
        file.parsed = ts.createSourceFile(file.fileName, file.text, ts.ScriptTarget.Latest, true);
        assert.equal(file.parsed.parseDiagnostics.length, 0, `cannot parse ${path.basename(file.fileName)}`);
    }
    const [core, performance] = files;
    const states = [coreState(ts, core.parsed).state, performanceState(ts, performance.parsed).state];
    assert(states.every(state => state === states[0]), `half-adapted tree: core.ts ${states[0]}, performanceCore.ts ${states[1]}`);
    if (states[0] === "adapted") return { changes: files.map(file => ({ fileName: file.fileName, before: file.text, after: file.text })), evidence: [] };
    // Only an original tree needs the checker, to prove the removed declarations are read nowhere else.
    const program = compiler(tree, ts);
    const checker = program.getTypeChecker();
    for (const file of files) {
        file.checked = program.getSourceFile(file.fileName);
        assert(file.checked && file.checked.text === file.text, `the program's ${path.basename(file.fileName)} is not the file on disk`);
    }
    core.found = coreState(ts, core.checked);
    performance.found = performanceState(ts, performance.checked);
    onlyAtSites(ts, checker, { file: core.checked, found: core.found }, { file: performance.checked, found: performance.found });
    const where = (file, node) => file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1;
    const evidence = [];
    const coreEdits = [];
    coreEdits.push({ start: core.found.declaration.body.getStart(core.checked), end: core.found.declaration.body.end, text: nodeLikeBody.adapted });
    evidence.push({ site: "isNodeLikeSystem", file: coreName, line: where(core.checked, core.found.declaration), before: nodeLikeBody.original, after: nodeLikeBody.adapted });
    const coreRequire = core.found.requireDeclaration.statement;
    coreEdits.push({ start: coreRequire.getFullStart(), end: coreRequire.end, text: "" });
    evidence.push({ site: "core.ts require declaration", file: coreName, line: where(core.checked, coreRequire), before: ` ${coreRequireText}`, after: "" });
    const performanceEdits = [];
    performanceEdits.push({ start: performance.found.anchor.end, end: performance.found.anchor.end, text: `\r\n${importText}` });
    evidence.push({ site: "perf_hooks import", file: performanceName, line: where(performance.checked, performance.found.anchor) + 1, before: "", after: importText });
    performanceEdits.push({ start: performance.found.statement.getStart(performance.checked), end: performance.found.statement.end, text: binding.adapted });
    evidence.push({ site: "tryGetPerformance binding", file: performanceName, line: where(performance.checked, performance.found.statement), before: binding.original, after: binding.adapted });
    const performanceRequire = performance.found.requireDeclaration.declaration;
    performanceEdits.push({ start: performanceRequire.getFullStart(), end: performance.text.length, text: "\r\n" });
    evidence.push({ site: "performanceCore.ts require declaration", file: performanceName, line: where(performance.checked, performanceRequire), before: performanceRequireText, after: "" });
    core.after = apply(core.text, coreEdits);
    performance.after = apply(performance.text, performanceEdits);
    // Re-read the planned texts: each must parse, be CRLF throughout, and be exactly this adaptation's state.
    for (const file of files) {
        const replanned = ts.createSourceFile(file.fileName, file.after, ts.ScriptTarget.Latest, true);
        assert.equal(replanned.parseDiagnostics.length, 0, `adapted syntax error in ${path.basename(file.fileName)}`);
        crlf(file.after, `adapted ${path.basename(file.fileName)}`);
        file.replanned = replanned;
    }
    assert.equal(coreState(ts, core.replanned).state, "adapted", "the planned core.ts is not this adaptation's");
    assert.equal(performanceState(ts, performance.replanned).state, "adapted", "the planned performanceCore.ts is not this adaptation's");
    provesJavaScript(ts, { before: core.text, after: core.after }, { before: performance.text, after: performance.after });
    for (const file of files) provesDeclarations(ts, { fileName: file.fileName, before: file.text, after: file.after });
    // The planned tree type-checks as before under 47's program: core.ts's diagnostics are the same, and
    // performanceCore.ts's are the same plus at most the unresolved specifier's, which sys.ts's node: imports share.
    const planned = plannedProgram(ts, program, new Map(files.map(file => [path.resolve(file.fileName), file.after])));
    const unresolved = plannedBinding(ts, planned, performance.fileName, path.join(tree, "src/compiler/sys.ts"));
    assert.deepEqual(diagnosticsOf(ts, planned, core.fileName), diagnosticsOf(ts, program, core.fileName), "core.ts's diagnostics changed");
    assert.deepEqual(diagnosticsOf(ts, planned, performance.fileName), [...diagnosticsOf(ts, program, performance.fileName), ...unresolved].sort(), "performanceCore.ts's diagnostics changed");
    evidence.push({ site: "47's program", file: performanceName, before: diagnosticsOf(ts, program, performance.fileName), after: diagnosticsOf(ts, planned, performance.fileName) });
    return { changes: files.map(file => ({ fileName: file.fileName, before: file.text, after: file.after })), evidence };
}

function adapt(tree, ts) {
    const change = plan(tree, ts);
    for (const file of change.changes) assert.equal(fs.readFileSync(file.fileName, "utf8"), file.before, "source changed while planning");
    const written = change.changes.filter(file => file.after !== file.before);
    for (const file of written) fs.writeFileSync(file.fileName, file.after);
    return { files: written.length, javascriptChanging: true, evidence: change.evidence };
}

module.exports = { adapt, plan, nodeLikeBody, binding, importText };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
