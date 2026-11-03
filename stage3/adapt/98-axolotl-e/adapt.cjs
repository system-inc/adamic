#!/usr/bin/env node
"use strict";
// sys.ts's last own-file diagnostic (1617:13) and the one fixing it unmasks (1524:15).
// Row 4: both ModuleImportResult failure members carry the caught value as unknown, and
// project.ts's two plugin-error readers read it through 94's propertyOf, reached through the
// compiler's ts namespace barrel as an @internal export. Row 5: System.createSHA256Hash
// becomes a property in adaptation 32's idiom. Needs 94's propertyOf and 98-c's sys.ts.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { nodes } = require("../47-host-errors/adapt.cjs");
const { helper } = require("../94-axolotl-property-read/adapt.cjs");

const pluginError = "{ stack?: string; message?: string; }";
const hashBefore = "createSHA256Hash?(data: string): string;";
const hashAfter = "createSHA256Hash?: { createSHA256Hash(data: string): string; }[\"createSHA256Hash\"] | undefined;";
const readerBefore = "result.error.stack || result.error.message || JSON.stringify(result.error)";
const readerAfter = "propertyOf(result.error, \"stack\") || propertyOf(result.error, \"message\") || JSON.stringify(result.error)";
const barrelAnchor = "export * from \"../sys.js\";";
// The barrel holds only `export *` lines, and upstream's dtsBundler.mjs refuses a named
// re-export there ("Failed to find namespace"). hostErrors.ts's three functions are all @internal.
const barrelExport = "export * from \"../hostErrors.js\";";
const readers = ["importServicePluginSync", "importServicePluginAsync"];

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

function internal(ts, node) {
    return ts.getJSDocTags(node).some(tag => tag.tagName.text === "internal");
}

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const hostErrors = fs.readFileSync(path.join(tree, "src/compiler/hostErrors.ts"), "utf8");
    assert(hostErrors.endsWith(helper), "hostErrors.ts lacks 94's propertyOf: adaptation 94 must run first");
    // No edit here needs the checker: each site is pinned by its parsed shape and text.
    const source = name => {
        const fileName = path.join(tree, "src", name);
        const file = ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
        assert(!file.parseDiagnostics.length, `cannot parse ${name}`);
        return file;
    };
    const sys = source("compiler/sys.ts");
    assert(!sys.text.includes("as typeof import(\"source-map-support\")") && sys.text.includes("propertyOf(e, \"code\")"), "sys.ts lacks 98-c's rows: adaptation 98-c must run first");
    const compilerTypes = source("compiler/types.ts");
    const serverTypes = source("server/types.ts");
    const barrel = source("compiler/_namespaces/ts.ts");
    const project = source("server/project.ts");
    const files = [sys, compilerTypes, serverTypes, barrel, project];
    const edits = new Map(files.map(file => [file, []]));
    const erasing = new Map(files.map(file => [file, []]));
    const evidence = [];
    function edit(file, start, end, text, erases, row) {
        evidence.push({ row, file: path.relative(tree, file.fileName), line: file.getLineAndCharacterOfPosition(start).line + 1, before: file.text.slice(start, end), after: text });
        const change = { start, end, text };
        edits.get(file).push(change);
        if (erases) erasing.get(file).push(change);
    }

    // Row 4, both failure members: the error is the caught value, unknown.
    function failureError(file, isInternal, success, failurePrefix) {
        const alias = one(file.statements.filter(node => ts.isTypeAliasDeclaration(node) && node.name.text === "ModuleImportResult"), `${file.fileName} ModuleImportResult`);
        assert.equal(internal(ts, alias), isInternal, `${path.basename(path.dirname(file.fileName))} ModuleImportResult's @internal moved`);
        assert(ts.isUnionTypeNode(alias.type) && alias.type.types.length === 2, "unreviewed ModuleImportResult shape");
        const [found, failed] = alias.type.types;
        assert.equal(found.getText(file), success, "unreviewed ModuleImportResult success member");
        assert(ts.isTypeLiteralNode(failed) && failed.getText(file).startsWith(failurePrefix), "unreviewed ModuleImportResult failure member");
        const error = one(failed.members.filter(member => member.name?.getText(file) === "error"), "failure member's error");
        assert(ts.isPropertySignature(error) && !error.questionToken && error.type, "unreviewed error property");
        const text = error.type.getText(file);
        assert(["unknown", pluginError].includes(text), `unreviewed error type: ${text}`);
        if (text === pluginError) edit(file, error.type.getStart(file), error.type.end, "unknown", true, 4);
    }
    failureError(compilerTypes, true, "{ module: T; modulePath?: string; error: undefined; }", "{ module: undefined; modulePath?: undefined; error: ");
    failureError(serverTypes, false, "{ module: {}; error: undefined; }", "{ module: undefined; error: ");

    // Row 4, the barrel: hostErrors.ts, every export of it @internal, beside the sys.js export.
    const exports = barrel.statements.filter(ts.isExportDeclaration);
    const existing = exports.filter(node => node.moduleSpecifier?.text === "../hostErrors.js");
    const helpers = fs.readFileSync(path.join(tree, "src/compiler/hostErrors.ts"), "utf8");
    const helperFile = ts.createSourceFile("hostErrors.ts", helpers, ts.ScriptTarget.Latest, true);
    assert(helperFile.statements.every(statement => ts.isFunctionDeclaration(statement) && internal(ts, statement)), "hostErrors.ts exports something not @internal");
    assert.deepEqual(helperFile.statements.map(statement => statement.name.text), ["errorCode", "errorMessage", "propertyOf"], "unreviewed hostErrors.ts exports");
    if (existing.length) assert.equal(one(existing, "hostErrors.js export").getText(barrel), barrelExport, "unreviewed hostErrors.js export");
    else {
        const anchor = one(exports.filter(node => node.getText(barrel) === barrelAnchor), "sys.js export");
        const newline = barrel.text.includes("\r\n") ? "\r\n" : "\n";
        edit(barrel, anchor.end, anchor.end, `${newline}${barrelExport}`, false, 4);
    }

    // Row 4, project.ts: the named import from the ts namespace gains propertyOf in sorted place,
    // and each reader reads the plugin error through it.
    const imports = project.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === "./_namespaces/ts.js" && node.importClause?.namedBindings && ts.isNamedImports(node.importClause.namedBindings));
    const named = one(imports, "named import from ./_namespaces/ts.js").importClause.namedBindings.elements;
    const names = named.map(element => element.getText(project));
    assert(names.every(name => /^[A-Za-z_$][\w$]*$/.test(name)), "unreviewed import element");
    if (!names.includes("propertyOf")) {
        const after = named.filter(element => element.getText(project).toLowerCase() < "propertyof");
        assert.equal(names.slice(0, after.length).join(), after.map(element => element.getText(project)).join(), "import list is not sorted where propertyOf goes");
        const previous = after.at(-1);
        assert.equal(previous.getText(project), "ProjectReference", "unreviewed import neighbor");
        assert.equal(project.text[previous.end], ",", "unreviewed import punctuation");
        const indent = project.text.slice(project.text.lastIndexOf("\n", previous.getStart(project)) + 1, previous.getStart(project));
        edit(project, previous.end + 1, previous.end + 1, `\n${indent}propertyOf,`, false, 4);
    }
    for (const name of readers) {
        const method = one(nodes(ts, project, node => ts.isMethodDeclaration(node) && node.name.getText(project) === name), name);
        const declaration = one(nodes(ts, method, node => ts.isVariableDeclaration(node) && node.name.getText(project) === "err"), `${name}'s err`);
        const text = declaration.initializer.getText(project);
        assert([readerBefore, readerAfter].includes(text), `unreviewed ${name} reader: ${text}`);
        if (text === readerBefore) edit(project, declaration.initializer.getStart(project), declaration.initializer.end, readerAfter, false, 4);
    }
    const otherReads = nodes(ts, project, node => ts.isPropertyAccessExpression(node) && ts.isPropertyAccessExpression(node.expression) && node.expression.name.text === "error" && ["stack", "message"].includes(node.name.text));
    assert.equal(otherReads.length, edits.get(project).filter(change => change.text === readerAfter).length * 2, "another reader of a plugin error's stack or message");

    // Row 5: System.createSHA256Hash, method to property in adaptation 32's idiom.
    const system = one(sys.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === "System"), "System");
    assert(!internal(ts, system), "System became @internal");
    const member = one(system.members.filter(node => node.name?.getText(sys) === "createSHA256Hash"), "System.createSHA256Hash");
    assert(!internal(ts, member), "System.createSHA256Hash became @internal");
    const memberText = member.getText(sys);
    assert([hashBefore, hashAfter].includes(memberText), `unreviewed createSHA256Hash: ${memberText}`);
    if (memberText === hashBefore) edit(sys, member.getStart(sys), member.end, hashAfter, true, 5);
    one(nodes(ts, sys, node => ts.isPropertyAssignment(node) && node.getText(sys) === "createSHA256Hash: _crypto ? createSHA256Hash : undefined"), "the nodeSystem store");

    // Plan every file, then prove the erasing edits erase and the rest change only what's labeled.
    const results = [];
    for (const [file, changes] of edits) {
        const after = apply(file.text, changes);
        assert.equal(ts.createSourceFile(file.fileName, after, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0, "adapted syntax error");
        const changedOnly = apply(file.text, changes.filter(change => !erasing.get(file).includes(change)));
        for (const removeComments of [false, true]) {
            assert.equal(transpile(ts, after, removeComments), transpile(ts, changedOnly, removeComments), `an erasing edit changed JavaScript in ${path.basename(file.fileName)}`);
        }
        results.push({ fileName: file.fileName, before: file.text, after });
    }
    // The barrel now exports 47's errorCode and errorMessage too. Every read 98-e adds goes through
    // propertyOf: no inserted or replaced text names either, project.ts names neither after, and
    // each edited file outside sys.ts keeps its count of references to each. Files 98-e does not
    // edit are unchanged, so their counts are too.
    const barrelOnly = /\b(errorCode|errorMessage)\b/;
    for (const changes of edits.values()) {
        for (const change of changes) assert(!barrelOnly.test(change.text), `98-e's edit names ${change.text.match(barrelOnly)?.[0]}: new reads go through propertyOf`);
    }
    const references = (fileName, text) => {
        const counts = { errorCode: 0, errorMessage: 0 };
        (function visit(node) {
            if (ts.isIdentifier(node) && Object.hasOwn(counts, node.text)) counts[node.text]++;
            ts.forEachChild(node, visit);
        })(ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true));
        return counts;
    };
    for (const result of results) {
        if (result.fileName === sys.fileName) continue;
        assert.deepEqual(references(result.fileName, result.after), references(result.fileName, result.before), `${path.relative(tree, result.fileName)} gains a reference to errorCode or errorMessage`);
    }
    const projectAfter = one(results.filter(result => result.fileName === project.fileName), "planned project.ts");
    assert.deepEqual(references(projectAfter.fileName, projectAfter.after), { errorCode: 0, errorMessage: 0 }, "project.ts names errorCode or errorMessage");
    return { results, evidence };
}

function adapt(tree, ts) {
    const { results, evidence } = plan(tree, ts);
    for (const result of results) assert.equal(fs.readFileSync(result.fileName, "utf8"), result.before, "source changed while planning");
    let files = 0;
    for (const result of results) if (result.after !== result.before) { fs.writeFileSync(result.fileName, result.after); files++; }
    return { files, javascriptChanging: true, evidence };
}

module.exports = { adapt, plan, hashBefore, hashAfter, readerBefore, readerAfter, pluginError };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
