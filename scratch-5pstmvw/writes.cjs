#!/usr/bin/env node
'use strict';
// writes.cjs: a scratch measurement for #5pstmvw, never to land. Adaptation 74 makes a flow node's node and antecedent
// readonly on FlowNodeBase and every arm of FlowNode but FlowLabel's antecedent; this finds, through the type checker and
// never by text, every write to a property named node or antecedent whose receiver is a flow node, so flow.sh can say
// whether the before tree writes node once (bindContainer's FlowStart) and antecedent only through a FlowLabel, and
// whether the after tree writes node at all.
//
//   node writes.cjs <tree> <output.json>
//
// One ts.createProgram over all upstream src of <tree>, with adaptation 70's options (the program flow.sh's part (2)
// builds), and stock typescript@6.0.3 from NODE_PATH. A write is an assignment of any kind (=, compound, logical), ++ or
// --, delete, a for-in or for-of target, or a destructuring assignment's target, to a property access named node or
// antecedent or an element access whose key is that literal (or whose key's type is). The receiver is a flow node when
// its type, or an arm of it (union or intersection, a type parameter through its constraint), is one of the flow
// interfaces declared in src/compiler/types.ts or an interface extending one. Also recorded: a write on a flow receiver
// whose key isn't a literal, Object.assign and the other property roads (defineProperty, defineProperties,
// setPrototypeOf, Reflect.set, Reflect.defineProperty, Reflect.deleteProperty) whose first argument is a flow node, and
// writes to node or antecedent through an any receiver, which the checker can't place. The whole result goes to
// <output.json>; stdout gets one line of counts.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 4) throw new Error('usage: node writes.cjs <tree> <output.json>');
const tree = path.resolve(process.argv[2]);
const output = path.resolve(process.argv[3]);
const sourceRoot = path.join(tree, 'src') + path.sep;
const typesFile = path.join(tree, 'src/compiler/types.ts');

// The interfaces adaptation 74's flowSlots names, and the slots it makes readonly.
const flowNames = new Set(['FlowNodeBase', 'FlowUnreachable', 'FlowStart', 'FlowLabel', 'FlowAssignment', 'FlowCall',
    'FlowCondition', 'FlowSwitchClause', 'FlowArrayMutation', 'FlowReduceLabel']);
const slotNames = new Set(['node', 'antecedent']);

// Adaptation 70's options, so the program is the one its audit checks and flow.sh's part (2) compares.
const options = {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: false,
    allowImportingTsExtensions: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
};
const roots = ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']);
const program = ts.createProgram(roots, options);
const checker = program.getTypeChecker();
const typesSource = program.getSourceFile(typesFile);
if (!typesSource) throw new Error(`the program holds no ${typesFile}`);
// Every assignment operator, =, the compound ones and the logical ones, lies between these two in SyntaxKind.
if (typeof ts.SyntaxKind.FirstAssignment !== 'number' || typeof ts.SyntaxKind.LastAssignment !== 'number')
    throw new Error('this TypeScript has no SyntaxKind.FirstAssignment or LastAssignment');

// The flow interfaces as declared, and FlowNode's arms, so a flow arm is a declaration in types.ts, never a name alone.
const declared = new Set();
let flowNodeAlias;
ts.forEachChild(typesSource, function find(node) {
    if (ts.isInterfaceDeclaration(node) && flowNames.has(node.name.text)) declared.add(node.name.text);
    if (ts.isTypeAliasDeclaration(node) && node.name.text === 'FlowNode') flowNodeAlias = node;
    ts.forEachChild(node, find);
});
const undeclared = [...flowNames].filter(name => !declared.has(name));
if (undeclared.length) throw new Error(`types.ts declares no interface ${undeclared.join(', ')}`);
if (!flowNodeAlias) throw new Error('types.ts declares no type FlowNode');

function where(node) {
    const source = node.getSourceFile();
    const at = source.getLineAndCharacterOfPosition(node.getStart(source));
    return { file: path.relative(tree, path.resolve(source.fileName)), line: at.line + 1, column: at.character + 1 };
}

function flat(node) {
    const text = node.getText().replace(/\s+/g, ' ');
    return text.length <= 200 ? text : `${text.slice(0, 200)} ...`;
}

// flowName is the flow interface a symbol declares in types.ts, if it declares one.
function flowName(symbol) {
    const declaration = symbol?.declarations?.find(declaration =>
        ts.isInterfaceDeclaration(declaration) && declaration.getSourceFile() === typesSource && flowNames.has(declaration.name.text));
    return declaration?.name.text;
}

// flowArms is every flow interface a type is or holds as an arm: its own name, or for an interface extending one,
// "Name (extends FlowLabel)".
function flowArms(type, seen = new Set()) {
    if (seen.has(type)) return [];
    seen.add(type);
    if (type.flags & ts.TypeFlags.TypeParameter) {
        const constraint = checker.getBaseConstraintOfType(type);
        return constraint ? flowArms(constraint, seen) : [];
    }
    if (type.isUnionOrIntersection()) return type.types.flatMap(arm => flowArms(arm, seen));
    const own = flowName(type.getSymbol());
    if (own) return [own];
    if (type.isClassOrInterface()) {
        for (const base of checker.getBaseTypes(type)) {
            const found = flowArms(base, seen);
            if (found.length) return [`${checker.typeToString(type)} (extends ${found[0].replace(/ \(extends .*\)$/, '')})`];
        }
    }
    return [];
}

function unique(values) {
    return [...new Set(values)];
}

const flowNodeArms = unique(flowArms(checker.getTypeAtLocation(flowNodeAlias.name))).sort();

// destructured: the target stands inside an object or array literal that is a destructuring assignment's left side or
// a for-in or for-of initializer, at any depth.
function destructured(target) {
    let current = target;
    let inLiteral = false;
    for (;;) {
        const parent = current.parent;
        if (!parent) return false;
        if (ts.isParenthesizedExpression(parent)) current = parent;
        else if (ts.isPropertyAssignment(parent) && parent.initializer === current) {
            current = parent.parent;
            inLiteral = true;
        }
        else if ((ts.isSpreadAssignment(parent) || ts.isSpreadElement(parent)) && parent.expression === current) {
            current = parent.parent;
            inLiteral = true;
        }
        else if (ts.isArrayLiteralExpression(parent)) {
            current = parent;
            inLiteral = true;
        }
        else {
            if (!inLiteral) return false;
            if (ts.isBinaryExpression(parent) && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken && parent.left === current) return true;
            return (ts.isForOfStatement(parent) || ts.isForInStatement(parent)) && parent.initializer === current;
        }
    }
}

// writeKind says how an access is written, or undefined when it is only read. A target may stand inside parentheses,
// a non-null assertion or a type assertion: (flow.node as X) = value writes flow.node.
function writeKind(access) {
    let current = access;
    while (current.parent && (ts.isParenthesizedExpression(current.parent) || ts.isNonNullExpression(current.parent) ||
        ts.isAsExpression(current.parent) || ts.isSatisfiesExpression(current.parent) || ts.isTypeAssertionExpression(current.parent)) &&
        current.parent.expression === current)
        current = current.parent;
    const parent = current.parent;
    if (!parent) return undefined;
    if (ts.isBinaryExpression(parent) && parent.left === current &&
        parent.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && parent.operatorToken.kind <= ts.SyntaxKind.LastAssignment)
        return { how: ts.tokenToString(parent.operatorToken.kind), statement: parent };
    const counts = [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken];
    if (ts.isPrefixUnaryExpression(parent) && counts.includes(parent.operator)) return { how: `prefix ${ts.tokenToString(parent.operator)}`, statement: parent };
    if (ts.isPostfixUnaryExpression(parent) && counts.includes(parent.operator)) return { how: `postfix ${ts.tokenToString(parent.operator)}`, statement: parent };
    if (ts.isDeleteExpression(parent)) return { how: 'delete', statement: parent };
    if (ts.isForOfStatement(parent) && parent.initializer === current) return { how: 'for-of target', statement: parent.initializer };
    if (ts.isForInStatement(parent) && parent.initializer === current) return { how: 'for-in target', statement: parent.initializer };
    if (destructured(current)) return { how: 'destructuring target', statement: current };
    return undefined;
}

// keyNames is the slot names an element access's key can be: its literal text, its type's string literals, or
// undefined when the key's type is wider than literals.
function keyNames(key) {
    if (ts.isStringLiteralLike(key)) return [key.text];
    const type = checker.getTypeAtLocation(key);
    const arms = type.isUnion() ? type.types : [type];
    if (!arms.every(arm => arm.isStringLiteral() || arm.isNumberLiteral())) return undefined;
    return arms.filter(arm => arm.isStringLiteral()).map(arm => arm.value);
}

const writes = [];
const computed = [];
const anyReceivers = [];
const roads = [];
let accesses = 0;
let objectAssignCalls = 0;
const roadNames = { Object: ['assign', 'defineProperty', 'defineProperties', 'setPrototypeOf'], Reflect: ['set', 'defineProperty', 'deleteProperty'] };

function visit(node) {
    if (ts.isPropertyAccessExpression(node) && slotNames.has(node.name.text)) {
        accesses++;
        const written = writeKind(node);
        if (written) {
            const receiver = checker.getTypeAtLocation(node.expression);
            const arms = unique(flowArms(receiver));
            const row = { ...where(node), slot: node.name.text, how: written.how, access: 'property', receiver: checker.typeToString(receiver), arms, text: flat(written.statement) };
            if (arms.length) writes.push(row);
            else if (receiver.flags & ts.TypeFlags.Any) anyReceivers.push(row);
        }
    }
    if (ts.isElementAccessExpression(node)) {
        const written = writeKind(node);
        if (written) {
            const receiver = checker.getTypeAtLocation(node.expression);
            const arms = unique(flowArms(receiver));
            const names = keyNames(node.argumentExpression);
            const row = { ...where(node), how: written.how, access: 'element', receiver: checker.typeToString(receiver), arms, text: flat(written.statement) };
            if (names === undefined) {
                if (arms.length) computed.push({ ...row, key: checker.typeToString(checker.getTypeAtLocation(node.argumentExpression)) });
            }
            else {
                for (const slot of names.filter(name => slotNames.has(name))) {
                    accesses++;
                    if (arms.length) writes.push({ ...row, slot });
                    else if (receiver.flags & ts.TypeFlags.Any) anyReceivers.push({ ...row, slot });
                }
            }
        }
    }
    if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && ts.isIdentifier(node.expression.expression) &&
        Object.hasOwn(roadNames, node.expression.expression.text) && roadNames[node.expression.expression.text].includes(node.expression.name.text)) {
        const road = `${node.expression.expression.text}.${node.expression.name.text}`;
        if (road === 'Object.assign') objectAssignCalls++;
        const first = node.arguments[0];
        if (first) {
            const target = checker.getTypeAtLocation(first);
            const arms = unique(flowArms(target));
            if (arms.length) roads.push({ ...where(node), road, target: checker.typeToString(target), arms, text: flat(node) });
        }
    }
    ts.forEachChild(node, visit);
}

let files = 0;
for (const source of program.getSourceFiles()) {
    if (!path.resolve(source.fileName).startsWith(sourceRoot)) continue;
    files++;
    visit(source);
}
const order = (a, b) => a.file.localeCompare(b.file) || a.line - b.line || a.column - b.column;
for (const rows of [writes, computed, anyReceivers, roads]) rows.sort(order);
fs.writeFileSync(output, JSON.stringify({ typescript: ts.version, tree, roots: roots.length, files, flowNodeArms, accesses,
    objectAssignCalls, writes, computed, anyReceivers, roads }, null, 2) + '\n');
console.log(`${roots.length} roots, ${files} src files walked, ${accesses} node or antecedent accesses, ` +
    `${writes.filter(row => row.slot === 'node').length} node and ${writes.filter(row => row.slot === 'antecedent').length} antecedent writes on a flow receiver, ` +
    `${computed.length} writes on a flow receiver with a key that isn't a literal, ${roads.length} property roads onto a flow node, ` +
    `${anyReceivers.length} writes through an any receiver`);
