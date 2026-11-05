#!/usr/bin/env node
"use strict";
// debug.ts's two writes to its own namespace: shouldAssertFunction caches an assertion and writes
// `(Debug as any)[name] = noop` below its level, and setAssertionLevel writes the cache entries back when
// the level rises. Adamic refuses both, the namespace used as a value and the write through any, and ruled
// adapt, don't lower (Nov 5, #0kpp2zc, #b1y988t): shouldAssertFunction checks the level on each call and
// patches nothing, setAssertionLevel only stores the level, and the cache and its two imports go. The
// restore loop also writes back the cache entry, not its assertion (LEDGER.md entry 8), which this removes.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { compiler, nodes } = require("../47-host-errors/adapt.cjs");

const crlfLines = lines => lines.join("\r\n");

// What the edit removes and writes, each as upstream (with 30's `| undefined`) and as 99 writes it.
const cacheText = "const assertionCache: Partial<Record<AssertionKeys, { level: AssertionLevel; assertion: AnyFunction; } | undefined>> = {};";
const setHead = "export function setAssertionLevel(level: AssertionLevel): void";
const setBody = {
    original: crlfLines([
        "{",
        "        const prevAssertionLevel = currentAssertionLevel;",
        "        currentAssertionLevel = level;",
        "",
        "        if (level > prevAssertionLevel) {",
        "            // restore assertion functions for the current assertion level (see `shouldAssertFunction`).",
        "            for (const key of getOwnKeys(assertionCache) as AssertionKeys[]) {",
        "                const cachedFunc = assertionCache[key];",
        "                if (cachedFunc !== undefined && Debug[key] !== cachedFunc.assertion && level >= cachedFunc.level) {",
        "                    (Debug as any)[key] = cachedFunc;",
        "                    assertionCache[key] = undefined;",
        "                }",
        "            }",
        "        }",
        "    }",
    ]),
    adapted: crlfLines([
        "{",
        "        currentAssertionLevel = level;",
        "    }",
    ]),
};
const shouldHead = "function shouldAssertFunction<K extends AssertionKeys>(level: AssertionLevel, name: K): boolean";
const shouldBody = {
    original: crlfLines([
        "{",
        "        if (!shouldAssert(level)) {",
        "            assertionCache[name] = { level, assertion: Debug[name] };",
        "            (Debug as any)[name] = noop;",
        "            return false;",
        "        }",
        "        return true;",
        "    }",
    ]),
    adapted: crlfLines([
        "{",
        "        return shouldAssert(level);",
        "    }",
    ]),
};
const shouldDoc = {
    original: crlfLines([
        "/**",
        "     * Tests whether an assertion function should be executed. If it shouldn't, it is cached and replaced with `ts.noop`.",
        "     * Replaced assertion functions are restored when `Debug.setAssertionLevel` is set to a high enough level.",
        "     * @param level The minimum assertion level required.",
        "     * @param name The name of the current assertion function.",
        "     */",
    ]),
    adapted: crlfLines([
        "/**",
        "     * Tests whether an assertion function should be executed at the current assertion level.",
        "     * The level is read on every call, so `Debug.setAssertionLevel` applies to every later assertion.",
        "     * @param level The minimum assertion level required.",
        "     * @param name The name of the current assertion function.",
        "     */",
    ]),
};
// shouldAssertFunction now reads the level through shouldAssert, which must still be this comparison.
const shouldAssertText = crlfLines([
    "export function shouldAssert(level: AssertionLevel): boolean {",
    "        return currentAssertionLevel >= level;",
    "    }",
]);
// The two imports that only the removed lines name, and the assertion functions that call the helper.
const imported = ["getOwnKeys", "noop"];
const assertions = ["assertEachNode", "assertNode", "assertNotNode", "assertOptionalNode", "assertOptionalToken", "assertMissingNode"];
// Names that exist only in the removed lines; the adapted file names none of them.
const removedNames = ["assertionCache", "prevAssertionLevel", "cachedFunc", ...imported];

function one(found, label) {
    assert.equal(found.length, 1, `${label}: want one, got ${found.length}`);
    return found[0];
}

function transpile(ts, text, removeComments) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
}

// debug.ts is CRLF throughout; every line break before and after is one.
function crlf(text, label) {
    assert.equal(text.split("\r\n").length, text.split("\n").length, `${label} has a line break that is not CRLF`);
}

// A body or comment is upstream's or 99's text exactly. Anything else is a tree that moved.
function textState(text, states, label) {
    const found = Object.keys(states).filter(state => states[state] === text);
    assert.equal(found.length, 1, `uninventoried ${label}: ${text}`);
    return found[0];
}

function debugNamespace(ts, file) {
    return one(file.statements.filter(node => ts.isModuleDeclaration(node) && node.name.text === "Debug"), "namespace Debug");
}

function member(ts, file, owner, name) {
    return one(owner.body.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body), `Debug.${name}`);
}

// The function's head, its body's state, and its one leading comment (the JSDoc, or none).
function functionSite(ts, file, owner, name, head, bodies) {
    const declaration = member(ts, file, owner, name);
    assert.equal(file.text.slice(declaration.getStart(file), declaration.body.getStart(file)).trimEnd(), head, `${name}: unreviewed signature`);
    return { declaration, state: textState(declaration.body.getText(file), bodies, `${name} body`) };
}

function docState(ts, file, declaration) {
    const ranges = ts.getLeadingCommentRanges(file.text, declaration.getFullStart()) ?? [];
    const range = one(ranges, "comments above shouldAssertFunction");
    const text = file.text.slice(range.pos, range.end);
    return { range, state: textState(text, shouldDoc, "shouldAssertFunction doc comment") };
}

// The cache declaration, present exactly as upstream leaves it after 30, or absent.
function cacheState(ts, file, owner) {
    const statements = owner.body.statements.filter(node => ts.isVariableStatement(node)
        && node.declarationList.declarations.some(declaration => ts.isIdentifier(declaration.name) && declaration.name.text === "assertionCache"));
    if (!statements.length) return { state: "adapted" };
    const statement = one(statements, "assertionCache declaration");
    assert.equal(statement.getText(file), cacheText, "unreviewed assertionCache declaration");
    assert.equal(file.text.slice(statement.getFullStart(), statement.getStart(file)), "\r\n\r\n    ", "unreviewed text above the assertionCache declaration");
    return { state: "original", statement };
}

// noop and getOwnKeys in the one named import from ./_namespaces/ts.js: both present or both gone.
function importState(ts, file) {
    const declaration = one(file.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === "./_namespaces/ts.js"
        && node.importClause?.namedBindings && ts.isNamedImports(node.importClause.namedBindings)), "./_namespaces/ts.js named import");
    const specifiers = imported.map(name => declaration.importClause.namedBindings.elements.filter(element => element.name.text === name));
    const states = specifiers.map(found => found.length ? "original" : "adapted");
    assert(states.every(state => state === states[0]), `half-adapted import: ${imported.map((name, index) => `${name} ${states[index]}`).join(", ")}`);
    if (states[0] === "adapted") return { state: "adapted", specifiers: [] };
    const found = specifiers.map((entry, index) => {
        const specifier = one(entry, `${imported[index]} import`);
        assert(!specifier.isTypeOnly && !specifier.propertyName, `unreviewed ${imported[index]} import`);
        assert.equal(file.text.slice(specifier.getFullStart(), specifier.getStart(file)), "\r\n    ", `unreviewed text before the ${imported[index]} import`);
        assert.equal(file.text[specifier.end], ",", `the ${imported[index]} import is not followed by a comma`);
        return specifier;
    });
    return { state: "original", specifiers: found };
}

// Each assertion function calls the helper once, with Normal and its own name, as both states do.
function callers(ts, file, owner) {
    const calls = nodes(ts, owner, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === "shouldAssertFunction");
    const names = calls.map(call => {
        let enclosing = call.parent;
        while (enclosing && !ts.isFunctionDeclaration(enclosing)) enclosing = enclosing.parent;
        assert(enclosing?.name, "shouldAssertFunction called outside a function");
        assert.equal(call.arguments.length, 2, `${enclosing.name.text}: unreviewed shouldAssertFunction call`);
        assert.equal(call.arguments[0].getText(file), "AssertionLevel.Normal", `${enclosing.name.text}: unreviewed assertion level`);
        assert(ts.isStringLiteral(call.arguments[1]) && call.arguments[1].text === enclosing.name.text, `${enclosing.name.text}: the assertion key is not its own name`);
        return enclosing.name.text;
    });
    assert.deepEqual(names.slice().sort(), assertions.slice().sort(), "the assertion functions that call shouldAssertFunction changed");
}

// The whole file's state: every piece upstream's, or every piece 99's. Anything else is half done.
function fileState(ts, file) {
    const owner = debugNamespace(ts, file);
    const shouldAssert = member(ts, file, owner, "shouldAssert");
    assert.equal(shouldAssert.getText(file), shouldAssertText, "unreviewed shouldAssert");
    callers(ts, file, owner);
    const set = functionSite(ts, file, owner, "setAssertionLevel", setHead, setBody);
    const should = functionSite(ts, file, owner, "shouldAssertFunction", shouldHead, shouldBody);
    const doc = docState(ts, file, should.declaration);
    const cache = cacheState(ts, file, owner);
    const imports = importState(ts, file);
    const pieces = { setAssertionLevel: set.state, shouldAssertFunction: should.state, "doc comment": doc.state, assertionCache: cache.state, imports: imports.state };
    const states = Object.values(pieces);
    assert(states.every(state => state === states[0]), `half-adapted debug.ts: ${Object.entries(pieces).map(([name, state]) => `${name} ${state}`).join(", ")}`);
    // The writes through any are the removed lines' two; the removed names appear nowhere after.
    const writes = nodes(ts, owner, node => ts.isAsExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === "Debug");
    assert.equal(writes.length, states[0] === "original" ? 2 : 0, "Debug cast outside a reviewed site");
    if (states[0] === "adapted") {
        const named = nodes(ts, file, node => ts.isIdentifier(node) && removedNames.includes(node.text)).map(node => node.text);
        assert.deepEqual(named, [], "a removed name is still named in debug.ts");
    }
    return { state: states[0], set, should, doc, cache, imports };
}

// Before the edit, by symbol: the cache, noop and getOwnKeys are read only inside the two bodies being replaced,
// so nothing else in debug.ts loses them.
function onlyInRemoved(ts, file, checker, found) {
    const removed = [found.set.declaration.body, found.should.declaration.body];
    const inside = node => removed.some(body => node.pos >= body.pos && node.end <= body.end);
    const declared = [
        { name: "assertionCache", at: found.cache.statement.declarationList.declarations[0].name },
        ...found.imports.specifiers.map(specifier => ({ name: specifier.name.text, at: specifier.name })),
    ];
    for (const entry of declared) {
        const symbol = checker.getSymbolAtLocation(entry.at);
        assert(symbol, `${entry.name} has no symbol`);
        const uses = nodes(ts, file, node => ts.isIdentifier(node) && node !== entry.at && checker.getSymbolAtLocation(node) === symbol);
        assert(uses.length > 0, `${entry.name} is never read`);
        const outside = uses.filter(node => !inside(node));
        assert.equal(outside.length, 0, `${entry.name} is read outside the replaced bodies, at line ${outside.map(node => file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1).join(", ")}`);
    }
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

// Stock 6.0.3's JavaScript of each replaced piece, before and after, with comments kept or removed.
function javascriptPieces(removeComments) {
    const lines = list => list.join("\n") + "\n";
    const restore = removeComments ? [] : ["            // restore assertion functions for the current assertion level (see `shouldAssertFunction`)."];
    const doc = text => removeComments ? [] : text.split("\r\n").map((line, index) => index ? line : `    ${line}`);
    return [
        [lines(["    const assertionCache = {};"]), ""],
        [lines([
            "    function setAssertionLevel(level) {",
            "        const prevAssertionLevel = currentAssertionLevel;",
            "        currentAssertionLevel = level;",
            "        if (level > prevAssertionLevel) {",
            ...restore,
            "            for (const key of getOwnKeys(assertionCache)) {",
            "                const cachedFunc = assertionCache[key];",
            "                if (cachedFunc !== undefined && Debug[key] !== cachedFunc.assertion && level >= cachedFunc.level) {",
            "                    Debug[key] = cachedFunc;",
            "                    assertionCache[key] = undefined;",
            "                }",
            "            }",
            "        }",
            "    }",
        ]), lines([
            "    function setAssertionLevel(level) {",
            "        currentAssertionLevel = level;",
            "    }",
        ])],
        [lines([
            ...doc(shouldDoc.original),
            "    function shouldAssertFunction(level, name) {",
            "        if (!shouldAssert(level)) {",
            "            assertionCache[name] = { level, assertion: Debug[name] };",
            "            Debug[name] = noop;",
            "            return false;",
            "        }",
            "        return true;",
            "    }",
        ]), lines([
            ...doc(shouldDoc.adapted),
            "    function shouldAssertFunction(level, name) {",
            "        return shouldAssert(level);",
            "    }",
        ])],
    ];
}

// Stock 6.0.3's JavaScript of the planned text is upstream's with exactly the three pieces replaced and the two
// names dropped from the import. The expected text is built from the before output, never from source.
function provesJavaScript(ts, before, after) {
    for (const removeComments of [false, true]) {
        const original = transpile(ts, before, removeComments);
        let expected = original;
        for (const [from, to] of javascriptPieces(removeComments)) {
            assert.equal(expected.split(from).length - 1, 1, `upstream JavaScript lacks ${from.split("\n")[0].trim()}`);
            expected = expected.replace(from, to);
        }
        const importLine = one(expected.split("\n").filter(line => line.startsWith("import {") && line.endsWith(" from \"./_namespaces/ts.js\";")), "JavaScript named import");
        let trimmed = importLine;
        for (const name of imported) {
            assert.equal(trimmed.split(` ${name}, `).length - 1, 1, `upstream JavaScript does not import ${name}`);
            trimmed = trimmed.replace(` ${name}, `, " ");
        }
        expected = expected.replace(importLine, trimmed);
        assert.equal(transpile(ts, after, removeComments), expected, "the adapted JavaScript differs by more than the three pieces and the import");
    }
}

// No declaration changes. With stripInternal, as typescript.d.ts is built, both declare the same nothing,
// since Debug is internal. Without it, the declarations are identical too: the cache and shouldAssertFunction
// are not exported, setAssertionLevel keeps its signature, and no declaration names noop or getOwnKeys.
function provesDeclarations(ts, before, after) {
    const declare = (text, stripInternal) => ts.transpileDeclaration(text, { compilerOptions: { stripInternal }, fileName: "debug.ts" }).outputText;
    assert.equal(declare(after, true), declare(before, true), "the public declarations changed");
    assert.equal(declare(after, false), declare(before, false), "the internal declarations changed");
}

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const fileName = path.join(tree, "src/compiler/debug.ts");
    const text = fs.readFileSync(fileName, "utf8");
    crlf(text, "debug.ts");
    const parsed = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(parsed.parseDiagnostics.length, 0, "cannot parse debug.ts");
    if (fileState(ts, parsed).state === "adapted") return { fileName, before: text, after: text, evidence: [] };
    // Only an upstream file needs the checker, to prove the removed names are read nowhere else.
    const program = compiler(tree, ts);
    // The checker binds the program's files, which sets the parent pointers fileState walks.
    const checker = program.getTypeChecker();
    const file = program.getSourceFile(fileName);
    assert(file && file.text === text, "the program's debug.ts is not the file on disk");
    const found = fileState(ts, file);
    onlyInRemoved(ts, file, checker, found);
    const where = node => file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1;
    const edits = [];
    const evidence = [];
    edits.push({ start: found.cache.statement.getFullStart(), end: found.cache.statement.end, text: "" });
    evidence.push({ site: "assertionCache", file: "src/compiler/debug.ts", line: where(found.cache.statement), before: cacheText, after: "" });
    for (const [name, site, bodies] of [["setAssertionLevel", found.set, setBody], ["shouldAssertFunction", found.should, shouldBody]]) {
        edits.push({ start: site.declaration.body.getStart(file), end: site.declaration.body.end, text: bodies.adapted });
        evidence.push({ site: name, file: "src/compiler/debug.ts", line: where(site.declaration), before: bodies.original, after: bodies.adapted });
    }
    edits.push({ start: found.doc.range.pos, end: found.doc.range.end, text: shouldDoc.adapted });
    evidence.push({ site: "shouldAssertFunction doc comment", file: "src/compiler/debug.ts", line: file.getLineAndCharacterOfPosition(found.doc.range.pos).line + 1, before: shouldDoc.original, after: shouldDoc.adapted });
    for (const specifier of found.imports.specifiers) {
        edits.push({ start: specifier.getFullStart(), end: specifier.end + 1, text: "" });
        evidence.push({ site: `${specifier.name.text} import`, file: "src/compiler/debug.ts", line: where(specifier), before: `${specifier.name.text},`, after: "" });
    }
    const after = apply(text, edits);
    // Re-read the planned text: it must parse, stay CRLF, and be exactly 99's whole state.
    const replanned = ts.createSourceFile(fileName, after, ts.ScriptTarget.Latest, true);
    assert.equal(replanned.parseDiagnostics.length, 0, "adapted syntax error");
    crlf(after, "adapted debug.ts");
    assert.equal(fileState(ts, replanned).state, "adapted", "the planned text is not 99's");
    provesJavaScript(ts, text, after);
    provesDeclarations(ts, text, after);
    return { fileName, before: text, after, evidence };
}

function adapt(tree, ts) {
    const change = plan(tree, ts);
    assert.equal(fs.readFileSync(change.fileName, "utf8"), change.before, "source changed while planning");
    if (change.after !== change.before) fs.writeFileSync(change.fileName, change.after);
    return { files: change.after === change.before ? 0 : 1, javascriptChanging: true, evidence: change.evidence };
}

module.exports = { adapt, plan, shouldBody };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
