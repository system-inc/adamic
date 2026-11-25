"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const pins = require("./pins.json");

const resolver = "src/compiler/moduleNameResolver.ts";
const services = "src/services/exportInfoMap.ts";
const name = "forEachAncestorDirectoryStoppingAtGlobalCache";
const normalize = text => text.replace(/\r\n/g, "\n");

function nodes(root, predicate) {
    const found = [];
    function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return found;
}

function one(found, label) {
    assert.equal(found.length, 1, label + " count drift");
    return found[0];
}

function locate(source, relative) {
    assert.equal(source.parseDiagnostics.length, 0, "Source parse error");
    if (relative === resolver) {
        const owner = one(nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === name), "Wrapper");
        assert(owner.parent === source && owner.body, "Wrapper scope or body drift");
        return { wrapper: owner };
    }
    const owner = one(nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === "getIsExcluded"), "Exclusion owner");
    assert(owner.parent === source && owner.body, "Exclusion owner scope or body drift");
    const imports = source.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === "./_namespaces/ts.js");
    const keys = nodes(owner, node => ts.isVariableStatement(node) &&
        node.declarationList.declarations.some(declaration => declaration.name.getText(source) === "symlinks"));
    return { imports: one(imports, "Services import"), key: one(keys, "Exclusion key") };
}

function start(node, source) {
    return node.getStart(source, ts.isFunctionDeclaration(node));
}

function matches(source, targets, side) {
    return Object.entries(targets).every(([label, node]) =>
        normalize(source.text.slice(start(node, source), node.end)) === pins[label][side]);
}

function rewrite(source, targets, side) {
    let result = source.text;
    const entries = Object.entries(targets).sort(([, a], [, b]) => b.pos - a.pos);
    for (const [label, node] of entries) {
        const replacement = source.text.includes("\r\n") ? pins[label][side].replace(/\n/g, "\r\n") : pins[label][side];
        result = result.slice(0, start(node, source)) + replacement + result.slice(node.end);
    }
    return result;
}

function verifyCalls(program, checker, tree, wrapper, side) {
    const symbol = checker.getSymbolAtLocation(wrapper.name);
    assert(symbol, "Ancestor wrapper symbol missing");
    const remaining = pins.calls.map(([relative, before, after]) =>
        JSON.stringify([relative, side === "after" ? after : before]));
    const resolvesToWrapper = node => {
        let resolved = checker.getSymbolAtLocation(node);
        if (resolved && resolved.flags & ts.SymbolFlags.Alias) resolved = checker.getAliasedSymbol(resolved);
        return resolved === symbol;
    };
    for (const source of program.getSourceFiles()) {
        if (!source.fileName.startsWith(path.join(tree, "src") + path.sep)) continue;
        const relative = path.relative(tree, source.fileName);
        function position(node) {
            const {line, character} = source.getLineAndCharacterOfPosition(node.getStart(source));
            return relative + ":" + (line + 1) + ":" + (character + 1);
        }
        function scan(node) {
            if (ts.isCallExpression(node) && resolvesToWrapper(node.expression)) {
                const pin = JSON.stringify([relative, normalize(node.getText(source))]);
                const index = remaining.indexOf(pin);
                assert(index !== -1, "Ancestor wrapper caller text drift at " + position(node));
                remaining.splice(index, 1);
            }
            if (ts.isIdentifier(node) && node !== wrapper.name &&
                !ts.isImportSpecifier(node.parent) && !ts.isExportSpecifier(node.parent) &&
                resolvesToWrapper(node)) {
                const parent = node.parent;
                if (!ts.isCallExpression(parent) || parent.expression !== node) {
                    assert.fail("Unpinned ancestor wrapper reference at " + position(node));
                }
            }
            ts.forEachChild(node, scan);
        }
        scan(source);
    }
    assert.equal(remaining.length, 0, "Ancestor wrapper caller count drift");
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const program = ts.createProgram(ts.sys.readDirectory(path.join(tree, "src"), [".ts"]), {
        module: ts.ModuleKind.NodeNext, strict: true, skipLibCheck: true,
    });
    const checker = program.getTypeChecker();
    const files = [resolver, services].map(relative => {
        const source = program.getSourceFile(path.join(tree, relative));
        assert(source, "Missing " + relative);
        return { relative, source, targets: locate(source, relative) };
    });
    const side = files.every(({source, targets}) => matches(source, targets, "after")) ? "after" : "before";
    for (const {source, targets} of files) {
        for (const [label, node] of Object.entries(targets)) {
            assert(matches(source, {[label]: node}, side), label + " exact text drift");
        }
    }
    verifyCalls(program, checker, tree, files[0].targets.wrapper, side);
    if (side === "after") return {files: 0, edits: 0};

    const writes = files.map(({relative, source, targets}) => {
        const after = rewrite(source, targets, "after");
        const parsed = ts.createSourceFile(source.fileName, after, ts.ScriptTarget.Latest, true);
        assert(matches(parsed, locate(parsed, relative), "after"), "Replacement drift");
        assert.equal(fs.readFileSync(source.fileName, "utf8"), source.text, "Source changed while planning");
        return [source.fileName, after];
    });
    for (const [file, after] of writes) fs.writeFileSync(file, after);
    return {files: 2, edits: 3};
}

module.exports = {adapt, locate, nodes, resolver, rewrite, services};
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
