// Stock TypeScript resolves declarations independently of the latent lowerer.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error(`expected TypeScript 6.0.3, got ${ts.version}`);
const [mode, rootArgument, input, output] = process.argv.slice(2);
const root = path.resolve(rootArgument);
const compilerRoot = mode === 'fixture' ? root : path.join(root, 'src/compiler');
const typeRoot = path.join(path.dirname(path.dirname(require.resolve('typescript/package.json'))), '@types');
const options = {strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
    noEmit: true, target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Bundler, allowImportingTsExtensions: true,
    lib: ['lib.es2024.d.ts'], types: ['node'], typeRoots: [typeRoot], skipLibCheck: true};
function files(directory) {
    return fs.readdirSync(directory, {withFileTypes: true}).flatMap(e => {
        const p = path.join(directory, e.name);
        return e.isDirectory() ? files(p) : /\.(ts|a)$/.test(p) ? [p] : [];
    });
}
if (mode !== 'fixture') {
    const manifest = JSON.parse(fs.readFileSync(path.join(path.dirname(input), 'source-manifest.json'), 'utf8'));
    for (const [name, expected] of Object.entries(manifest)) {
        const bytes = fs.readFileSync(path.join(compilerRoot, name));
        if (bytes.length !== expected.bytes || crypto.createHash('sha256').update(bytes).digest('hex') !== expected.sha256)
            throw Error(`frozen source hash mismatch: ${name}`);
    }
}
const roots = mode === 'fixture' ? [input] : files(compilerRoot);
const host = ts.createCompilerHost(options);
const readSource = host.getSourceFile;
host.getSourceFile = (name, version, onError) => name.endsWith('.a')
    ? ts.createSourceFile(name, fs.readFileSync(name, 'utf8'), version, true, ts.ScriptKind.TS)
    : readSource(name, version, onError);
const program = ts.createProgram(roots, {...options, allowNonTsExtensions: true}, host);
const checker = program.getTypeChecker();
function local(file) {
    if (file.fileName.startsWith(root + '/')) return path.relative(root, file.fileName).split(path.sep).join('/');
    if (program.isSourceFileDefaultLibrary(file)) return 'typescript/lib/' + path.basename(file.fileName);
    if (file.fileName.includes('/node_modules/')) return file.fileName.split('/node_modules/').pop();
    return file.fileName;
}
function location(node) {
    const file = node.getSourceFile();
    const point = file.getLineAndCharacterOfPosition(node.getStart(file));
    return `${local(file)}:${point.line+1}:${point.character+1}`;
}
function owner(declaration) {
    const file = declaration.getSourceFile();
    if (program.isSourceFileDefaultLibrary(file)) return 'lib.d.ts';
    if (file.fileName.includes('/node_modules/@types/node/')) return 'Node typings';
    if (file.fileName.startsWith(compilerRoot + '/')) return 'tsc';
    return 'other';
}
function typeText(type) {
    return checker.typeToString(type, undefined, ts.TypeFormatFlags.NoTruncation);
}
function any(type) { return !!(type.flags & ts.TypeFlags.Any); }
function qualifiedName(declaration) {
    if (declaration.name) {
        const names = [declaration.name.getText()];
        let ancestor = declaration.parent;
        while (ancestor && !ts.isSourceFile(ancestor)) {
            if (ancestor.name && (ts.isFunctionLike(ancestor) || ts.isClassDeclaration(ancestor) ||
                ts.isInterfaceDeclaration(ancestor) || ts.isModuleDeclaration(ancestor))) names.unshift(ancestor.name.getText());
            ancestor = ancestor.parent;
        }
        return names.join('.');
    }
    if (ts.isPropertyAssignment(declaration.parent)) {
        const property = declaration.parent.name.getText();
        const object = declaration.parent.parent;
        if (ts.isVariableDeclaration(object.parent)) return object.parent.name.getText() + '.' + property;
        return property + ' (at ' + location(declaration) + ')';
    }
    if (ts.isCallSignatureDeclaration(declaration)) return `${declaration.parent.name?.getText() || '(anonymous)'}.(call)`;
    if (ts.isVariableDeclaration(declaration.parent)) return declaration.parent.name.getText();
    return `(anonymous ${ts.SyntaxKind[declaration.kind]} at ${location(declaration)})`;
}
function returns(declaration) {
    if (!declaration.body) return [];
    if (!ts.isBlock(declaration.body)) return [declaration.body];
    const result = [];
    function walk(node) {
        if (node !== declaration.body && ts.isFunctionLike(node)) return;
        if (ts.isReturnStatement(node)) { result.push(node.expression || node); return; }
        ts.forEachChild(node, walk);
    }
    walk(declaration.body);
    return result;
}
function traceAny(expression, seen = new Set(), depth = 0) {
    if (!expression || depth > 8 || seen.has(expression)) return [];
    seen.add(expression);
    if (ts.isParenthesizedExpression(expression) || ts.isNonNullExpression(expression)) return traceAny(expression.expression, seen, depth+1);
    if (ts.isAsExpression(expression) || ts.isTypeAssertionExpression(expression)) {
        if (expression.type.kind === ts.SyntaxKind.AnyKeyword)
            return [{owner: owner(expression), where: location(expression.type), evidence: 'explicit any assertion'}];
        return traceAny(expression.expression, seen, depth+1);
    }
    if (ts.isCallExpression(expression)) {
        const signature = checker.getResolvedSignature(expression);
        const declaration = signature?.declaration;
        if (declaration && any(checker.getReturnTypeOfSignature(signature))) {
            return [{owner: owner(declaration), where: location(declaration), callee: qualifiedName(declaration),
                evidence: 'resolved call signature returns any', declared_return: declaration.type?.getText() || null}];
        }
    }
    const symbol = checker.getSymbolAtLocation(ts.isPropertyAccessExpression(expression) ? expression.name : expression);
    const resolved = symbol && symbol.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(symbol) : symbol;
    const declaration = resolved?.valueDeclaration;
    if (declaration) {
        if (declaration.type?.kind === ts.SyntaxKind.AnyKeyword)
            return [{owner: owner(declaration), where: location(declaration.type), evidence: 'explicit any binding annotation'}];
        if (declaration.initializer) return traceAny(declaration.initializer, seen, depth+1);
    }
    const children = [];
    ts.forEachChild(expression, child => { if (any(checker.getTypeAtLocation(child))) children.push(...traceAny(child, seen, depth+1)); });
    return children;
}
function describe(declaration, signature = checker.getSignatureFromDeclaration(declaration)) {
    if (!signature) throw Error(`missing signature at ${location(declaration)}`);
    const returnType = checker.getReturnTypeOfSignature(signature);
    const expressions = returns(declaration);
    const returnEvidence = expressions.map(expression => {
        const type = ts.isReturnStatement(expression) ? checker.getUndefinedType() : checker.getTypeAtLocation(expression);
        return {where: location(expression), expression: expression.getText().slice(0, 500),
            type: typeText(type), widened_type: typeText(checker.getBaseTypeOfLiteralType(type)),
            any: any(type), any_sources: any(type) ? traceAny(expression) : []};
    });
    const declared = declaration.type?.getText() || null;
    let source;
    if (!any(returnType)) source = 'stock return is not any; latent checker/substitution/context disagreement';
    else if (declaration.type?.kind === ts.SyntaxKind.AnyKeyword) source = 'explicit any return annotation';
    else source = 'inferred or contextual any return; see return-expression evidence';
    // A body-derived candidate needs every return expression to avoid any. This
    // is deliberately not a proof that arbitrary control flow has no fallthrough.
    const bodyCandidate = returnEvidence.length && returnEvidence.every(e => !e.any)
        ? [...new Set(returnEvidence.map(e => e.widened_type))].join(' | ') : null;
    return {callee: qualifiedName(declaration), declaration: location(declaration), declaration_owner: owner(declaration),
        declared_return: declared, stock_return: typeText(returnType), stock_return_is_any: any(returnType),
        any_explanation: source, type_parameters: declaration.typeParameters?.map(p => p.getText()) || [],
        body_return_candidate: bodyCandidate, candidate_limit: 'return-expression types only; fallthrough, casts and generic substitutions require validation',
        return_evidence: returnEvidence,
        declaration_sha256: crypto.createHash('sha256').update(fs.readFileSync(declaration.getSourceFile().fileName)).digest('hex')};
}
const callees = {};
const observations = [];
if (mode === 'fixture') {
    const file = program.getSourceFile(input);
    function walk(node) {
        if (ts.isCallExpression(node)) {
            const signature = checker.getResolvedSignature(node);
            if (signature?.declaration && any(checker.getReturnTypeOfSignature(signature))) {
                const result = describe(signature.declaration, signature);
                if (process.env.ANY_RETURNS_MUTANT === 'lib-to-tsc' && result.declaration_owner === 'lib.d.ts') result.declaration_owner = 'tsc';
                callees[result.declaration] = result;
                const start = Buffer.byteLength(file.text.slice(0, node.getStart(file)));
                const end = Buffer.byteLength(file.text.slice(0, node.end));
                observations.push({where: location(node), callee: result.declaration, expression: node.expression.getText(),
                    file: local(file), start, end, attributed_hidden_bytes: end-start});
            }
        }
        ts.forEachChild(node, walk);
    }
    walk(file);
} else {
    const boundaries = JSON.parse(fs.readFileSync(input, 'utf8')).boundaries;
    const byFile = new Map();
    for (const boundary of boundaries) {
        const match = boundary.diagnostic.match(/^(.+):(\d+):(\d+): stage 0 can't lower a function returning any yet$/);
        if (!match) throw Error(`invalid diagnostic: ${boundary.diagnostic}`);
        const relative = match[1].split('/src/compiler/')[1];
        const name = path.join(compilerRoot, relative);
        const file = program.getSourceFile(name);
        if (!file) throw Error(`missing source ${name}`);
        if (!byFile.has(name)) {
            const functions = [];
            function walk(node) { if (ts.isFunctionLike(node)) functions.push(node); ts.forEachChild(node, walk); }
            walk(file); byFile.set(name, functions);
        }
        const position = file.getPositionOfLineAndCharacter(Number(match[2])-1, Number(match[3])-1);
        const declarations = byFile.get(name).filter(node => node.name ? node.name.getStart(file) <= position && position < node.name.end : node.getStart(file) === position);
        if (declarations.length !== 1) throw Error(`expected one function at ${boundary.diagnostic}, got ${declarations.length}`);
        const declaration = declarations[0];
        const key = location(declaration);
        if (!callees[key]) callees[key] = describe(declaration);
        observations.push({...boundary, callee: key});
    }
}
if (mode !== 'fixture') {
    const callsByCallee = new Map();
    for (const file of program.getSourceFiles().filter(file => file.fileName.startsWith(compilerRoot + '/'))) {
        const offsets = new Uint32Array(file.text.length + 1);
        let bytes = 0;
        for (let i = 0; i < file.text.length; i++) {
            offsets[i] = bytes;
            const code = file.text.codePointAt(i);
            if (code > 0xffff) { offsets[++i] = bytes + 3; bytes += 4; }
            else bytes += code < 0x80 ? 1 : code < 0x800 ? 2 : 3;
            offsets[i+1] = bytes;
        }
        function walk(node) {
            if (ts.isCallExpression(node)) {
                const signature = checker.getResolvedSignature(node);
                const declaration = signature?.declaration;
                if (declaration) {
                    const symbol = declaration.name && checker.getSymbolAtLocation(declaration.name);
                    const targets = new Set([location(declaration), ...(symbol?.declarations || []).map(location)]);
                    for (const target of targets) {
                        if (!callees[target]) continue;
                        const key = local(file) + ':' + target;
                        if (!callsByCallee.has(key)) callsByCallee.set(key, []);
                        const result = checker.getReturnTypeOfSignature(signature);
                        callsByCallee.get(key).push({where: location(node), expression: node.expression.getText(),
                            start: offsets[node.getStart(file)], end: offsets[node.end],
                            stock_resolved_return: typeText(result), stock_resolved_return_is_any: any(result),
                            any_arguments: node.arguments.filter(argument => any(checker.getTypeAtLocation(argument)))
                                .map(argument => ({where: location(argument), expression: argument.getText().slice(0,200), sources: traceAny(argument)}))});
                    }
                }
            }
            ts.forEachChild(node, walk);
        }
        walk(file);
    }
    for (const observation of observations) {
        observation.matching_calls = (callsByCallee.get('src/compiler/' + observation.file + ':' + observation.callee) || [])
            .filter(call => observation.start <= call.start && call.end <= observation.end);
        observation.call_evidence_limit = observation.matching_calls.length
            ? 'syntactic calls inside the failed span; runtime failing call/instantiation not recorded'
            : 'no direct matching call in span; signature preparation or transitive lowering may have failed';
    }
}
const diagnostics = ts.getPreEmitDiagnostics(program);
const externalManifest = program.getSourceFiles().filter(file => !file.fileName.startsWith(root + '/'))
    .map(file => ({file: local(file), sha256: crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex')}));
fs.writeFileSync(output, JSON.stringify({typescript: ts.version, options, diagnostics: diagnostics.length,
    diagnostic_codes: diagnostics.reduce((counts, d) => { counts[d.code] = (counts[d.code] || 0)+1; return counts; }, {}),
    external_manifest: externalManifest, callees, observations}, null, 2)+'\n');
console.log(`${observations.length} observations mapped to ${Object.keys(callees).length} callees; ${diagnostics.length} stock diagnostics`);
