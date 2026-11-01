'use strict';
const path = require('node:path'), ts = require(path.resolve(process.argv[2]));
const text = '/** @type {string} */ let key="xyz";\nclass Computed { static [key]=1; }\nComputed.xyz;';
const name = '/95g/control.js', options = {allowJs: true, checkJs: true, noEmit: true};
const host = ts.createCompilerHost(options), original = host.getSourceFile;
host.getSourceFile = (file, ...args) => file === name ? ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true) : original.call(host, file, ...args);
const program = ts.createProgram([name], options, host), checker = program.getTypeChecker(), node = program.getSourceFile(name).statements[2].expression;
const symbol = checker.getSymbolAtLocation(node);
console.log(JSON.stringify({result: symbol?.name, declarations: symbol?.declarations?.length, parent: symbol?.parent?.name, diagnostics: ts.getPreEmitDiagnostics(program).map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}))}, null, 2));
