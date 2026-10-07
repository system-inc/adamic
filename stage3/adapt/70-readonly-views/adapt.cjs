#!/usr/bin/env node
'use strict';
// Explicit owners are audited through the stock checker before any edit.
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
    { file: 'src/compiler/moduleNameResolver.ts', declaration: 'SearchResult', member: 'value' },
    { file: 'src/compiler/moduleNameResolver.ts', declaration: 'nodeModuleNameResolverWorker.tryResolve', returnMember: true, member: 'isExternalLibraryImport' },
    { file: 'src/compiler/utilities.ts', declaration: 'ParsedPatterns', member: 'matchableStringSet' },
    { file: 'src/compiler/utilities.ts', declaration: 'ParsedPatterns', member: 'patterns' },
    { file: 'src/compiler/moduleNameResolver.ts', declaration: 'Resolved', member: 'originalPath' },
    { file: 'src/compiler/moduleNameResolver.ts', declaration: 'ModuleResolutionState', member: 'reportDiagnostic' },
    { file: 'src/compiler/moduleSpecifiers.ts', declaration: 'ModuleSpecifierResult', member: 'kind' },
    { file: 'src/compiler/moduleSpecifiers.ts', declaration: 'ModuleSpecifierResult', member: 'moduleSpecifiers' },
    { file: 'src/compiler/moduleSpecifiers.ts', declaration: 'ModuleSpecifierResult', member: 'computedWithoutCache' },
];
const parameters = require('./parameter-audit.cjs')(program, checker, tree, [
    { file: 'src/compiler/utilities.ts', function: 'createDiagnosticForRange', parameter: 'range' },
    { file: 'src/compiler/utilities.ts', function: 'nodeIsSynthesized', parameter: 'range' },
    { file: 'src/compiler/utilities.ts', function: 'moveRangeEnd', parameter: 'range' },
    { file: 'src/compiler/utilities.ts', function: 'moveRangePos', parameter: 'range' },
    { file: 'src/compiler/utilities.ts', function: 'rangeIsOnSingleLine', parameter: 'range' },
    { file: 'src/compiler/utilities.ts', function: 'rangeStartIsOnSameLineAsRangeEnd', parameter: 'range1' },
    { file: 'src/compiler/utilities.ts', function: 'rangeStartIsOnSameLineAsRangeEnd', parameter: 'range2' },
    { file: 'src/compiler/utilities.ts', function: 'getStartPositionOfRange', parameter: 'range' },
    { file: 'src/compiler/utilities.ts', function: 'emitDetachedComments', parameter: 'node' },
    { file: 'src/compiler/utilities.ts', function: 'emitNewLineBeforeLeadingComments', parameter: 'node' },
    { file: 'src/compiler/parser.ts', function: 'parseErrorAtRange', parameter: 'range' },
    { file: 'src/compiler/factory/utilities.ts', function: 'createMemberAccessForPropertyName', parameter: 'location' },
    { file: 'src/compiler/factory/utilitiesPublic.ts', function: 'setTextRange', parameter: 'location', public: true },
]);
function location(node) {
    const file = node.getSourceFile();
    const point = file.getLineAndCharacterOfPosition(node.getStart(file));
    return `${path.relative(tree, file.fileName)}:${point.line + 1}:${point.character + 1}`;
}
function ownerOf(symbol, owner) { return symbol?.declarations?.includes(owner); }
const owners = selected.map(spec => {
    const file = program.getSourceFile(path.join(tree, spec.file));
    if (!file) throw new Error(`missing source: ${spec.file}`);
    const declarations = [];
    function find(node) {
        if (spec.returnMember && ts.isFunctionDeclaration(node)) {
            const names = [];
            for (let current = node; current; current = current.parent) {
                if (ts.isFunctionDeclaration(current) && current.name) names.unshift(current.name.text);
            }
            if (names.join('.') === spec.declaration) declarations.push(node);
        } else if (!spec.returnMember && (ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node)) &&
            node.name.text === spec.declaration) declarations.push(node);
        ts.forEachChild(node, find);
    }
    find(file);
    if (declarations.length !== 1) throw new Error(`expected one declaration: ${spec.declaration}`);
    const declaration = declarations[0];
    // An exported declaration must be explicitly internal. An unexported alias
    // stays local; this does not infer public API permission from optionality.
    if (ts.getCombinedModifierFlags(declaration) & ts.ModifierFlags.Export &&
        !ts.getJSDocTags(declaration).some(tag => tag.tagName.text === 'internal')) {
        throw new Error(`public declaration excluded: ${spec.declaration}`);
    }
    const literals = spec.returnMember ?
        (ts.isTypeReferenceNode(declaration.type) && declaration.type.typeName.getText(file) === 'SearchResult' ?
            declaration.type.typeArguments?.filter(ts.isTypeLiteralNode) || [] : []) :
        ts.isInterfaceDeclaration(declaration) ? [declaration] :
        (ts.isUnionTypeNode(declaration.type) ? declaration.type.types : [declaration.type]).filter(ts.isTypeLiteralNode);
    const members = literals.flatMap(node => [...node.members]).filter(node =>
        ts.isPropertySignature(node) && ts.isIdentifier(node.name) && node.name.text === spec.member);
    if (members.length !== 1 || !members[0].type) throw new Error(`unexpected slot: ${spec.declaration}.${spec.member}`);
    return { ...spec, node: members[0], references: 0, copies: [], aliases: [], writes: [], escapes: [] };
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
        if (ts.isDeleteExpression(parent) || ((ts.isForOfStatement(parent) || ts.isForInStatement(parent)) && parent.initializer === current)) return parent;
        return undefined;
    }
}
function copiedReceiver(node) {
    let current = node;
    while (current.parent) {
        const parent = current.parent;
        if (ts.isParenthesizedExpression(parent) || ts.isNonNullExpression(parent) ||
            (ts.isBinaryExpression(parent) && [ts.SyntaxKind.AmpersandAmpersandToken,
                ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken].includes(parent.operatorToken.kind)) ||
            (ts.isConditionalExpression(parent) && parent.condition !== current)) {
            current = parent;
            continue;
        }
        return ts.isVariableDeclaration(parent) && parent.initializer === current &&
            (ts.isObjectBindingPattern(parent.name) || ts.isArrayBindingPattern(parent.name));
    }
    return false;
}
const inferredLocal = require('./local-alias.cjs')(checker, tree);
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
                // Binding patterns copy values out of the receiver. They do not
                // retain the receiver object or permit replacement of its slots.
                // Rest bindings create a separate object too. Nested mutable
                // values are a different owner from the selected shallow slot.
                if (copiedReceiver(node)) {
                    owner.copies.push(location(node));
                    continue;
                }
                let target = checker.getContextualType(node);
                if (ts.isAsExpression(node.parent) || ts.isTypeAssertionExpression(node.parent)) {
                    target = checker.getTypeFromTypeNode(node.parent.type);
                }
                if (target) {
                    const slot = property(target, owner.member);
                    if (!ownerOf(slot, owner.node)) {
                        const alias = target.flags & ts.TypeFlags.Any ? inferredLocal(node, owner.member) : undefined;
                        if (alias && !alias.writes.length && !alias.escapes.length) {
                            if (!owner.aliases.some(p => p.declaration === alias.declaration)) owner.aliases.push(alias);
                            continue;
                        }
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
for (const parameter of parameters) {
    if (parameter.writes.length || parameter.escapes.length) throw new Error(
        `cannot prove ${parameter.function}.${parameter.parameter} readonly: ${JSON.stringify({ writes: parameter.writes, escapes: parameter.escapes })}`);
}
const edits = new Map();
for (const owner of owners) {
    if (ts.getCombinedModifierFlags(owner.node) & ts.ModifierFlags.Readonly) continue;
    const file = owner.node.getSourceFile();
    if (!edits.has(file)) edits.set(file, []);
    edits.get(file).push({ start: owner.node.getStart(file), end: owner.node.getStart(file), text: 'readonly ' });
}
for (const parameter of parameters) {
    const type = parameter.node.type;
    const parts = ts.isUnionTypeNode(type) ? type.types : [type];
    const isRange = n => ts.isTypeReferenceNode(n) && ts.isIdentifier(n.typeName) && n.typeName.text === 'TextRange';
    const readonlyRange = n => ts.isTypeReferenceNode(n) && ts.isIdentifier(n.typeName) && n.typeName.text === 'Readonly' &&
        n.typeArguments?.length === 1 && isRange(n.typeArguments[0]);
    if (parts.filter(n => isRange(n) || readonlyRange(n)).length !== 1 ||
        parts.some(n => !isRange(n) && !readonlyRange(n) && n.kind !== ts.SyntaxKind.UndefinedKeyword))
        throw new Error(`unexpected parameter type: ${parameter.function}.${parameter.parameter}`);
    if (parts.some(readonlyRange)) continue;
    const range = parts.find(isRange);
    const file = parameter.node.getSourceFile();
    if (!edits.has(file)) edits.set(file, []);
    edits.get(file).push({ start: range.getStart(file), end: range.end, text: `Readonly<${range.getText(file)}>` });
}
for (const [file, positions] of edits) {
    let text = file.text;
    for (const edit of positions.sort((a, b) => b.start - a.start)) text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
    if (fs.readFileSync(file.fileName, 'utf8') !== file.text) throw new Error(`source changed during audit: ${file.fileName}`);
    fs.writeFileSync(file.fileName, text);
}
console.log(JSON.stringify({ typescript: ts.version, files: edits.size,
    owners: owners.map(({ node, ...owner }) => ({ ...owner, where: location(node) })),
    parameters: parameters.map(({ node, ...parameter }) => parameter) }, null, 2));
