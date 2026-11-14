'use strict';
const fs = require('node:fs');
const ts = require('/tmp/watch-file-check-1532/work/stage3/api/node_modules/typescript');
const { compiler } = require('/tmp/watch-file-check-1532/work/stage3/adapt/47-host-errors/adapt.cjs');
const s = '/tmp/watch-file-check-1532';
const tree = '/tmp/watch-functions-1532/main-tree';
const target = tree + '/src/compiler/sys.ts';
const baseline = fs.readFileSync(s + '/sys-main.ts');
try {
 const results = {};
 for (const side of ['propagated']) {
  fs.writeFileSync(target, fs.readFileSync(s + '/sys-' + side + '.ts'));
  const program = compiler(tree, ts);
  const diagnostics = ts.getPreEmitDiagnostics(program).map(d => {
   const loc = d.file && d.start !== undefined ? d.file.getLineAndCharacterOfPosition(d.start) : null;
   return { file: d.file?.fileName, line: loc && loc.line+1, column: loc && loc.character+1, code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n') };
  });
  results[side] = { roots: program.getRootFileNames().length, diagnostics };
 }
 fs.writeFileSync(s + '/stock-propagated.json', JSON.stringify(results, null, 2) + '\n');
 console.log(JSON.stringify(results, null, 2));
} finally { fs.writeFileSync(target, baseline); }
