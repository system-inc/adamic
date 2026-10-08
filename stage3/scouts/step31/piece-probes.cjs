'use strict';
// Small independent API witnesses for the proposed checker and emitter seams.
const assert = require('node:assert/strict');
function probes(ts) {
    const token = {isCancellationRequested: () => false, throwIfCancellationRequested() {}};
    const options = {target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.ESNext, strict: true, noLib: true};
    const file = ts.createSourceFile('/probe.ts', '// kept comment\nexport const value: number = "wrong";', options.target, true);
    // program.ts:3318-3361 normally records these host-side reference lists.
    // This fixed probe has no imports, ambient modules or module augmentations.
    file.imports = [];
    file.moduleAugmentations = [];
    file.ambientModuleNames = [];
    const calls = new Set();
    const available = {
        getCompilerOptions: () => options,
        getSourceFiles: () => [file],
        getSourceFile: name => name === file.fileName ? file : undefined,
        getCurrentDirectory: () => '/',
        getCanonicalFileName: name => name,
        getCommonSourceDirectory: () => '/',
        isSourceFileDefaultLibrary: () => false,
        isSourceFileFromExternalLibrary: () => false,
        isSourceOfProjectReferenceRedirect: () => false,
        getEmitModuleFormatOfFile: () => options.module,
        getImpliedNodeFormatForEmit: () => undefined,
        redirectTargetsMap: new Map(),
        getSymlinkCache: undefined,
        getPackageJsonInfoCache: undefined,
        getGlobalTypingsCacheLocation: undefined,
        readFile: undefined,
        useCaseSensitiveFileNames: () => true,
    };
    const host = new Proxy(available, {get(target, key) {
        calls.add(String(key));
        if (!(key in target)) throw new Error('unimplemented probe host capability: ' + String(key));
        return target[key];
    }});
    const checker = ts.createTypeChecker(host);
    const diagnostics = checker.getDiagnostics(file, token);
    assert.deepEqual(diagnostics.map(value => value.code), [2322]);
    assert.equal(diagnostics[0].start, 29);
    const printed = ts.createPrinter({newLine: ts.NewLineKind.LineFeed}).printFile(file);
    assert.equal(printed, '// kept comment\nexport const value: number = "wrong";\n');
    const emitted = [];
    Object.assign(available, {
        isEmitBlocked: () => false,
        shouldTransformImportCall: () => false,
        getBuildInfo: () => undefined,
        getSourceFileFromReference: () => undefined,
        writeFile: (name, text, bom) => emitted.push({name, text, bom}),
    });
    const resolver = checker.getEmitResolver(file, token);
    const result = ts.emitFiles(resolver, host, file, ts.getTransformers(options), false, false);
    assert.equal(result.emitSkipped, false);
    assert.deepEqual(result.diagnostics, []);
    assert.deepEqual(emitted, [{name: '/probe.js', text: '// kept comment\nexport const value = "wrong";\n', bom: false}]);
    return {diagnostics: diagnostics.map(value => ({code: value.code, start: value.start,
        length: value.length, message: ts.flattenDiagnosticMessageText(value.messageText, '\n')})),
        printed, emitted, hostCapabilitiesUsed: [...calls].sort()};
}
module.exports = {probes};
if (require.main === module) console.log(JSON.stringify(probes(require(process.env.STEP31_TYPESCRIPT)), null, 2));
