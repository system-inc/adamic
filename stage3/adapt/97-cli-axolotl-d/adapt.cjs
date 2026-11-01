'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// tsc stores possibly-undefined values, mostly maybeBind(host, host.x), into public optional host methods declared x?(…): R (#qg1ry2t).
// Each becomes adaptation 32's indexed method type plus undefined, which keeps method parameter bivariance, under Adamic's
// Nov 1 ruling on #axochk for these 19 by name. I12, I13 and I14 are the internal cascade owners those 19 need (README.md).
// scope: 'Public' (an exported interface and a member, neither @internal), 'Member' (its own @internal), 'Owner' (its interface's @internal).
// An array spells a multi-line member, joined with the file's own line breaks.
const targets = [
    {id: 'P1', file: 'types.ts', owner: 'ModuleResolutionHost', member: 'trace', scope: 'Public',
        before: 'trace?(s: string): void;',
        after: 'trace?: { trace(s: string): void; }["trace"] | undefined;'},
    {id: 'P2', file: 'types.ts', owner: 'ModuleResolutionHost', member: 'directoryExists', scope: 'Public',
        before: 'directoryExists?(directoryName: string): boolean;',
        after: 'directoryExists?: { directoryExists(directoryName: string): boolean; }["directoryExists"] | undefined;'},
    {id: 'P3', file: 'types.ts', owner: 'ModuleResolutionHost', member: 'realpath', scope: 'Public',
        before: 'realpath?(path: string): string;',
        after: 'realpath?: { realpath(path: string): string; }["realpath"] | undefined;'},
    {id: 'P4', file: 'types.ts', owner: 'ModuleResolutionHost', member: 'getDirectories', scope: 'Public',
        before: 'getDirectories?(path: string): string[];',
        after: 'getDirectories?: { getDirectories(path: string): string[]; }["getDirectories"] | undefined;'},
    {id: 'P5', file: 'types.ts', owner: 'ParseConfigHost', member: 'trace', scope: 'Public',
        before: 'trace?(s: string): void;',
        after: 'trace?: { trace(s: string): void; }["trace"] | undefined;'},
    {id: 'P6', file: 'types.ts', owner: 'CompilerHost', member: 'getDefaultLibLocation', scope: 'Public',
        before: 'getDefaultLibLocation?(): string;',
        after: 'getDefaultLibLocation?: { getDefaultLibLocation(): string; }["getDefaultLibLocation"] | undefined;'},
    {id: 'P7', file: 'types.ts', owner: 'CompilerHost', member: 'readDirectory', scope: 'Public',
        before: 'readDirectory?(rootDir: string, extensions: readonly string[], excludes: readonly string[] | undefined, includes: readonly string[], depth?: number): string[];',
        after: 'readDirectory?: { readDirectory(rootDir: string, extensions: readonly string[], excludes: readonly string[] | undefined, includes: readonly string[], depth?: number): string[]; }["readDirectory"] | undefined;'},
    {id: 'P8', file: 'types.ts', owner: 'CompilerHost', member: 'resolveModuleNames', scope: 'Public',
        before: 'resolveModuleNames?(moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile): (ResolvedModule | undefined)[];',
        after: 'resolveModuleNames?: { resolveModuleNames(moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile): (ResolvedModule | undefined)[]; }["resolveModuleNames"] | undefined;'},
    {id: 'P9', file: 'types.ts', owner: 'CompilerHost', member: 'getModuleResolutionCache', scope: 'Public',
        before: 'getModuleResolutionCache?(): ModuleResolutionCache | undefined;',
        after: 'getModuleResolutionCache?: { getModuleResolutionCache(): ModuleResolutionCache | undefined; }["getModuleResolutionCache"] | undefined;'},
    {id: 'P10', file: 'types.ts', owner: 'CompilerHost', member: 'resolveTypeReferenceDirectives', scope: 'Public',
        before: 'resolveTypeReferenceDirectives?(typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode): (ResolvedTypeReferenceDirective | undefined)[];',
        after: 'resolveTypeReferenceDirectives?: { resolveTypeReferenceDirectives(typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode): (ResolvedTypeReferenceDirective | undefined)[]; }["resolveTypeReferenceDirectives"] | undefined;'},
    {id: 'P11', file: 'types.ts', owner: 'CompilerHost', member: 'resolveModuleNameLiterals', scope: 'Public',
        before: ['resolveModuleNameLiterals?(', '        moduleLiterals: readonly StringLiteralLike[],', '        containingFile: string,', '        redirectedReference: ResolvedProjectReference | undefined,', '        options: CompilerOptions,', '        containingSourceFile: SourceFile,', '        reusedNames: readonly StringLiteralLike[] | undefined,', '    ): readonly ResolvedModuleWithFailedLookupLocations[];'],
        after: ['resolveModuleNameLiterals?: { resolveModuleNameLiterals(', '        moduleLiterals: readonly StringLiteralLike[],', '        containingFile: string,', '        redirectedReference: ResolvedProjectReference | undefined,', '        options: CompilerOptions,', '        containingSourceFile: SourceFile,', '        reusedNames: readonly StringLiteralLike[] | undefined,', '    ): readonly ResolvedModuleWithFailedLookupLocations[]; }["resolveModuleNameLiterals"] | undefined;']},
    {id: 'P12', file: 'types.ts', owner: 'CompilerHost', member: 'resolveTypeReferenceDirectiveReferences', scope: 'Public',
        before: ['resolveTypeReferenceDirectiveReferences?<T extends FileReference | string>(', '        typeDirectiveReferences: readonly T[],', '        containingFile: string,', '        redirectedReference: ResolvedProjectReference | undefined,', '        options: CompilerOptions,', '        containingSourceFile: SourceFile | undefined,', '        reusedNames: readonly T[] | undefined,', '    ): readonly ResolvedTypeReferenceDirectiveWithFailedLookupLocations[];'],
        after: ['resolveTypeReferenceDirectiveReferences?: { resolveTypeReferenceDirectiveReferences<T extends FileReference | string>(', '        typeDirectiveReferences: readonly T[],', '        containingFile: string,', '        redirectedReference: ResolvedProjectReference | undefined,', '        options: CompilerOptions,', '        containingSourceFile: SourceFile | undefined,', '        reusedNames: readonly T[] | undefined,', '    ): readonly ResolvedTypeReferenceDirectiveWithFailedLookupLocations[]; }["resolveTypeReferenceDirectiveReferences"] | undefined;']},
    {id: 'P13', file: 'types.ts', owner: 'CompilerHost', member: 'createHash', scope: 'Public',
        before: 'createHash?(data: string): string;',
        after: 'createHash?: { createHash(data: string): string; }["createHash"] | undefined;'},
    {id: 'P14', file: 'watchPublic.ts', owner: 'ProgramHost', member: 'createHash', scope: 'Public',
        before: 'createHash?(data: string): string;',
        after: 'createHash?: { createHash(data: string): string; }["createHash"] | undefined;'},
    {id: 'P15', file: 'watchPublic.ts', owner: 'ProgramHost', member: 'directoryExists', scope: 'Public',
        before: 'directoryExists?(path: string): boolean;',
        after: 'directoryExists?: { directoryExists(path: string): boolean; }["directoryExists"] | undefined;'},
    {id: 'P16', file: 'watchPublic.ts', owner: 'ProgramHost', member: 'realpath', scope: 'Public',
        before: 'realpath?(path: string): string;',
        after: 'realpath?: { realpath(path: string): string; }["realpath"] | undefined;'},
    {id: 'P17', file: 'watchPublic.ts', owner: 'ProgramHost', member: 'getEnvironmentVariable', scope: 'Public',
        before: 'getEnvironmentVariable?(name: string): string | undefined;',
        after: 'getEnvironmentVariable?: { getEnvironmentVariable(name: string): string | undefined; }["getEnvironmentVariable"] | undefined;'},
    {id: 'P18', file: 'tsbuildPublic.ts', owner: 'SolutionBuilderHostBase', member: 'createDirectory', scope: 'Public',
        before: 'createDirectory?(path: string): void;',
        after: 'createDirectory?: { createDirectory(path: string): void; }["createDirectory"] | undefined;'},
    {id: 'P19', file: 'tsbuildPublic.ts', owner: 'SolutionBuilderHostBase', member: 'writeFile', scope: 'Public',
        before: 'writeFile?(path: string, data: string, writeByteOrderMark?: boolean): void;',
        after: 'writeFile?: { writeFile(path: string, data: string, writeByteOrderMark?: boolean): void; }["writeFile"] | undefined;'},
    {id: 'I12', file: 'program.ts', owner: 'CompilerHostLikeForCache', member: 'writeFile', scope: 'Owner',
        before: 'writeFile?: WriteFileCallback;',
        after: 'writeFile?: WriteFileCallback | undefined;'},
    {id: 'I13', file: 'watchPublic.ts', owner: 'ProgramHost', member: 'createDirectory', scope: 'Member',
        before: 'createDirectory?(path: string): void;',
        after: 'createDirectory?: { createDirectory(path: string): void; }["createDirectory"] | undefined;'},
    {id: 'I14', file: 'watchPublic.ts', owner: 'ProgramHost', member: 'writeFile', scope: 'Member',
        before: 'writeFile?(path: string, data: string, writeByteOrderMark?: boolean): void;',
        after: 'writeFile?: { writeFile(path: string, data: string, writeByteOrderMark?: boolean): void; }["writeFile"] | undefined;'},
];
// 97b's 11 run first, in number order: each must already read as 97b's after, or this head refuses (97b's README.md).
// scope 'Local' is I15, the parameter type literal of tsbuildPublic.ts's unexported getCurrentTime.
const prerequisites = [
    {id: 'I1', file: 'types.ts', owner: 'CompilerHost', member: 'resolveLibrary', scope: 'Member', after: ['resolveLibrary?: { resolveLibrary(', '        libraryName: string,', '        resolveFrom: string,', '        options: CompilerOptions,', '        libFileName: string,', '    ): ResolvedModuleWithFailedLookupLocations; }["resolveLibrary"] | undefined;']},
    {id: 'I2', file: 'types.ts', owner: 'CompilerHost', member: 'useSourceOfProjectReferenceRedirect', scope: 'Member', after: 'useSourceOfProjectReferenceRedirect?: { useSourceOfProjectReferenceRedirect(): boolean; }["useSourceOfProjectReferenceRedirect"] | undefined;'},
    {id: 'I3', file: 'types.ts', owner: 'CompilerHost', member: 'createDirectory', scope: 'Member', after: 'createDirectory?: { createDirectory(directory: string): void; }["createDirectory"] | undefined;'},
    {id: 'I4', file: 'types.ts', owner: 'EmitHost', member: 'createHash', scope: 'Owner', after: 'createHash?: { createHash(data: string): string; }["createHash"] | undefined;'},
    {id: 'I5', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'directoryExists', scope: 'Owner', after: 'directoryExists?: { directoryExists(path: string): boolean; }["directoryExists"] | undefined;'},
    {id: 'I6', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'realpath', scope: 'Owner', after: 'realpath?: { realpath(path: string): string; }["realpath"] | undefined;'},
    {id: 'I7', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'getGlobalTypingsCacheLocation', scope: 'Owner', after: 'getGlobalTypingsCacheLocation?: { getGlobalTypingsCacheLocation(): string | undefined; }["getGlobalTypingsCacheLocation"] | undefined;'},
    {id: 'I8', file: 'types.ts', owner: 'ModuleSpecifierResolutionHost', member: 'trace', scope: 'Owner', after: 'trace?: { trace(s: string): void; }["trace"] | undefined;'},
    {id: 'I9', file: 'tsbuildPublic.ts', owner: 'SolutionBuilderHostBase', member: 'now', scope: 'Member', after: 'now?: { now(): Date; }["now"] | undefined;'},
    {id: 'I11', file: 'program.ts', owner: 'CompilerHostLikeForCache', member: 'createDirectory', scope: 'Owner', after: 'createDirectory?: { createDirectory(directory: string): void; }["createDirectory"] | undefined;'},
    {id: 'I15', file: 'tsbuildPublic.ts', owner: 'getCurrentTime(host)', member: 'now', scope: 'Local', after: 'now?: { now(): Date; }["now"] | undefined;'},
];
const files = [...new Set(targets.map(t => t.file))];
// Prerequisites are checked while each edited file is read, so each must live in one of them.
assert(prerequisites.every(p => files.includes(p.file)), 'a 97b prerequisite lives outside the edited files');
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

// The members list a site lives in: a top-level exported interface by name, or I15's parameter type literal.
function ownerMembers(source, site) {
    if (site.scope === 'Local') {
        const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getCurrentTime');
        assert.equal(functions.length, 1, `${site.file}: ${functions.length} getCurrentTime functions, not one`);
        const owner = functions[0];
        assert(ts.isSourceFile(owner.parent) && !exported(owner), `${site.file}: getCurrentTime is no longer a local function`);
        assert.equal(owner.parameters.length, 1, `${site.file}: getCurrentTime's parameters changed`);
        const parameter = owner.parameters[0];
        assert(ts.isIdentifier(parameter.name) && parameter.name.text === 'host', `${site.file}: getCurrentTime's parameter is not host`);
        assert(parameter.type && ts.isTypeLiteralNode(parameter.type), `${site.file}: getCurrentTime's host is not a type literal`);
        assert.equal(parameter.type.members.length, 1, `${site.file}: getCurrentTime's host type grew or shrank`);
        return parameter.type.members;
    }
    const owners = nodes(source, n => ts.isInterfaceDeclaration(n) && n.name.text === site.owner);
    assert.equal(owners.length, 1, `${site.file}: ${owners.length} ${site.owner} interfaces, not one`);
    const owner = owners[0];
    assert(ts.isSourceFile(owner.parent) && exported(owner), `${site.file}: ${site.owner} is no longer a top-level exported interface`);
    if (site.scope === 'Owner') assert(internalTag(owner), `${site.file}: ${site.owner} is no longer @internal`);
    if (site.scope === 'Public') assert(!internalTag(owner), `${site.file}: ${site.owner} is @internal, not public`);
    return owner.members;
}

// Finds one site's member and names its shape, 'Before' or 'After'. Anything else is not inventoried.
function shapeOf(source, file, newline, site) {
    const members = ownerMembers(source, site).filter(m => m.name && ts.isIdentifier(m.name) && m.name.text === site.member);
    assert.equal(members.length, 1, `${file}: ${members.length} ${site.owner}.${site.member} members, not one`);
    const member = members[0];
    if (site.scope === 'Member') assert(internalTag(member), `${file}: ${site.owner}.${site.member} is no longer @internal`);
    if (site.scope === 'Public') assert(!internalTag(member), `${file}: ${site.owner}.${site.member} is @internal, not public`);
    const before = site.before === undefined ? undefined : [site.before].flat().join(newline);
    const after = [site.after].flat().join(newline);
    const written = member.getText(source);
    const shape = written === before ? 'Before' : written === after ? 'After' : undefined;
    // A prerequisite has no before: anything but 97b's after means 97b has not run.
    if (before === undefined) assert(shape, `${file}: 97b's ${site.id} ${site.owner}.${site.member} is not applied`);
    assert(shape, `${file}: ${site.id} ${site.owner}.${site.member} is ${JSON.stringify(written)}, not an inventoried shape`);
    return {target: site, shape, start: member.getStart(source), end: member.end, before, after};
}

function locate(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
    // A multi-line member is spelled with the file's own line breaks, so the file must use one kind throughout.
    const crlf = text.split('\r\n').length - 1;
    const lf = text.split('\n').length - 1;
    assert(crlf === 0 || crlf === lf, `${file} mixes line breaks`);
    const newline = crlf ? '\r\n' : '\n';
    for (const site of prerequisites.filter(p => p.file === file)) shapeOf(source, file, newline, site);
    return targets.filter(t => t.file === file).map(site => shapeOf(source, file, newline, site));
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
        publicApiChanging: true,
    };
}
module.exports = {adapt, targets, prerequisites};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
