const ts = require(process.argv[2]);
const prelude = process.argv[3];
// Kept identical to internal/load/load.go, including the declaration prelude.
const options = {
  strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
  noImplicitReturns: true, noFallthroughCasesInSwitch: true, erasableSyntaxOnly: true,
  verbatimModuleSyntax: true, allowImportingTsExtensions: true, noEmit: true,
  module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
  moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
  lib: ['lib.es2024.d.ts'], types: []
};
const file = process.argv[4];
const host = ts.createCompilerHost(options);
const original = host.getSourceFile.bind(host);
const libraries = new Map();
let source = '';
host.getSourceFile = (name, version, onError) => {
  if (name === file) return ts.createSourceFile(name, source, version, true);
  if (!libraries.has(name)) libraries.set(name, original(name, version, onError));
  return libraries.get(name);
};
console.log(JSON.stringify(ts.version));
require('node:readline').createInterface({input: process.stdin}).on('line', line => {
  source = JSON.parse(line);
  const program = ts.createProgram([file, prelude], options, host);
  let diagnostics = program.getSyntacticDiagnostics();
  if (diagnostics.length === 0) diagnostics = ts.getPreEmitDiagnostics(program);
  const codes = [...new Set(diagnostics.filter(d => d.category === ts.DiagnosticCategory.Error).map(d => 'TS' + d.code))].sort();
  console.log(JSON.stringify(codes));
});
