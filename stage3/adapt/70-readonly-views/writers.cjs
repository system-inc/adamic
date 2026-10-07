#!/usr/bin/env node
'use strict';
// This reports two inspected writer families. It makes no claim about the other
// declined sites, and does not infer a reachable type violation from a write.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error('expected stock TypeScript 6.0.3');
if (process.argv.length !== 4) throw new Error('usage: node writers.cjs <tree> <output.json>');
const tree = path.resolve(process.argv[2]);
const roots = ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']);
const program = ts.createProgram(roots, {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    noImplicitReturns: true, noFallthroughCasesInSwitch: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: true, noEmit: true,
    allowImportingTsExtensions: true, module: ts.ModuleKind.ESNext,
    moduleDetection: ts.ModuleDetectionKind.Force, moduleResolution: ts.ModuleResolutionKind.Bundler,
    target: ts.ScriptTarget.ES2024, lib: ['lib.es2024.d.ts'], types: [],
});
const checker = program.getTypeChecker();
const types = program.getSourceFile(path.join(tree, 'src/compiler/types.ts'));
const sourceFile = types.statements.find(node => ts.isInterfaceDeclaration(node) && node.name.text === 'SourceFile');
const originalType = checker.getTypeAtLocation(sourceFile);
const slot = checker.getPropertyOfType(originalType, 'lineMap');
const slotType = checker.getTypeOfSymbolAtLocation(slot, sourceFile);
const wider = types.statements.find(node => ts.isInterfaceDeclaration(node) && node.name.text === 'SourceFileLike');
const widerSlot = wider.members.find(node => ts.isPropertySignature(node) && node.name.text === 'lineMap');
const core = program.getSourceFile(path.join(tree, 'src/compiler/core.ts'));
const clear = core.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'clear');
const parameter = clear.parameters[0];
const parameterSymbol = checker.getSymbolAtLocation(parameter.name);
function where(node) {
    const file = node.getSourceFile();
    const p = file.getLineAndCharacterOfPosition(node.getStart(file));
    return `${path.relative(tree,file.fileName)}:${p.line+1}:${p.character+1}`;
}
const writers = [];
for (const file of program.getSourceFiles()) {
    if (!path.resolve(file.fileName).startsWith(tree + path.sep + 'src' + path.sep)) continue;
    function visit(node) {
        if (ts.isBinaryExpression(node) && node.operatorToken.kind >= ts.SyntaxKind.FirstAssignment &&
            node.operatorToken.kind <= ts.SyntaxKind.LastAssignment && ts.isPropertyAccessExpression(node.left)) {
            const property = checker.getSymbolAtLocation(node.left.name);
            if (property?.declarations?.includes(widerSlot)) {
                const value = checker.getTypeAtLocation(node.right);
                writers.push({ view: 'SourceFileLike.lineMap', declaration: where(widerSlot), where: where(node),
                    expression: node.getText(), value: checker.typeToString(value),
                    original_slot: checker.typeToString(slotType), fits_original_slot: checker.isTypeAssignableTo(value, slotType),
                    classification: 'latent wider-view capacity; this observed write fits SourceFile.lineMap',
                    public: true });
            }
            if (checker.getSymbolAtLocation(node.left.expression) === parameterSymbol) {
                writers.push({ view: 'clear(array: unknown[])', declaration: where(parameter), where: where(node),
                    expression: node.getText(), value: checker.typeToString(checker.getTypeAtLocation(node.right)),
                    classification: 'latent element-widening capacity; setting length to zero inserts no wrong-typed value', public: false });
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(file);
}
fs.writeFileSync(process.argv[3], JSON.stringify({ typescript: ts.version, writers,
    limit: 'Direct checker-resolved writes for two inspected families; not a transitive write census of every declined target.' },null,2)+'\n');
console.log(`${writers.length} inspected writes`);
