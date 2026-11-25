'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

// Each site is found by its nodes and pinned by its text, before and after; no line number or digest names one.
const aliasBefore = `export type BindableObjectDefinePropertyCall = CallExpression & {
    readonly arguments: readonly [BindableStaticNameExpression, StringLiteralLike | NumericLiteral, ObjectLiteralExpression] & Readonly<TextRange>;
};`;
const aliasAfter = aliasBefore.replace('NumericLiteral, ObjectLiteralExpression]', 'NumericLiteral, Expression]');

const expandoCall = 'hasExpandoValueProperty(node.arguments[2], node.arguments[1].text === "prototype")';
const expandoCheck = 'isObjectLiteralExpression(node.arguments[2]) ? ';
const expandoBefore = `export function getAssignedExpandoInitializer(node: Node | undefined): Expression | undefined {
    if (node && node.parent && isBinaryExpression(node.parent) && node.parent.operatorToken.kind === SyntaxKind.EqualsToken) {
        const isPrototypeAssignment = isPrototypeAccess(node.parent.left);
        return getExpandoInitializer(node.parent.right, isPrototypeAssignment) ||
            getDefaultedExpandoInitializer(node.parent.left, node.parent.right, isPrototypeAssignment);
    }
    if (node && isCallExpression(node) && isBindableObjectDefinePropertyCall(node)) {
        const result = ${expandoCall};
        if (result) {
            return result;
        }
    }
}`;
const expandoAfter = expandoBefore.replace(expandoCall, expandoCheck + expandoCall + ' : undefined');

const accessorCheck = 'isObjectLiteralExpression(declaration.arguments[2]) && ';
const accessorBefore = `function bindPotentiallyNewExpandoMemberToNamespace(declaration: BindableStaticAccessExpression | CallExpression, namespaceSymbol: Symbol | undefined, isPrototypeProperty: boolean) {
        if (!namespaceSymbol || !isExpandoSymbol(namespaceSymbol)) {
            return;
        }

        // Set up the members collection if it doesn't exist already
        const symbolTable = isPrototypeProperty ?
            (namespaceSymbol.members || (namespaceSymbol.members = createSymbolTable())) :
            (namespaceSymbol.exports || (namespaceSymbol.exports = createSymbolTable()));

        let includes = SymbolFlags.None;
        let excludes = SymbolFlags.None;
        // Method-like
        if (isFunctionLikeDeclaration(getAssignedExpandoInitializer(declaration)!)) {
            includes = SymbolFlags.Method;
            excludes = SymbolFlags.MethodExcludes;
        }
        // Maybe accessor-like
        else if (isCallExpression(declaration) && isBindableObjectDefinePropertyCall(declaration)) {
            if (
                some(declaration.arguments[2].properties, p => {
                    const id = getNameOfDeclaration(p);
                    return !!id && isIdentifier(id) && idText(id) === "set";
                })
            ) {
                // We mix in \`SymbolFLags.Property\` so in the checker \`getTypeOfVariableParameterOrProperty\` is used for this
                // symbol, instead of \`getTypeOfAccessor\` (which will assert as there is no real accessor declaration)
                includes |= SymbolFlags.SetAccessor | SymbolFlags.Property;
                excludes |= SymbolFlags.SetAccessorExcludes;
            }
            if (
                some(declaration.arguments[2].properties, p => {
                    const id = getNameOfDeclaration(p);
                    return !!id && isIdentifier(id) && idText(id) === "get";
                })
            ) {
                includes |= SymbolFlags.GetAccessor | SymbolFlags.Property;
                excludes |= SymbolFlags.GetAccessorExcludes;
            }
        }

        if (includes === SymbolFlags.None) {
            includes = SymbolFlags.Property;
            excludes = SymbolFlags.PropertyExcludes;
        }

        declareSymbol(symbolTable, namespaceSymbol, declaration, includes | SymbolFlags.Assignment, excludes & ~SymbolFlags.Assignment);
    }`;
const accessorAfter = accessorBefore.replaceAll('                some(declaration.arguments[2].properties, p => {',
    '                ' + accessorCheck + 'some(declaration.arguments[2].properties, p => {');

const setterCheck = 'isObjectLiteralExpression(d.arguments[2]) ? ';
const setterBefore = `d => {
                                if (d.kind === SyntaxKind.SetAccessor) {
                                    return d as SetAccessorDeclaration;
                                }
                                if (isCallExpression(d) && isBindableObjectDefinePropertyCall(d)) {
                                    return forEach(d.arguments[2].properties, propDecl => {
                                        const id = getNameOfDeclaration(propDecl);
                                        if (!!id && isIdentifier(id) && idText(id) === "set") {
                                            return propDecl;
                                        }
                                    });
                                }
                            }`;
const setterAfter = setterBefore
    .replace('return forEach(d.arguments[2].properties', 'return ' + setterCheck + 'forEach(d.arguments[2].properties')
    .replace(/\}\);\n( {32}\}\n {28}\})$/, '}) : undefined;\n$1');

function parse(tree, relative) {
    const file = path.join(tree, relative);
    const text = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `Cannot parse ${relative}`);
    return { file, relative, text, source, newline: text.includes('\r\n') ? '\r\n' : '\n' };
}

function find(root, predicate) {
    const found = [];
    (function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(root);
    return found;
}

function one(nodes, label) {
    assert.equal(nodes.length, 1, `expected one ${label}, found ${nodes.length}`);
    return nodes[0];
}

function functionNamed(unit, name) {
    return one(find(unit.source, node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body !== undefined), name);
}

// A call whose callee is the plain identifier named and whose first argument reads the third argument's properties
// (or, for the expando check, is the third argument itself).
function callsOf(root, unit, callee, firstArgument) {
    return find(root, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === callee &&
        node.arguments.length > 0 && node.arguments[0].getText(unit.source) === firstArgument);
}

// The check is the ts namespace's own isObjectLiteralExpression, imported by name, so each file must import it already.
function assertImportsCheck(unit) {
    const imports = unit.source.statements.filter(node => ts.isImportDeclaration(node) &&
        ts.isStringLiteral(node.moduleSpecifier) && node.moduleSpecifier.text === './_namespaces/ts.js' &&
        !node.importClause?.isTypeOnly && node.importClause?.namedBindings && ts.isNamedImports(node.importClause.namedBindings));
    const specifiers = imports.flatMap(node => node.importClause.namedBindings.elements)
        .filter(node => node.name.text === 'isObjectLiteralExpression');
    assert.equal(specifiers.length, 1, `${unit.relative} imports isObjectLiteralExpression once from ./_namespaces/ts.js`);
    assert(!specifiers[0].isTypeOnly && !specifiers[0].propertyName, `${unit.relative} imports isObjectLiteralExpression as a value`);
}

function sites(tree) {
    const types = parse(tree, 'src/compiler/types.ts');
    const utilities = parse(tree, 'src/compiler/utilities.ts');
    const binder = parse(tree, 'src/compiler/binder.ts');
    const checker = parse(tree, 'src/compiler/checker.ts');
    for (const unit of [utilities, binder, checker]) assertImportsCheck(unit);

    const alias = one(types.source.statements.filter(node => ts.isTypeAliasDeclaration(node) &&
        node.name.text === 'BindableObjectDefinePropertyCall'), 'BindableObjectDefinePropertyCall');
    assert(ts.isIntersectionTypeNode(alias.type) && alias.type.types.length === 2);
    const literal = alias.type.types[1];
    assert(ts.isTypeLiteralNode(literal) && literal.members.length === 1);
    const member = literal.members[0];
    assert(ts.isPropertySignature(member) && member.name.getText(types.source) === 'arguments' && member.type);
    assert(ts.isIntersectionTypeNode(member.type));
    const view = member.type.types[0];
    assert(ts.isTypeOperatorNode(view) && view.operator === ts.SyntaxKind.ReadonlyKeyword && ts.isTupleTypeNode(view.type));
    assert.equal(view.type.elements.length, 3);
    const third = view.type.elements[2];
    assert(ts.isTypeReferenceNode(third) && ts.isIdentifier(third.typeName) && !third.typeArguments);

    const expando = functionNamed(utilities, 'getAssignedExpandoInitializer');
    const expandoRead = one(callsOf(expando, utilities, 'hasExpandoValueProperty', 'node.arguments[2]'), 'hasExpandoValueProperty read');
    const accessor = functionNamed(binder, 'bindPotentiallyNewExpandoMemberToNamespace');
    const accessorReads = callsOf(accessor, binder, 'some', 'declaration.arguments[2].properties');
    assert.equal(accessorReads.length, 2, 'expected the set and get reads in bindPotentiallyNewExpandoMemberToNamespace');
    const serializer = functionNamed(checker, 'makeSerializePropertySymbol');
    const setterRead = one(callsOf(serializer, checker, 'forEach', 'd.arguments[2].properties'), 'makeSerializePropertySymbol setter read');
    // The reads' places are pinned below by the whole text of the function or arrow holding them, before and after.
    let setterFinder = setterRead.parent;
    while (!ts.isArrowFunction(setterFinder)) setterFinder = setterFinder.parent;
    return {
        types, utilities, binder, checker,
        alias, third, expando, expandoRead, accessor, accessorReads, setterFinder, setterRead,
    };
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const found = sites(tree);
    const { types, utilities, binder, checker } = found;
    const snapshot = (unit, text) => text.replaceAll('\n', unit.newline);
    const texts = [
        [types, found.alias, aliasBefore, aliasAfter],
        [utilities, found.expando, expandoBefore, expandoAfter],
        [binder, found.accessor, accessorBefore, accessorAfter],
        [checker, found.setterFinder, setterBefore, setterAfter],
    ];
    const adapted = texts.map(([unit, node, , after]) => node.getText(unit.source) === snapshot(unit, after));
    if (adapted.every(Boolean)) {
        assert.equal(found.third.typeName.text, 'Expression');
        return { files: 0, edits: 0 };
    }
    assert(!adapted.some(Boolean), 'Some define-property sites are adapted and some are not');
    for (const [unit, node, before] of texts) {
        assert.equal(node.getText(unit.source), snapshot(unit, before), `Unreviewed ${unit.relative} site`);
    }
    assert.equal(found.third.typeName.text, 'ObjectLiteralExpression');

    // [unit, start, end, replacement]: the retype, then a check at each of the four reads that needs an object literal.
    const edits = [
        [types, found.third.getStart(types.source), found.third.end, 'Expression'],
        [utilities, found.expandoRead.getStart(utilities.source), found.expandoRead.getStart(utilities.source), expandoCheck],
        [utilities, found.expandoRead.end, found.expandoRead.end, ' : undefined'],
        ...found.accessorReads.map(read => [binder, read.getStart(binder.source), read.getStart(binder.source), accessorCheck]),
        [checker, found.setterRead.getStart(checker.source), found.setterRead.getStart(checker.source), setterCheck],
        [checker, found.setterRead.end, found.setterRead.end, ' : undefined'],
    ];
    // Every file is edited and checked before any is written, so a refusal leaves the tree as it was.
    const outputs = [];
    for (const [unit, node, , after] of texts) {
        let output = unit.text;
        for (const [, start, end, replacement] of edits.filter(edit => edit[0] === unit).sort((a, b) => b[1] - a[1])) {
            output = output.slice(0, start) + replacement + output.slice(end);
        }
        const written = ts.createSourceFile(unit.file, output, ts.ScriptTarget.Latest, true);
        assert.equal(written.parseDiagnostics.length, 0, `Adapted ${unit.relative} must parse`);
        const start = node.getStart(unit.source);
        assert.equal(output.slice(start, start + snapshot(unit, after).length), snapshot(unit, after), `Adapted ${unit.relative} site`);
        outputs.push([unit.file, output]);
    }
    for (const [file, output] of outputs) fs.writeFileSync(file, output);
    return { files: 4, edits: 5, javascriptChanging: true };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
