#!/usr/bin/env node
'use strict';
// w2scout.cjs: a scratch measurement, never to land. tsc writes AST nodes through a Mutable<X> view, a mapped type that
// lifts readonly from every field, so a compiler that refuses unproven writes through Mutable needs to know, field by
// field, what tsc actually writes through those views. This reads a list of sites where a value is seen as a Mutable
// view, finds the writes each view carries, and says for each write whether the receiver is an object its own function
// is still making, and how the written field is declared (which interface, whether readonly, whether public).
//
//   node w2scout.cjs <tree> <rows.tsv> <out.json>
//
// One ts.createProgram over every .ts under <tree>/src but **/lib/**, with stock typescript@6.0.3 from NODE_PATH, and
// a type checker. Each row's where (src/compiler/file.ts:line:column, one based) names a position. The script takes the
// smallest node starting there and climbs, only through nodes starting at that same position or through parentheses,
// casts and non-null assertions, to the expression seen as Mutable: a cast (as or <>) to a Mutable<...> type, or an
// argument whose parameter is declared Mutable<...> or whose contextual type is a Mutable instantiation. That road is
// recorded as cast, argument, or other.
//
// The writes made through the view depend on how the view is used. Directly under a property access, the access is one
// write when it is assigned, compounded, incremented, deleted or destructured into. As a variable's initializer, every
// such write through that variable is counted. As a call argument, the callee is followed one level through the
// checker, and every write through the matching parameter inside the callee is counted; calls the callee passes the
// parameter on to are listed, not followed.
//
// A write is making when the receiver (the expression seen as Mutable, with its casts taken off) is a local declared in
// the same function as the site, initialized by a call whose callee starts with factory.create or whose last name
// starts with create, new or base, or by an object literal, and between that initializer and the write there is no
// assignment to the local, no call passing the local or the view variable as an argument, no return mentioning either,
// and no nested function mentioning either. Anything else is elsewhere, including a fresh create call passed straight
// to a parameter, which is flagged separately as freshArgument.
//
// Each written field is resolved on the source type (the type before the view, through its constraint when generic),
// then on the view's type when the source type has no such property. The declaring interfaces are recorded, whether
// the declarations are readonly, and whether they carry an @internal tag, read with ts.getJSDocTags and from the
// leading comments. A field is public when some declaration is neither tagged internal nor inside an interface tagged
// internal. A row whose node can't be found is recorded as unfound with a reason, and the run goes on.
//
// The result goes to <out.json>. Stdout gets one FIELD line per written field, sorted by writes, then PASS.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

function fail(message) {
    console.error(`w2scout: ${message}`);
    process.exit(1);
}

if (ts.version !== '6.0.3') fail(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 5) fail('usage: node w2scout.cjs <tree> <rows.tsv> <out.json>');
const tree = path.resolve(process.argv[2]);
const rowsFile = path.resolve(process.argv[3]);
const output = path.resolve(process.argv[4]);
const sourceRoot = path.join(tree, 'src');
if (!fs.existsSync(path.join(sourceRoot, 'compiler', 'types.ts'))) fail(`${tree} holds no src/compiler/types.ts`);
if (typeof ts.SyntaxKind.FirstAssignment !== 'number' || typeof ts.SyntaxKind.LastAssignment !== 'number')
    fail('this TypeScript has no SyntaxKind.FirstAssignment or LastAssignment');

// The rows: a header naming the columns, then one site per line. Column names become camelCase keys.
if (!fs.existsSync(rowsFile)) fail(`no such TSV ${rowsFile}`);
const lines = fs.readFileSync(rowsFile, 'utf8').split(/\r?\n/).filter(line => line.trim() !== '');
if (lines.length < 2) fail(`${rowsFile} holds no rows under its header`);
const header = lines[0].split('\t').map(name => name.trim().replace(/[\s-]+(\w)/g, (_, letter) => letter.toUpperCase()));
for (const needed of ['where', 'field']) if (!header.includes(needed)) fail(`${rowsFile}'s header names no ${needed} column`);
const rows = lines.slice(1).map(line => {
    const cells = line.split('\t');
    return Object.fromEntries(header.map((name, index) => [name, (cells[index] ?? '').trim()]));
});

const options = {
    strict: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [], allowImportingTsExtensions: true, moduleDetection: ts.ModuleDetectionKind.Force,
};
const roots = ts.sys.readDirectory(sourceRoot, ['.ts'], ['**/lib/**']);
if (!roots.length) fail(`${sourceRoot} holds no .ts files`);
const program = ts.createProgram(roots, options);
const checker = program.getTypeChecker();

function where(node) {
    const source = node.getSourceFile();
    const at = source.getLineAndCharacterOfPosition(node.getStart(source));
    return `${path.relative(tree, path.resolve(source.fileName)).split(path.sep).join('/')}:${at.line + 1}:${at.character + 1}`;
}

function flat(node, limit = 160) {
    const text = node.getText().replace(/\s+/g, ' ');
    return text.length <= limit ? text : `${text.slice(0, limit - 4)} ...`;
}

function unique(values) {
    return [...new Set(values)];
}

// A wrapper changes how a value is typed or grouped but not which object it is.
function isWrapper(node) {
    return ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) ||
        ts.isNonNullExpression(node) || ts.isSatisfiesExpression(node);
}

function stripWrappers(node) {
    let current = node;
    while (isWrapper(current)) current = current.expression;
    return current;
}

// outermost climbs from a node through the wrappers it stands in, so (m as X).f finds the access above the cast.
function outermost(node) {
    let current = node;
    while (current.parent && isWrapper(current.parent) && current.parent.expression === current) current = current.parent;
    return current;
}

function isFunctionBody(node) {
    return ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node) ||
        ts.isMethodDeclaration(node) || ts.isConstructorDeclaration(node) || ts.isGetAccessorDeclaration(node) ||
        ts.isSetAccessorDeclaration(node);
}

function enclosingFunction(node) {
    return ts.findAncestor(node.parent, isFunctionBody);
}

function functionName(fn) {
    if (!fn) return '<top level>';
    if (ts.isConstructorDeclaration(fn)) {
        const owner = fn.parent;
        return `${owner && owner.name ? owner.name.text : '<class>'}.constructor`;
    }
    if (fn.name && (ts.isIdentifier(fn.name) || ts.isStringLiteral(fn.name) || ts.isPrivateIdentifier(fn.name))) return fn.name.text;
    const parent = fn.parent;
    if (parent && ts.isVariableDeclaration(parent) && ts.isIdentifier(parent.name)) return parent.name.text;
    if (parent && ts.isPropertyAssignment(parent) && ts.isIdentifier(parent.name)) return parent.name.text;
    return '<anonymous>';
}

// isMutableTypeNode: a reference to the type alias Mutable, or a union holding one. The name is checked through the
// checker (an import of Mutable resolves to the alias), and the bare name is trusted only when the checker can't say.
function isMutableTypeNode(node) {
    if (!node) return false;
    if (ts.isParenthesizedTypeNode(node)) return isMutableTypeNode(node.type);
    if (ts.isUnionTypeNode(node)) return node.types.some(isMutableTypeNode);
    if (!ts.isTypeReferenceNode(node)) return false;
    const name = ts.isIdentifier(node.typeName) ? node.typeName : node.typeName.right;
    if (name.text !== 'Mutable') return false;
    let symbol = checker.getSymbolAtLocation(name);
    if (symbol && symbol.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
    if (!symbol || !symbol.declarations || !symbol.declarations.length) return true;
    return symbol.getName() === 'Mutable' && symbol.declarations.some(ts.isTypeAliasDeclaration);
}

function isMutableType(type) {
    if (!type) return false;
    if (type.aliasSymbol && type.aliasSymbol.getName() === 'Mutable') return true;
    return type.isUnion() && type.types.some(isMutableType);
}

function isMutableCast(node) {
    return (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) && isMutableTypeNode(node.type);
}

// implementation finds the declaration with a body behind a signature, so an overload leads to the function itself.
function implementation(declaration) {
    if (!declaration) return undefined;
    if (isFunctionBody(declaration) && declaration.body) return declaration;
    const symbol = declaration.symbol;
    return symbol && symbol.declarations ? symbol.declarations.find(other => isFunctionBody(other) && other.body) : undefined;
}

function calleeKind(name, declaration) {
    if (name === 'update') return 'update';
    if (name.startsWith('finishUpdate')) return 'finishUpdate';
    const file = declaration ? path.resolve(declaration.getSourceFile().fileName) : '';
    if (file.startsWith(path.join(sourceRoot, 'compiler', 'factory') + path.sep)) return 'factory helper';
    return 'other';
}

// calleeOf: the call an argument stands in, the parameter it meets (in the implementation), and whether that
// parameter is a Mutable view, by its declared type in the signature the checker chose or by the argument's contextual
// type.
function calleeOf(argument) {
    const call = argument.parent;
    if (!call || !(ts.isCallExpression(call) || ts.isNewExpression(call)) || !call.arguments) return undefined;
    const index = call.arguments.indexOf(argument);
    if (index < 0) return undefined;
    const signature = checker.getResolvedSignature(call);
    const chosen = signature ? signature.declaration : undefined;
    const body = implementation(chosen);
    const parameterAt = declaration => {
        if (!declaration || !declaration.parameters) return undefined;
        const parameters = declaration.parameters;
        if (index < parameters.length) return parameters[index];
        const last = parameters[parameters.length - 1];
        return last && last.dotDotDotToken ? last : undefined;
    };
    const chosenParameter = parameterAt(chosen);
    const parameter = parameterAt(body);
    const mutable = isMutableTypeNode(chosenParameter && chosenParameter.type) || isMutableTypeNode(parameter && parameter.type) ||
        isMutableType(checker.getContextualType(argument));
    const name = body ? functionName(body) : flat(call.expression, 80);
    return { call, index, body, parameter, mutable, name, kind: calleeKind(name, body) };
}

// deepestAt is the smallest node starting exactly at a position.
function deepestAt(source, position) {
    let found;
    (function descend(node) {
        ts.forEachChild(node, child => {
            const start = child.getStart(source);
            if (start <= position && position < child.getEnd()) {
                if (start === position) found = child;
                descend(child);
                return true;
            }
            return undefined;
        });
    })(source);
    return found;
}

// mutableSite climbs from the node at a row's position to the expression seen as Mutable.
function mutableSite(start, position) {
    let node = start;
    for (;;) {
        if (isMutableCast(node)) return { road: 'cast', expression: node };
        if (ts.isParenthesizedExpression(node) && isMutableCast(stripParentheses(node))) return { road: 'cast', expression: stripParentheses(node) };
        const callee = calleeOf(node);
        if (callee && callee.mutable) return { road: 'argument', expression: node, callee };
        const parent = node.parent;
        if (!parent || ts.isSourceFile(parent)) break;
        if (!(parent.getStart() === position || (isWrapper(parent) && parent.expression === node))) break;
        node = parent;
    }
    // No cast and no Mutable parameter: the value may still meet a Mutable contextual type, as in const m: Mutable<X> = x.
    let widest = start;
    while (widest.parent && !ts.isSourceFile(widest.parent) && widest.parent.getStart() === position) widest = widest.parent;
    for (let node = start; ; node = node.parent) {
        if (ts.isExpression(node) && isMutableType(checker.getContextualType(node))) return { road: 'other', expression: node, contextual: true };
        if (node === widest || !node.parent) break;
    }
    return { road: 'other', expression: widest, contextual: false };
}

function stripParentheses(node) {
    let current = node;
    while (ts.isParenthesizedExpression(current)) current = current.expression;
    return current;
}

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

// writeKind says how a target is written, or undefined when it is only read. A target may stand inside parentheses,
// a non-null assertion or a type assertion.
function writeKind(target) {
    const current = outermost(target);
    const parent = current.parent;
    if (!parent) return undefined;
    if (ts.isBinaryExpression(parent) && parent.left === current &&
        parent.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && parent.operatorToken.kind <= ts.SyntaxKind.LastAssignment)
        return { operator: ts.tokenToString(parent.operatorToken.kind), statement: parent };
    const counts = [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken];
    if (ts.isPrefixUnaryExpression(parent) && counts.includes(parent.operator)) return { operator: `prefix ${ts.tokenToString(parent.operator)}`, statement: parent };
    if (ts.isPostfixUnaryExpression(parent) && counts.includes(parent.operator)) return { operator: `postfix ${ts.tokenToString(parent.operator)}`, statement: parent };
    if (ts.isDeleteExpression(parent)) return { operator: 'delete', statement: parent };
    if (ts.isForOfStatement(parent) && parent.initializer === current) return { operator: 'for-of target', statement: parent };
    if (ts.isForInStatement(parent) && parent.initializer === current) return { operator: 'for-in target', statement: parent };
    if (destructured(current)) return { operator: 'destructuring target', statement: current };
    return undefined;
}

// fieldWrite: when a value (already climbed out of its wrappers) is the receiver of a written property access, the
// field and how it is written. An element access counts when its key is a string literal.
function fieldWrite(receiver) {
    const parent = receiver.parent;
    if (!parent) return undefined;
    let field;
    if (ts.isPropertyAccessExpression(parent) && parent.expression === receiver) field = parent.name.text;
    else if (ts.isElementAccessExpression(parent) && parent.expression === receiver && ts.isStringLiteralLike(parent.argumentExpression))
        field = parent.argumentExpression.text;
    if (field === undefined) return undefined;
    const written = writeKind(parent);
    return written ? { field, access: parent, ...written } : undefined;
}

// references: every identifier under a scope that the checker resolves to one of the symbols.
function references(scope, symbols) {
    const found = [];
    (function visit(node) {
        if (ts.isIdentifier(node) && symbols.has(checker.getSymbolAtLocation(node))) found.push(node);
        ts.forEachChild(node, visit);
    })(scope);
    return found;
}

// throughSymbol: the writes through a variable or parameter, and the calls it is passed on to.
function throughSymbol(symbol, scope) {
    const writes = [];
    const onward = [];
    for (const reference of references(scope, new Set([symbol]))) {
        if (reference.parent && ts.isVariableDeclaration(reference.parent) && reference.parent.name === reference) continue;
        if (reference.parent && ts.isParameter(reference.parent) && reference.parent.name === reference) continue;
        const receiver = outermost(reference);
        const written = fieldWrite(receiver);
        if (written) writes.push(written);
        else {
            const callee = calleeOf(receiver);
            if (callee) onward.push({ callee: callee.name, calleeKind: callee.kind, where: where(callee.call) });
        }
    }
    return { writes, onward };
}

// mentions: whether a subtree names one of the symbols as a value of its own, not only as the object of a property
// or element access (node.parent reads a field; node and cond ? node : other hand the object on).
function mentions(subtree, symbols) {
    let hit = false;
    (function visit(node) {
        if (hit) return;
        if (ts.isIdentifier(node) && symbols.has(checker.getSymbolAtLocation(node))) {
            const receiver = outermost(node);
            const parent = receiver.parent;
            const accessed = parent && (ts.isPropertyAccessExpression(parent) || ts.isElementAccessExpression(parent)) && parent.expression === receiver;
            if (!accessed) hit = true;
            return;
        }
        ts.forEachChild(node, visit);
    })(subtree);
    return hit;
}

function freshInitializer(initializer) {
    const value = stripWrappers(initializer);
    if (ts.isObjectLiteralExpression(value)) return 'object literal';
    if (!ts.isCallExpression(value)) return undefined;
    const text = value.expression.getText().replace(/\s+/g, '');
    const last = ts.isPropertyAccessExpression(value.expression) ? value.expression.name.text
        : ts.isIdentifier(value.expression) ? value.expression.text : '';
    return text.startsWith('factory.create') || /^(create|new|base)/.test(last) ? `call ${text}` : undefined;
}

function describeReceiver(receiver) {
    const value = stripWrappers(receiver);
    if (ts.isIdentifier(value)) {
        const symbol = checker.getSymbolAtLocation(value);
        const declaration = symbol && symbol.valueDeclaration;
        if (declaration && ts.isParameter(declaration)) return `parameter ${value.text}`;
        if (declaration && ts.isVariableDeclaration(declaration))
            return declaration.initializer ? `local ${value.text} = ${flat(declaration.initializer, 80)}` : `local ${value.text}`;
        return `identifier ${value.text}`;
    }
    if (ts.isCallExpression(value)) return `call ${flat(value, 80)}`;
    if (ts.isObjectLiteralExpression(value)) return 'object literal';
    return flat(value, 80);
}

// makingCheck applies the making heuristic to one write: receiver is the expression seen as Mutable, site is where the
// view is taken, end is where the write's own expression ends, views are variables holding the view, and own is the
// call carrying the view to a callee (the road itself, so not a pass).
function makingCheck(receiver, site, end, views, own) {
    const value = stripWrappers(receiver);
    if (!ts.isIdentifier(value)) return { making: 'elsewhere', reason: 'the receiver is not a local' };
    const symbol = checker.getSymbolAtLocation(value);
    const declaration = symbol && symbol.valueDeclaration;
    if (!declaration || !ts.isVariableDeclaration(declaration) || !ts.isIdentifier(declaration.name) || !declaration.initializer)
        return { making: 'elsewhere', reason: 'the receiver is not an initialized local' };
    const fn = enclosingFunction(site);
    if (enclosingFunction(declaration) !== fn) return { making: 'elsewhere', reason: 'the local is declared in another function' };
    const fresh = freshInitializer(declaration.initializer);
    if (!fresh) return { making: 'elsewhere', reason: `the local is initialized by ${flat(declaration.initializer, 60)}` };
    const low = declaration.getEnd();
    if (end <= low) return { making: 'elsewhere', reason: 'the write does not follow the initializer' };
    const watched = new Set([symbol, ...views]);
    const scope = fn ? fn.body : site.getSourceFile();
    let reason;
    (function visit(node) {
        if (reason || node.getEnd() <= low || node.getStart() >= end) return;
        const inside = node.getStart() >= low;
        if (inside && node !== own && (ts.isCallExpression(node) || ts.isNewExpression(node)) &&
            (node.arguments || []).some(argument => mentions(argument, watched))) {
            reason = `${flat(node, 60)} passes it on first`;
            return;
        }
        if (inside && ts.isReturnStatement(node) && node.expression && mentions(node.expression, watched)) {
            reason = `${flat(node, 60)} returns it first`;
            return;
        }
        if (inside && ts.isIdentifier(node) && checker.getSymbolAtLocation(node) === symbol && writeKind(node)) {
            reason = 'the local is reassigned first';
            return;
        }
        if (isFunctionBody(node) && node !== fn) {
            if (inside && mentions(node, watched)) reason = `a nested function at ${where(node)} captures it first`;
            return;
        }
        ts.forEachChild(node, visit);
    })(scope);
    return reason ? { making: 'elsewhere', reason } : { making: 'making', reason: `initialized by ${fresh}` };
}

function hasInternal(node) {
    if (!node) return false;
    if (ts.getJSDocTags(node).some(tag => tag.tagName.text === 'internal')) return true;
    const text = node.getSourceFile().text;
    const ranges = ts.getLeadingCommentRanges(text, node.pos) || [];
    return ranges.some(range => /@internal\b/.test(text.slice(range.pos, range.end)));
}

function containerName(declaration) {
    const container = declaration.parent;
    if (!container) return '<unknown>';
    if ((ts.isInterfaceDeclaration(container) || ts.isClassDeclaration(container)) && container.name) return container.name.text;
    if (ts.isTypeLiteralNode(container) && container.parent && ts.isTypeAliasDeclaration(container.parent)) return container.parent.name.text;
    if (ts.isConstructorDeclaration(container) && container.parent && container.parent.name) return container.parent.name.text;
    return `<${ts.SyntaxKind[container.kind]}>`;
}

function propertyDeclarations(type, field) {
    if (!type) return [];
    const apparent = checker.getApparentType(type);
    const property = checker.getPropertyOfType(apparent, field);
    let declarations = property && property.declarations ? property.declarations : [];
    if (!declarations.length && apparent.isUnionOrIntersection())
        declarations = apparent.types.flatMap(arm => {
            const armProperty = checker.getPropertyOfType(checker.getApparentType(arm), field);
            return armProperty && armProperty.declarations ? armProperty.declarations : [];
        });
    return unique(declarations);
}

// declarationOf resolves a field on the source type, then on the view's type, and reads how it is declared.
function declarationOf(field, sourceType, viewType) {
    let resolvedOn = 'source';
    let declarations = propertyDeclarations(sourceType, field);
    if (!declarations.length) {
        resolvedOn = 'view';
        declarations = propertyDeclarations(viewType, field);
    }
    if (!declarations.length) return { declaredIn: [], readonly: null, public: null, internal: null, resolvedOn: 'none', declaredAt: [] };
    const read = declarations.map(declaration => ({
        container: containerName(declaration),
        readonly: (ts.getCombinedModifierFlags(declaration) & ts.ModifierFlags.Readonly) !== 0,
        internal: hasInternal(declaration),
        containerInternal: hasInternal(declaration.parent),
        at: where(declaration),
    }));
    const readonly = read.every(entry => entry.readonly) ? true : read.some(entry => entry.readonly) ? 'mixed' : false;
    return {
        declaredIn: unique(read.map(entry => entry.container)),
        readonly,
        public: read.some(entry => !entry.internal && !entry.containerInternal),
        internal: read.every(entry => entry.internal) ? true : read.some(entry => entry.internal) ? 'mixed' : false,
        containerInternal: read.some(entry => entry.containerInternal),
        resolvedOn,
        declaredAt: read.map(entry => entry.at),
    };
}

function typeText(type) {
    return type ? checker.typeToString(type, undefined, ts.TypeFormatFlags.NoTruncation) : null;
}

function measure(row) {
    const match = /^(.+):(\d+):(\d+)$/.exec(row.where || '');
    if (!match) return { unfound: `where ${JSON.stringify(row.where)} is not file:line:column` };
    const file = path.join(tree, match[1]);
    const source = program.getSourceFile(file);
    if (!source) return { unfound: `the program holds no ${match[1]}` };
    const line = Number(match[2]) - 1;
    const column = Number(match[3]) - 1;
    const starts = source.getLineStarts();
    if (line < 0 || line >= starts.length) return { unfound: `${match[1]} has no line ${line + 1}` };
    const lineEnd = line + 1 < starts.length ? starts[line + 1] : source.text.length;
    const position = starts[line] + column;
    if (column < 0 || position >= lineEnd) return { unfound: `line ${line + 1} of ${match[1]} has no column ${column + 1}` };
    const start = deepestAt(source, position);
    if (!start) return { unfound: `no node starts at ${row.where}; the line reads ${JSON.stringify(source.text.slice(starts[line], lineEnd).trim().slice(0, 120))}` };

    const site = mutableSite(start, position);
    const expression = site.expression;
    const result = {
        road: site.road,
        expression: flat(expression),
        function: functionName(enclosingFunction(expression)),
        callee: null,
        calleeKind: null,
        use: 'none',
        receiver: null,
        sourceTypeSeen: null,
        viewTypeSeen: null,
        freshArgument: false,
        onward: [],
        writes: [],
        notes: [],
    };
    if (site.road === 'other') result.notes.push(site.contextual ? 'no cast or Mutable parameter here, but the contextual type is Mutable'
        : `no cast to Mutable or Mutable parameter climbing from ${ts.SyntaxKind[start.kind]} ${flat(start, 60)}`);

    // The receiver is the value before the view: a cast's operand, or the argument itself.
    const receiver = site.road === 'cast' ? expression.expression : expression;
    result.receiver = describeReceiver(receiver);
    const sourceType = checker.getTypeAtLocation(stripWrappers(receiver));
    const viewType = checker.getTypeAtLocation(expression);
    result.sourceTypeSeen = typeText(sourceType);
    result.viewTypeSeen = typeText(viewType);

    const record = (written, making, viewTypeHere) => {
        const declared = declarationOf(written.field, sourceType, viewTypeHere);
        result.writes.push({
            field: written.field, operator: written.operator, where: where(written.access),
            making: making.making, makingReason: making.reason, statement: flat(written.statement, 120), ...declared,
        });
    };

    // How the view is used decides where its writes are.
    const used = outermost(expression);
    const direct = fieldWrite(used);
    const parent = used.parent;
    const callee = site.road === 'argument' ? site.callee : calleeOf(used);
    if (direct) {
        result.use = 'property';
        record(direct, makingCheck(receiver, expression, direct.statement.getEnd(), [], undefined), viewType);
    }
    else if (parent && ts.isVariableDeclaration(parent) && parent.initializer === used) {
        result.use = 'variable';
        if (!ts.isIdentifier(parent.name)) result.notes.push('the view initializes a binding pattern, which is not followed');
        else {
            const symbol = checker.getSymbolAtLocation(parent.name);
            const fn = enclosingFunction(parent);
            const through = throughSymbol(symbol, fn ? fn.body : source);
            result.onward = through.onward;
            for (const written of through.writes)
                record(written, makingCheck(receiver, expression, written.statement.getEnd(), [symbol], undefined), viewType);
        }
    }
    else if (callee) {
        result.use = 'argument';
        result.callee = callee.name;
        result.calleeKind = callee.kind;
        const value = stripWrappers(receiver);
        result.freshArgument = ts.isObjectLiteralExpression(value) || (ts.isCallExpression(value) && freshInitializer(value) !== undefined);
        if (!callee.body) result.notes.push('the checker resolves no callee body to follow');
        else if (!callee.parameter) result.notes.push(`${callee.name} has no parameter at index ${callee.index}`);
        else if (!ts.isIdentifier(callee.parameter.name)) result.notes.push(`${callee.name}'s parameter is a binding pattern, which is not followed`);
        else {
            const symbol = checker.getSymbolAtLocation(callee.parameter.name);
            const through = throughSymbol(symbol, callee.body);
            result.onward = through.onward;
            const making = makingCheck(receiver, expression, callee.call.getEnd(), [], callee.call);
            const parameterType = checker.getTypeAtLocation(callee.parameter.name);
            for (const written of through.writes) record(written, making, parameterType);
            if (!through.writes.length) result.notes.push(`${callee.name} writes nothing through ${callee.parameter.name.text}` +
                (through.onward.length ? `; it passes it on to ${unique(through.onward.map(entry => entry.callee)).join(', ')}` : ''));
        }
    }
    else if (parent && ts.isPropertyAccessExpression(parent) && parent.expression === used) {
        result.use = 'property';
        result.notes.push(`the view's .${parent.name.text} is read, not written`);
    }
    else result.notes.push(`the view is used by ${parent ? ts.SyntaxKind[parent.kind] : 'nothing'}, which is not followed`);
    return result;
}

const entries = rows.map(row => {
    try {
        const measured = measure(row);
        if (measured.unfound) return { ...row, road: 'unfound', reason: measured.unfound, writes: [] };
        return { ...row, ...measured };
    }
    catch (error) {
        return { ...row, road: 'unfound', reason: `the scout failed on this row: ${error && error.message ? error.message : String(error)}`, writes: [] };
    }
});

const byField = {};
entries.forEach((entry, index) => {
    for (const write of entry.writes) {
        const field = byField[write.field] || (byField[write.field] = {
            writes: 0, sites: 0, making: 0, elsewhere: 0, publicAnywhere: false, declaredIn: [], distinctWheres: 0, rows: new Set(), wheres: new Set(),
        });
        field.writes++;
        field.rows.add(index);
        field.wheres.add(write.where);
        if (write.making === 'making') field.making++;
        else field.elsewhere++;
        if (write.public === true) field.publicAnywhere = true;
        field.declaredIn = unique([...field.declaredIn, ...write.declaredIn]);
    }
});
for (const field of Object.values(byField)) {
    field.sites = field.rows.size;
    field.distinctWheres = field.wheres.size;
    delete field.rows;
    delete field.wheres;
}

const unfound = entries.filter(entry => entry.road === 'unfound');
const roads = {};
for (const entry of entries) roads[entry.road] = (roads[entry.road] || 0) + 1;
fs.writeFileSync(output, JSON.stringify({
    typescript: ts.version, tree, rowsFile, roots: roots.length, rows: entries.length, roads, unfound: unfound.length, entries, byField,
}, null, 2) + '\n');

for (const entry of unfound) console.error(`UNFOUND ${entry.where} ${entry.reason}`);
console.error(`${entries.length} rows: ${Object.entries(roads).map(([road, count]) => `${road} ${count}`).join(', ')}`);
const order = Object.entries(byField).sort(([nameA, a], [nameB, b]) => b.writes - a.writes || nameA.localeCompare(nameB));
for (const [name, field] of order)
    console.log(`FIELD ${name} writes ${field.writes} sites ${field.sites} making ${field.making} elsewhere ${field.elsewhere} ` +
        `public ${field.publicAnywhere ? 'yes' : 'no'} declared ${field.declaredIn.length ? field.declaredIn.join(',') : 'none'}`);
console.log('PASS');
