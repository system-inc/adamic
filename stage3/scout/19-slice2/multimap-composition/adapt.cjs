#!/usr/bin/env node
"use strict";
// Step 19 ruling 1, tracked for #pcwk8fh. No Set rewriting is authorized here.
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
if (ts.version !== "6.0.3") throw new Error("requires TypeScript 6.0.3");

function adapt(tree) {
    const directory = path.join(tree, "src/compiler");
    const files = fs.readdirSync(directory).filter(name => name.endsWith(".ts")).map(name => path.join(directory, name));
    const program = ts.createProgram(files, {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.NodeNext, moduleResolution: ts.ModuleResolutionKind.NodeNext});
    const checker = program.getTypeChecker();
    const core = program.getSourceFile(path.join(directory, "core.ts"));
    if (!core) throw new Error("missing core.ts");
    const declaration = core.statements.find(node => ts.isInterfaceDeclaration(node) && node.name.text === "MultiMap");
    if (!declaration) throw new Error("missing MultiMap");
    const port = fs.readFileSync(path.join(__dirname, "../../../../stage1/typescript/collections/multimap.a"), "utf8");
    const replacement = port.slice(port.indexOf("export interface MultiMap"));
    const replacementTree = ts.createSourceFile("multimap.a", replacement, ts.ScriptTarget.Latest, true);
    const normalize = text => text.replace(/\s+/g, " ").trim();
    if (!declaration.heritageClauses) {
        const expected = replacementTree.statements.slice(0, 5).map(node => normalize(node.getText(replacementTree)));
        const actual = core.statements.filter(node => node.name && ["MultiMap", "createMultiMap", "multiMapAdd", "multiMapRemove", "multiMapForEach"].includes(node.name.text)).map(node => normalize(node.getText(core)));
        if (JSON.stringify(expected) !== JSON.stringify(actual)) throw new Error("composition implementation drift");
        return [];
    }
    const remove = core.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "multiMapRemove");
    if (!remove) throw new Error("missing multiMapRemove");
    const original = fs.readFileSync(path.join(__dirname, "original.a"), "utf8");
    if (normalize(core.text.slice(declaration.getStart(core), remove.end)) !== normalize(original)) throw new Error("unreviewed MultiMap implementation");
    const isMultiMap = type => type.isUnion() ? type.types.some(isMultiMap) : ["MultiMap", "RedirectTargetsMap"].includes(type.aliasSymbol?.name || type.symbol?.name);
    const changes = [];
    for (const source of program.getSourceFiles()) {
        if (!files.includes(source.fileName)) continue;
        const edits = [], helpers = new Set();
        const record = (node, text) => edits.push({start: node.getStart(source), end: node.end, text});
        function visit(node) {
            if (source === core && node.getStart(source) >= declaration.getStart(core) && node.end <= remove.end) return;
            if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && ["add", "remove", "forEach"].includes(node.expression.name.text) && isMultiMap(checker.getTypeAtLocation(node.expression.expression))) {
                const helper = {add: "multiMapAdd", remove: "multiMapRemove", forEach: "multiMapForEach"}[node.expression.name.text];
                const receiver = node.expression.expression.getText(source);
                record(node, `${helper}(${receiver}${node.arguments.length ? ", " + node.arguments.map(argument => argument.getText(source)).join(", ") : ""})`);
                helpers.add(helper);
                return;
            }
            if (ts.isPropertyAccessExpression(node) && isMultiMap(checker.getTypeAtLocation(node.expression))) {
                if (["add", "remove"].includes(node.name.text)) throw new Error("unbound MultiMap helper at " + source.fileName);
                record(node, `${node.expression.getText(source)}.map.${node.name.text}`);
                return;
            }
            if (ts.isElementAccessExpression(node) && isMultiMap(checker.getTypeAtLocation(node.expression))) throw new Error("unreviewed computed MultiMap member");
            if (ts.isTypeAliasDeclaration(node) && node.name.text === "RedirectTargetsMap") {
                if (normalize(node.getText(source)) !== "export type RedirectTargetsMap = ReadonlyMap<Path, readonly string[]>;") throw new Error("RedirectTargetsMap drift");
                record(node, "export type RedirectTargetsMap = { readonly map: ReadonlyMap<Path, readonly string[]>; };");
                return;
            }
            ts.forEachChild(node, visit);
        }
        ts.forEachChild(source, visit);
        if (source === core) edits.push({start: declaration.getStart(core), end: remove.end, text: replacement.trimEnd()});
        if (helpers.size && source !== core) edits.push({start: 0, end: 0, text: `import { ${[...helpers].sort().join(", ")} } from "./_namespaces/ts.js";\n`});
        if (!edits.length) continue;
        edits.sort((a, b) => b.start - a.start);
        for (let index = 1; index < edits.length; index++) if (edits[index].end > edits[index - 1].start) throw new Error("overlapping MultiMap edits");
        let text = source.text;
        for (const edit of edits) text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
        if (ts.createSourceFile(source.fileName, text, ts.ScriptTarget.Latest, true).parseDiagnostics.length) throw new Error("adapted syntax error");
        changes.push({file: source.fileName, text, edits: edits.length});
    }
    // Validate every input before the first write, so source drift never leaves a half edit.
    for (const change of changes) fs.writeFileSync(change.file, change.text);
    return changes.map(change => ({file: path.relative(tree, change.file), edits: change.edits}));
}
module.exports = {adapt};
if (require.main === module) {
    if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
