#!/usr/bin/env node
'use strict';
// Scratch-tree mutants only. The caller must restore its source after each run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error('expected stock TypeScript 6.0.3');
if (process.argv.length !== 4) throw new Error('usage: node mutant.cjs <tree> <mode>');
const tree = path.resolve(process.argv[2]);
const mode = process.argv[3];
const name = path.join(tree, mode === 'public-input-writer' ? 'src/compiler/factory/utilitiesPublic.ts' : (mode.startsWith('inferred-local-') || mode === 'payload-writer') ? 'src/compiler/moduleNameResolver.ts' :
    mode.startsWith('parameter-') ? 'src/compiler/utilities.ts' : 'src/compiler/types.ts');
const text = fs.readFileSync(name, 'utf8');
const source = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
if (process.argv[3] === 'writer-view') {
    const declaration = source.statements.find(node => ts.isInterfaceDeclaration(node) && node.name.text === 'SourceFileLike');
    const member = declaration?.members.find(node => ts.isPropertySignature(node) && node.name.text === 'lineMap');
    if (!member || ts.getCombinedModifierFlags(member) & ts.ModifierFlags.Readonly) throw new Error('unexpected writer view');
    const position = member.getStart(source);
    fs.writeFileSync(name, text.slice(0, position) + 'readonly ' + text.slice(position));
} else if (process.argv[3] === 'audit-writer') {
    // This function takes the actual internal view and writes the selected slot.
    // The adaptation must reject it before any edit, independently of a build.
    const declaration = ts.factory.createFunctionDeclaration(undefined, undefined, 'readonlyViewAuditMutant', undefined,
        [ts.factory.createParameterDeclaration(undefined, undefined, 'result', undefined,
            ts.factory.createTypeReferenceNode('EvaluatorResult'))], ts.factory.createKeywordTypeNode(ts.SyntaxKind.VoidKeyword),
        ts.factory.createBlock([ts.factory.createExpressionStatement(ts.factory.createBinaryExpression(
            ts.factory.createPropertyAccessExpression(ts.factory.createIdentifier('result'), 'value'),
            ts.SyntaxKind.EqualsToken, ts.factory.createIdentifier('undefined')))], true));
    fs.writeFileSync(name, text + '\n' + ts.createPrinter().printNode(ts.EmitHint.Unspecified, declaration, source) + '\n');
} else if (process.argv[3] === 'audit-escape') {
    const declaration = ts.factory.createFunctionDeclaration(undefined, undefined, 'readonlyViewEscapeMutant', undefined,
        [ts.factory.createParameterDeclaration(undefined, undefined, 'result', undefined,
            ts.factory.createTypeReferenceNode('EvaluatorResult'))], ts.factory.createKeywordTypeNode(ts.SyntaxKind.VoidKeyword),
        ts.factory.createBlock([
            ts.factory.createVariableStatement(undefined, ts.factory.createVariableDeclarationList([
                ts.factory.createVariableDeclaration('wider', undefined, ts.factory.createTypeLiteralNode([
                    ts.factory.createPropertySignature(undefined, 'value', undefined, ts.factory.createUnionTypeNode([
                        ts.factory.createKeywordTypeNode(ts.SyntaxKind.StringKeyword),
                        ts.factory.createKeywordTypeNode(ts.SyntaxKind.NumberKeyword),
                        ts.factory.createKeywordTypeNode(ts.SyntaxKind.UndefinedKeyword),
                    ])),
                ]), ts.factory.createIdentifier('result')),
            ], ts.NodeFlags.Const)),
            ts.factory.createExpressionStatement(ts.factory.createBinaryExpression(
                ts.factory.createPropertyAccessExpression(ts.factory.createIdentifier('wider'), 'value'),
                ts.SyntaxKind.EqualsToken, ts.factory.createIdentifier('undefined'))),
        ], true));
    fs.writeFileSync(name, text + '\n' + ts.createPrinter().printNode(ts.EmitHint.Unspecified, declaration, source) + '\n');
} else if (process.argv[3] === 'public-owner') {
    const declaration = source.statements.find(node => ts.isInterfaceDeclaration(node) && node.name.text === 'EvaluatorResult');
    const tag = declaration && ts.getJSDocTags(declaration).find(node => node.tagName.text === 'internal');
    if (!tag) throw new Error('missing internal tag');
    fs.writeFileSync(name, text.slice(0, tag.getStart(source)) + text.slice(tag.end));
} else if (mode === 'public-input-writer') {
    const fn = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'setTextRange' && n.body);
    if (!fn) throw new Error('missing public input owner');
    const fragment = ts.createSourceFile('mutant-input.a', 'if (location) location.pos = 0;', ts.ScriptTarget.Latest, true);
    const printed = ts.createPrinter().printNode(ts.EmitHint.Unspecified, fragment.statements[0], fragment);
    const at = fn.body.getStart(source) + 1;
    fs.writeFileSync(name, text.slice(0, at) + '\n' + printed + '\n' + text.slice(at));
} else if (mode === 'payload-writer') {
    const worker = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'nodeModuleNameResolverWorker');
    const matches = [];
    function find(n) {
        if (ts.isFunctionDeclaration(n) && n.name?.text === 'tryResolve' && n.type &&
            ts.isTypeReferenceNode(n.type) && n.type.typeName.getText(source) === 'SearchResult') matches.push(n);
        ts.forEachChild(n, find);
    }
    find(worker);
    if (matches.length !== 1 || !matches[0].body) throw new Error('missing payload owner');
    const fragment = ts.createSourceFile('mutant-input.a',
        '((view: NonNullable<NonNullable<ReturnType<typeof tryResolve>>["value"]>) => { view.isExternalLibraryImport = true; });',
        ts.ScriptTarget.Latest, true);
    if (fragment.parseDiagnostics.length) throw new Error('mutant parse failed');
    const printed = ts.createPrinter().printNode(ts.EmitHint.Unspecified, fragment.statements[0], fragment);
    const at = matches[0].body.getStart(source) + 1;
    fs.writeFileSync(name, text.slice(0, at) + '\n' + printed + '\n' + text.slice(at));
} else if (mode.startsWith('inferred-local-')) {
    const fn = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'nodeModuleNameResolverWorker');
    let declaration;
    function find(n) {
        if (ts.isVariableDeclaration(n) && ts.isIdentifier(n.name) && n.name.text === 'result' && !n.initializer) declaration = n;
        ts.forEachChild(n, find);
    }
    find(fn);
    if (!declaration || declaration.type) throw new Error('missing inferred local receiver');
    if (mode === 'inferred-local-any') {
        fs.writeFileSync(name, text.slice(0, declaration.name.end) + ': any' + text.slice(declaration.name.end));
    } else if (mode === 'inferred-local-escape') {
        const resultReturn = fn.body.statements.find(ts.isReturnStatement);
        const fragment = ts.createSourceFile('mutant-input.a',
            'if (result) { const wider: any = result; wider.value = undefined; }', ts.ScriptTarget.Latest, true);
        const insertion = fragment.statements.map(n => ts.createPrinter().printNode(ts.EmitHint.Unspecified, n, fragment)).join('\n');
        const at = resultReturn.getStart(source);
        fs.writeFileSync(name, text.slice(0, at) + insertion + '\n' + text.slice(at));
    } else throw new Error('unknown inferred-local mutant');
} else if (mode.startsWith('parameter-')) {
    const fnName = mode === 'parameter-callee-writer' ? 'getStartPositionOfRange' : 'nodeIsSynthesized';
    const declaration = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === fnName && n.body);
    if (!declaration) throw new Error('missing range reader body');
    const statements = {
        'parameter-writer': 'range.pos = 0;',
        'parameter-callee-writer': 'range.pos = 0;',
        'parameter-escape': 'const wider: TextRange = range; wider.pos = 0;',
        'parameter-destructure-writer': '[range.pos] = [0];',
        'parameter-iteration-writer': 'for (range.pos of [0]) {}',
    };
    if (!statements[mode]) throw new Error('unknown parameter mutant');
    const fragment = ts.createSourceFile('mutant-input.a', statements[mode], ts.ScriptTarget.Latest, true);
    if (fragment.parseDiagnostics.length) throw new Error('mutant parse failed');
    const insertion = fragment.statements.map(n => ts.createPrinter().printNode(ts.EmitHint.Unspecified, n, fragment)).join('\n');
    const at = declaration.body.getStart(source) + 1;
    fs.writeFileSync(name, text.slice(0, at) + '\n' + insertion + '\n' + text.slice(at));
} else if (mode === 'diagnostic-writer') {
    const fragment = ts.createSourceFile('mutant-input.a',
        '/** @internal */ export function diagnosticViewWriterMutant(diagnostic: DiagnosticRelatedInformation): void { diagnostic.file = undefined; }',
        ts.ScriptTarget.Latest, true);
    const printed = ts.createPrinter().printNode(ts.EmitHint.Unspecified, fragment.statements[0], fragment);
    fs.writeFileSync(name, text + '\n' + printed + '\n');
} else if (mode === 'audit-iteration-writer') {
    const fragment = ts.createSourceFile('mutant-input.a',
        'function readonlyViewIterationMutant(result: EvaluatorResult) { for (result.value of [undefined]) {} }',
        ts.ScriptTarget.Latest, true);
    const printed = ts.createPrinter().printNode(ts.EmitHint.Unspecified, fragment.statements[0], fragment);
    fs.writeFileSync(name, text + '\n' + printed + '\n');
} else throw new Error('unknown mutant');
console.log(process.argv[3]);
