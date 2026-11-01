'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// moduleNameResolver.ts, resolutionCache.ts and sourcemap.ts to zero own-file checker diagnostics (#7xw984y).
// Five kinds of edit, each found by its owner and shape in freshly parsed text, never by a line:
// truthful optional declarations (20's model) on six members, two required values, and one read of caller data
// in sourcemap.ts's appendSourceMap instead of two. README.md gives each invariant; sites.json records each site.
// Every site must be in its original shape or its adapted shape, and all sites in the same one: anything else
// is drift, and the adapter refuses it before writing.

// The members widened from `p?: T` to `p?: T | undefined`. Only `public` ones move typescript.d.ts.
const declarations = [
    {file: 'src/compiler/types.ts', owner: 'ResolvedModuleFull', member: 'originalPath', type: 'string', readonly: true, internal: true},
    {file: 'src/compiler/types.ts', owner: 'ResolvedModuleFull', member: 'packageId', type: 'PackageId', readonly: false, internal: false},
    {file: 'src/compiler/types.ts', owner: 'ResolvedTypeReferenceDirective', member: 'originalPath', type: 'string', readonly: false, internal: true},
    {file: 'src/compiler/types.ts', owner: 'ResolvedTypeReferenceDirective', member: 'packageId', type: 'PackageId', readonly: false, internal: false},
    {file: 'src/compiler/program.ts', owner: 'ResolutionWithResolvedFileName', member: 'packageId', type: 'PackageId', readonly: false, internal: true},
    {file: 'src/compiler/resolutionCache.ts', owner: 'ResolutionWithFailedLookupLocations', member: 'files', type: 'Set<Path>', readonly: false, internal: true},
];
const resolver = 'src/compiler/moduleNameResolver.ts';
const sourcemap = 'src/compiler/sourcemap.ts';
const originalGuard = 'map.sourcesContent && typeof map.sourcesContent[raw.sourceIndex] === "string"';
const originalPayload = 'setSourceContent(newSourceIndex, map.sourcesContent[raw.sourceIndex]);';

function nodes(root, predicate) {
    const found = [];
    (function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); })(root);
    return found;
}
function one(found, what) {
    assert.equal(found.length, 1, `${what}: found ${found.length}, want 1`);
    return found[0];
}
function parse(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${fileName}`);
    return source;
}
function functionNamed(source, name) {
    return one(nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === name && n.body !== undefined), `function ${name}`);
}
function isInternal(source, node) {
    return (ts.getJSDocTags(node).some(tag => tag.tagName.text === 'internal')) ||
        /@internal/.test(source.text.slice(node.getFullStart(), node.getStart(source)));
}

// Each planner returns {state: 'before' | 'after', edits: [{at, end, text}], site}.
function planDeclaration(source, row) {
    const owners = nodes(source, n => ts.isInterfaceDeclaration(n) && n.name.text === row.owner);
    const owner = one(owners, `interface ${row.owner}`);
    const member = one(owner.members.filter(m => m.name && ts.isIdentifier(m.name) && m.name.text === row.member), `${row.owner}.${row.member}`);
    assert(ts.isPropertySignature(member) && member.questionToken && member.type, `${row.owner}.${row.member} is not an optional property signature`);
    const readonly = (member.modifiers ?? []).some(m => m.kind === ts.SyntaxKind.ReadonlyKeyword);
    assert.equal(readonly, row.readonly, `${row.owner}.${row.member} readonly changed`);
    // A member is internal when it or its interface is marked; only a public member reaches typescript.d.ts.
    assert.equal(isInternal(source, member) || isInternal(source, owner), row.internal, `${row.owner}.${row.member} @internal changed`);
    const type = member.type;
    const site = {file: row.file, owner: row.owner, member: row.member};
    if (type.getText(source) === row.type) {
        return {state: 'before', edits: [{at: type.end, end: type.end, text: ' | undefined'}], site};
    }
    assert(ts.isUnionTypeNode(type) && type.types.length === 2 && type.types[0].getText(source) === row.type &&
        type.types[1].kind === ts.SyntaxKind.UndefinedKeyword, `${row.owner}.${row.member} has an uninventoried type: ${type.getText(source)}`);
    return {state: 'after', edits: [], site};
}

// Row 1: `typesVersions[key]` in getPackageJsonTypesVersionsPaths' returned literal, under for-in plus hasProperty.
function planTypesVersions(source) {
    const owner = functionNamed(source, 'getPackageJsonTypesVersionsPaths');
    const loop = one(nodes(owner, n => ts.isForInStatement(n)), 'for-in in getPackageJsonTypesVersionsPaths');
    assert.equal(loop.initializer.getText(source), 'const key');
    assert.equal(loop.expression.getText(source), 'typesVersions');
    assert(ts.isBlock(loop.statement));
    assert.equal(loop.statement.statements[0]?.getText(source), 'if (!hasProperty(typesVersions, key)) continue;', 'hasProperty guard moved');
    const reads = nodes(owner, n => ts.isElementAccessExpression(n) && n.getText(source) === 'typesVersions[key]');
    const read = one(reads, 'typesVersions[key] in getPackageJsonTypesVersionsPaths');
    assert(loop.statement.statements.some(s => s.pos <= read.pos && read.end <= s.end), 'read left the loop');
    const site = {file: resolver, function: 'getPackageJsonTypesVersionsPaths', expression: 'typesVersions[key]'};
    if (ts.isNonNullExpression(read.parent)) {
        assert(ts.isPropertyAssignment(read.parent.parent) && read.parent.parent.name.getText(source) === 'paths');
        return {state: 'after', edits: [], site};
    }
    assert(ts.isPropertyAssignment(read.parent) && read.parent.name.getText(source) === 'paths', 'read is not the paths value');
    return {state: 'before', edits: [{at: read.end, end: read.end, text: '!'}], site};
}

// Row 2 (iii): the literal's isExternalLibraryImport, built only inside `resolved && { ... }`.
function planExternalLibraryImport(source) {
    const owner = functionNamed(source, 'createResolvedModuleWithFailedLookupLocations');
    const parameter = one(owner.parameters.filter(p => p.name.getText(source) === 'isExternalLibraryImport'), 'isExternalLibraryImport parameter');
    assert.equal(parameter.type?.getText(source), 'boolean | undefined');
    const resolvedParameter = one(owner.parameters.filter(p => p.name.getText(source) === 'resolved'), 'resolved parameter');
    assert.equal(resolvedParameter.type?.getText(source), 'Resolved | undefined');
    const property = one(nodes(owner, n => ts.isPropertyAssignment(n) && n.name.getText(source) === 'resolvedModule'), 'resolvedModule property');
    const guard = property.initializer;
    assert(ts.isBinaryExpression(guard) && guard.operatorToken.kind === ts.SyntaxKind.AmpersandAmpersandToken &&
        guard.left.getText(source) === 'resolved' && ts.isObjectLiteralExpression(guard.right), 'literal is no longer built under resolved &&');
    const entries = guard.right.properties.filter(p => p.name?.getText(source) === 'isExternalLibraryImport');
    const entry = one(entries, 'isExternalLibraryImport in the resolvedModule literal');
    // Every other write of the parameter would invalidate the callers' proof.
    const writes = nodes(owner, n => ts.isBinaryExpression(n) && ts.isIdentifier(n.left) && n.left.text === 'isExternalLibraryImport' &&
        n.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && n.operatorToken.kind <= ts.SyntaxKind.LastAssignment);
    assert.equal(writes.length, 0, 'isExternalLibraryImport is reassigned');
    const site = {file: resolver, function: 'createResolvedModuleWithFailedLookupLocations', expression: 'isExternalLibraryImport'};
    if (ts.isShorthandPropertyAssignment(entry)) {
        return {state: 'before', edits: [{at: entry.getStart(source), end: entry.end, text: 'isExternalLibraryImport: isExternalLibraryImport!'}], site};
    }
    assert(ts.isPropertyAssignment(entry) && entry.initializer.getText(source) === 'isExternalLibraryImport!', 'uninventoried isExternalLibraryImport entry');
    return {state: 'after', edits: [], site};
}

// Every caller of the producer, by callee name, with the exact isExternalLibraryImport argument it passes.
const callers = [
    {function: 'nodeModuleNameResolverWorker', callee: 'createResolvedModuleWithFailedLookupLocationsHandlingSymlink', resolved: 'result?.value?.resolved', external: 'result?.value?.isExternalLibraryImport'},
    {function: 'classicNameResolver', callee: 'createResolvedModuleWithFailedLookupLocationsHandlingSymlink', resolved: 'resolved && resolved.value', external: 'resolved?.value && pathContainsNodeModules(resolved.value.path)'},
    {function: 'loadModuleFromGlobalCache', callee: 'createResolvedModuleWithFailedLookupLocations', resolved: 'resolved', external: '/*isExternalLibraryImport*/ true'},
    {function: 'createResolvedModuleWithFailedLookupLocationsHandlingSymlink', callee: 'createResolvedModuleWithFailedLookupLocations', resolved: 'resolved', external: 'isExternalLibraryImport'},
];
function checkCallers(source) {
    const calls = nodes(source, n => ts.isCallExpression(n) && ts.isIdentifier(n.expression) &&
        ['createResolvedModuleWithFailedLookupLocations', 'createResolvedModuleWithFailedLookupLocationsHandlingSymlink'].includes(n.expression.text));
    assert.equal(calls.length, callers.length, `producer has ${calls.length} callers, want ${callers.length}`);
    for (const caller of callers) {
        const owner = functionNamed(source, caller.function);
        const call = one(calls.filter(c => c.expression.text === caller.callee && owner.pos <= c.pos && c.end <= owner.end), `${caller.function} calling ${caller.callee}`);
        const offset = caller.callee.endsWith('HandlingSymlink') ? 1 : 0;
        assert.equal(call.arguments[offset].getText(source), caller.resolved, `${caller.function}'s resolved argument moved`);
        const external = call.arguments[offset + 1];
        assert.equal(source.text.slice(external.getFullStart(), external.end).trim(), caller.external, `${caller.function}'s isExternalLibraryImport argument moved`);
    }
    // The worker's tryResolve declares the value type the worker passes on.
    const worker = functionNamed(source, 'nodeModuleNameResolverWorker');
    const tryResolve = one(nodes(worker, n => ts.isFunctionDeclaration(n) && n.name?.text === 'tryResolve'), 'worker tryResolve');
    // Upstream spells it `SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>`; an earlier
    // adaptation may mark the members readonly. Either way the member must be required and exactly boolean.
    const returned = tryResolve.type;
    assert(returned && ts.isTypeReferenceNode(returned) && returned.typeName.getText(source) === 'SearchResult' &&
        returned.typeArguments?.length === 1 && ts.isTypeLiteralNode(returned.typeArguments[0]), 'worker tryResolve return type moved');
    const members = returned.typeArguments[0].members;
    assert.equal(members.length, 2);
    for (const [name, type] of [['resolved', 'Resolved'], ['isExternalLibraryImport', 'boolean']]) {
        const member = one(members.filter(m => m.name?.getText(source) === name), `tryResolve value member ${name}`);
        assert(ts.isPropertySignature(member) && !member.questionToken && member.type?.getText(source) === type, `tryResolve value member ${name} is not a required ${type}`);
    }
    return callers.length;
}

// Row 4: one read of the caller's sourcesContent element instead of two.
function planSourcesContent(source, newline) {
    const owner = one(nodes(source, n => (ts.isFunctionDeclaration(n) || ts.isFunctionExpression(n)) && n.name?.text === 'appendSourceMap'), 'appendSourceMap');
    const declared = nodes(owner, n => (ts.isVariableDeclaration(n) || ts.isParameter(n)) && n.name.getText(source) === 'content');
    const site = {file: sourcemap, function: 'appendSourceMap', expression: 'map.sourcesContent[raw.sourceIndex]'};
    const before = nodes(owner, n => ts.isIfStatement(n) && n.expression.getText(source) === originalGuard);
    const after = nodes(owner, n => ts.isIfStatement(n) && n.expression.getText(source) === 'map.sourcesContent');
    assert.equal(before.length + after.length, 1, 'sourcesContent guard: want exactly one, original or adapted');
    if (before.length) {
        assert.equal(declared.length, 0, '`content` is already declared in appendSourceMap');
        const statement = before[0];
        assert(ts.isBlock(statement.thenStatement) && statement.elseStatement === undefined);
        assert.equal(statement.thenStatement.statements.length, 1);
        assert.equal(statement.thenStatement.statements[0].getText(source), originalPayload);
        assert.equal(nodes(owner, n => ts.isElementAccessExpression(n) && n.getText(source) === 'map.sourcesContent[raw.sourceIndex]').length, 2);
        const start = statement.getStart(source);
        const line = source.text.lastIndexOf('\n', start) + 1;
        const indent = source.text.slice(line, start);
        assert.match(indent, /^ *$/);
        const inner = indent + '    ';
        const text = [
            'if (map.sourcesContent) {',
            inner + 'const content = map.sourcesContent[raw.sourceIndex];',
            inner + 'if (typeof content === "string") {',
            inner + '    setSourceContent(newSourceIndex, content);',
            inner + '}',
            indent + '}',
        ].join(newline);
        return {state: 'before', edits: [{at: start, end: statement.end, text}], site};
    }
    const statement = after[0];
    assert(ts.isBlock(statement.thenStatement) && statement.elseStatement === undefined);
    const [read, test] = statement.thenStatement.statements;
    assert.equal(statement.thenStatement.statements.length, 2);
    assert.equal(read.getText(source), 'const content = map.sourcesContent[raw.sourceIndex];');
    assert(ts.isIfStatement(test) && test.expression.getText(source) === 'typeof content === "string"' && test.elseStatement === undefined);
    assert(ts.isBlock(test.thenStatement) && test.thenStatement.statements.length === 1 &&
        test.thenStatement.statements[0].getText(source) === 'setSourceContent(newSourceIndex, content);');
    assert.equal(declared.length, 1, '`content` is declared elsewhere in appendSourceMap');
    return {state: 'after', edits: [], site};
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const files = new Map();
    function file(relative) {
        if (!files.has(relative)) {
            const fileName = path.join(tree, relative);
            const text = fs.readFileSync(fileName, 'utf8');
            files.set(relative, {fileName, text, source: parse(fileName, text), edits: []});
        }
        return files.get(relative);
    }
    const plans = [];
    for (const row of declarations) plans.push({kind: 'declaration', ...planDeclaration(file(row.file).source, row)});
    plans.push({kind: 'required-read', ...planTypesVersions(file(resolver).source)});
    const callerCount = checkCallers(file(resolver).source);
    plans.push({kind: 'required-value', ...planExternalLibraryImport(file(resolver).source)});
    const sourcemapFile = file(sourcemap);
    const newline = sourcemapFile.text.includes('\r\n') ? '\r\n' : '\n';
    plans.push({kind: 'single-read', ...planSourcesContent(sourcemapFile.source, newline)});
    const states = new Set(plans.map(plan => plan.state));
    assert.equal(states.size, 1, `partly adapted: ${plans.map(plan => `${plan.site.owner ?? plan.site.function}.${plan.site.member ?? plan.site.expression}=${plan.state}`).join(', ')}`);
    for (const plan of plans) for (const edit of plan.edits) file(plan.site.file).edits.push(edit);
    const outputs = [];
    for (const [relative, entry] of files) {
        if (!entry.edits.length) continue;
        let result = entry.text;
        for (const edit of entry.edits.sort((a, b) => b.at - a.at)) result = result.slice(0, edit.at) + edit.text + result.slice(edit.end);
        parse(entry.fileName, result);
        outputs.push({relative, entry, result});
    }
    // Validate every file before writing any.
    for (const {entry} of outputs) assert.equal(fs.readFileSync(entry.fileName, 'utf8'), entry.text, 'source changed while planning');
    for (const {entry, result} of outputs) fs.writeFileSync(entry.fileName, result);
    return {
        state: [...states][0],
        files: outputs.map(output => output.relative),
        edits: plans.reduce((sum, plan) => sum + plan.edits.length, 0),
        callers: callerCount,
        javascriptChanging: true,
        sites: plans.map(plan => ({kind: plan.kind, ...plan.site})),
    };
}
module.exports = {adapt, declarations, callers};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
