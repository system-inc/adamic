'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

function one(nodes, matches, description) {
    const found = nodes.filter(matches);
    assert.equal(found.length, 1, `expected one ${description}`);
    return found[0];
}

function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
    return source;
}

function namedFunction(node) {
    for (let parent = node.parent; parent; parent = parent.parent) {
        if (ts.isFunctionLike(parent) && parent.name) return parent;
    }
}

const auditedReads = [
    ['getIntrinsicTagSymbol', "                    links.jsxFlags |= JsxFlags.IntrinsicNamedElement;"],
    ['getIntrinsicTagSymbol', "                    links.jsxFlags |= JsxFlags.IntrinsicIndexedElement;"],
    ['getIntrinsicTagSymbol', "                    links.jsxFlags |= JsxFlags.IntrinsicIndexedElement;"],
    ['getIntrinsicAttributesTypeFromJsxOpeningLikeElement', "            if (links.jsxFlags & JsxFlags.IntrinsicNamedElement) {"],
    ['getIntrinsicAttributesTypeFromJsxOpeningLikeElement', "            else if (links.jsxFlags & JsxFlags.IntrinsicIndexedElement) {"],
    ['calculateNodeCheckFlagWorker', "        if (links.calculatedFlags & flag) {"],
    ['checkSuperExpressions', "            if (links.calculatedFlags & flag) return \"skip\";"],
    ['checkSuperExpressions', "            links.calculatedFlags |= NodeCheckFlags.MethodWithSuperPropertyAccessInAsync | NodeCheckFlags.MethodWithSuperPropertyAssignmentInAsync | NodeCheckFlags.ContainsSuperPropertyInStaticInitializer;"],
    ['checkSingleSuperExpression', "            nodeLinks.calculatedFlags |= NodeCheckFlags.SuperInstance | NodeCheckFlags.SuperStatic; // Yes, we set this on non-applicable nodes, so we can entirely skip the traversal on future calls"],
    ['checkIdentifiers', "            if (links.calculatedFlags & flag) return \"skip\";"],
    ['checkIdentifiers', "            links.calculatedFlags |= NodeCheckFlags.CaptureArguments | NodeCheckFlags.ContainsCapturedBlockScopeBinding | NodeCheckFlags.NeedsLoopOutParameter | NodeCheckFlags.ContainsConstructorReference;"],
    ['checkSingleIdentifier', "            nodeLinks.calculatedFlags |= NodeCheckFlags.ConstructorReference;"],
    ['checkSingleIdentifier', "                nodeLinks.calculatedFlags |= NodeCheckFlags.BlockScopedBindingInLoop | NodeCheckFlags.CapturedBlockScopedBinding; // Can't set on all arbitrary nodes (these nodes have this flag set by `checkSingleBlockScopeBinding` only)"],
    ['checkBlockScopeBindings', "            if (links.calculatedFlags & flag) return \"skip\";"],
    ['checkBlockScopeBindings', "            links.calculatedFlags |= NodeCheckFlags.LoopWithCapturedBlockScopedBinding | NodeCheckFlags.BlockScopedBindingInLoop | NodeCheckFlags.CapturedBlockScopedBinding;"],
];

function verifyFieldTypes(tree) {
    const file = path.join(tree, 'src/compiler/types.ts');
    const source = parse(file, fs.readFileSync(file, 'utf8'));
    const shape = one(source.statements, node => ts.isInterfaceDeclaration(node) &&
        node.name.text === 'NodeLinks', 'exported NodeLinks shape');
    for (const [field, type] of [['flags', 'NodeCheckFlags'], ['calculatedFlags', 'NodeCheckFlags'], ['jsxFlags', 'JsxFlags']]) {
        const member = one(shape.members, node => ts.isPropertySignature(node) && node.name.getText(source) === field, field);
        assert.equal(member.getText(source), `${field}: ${type};`);
    }
    for (const name of ['NodeCheckFlags', 'JsxFlags']) {
        const declaration = one(source.statements, node => ts.isEnumDeclaration(node) && node.name.text === name, name);
        const none = one(declaration.members, node => node.name.getText(source) === 'None', `${name}.None`);
        assert.equal(none.initializer?.getText(source), '0');
    }
    return new Set(shape.members.map(member => member.name.getText(source)));
}

const flagNames = new Set(['calculatedFlags', 'jsxFlags']);
const trackedNames = new Set([...flagNames, 'SymbolLinks', 'SymbolLinksShape', 'NodeLinks', 'NodeLinksShape']);

function inventory(source) {
    const identifiers = new Map();
    const reads = [];
    const stringKeys = [];
    function visit(node) {
        if (ts.isIdentifier(node) && trackedNames.has(node.text)) {
            if (!identifiers.has(node.text)) identifiers.set(node.text, []);
            identifiers.get(node.text).push(node);
            if (flagNames.has(node.text) && ts.isPropertyAccessExpression(node.parent) && node.parent.name === node) {
                reads.push(node);
            }
        }
        if (ts.isStringLiteralLike(node) && flagNames.has(node.text)) stringKeys.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    return { identifiers, reads, stringKeys };
}

function flagRead(source, lines, node) {
    const line = source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
    return { line, text: lines[line - 1], owner: namedFunction(node)?.name.getText(source) };
}

function verifyAuditedReads(source, text, reads) {
    const lines = text.split(/\r?\n/);
    for (const [index, [owner, written]] of auditedReads.entries()) {
        assert(reads[index], `missing audited flag access in ${owner}: ${written.trim()}`);
        const read = flagRead(source, lines, reads[index]);
        assert.deepEqual([read.owner, read.text], [owner, written],
            `unaudited flag access at checker.ts:${read.line}`);
    }
}

// Close the field spellings and constructor references here. The record-use
// walk separately follows lexical bindings and checks whole-record escapes.
function verifyClosedSet(source, seen, nodeLinksClass, adapted) {
    const allowedReads = new Set(seen.reads.slice(0, auditedReads.length));
    const locate = node => source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
    for (const name of flagNames) {
        for (const node of seen.identifiers.get(name) || []) {
            const member = node.parent;
            assert(allowedReads.has(node) || (nodeLinksClass && ts.isPropertyDeclaration(member) &&
                member.parent === nodeLinksClass && member.name === node),
            `unaudited flag access at checker.ts:${locate(node)}: ${name}`);
        }
    }
    for (const node of seen.stringKeys) {
        assert.fail(`unaudited string flag key at checker.ts:${locate(node)}: ${node.text}`);
    }

    // Also close the constructor uses, including references outside the edited owners.
    for (const [name, count] of adapted
        ? [['SymbolLinks', 5], ['SymbolLinksShape', 5], ['NodeLinks', 4], ['NodeLinksShape', 5]]
        : [['SymbolLinks', 8], ['SymbolLinksShape', 0], ['NodeLinks', 7], ['NodeLinksShape', 0]]) {
        assert.equal(seen.identifiers.get(name)?.length || 0, count, `unreviewed ${name} reference`);
    }
}

function verifyRecordUses(source, adapted, fields) {
    // Bind this file only. Symbol identity distinguishes unrelated locals named links.
    const options = { noResolve: true, noLib: true, types: [] };
    const host = ts.createCompilerHost(options);
    host.getSourceFile = name => name === source.fileName ? source : undefined;
    const checker = ts.createProgram([source.fileName], options, host).getTypeChecker();
    const names = new Map();
    function visit(node) {
        if (ts.isIdentifier(node)) {
            if (!names.has(node.text)) names.set(node.text, []);
            names.get(node.text).push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);

    const checkerBody = one(source.statements, node => ts.isFunctionDeclaration(node) &&
        node.name?.text === 'createTypeChecker', 'createTypeChecker').body;
    const functionNamed = name => one(checkerBody.statements, node =>
        ts.isFunctionDeclaration(node) && node.name?.text === name, name);
    const getter = functionNamed('getNodeLinks');
    const importer = functionNamed('resolveImportSymbolType');
    const cacheOwner = functionNamed('runWithoutResolvedSignatureCaching');
    const symbols = new Map();
    function references(binding) {
        const symbol = checker.getSymbolAtLocation(binding);
        assert(symbol, `missing binding for ${binding.text}`);
        if (!symbols.has(symbol)) {
            symbols.set(symbol, (names.get(binding.text) || []).filter(node =>
                (ts.isShorthandPropertyAssignment(node.parent) && node.parent.name === node
                    ? checker.getShorthandAssignmentValueSymbol(node.parent)
                    : checker.getSymbolAtLocation(node)) === symbol));
        }
        return symbols.get(symbol).filter(node => node !== binding && !ts.isTypeQueryNode(node.parent));
    }
    const lines = source.text.split(/\r?\n/);
    const lineOf = node => source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
    const written = node => lines[lineOf(node) - 1].trim();
    function refuse(node, role) {
        assert.fail(`unaudited NodeLinks ${role} at checker.ts:${lineOf(node)} (${namedFunction(node)?.name.getText(source)})`);
    }

    const storage = one(names.get('nodeLinks'), node => ts.isVariableDeclaration(node.parent) &&
        node.parent.name === node && node.parent.parent.parent.parent === checkerBody, 'NodeLinks storage');
    assert.equal(storage.parent.initializer?.getText(source), '[]', 'NodeLinks storage must be a fresh private array');
    const flagsGetter = functionNamed('getNodeCheckFlags');
    const storageUses = references(storage);
    const bounds = storageUses.filter(node => namedFunction(node) === flagsGetter);
    assert.deepEqual(bounds.map(written), [
        'if (nodeId < 0 || nodeId >= nodeLinks.length) return 0;',
        'return nodeLinks[nodeId]?.flags || 0;',
    ]);
    assert(ts.isPropertyAccessExpression(bounds[0].parent) &&
        bounds[0].parent.expression === bounds[0] && bounds[0].parent.name.text === 'length');
    const slots = storageUses.filter(node => !bounds.includes(node)).map(node => {
        const slot = node.parent;
        if (!ts.isElementAccessExpression(slot) || slot.expression !== node ||
            slot.getText(source) !== 'nodeLinks[nodeId]' || namedFunction(slot) !== getter) refuse(node, 'storage use');
        return slot;
    });
    assert.equal(slots.length, 2, 'reviewed NodeLinks storage accesses');
    const assignment = slots[1].parent;
    assert(ts.isBinaryExpression(assignment) && assignment.left === slots[1]);
    assert.equal(assignment.getText(source), adapted
        ? 'nodeLinks[nodeId] = new NodeLinks()' : 'nodeLinks[nodeId] = new (NodeLinks as any)()');

    // Pin the existing tuple cache and import-helper transfers. Track the record
    // again at the tuple's destructuring binding and in the helper's parameter.
    const cache = one(names.get('cachedResolvedSignatures'), node => ts.isVariableDeclaration(node.parent) &&
        node.parent.name === node && namedFunction(node) === cacheOwner, 'signature cache');
    assert.equal(written(cache), 'const cachedResolvedSignatures = [];');
    const [push, loop] = references(cache);
    assert.equal(references(cache).length, 2);
    assert.equal(written(push), 'cachedResolvedSignatures.push([nodeLinks, nodeLinks.resolvedSignature] as const);');
    assert.equal(written(loop), 'for (const [nodeLinks, resolvedSignature] of cachedResolvedSignatures) {');
    assert(ts.isForOfStatement(loop.parent) && loop.parent.expression === loop);
    const [tuple] = loop.parent.initializer.declarations;
    assert(ts.isArrayBindingPattern(tuple.name));
    const [cachedLinks] = tuple.name.elements;
    const cacheCall = push.parent.parent;
    assert(ts.isCallExpression(cacheCall));
    const tupleArgument = cacheCall.arguments[0];
    assert(ts.isAsExpression(tupleArgument) && ts.isArrayLiteralExpression(tupleArgument.expression));
    const cachedValue = tupleArgument.expression.elements[0];

    const helperCalls = references(importer.name);
    assert.deepEqual(helperCalls.map(node => [namedFunction(node)?.name.getText(source), written(node)]), [
        ['getTypeFromImportTypeNode', 'links.resolvedType = resolveImportSymbolType(node, links, currentNamespace, targetMeaning);'],
        ['getTypeFromImportTypeNode', 'links.resolvedType = resolveImportSymbolType(node, links, moduleSymbol, targetMeaning);'],
    ]);
    for (const node of helperCalls) assert(ts.isCallExpression(node.parent) && node.parent.expression === node);
    const helperArguments = new Set(helperCalls.map(node => node.parent.arguments[1]));
    const parameter = importer.parameters[1];
    assert(ts.isIdentifier(parameter.name) && parameter.name.text === 'links');
    assert.equal(parameter.type.getText(source), adapted ? 'NodeLinksShape' : 'NodeLinks');

    const values = [slots[0], assignment, bounds[1].parent, ...references(parameter.name), ...references(cachedLinks.name)];
    for (const node of references(getter.name)) {
        if (!ts.isCallExpression(node.parent) || node.parent.expression !== node) refuse(node, 'accessor escape');
        values.push(node.parent);
    }
    const visited = new Set();
    for (const value of values) {
        if (visited.has(value)) continue;
        visited.add(value);
        const parent = value.parent;
        if (ts.isPropertyAccessExpression(parent) && parent.expression === value) {
            if (!fields.has(parent.name.text)) refuse(parent, 'unknown field');
            let callee = parent;
            while (ts.isParenthesizedExpression(callee.parent) || ts.isAsExpression(callee.parent) ||
                ts.isNonNullExpression(callee.parent)) callee = callee.parent;
            if ((ts.isCallExpression(callee.parent) && callee.parent.expression === callee) ||
                (ts.isTaggedTemplateExpression(callee.parent) && callee.parent.tag === callee)) refuse(parent, 'method receiver');
            continue;
        }
        if (ts.isElementAccessExpression(parent) && parent.expression === value) {
            refuse(parent, 'computed access');
            continue;
        }
        if (ts.isVariableDeclaration(parent) && parent.initializer === value) {
            if (ts.isIdentifier(parent.name)) {
                values.push(...references(parent.name));
                continue;
            }
            if (parent.name.getText(source) === '{ resolvedSymbol }' &&
                written(parent) === 'const { resolvedSymbol } = getNodeLinks(errorTarget);' &&
                namedFunction(parent)?.name.getText(source) === 'invocationErrorDetails') continue;
        }
        if (ts.isParenthesizedExpression(parent) || ts.isAsExpression(parent) || ts.isNonNullExpression(parent) ||
            (ts.isBinaryExpression(parent) && [ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken].includes(parent.operatorToken.kind)) ||
            (ts.isConditionalExpression(parent) && parent.condition !== value)) {
            values.push(parent);
            continue;
        }
        if ((ts.isReturnStatement(parent) && namedFunction(parent) === getter) ||
            (ts.isIfStatement(parent) && parent.expression === value) ||
            (ts.isPrefixUnaryExpression(parent) && parent.operator === ts.SyntaxKind.ExclamationToken) ||
            value === cachedValue || helperArguments.has(value)) continue;
        refuse(parent, ts.SyntaxKind[parent.kind]);
    }
}

function importsFor(source, seen, name) {
    const specifier = one(seen.identifiers.get(name) || [], node => ts.isImportSpecifier(node.parent),
        `${name} import`).parent;
    const importedFrom = specifier.parent.parent.parent;
    assert(ts.isImportDeclaration(importedFrom));
    assert.equal(importedFrom.moduleSpecifier.getText(source), '"./_namespaces/ts.js"');
    const written = specifier.getText(source);
    assert(written === `type ${name}` || written === `type ${name} as ${name}Shape`, `unreviewed ${name} import`);
    return specifier;
}

const links = [
    {
        name: 'SymbolLinks',
        owners: ['symbolLinks', 'getSymbolLinks', 'aliasDeclarationLinks'],
        allocationOwners: ['createSymbol', 'getSymbolLinks'],
        upstream: [
            'const SymbolLinks = class implements SymbolLinks {',
            '    declare _symbolLinksBrand: undefined;',
            '};',
        ],
        replacement: [
            'class SymbolLinks implements SymbolLinksShape {}',
            '',
            'interface SymbolLinks { _symbolLinksBrand: undefined; }',
        ],
    },
    {
        name: 'NodeLinks',
        owners: ['nodeLinks', 'getNodeLinks', 'links'],
        allocationOwners: ['getNodeLinks'],
        upstream: [
            'function NodeLinks(this: NodeLinks) {',
            '    this.flags = NodeCheckFlags.None;',
            '}',
        ],
        replacement: [
            'class NodeLinks implements NodeLinksShape { flags = NodeCheckFlags.None; calculatedFlags = NodeCheckFlags.None; jsxFlags = JsxFlags.None; }',
            '',
            'interface NodeLinks { _nodeLinksBrand: undefined; }',
        ],
    },
];

function typeOwner(node) {
    const parent = node.parent.parent;
    return (ts.isArrayTypeNode(parent) ? parent.parent : parent).name?.getText();
}

function plan(source, seen, specifier, definition, adapted, edits) {
    const { name, owners, allocationOwners, upstream, replacement } = definition;
    const shape = `${name}Shape`;
    const isSymbol = name === 'SymbolLinks';
    const uses = seen.identifiers.get(name);
    const classNode = adapted ? one(source.statements, node =>
        ts.isClassDeclaration(node) && node.name?.text === name, `${name} class`) : undefined;
    const binding = adapted ? classNode : one(uses, node => isSymbol
        ? ts.isVariableDeclaration(node.parent) && node.parent.name === node
        : ts.isFunctionDeclaration(node.parent) && node.parent.name === node, `${name} declaration`).parent;
    const statement = !adapted && isSymbol ? binding.parent.parent : binding;
    assert.equal(statement.parent, source);
    if (!adapted && isSymbol) assert(ts.isClassExpression(binding.initializer));

    const end = adapted ? one(source.statements, node => ts.isInterfaceDeclaration(node) &&
        node.name.text === name, `${name} brand interface`).end : statement.end;
    const start = statement.getStart(source);
    const text = source.text.slice(start, end);
    const eol = text.includes('\r\n') ? '\r\n' : '\n';
    assert.equal(text, (adapted ? replacement : upstream).join(eol), `unreviewed ${name} declaration`);
    const allocationText = !adapted && !isSymbol ? 'new (NodeLinks as any)()' : `new ${name}()`;
    const allocations = uses.flatMap(node => {
        if (ts.isNewExpression(node.parent) && node.parent.expression === node) return [node.parent];
        if (ts.isAsExpression(node.parent) && ts.isParenthesizedExpression(node.parent.parent) &&
            ts.isNewExpression(node.parent.parent.parent)) return [node.parent.parent.parent];
        return [];
    });
    assert.deepEqual(allocations.map(node => node.getText(source)), allocationOwners.map(() => allocationText));
    const checker = one(source.statements, node => ts.isFunctionDeclaration(node) &&
        node.name?.text === 'createTypeChecker', 'createTypeChecker');
    assert(statement.end < checker.getStart(source), `${name} must initialize before createTypeChecker`);
    assert.deepEqual(allocations.map(node => {
        const owner = namedFunction(node);
        assert(owner && owner.parent === checker.body, `${name} allocation outside its reviewed scope`);
        return owner.name.getText(source);
    }), allocationOwners);

    const types = (seen.identifiers.get(adapted ? shape : name) || []).filter(node => {
        if (!ts.isTypeReferenceNode(node.parent) || node.parent.typeName !== node) return false;
        const owner = typeOwner(node);
        return owners.includes(owner) &&
            (owner !== 'links' || namedFunction(node)?.name.getText(source) === 'resolveImportSymbolType');
    });
    assert.deepEqual(types.map(typeOwner), owners, `reviewed ${name} type uses`);
    for (const type of types) assert.equal(type.getText(source), adapted ? shape : name);
    if (!adapted) {
        edits.push({ start, end, after: replacement.join(eol) });
        edits.push({ start: specifier.getStart(source), end: specifier.end, after: `type ${name} as ${shape}` });
        for (const node of types) edits.push({ start: node.getStart(source), end: node.end, after: shape });
        if (!isSymbol) edits.push({ start: allocations[0].getStart(source), end: allocations[0].end, after: 'new NodeLinks()' });
    }
    return classNode;
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const file = path.join(tree, 'src/compiler/checker.ts');
    const text = fs.readFileSync(file, 'utf8');
    const source = parse(file, text);
    const fields = verifyFieldTypes(tree);
    const seen = inventory(source);
    verifyAuditedReads(source, text, seen.reads);
    const specifiers = links.map(definition => importsFor(source, seen, definition.name));
    const adapted = !!specifiers[0].propertyName;
    assert(specifiers.every(specifier => !!specifier.propertyName === adapted), 'partially adapted imports');
    const edits = [];
    const classes = links.map((definition, index) => plan(source, seen, specifiers[index], definition, adapted, edits));
    verifyClosedSet(source, seen, classes[1], adapted);
    verifyRecordUses(source, adapted, fields);
    if (adapted) return { files: 0, edits: 0, javascriptChanging: true };
    assert.equal(edits.length, 11);

    let output = text;
    for (const { start, end, after } of edits.sort((a, b) => b.start - a.start)) {
        output = output.slice(0, start) + after + output.slice(end);
    }
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'checker.ts changed during planning');
    fs.writeFileSync(file, output);
    return { files: 1, edits: edits.length, javascriptChanging: true };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
