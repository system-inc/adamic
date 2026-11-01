#!/usr/bin/env node
"use strict";
// sys.ts's own-file checker diagnostics, rows 1, 3 and 4 of sites.json; row 2 is held.
// Needs adaptation 94: the caught-value read goes through its propertyOf.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");
const { helper } = require("../94-axolotl-property-read/adapt.cjs");

const profileArgument = "JSON.stringify(cleanupPaths(profile))";
const sourceMapSupport = "(require(\"source-map-support\") as typeof import(\"source-map-support\")).install()";
const pluginError = "{ stack?: string; message?: string; }";
const importBefore = "errorCode";
const importAfter = "errorCode, propertyOf";

function one(found, label) {
    assert.equal(found.length, 1, `${label}: want one, got ${found.length}`);
    return found[0];
}

function transpile(ts, text, removeComments) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
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

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const hostErrors = fs.readFileSync(path.join(tree, "src/compiler/hostErrors.ts"), "utf8");
    assert(hostErrors.endsWith(helper), "hostErrors.ts lacks 94's propertyOf: adaptation 94 must run first");
    const program = compiler(tree, ts);
    const checker = program.getTypeChecker();
    const source = name => {
        const file = program.getSourceFile(path.join(tree, "src/compiler", name));
        assert(file && !file.parseDiagnostics.length, `cannot parse ${name}`);
        return file;
    };
    const sys = source("sys.ts");
    const types = source("types.ts");
    const edits = new Map([[sys, []], [types, []]]);
    const erasing = new Map([[sys, []], [types, []]]);
    const evidence = [];
    function edit(file, node, text, erases, row) {
        const where = file.getLineAndCharacterOfPosition(node.getStart(file));
        evidence.push({ row, file: path.relative(tree, file.fileName), line: where.line + 1, before: node.getText(file), after: text });
        const change = { start: node.getStart(file), end: node.end, text };
        edits.get(file).push(change);
        if (erases) erasing.get(file).push(change);
    }

    // Row 1: the CPU profile's serialization is a required result.
    const disable = one(nodes(ts, sys, node => ts.isFunctionDeclaration(node) && node.name?.text === "disableCPUProfiler"), "disableCPUProfiler");
    const write = one(nodes(ts, disable, node => ts.isCallExpression(node) && node.expression.getText(sys) === "_fs.writeFileSync"), "profile write");
    assert.equal(write.arguments.length, 2, "unreviewed profile write");
    assert.equal(write.arguments[0].getText(sys), "profilePath", "unreviewed profile path");
    const serialized = write.arguments[1];
    if (ts.isNonNullExpression(serialized)) assert.equal(serialized.expression.getText(sys), profileArgument, "unreviewed profile argument");
    else {
        assert.equal(serialized.getText(sys), profileArgument, "unreviewed profile argument");
        edit(sys, serialized, `${profileArgument}!`, true, 1);
    }
    assert.equal(nodes(ts, sys, node => ts.isCallExpression(node) && node.expression.getText(sys) === "cleanupPaths").length, 1, "cleanupPaths has another caller");

    // Row 2 is held: the site must still read exactly as reviewed.
    const sourceMaps = one(nodes(ts, sys, node => ts.isMethodDeclaration(node) && node.name.getText(sys) === "tryEnableSourceMapsForHost"), "tryEnableSourceMapsForHost");
    one(nodes(ts, sourceMaps, node => ts.isCallExpression(node) && node.getText(sys) === sourceMapSupport), "held source-map-support site");

    // Row 3: the caught value's code is read with JavaScript's semantics.
    const watch = one(nodes(ts, sys, node => ts.isFunctionDeclaration(node) && node.name?.text === "watchPresentFileSystemEntry"), "watchPresentFileSystemEntry");
    const clause = one(nodes(ts, watch, ts.isCatchClause), "watcher catch");
    const binding = clause.variableDeclaration;
    assert(binding && ts.isIdentifier(binding.name) && binding.name.text === "e" && !binding.type, "unreviewed catch binding");
    assert(checker.getTypeAtLocation(binding.name).flags & ts.TypeFlags.Unknown, "catch is not unknown");
    const caught = checker.getSymbolAtLocation(binding.name);
    const uses = nodes(ts, clause.block, node => ts.isIdentifier(node) && checker.getSymbolAtLocation(node) === caught);
    const use = one(uses, "reads of the caught value");
    const read = use.parent;
    const original = ts.isPropertyAccessExpression(read) && read.expression === use && read.name.text === "code";
    const adapted = ts.isCallExpression(read) && read.expression.getText(sys) === "propertyOf" && read.arguments.length === 2 && read.arguments[0] === use && ts.isStringLiteral(read.arguments[1]) && read.arguments[1].text === "code";
    assert(original || adapted, "unreviewed caught-value read");
    const comparison = read.parent;
    assert(ts.isBinaryExpression(comparison) && comparison.left === read && comparison.operatorToken.kind === ts.SyntaxKind.EqualsEqualsEqualsToken && ts.isStringLiteral(comparison.right) && comparison.right.text === "ENOSPC", "unreviewed ENOSPC test");
    const assignment = comparison.parent;
    assert(ts.isBinaryExpression(assignment) && assignment.right === comparison && assignment.operatorToken.kind === ts.SyntaxKind.BarBarEqualsToken && assignment.left.getText(sys) === "hitSystemWatcherLimit", "unreviewed ||= assignment");
    if (original) edit(sys, read, "propertyOf(e, \"code\")", false, 3);

    // The import 47 added gains propertyOf, the way 47 imports errorCode.
    const imports = sys.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === "./hostErrors.js");
    const helpers = one(imports, "hostErrors.js import").importClause;
    assert(helpers && !helpers.isTypeOnly && !helpers.name && helpers.namedBindings && ts.isNamedImports(helpers.namedBindings), "unreviewed hostErrors.js import");
    const named = helpers.namedBindings.elements.map(element => element.getText(sys)).join(", ");
    assert([importBefore, importAfter].includes(named), `unreviewed hostErrors.js bindings: ${named}`);
    if (named === importBefore) {
        const elements = helpers.namedBindings.elements;
        edits.get(sys).push({ start: elements[0].getStart(sys), end: elements[elements.length - 1].end, text: importAfter });
        evidence.push({ row: 3, file: "src/compiler/sys.ts", line: sys.getLineAndCharacterOfPosition(elements[0].getStart(sys)).line + 1, before: importBefore, after: importAfter });
    }

    // Row 4: the @internal ModuleImportResult's failed member carries the caught value as it is.
    const alias = one(types.statements.filter(node => ts.isTypeAliasDeclaration(node) && node.name.text === "ModuleImportResult"), "ModuleImportResult");
    assert(ts.getJSDocTags(alias).some(tag => tag.tagName.text === "internal"), "ModuleImportResult is not @internal");
    assert(ts.isUnionTypeNode(alias.type) && alias.type.types.length === 2, "unreviewed ModuleImportResult shape");
    const [found, failed] = alias.type.types;
    assert.equal(found.getText(types), "{ module: T; modulePath?: string; error: undefined; }", "unreviewed ModuleImportResult success member");
    assert(ts.isTypeLiteralNode(failed), "unreviewed ModuleImportResult failure member");
    const error = one(failed.members.filter(member => member.name?.getText(types) === "error"), "failure member's error");
    assert(ts.isPropertySignature(error) && !error.questionToken && error.type, "unreviewed error property");
    assert(failed.getText(types).startsWith("{ module: undefined; modulePath?: undefined; error: "), "unreviewed ModuleImportResult failure member");
    const errorType = error.type.getText(types);
    assert(["unknown", pluginError].includes(errorType), `unreviewed error type: ${errorType}`);
    if (errorType === pluginError) edit(types, error.type, "unknown", true, 4);

    // Plan every file, then prove the erasing edits leave JavaScript byte for byte.
    const results = [];
    for (const [file, changes] of edits) {
        const after = apply(file.text, changes);
        assert.equal(ts.createSourceFile(file.fileName, after, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0, "adapted syntax error");
        const changing = changes.filter(change => !erasing.get(file).includes(change));
        const changedOnly = apply(file.text, changing);
        for (const removeComments of [false, true]) {
            assert.equal(transpile(ts, after, removeComments), transpile(ts, changedOnly, removeComments), `an erasing edit changed JavaScript in ${path.basename(file.fileName)}`);
        }
        results.push({ fileName: file.fileName, before: file.text, after });
    }
    return { results, evidence };
}

function adapt(tree, ts) {
    const { results, evidence } = plan(tree, ts);
    for (const result of results) assert.equal(fs.readFileSync(result.fileName, "utf8"), result.before, "source changed while planning");
    let files = 0;
    for (const result of results) if (result.after !== result.before) { fs.writeFileSync(result.fileName, result.after); files++; }
    return { files, javascriptChanging: true, held: ["row 2: source-map-support require"], evidence };
}

module.exports = { adapt, plan, transpile };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
