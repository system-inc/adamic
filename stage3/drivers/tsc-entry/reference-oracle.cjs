// Stock TypeScript checks the explicit, merged source program independently of Adamic.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SCANNER_TYPESCRIPT);
const configName = path.resolve(process.argv[2]);
const includeTypes = process.argv[3] === 'with-types';
const roots = new Set(), configs = [], seen = new Set(), types = new Set();
function visit(name) {
  name = path.resolve(name);
  if (seen.has(name)) return;
  seen.add(name);
  const read = ts.readConfigFile(name, ts.sys.readFile);
  if (read.error) throw new Error(ts.flattenDiagnosticMessageText(read.error.messageText, '\n'));
  const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(name), undefined, name);
  if (config.errors.length) throw new Error(config.errors.map(d => ts.flattenDiagnosticMessageText(d.messageText, '\n')).join('\n'));
  configs.push({name, roots: config.fileNames, options: config.options});
  for (const root of config.fileNames) roots.add(root);
  for (const type of config.options.types || []) types.add(type);
  for (const ref of config.projectReferences || []) {
    let file = ref.path;
    if (fs.statSync(file).isDirectory()) file = path.join(file, 'tsconfig.json');
    visit(file);
  }
}
visit(configName);
const options = {...configs[0].options, noEmit: true};
let directory = path.dirname(configName);
for (const root of roots) {
  if (root.endsWith('.d.ts')) continue;
  while (path.relative(directory, root).startsWith('..'+path.sep)) directory = path.dirname(directory);
}
options.rootDir = directory;
if (includeTypes) options.types = [...types];
const program = ts.createProgram({rootNames: [...roots], options});
const diagnostics = ts.getPreEmitDiagnostics(program).map(d => ({
  file: d.file?.fileName,
  ...(d.file ? (() => {const p=d.file.getLineAndCharacterOfPosition(d.start); return {line:p.line+1,column:p.character+1};})() : {}),
  code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')
}));
console.log(JSON.stringify({configs, roots:[...roots], options, sources:program.getSourceFiles().filter(f=>!f.isDeclarationFile).map(f=>f.fileName),diagnostics},null,2));
process.exitCode = diagnostics.length ? 1 : 0;
