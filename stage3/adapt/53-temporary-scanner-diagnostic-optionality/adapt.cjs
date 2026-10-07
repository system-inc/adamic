const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT || 'typescript');
const tree = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree, 'slice.json'))) throw new Error('adaptation 53 requires a declaration slice');
const file = path.join(tree, 'src/compiler/diagnosticInformationMap.generated.ts');
let text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const declaration = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'diag');
if (!declaration) throw new Error('diag missing');
let changed = 0;
if (declaration.type) {
    if (declaration.type.getText(source) !== 'DiagnosticMessage') throw new Error('unexpected diag return type');
    const start = text.lastIndexOf(':', declaration.type.getStart(source));
    text = text.slice(0,start)+text.slice(declaration.type.end);
    fs.writeFileSync(file,text); changed = 1;
}
console.log(JSON.stringify({inferredPrivateReturn: changed}));
