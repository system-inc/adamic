// Checker-only witnesses. .a inputs are exposed as virtual TypeScript sources.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('expected TypeScript 6.0.3');
const directory = path.join(__dirname, 'witnesses');
const prelude = '/adamic-prelude/adamic.d.ts';
const options = {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    erasableSyntaxOnly: false, verbatimModuleSyntax: true, allowImportingTsExtensions: true,
    noEmit: true, module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
};
for (const file of fs.readdirSync(directory).filter(name => name.endsWith('.a')).sort()) {
    const source = fs.readFileSync(path.join(directory, file), 'utf8');
    const virtual = path.join(directory, file + '.ts');
    const modes = [['stock', options, false], ['adamic-prelude', options, true]];
    const control = { 'index-read.a': 'noUncheckedIndexedAccess', 'optional-write.a': 'exactOptionalPropertyTypes', 'catch.a': 'useUnknownInCatchVariables' }[file];
    if (control) modes.push(['without-' + control, { ...options, [control]: false }, false]);
    for (const [mode, configuration, ambient] of modes) {
        const host = ts.createCompilerHost(configuration);
        const getSource = host.getSourceFile.bind(host);
        host.getSourceFile = (name, language, ...rest) => {
            if (name === virtual) return ts.createSourceFile(name, source, language, true);
            if (name === prelude) return ts.createSourceFile(name, fs.readFileSync(path.join(__dirname, 'evidence/prelude-ordinary.txt'), 'utf8'), language, true);
            return getSource(name, language, ...rest);
        };
        const program = ts.createProgram({ rootNames: ambient ? [virtual, prelude] : [virtual], options: configuration, host });
        const diagnostics = ts.getPreEmitDiagnostics(program).map(diagnostic => {
            const point = diagnostic.file?.getLineAndCharacterOfPosition(diagnostic.start || 0);
            return { line: point ? point.line + 1 : 0, column: point ? point.character + 1 : 0,
                code: 'TS' + diagnostic.code, message: ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n') };
        });
        console.log(JSON.stringify({ file, mode, diagnostics }));
    }
}
