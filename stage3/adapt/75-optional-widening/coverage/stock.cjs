#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [configFile, sitesFile, output] = process.argv.slice(2);
const config = ts.readConfigFile(configFile, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configFile), {}, configFile);
const program = ts.createProgram(parsed.fileNames, parsed.options);
const c = program.getTypeChecker();
const rows = JSON.parse(fs.readFileSync(sitesFile)).map(r => {
 const file = program.getSourceFile(path.join(path.dirname(configFile), 'src/compiler', r.File));
 assert(file);
 const pos = file.getPositionOfLineAndCharacter(r.Line - 1, r.Column - 1);
 let found;
 function visit(n) { if (n.getStart(file) === pos && n.getText(file) === r.Text && 'Kind' + ts.SyntaxKind[n.kind] === r.Kind) found = n; ts.forEachChild(n, visit); }
 visit(file); assert(found, r.File);
 const t = c.getTypeAtLocation(found);
 const parts = t.isUnion() ? t.types : [t];
 return {id:`${r.File}:${r.Line}:${r.Column}`, type:c.typeToString(t), flags:t.flags, never:!!(t.flags & ts.TypeFlags.Never), members:parts.map(m => ({type:c.typeToString(m), flags:m.flags, never:!!(m.flags & ts.TypeFlags.Never), has_id:!!c.getPropertyOfType(m, 'id')}))};
});
const diagnostics = ts.getPreEmitDiagnostics(program).map(d => ({code:d.code, message:ts.flattenDiagnosticMessageText(d.messageText, '\n')}));
fs.writeFileSync(output, JSON.stringify({typescript:ts.version, options:parsed.options, diagnostics, rows}, null, 2) + '\n');
