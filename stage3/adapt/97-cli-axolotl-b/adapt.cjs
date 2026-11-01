'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// tsc stores possibly-undefined values, mostly maybeBind(host, host.x), into internal optional host members declared x?(…): R (#qg1ry2t).
// Each declaration says what tsc stores: a method becomes adaptation 32's indexed method type, which keeps method
// parameter bivariance, plus undefined; a property gains | undefined. Reads of x? see undefined either way (README.md).
// 11 internal declarations of the inventory, each pinned by its own restore mutant (README.md). I10 is 97c's; I12, I13 and I14
// follow the public P18 and P19 and pin nothing until those widen, so they travel with 97b-public. An array spells a multi-line member, joined with the file's own line breaks.
// internal: 'Member' (its own @internal), 'Owner' (its interface's @internal), 'Local' (a parameter of a function the file doesn't export).
const targets = [
    {id: 'I1', file: 'types.ts', owner: 'CompilerHost', member: 'resolveLibrary', internal: 'Member',
        before: ['resolveLibrary?(', '        libraryName: string,', '        resolveFrom: string,', '        options: CompilerOptions,', '        libFileName: string,', '    ): ResolvedModuleWithFailedLookupLocations;'],
        after: ['resolveLibrary?: { resolveLibrary(', '        libraryName: string,', '        resolveFrom: string,', '        options: CompilerOptions,', '        libFileName: string,', '    ): ResolvedModuleWithFailedLookupLocations; }["resolveLibrary"] | undefined;']},
    {id: 'I2', file: 'types.ts', owner: 'CompilerHost', member: 'useSourceOfProjectReferenceRedirect', internal: 'Member',
        before: 'useSourceOfProjectReferenceRedirect?(): boolean;',
        after: 'useSourceOfProjectReferenceRedirect?: { useSourceOfProjectReferenceRedirect(): boolean; }["useSourceOfProjectReferenceRedirect"] | undefined;'},
    {id: 'I3', file: 'types.ts', owner: 'CompilerHost', member: 'createDirectory', internal: 'Member',
        before: 'createDirectory?(directory: string): void;',
        after: 'createDirectory?: { createDirectory(directory: string): void; }["createDirectory"] | undefined;'},
    {id: 'I4', file: 'types.ts', owner: 'EmitHost', member: 'createHash', internal: 'Owner',
        before: 'createHash?(data: string): string;',
        after: 'createHash?: { createHash(data: string): string; }["createHash"] | undefined;'},
    {id: 'I5', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'directoryExists', internal: 'Owner',
        before: 'directoryExists?(path: string): boolean;',
        after: 'directoryExists?: { directoryExists(path: string): boolean; }["directoryExists"] | undefined;'},
    {id: 'I6', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'realpath', internal: 'Owner',
        before: 'realpath?(path: string): string;',
        after: 'realpath?: { realpath(path: string): string; }["realpath"] | undefined;'},
    {id: 'I7', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'getGlobalTypingsCacheLocation', internal: 'Owner',
        before: 'getGlobalTypingsCacheLocation?(): string | undefined;',
        after: 'getGlobalTypingsCacheLocation?: { getGlobalTypingsCacheLocation(): string | undefined; }["getGlobalTypingsCacheLocation"] | undefined;'},
    {id: 'I8', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'trace', internal: 'Owner',
        before: 'trace?(s: string): void;',
        after: 'trace?: { trace(s: string): void; }["trace"] | undefined;'},
    {id: 'I9', file: 'tsbuildPublic.ts', owner: 'SolutionBuilderHostBase', member: 'now', internal: 'Member',
        before: 'now?(): Date;',
        after: 'now?: { now(): Date; }["now"] | undefined;'},
    // CompilerHostLikeForCache takes I3's CompilerHost, so it must admit I3's stored undefined too.
    {id: 'I11', file: 'program.ts', owner: 'CompilerHostLikeForCache', member: 'createDirectory', internal: 'Owner',
        before: 'createDirectory?(directory: string): void;',
        after: 'createDirectory?: { createDirectory(directory: string): void; }["createDirectory"] | undefined;'},
    // getCurrentTime(host: { now?(): Date; }) takes every SolutionBuilderHostBase, so it must admit I9's stored undefined too.
    {id: 'I15', file: 'tsbuildPublic.ts', owner: 'getCurrentTime(host)', member: 'now', internal: 'Local',
        before: 'now?(): Date;',
        after: 'now?: { now(): Date; }["now"] | undefined;'},
];
const files = [...new Set(targets.map(t => t.file))];
// The JavaScript must not move: these are type-only edits, and every edited file is transpiled before and after to prove it.
const javascriptOptions = {target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS};

function nodes(source, test) {
    const found = [];
    (function visit(node) {
        if (test(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(source);
    return found;
}

function internalTag(node) {
    return ts.getJSDocTags(node).some(tag => tag.tagName.text === 'internal');
}

function exported(node) {
    return !!node.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword);
}

// The members list a target lives in: a top-level interface by name, or I15's parameter type literal.
function ownerMembers(source, target) {
    if (target.internal === 'Local') {
        const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getCurrentTime');
        assert.equal(functions.length, 1, `${target.file}: ${functions.length} getCurrentTime functions, not one`);
        const owner = functions[0];
        assert(ts.isSourceFile(owner.parent) && !exported(owner), `${target.file}: getCurrentTime is no longer a local function`);
        assert.equal(owner.parameters.length, 1, `${target.file}: getCurrentTime's parameters changed`);
        const parameter = owner.parameters[0];
        assert(ts.isIdentifier(parameter.name) && parameter.name.text === 'host', `${target.file}: getCurrentTime's parameter is not host`);
        assert(parameter.type && ts.isTypeLiteralNode(parameter.type), `${target.file}: getCurrentTime's host is not a type literal`);
        assert.equal(parameter.type.members.length, 1, `${target.file}: getCurrentTime's host type grew or shrank`);
        return parameter.type.members;
    }
    const owners = nodes(source, n => ts.isInterfaceDeclaration(n) && n.name.text === target.owner);
    assert.equal(owners.length, 1, `${target.file}: ${owners.length} ${target.owner} interfaces, not one`);
    const owner = owners[0];
    assert(ts.isSourceFile(owner.parent) && exported(owner), `${target.file}: ${target.owner} is no longer a top-level exported interface`);
    if (target.internal === 'Owner') assert(internalTag(owner), `${target.file}: ${target.owner} is no longer @internal`);
    return owner.members;
}

// Finds each target of one file and names its shape, 'Before' or 'After'. Anything else is not inventoried.
function locate(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
    // A multi-line member is spelled with the file's own line breaks, so the file must use one kind throughout.
    const crlf = text.split('\r\n').length - 1;
    const lf = text.split('\n').length - 1;
    assert(crlf === 0 || crlf === lf, `${file} mixes line breaks`);
    const newline = crlf ? '\r\n' : '\n';
    return targets.filter(t => t.file === file).map(target => {
        const members = ownerMembers(source, target).filter(m => m.name && ts.isIdentifier(m.name) && m.name.text === target.member);
        assert.equal(members.length, 1, `${file}: ${members.length} ${target.owner}.${target.member} members, not one`);
        const member = members[0];
        if (target.internal === 'Member') assert(internalTag(member), `${file}: ${target.owner}.${target.member} is no longer @internal`);
        const before = [target.before].flat().join(newline);
        const after = [target.after].flat().join(newline);
        const written = member.getText(source);
        const shape = written === before ? 'Before' : written === after ? 'After' : undefined;
        assert(shape, `${file}: ${target.id} ${target.owner}.${target.member} is ${JSON.stringify(written)}, not an inventoried shape`);
        return {target, shape, start: member.getStart(source), end: member.end, before, after};
    });
}

function javascript(file, text) {
    return ts.transpileModule(text, {compilerOptions: javascriptOptions, fileName: file, reportDiagnostics: false}).outputText;
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = files.map(file => {
        const fileName = path.join(tree, 'src/compiler', file);
        const text = fs.readFileSync(fileName, 'utf8');
        const sites = locate(file, text);
        // Either untouched or already adapted, never half: a partial shape is drift.
        const adapted = sites.filter(s => s.shape === 'After').map(s => s.target.id);
        assert(adapted.length === 0 || adapted.length === sites.length, `${file} is partly adapted: ${adapted.join(', ')} of ${sites.map(s => s.target.id).join(', ')}`);
        const edits = sites.filter(s => s.shape === 'Before');
        let result = text;
        for (const edit of [...edits].sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.after + result.slice(edit.end);
        assert(locate(file, result).every(s => s.shape === 'After'), `the planned edits do not adapt ${file}`);
        assert.equal(result.length - text.length, edits.reduce((sum, e) => sum + e.after.length - e.before.length, 0), `the planned edits touch other bytes of ${file}`);
        const bytes = edits.length ? javascript(file, text).length : undefined;
        if (edits.length) assert.equal(javascript(file, result), javascript(file, text), `${file}'s JavaScript changes`);
        return {file, fileName, text, result, edits, adapted, bytes};
    });
    // The head is one unit: a tree with some files adapted and others not is drift too.
    const done = plans.filter(p => p.adapted.length);
    assert(done.length === 0 || done.length === plans.length, `the tree is partly adapted: ${done.map(p => p.file).join(', ')} done, ${plans.filter(p => !p.adapted.length).map(p => p.file).join(', ')} not`);
    for (const plan of plans) assert.equal(fs.readFileSync(plan.fileName, 'utf8'), plan.text, `${plan.file} changed while planning`);
    for (const plan of plans) if (plan.edits.length) fs.writeFileSync(plan.fileName, plan.result);
    const edited = plans.filter(p => p.edits.length);
    return {
        files: edited.length,
        edits: edited.reduce((sum, p) => sum + p.edits.length, 0),
        perFile: Object.fromEntries(edited.map(p => [p.file, {declarations: p.edits.map(e => e.target.id), javascriptBytes: p.bytes}])),
        javascriptChanging: false,
        publicApiChanging: false,
    };
}
module.exports = {adapt, targets};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
