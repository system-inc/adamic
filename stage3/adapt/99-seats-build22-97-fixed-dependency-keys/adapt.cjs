#!/usr/bin/env node
'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
const relative = 'src/compiler/utilities.ts';
const file = path.join(process.argv[2], relative);
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0);

const declarations = source.statements.filter(statement =>
    ts.isFunctionDeclaration(statement) && statement.name?.text === 'createComputedCompilerOptions');
assert.equal(declarations.length, 1, 'expected one createComputedCompilerOptions declaration');
const [typeParameter] = declarations[0].typeParameters;
assert.equal(declarations[0].typeParameters.length, 1);
assert.equal(typeParameter.name.text, 'T');
const constraint = typeParameter.constraint;
assert(constraint, 'expected a constrained dependency table');

const dependencies = {
    allowImportingTsExtensions: ["rewriteRelativeImportExtensions"],
    target: [],
    module: ["target"],
    moduleResolution: ["module", "target"],
    moduleDetection: ["module", "target"],
    isolatedModules: ["verbatimModuleSyntax"],
    esModuleInterop: [],
    allowSyntheticDefaultImports: [],
    resolvePackageJsonExports: ["moduleResolution", "module", "target"],
    resolvePackageJsonImports: ["moduleResolution", "resolvePackageJsonExports", "module", "target"],
    resolveJsonModule: ["moduleResolution", "module", "target"],
    declaration: ["composite"],
    preserveConstEnums: ["isolatedModules", "verbatimModuleSyntax"],
    incremental: ["composite"],
    declarationMap: ["declaration", "composite"],
    allowJs: ["checkJs"],
    useDefineForClassFields: ["target", "module"],
    noImplicitAny: ["strict"],
    noImplicitThis: ["strict"],
    strictNullChecks: ["strict"],
    strictFunctionTypes: ["strict"],
    strictBindCallApply: ["strict"],
    strictPropertyInitialization: ["strict"],
    strictBuiltinIteratorReturn: ["strict"],
    alwaysStrict: [],
    useUnknownInCatchVariables: ["strict"],
};

const tables = source.statements.filter(ts.isVariableStatement)
    .flatMap(statement => [...statement.declarationList.declarations])
    .filter(declaration => ts.isIdentifier(declaration.name) && declaration.name.text === '_computedOptions');
assert.equal(tables.length, 1, 'expected one computed options table');
const call = tables[0].initializer;
assert(call && ts.isCallExpression(call) && call.expression.getText(source) === 'createComputedCompilerOptions');
assert.equal(call.arguments.length, 1);
const [table] = call.arguments;
assert(ts.isObjectLiteralExpression(table), 'expected a literal dependency table');
assert.deepEqual(table.properties.map(property => {
    assert(ts.isPropertyAssignment(property) && ts.isIdentifier(property.name));
    assert(ts.isObjectLiteralExpression(property.initializer));
    const list = property.initializer.properties.find(member =>
        ts.isPropertyAssignment(member) && ts.isIdentifier(member.name) && member.name.text === 'dependencies');
    assert(list && ts.isArrayLiteralExpression(list.initializer));
    const values = list.initializer.elements.map(element => {
        assert(ts.isStringLiteral(element));
        return element.text;
    });
    return [property.name.text, values];
}), Object.entries(dependencies), 'computed option dependencies changed');

const aliases = source.statements.filter(statement => ts.isTypeAliasDeclaration(statement)
    && ['CompilerOptionKeys', 'ComputedCompilerOptionDependencies'].includes(statement.name.text));
assert.equal(aliases.length, 1, 'expected one dependency type alias');
const mapped = declarations[0].parameters[0]?.type;
assert(mapped && ts.isMappedTypeNode(mapped) && mapped.typeParameter.constraint, 'expected mapped options parameter');

const newline = text.includes('\r\n') ? '\r\n' : '\n';
const fields = Object.entries(dependencies).map(([key, values]) => {
    const union = values.map(value => JSON.stringify(value)).join(' | ');
    const elementType = values.length > 1 ? `(${union})` : union || 'never';
    return `    ${key}: ${elementType}[];`;
});
const fixed = 'ComputedCompilerOptionDependencies';
const edits = [
    {
        node: aliases[0],
        before: 'type CompilerOptionKeys = keyof { [K in keyof CompilerOptions as string extends K ? never : K]: never; };',
        after: [`type ${fixed} = {`, ...fields, '};'].join(newline),
    },
    { node: constraint, before: 'Record<string, CompilerOptionKeys[]>', after: fixed },
    { node: mapped.typeParameter.constraint, before: 'keyof T & CompilerOptionKeys | StrictOptionName', after: `keyof ${fixed}` },
];
for (const edit of edits) {
    edit.start = edit.node.getStart(source);
    const actual = edit.node.getText(source);
    assert(actual === edit.before || actual === edit.after, `dependency type snapshot changed at offset ${edit.start}`);
    edit.pending = actual === edit.before;
}
const pending = edits.filter(edit => edit.pending);
assert(pending.length === 0 || pending.length === edits.length, 'partial dependency adaptation');
if (pending.length) {
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'source changed during adaptation');
    let result = text;
    for (const edit of pending.sort((left, right) => right.start - left.start)) {
        result = result.slice(0, edit.start) + edit.after + result.slice(edit.node.end);
    }
    fs.writeFileSync(file, result);
}
console.log(JSON.stringify({
    file: relative,
    edits: edits.map(edit => {
        const location = source.getLineAndCharacterOfPosition(edit.start);
        return { line: location.line + 1, column: location.character + 1, edited: edit.pending };
    }),
}));
