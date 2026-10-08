#!/usr/bin/env node
'use strict';
// A closed-program, flow-insensitive may-write analysis of checker-bound values.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
const [treeArg, input, output, mode] = process.argv.slice(2);
assert(treeArg && input && output);
const tree = path.resolve(treeArg);
const configPath = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
assert(!config.error);
const configOptions = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath), {}, configPath);
const analysisOptions = {...configOptions.options, strict:true, exactOptionalPropertyTypes:true, noUncheckedIndexedAccess:true, verbatimModuleSyntax:true, erasableSyntaxOnly: false};
if (mode === '--fixture') analysisOptions.allowNonTsExtensions = true;
const rootNames = mode === '--fixture' ? config.config.files.map(f => path.resolve(path.dirname(configPath), f)) : configOptions.fileNames;
const program = ts.createProgram(rootNames, analysisOptions);
const checker = program.getTypeChecker();
const graph = [];
const nodes = new WeakMap();
const symbols = new Map();
const returns = new WeakMap();
const propertyRoots = new Map();
const functions = [];
const calls = [];
const writes = [];
const namedTypes = new Map();
let functionPropagation;
function fresh() { graph.push(new Set()); return graph.length - 1; }
function id(node) { if (!nodes.has(node)) nodes.set(node, fresh()); return nodes.get(node); }
function canonical(symbol) {
    if (symbol?.flags & ts.SymbolFlags.Alias) return checker.getAliasedSymbol(symbol);
    return symbol;
}
function sid(symbol) {
    symbol = canonical(symbol);
    if (!symbol) return undefined;
    if (!symbols.has(symbol)) symbols.set(symbol, fresh());
    return symbols.get(symbol);
}
function roots(symbol) { return symbol ? checker.getRootSymbols(canonical(symbol)) : []; }
function field(symbol) {
    const root = roots(symbol)[0];
    if (!root) return undefined;
    if (!propertyRoots.has(root)) propertyRoots.set(root, fresh());
    return propertyRoots.get(root);
}
function edge(a, b) {
    if (a === undefined || b === undefined || a === b || graph[a].has(b)) return;
    graph[a].add(b);
    if (functionPropagation) functionPropagation(a, b);
}
function ret(fn) { if (!returns.has(fn)) returns.set(fn, fresh()); return returns.get(fn); }
function location(n) {
    const file = n.getSourceFile();
    const p = file.getLineAndCharacterOfPosition(n.getStart(file));
    return `${path.relative(tree, file.fileName)}:${p.line + 1}:${p.character + 1}`;
}
function bind(name, value) {
    if (ts.isIdentifier(name)) edge(value, sid(checker.getSymbolAtLocation(name)));
    else if (ts.isObjectBindingPattern(name) || ts.isArrayBindingPattern(name)) {
        for (const e of name.elements) if (ts.isBindingElement(e)) {
            if (ts.isArrayBindingPattern(name)) bind(e.name, value);
            else {
                const type = checker.getTypeAtLocation(name);
                const key = e.propertyName || e.name;
                let keyText = key.text;
                if (ts.isComputedPropertyName(key)) { const t = checker.getTypeAtLocation(key.expression); if (t.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral)) keyText = String(t.value); }
                if (keyText === undefined || e.dotDotDotToken) bind(e.name, value);
                else bind(e.name, field(checker.getPropertyOfType(type, keyText)));
            }
        }
    }
}
function ownerFunction(n) {
    for (let p = n.parent; p; p = p.parent) if (ts.isFunctionLike(p)) return p;
}
function assignment(left, value, whole) {
    if (ts.isIdentifier(left)) edge(value, sid(checker.getSymbolAtLocation(left)));
    else if (ts.isPropertyAccessExpression(left) || ts.isElementAccessExpression(left)) {
        const symbol = accessSymbol(left);
        edge(value, field(symbol));
        writes.push({ receiver: id(left.expression), symbol, access: left, node: whole,
            literalKey: !symbol && ts.isPropertyAccessExpression(left) ? left.name.escapedText : undefined,
            dynamicKey: !symbol && ts.isElementAccessExpression(left) ? checker.getTypeAtLocation(left.argumentExpression) : undefined,
            value, valueNode: ts.isBinaryExpression(whole) ? whole.right : undefined,
            operation: ts.isBinaryExpression(whole) ? ts.tokenToString(whole.operatorToken.kind) : ts.isDeleteExpression(whole) ? 'delete' : ts.tokenToString(whole.operator) });
    } else if (ts.isObjectLiteralExpression(left) || ts.isArrayLiteralExpression(left)) {
        for (const p of left.properties || left.elements) {
            if (ts.isShorthandPropertyAssignment(p)) edge(value, sid(checker.getShorthandAssignmentValueSymbol(p)));
            else if (ts.isPropertyAssignment(p)) assignment(p.initializer, value, whole);
            else if (ts.isSpreadElement(p) || ts.isSpreadAssignment(p)) assignment(p.expression, value, whole);
            else if (ts.isExpression(p)) assignment(p, value, whole);
        }
    }
}
function accessSymbol(n) {
    if (ts.isPropertyAccessExpression(n)) return checker.getSymbolAtLocation(n.name);
    if (!ts.isElementAccessExpression(n)) return undefined;
    const key = checker.getTypeAtLocation(n.argumentExpression);
    if (key.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral))
        return checker.getPropertyOfType(checker.getTypeAtLocation(n.expression), String(key.value));
}
function visit(n) {
    if (ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n) || ts.isClassDeclaration(n)) {
        if (n.name) {
            const symbol = checker.getSymbolAtLocation(n.name);
            const type = symbol && checker.getDeclaredTypeOfSymbol(symbol);
            if (type) { if (!namedTypes.has(n.name.text)) namedTypes.set(n.name.text, []); namedTypes.get(n.name.text).push(type); }
        }
    }
    if (ts.isFunctionLike(n)) {
        functions.push(n);
        if (n.name && ts.isIdentifier(n.name)) edge(id(n), sid(checker.getSymbolAtLocation(n.name)));
        if (n.body && !ts.isBlock(n.body)) edge(id(n.body), ret(n));
    }
    if (ts.isIdentifier(n)) edge(sid(checker.getSymbolAtLocation(n)), id(n));
    if (ts.isVariableDeclaration(n) || ts.isParameter(n)) { if (n.initializer) bind(n.name, id(n.initializer)); }
    if (ts.isParenthesizedExpression(n) || ts.isAsExpression(n) || ts.isTypeAssertionExpression(n) || ts.isNonNullExpression(n)) edge(id(n.expression), id(n));
    if (ts.isConditionalExpression(n)) { edge(id(n.whenTrue), id(n)); edge(id(n.whenFalse), id(n)); }
    if (ts.isBinaryExpression(n)) {
        if (n.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && n.operatorToken.kind <= ts.SyntaxKind.LastAssignment) {
            assignment(n.left, id(n.right), n); edge(id(n.right), id(n));
        } else if ([ts.SyntaxKind.BarBarToken, ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.QuestionQuestionToken, ts.SyntaxKind.CommaToken].includes(n.operatorToken.kind)) {
            edge(id(n.left), id(n)); edge(id(n.right), id(n));
        }
    }
    if ((ts.isPrefixUnaryExpression(n) || ts.isPostfixUnaryExpression(n)) && [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(n.operator)) assignment(n.operand, id(n), n);
    if (ts.isDeleteExpression(n)) assignment(n.expression, id(n), n);
    if (ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n)) {
        edge(field(accessSymbol(n)), id(n));
        if (ts.isElementAccessExpression(n) && (checker.isArrayType(checker.getTypeAtLocation(n.expression)) || checker.isTupleType(checker.getTypeAtLocation(n.expression)))) edge(id(n.expression), id(n));
    }
    if (ts.isReturnStatement(n) && n.expression) { const fn = ownerFunction(n); if (fn) edge(id(n.expression), ret(fn)); }
    if (ts.isPropertyAssignment(n)) {
        edge(id(n.initializer), field(checker.getSymbolAtLocation(n.name)));
        const context = checker.getContextualType(n.parent);
        const symbol = checker.getSymbolAtLocation(n.name);
        if (context && symbol) edge(id(n.initializer), field(checker.getPropertyOfType(context, symbol.escapedName)));
    }
    if (ts.isShorthandPropertyAssignment(n)) edge(sid(checker.getShorthandAssignmentValueSymbol(n)), field(checker.getSymbolAtLocation(n.name)));
    if (ts.isArrayLiteralExpression(n)) for (const e of n.elements) edge(id(ts.isSpreadElement(e) ? e.expression : e), id(n));
    if (ts.isSpreadAssignment(n) || ts.isSpreadElement(n)) edge(id(n.expression), id(n));
    if (ts.isForOfStatement(n)) {
        if (ts.isVariableDeclarationList(n.initializer)) for (const d of n.initializer.declarations) bind(d.name, id(n.expression));
        else assignment(n.initializer, id(n.expression), n);
    }
    if (ts.isCallExpression(n) || ts.isNewExpression(n)) calls.push(n);
    ts.forEachChild(n, visit);
}
const sources = program.getSourceFiles().filter(f => !f.isDeclarationFile && f.fileName.startsWith(tree + path.sep + 'src' + path.sep));
for (const file of sources) visit(file);
const thisUses = new WeakMap();
for (const file of sources) {
    function visitThis(n) { if(n.kind===ts.SyntaxKind.ThisKeyword) { const fn=ownerFunction(n); if(fn) { if(!thisUses.has(fn))thisUses.set(fn,[]); thisUses.get(fn).push(n); } } ts.forEachChild(n,visitThis); }
    visitThis(file);
}
function connectCall(call, signature) {
    const decl = signature?.declaration;
    if (!decl) return;
    const args = call.arguments || [];
    for (let i = 0; i < args.length; i++) {
        const param = decl.parameters?.[Math.min(i, decl.parameters.length - 1)];
        if (param) bind(param.name, id(ts.isSpreadElement(args[i]) ? args[i].expression : args[i]));
    }
    if (decl.body) {
        edge(ret(decl), id(call));
        if(ts.isPropertyAccessExpression(call.expression)) for(const use of thisUses.get(decl) || []) edge(id(call.expression.expression),id(use));
    }
}
for (const call of calls) {
    connectCall(call, checker.getResolvedSignature(call));
    // Include implementations behind overload signatures, by their checker symbol.
    const symbol = ts.isPropertyAccessExpression(call.expression) ? checker.getSymbolAtLocation(call.expression.name) : checker.getSymbolAtLocation(call.expression);
    for (const d of canonical(symbol)?.declarations || []) if (ts.isFunctionLike(d) && d.body) connectCall(call, checker.getSignatureFromDeclaration(d));
    // Intrinsic array callbacks pass the receiver's elements to the callback.
    if (ts.isPropertyAccessExpression(call.expression)) {
        const method = checker.getSymbolAtLocation(call.expression.name);
        const receiver = call.expression.expression;
        const receiverType = checker.getTypeAtLocation(receiver);
        if (checker.isArrayType(receiverType) || checker.isTupleType(receiverType)) {
            for (const arg of call.arguments || []) {
                for (const sig of checker.getSignaturesOfType(checker.getTypeAtLocation(arg), ts.SignatureKind.Call)) {
                    const fn = sig.declaration;
                    if (fn?.parameters?.[0]) bind(fn.parameters[0].name, id(receiver));
                    if (fn?.body) edge(ret(fn), id(call));
                }
            }
            if (method && ['push','unshift','splice'].includes(method.getName())) for(const arg of call.arguments || [])edge(id(arg),id(receiver));
            // Returned array elements, copied and filtered, retain value provenance.
            const copying = new Set(['filter','slice','concat','toSorted','toReversed','with','splice','sort','reverse']);
            if (method && copying.has(method.getName()) && checker.isArrayType(checker.getTypeAtLocation(call))) edge(id(receiver), id(call));
        }
        // Model standard collection storage by its resolved library declaration.
        if (method && roots(method).some(s => (s.declarations || []).some(d => d.getSourceFile().isDeclarationFile && ['Map','ReadonlyMap','Set','ReadonlySet'].includes(d.parent.name?.text)))) {
            if (['set','add'].includes(method.getName())) {
                const arg = call.arguments?.[method.getName()==='set' ? 1 : 0];
                if(arg)edge(id(arg),id(receiver));
            } else if (['get','values','entries'].includes(method.getName()))edge(id(receiver),id(call));
        }
        // Object.assign / defineProperty write through their first argument.
        const objectSymbol = checker.resolveName('Object',call,ts.SymbolFlags.Value,false);
        const objectType = objectSymbol && roots(objectSymbol).some(s => (s.declarations || []).some(d => d.getSourceFile().isDeclarationFile)) && checker.getTypeOfSymbolAtLocation(objectSymbol,call);
        const assign = objectType && checker.getPropertyOfType(objectType,'assign');
        const define = objectType && checker.getPropertyOfType(objectType,'defineProperty');
        const isIntrinsic = expected => expected && roots(expected).some(s => roots(method).includes(s));
        if ((isIntrinsic(assign) || isIntrinsic(define)) && call.arguments?.[0])edge(id(call.arguments[0]),id(call));
        if (isIntrinsic(assign) && call.arguments?.[0]) for (const source of call.arguments.slice(1)) {
            if (!ts.isObjectLiteralExpression(source)) continue;
            for (const prop of source.properties) if(ts.isPropertyAssignment(prop)) {
                const keySymbol = checker.getSymbolAtLocation(prop.name);
                const symbol = keySymbol && checker.getPropertyOfType(checker.getTypeAtLocation(call.arguments[0]),keySymbol.escapedName);
                if(symbol)writes.push({receiver:id(call.arguments[0]),symbol,access:{expression:call.arguments[0],getText:()=>call.getText()},node:call,valueNode:prop.initializer,operation:'Object.assign'});
            }
        }
        if (isIntrinsic(define) && call.arguments?.[1]) {
            const keyType=checker.getTypeAtLocation(call.arguments[1]);
            const symbol=(keyType.flags & ts.TypeFlags.StringLiteral) && checker.getPropertyOfType(checker.getTypeAtLocation(call.arguments[0]),keyType.value);
            if(symbol)writes.push({receiver:id(call.arguments[0]),symbol,access:{expression:call.arguments[0],getText:()=>call.getText()},node:call,valueNode:call.arguments[2],operation:'Object.defineProperty'});
        }
    }
}
// Propagate known function identities incrementally through the alias graph.
const callExpressions = new Map();
for (const call of calls) {
    const key = id(call.expression);
    if (!callExpressions.has(key)) callExpressions.set(key, []);
    callExpressions.get(key).push(call);
}
const functionValues = new Map();
const work = [];
function enqueue(node, fn) {
    if (!functionValues.has(node)) functionValues.set(node, new Set());
    const set = functionValues.get(node);
    if (!set.has(fn)) { set.add(fn); work.push([node, fn]); }
}
functionPropagation = (from, to) => { for (const fn of functionValues.get(from) || []) enqueue(to, fn); };
for (const fn of functions) if (fn.body) enqueue(id(fn), fn);
let aliasRounds = 0;
for (let at = 0; at < work.length; at++) {
    const [node, fn] = work[at];
    for (const call of callExpressions.get(node) || []) connectCall(call, checker.getSignatureFromDeclaration(fn));
    for (const next of graph[node]) enqueue(next, fn);
    aliasRounds++;
}
functionPropagation = undefined;
console.error(`function alias fixed point: ${aliasRounds} propagated identities`);
function findNode(r) {
    const file = program.getSourceFile(path.join(tree, 'src/compiler', r.File)); assert(file, r.File);
    const start = file.getPositionOfLineAndCharacter(r.Line - 1, r.Column - 1);
    let node;
    function visit(n) {
        if (n.getStart(file) === start && [ts.SyntaxKind[n.kind], 'Kind' + ts.SyntaxKind[n.kind]].includes(r.Kind) && n.getText(file) === r.Text) node = n;
        if (n.pos <= start && start < n.end) ts.forEachChild(n, visit);
    }
    visit(file); assert(node, `site moved: ${r.File}:${r.Line}:${r.Column}`); return node;
}
function targetProperty(r, n) {
    const candidates = [];
    if (ts.isAsExpression(n) || ts.isTypeAssertionExpression(n)) candidates.push(checker.getTypeFromTypeNode(n.type));
    candidates.push(checker.getContextualType(n));
    candidates.push(checker.getTypeAtLocation(n));
    candidates.push(...(namedTypes.get(r.Target) || []));
    let parent = n.parent;
    if (ts.isSpreadAssignment(n) || ts.isShorthandPropertyAssignment(n)) candidates.push(checker.getContextualType(parent));
    if (ts.isCallExpression(parent) || ts.isNewExpression(parent)) {
        const sig = checker.getResolvedSignature(parent);
        const index = parent.arguments?.indexOf(n);
        const param = sig?.parameters[Math.min(index, sig.parameters.length - 1)];
        if (param) candidates.push(checker.getTypeOfSymbolAtLocation(param, n));
    }
    const seen = new Set();
    function search(type) {
        if (!type || seen.has(type)) return undefined; seen.add(type);
        if (checker.typeToString(type) === r.Target) { const p = checker.getPropertyOfType(type, r.Property); if (p) return {symbol:p,type}; }
        if (type.isUnionOrIntersection()) for (const t of type.types) { const p = search(t); if (p) return p; }
        if (checker.isArrayType(type) || checker.isTupleType(type)) for (const t of checker.getTypeArguments(type)) { const p = search(t); if (p) return p; }
        for (const sig of checker.getSignaturesOfType(type, ts.SignatureKind.Call)) { const p = search(checker.getReturnTypeOfSignature(sig)); if (p) return p; }
    }
    for (const t of candidates) { const p = search(t); if (p) return p; }
    // Anonymous contextual types may print differently in the two checkers.
    for (const t of candidates) { if (t) { const p = checker.getPropertyOfType(t, r.Property); if (p) return {symbol:p,type:t}; } }
    function loose(type) {
        if (!type) return;
        const property = checker.getPropertyOfType(type, r.Property);
        if (property) return {symbol:property,type};
        if (type.isUnionOrIntersection()) for (const member of type.types) { const p=loose(member); if(p)return p; }
    }
    for (const type of candidates) { const p=loose(type); if(p)return p; }
    throw new Error(`unresolved target property: ${location(n)} ${r.Target}.${r.Property}`);
}
function reachable(seed) {
    const seen = new Set([seed]); const todo = [seed];
    while (todo.length) for (const next of graph[todo.pop()]) if (!seen.has(next)) { seen.add(next); todo.push(next); }
    return seen;
}
const rows = JSON.parse(fs.readFileSync(input, 'utf8'));
const results = [];
for (const r of rows) {
    const n = findNode(r);
    const target = targetProperty(r, n);
    const p = target.symbol;
    let sourceType = checker.getTypeAtLocation(ts.isAsExpression(n) ? n.expression : n);
    if (namedTypes.get(r.Source)?.length === 1) sourceType = namedTypes.get(r.Source)[0];
    else if (checker.isArrayType(sourceType) || checker.isTupleType(sourceType)) sourceType = checker.getTypeArguments(sourceType)[0] || sourceType;
    else { const sig=checker.getSignaturesOfType(sourceType,ts.SignatureKind.Call)[0]; if(sig)sourceType=checker.getReturnTypeOfSignature(sig); }
    sourceType = checker.getNonNullableType(sourceType);
    const missing = (r.Source === 'never' ? [] : checker.getPropertiesOfType(target.type)).filter(s => (s.flags & ts.SymbolFlags.Optional) && !checker.getPropertyOfType(sourceType,s.escapedName));
    if (!missing.some(s => roots(s).some(root => roots(p).includes(root)))) missing.push(p);
    const reach = reachable(id(n));
    // Function-return and parameter relations are recorded at the function value.
    const valueType = checker.getTypeAtLocation(n);
    for (const sig of checker.getSignaturesOfType(valueType, ts.SignatureKind.Call)) {
        const d = sig.declaration;
        if (d?.body) {
            for (const x of reachable(ret(d))) reach.add(x);
            for (const parameter of d.parameters) { const symbol=checker.getSymbolAtLocation(parameter.name); if(symbol)for(const x of reachable(sid(symbol)))reach.add(x); }
        }
    }
    const hits = [];
    for (const w of writes) {
        if (!reach.has(w.receiver)) continue;
        const receiverType = checker.getNonNullableType(checker.getTypeAtLocation(w.access.expression));
        if (!checker.isTypeAssignableTo(receiverType,target.type) && !checker.isTypeAssignableTo(target.type,receiverType)) continue;
        // Project the target member into this alias's bound receiver type. Compare
        // checker root symbols, never property spelling in source text.
        const matched = missing.find(property => {
            const projected = checker.getPropertyOfType(checker.getTypeAtLocation(w.access.expression), property.escapedName);
            if (!projected && !(receiverType.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown))) return false;
            if (w.literalKey !== undefined) return w.literalKey === property.escapedName;
            if (w.dynamicKey) {
                const types=w.dynamicKey.isUnion() ? w.dynamicKey.types : [w.dynamicKey];
                return types.some(t => (t.flags & (ts.TypeFlags.Any | ts.TypeFlags.String)) || ((t.flags & ts.TypeFlags.StringLiteral) && t.value===property.getName()));
            }
            return projected && roots(projected).some(s => roots(w.symbol).includes(s));
        });
        if (!matched) continue;
        hits.push({untyped_receiver:!!(receiverType.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown)),property:matched.getName(),where:location(w.node), access:w.access.getText(), operation:w.operation,
            written:w.valueNode?.getText() || w.node.getText(),
            written_type:w.valueNode ? checker.typeToString(checker.getTypeAtLocation(w.valueNode)) : undefined,
            receiver_type:checker.typeToString(checker.getTypeAtLocation(w.access.expression)),
            property_declarations:roots(w.symbol || matched).flatMap(s => (s.declarations || []).map(location))});
    }
    const unique = [...new Map(hits.map(w => [w.where + w.access,w])).values()];
    results.push({...r,stock_source:checker.typeToString(checker.getTypeAtLocation(ts.isAsExpression(n) ? n.expression : n)),
        target_property_declarations:roots(p).flatMap(s => (s.declarations || []).map(location)),
        checker_source_declares_first_property:!!checker.getPropertyOfType(sourceType,p.escapedName),
        missing_optional_properties:missing.map(s => s.getName()),classification:unique.length ? 'write' : 'read-only',writes:unique});
}
const counts = {total:results.length,read_only:results.filter(r => r.classification === 'read-only').length,write:results.filter(r => r.classification === 'write').length};
fs.writeFileSync(output,JSON.stringify({typescript:ts.version,counts,scope:'compiler project; flow-insensitive may-write reachability with checker-bound variable, parameter and member symbols',options:analysisOptions,function_alias_propagations:aliasRounds,graph_nodes:graph.length,write_operations:writes.length,rows:results},null,2)+'\n');
console.log(JSON.stringify(counts));
