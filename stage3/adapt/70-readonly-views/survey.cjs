#!/usr/bin/env node
'use strict';
// An owner inventory, not an interprocedural mutation proof for unselected sites.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 5) throw new Error('usage: node survey.cjs <tree> <latent.jsonl> <output.json>');
const tree = path.resolve(process.argv[2]);
const raw = fs.readFileSync(process.argv[3], 'utf8').trim().split('\n').map(line => JSON.parse(line));
if (!raw[0].checker_rejected) throw new Error('expected checker-rejected measurement');
const sites = [...new Map(raw.slice(1).flatMap(row => row.findings)
    .filter(row => row.reason.includes(' seen as '))
    .map(row => [JSON.stringify([row.kind, row.where, row.reason, row.text]), row])).values()];
const roots = ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']);
const program = ts.createProgram(roots, {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: false, noEmit: true,
    allowImportingTsExtensions: true, module: ts.ModuleKind.ESNext,
    moduleDetection: ts.ModuleDetectionKind.Force, moduleResolution: ts.ModuleResolutionKind.Bundler,
    target: ts.ScriptTarget.ES2024, lib: ['lib.es2024.d.ts'], types: [],
});
const checker = program.getTypeChecker();
function location(node) {
    const file = node.getSourceFile();
    const p = file.getLineAndCharacterOfPosition(node.getStart(file));
    return `${path.relative(tree, file.fileName)}:${p.line + 1}:${p.character + 1}`;
}
function record(node) {
    let declaration = node;
    let internal = false;
    while (declaration.parent && !ts.isSourceFile(declaration.parent)) {
        internal ||= ts.getJSDocTags(declaration).some(tag => tag.tagName.text === 'internal');
        declaration = declaration.parent;
    }
    internal ||= ts.getJSDocTags(declaration).some(tag => tag.tagName.text === 'internal');
    const exported = !!(ts.getCombinedModifierFlags(declaration) & ts.ModifierFlags.Export);
    return { where: location(node), kind: ts.SyntaxKind[node.kind], text: node.getText().slice(0,500),
        api: exported && !internal ? 'public, no edit without ruling' : 'internal or local' };
}
function leaf(node, position) {
    let result = node;
    ts.forEachChild(node, child => {
        if (child.getFullStart() <= position && position < child.end) result = leaf(child, position);
    });
    return result;
}
function parts(type) { return type.isUnionOrIntersection() ? type.types.flatMap(parts) : [type]; }
const result = [];
for (const site of sites) {
    const point = site.where.slice(tree.length + 1).split(':');
    const source = program.getSourceFile(path.join(tree, point[0]));
    const row = { where: `${point[0]}:${point[1]}:${point[2]}`, reason: site.reason, contexts: [] };
    if (!source) { row.decline = 'source outside compiler tree'; result.push(row); continue; }
    const offset = source.getPositionOfLineAndCharacter(+point[1]-1,+point[2]-1);
    const node = leaf(source, offset);
    for (let n = node; n && !ts.isSourceFile(n); n = n.parent) {
        if (!ts.isExpressionNode(n)) continue;
        let target = checker.getContextualType(n);
        let owner;
        const parent = n.parent;
        if (ts.isCallExpression(parent)) {
            const signature = checker.getResolvedSignature(parent);
            const index = parent.arguments.indexOf(n);
            owner = signature?.parameters[Math.min(index,signature.parameters.length-1)]?.valueDeclaration;
        } else if (ts.isAsExpression(parent) || ts.isTypeAssertionExpression(parent)) {
            target = checker.getTypeFromTypeNode(parent.type);
            owner = parent.type;
        } else if (ts.isVariableDeclaration(parent) || ts.isPropertyAssignment(parent)) {
            owner = parent;
        } else if (ts.isBinaryExpression(parent) && parent.right === n && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
            target = checker.getTypeAtLocation(parent.left);
            owner = checker.getSymbolAtLocation(parent.left)?.valueDeclaration;
        } else if (ts.isReturnStatement(parent)) {
            let fn = parent.parent;
            while (fn && !ts.isFunctionLike(fn)) fn = fn.parent;
            const signature = fn && checker.getSignatureFromDeclaration(fn);
            if (signature) target ||= checker.getReturnTypeOfSignature(signature);
            owner = fn?.type || fn;
        }
        if (!target) continue;
        const declarations = parts(target).flatMap(type => (type.aliasSymbol || type.symbol)?.declarations || []);
        const fields = parts(target).flatMap(type => checker.getPropertiesOfType(type))
            .filter(symbol => symbol.declarations?.some(d => ts.isPropertySignature(d) || ts.isPropertyDeclaration(d)))
            .map(symbol => ({ name: symbol.name, declarations: symbol.declarations.map(record) }));
        row.contexts.push({ expression: n.getText().slice(0,180), target: checker.typeToString(target),
            owner: owner && record(owner), declarations: declarations.map(record), fields });
    }
    row.decline = row.contexts.length ? 'owner candidates inventoried; mutation and escape proof not completed for this site' :
        'no explicit contextual owner resolved; inferred or method-variance view requires a separate proof';
    result.push(row);
}
fs.writeFileSync(process.argv[4], JSON.stringify({ typescript: ts.version, measurement: raw[0].measurement,
    count: result.length, sites: result }, null, 2) + '\n');
console.log(`${result.length} sites inventoried with the checker`);
