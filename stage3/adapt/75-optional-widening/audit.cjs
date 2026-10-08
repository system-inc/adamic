#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
const tree = path.resolve(process.argv[2]);
const configPath = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
assert(!config.error);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath));
const options = { ...parsed.options, strict: true, exactOptionalPropertyTypes: true,
    noUncheckedIndexedAccess: true, verbatimModuleSyntax: true, erasableSyntaxOnly: false, noEmit: true };
const program = ts.createProgram(parsed.fileNames, options);
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(tree, 'src/compiler/types.ts'));
const specs = [
    ['Type', 'InstantiableType', 'resolvedBaseConstraint'],
    ['Type', 'InstantiableType', 'resolvedIndexType'],
    ['Type', 'InstantiableType', 'resolvedStringIndexType'],
    ['SymbolVisibilityResult', 'SymbolAccessibilityResult', 'errorModuleName'],
    ['JsonSourceFile', 'TsConfigSourceFile', 'extendedSourceFiles'],
    ['JsonSourceFile', 'TsConfigSourceFile', 'configFileSpecs'],
];
if (process.argv.includes('--candidate-text-range')) specs.push(['TextRange', 'SourceMapRange', 'source']);
const interfaces = new Map();
function discover(n) { if (ts.isInterfaceDeclaration(n)) interfaces.set(n.name.text, n); ts.forEachChild(n, discover); }
discover(source);
const owners = specs.map(([name, target, property]) => {
    const node = interfaces.get(name);
    const targetNode = interfaces.get(target);
    assert(node && targetNode);
    const type = checker.getTypeAtLocation(node);
    const contract = targetNode.members.find(m => m.name?.getText(source) === property);
    assert(contract?.questionToken && contract.type);
    return { name, target, property, type, contract: checker.getTypeAtLocation(contract), writes: [], declarations: [] };
});
function where(node) {
    const file = node.getSourceFile();
    const p = file.getLineAndCharacterOfPosition(node.getStart(file));
    return `${path.relative(tree, file.fileName)}:${p.line + 1}:${p.character + 1}`;
}
function forbidden(type) { return !!(type.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown)); }
for (const file of program.getSourceFiles()) {
    if (!file.fileName.startsWith(tree + path.sep + 'src' + path.sep)) continue;
    function visit(n) {
        if (ts.isPropertySignature(n) && n.name && n.type) {
            for (const owner of owners) {
                if (n.name.getText(file) !== owner.property || !ts.isInterfaceDeclaration(n.parent)) continue;
                const receiver = checker.getTypeAtLocation(n.parent);
                if (forbidden(receiver) || !checker.isTypeAssignableTo(receiver, owner.type)) continue;
                const value = checker.getTypeAtLocation(n);
                assert(!forbidden(value) && checker.isTypeAssignableTo(value, owner.contract), `incompatible declaration at ${where(n)}`);
                owner.declarations.push({ where: where(n), owner: n.parent.name.text, type: checker.typeToString(value) });
            }
        }
        if (ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
            (ts.isPropertyAccessExpression(n.left) || ts.isElementAccessExpression(n.left))) {
            const access = n.left;
            const key = ts.isPropertyAccessExpression(access) ? access.name.text :
                ts.isStringLiteralLike(access.argumentExpression) ? access.argumentExpression.text : undefined;
            const receiver = checker.getNonNullableType(checker.getTypeAtLocation(access.expression));
            for (const owner of owners) {
                if (key !== owner.property || forbidden(receiver) || !checker.isTypeAssignableTo(receiver, owner.type)) continue;
                const value = checker.getTypeAtLocation(n.right);
                assert(!forbidden(value) && checker.isTypeAssignableTo(value, owner.contract), `incompatible write at ${where(n)}`);
                owner.writes.push({ where: where(n), type: checker.typeToString(value), text: n.getText(file) });
            }
        }
        if (ts.isPropertyAssignment(n) && n.name) {
            const contextual = checker.getContextualType(n.parent);
            const context = contextual && checker.getNonNullableType(contextual);
            for (const owner of owners) {
                if (n.name.getText(file) !== owner.property || !context || forbidden(context) || !checker.isTypeAssignableTo(context, owner.type)) continue;
                const value = checker.getTypeAtLocation(n.initializer);
                assert(!forbidden(value) && checker.isTypeAssignableTo(value, owner.contract), `incompatible initializer at ${where(n)}`);
                owner.writes.push({ where: where(n), type: checker.typeToString(value), text: n.getText(file) });
            }
        }
        ts.forEachChild(n, visit);
    }
    visit(file);
}
const diagnostics = ts.getPreEmitDiagnostics(program).map(d => ({
    code: d.code, where: d.file && d.start !== undefined ? whereAt(d.file, d.start) : '',
    text: ts.flattenDiagnosticMessageText(d.messageText, '\n'),
}));
function whereAt(file, start) { const p=file.getLineAndCharacterOfPosition(start); return `${path.relative(tree,file.fileName)}:${p.line+1}:${p.character+1}`; }
console.log(JSON.stringify({ typescript: ts.version, options, owners: owners.map(({type,contract,...o}) => o), diagnostics }, null, 2));
