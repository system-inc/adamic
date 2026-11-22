#!/usr/bin/env node
'use strict';
// Adaptation 99-cli-factory-creator-properties, #s8mc8gs: createNodeFactory's memoized creators become plain properties.
// Upstream 6.0.3 defines 70 members of its factory object literal as getters, each returning a memoizeOne helper's
// creator for one SyntaxKind (`get createAssignment() { return getBinaryCreateFunction(SyntaxKind.EqualsToken); }`),
// while the NodeFactory interface declares them as methods. Adamic refuses a method that shares its name with a getter,
// so every call through the factory stopped there (112 latent sites, J-vt1zf5t). Each getter becomes a property holding
// the very expression it returned, evaluated once when the factory is built: `createAssignment: getBinaryCreateFunction(
// SyntaxKind.EqualsToken),`. Adamic ruled the internal form (Nov 22, on #s8mc8gs). It changes the emitted JavaScript of
// nodeFactory.ts alone, and keeps typescript.d.ts byte-identical, since the interface doesn't move.
//
// What the adapter holds, refusing before any write when it isn't so:
// - Every getter in the literal but `parenthesizer` and `converters` (left as they are: they're readonly properties in
//   the interface, and their bodies build the factory's own rules, lazily) is exactly `return <helper>(SyntaxKind.<X>)`,
//   optionally cast with `as NodeFactory["<its own name>"]`, where <helper> is a `const` of createNodeFactory's own body
//   declared before the literal and initialized with `memoizeOne(<an arrow returning an arrow>)`. So evaluating it at
//   construction runs no program code but building one closure, and reads no binding before its declaration.
// - There are exactly 70 such getters, by name, in upstream's order. A property already adapted keeps its bytes, so a
//   second run writes nothing; a tree with some getters and some properties is refused as half adapted.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/factory/nodeFactory.ts');
const relative = 'src/compiler/factory/nodeFactory.ts';

// The 70 memoized creators, in the literal's order, each with the helper upstream 6.0.3 calls for it.
const jsdocTypes = [
    ['createJSDocAllType', 'getJSDocPrimaryTypeCreateFunction'],
    ['createJSDocUnknownType', 'getJSDocPrimaryTypeCreateFunction'],
    ['createJSDocNonNullableType', 'getJSDocPrePostfixUnaryTypeCreateFunction'],
    ['updateJSDocNonNullableType', 'getJSDocPrePostfixUnaryTypeUpdateFunction'],
    ['createJSDocNullableType', 'getJSDocPrePostfixUnaryTypeCreateFunction'],
    ['updateJSDocNullableType', 'getJSDocPrePostfixUnaryTypeUpdateFunction'],
    ['createJSDocOptionalType', 'getJSDocUnaryTypeCreateFunction'],
    ['updateJSDocOptionalType', 'getJSDocUnaryTypeUpdateFunction'],
    ['createJSDocVariadicType', 'getJSDocUnaryTypeCreateFunction'],
    ['updateJSDocVariadicType', 'getJSDocUnaryTypeUpdateFunction'],
    ['createJSDocNamepathType', 'getJSDocUnaryTypeCreateFunction'],
    ['updateJSDocNamepathType', 'getJSDocUnaryTypeUpdateFunction'],
];
const jsdocTags = [];
for (const [tag, kind] of [['Type', 'TypeLike'], ['Return', 'TypeLike'], ['This', 'TypeLike'], ['Author', 'Simple'], ['Class', 'Simple'],
    ['Public', 'Simple'], ['Private', 'Simple'], ['Protected', 'Simple'], ['Readonly', 'Simple'], ['Override', 'Simple'],
    ['Deprecated', 'Simple'], ['Throws', 'TypeLike'], ['Satisfies', 'TypeLike']]) {
    jsdocTags.push([`createJSDoc${tag}Tag`, `getJSDoc${kind}TagCreateFunction`], [`updateJSDoc${tag}Tag`, `getJSDoc${kind}TagUpdateFunction`]);
}
const operators = [
    ...['Comma', 'Assignment', 'LogicalOr', 'LogicalAnd', 'BitwiseOr', 'BitwiseXor', 'BitwiseAnd', 'StrictEquality', 'StrictInequality',
        'Equality', 'Inequality', 'LessThan', 'LessThanEquals', 'GreaterThan', 'GreaterThanEquals', 'LeftShift', 'RightShift',
        'UnsignedRightShift', 'Add', 'Subtract', 'Multiply', 'Divide', 'Modulo', 'Exponent'].map(name => [`create${name}`, 'getBinaryCreateFunction']),
    ...['PrefixPlus', 'PrefixMinus', 'PrefixIncrement', 'PrefixDecrement', 'BitwiseNot', 'LogicalNot'].map(name => [`create${name}`, 'getPrefixUnaryCreateFunction']),
    ...['PostfixIncrement', 'PostfixDecrement'].map(name => [`create${name}`, 'getPostfixUnaryCreateFunction']),
];
const creators = [...jsdocTypes, ...jsdocTags, ...operators];
if (creators.length !== 70 || new Set(creators.map(([name]) => name)).size !== 70) throw new Error('the creator list is not 70 distinct names');
const helperOf = new Map(creators);
const kept = new Set(['parenthesizer', 'converters']);

const text = fs.readFileSync(fileName, 'utf8');
// The tree keeps upstream's CRLF line endings. Each edit replaces one member's text with one line and writes no line
// ending, so the file's endings are kept as they are.
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const factoryFunctions = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'createNodeFactory');
if (factoryFunctions.length !== 1) throw new Error(`expected one createNodeFactory in ${relative}, found ${factoryFunctions.length}`);
const body = factoryFunctions[0].body.statements;
const literalIndex = body.findIndex(statement => ts.isVariableStatement(statement)
    && statement.declarationList.declarations.length === 1 && statement.declarationList.declarations[0].name.getText(source) === 'factory');
if (literalIndex < 0) throw new Error('createNodeFactory declares no factory');
const declaration = body[literalIndex].declarationList.declarations[0];
if (!(body[literalIndex].declarationList.flags & ts.NodeFlags.Const) || declaration.type?.getText(source) !== 'NodeFactory'
    || !declaration.initializer || !ts.isObjectLiteralExpression(declaration.initializer))
    throw new Error('createNodeFactory\'s factory is not `const factory: NodeFactory = { ... }`');
const literal = declaration.initializer;

// The helpers a creator may call: consts of createNodeFactory's body before the literal, each `memoizeOne(<arrow
// returning an arrow>)`, whose construction-time call builds one closure and nothing else.
const helpers = new Map();
for (const statement of body.slice(0, literalIndex)) {
    if (!ts.isVariableStatement(statement) || !(statement.declarationList.flags & ts.NodeFlags.Const)) continue;
    for (const variable of statement.declarationList.declarations) {
        const initializer = variable.initializer;
        if (!ts.isIdentifier(variable.name) || !initializer || !ts.isCallExpression(initializer)) continue;
        if (initializer.expression.getText(source) !== 'memoizeOne' || initializer.arguments.length !== 1) continue;
        const make = initializer.arguments[0];
        if (ts.isArrowFunction(make) && make.parameters.length === 1 && ts.isArrowFunction(make.body)) helpers.set(variable.name.text, make);
    }
}
for (const helper of new Set(helperOf.values())) {
    if (!helpers.has(helper)) throw new Error(`${helper} is not a const memoizeOne(arrow => arrow) declared before the factory literal`);
}

// The creator a getter returns: `<helper>(SyntaxKind.<X>)` or `<helper><T>(SyntaxKind.<X>)`, optionally cast to
// NodeFactory["<name>"]. Anything else refuses, so an expression with a side effect never moves to construction.
function creatorExpression(name, expression) {
    let call = expression;
    if (ts.isAsExpression(call)) {
        if (call.type.getText(source) !== `NodeFactory["${name}"]`) throw new Error(`${name} casts its creator to ${call.type.getText(source)}`);
        call = call.expression;
    }
    if (!ts.isCallExpression(call) || !ts.isIdentifier(call.expression) || call.expression.text !== helperOf.get(name)
        || call.arguments.length !== 1 || !ts.isPropertyAccessExpression(call.arguments[0])
        || call.arguments[0].expression.getText(source) !== 'SyntaxKind' || (call.typeArguments || []).length > 1)
        throw new Error(`${name} returns ${expression.getText(source)}, not ${helperOf.get(name)}(SyntaxKind.<kind>)`);
    return expression.getText(source);
}

const seen = [];
const edits = [];
for (const member of literal.properties) {
    const name = member.name && ts.isIdentifier(member.name) ? member.name.text : undefined;
    if (ts.isGetAccessorDeclaration(member)) {
        if (kept.has(name)) continue;
        if (!helperOf.has(name)) throw new Error(`unexpected getter ${name} in createNodeFactory's literal`);
        const statements = member.body?.statements || [];
        if (member.parameters.length || member.type || member.modifiers?.length || statements.length !== 1
            || !ts.isReturnStatement(statements[0]) || !statements[0].expression)
            throw new Error(`getter ${name} is not a single return`);
        const expression = creatorExpression(name, statements[0].expression);
        seen.push({ name, form: 'getter' });
        edits.push({ name, at: member.getStart(source), end: member.end, text: `${name}: ${expression}` });
    } else if (helperOf.has(name)) {
        if (!ts.isPropertyAssignment(member)) throw new Error(`${name} is neither its getter nor an adapted property`);
        creatorExpression(name, member.initializer);
        seen.push({ name, form: 'property' });
    }
}
const order = seen.map(entry => entry.name).join(',');
if (order !== creators.map(([name]) => name).join(','))
    throw new Error(`createNodeFactory's creators are not upstream's 70 in order: ${order}`);
const forms = new Set(seen.map(entry => entry.form));
if (forms.size !== 1) throw new Error('half adapted: createNodeFactory holds both getters and adapted properties');

let adapted = text;
for (const edit of [...edits].sort((a, b) => b.at - a.at)) adapted = adapted.slice(0, edit.at) + edit.text + adapted.slice(edit.end);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${relative}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#s8mc8gs', ruling: 'Adamic, Nov 22, on #s8mc8gs: the internal form',
    file: relative, edited: edits.length, alreadyAdapted: forms.has('property') ? 70 : 0,
    helpers: [...new Set(helperOf.values())] }, null, 2));
