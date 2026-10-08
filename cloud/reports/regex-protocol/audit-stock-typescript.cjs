// Independent audit only. Options match the runner owner's stock-tsc oracle;
// no outcome or classifier code is changed by this script.
const fs = require('node:fs');
const ts = require(process.argv[2]);
const rows = JSON.parse(fs.readFileSync(process.argv[3], 'utf8'));
const prelude = process.argv[4];
const file = '/tmp/regex-protocol-tsc-input.ts';
const options = {
 strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
 noImplicitReturns: true, noFallthroughCasesInSwitch: true, erasableSyntaxOnly: true,
 verbatimModuleSyntax: true, allowImportingTsExtensions: true, noEmit: true,
 module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
 moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
 lib: ['lib.es2024.d.ts'], types: []
};
const host = ts.createCompilerHost(options);
const original = host.getSourceFile.bind(host), libraries = new Map();
let source = '';
host.getSourceFile = (name, version, onError) => {
 if (name === file) return ts.createSourceFile(name, source, version, true);
 if (!libraries.has(name)) libraries.set(name, original(name, version, onError));
 return libraries.get(name);
};
const results = rows.map(row => {
 source = row.Program;
 const program = ts.createProgram([file, prelude], options, host);
 let diagnostics = program.getSyntacticDiagnostics();
 if (diagnostics.length === 0) diagnostics = ts.getPreEmitDiagnostics(program);
 const codes = [...new Set(diagnostics.filter(d => d.category === ts.DiagnosticCategory.Error).map(d => 'TS' + d.code))].sort();
 return {path: row.Path, adamic: row.Adamic, stockCodes: codes};
});
process.stdout.write(JSON.stringify({version: ts.version, results}, null, 2)+'\n');
