// Typed call census. Redirect stdout outside the checkout. No source is changed.
const path = require('path');
const checkout = process.env.ADAMIC_TYPESCRIPT_SOURCE;
if (!checkout)
    throw new Error('ADAMIC_TYPESCRIPT_SOURCE must name the pinned TypeScript 6.0.3 checkout');
const ts = require(path.join(checkout, 'lib/typescript.js'));
if (ts.version !== '6.0.3')
    throw new Error('expected TypeScript 6.0.3, got ' + ts.version);
const root = path.resolve(process.argv[2] || path.resolve(__dirname, '../cohere/TypeScript/tsc/testdata/fixtures/compiler'));
const cfg = ts.readConfigFile(path.join(root, 'tsconfig.json'), ts.sys.readFile);
if (cfg.error)
    throw new Error(ts.flattenDiagnosticMessageText(cfg.error.messageText, ' '));
const parsed = ts.parseJsonConfigFileContent(cfg.config, ts.sys, root);
if (parsed.errors.length)
    throw new Error(parsed.errors.map(d => ts.flattenDiagnosticMessageText(d.messageText, ' ')).join('\n'));
const program = ts.createProgram(parsed.fileNames, parsed.options), checker = program.getTypeChecker();
function kind(t, seen = new Set()) {
    if (seen.has(t))
        return 'other';
    seen.add(t);
    if (t.flags & ts.TypeFlags.Any)
        return 'any';
    if (t.flags & ts.TypeFlags.Unknown)
        return 'unknown';
    if (t.flags & ts.TypeFlags.StringLike)
        return 'string';
    if (t.flags & ts.TypeFlags.NumberLike)
        return 'number';
    if (t.flags & ts.TypeFlags.BooleanLike)
        return 'boolean';
    if (t.flags & ts.TypeFlags.Null)
        return 'null';
    if (t.flags & ts.TypeFlags.Never)
        return 'never';
    if (checker.isTupleType(t))
        return checker.typeToString(t).startsWith('readonly') ? 'readonly tuple' : 'tuple';
    if (t.flags & ts.TypeFlags.Undefined)
        return 'undefined';
    if (checker.isArrayType(t) || checker.isTupleType(t))
        return checker.typeToString(t).startsWith('readonly') ? 'readonly array' : 'array';
    if (t.isUnion())
        return [...new Set(t.types.map(x => kind(x, new Set(seen))))].sort().join('|');
    if (t.symbol?.name === 'RegExp')
        return 'regexp';
    if (t.isIntersection()) {
        for (const m of t.types) {
            const k = kind(m, new Set(seen));
            if (k === 'string' || k === 'number' || k.includes('array'))
                return k;
        }
    }
    if (t.flags & ts.TypeFlags.TypeParameter) {
        const c = checker.getBaseConstraintOfType(t);
        return c ? kind(c, seen) : 'type parameter';
    }
    if (t.flags & ts.TypeFlags.Object) {
        for (const b of checker.getBaseTypes(t) || []) {
            const k = kind(b, seen);
            if (k.includes('array'))
                return k;
        }
    }
    return checker.getSignaturesOfType(t, ts.SignatureKind.Call).length ? 'function' : 'object';
}
const rows = [], unresolved = [];
const methodNames = new Set();
for (const sf of program.getSourceFiles()) {
    if (!program.isSourceFileDefaultLibrary(sf))
        continue;
    ts.forEachChild(sf, n => { if (ts.isInterfaceDeclaration(n) && ['String', 'Array', 'ReadonlyArray', 'StringConstructor', 'ArrayConstructor'].includes(n.name.text))
        for (const member of n.members)
            if (member.name && ts.isIdentifier(member.name) && ts.isMethodSignature(member))
                methodNames.add(member.name.text); });
}
for (const sf of program.getSourceFiles()) {
    if (!sf.fileName.startsWith(root + '/') || sf.isDeclarationFile)
        continue;
    function visit(n) {
        if (ts.isCallExpression(n)) {
            let recv, name;
            if (ts.isPropertyAccessExpression(n.expression)) {
                recv = n.expression.expression;
                name = n.expression.name.text;
            }
            else if (ts.isElementAccessExpression(n.expression) && ts.isStringLiteral(n.expression.argumentExpression)) {
                recv = n.expression.expression;
                name = n.expression.argumentExpression.text;
            }
            if (recv) {
                const sig = checker.getResolvedSignature(n), decl = sig?.declaration;
                let builtinDeclaration = decl;
                let owner = decl?.parent?.name?.text;
                if (name === 'call' || name === 'apply') {
                    const calledType = checker.getTypeAtLocation(recv);
                    const target = checker.getSignaturesOfType(calledType, ts.SignatureKind.Call)[0]?.declaration;
                    if (['String', 'Array', 'ReadonlyArray'].includes(target?.parent?.name?.text)) {
                        builtinDeclaration = target;
                        owner = target.parent.name.text;
                        name = (target.name?.getText() || 'unknown') + '.' + name;
                    }
                }
                const family = ['String', 'Array', 'ReadonlyArray', 'StringConstructor', 'ArrayConstructor'].includes(owner) ? owner : null;
                if (family && builtinDeclaration && program.isSourceFileDefaultLibrary(builtinDeclaration.getSourceFile())) {
                    const args = n.arguments.map(a => { const t = checker.getTypeAtLocation(a); return { kind: kind(t), type: checker.typeToString(t), syntax: ts.SyntaxKind[a.kind], spread: ts.isSpreadElement(a), spreadSourceType: ts.isSpreadElement(a) ? checker.typeToString(checker.getTypeAtLocation(a.expression)) : undefined, functionArity: (ts.isArrowFunction(a) || ts.isFunctionExpression(a)) ? a.parameters.length : checker.getSignaturesOfType(t, ts.SignatureKind.Call)[0]?.parameters.length }; });
                    for (let i = 0; i < args.length; i++) {
                        if ((i === 0 && ['map', 'flatMap', 'filter', 'sort', 'forEach', 'some', 'every', 'find', 'findIndex', 'findLast', 'findLastIndex', 'reduce', 'reduceRight'].includes(name)) || (i === 1 && ['replace', 'replaceAll'].includes(name)))
                            args[i].callbackArity = args[i].functionArity;
                    }
                    const loc = sf.getLineAndCharacterOfPosition(n.getStart(sf));
                    const receiverType = checker.typeToString(checker.getTypeAtLocation(recv));
                    const shape = family + '.' + name + ' receiver=' + kind(checker.getTypeAtLocation(recv)) + ' argc=' + args.length + ' args=' + args.map(a => a.kind + (a.functionArity === undefined ? '' : (a.callbackArity === undefined ? ':function' : ':callback') + a.functionArity) + (a.spread ? ':spread' : '')).join(',');
                    rows.push({ file: path.relative(root, sf.fileName), line: loc.line + 1, column: loc.character + 1, method: family + '.' + name, shape, receiverType, optionalChain: ts.isOptionalChain(n), args, text: n.getText(sf) });
                }
                else if (methodNames.has(name) && (!decl || (checker.getTypeAtLocation(recv).flags & ts.TypeFlags.Any))) {
                    const loc = sf.getLineAndCharacterOfPosition(n.getStart(sf));
                    unresolved.push({ file: path.relative(root, sf.fileName), line: loc.line + 1, column: loc.character + 1, name, receiverType: checker.typeToString(checker.getTypeAtLocation(recv)), text: n.getText(sf) });
                }
            }
        }
        ts.forEachChild(n, visit);
    }
    visit(sf);
}
const methods = {}, shapes = {};
for (const r of rows) {
    methods[r.method] = (methods[r.method] || 0) + 1;
    (shapes[r.shape] ??= []).push(r);
}
const diagnostics = ts.getPreEmitDiagnostics(program).map(d => ({ code: d.code, file: d.file && path.relative(root, d.file.fileName), message: ts.flattenDiagnosticMessageText(d.messageText, ' ') }));
const typedShapes = {};
for (const row of rows) {
    const typedShape = row.shape + ' type=' + row.receiverType;
    (typedShapes[typedShape] ??= []).push(row);
}
const report = { typedShapes, unresolved, version: ts.version, node: process.version, root, roots: parsed.fileNames.length, total: rows.length, methods, shapes, rows, diagnostics };
process.stdout.write(JSON.stringify(report, null, 2) + '\n');
