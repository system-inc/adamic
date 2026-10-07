#!/usr/bin/env node
'use strict';
// Only the audited internal evaluator result slot is owned by this wave.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const options = {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    noImplicitReturns: true, noFallthroughCasesInSwitch: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: true,
    allowImportingTsExtensions: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
};
const roots = ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']);
const program = ts.createProgram(roots, options);
const checker = program.getTypeChecker();
const selected = [
    { file: 'src/compiler/types.ts', declaration: 'EvaluatorResult', member: 'value' },
];
function location(node) {
    const file = node.getSourceFile();
    const point = file.getLineAndCharacterOfPosition(node.getStart(file));
    return `${path.relative(tree, file.fileName)}:${point.line + 1}:${point.character + 1}`;
}
function ownerOf(symbol, owner) { return symbol?.declarations?.includes(owner); }
const owners = selected.map(spec => {
    const file = program.getSourceFile(path.join(tree, spec.file));
    if (!file) throw new Error(`missing source: ${spec.file}`);
    const declarations = file.statements.filter(node =>
        (ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node)) && node.name.text === spec.declaration);
    if (declarations.length !== 1) throw new Error(`expected one declaration: ${spec.declaration}`);
    const declaration = declarations[0];
    // An exported declaration must be explicitly internal. An unexported alias
    // stays local; this does not infer public API permission from optionality.
    if (ts.getCombinedModifierFlags(declaration) & ts.ModifierFlags.Export &&
        !ts.getJSDocTags(declaration).some(tag => tag.tagName.text === 'internal')) {
        throw new Error(`public declaration excluded: ${spec.declaration}`);
    }
    const literals = ts.isInterfaceDeclaration(declaration) ? [declaration] :
        (ts.isUnionTypeNode(declaration.type) ? declaration.type.types : [declaration.type]).filter(ts.isTypeLiteralNode);
    const members = literals.flatMap(node => [...node.members]).filter(node =>
        ts.isPropertySignature(node) && ts.isIdentifier(node.name) && node.name.text === spec.member);
    if (members.length !== 1 || !members[0].type) throw new Error(`unexpected slot: ${spec.declaration}.${spec.member}`);
    return { ...spec, node: members[0], references: 0, writes: [], escapes: [] };
});
function property(type, name) { return checker.getPropertyOfType(checker.getNonNullableType(type), name); }
function assignmentTarget(node) {
    let current = node;
    while (current.parent) {
        const parent = current.parent;
        if (ts.isParenthesizedExpression(parent) || ts.isArrayLiteralExpression(parent) ||
            ts.isObjectLiteralExpression(parent) || ts.isPropertyAssignment(parent) ||
            ts.isSpreadAssignment(parent) || ts.isSpreadElement(parent)) { current = parent; continue; }
        if (ts.isBinaryExpression(parent) && parent.left === current &&
            parent.operatorToken.kind >= ts.SyntaxKind.FirstAssignment &&
            parent.operatorToken.kind <= ts.SyntaxKind.LastAssignment) return parent;
        if ((ts.isPrefixUnaryExpression(parent) || ts.isPostfixUnaryExpression(parent)) &&
            (parent.operator === ts.SyntaxKind.PlusPlusToken || parent.operator === ts.SyntaxKind.MinusMinusToken)) return parent;
        if (ts.isDeleteExpression(parent)) return parent;
        return undefined;
    }
}
for (const file of program.getSourceFiles()) {
    if (!path.resolve(file.fileName).startsWith(tree + path.sep + 'src' + path.sep)) continue;
    function visit(node) {
        if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) {
            const receiver = checker.getTypeAtLocation(node.expression);
            const name = ts.isPropertyAccessExpression(node) ? node.name.text :
                ts.isStringLiteralLike(node.argumentExpression) ? node.argumentExpression.text : undefined;
            const write = assignmentTarget(node);
            for (const owner of owners) {
                if (name === owner.member && ownerOf(property(receiver, name), owner.node)) {
                    owner.references++;
                    if (write) owner.writes.push(location(write));
                } else if (name === undefined && write && ownerOf(property(receiver, owner.member), owner.node)) {
                    // An unresolved computed write could replace the selected slot.
                    owner.writes.push(location(write));
                }
            }
        }
        if (ts.isExpressionNode(node) && node.parent?.name !== node) {
            const source = checker.getTypeAtLocation(node);
            for (const owner of owners) {
                if (!ownerOf(property(source, owner.member), owner.node)) continue;
                let target = checker.getContextualType(node);
                if (ts.isAsExpression(node.parent) || ts.isTypeAssertionExpression(node.parent)) {
                    target = checker.getTypeFromTypeNode(node.parent.type);
                }
                if (target) {
                    const slot = property(target, owner.member);
                    if (!ownerOf(slot, owner.node)) {
                        owner.escapes.push({ where: location(node), target: checker.typeToString(target) });
                    }
                }
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(file);
}
// Fail before any write if upstream has gained a writer, an erased escape, or
// another structural view. Even a different readonly declaration is declined
// because forwarding from it could erase readonly. Symbol identity, not spelling,
// identifies the audited slot through calls, inferred aliases and instantiations.
for (const owner of owners) {
    if (owner.writes.length || owner.escapes.length) {
        throw new Error(`cannot prove ${owner.declaration}.${owner.member} readonly: ${JSON.stringify({ writes: owner.writes, escapes: owner.escapes })}`);
    }
}
const edits = new Map();
for (const owner of owners) {
    if (ts.getCombinedModifierFlags(owner.node) & ts.ModifierFlags.Readonly) continue;
    const file = owner.node.getSourceFile();
    if (!edits.has(file)) edits.set(file, []);
    edits.get(file).push(owner.node.getStart(file));
}
for (const [file, positions] of edits) {
    let text = file.text;
    for (const position of positions.sort((a, b) => b - a)) text = text.slice(0, position) + 'readonly ' + text.slice(position);
    if (fs.readFileSync(file.fileName, 'utf8') !== file.text) throw new Error(`source changed during audit: ${file.fileName}`);
    fs.writeFileSync(file.fileName, text);
}
console.log(JSON.stringify({ typescript: ts.version, files: edits.size,
    owners: owners.map(({ node, ...owner }) => ({ ...owner, where: location(node) })) }, null, 2));
