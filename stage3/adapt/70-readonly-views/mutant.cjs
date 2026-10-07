#!/usr/bin/env node
'use strict';
// Scratch-tree mutants only. The caller must restore its source after each run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error('expected stock TypeScript 6.0.3');
if (process.argv.length !== 4) throw new Error('usage: node mutant.cjs <tree> <writer-view|audit-writer|audit-escape|public-owner>');
const tree = path.resolve(process.argv[2]);
const name = path.join(tree, 'src/compiler/types.ts');
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
} else throw new Error('unknown mutant');
console.log(process.argv[3]);
