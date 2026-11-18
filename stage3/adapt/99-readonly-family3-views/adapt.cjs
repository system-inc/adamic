#!/usr/bin/env node
'use strict';
// Adaptation 99-readonly-family3-views, #k8ed1f2, family 3's roads 1 and 2. Family 3 is the refusal "a value of type
// A seen as B, which can write X where Y is read". At 57 slots in 17 files of src/compiler and src/services, a slot
// tsc only reads becomes a read-only view (`readonly` on a property, `Readonly<T>` on a parameter, `ReadonlyMap`,
// `readonly T[]`, a readonly constraint), one parameter takes its true key type, and nine calls get explicit type
// arguments, so 83 of the family's sites clear. Adamic delegated the ruling on #k8ed1f2: readonly where tsc never
// writes, a shape earlier adaptations established. README.md says why each slot, which 25 sites were left out, and
// every write found to a touched slot. Type-only: the emitted JavaScript does not change. No declaration is public:
// each is local to a function, file-local, or `@internal` on itself or an enclosing declaration, so the published
// declaration files do not change and no api-sanctions.json is needed. Through the stock AST each slot is found by
// name, never by line, offset or whole-line text; the edit replaces a type node's span, inserts `readonly ` before a
// property's name, inserts `: <type>` after a binding's name, or inserts `<...>` after a callee. It refuses before any
// write unless every slot is found exactly once, holds upstream's text or exactly its adapted text, and is local,
// file-local or `@internal`, and a second run changes nothing. It sorts after every adaptation whose output text it
// reads, among them 99-correctness-supported-extensions, which pins getSupportedExtensions' whole overload group,
// 99-correctness-binding-views, which pins resolutionCache.ts's edited owners, and 99-correctness-receiver-printer,
// which pins createPrinter, so each pin reads its declarations as it expects. The three that sort after it,
// 99-readonly-inferred-bindings, 99-readonly-script-target-entries and 99-wasm-crypto-import, neither read nor pin
// any declaration it touches.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const builder = 'src/compiler/builder.ts';
const checker = 'src/compiler/checker.ts';
const commandLineParser = 'src/compiler/commandLineParser.ts';
const core = 'src/compiler/core.ts';
const emitter = 'src/compiler/emitter.ts';
const executeCommandLine = 'src/compiler/executeCommandLine.ts';
const moduleNameResolver = 'src/compiler/moduleNameResolver.ts';
const moduleSpecifiers = 'src/compiler/moduleSpecifiers.ts';
const program = 'src/compiler/program.ts';
const resolutionCache = 'src/compiler/resolutionCache.ts';
const tsbuildPublic = 'src/compiler/tsbuildPublic.ts';
const types = 'src/compiler/types.ts';
const utilities = 'src/compiler/utilities.ts';
const watch = 'src/compiler/watch.ts';
const watchPublic = 'src/compiler/watchPublic.ts';
const watchUtilities = 'src/compiler/watchUtilities.ts';
const stringCompletions = 'src/services/stringCompletions.ts';
// The extension tables of cause 13: the outer array was already readonly, and the inner ones become so too.
const extensionTable = { upstream: 'readonly Extension[][]', adapted: 'readonly (readonly Extension[])[]' };
const stringTable = { upstream: 'readonly string[][]', adapted: 'readonly (readonly string[])[]' };
const options = { upstream: 'CompilerOptions | BuildOptions', adapted: 'Readonly<CompilerOptions | BuildOptions>' };
const builderState = { upstream: 'BuilderProgramState', adapted: 'Readonly<BuilderProgramState>' };
const diagnostics = { upstream: 'Diagnostic[]', adapted: 'readonly Diagnostic[]' };
// Cause 8's nine calls, each by its first argument: the map's own value type, from SolutionBuilderState's fields, so
// `T` is no longer inferred as unknown from noopOnDelete.
const statePrunes = [
    ['state.configFileCache', 'ConfigFileCacheEntry'],
    ['state.projectStatus', 'UpToDateStatus'],
    ['state.builderPrograms', 'T'],
    ['state.diagnostics', 'readonly Diagnostic[]'],
    ['state.projectPendingBuild', 'ProgramUpdateLevel'],
    ['state.projectErrorsReported', 'true'],
    ['state.buildInfoCache', 'BuildInfoCacheEntry'],
    ['state.outputTimeStamps', 'Map<Path, Date>'],
    ['state.lastCachedPackageJsonLookups', 'Set<string> | undefined'],
];
// Each slot, addressed by name, with the cause it clears. A slot is one of:
// - a function's return type, named parameter's type or type parameter's constraint (`function`, and `parameter` or
//   `typeParameter`); an overload by its index among the `overloads` declarations of that name, which must be one
//   overload group;
// - an interface's or a type alias's property (`interface` or `typeAlias`, `property`), whose type changes
//   (`upstream`, `adapted`) or which gains `readonly` (`readonly: true`, its type `type`);
// - an interface's method's return type (`interface`, `method`);
// - a const binding's type in the function declaration nearest it (`within`, `variable`), its initializer starting
//   `starts`, with `upstream: null` where upstream writes no annotation;
// - a call in the function declaration nearest it (`within`, `call`), by its callee and its first argument's text,
//   which gains the type arguments `typeArguments`.
// The type text as upstream 6.0.3 writes it, and as this adaptation leaves it: any other text is not the slot the
// ruling covers.
const slots = [
    { cause: 4, file: builder, interface: 'ReusableBuilderProgramState', property: 'semanticDiagnosticsPerFile',
        upstream: 'Map<Path, readonly ReusableDiagnostic[] | readonly Diagnostic[]>',
        adapted: 'ReadonlyMap<Path, readonly ReusableDiagnostic[] | readonly Diagnostic[]>' },
    { cause: 4, file: builder, interface: 'ReusableBuilderProgramState', property: 'emitDiagnosticsPerFile',
        upstream: 'Map<Path, readonly ReusableDiagnostic[] | readonly Diagnostic[]> | undefined',
        adapted: 'ReadonlyMap<Path, readonly ReusableDiagnostic[] | readonly Diagnostic[]> | undefined' },
    { cause: 6, file: builder, function: 'assertSourceFileOkWithoutNextAffectedCall', parameter: 'state', ...builderState },
    { cause: 6, file: builder, function: 'removeSemanticDiagnosticsOf', parameter: 'state', ...builderState },
    { cause: 6, file: builder, function: 'isChangedSignature', parameter: 'state', ...builderState },
    { cause: 7, file: types, interface: 'SyntacticTypeNodeBuilderContext', property: 'tracker', readonly: true,
        type: 'Required<Pick<SymbolTracker, "reportInferenceFallback">>' },
    ...statePrunes.map(([argument, value]) => ({ cause: 8, file: tsbuildPublic, within: 'createStateBuildOrder',
        call: 'mutateMapSkippingNewValues', argument, typeArguments: `ResolvedConfigFilePath, ${value}` })),
    { cause: 12, file: executeCommandLine, function: 'updateReportDiagnostic', parameter: 'options', ...options },
    { cause: 12, file: executeCommandLine, function: 'shouldBePretty', parameter: 'options', ...options },
    { cause: 12, file: executeCommandLine, function: 'createReportErrorSummary', parameter: 'options', ...options },
    { cause: 12, file: executeCommandLine, function: 'createWatchStatusReporter', parameter: 'options', ...options },
    { cause: 13, file: utilities, within: null, variable: 'supportedTSExtensions', starts: '[[Extension.Ts,', ...extensionTable },
    { cause: 13, file: utilities, within: null, variable: 'supportedTSExtensionsWithJson',
        starts: '[...supportedTSExtensions,', ...extensionTable },
    { cause: 13, file: utilities, within: null, variable: 'allSupportedExtensions', starts: '[[Extension.Ts,', ...extensionTable },
    { cause: 13, file: utilities, within: null, variable: 'allSupportedExtensionsWithJson',
        starts: '[...allSupportedExtensions,', ...extensionTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensions', overload: 0, overloads: 3, ...extensionTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensions', overload: 1, overloads: 3, ...stringTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensions', overload: 2, overloads: 3, ...stringTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensionsWithJsonIfResolveJsonModule', overload: 0, overloads: 3,
        parameter: 'supportedExtensions', ...extensionTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensionsWithJsonIfResolveJsonModule', overload: 0, overloads: 3,
        ...extensionTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensionsWithJsonIfResolveJsonModule', overload: 1, overloads: 3,
        parameter: 'supportedExtensions', ...stringTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensionsWithJsonIfResolveJsonModule', overload: 1, overloads: 3,
        ...stringTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensionsWithJsonIfResolveJsonModule', overload: 2, overloads: 3,
        parameter: 'supportedExtensions', ...stringTable },
    { cause: 13, file: utilities, function: 'getSupportedExtensionsWithJsonIfResolveJsonModule', overload: 2, overloads: 3,
        ...stringTable },
    { cause: 13, file: commandLineParser, function: 'hasFileWithHigherPriorityExtension', parameter: 'extensions',
        ...stringTable },
    { cause: 13, file: commandLineParser, function: 'removeWildcardFilesWithLowerPriorityExtension',
        parameter: 'extensions', ...stringTable },
    // Not a site of its own: services' one reader of getSupportedExtensions' result, which returns it, so the readonly
    // inner arrays reach its two callers as they are (J-xa7rf9z).
    { cause: 13, file: stringCompletions, function: 'getSupportedExtensionsForModuleResolution', ...stringTable },
    { cause: 14, file: types, interface: 'TypeChecker', method: 'getDiagnostics', ...diagnostics },
    { cause: 14, file: checker, function: 'getDiagnostics', ...diagnostics },
    { cause: 14, file: checker, function: 'getDiagnosticsWorker', ...diagnostics },
    { cause: 15, file: builder, typeAlias: 'IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo', property: 'signature',
        readonly: true, type: 'string | false | undefined' },
    { cause: 16, file: watchUtilities, function: 'createCachedDirectoryStructureHost', parameter: 'host',
        upstream: 'DirectoryStructureHost', adapted: 'Readonly<DirectoryStructureHost>' },
    { cause: 16, file: program, function: 'parseConfigHostFromCompilerHostLike', parameter: 'host',
        upstream: '(CompilerHost | ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }',
        adapted: 'Readonly<(CompilerHost | ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }>' },
    { cause: 16, file: watchPublic, within: 'createWatchProgram', variable: 'directoryStructureHost',
        starts: 'cachedDirectoryStructureHost || host', upstream: 'DirectoryStructureHost',
        adapted: 'Readonly<DirectoryStructureHost>' },
    { cause: 17, file: checker, function: 'discriminateTypeByDiscriminableItems', parameter: 'discriminators',
        upstream: '(readonly [() => Type, __String])[]', adapted: 'readonly (readonly [() => Type, __String])[]' },
    // createDiagnosticCollection's implementation, the third of its getDiagnostics declarations: its overloads, which
    // callers see, stay as they are.
    { cause: 21, file: utilities, function: 'getDiagnostics', overload: 2, overloads: 3, ...diagnostics },
    { cause: 23, file: core, function: 'not', typeParameter: 'T', upstream: 'unknown[]', adapted: 'readonly unknown[]' },
    { cause: 25, file: moduleSpecifiers, function: 'tryGetModuleNameFromPaths', parameter: 'paths',
        upstream: 'MapLike<readonly string[]>', adapted: 'Readonly<MapLike<readonly string[]>>' },
    { cause: 26, file: watchUtilities, function: 'closeFileWatcherOf', typeParameter: 'T',
        upstream: '{ watcher: FileWatcher; }', adapted: '{ readonly watcher: FileWatcher; }' },
    // The true key type, not a readonly view: the body deletes from the map, which ReadonlyMap can't.
    { cause: 27, file: resolutionCache, function: 'removeResolutionsOfFileFromCache', parameter: 'cache',
        upstream: 'Map<string, ModeAwareCache<T>>', adapted: 'Map<Path, ModeAwareCache<T>>' },
    { cause: 28, file: checker, function: 'getElaborationElementForJsxChild', parameter: 'nameType',
        upstream: 'LiteralType', adapted: 'Readonly<LiteralType>' },
    { cause: 32, file: resolutionCache, interface: 'ResolutionCache', property: 'directoryWatchesOfFailedLookups',
        upstream: 'Map<string, DirectoryWatchesOfFailedLookup>', adapted: 'ReadonlyMap<string, DirectoryWatchesOfFailedLookup>' },
    { cause: 37, file: checker, function: 'combineIterationTypes', parameter: 'array',
        upstream: '(IterationTypes | undefined)[]', adapted: 'readonly (IterationTypes | undefined)[]' },
    { cause: 38, file: watch, function: 'createTabularErrorsDisplay', parameter: 'filesInError',
        upstream: '(ReportFileInError | undefined)[]', adapted: 'readonly (ReportFileInError | undefined)[]' },
    { cause: 42, file: emitter, function: 'commentWillEmitNewLine', parameter: 'node',
        upstream: 'CommentRange', adapted: 'Readonly<CommentRange>' },
    { cause: 44, file: moduleNameResolver, interface: 'ModuleResolutionState', property: 'host', readonly: true,
        type: 'ModuleResolutionHost' },
    { cause: 45, file: program, function: 'getReferencedFileLocation', parameter: 'ref',
        upstream: 'ReferencedFile', adapted: 'Readonly<ReferencedFile>' },
    { cause: 46, file: resolutionCache, function: 'invalidateResolutions', parameter: 'resolutions',
        upstream: 'Set<ResolutionWithFailedLookupLocations> | Map<string, ResolutionWithFailedLookupLocations> | undefined',
        adapted: 'ReadonlySet<ResolutionWithFailedLookupLocations> | ReadonlyMap<string, ResolutionWithFailedLookupLocations> | undefined' },
    { cause: 47, file: checker, within: 'getObjectTypeInstantiation', variable: 'allDeclarations',
        starts: 'type.objectFlags & (ObjectFlags.Reference | ObjectFlags.InstantiationExpressionType) ? [declaration] :',
        upstream: null, adapted: 'readonly Node[]' },
];
const nodes = (root, predicate) => {
    const found = [];
    ts.forEachChild(root, function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    });
    return found;
};
const one = (found, what) => {
    if (found.length !== 1) throw new Error(`expected one ${what}, found ${found.length}`);
    return found[0];
};
const name = slot => {
    if (slot.call) return `${slot.within} ${slot.call}(${slot.argument})`;
    if (slot.variable) return `${slot.within ?? 'module'} const ${slot.variable}`;
    if (slot.property) return `${slot.interface ?? slot.typeAlias}.${slot.property}`;
    if (slot.method) return `${slot.interface}.${slot.method} return`;
    const head = slot.overload === undefined ? slot.function : `${slot.function} overload ${slot.overload}`;
    return `${head} ${slot.parameter ?? (slot.typeParameter ? `<${slot.typeParameter}> constraint` : 'return')}`;
};
const functionLike = node => ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node)
    || ts.isMethodDeclaration(node) || ts.isConstructorDeclaration(node) || ts.isGetAccessorDeclaration(node)
    || ts.isSetAccessorDeclaration(node);
const enclosingFunction = node => {
    let outer = node.parent;
    while (outer && !functionLike(outer)) outer = outer.parent;
    return outer;
};
// The function declaration nearest a node, by name, or the module itself when `within` is null.
function inside(node, within) {
    const outer = enclosingFunction(node);
    if (within === null) return outer === undefined;
    return outer !== undefined && ts.isFunctionDeclaration(outer) && outer.name?.text === within;
}
// The declarations that can carry a JSDoc tag: an `@internal` on any of them, the slot's own or an enclosing one,
// keeps it out of the published declaration files.
const documented = node => ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node)
    || ts.isFunctionDeclaration(node) || ts.isVariableDeclaration(node) || ts.isVariableStatement(node)
    || ts.isPropertySignature(node) || ts.isMethodSignature(node) || ts.isModuleDeclaration(node)
    || ts.isClassDeclaration(node);
const internal = node => ts.getJSDocTags(node).some(tag => tag.tagName.text === 'internal');
const exported = node => node.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword) ?? false;
// Why a declaration can't reach typescript.d.ts: it is local to a function, it or a declaration enclosing it carries
// `@internal`, or it is a top-level statement of a module that does not export it. Anything else refuses.
function visibility(source, declaration, what) {
    const outer = enclosingFunction(declaration);
    if (outer) return `local to ${outer.name && ts.isIdentifier(outer.name) ? outer.name.text : 'a function'}`;
    for (let node = declaration; node && !ts.isSourceFile(node); node = node.parent)
        if (documented(node) && internal(node)) return '@internal';
    const statement = ts.isVariableDeclaration(declaration) ? declaration.parent.parent : declaration;
    if (statement.parent === source && ts.isExternalModule(source) && !exported(statement)) return 'file-local';
    throw new Error(`expected ${what} to be local to a function, file-local or @internal`);
}
// A function by name, or one overload of a group by its index: the group must be all the declarations of that name
// in the file, siblings in one scope.
function functionDeclaration(source, slot) {
    const found = nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === slot.function);
    const expected = slot.overloads ?? 1;
    const group = expected === 1 ? 'one function' : `an overload group of ${expected}`;
    if (found.length !== expected || found.some(node => node.parent !== found[0].parent))
        throw new Error(`expected ${group} ${slot.function} in one scope of ${slot.file}, found ${found.length}`);
    return found[slot.overload ?? 0];
}
// The slot's node: the declaration that owns it, for the visibility check, and what the edit addresses.
function find(source, slot) {
    const what = `${name(slot)} in ${slot.file}`;
    if (slot.call) {
        const call = one(nodes(source, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression)
            && node.expression.text === slot.call && node.arguments.length > 0
            && node.arguments[0].getText(source) === slot.argument), `call ${what}`);
        if (!inside(call, slot.within)) throw new Error(`expected ${what} to sit in ${slot.within}`);
        return { call };
    }
    if (slot.variable) {
        const declaration = one(nodes(source, node => ts.isVariableDeclaration(node) && ts.isIdentifier(node.name)
            && node.name.text === slot.variable && inside(node, slot.within)), `binding ${what}`);
        if (!ts.isVariableDeclarationList(declaration.parent)
            || (declaration.parent.flags & ts.NodeFlags.BlockScoped) !== ts.NodeFlags.Const
            || !declaration.initializer?.getText(source).startsWith(slot.starts))
            throw new Error(`expected ${what} to be a const starting ${slot.starts}: ${declaration.getText(source).split('\n')[0]}`);
        return { owner: declaration, binding: declaration, type: declaration.type };
    }
    if (slot.property || slot.method) {
        const container = slot.interface
            ? one(nodes(source, node => ts.isInterfaceDeclaration(node) && node.name.text === slot.interface),
                `interface ${slot.interface} in ${slot.file}`)
            : one(nodes(source, node => ts.isTypeAliasDeclaration(node) && node.name.text === slot.typeAlias),
                `type ${slot.typeAlias} in ${slot.file}`);
        // An interface's own members, or the members of any type literal inside the alias. A call or index
        // signature has no name, so the name is checked before it is read.
        const members = slot.interface ? container.members : nodes(container, ts.isTypeElement);
        const member = one(members.filter(member => member.name && ts.isIdentifier(member.name)
            && member.name.text === (slot.property ?? slot.method)
            && (slot.property ? ts.isPropertySignature(member) : ts.isMethodSignature(member))),
            `${slot.property ? 'property' : 'method'} ${what}`);
        return { owner: member, member, type: member.type };
    }
    const declaration = functionDeclaration(source, slot);
    let type = declaration.type;
    if (slot.parameter)
        type = one(declaration.parameters.filter(parameter => ts.isIdentifier(parameter.name)
            && parameter.name.text === slot.parameter), `parameter ${what}`).type;
    if (slot.typeParameter)
        type = one((declaration.typeParameters ?? []).filter(parameter => parameter.name.text === slot.typeParameter),
            `type parameter ${what}`).constraint;
    return { owner: declaration, type };
}
// A slot whose type node is replaced whole: upstream's text is edited, the adapted text is left as it is.
function retype(source, type, slot, record) {
    if (!type) throw new Error(`expected ${record.slot} at ${record.where} to declare its type`);
    const text = type.getText(source);
    if (text === slot.adapted) return { ...record, edited: false };
    if (text !== slot.upstream) throw new Error(`unexpected slot ${record.slot} at ${record.where}: ${text}`);
    return { ...record, edited: true, at: type.getStart(source), end: type.end, text: slot.adapted };
}
const files = [...new Set(slots.map(slot => slot.file))].map(file => {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    return { file, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true) };
});
const records = [];
const addressed = new Set();
for (const slot of slots) {
    const { source } = files.find(file => file.file === slot.file);
    const found = find(source, slot);
    // Where the slot is reported: the call, the property gaining `readonly`, the type node, or else the binding's name
    // or the owning declaration, so a slot missing its type is named before it refuses.
    const anchor = slot.readonly ? found.member : found.call ?? found.type ?? found.binding?.name ?? found.owner;
    const at = source.getLineAndCharacterOfPosition(anchor.getStart(source));
    const record = { slot: name(slot), where: `${slot.file}:${at.line + 1}:${at.character + 1}`, file: slot.file };
    // Found exactly once across the table too: two entries naming one node would be one slot read twice.
    const key = `${slot.file}:${anchor.getStart(source)}`;
    if (addressed.has(key)) throw new Error(`slot ${record.slot} at ${record.where} is already in the table`);
    addressed.add(key);
    if (found.call) {
        // A call changes no declaration; its type arguments erase.
        record.visibility = `a call in ${slot.within}`;
        const carried = found.call.typeArguments?.map(type => type.getText(source)).join(', ');
        if (carried === slot.typeArguments) records.push({ ...record, edited: false });
        else if (carried === undefined)
            records.push({ ...record, edited: true, at: found.call.expression.end, end: found.call.expression.end,
                text: `<${slot.typeArguments}>` });
        else throw new Error(`unexpected type arguments on ${record.slot} at ${record.where}: <${carried}>`);
        continue;
    }
    record.visibility = visibility(source, found.owner, record.slot);
    if (found.member && slot.readonly) {
        // The property keeps its type and gains the one modifier.
        const modifiers = found.member.modifiers?.map(modifier => modifier.getText(source)) ?? [];
        if (found.member.questionToken || found.type?.getText(source) !== slot.type)
            throw new Error(`unexpected slot ${record.slot} at ${record.where}: ${found.member.getText(source)}`);
        if (modifiers.length === 1 && modifiers[0] === 'readonly') records.push({ ...record, edited: false });
        else if (modifiers.length === 0) {
            const start = found.member.name.getStart(source);
            records.push({ ...record, edited: true, at: start, end: start, text: 'readonly ' });
        }
        else throw new Error(`unexpected modifiers on ${record.slot} at ${record.where}: ${modifiers.join(' ')}`);
        continue;
    }
    // A property whose type changes keeps no modifier: the view is its type, not a readonly slot.
    if (found.member?.modifiers?.length)
        throw new Error(`unexpected modifiers on ${record.slot} at ${record.where}: ${found.member.getText(source)}`);
    if (found.binding && slot.upstream === null && !found.type) {
        // Upstream writes no annotation: the type goes right after the binding's name.
        const end = found.binding.name.end;
        records.push({ ...record, edited: true, at: end, end, text: `: ${slot.adapted}` });
        continue;
    }
    records.push(retype(source, found.type, slot, record));
}
// Each file's edits from the end, so offsets hold, and no two of them touching the same bytes.
const planned = files.map(({ file, fileName, text }) => {
    const edits = records.filter(record => record.edited && record.file === file).sort((a, b) => b.at - a.at);
    edits.forEach((edit, index) => {
        if (index > 0 && edit.end > edits[index - 1].at)
            throw new Error(`overlapping edits in ${file}: ${edit.slot} and ${edits[index - 1].slot}`);
    });
    let adapted = text;
    for (const edit of edits) adapted = adapted.slice(0, edit.at) + edit.text + adapted.slice(edit.end);
    return { fileName, text, adapted };
});
// Every slot passed: only now is anything written, each file once.
for (const { fileName, text, adapted } of planned) {
    if (adapted === text) continue;
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
for (const record of records)
    console.log(`${record.slot}: ${record.edited ? 'edited' : 'already adapted'} (${record.where}, ${record.visibility})`);
