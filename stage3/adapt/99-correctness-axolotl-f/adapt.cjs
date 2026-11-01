'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

const declaration = "/** @internal */\nexport interface HashedSet<T> {\n    has(element: T): boolean;\n    add(element: T): HashedSet<T>;\n    delete(element: T): boolean;\n    clear(): void;\n    readonly size: number;\n    forEach(action: (value: T, key: T, set: HashedSet<T>) => void): void;\n    keys(): SetIterator<T>;\n    values(): SetIterator<T>;\n    entries(): SetIterator<[T, T]>;\n    [Symbol.iterator](): SetIterator<T>;\n    [Symbol.toStringTag]: string;\n}\n\n";
const sites = [
    {
        "file": "src/compiler/core.ts",
        "scope": "createSet",
        "kind": "function",
        "old": "export function createSet<TElement, THash = number>(getHashCode: (element: TElement) => THash, equals: EqualityComparer<TElement>): Set<TElement> {",
        "new": "export function createSet<TElement, THash = number>(getHashCode: (element: TElement) => THash, equals: EqualityComparer<TElement>): HashedSet<TElement> {"
    },
    {
        "file": "src/compiler/core.ts",
        "scope": "createSet",
        "kind": "function",
        "old": "const set: Set<TElement> = {",
        "new": "const set: HashedSet<TElement> = {"
    },
    {
        "file": "src/compiler/core.ts",
        "scope": "createSet",
        "kind": "function",
        "old": "add(element: TElement): Set<TElement> {",
        "new": "add(element: TElement): HashedSet<TElement> {"
    },
    {
        "file": "src/compiler/core.ts",
        "scope": "createSet",
        "kind": "function",
        "old": "forEach(action: (value: TElement, key: TElement, set: Set<TElement>) => void): void {",
        "new": "forEach(action: (value: TElement, key: TElement, set: HashedSet<TElement>) => void): void {"
    },
    {
        "file": "src/server/session.ts",
        "scope": "./_namespaces/ts.js",
        "kind": "import",
        "old": "    DocumentSpan,",
        "new": "    DocumentSpan,\n    type HashedSet,"
    },
    {
        "file": "src/services/services.ts",
        "scope": "./_namespaces/ts.js",
        "kind": "import",
        "old": "    DocumentSpan,",
        "new": "    DocumentSpan,\n    type HashedSet,"
    },
    {
        "file": "src/services/types.ts",
        "scope": "./_namespaces/ts.js",
        "kind": "import",
        "old": "    HasChangedAutomaticTypeDirectiveNames,",
        "new": "    HasChangedAutomaticTypeDirectiveNames,\n    type HashedSet,"
    },
    {
        "file": "src/harness/client.ts",
        "scope": "./_namespaces/ts.js",
        "kind": "import",
        "old": "    DocumentSpan,",
        "new": "    DocumentSpan,\n    type HashedSet,"
    },
    {
        "file": "src/server/session.ts",
        "scope": "createDocumentSpanSet",
        "kind": "function",
        "old": "function createDocumentSpanSet(useCaseSensitiveFileNames: boolean): Set<DocumentSpan> {",
        "new": "function createDocumentSpanSet(useCaseSensitiveFileNames: boolean): HashedSet<DocumentSpan> {"
    },
    {
        "file": "src/services/services.ts",
        "scope": "updateIsDefinitionOfReferencedSymbols",
        "kind": "function",
        "old": "knownSymbolSpans: Set<DocumentSpan>",
        "new": "knownSymbolSpans: HashedSet<DocumentSpan>"
    },
    {
        "file": "src/services/types.ts",
        "scope": "updateIsDefinitionOfReferencedSymbols",
        "kind": "method",
        "old": "knownSymbolSpans: Set<DocumentSpan>",
        "new": "knownSymbolSpans: HashedSet<DocumentSpan>"
    },
    {
        "file": "src/harness/client.ts",
        "scope": "updateIsDefinitionOfReferencedSymbols",
        "kind": "method",
        "old": "_knownSymbolSpans: Set<DocumentSpan>",
        "new": "_knownSymbolSpans: HashedSet<DocumentSpan>"
    }
];
function nodes(root, predicate) {
    const found = [];
    (function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); })(root);
    return found;
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = [];
    for (const file of new Set(sites.map(site => site.file))) {
        const fileName = path.join(tree, file);
        const text = fs.readFileSync(fileName, 'utf8');
        const newline = text.includes('\r\n') ? '\r\n' : '\n';
        const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
        const edits = [];
        let adapted = 0;
        let expected = 0;
        for (const site of sites.filter(site => site.file === file)) {
            const scopes = nodes(source, node => site.kind === 'import'
                ? ts.isImportDeclaration(node) && node.moduleSpecifier.text === site.scope
                : (site.kind === 'method' ? ts.isMethodSignature(node) || ts.isMethodDeclaration(node) : ts.isFunctionDeclaration(node))
                    && node.name?.getText(source) === site.scope);
            assert.equal(scopes.length, 1, `${file}: ${site.scope} moved`);
            const scope = scopes[0];
            const start = scope.getStart(source);
            const body = text.slice(start, scope.end);
            const old = site.old.replaceAll('\n', newline);
            const replacement = site.new.replaceAll('\n', newline);
            const count = value => body.split(value).length - 1;
            const newCount = count(replacement), oldCount = body.split(replacement).join('').split(old).length - 1;
            assert(oldCount + newCount === 1, `${file}: ${site.scope}: exact text drift`);
            expected++;
            if (newCount) adapted++;
            else edits.push({at: start + body.indexOf(old), length: old.length, text: replacement});
        }
        if (file === 'src/compiler/core.ts') {
            const declarations = nodes(source, node => ts.isInterfaceDeclaration(node) && node.name.text === 'HashedSet');
            assert(declarations.length <= 1, 'duplicate HashedSet');
            expected++;
            if (declarations.length) {
                const node = declarations[0];
                assert.equal(text.slice(node.getFullStart(), node.end).trim(), declaration.replaceAll('\n', newline).trim(), 'HashedSet drift');
                adapted++;
            }
            else {
                const fn = nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === 'createSet')[0];
                edits.push({at: fn.getFullStart(), length: 0, text: newline + newline + declaration.replaceAll('\n', newline).trimEnd()});
            }
        }
        assert(adapted === 0 || adapted === expected, `${file} is partly adapted (${adapted} of ${expected})`);
        let result = text;
        for (const edit of edits.sort((a, b) => b.at - a.at)) result = result.slice(0, edit.at) + edit.text + result.slice(edit.at + edit.length);
        assert.equal(ts.createSourceFile(fileName, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0, `cannot parse adapted ${file}`);
        plans.push({fileName, text, result, edits: edits.length});
    }
    for (const plan of plans) assert.equal(fs.readFileSync(plan.fileName, 'utf8'), plan.text, 'source changed while planning');
    for (const plan of plans) if (plan.edits) fs.writeFileSync(plan.fileName, plan.result);
    return {files: plans.filter(plan => plan.edits).length, edits: plans.reduce((sum, plan) => sum + plan.edits, 0), javascriptChanging: false};
}
module.exports = {adapt, sites};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
