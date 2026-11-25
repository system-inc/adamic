#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const helperBefore = "function widenTypeForVariableLikeDeclaration(type: Type | undefined, declaration: any, reportErrors?: boolean) {\r\n        if (type) {\r\n            // TODO: If back compat with pre-3.0/4.0 libs isn't required, remove the following SymbolConstructor special case transforming `symbol` into `unique symbol`\r\n            if (type.flags & TypeFlags.ESSymbol && isGlobalSymbolConstructor(declaration.parent)) {\r\n                type = getESSymbolLikeTypeForNode(declaration);\r\n            }\r\n            if (reportErrors) {\r\n                reportErrorsFromWidening(declaration, type);\r\n            }\r\n\r\n            // always widen a 'unique symbol' type if the type was created for a different declaration.\r\n            if (type.flags & TypeFlags.UniqueESSymbol && (isBindingElement(declaration) || !tryGetTypeFromEffectiveTypeNode(declaration)) && type.symbol !== getSymbolOfDeclaration(declaration)) {\r\n                type = esSymbolType;\r\n            }\r\n\r\n            return getWidenedType(type);\r\n        }\r\n\r\n        // Rest parameters default to type any[], other parameters default to type any\r\n        type = isParameter(declaration) && declaration.dotDotDotToken ? anyArrayType : anyType;\r\n\r\n        // Report implicit any errors unless this is a private property within an ambient declaration\r\n        if (reportErrors) {\r\n            if (!declarationBelongsToPrivateAmbientMember(declaration)) {\r\n                reportImplicitAny(declaration, type);\r\n            }\r\n        }\r\n        return type;\r\n    }";
const branchBefore = "if (declaration.kind === SyntaxKind.ExportAssignment) {\r\n            type = widenTypeForVariableLikeDeclaration(tryGetTypeFromEffectiveTypeNode(declaration) || checkExpressionCached((declaration as ExportAssignment).expression), declaration);\r\n        }";
const parameterType = "Declaration";
const conditionBefore = 'declaration.kind === SyntaxKind.ExportAssignment';
const guardBefore = '!declarationBelongsToPrivateAmbientMember(declaration)';
const guardAfter = '!(isVariableLike(declaration) && declarationBelongsToPrivateAmbientMember(declaration))';

function one(values, label) {
    assert.equal(values.length, 1, label);
    return values[0];
}

const file = path.join(process.argv[2], 'src/compiler/checker.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(ts.version, '6.0.3');
assert.equal(source.parseDiagnostics.length, 0);

function declaration(statements, name) {
    return one(statements.filter(node => ts.isFunctionDeclaration(node) &&
        node.name?.text === name && node.body), name);
}

const owner = declaration(source.statements, 'createTypeChecker');
const helper = declaration(owner.body.statements, 'widenTypeForVariableLikeDeclaration');
const worker = declaration(owner.body.statements, 'getTypeOfVariableOrParameterOrPropertyWorker');
const parameter = one(helper.parameters.filter(node => node.name.getText(source) === 'declaration'), 'declaration parameter');

function find(root, predicate) {
    const matches = [];
    function visit(node) {
        if (predicate(node)) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return matches;
}

const branch = one(find(worker.body, node => ts.isIfStatement(node) &&
    node.expression.getText(source) === conditionBefore), 'export assignment branch');
const branchText = text.slice(branch.getStart(source), branch.thenStatement.end);
const guard = one(find(helper.body, node => ts.isIfStatement(node) &&
    [guardBefore, guardAfter].includes(node.expression.getText(source))), 'private ambient guard');
const helperWithParameter = helperBefore.replace('declaration: any', 'declaration: ' + parameterType);
const helperAfter = helperWithParameter.replace(guardBefore, guardAfter);

if (helper.getText(source) === helperAfter && branchText === branchBefore) {
    console.log(JSON.stringify({ edits: 0 }));
} else {
    assert.equal(helper.getText(source), helperBefore, 'helper snapshot');
    assert.equal(branchText, branchBefore, 'export branch snapshot');
    assert.equal(parameter.type.getText(source), 'any', 'parameter type snapshot');
    assert.equal(guard.expression.getText(source), guardBefore, 'private ambient guard snapshot');
    const edits = [[parameter.type, parameterType], [guard.expression, guardAfter]];
    let output = text;
    for (const [node, replacement] of edits.sort(([a], [b]) => b.getStart(source) - a.getStart(source))) {
        output = output.slice(0, node.getStart(source)) + replacement + output.slice(node.end);
    }
    fs.writeFileSync(file, output);
    console.log(JSON.stringify({ edits: edits.length }));
}
