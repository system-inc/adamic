'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

function one(nodes, predicate, label) {
    const matches = nodes.filter(predicate);
    assert.equal(matches.length, 1, `expected one ${label}`);
    return matches[0];
}

function functionNamed(statements, name) {
    return one(statements, node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body, name);
}

const before = `function getRequiresAtTopOfFile(sourceFile: SourceFile): readonly RequireOrImportCall[] {
    let nonRequireStatementCount = 0;
    let requires: RequireOrImportCall[] | undefined;
    for (const statement of sourceFile.statements) {
        if (nonRequireStatementCount > 3) {
            break;
        }
        if (isRequireVariableStatement(statement)) {
            requires = concatenate(requires, statement.declarationList.declarations.map(d => d.initializer));
        }
        else if (isExpressionStatement(statement) && isRequireCall(statement.expression, /*requireStringLiteralLikeArgument*/ true)) {
            requires = append(requires, statement.expression);
        }
        else {
            nonRequireStatementCount++;
        }
    }
    return requires || emptyArray;
}`;

const after = `function getRequiresAtTopOfFile(sourceFile: SourceFile): readonly StringLiteralLike[] {
    let nonRequireStatementCount = 0;
    const requires: StringLiteralLike[] = [];
    for (const statement of sourceFile.statements) {
        if (nonRequireStatementCount > 3) {
            break;
        }
        if (isRequireVariableStatement(statement)) {
            for (const declaration of statement.declarationList.declarations) {
                requires.push(declaration.initializer.arguments[0]);
            }
        }
        else if (isExpressionStatement(statement) && isRequireCall(statement.expression, /*requireStringLiteralLikeArgument*/ true)) {
            requires.push(statement.expression.arguments[0]);
        }
        else {
            nonRequireStatementCount++;
        }
    }
    return requires;
}`;

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const file = path.join(tree, 'src/compiler/utilities.ts');
    const text = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const newline = text.includes('\r\n') ? '\r\n' : '\n';

    const collector = functionNamed(source.statements, 'getRequiresAtTopOfFile');
    const preference = functionNamed(source.statements, 'getModuleSpecifierEndingPreference');
    const infer = functionNamed(preference.body.statements, 'inferPreference');
    const bindings = infer.body.statements.filter(ts.isVariableStatement)
        .flatMap(statement => [...statement.declarationList.declarations]);
    const specifiers = one(bindings, node => ts.isIdentifier(node.name) && node.name.text === 'specifiers', 'specifiers').initializer;
    assert(specifiers && ts.isConditionalExpression(specifiers) && ts.isConditionalExpression(specifiers.whenFalse));
    const requireSpecifiers = specifiers.whenFalse.whenTrue;
    const call = 'getRequiresAtTopOfFile(sourceFile)';

    const namespace = one(source.statements, node => ts.isImportDeclaration(node) &&
        node.moduleSpecifier.getText(source) === '"./_namespaces/ts.js"', 'compiler namespace import');
    const imports = namespace.importClause?.namedBindings;
    assert(imports && ts.isNamedImports(imports));
    const append = imports.elements.filter(node => node.name.text === 'append');

    const oldCollector = before.replaceAll('\n', newline);
    const newCollector = after.replaceAll('\n', newline);
    if (collector.getText(source) === newCollector) {
        assert.equal(requireSpecifiers.getText(source), call);
        assert.equal(append.length, 0);
        return { files: 0, edits: 0 };
    }

    assert.equal(collector.getText(source), oldCollector);
    assert.equal(requireSpecifiers.getText(source), call + '.map(r => r.arguments[0])');
    assert.equal(append.length, 1);
    const importStart = text.lastIndexOf('\n', append[0].getStart(source)) + 1;
    const importEnd = text.indexOf('\n', append[0].end) + 1;
    assert.equal(text.slice(importStart, importEnd), '    append,' + newline);

    const edits = [
        [collector.getStart(source), collector.end, newCollector],
        [requireSpecifiers.getStart(source), requireSpecifiers.end, call],
        [importStart, importEnd, ''],
    ];
    let output = text;
    for (const [start, end, replacement] of edits.sort((a, b) => b[0] - a[0])) {
        output = output.slice(0, start) + replacement + output.slice(end);
    }
    fs.writeFileSync(file, output);
    return { files: 1, edits: edits.length, javascriptChanging: true };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
