'use strict';
const path = require('node:path'), ts = require(path.resolve(process.argv[2]));
const name = '/95g/control.ts';
const text = 'declare const key: string;\nclass Computed { static [key] = 1; }\nComputed.xyz;\ntype Named = { [k: string]: number; [k: symbol]: number };\ndeclare const alias: Named;\nalias.xyz;\n';
const options = {noEmit: true, strict: true, target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.CommonJS};
const host = ts.createCompilerHost(options), original = host.getSourceFile;
host.getSourceFile = (f, ...args) => f === name ? ts.createSourceFile(f, text, ts.ScriptTarget.Latest, true) : original.call(host, f, ...args);
const program = ts.createProgram([name], options, host), checker = program.getTypeChecker(), source = program.getSourceFile(name);
const result = {diagnostics: ts.getPreEmitDiagnostics(program).map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')})), queries: []};
for (const n of [source.statements[2].expression, source.statements[5].expression]) {
    const symbol = checker.getSymbolAtLocation(n);
    result.queries.push({text: n.getText(), symbol: symbol?.name, declarations: symbol?.declarations?.length, parent: symbol?.parent?.name});
}
console.log(JSON.stringify(result, null, 2));
