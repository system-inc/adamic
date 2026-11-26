#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const pins = require("./premises.json");
const barrelStatements = require("./barrel.json");

const utilities = "utilities.ts";
const diagnostics = "transformers/declarations/diagnostics.ts";
const compilerNamespace = "_namespaces/ts.ts";
const slots = [
    {
        file: utilities,
        owner: "getAllAccessorDeclarations",
        before: "hasDynamicName(accessor)",
        after: "isDynamicName(accessor.name)",
        locate(owner) {
            const branch = one(owner.body.statements.filter(node =>
                ts.isIfStatement(node) && ts.isCallExpression(node.expression)
                && ["hasDynamicName", "isDynamicName"].includes(node.expression.expression.getText())),
            "accessor condition");
            return branch.expression;
        },
    },
    {
        file: diagnostics,
        owner: "findNearestDeclaration",
        before: "isStatement(result) ? undefined : result",
        after: "isVariableDeclaration(result) || isPropertyDeclaration(result) || isParameter(result) ? result : undefined",
        locate(owner) {
            const tail = owner.body.statements.at(-1);
            assert(tail && ts.isReturnStatement(tail) && tail.expression && ts.isAsExpression(tail.expression)
                && ts.isParenthesizedExpression(tail.expression.expression)
                && ts.isConditionalExpression(tail.expression.expression.expression), "nearest declaration cast");
            return tail.expression.expression.expression;
        },
    },
];

function one(matches, label) {
    assert.equal(matches.length, 1, label);
    return matches[0];
}

function collect(root, matches) {
    const found = [];
    function visit(node) {
        if (matches(node)) found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return found;
}

function named(source, name) {
    return collect(source, node => (ts.isFunctionDeclaration(node) || ts.isInterfaceDeclaration(node)
        || ts.isTypeAliasDeclaration(node) || ts.isVariableDeclaration(node) || ts.isEnumDeclaration(node)
        || ts.isModuleDeclaration(node))
        && node.name?.getText(source) === name);
}

function canonical(pin, source, sites) {
    const declarations = named(source, pin.name);
    assert(declarations.length, `missing premise ${pin.file} ${pin.name}`);
    return declarations.map(declaration => {
        const start = declaration.getStart(source);
        const end = declaration.end;
        // An unchanged branch needs the binding identity, not a new body premise.
        if (pin.headerOnly && declaration.body) {
            return source.text.slice(start, declaration.body.getStart(source)).trimEnd().replaceAll("\r\n", "\n");
        }
        let text = declaration.getText(source);
        const spans = sites.filter(site => site.source === source && site.node.getStart(source) >= start
            && site.node.end <= end).map(site => site.node);
        for (const span of spans.sort((a, b) => b.getStart(source) - a.getStart(source))) {
            text = text.slice(0, span.getStart(source) - start) + "__REVIEWED_SLOT__" + text.slice(span.end - start);
        }
        return text.replaceAll("\r\n", "\n");
    });
}

function lexicalBindings(source) {
    const bindings = new Map();
    const names = collect(source, node => {
        if (!ts.isIdentifier(node) || node.parent.name !== node) return false;
        const owner = node.parent;
        return ts.isVariableDeclaration(owner) || ts.isParameter(owner) || ts.isBindingElement(owner)
            || ts.isFunctionDeclaration(owner) || ts.isFunctionExpression(owner)
            || ts.isClassDeclaration(owner) || ts.isClassExpression(owner)
            || ts.isEnumDeclaration(owner) || ts.isModuleDeclaration(owner)
            || ts.isInterfaceDeclaration(owner) || ts.isTypeAliasDeclaration(owner)
            || ts.isImportSpecifier(owner) || ts.isImportClause(owner)
            || ts.isNamespaceImport(owner) || ts.isImportEqualsDeclaration(owner);
    });
    for (const name of names) {
        if (!bindings.has(name.text)) bindings.set(name.text, []);
        bindings.get(name.text).push(name.parent);
    }
    return bindings;
}

function modulePath(from, to) {
    const relative = path.posix.relative(path.posix.dirname(from), to).replace(/\.ts$/, ".js");
    return relative.startsWith(".") ? relative : "./" + relative;
}

function verifyBindings(files) {
    const barrel = files.get(compilerNamespace);
    assert(barrel.statements.length === barrelStatements.length
        && barrel.statements.every((node, index) => node.getText(barrel) === barrelStatements[index]),
        "premise drift in _namespaces/ts.ts ordered statements");
    const exports = barrel.statements.filter(ts.isExportDeclaration);
    const namesByFile = new Map([...files].map(([file, source]) => [file, lexicalBindings(source)]));
    for (const pin of pins) {
        for (const file of pin.usedIn || []) {
            const label = `premise drift in ${file} ${pin.name}`;
            const bindings = namesByFile.get(file).get(pin.name) || [];
            const declarations = named(files.get(pin.file), pin.name);
            assert(declarations.every(node => node.parent === files.get(pin.file)), label);
            if (file === pin.file) {
                assert(bindings.length === declarations.length && bindings.every(node => declarations.includes(node)), label);
                continue;
            }

            const binding = bindings[0];
            assert(bindings.length === 1 && ts.isImportSpecifier(binding) && !binding.propertyName
                && !binding.isTypeOnly && !binding.parent.parent.isTypeOnly, label);
            const statement = binding.parent.parent.parent;
            assert(statement.parent === files.get(file) && ts.isImportDeclaration(statement)
                && ts.isStringLiteral(statement.moduleSpecifier)
                && statement.moduleSpecifier.text === modulePath(file, compilerNamespace), label);

            const route = modulePath(compilerNamespace, pin.file);
            assert(exports.filter(node => !node.isTypeOnly && !node.exportClause && node.moduleSpecifier?.text === route).length === 1
                && !exports.some(node => node.exportClause && (ts.isNamedExports(node.exportClause)
                    ? node.exportClause.elements.some(element => element.name.text === pin.name)
                    : node.exportClause.name.text === pin.name))
                && !namesByFile.get(compilerNamespace).has(pin.name)
                && declarations.every(node => node.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword)), label);
        }
    }
    // The existing undefined value has no compiler declaration to import or shadow.
    for (const file of [diagnostics, "utilitiesPublic.ts"]) {
        assert(!namesByFile.get(file).has("undefined"), `premise drift in ${file} undefined`);
    }
}

function reviewed(tree) {
    const files = new Map();
    for (const file of new Set([compilerNamespace, ...pins.map(pin => pin.file)])) {
        const filename = path.join(tree, "src/compiler", file);
        const source = ts.createSourceFile(filename, fs.readFileSync(filename, "utf8"), ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
        files.set(file, source);
    }
    const sites = slots.map(slot => {
        const source = files.get(slot.file);
        const owner = one(named(source, slot.owner).filter(node => ts.isFunctionDeclaration(node) && node.body), slot.owner);
        const node = slot.locate(owner);
        const text = node.getText(source);
        assert(text === slot.before || text === slot.after, `unexpected slot in ${slot.owner}: ${text}`);
        return { slot, source, node, text };
    });
    for (const pin of pins) {
        assert(JSON.stringify(canonical(pin, files.get(pin.file), sites)) === JSON.stringify(pin.declarations),
            `premise drift in ${pin.file} ${pin.name}`);
    }
    verifyBindings(files);
    return sites;
}

if (ts.version !== "6.0.3") throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");

const sites = reviewed(path.resolve(process.argv[2]));
const edits = sites.map(({ slot, source, node, text }) => ({
    slot, source,
    changed: text === slot.before,
    adapted: source.text.slice(0, node.getStart(source)) + slot.after + source.text.slice(node.end),
}));
for (const { source } of edits) assert.equal(fs.readFileSync(source.fileName, "utf8"), source.text, "source changed while planning");
for (const { slot, source, changed, adapted } of edits) {
    if (changed) fs.writeFileSync(source.fileName, adapted);
    console.log(`${slot.owner}: ${changed ? "edited" : "already adapted"}`);
}
