'use strict';
const path = require('node:path'), ts = require(path.resolve(process.argv[2]));
Error.stackTraceLimit = 50;
const exported = "/** @type {string} */\nlet key = 'xyz';\nclass Computed { static [key] = 1; }\nmodule.exports = /** @type {typeof Computed} */ (class {});\nmodule.exports.extra = 2;\n";
const cases = [
    {name: 'original', file: '/95g/consumer.ts', text: "import v = require('./exported');\nv.xyz;\n"},
    {name: 'original-super', file: '/95g/derived.ts', text: "import v = require('./exported');\nclass Derived extends v { static read() { return super.xyz; } }\n"},
    {name: 'successful-super', file: '/95g/super.js', text: '/** @type {string} */\nlet key = "xyz";\nclass Computed { static [key] = 1; }\nclass Derived extends Computed { static read() { return super.xyz; } }\n'},
];
const operations = [
    ['quickInfo', (s, f, p) => s.getQuickInfoAtPosition(f, p)],
    ['definition', (s, f, p) => s.getDefinitionAtPosition(f, p)],
    ['definitionAndSpan', (s, f, p) => s.getDefinitionAndBoundSpan(f, p)],
    ['typeDefinition', (s, f, p) => s.getTypeDefinitionAtPosition(f, p)],
    ['implementation', (s, f, p) => s.getImplementationAtPosition(f, p)],
    ['findReferences', (s, f, p) => s.findReferences(f, p)],
    ['references', (s, f, p) => s.getReferencesAtPosition(f, p)],
    ['renameInfo', (s, f, p) => s.getRenameInfo(f, p, {})],
    ['renameLocations', (s, f, p) => s.findRenameLocations(f, p, false, false, {})],
    ['highlights', (s, f, p) => s.getDocumentHighlights(f, p, [f])],
    ['signatureHelp', (s, f, p) => s.getSignatureHelpItems(f, p, {})],
    ['completions', (s, f, p) => s.getCompletionsAtPosition(f, p, {includeCompletionsForModuleExports: false})],
    ['completionDetails', (s, f, p) => s.getCompletionEntryDetails(f, p, 'xyz', {}, undefined, {}, undefined)],
    ['semanticClassifications', (s, f, p, n) => s.getSemanticClassifications(f, {start: 0, length: n})],
    ['semanticClassifications2020', (s, f, p, n) => s.getEncodedSemanticClassifications(f, {start: 0, length: n}, ts.SemanticClassificationFormat.TwentyTwenty)],
    ['prepareCallHierarchy', (s, f, p) => s.prepareCallHierarchy(f, p)],
    ['incomingCalls', (s, f, p) => s.provideCallHierarchyIncomingCalls(f, p)],
    ['outgoingCalls', (s, f, p) => s.provideCallHierarchyOutgoingCalls(f, p)],
    ['refactors', (s, f, p) => s.getApplicableRefactors(f, p, {})],
];
const results = [];
for (const fixture of cases) for (const [operation, invoke] of operations) {
    const files = new Map([[fixture.file, fixture.text], ['/95g/exported.js', exported]]);
    const options = {strict: true, allowJs: true, checkJs: true, noEmit: true, target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.CommonJS, moduleResolution: ts.ModuleResolutionKind.Node10, ignoreDeprecations: '6.0'};
    const host = {getScriptFileNames: () => [fixture.file], getScriptVersion: () => '1',
        getScriptSnapshot: f => {const data = files.has(f) ? files.get(f) : ts.sys.readFile(f); return data === undefined ? undefined : ts.ScriptSnapshot.fromString(data);},
        getCurrentDirectory: () => '/95g', getCompilationSettings: () => options, getDefaultLibFileName: ts.getDefaultLibFilePath,
        fileExists: f => files.has(f) || ts.sys.fileExists(f), readFile: f => files.has(f) ? files.get(f) : ts.sys.readFile(f),
        readDirectory: ts.sys.readDirectory, directoryExists: d => d === '/95g' || ts.sys.directoryExists(d)};
    const service = ts.createLanguageService(host), row = {fixture: fixture.name, operation};
    row.diagnostics = [...service.getSyntacticDiagnostics(fixture.file), ...service.getSemanticDiagnostics(fixture.file)].map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}));
    global.__95gEvents = [];
    try { const value = invoke(service, fixture.file, fixture.text.lastIndexOf('xyz') + 1, fixture.text.length); row.result = value === undefined ? 'undefined' : value; }
    catch (error) { row.error = error.name + ': ' + error.message; row.stack = error.stack; }
    row.scopedExitStacks = global.__95gEvents;
    results.push(row); service.dispose();
}
console.log(JSON.stringify(results, null, 2));
