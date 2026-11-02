#!/usr/bin/env node
"use strict";
// program.ts's two TS18046 rows: `onError(e.message)` in the catch of createGetSourceFile and of
// createWriteFileMeasuringIO. Each read goes through 94's propertyOf and is converted with the global
// String, "convert, don't cast" (Adamic, Nov 1, #axochk): onError stays `(message: string) => void` in
// types.ts, and now receives text for every thrown value that reaches it.
// Needs adaptation 94: the caught-value read goes through its propertyOf.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");
const { helper } = require("../94-axolotl-property-read/adapt.cjs");

const functions = ["createGetSourceFile", "createWriteFileMeasuringIO"];
const readAfter = "String(propertyOf(e, \"message\"))";
const anchorText = "import * as performance from \"./_namespaces/ts.performance.js\";";
const importText = "import { propertyOf } from \"./hostErrors.js\";";

function one(found, label) {
    assert.equal(found.length, 1, `${label}: want one, got ${found.length}`);
    return found[0];
}

function transpile(ts, text, removeComments) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
}

// program.ts is CRLF throughout; every line break before and after is one.
function crlf(text, label) {
    assert.equal(text.split("\r\n").length, text.split("\n").length, `${label} has a line break that is not CRLF`);
}

// The read inside onError's one argument, as upstream writes it or as 97e writes it. The withdrawn
// `propertyOf(e, "message") as string` and a bare propertyOf call are shapes 97e never writes.
function readState(ts, file, argument, caught) {
    const original = ts.isPropertyAccessExpression(argument) && caught(argument.expression) && argument.name.text === "message";
    if (original) return { state: "original", read: argument };
    assert(!ts.isAsExpression(argument), `uninventoried onError argument, the withdrawn cast: ${argument.getText(file)}`);
    const read = ts.isCallExpression(argument) && ts.isIdentifier(argument.expression) && argument.expression.text === "String" && argument.arguments.length === 1 ? argument.arguments[0] : undefined;
    const adapted = read !== undefined && ts.isCallExpression(read) && read.expression.getText(file) === "propertyOf"
        && read.arguments.length === 2 && caught(read.arguments[0])
        && ts.isStringLiteral(read.arguments[1]) && read.arguments[1].text === "message";
    assert(adapted, `uninventoried onError argument: ${argument.getText(file)}`);
    return { state: "adapted", read: argument };
}

// String is the global one: program.ts declares no String of its own, and at the site the value named
// String resolves only to the default library's declarations.
function globalString(ts, file, program, checker, node) {
    const bindings = [
        ts.SyntaxKind.VariableDeclaration, ts.SyntaxKind.Parameter, ts.SyntaxKind.BindingElement,
        ts.SyntaxKind.FunctionDeclaration, ts.SyntaxKind.FunctionExpression, ts.SyntaxKind.ClassDeclaration, ts.SyntaxKind.ClassExpression,
        ts.SyntaxKind.EnumDeclaration, ts.SyntaxKind.ModuleDeclaration,
        ts.SyntaxKind.ImportClause, ts.SyntaxKind.NamespaceImport, ts.SyntaxKind.ImportSpecifier, ts.SyntaxKind.ImportEqualsDeclaration,
    ];
    const declared = nodes(ts, file, child => ts.isIdentifier(child) && child.text === "String" && bindings.includes(child.parent.kind) && child.parent.name === child);
    assert.equal(declared.length, 0, "program.ts declares its own String");
    if (!checker) return;
    const symbol = checker.getSymbolsInScope(node, ts.SymbolFlags.Value).find(candidate => candidate.name === "String");
    assert(symbol?.declarations?.length && symbol.declarations.every(declaration => program.isSourceFileDefaultLibrary(declaration.getSourceFile())), "String at the site is not the global String");
}

// One site: the named function's one catch clause, its binding e, read once, as onError's argument
// inside `if (onError) { ... }`. checker and program are absent when re-reading the planned text.
function site(ts, file, name, checker, program) {
    const declaration = one(file.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name), name);
    const clause = one(nodes(ts, declaration, ts.isCatchClause), `${name} catch`);
    const binding = clause.variableDeclaration;
    assert(binding && ts.isIdentifier(binding.name) && binding.name.text === "e" && !binding.type, `${name}: unreviewed catch binding`);
    let caught = node => ts.isIdentifier(node) && node.text === "e";
    if (checker) {
        assert(checker.getTypeAtLocation(binding.name).flags & ts.TypeFlags.Unknown, `${name}: catch is not unknown`);
        const symbol = checker.getSymbolAtLocation(binding.name);
        caught = node => ts.isIdentifier(node) && checker.getSymbolAtLocation(node) === symbol;
    }
    const use = one(nodes(ts, clause.block, caught), `${name}: reads of the caught value`);
    // Upstream's e.message is the argument itself. 97e's propertyOf(e, ...) sits inside String(...), and
    // the withdrawn form inside `as string`; readState refuses that, and a bare propertyOf call.
    const read = use.parent;
    const wrapped = ts.isCallExpression(read) && (ts.isAsExpression(read.parent) || (ts.isCallExpression(read.parent) && read.parent.expression.getText(file) === "String"));
    const argument = wrapped ? read.parent : read;
    const call = argument.parent;
    assert(ts.isCallExpression(call) && ts.isIdentifier(call.expression) && call.expression.text === "onError" && call.arguments.length === 1 && call.arguments[0] === argument, `${name}: the caught value's read is not onError's one argument`);
    const statement = call.parent;
    assert(ts.isExpressionStatement(statement) && ts.isBlock(statement.parent) && statement.parent.statements.length === 1, `${name}: unreviewed onError call statement`);
    const guard = statement.parent.parent;
    assert(ts.isIfStatement(guard) && !guard.elseStatement && ts.isIdentifier(guard.expression) && guard.expression.text === "onError", `${name}: onError is not called under if (onError)`);
    if (checker) {
        // onError is the returned arrow's own parameter, declared (message: string) => void by types.ts.
        const callee = checker.getSymbolAtLocation(call.expression);
        assert(callee?.valueDeclaration && ts.isParameter(callee.valueDeclaration), `${name}: onError is not the callback's parameter`);
        const signature = checker.getResolvedSignature(call);
        assert.equal(signature.parameters.length, 1, `${name}: onError's signature changed`);
        assert.equal(checker.typeToString(checker.getTypeOfSymbol(signature.parameters[0])), "string", `${name}: onError's message is no longer declared string`);
    }
    globalString(ts, file, program, checker, call);
    return { name, ...readState(ts, file, argument, caught) };
}

// The hostErrors.js import, absent or exactly 97e's, directly after the performance import as 47 places its own.
function importState(ts, file) {
    const anchor = one(file.statements.filter(node => node.getText(file) === anchorText), "ts.performance.js import");
    const index = file.statements.indexOf(anchor);
    const imports = file.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === "./hostErrors.js");
    if (!imports.length) return { state: "original", anchor };
    const declaration = one(imports, "hostErrors.js import");
    assert.equal(declaration.getText(file), importText, "unreviewed hostErrors.js import");
    assert.equal(file.statements[index + 1], declaration, "the hostErrors.js import is not after the ts.performance.js import");
    assert.equal(file.text.slice(anchor.end, declaration.getStart(file)), "\r\n", "unreviewed text between the two imports");
    return { state: "adapted", anchor, declaration };
}

// The whole file's state: both sites and the import upstream's, or all three 97e's. Anything else is half done.
function fileState(ts, file, checker, program) {
    const sites = functions.map(name => site(ts, file, name, checker, program));
    const imported = importState(ts, file);
    const states = [...sites.map(entry => entry.state), imported.state];
    assert(states.every(state => state === states[0]), `half-adapted program.ts: ${functions.map((name, index) => `${name} ${states[index]}`).join(", ")}, import ${imported.state}`);
    // propertyOf is named only where 97e names it: nowhere before, the two reads and the import after.
    const named = nodes(ts, file, node => ts.isIdentifier(node) && node.text === "propertyOf").length;
    assert.equal(named, states[0] === "original" ? 0 : 3, "propertyOf named outside a reviewed site");
    return { state: states[0], sites, imported };
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

// Stock 6.0.3's JavaScript of the planned text is upstream's with exactly the two reads converted and the
// import added. The expected text is built from the before output, never from source.
function provesJavaScript(ts, before, after) {
    const readBefore = "onError(e.message);";
    const readCall = "onError(String(propertyOf(e, \"message\")));";
    for (const removeComments of [false, true]) {
        const original = transpile(ts, before, removeComments);
        assert.equal(original.split(readBefore).length - 1, 2, "upstream JavaScript lacks the two reads");
        assert.equal(original.split(`${anchorText}\n`).length - 1, 1, "upstream JavaScript lacks the ts.performance.js import");
        const expected = original.split(readBefore).join(readCall).replace(`${anchorText}\n`, `${anchorText}\n${importText}\n`);
        assert.equal(transpile(ts, after, removeComments), expected, "the adapted JavaScript differs by more than the two conversions and the import");
    }
}

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const hostErrors = path.join(tree, "src/compiler/hostErrors.ts");
    assert(fs.existsSync(hostErrors) && fs.readFileSync(hostErrors, "utf8").endsWith(helper), "hostErrors.ts lacks 94's propertyOf: adaptation 94 must run first");
    const program = compiler(tree, ts);
    const checker = program.getTypeChecker();
    const fileName = path.join(tree, "src/compiler/program.ts");
    const file = program.getSourceFile(fileName);
    assert(file && !file.parseDiagnostics.length, "cannot parse program.ts");
    crlf(file.text, "program.ts");
    const found = fileState(ts, file, checker, program);
    const evidence = [];
    const where = node => file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1;
    if (found.state === "adapted") return { fileName, before: file.text, after: file.text, evidence };
    const edits = [];
    for (const entry of found.sites) {
        edits.push({ start: entry.read.getStart(file), end: entry.read.end, text: readAfter });
        evidence.push({ function: entry.name, file: "src/compiler/program.ts", line: where(entry.read), before: entry.read.getText(file), after: readAfter });
    }
    edits.push({ start: found.imported.anchor.end, end: found.imported.anchor.end, text: `\r\n${importText}` });
    evidence.push({ import: true, file: "src/compiler/program.ts", line: where(found.imported.anchor) + 1, after: importText });
    const after = apply(file.text, edits);
    // Re-read the planned text: it must parse, stay CRLF, and be exactly 97e's whole state.
    const parsed = ts.createSourceFile(fileName, after, ts.ScriptTarget.Latest, true);
    assert.equal(parsed.parseDiagnostics.length, 0, "adapted syntax error");
    crlf(after, "adapted program.ts");
    assert.equal(fileState(ts, parsed).state, "adapted", "the planned text is not 97e's");
    provesJavaScript(ts, file.text, after);
    return { fileName, before: file.text, after, evidence };
}

function adapt(tree, ts) {
    const change = plan(tree, ts);
    assert.equal(fs.readFileSync(change.fileName, "utf8"), change.before, "source changed while planning");
    if (change.after !== change.before) fs.writeFileSync(change.fileName, change.after);
    return { files: change.after === change.before ? 0 : 1, javascriptChanging: true, evidence: change.evidence };
}

module.exports = { adapt, plan, transpile };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
