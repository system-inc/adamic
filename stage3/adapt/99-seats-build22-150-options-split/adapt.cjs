#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
const file = path.join(process.argv[2], "src/compiler/commandLineParser.ts");
const text = fs.readFileSync(file, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0);

function only(found, name) {
    assert.equal(found.length, 1, `expected one ${name}, found ${found.length}`);
    return found[0];
}

const worker = only(source.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === "parseCommandLineWorker"),
"parseCommandLineWorker");
const parseStrings = only(worker.body.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === "parseStrings"),
"parseStrings");
const watchBranch = only(nodes(ts, parseStrings, node =>
    ts.isIfStatement(node) && ts.isIdentifier(node.expression) && node.expression.text === "watchOpt"),
"watchOpt branch");
const branch = watchBranch.thenStatement;

// Collect lexical bindings, including names declared through destructuring.
const bindingKinds = new Set([
    ts.SyntaxKind.VariableDeclaration, ts.SyntaxKind.BindingElement, ts.SyntaxKind.Parameter,
    ts.SyntaxKind.FunctionDeclaration, ts.SyntaxKind.FunctionExpression,
    ts.SyntaxKind.ClassDeclaration, ts.SyntaxKind.ClassExpression,
    ts.SyntaxKind.EnumDeclaration, ts.SyntaxKind.ModuleDeclaration,
    ts.SyntaxKind.ImportClause, ts.SyntaxKind.ImportSpecifier, ts.SyntaxKind.NamespaceImport,
    ts.SyntaxKind.ImportEqualsDeclaration, ts.SyntaxKind.TypeAliasDeclaration,
    ts.SyntaxKind.InterfaceDeclaration, ts.SyntaxKind.TypeParameter,
]);

function isBinding(node) {
    return bindingKinds.has(node.kind) && node.name && ts.isIdentifier(node.name);
}

function bindingScope(node) {
    if (ts.isFunctionExpression(node) || ts.isClassExpression(node)) return node;
    while (ts.isBindingElement(node)) node = node.parent.parent;
    const list = node.parent;
    const functionScoped = ts.isVariableDeclaration(node) && ts.isVariableDeclarationList(list)
        && !(list.flags & ts.NodeFlags.BlockScoped);
    for (let parent = list; parent; parent = parent.parent) {
        if (ts.isSourceFile(parent) || ts.isFunctionLike(parent)
            || ts.isModuleBlock(parent) || ts.isClassStaticBlockDeclaration(parent)
            || (!functionScoped && ts.isBlockScope(parent, parent.parent))) return parent;
    }
}

function variableStatement(node, block, flags) {
    if (!ts.isVariableDeclaration(node)) return undefined;
    const list = node.parent;
    const statement = list.parent;
    if (ts.isVariableDeclarationList(list) && list.declarations.length === 1
        && (list.flags & ts.NodeFlags.BlockScoped) === flags
        && ts.isVariableStatement(statement) && statement.parent === block) return statement;
}

const slots = nodes(ts, worker, node => isBinding(node) && node.name.text === "watchOptions");
const slot = slots[0];
assert(slots.length === 1 && !slot.initializer
    && slot.type?.getText(source) === "WatchOptions | undefined"
    && variableStatement(slot, worker.body, ts.NodeFlags.Let)?.getText(source)
        === "let watchOptions: WatchOptions | undefined;", "watchOptions declaration drift");

const branchScopes = new Set();
for (let node = watchBranch; node; node = node.parent) {
    assert(!ts.isWithStatement(node), "watchOpt branch is inside a with statement");
    branchScopes.add(node);
}
const bindings = nodes(ts, source, isBinding);
function pinBinding(name, check) {
    const matches = bindings.filter(node => node.name.text === name && branchScopes.has(bindingScope(node)));
    assert(matches.length === 1 && check(matches[0]), `${name} binding drift`);
    return matches[0];
}

const callee = pinBinding("parseOptionValue", node =>
    ts.isFunctionDeclaration(node) && node.parent === source && node.body);
const options = callee.parameters[4];
assert(options && ts.isIdentifier(options.name) && options.name.text === "options"
    && !options.questionToken && !options.initializer && !options.dotDotDotToken
    && options.type?.getText(source) === "OptionsBase", "parseOptionValue options parameter drift");
assert.equal(text.slice(callee.getStart(source), callee.body.getStart(source) + 1), [
    "function parseOptionValue(",
    "    args: readonly string[],",
    "    i: number,",
    "    diagnostics: ParseCommandLineWorkerDiagnostics,",
    "    opt: CommandLineOption,",
    "    options: OptionsBase,",
    "    errors: Diagnostic[],",
    ") {",
].join("\r\n"), "parseOptionValue declaration drift");

pinBinding("args", node => parseStrings.parameters.length === 1
    && parseStrings.parameters[0] === node && node.getText(source) === "args: readonly string[]");
pinBinding("i", node =>
    variableStatement(node, parseStrings.body, ts.NodeFlags.Let)?.getText(source) === "let i = 0;");
pinBinding("watchOptionsDidYouMeanDiagnostics", node =>
    variableStatement(node, source, ts.NodeFlags.Const)?.getText(source) === [
        "const watchOptionsDidYouMeanDiagnostics: ParseCommandLineWorkerDiagnostics = {",
        "    getOptionsNameMap: getWatchOptionsNameMap,",
        "    optionDeclarations: optionsForWatch,",
        "    unknownOptionDiagnostic: Diagnostics.Unknown_watch_option_0,",
        "    unknownDidYouMeanDiagnostic: Diagnostics.Unknown_watch_option_0_Did_you_mean_1,",
        "    optionTypeMismatchDiagnostic: Diagnostics.Watch_option_0_requires_a_value_of_type_1,",
        "};",
    ].join("\r\n"));
pinBinding("watchOpt", node => ts.isBlock(watchBranch.parent) && node.end < watchBranch.getStart(source)
    && variableStatement(node, watchBranch.parent, ts.NodeFlags.Const)?.getText(source)
        === "const watchOpt = getOptionDeclarationFromName(watchOptionsDidYouMeanDiagnostics.getOptionsNameMap, inputOptionName, /*allowShort*/ true);");

const before = [
    "{",
    "                        i = parseOptionValue(args, i, watchOptionsDidYouMeanDiagnostics, watchOpt, watchOptions || (watchOptions = {}), errors);",
    "                    }",
].join("\r\n");
const after = [
    "{",
    "                        if (watchOptions === undefined) watchOptions = {};",
    "                        i = parseOptionValue(args, i, watchOptionsDidYouMeanDiagnostics, watchOpt, watchOptions, errors);",
    "                    }",
].join("\r\n");
const snapshot = branch.getText(source);
if (snapshot !== after) {
    assert.equal(snapshot, before, "watchOpt branch snapshot changed");
    const start = branch.getStart(source);
    assert.equal(fs.readFileSync(file, "utf8"), text, "source changed during adaptation");
    fs.writeFileSync(file, text.slice(0, start) + after + text.slice(branch.end));
}
