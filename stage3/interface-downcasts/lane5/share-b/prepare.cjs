// Original declaration/read evidence; all fixture carriers are reductions.
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const child = require('child_process');
const original = process.argv[2];
const ts = require(path.join(original, 'lib/typescript.js'));
const pin = '050880ce59e30b356b686bd3144efe24f875ebc8';
if (child.execFileSync('git', ['-C', original, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== pin) throw Error('wrong original pin');
const ranks = new Set(process.argv[3].split(',').map(Number));
const pairs = JSON.parse(fs.readFileSync(path.join(__dirname, '../unknown-callable-pairs-ranked.json'))).filter(p => ranks.has(p.rank));
const interfaces = new Map();
const functions = new Map();
function scan(directory) {
 for (const entry of fs.readdirSync(directory, {withFileTypes:true})) {
  const file = path.join(directory, entry.name);
  if (entry.isDirectory()) scan(file);
  else if (file.endsWith('.ts')) {
   const source = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
   function visit(node) {
    if (ts.isInterfaceDeclaration(node)) {
     const name = node.name.text;
     interfaces.set(name, [...(interfaces.get(name) || []), {node, source}]);
    }
    if (ts.isFunctionDeclaration(node) && node.name && node.body) functions.set(node.name.text, {node, source});
    ts.forEachChild(node, visit);
   }
   visit(source);
  }
 }
}
scan(path.join(original, 'src/compiler'));
scan(path.join(original, 'src/lib'));
function member(name, field, chain = []) {
 if (chain.includes(name)) return;
 for (const {node, source} of interfaces.get(name) || []) {
  const declarations = node.members.filter(m => m.name?.getText(source) === field);
  if (declarations.length) return {declarations, source, inheritancePath:[...chain, name]};
 }
 for (const {node, source} of interfaces.get(name) || []) for (const clause of node.heritageClauses || []) for (const base of clause.types) {
  const found = member(base.expression.getText(source), field, [...chain, name]);
  if (found) return found;
 }
}
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const members = [];
for (const pair of pairs) {
 const source = ts.createSourceFile(pair.witness.file, fs.readFileSync(path.join(original, pair.witness.file), 'utf8'), ts.ScriptTarget.Latest, true);
 let read;
 function visit(node) {
  if (ts.isPropertyAccessExpression(node) && node.name.text === pair.field || ts.isBindingElement(node) && (node.propertyName || node.name).getText(source) === pair.field) {
   const lc = source.getLineAndCharacterOfPosition(node.getStart(source));
   if (lc.line + 1 === pair.witness.line && lc.character + 1 === pair.witness.column) read = node;
  }
  ts.forEachChild(node, visit);
 }
 visit(source);
 const name = pair.type.endsWith('[]') ? 'Array' : pair.type.replace(/ \| undefined$/, '').replace(/<.*>$/, '');
 const found = member(name, pair.field);
 if (!read || !found) throw Error('missing original ' + pair.rank + ' ' + pair.type + '.' + pair.field);
 const declarationFile = path.relative(original, found.source.fileName);
 members.push({rank:pair.rank, receiver_type_id:pair.receiver_type_id, type:pair.type, field:pair.field, candidateReads:pair.reads, witness:pair.witness, read:read.getText(source), call:read.parent.getText(source), utf16Start:read.getStart(source), utf16End:read.end, fileSha256:hash(source.text), declaration:found.declarations.map(d => d.getText(found.source)).join('\n'), declarationFile, declarationSha256:hash(found.source.text), inheritancePath:found.inheritancePath});
}
if (members.length !== ranks.size) throw Error('missing rank');
fs.writeFileSync(path.join(__dirname, process.argv[4] || 'original-witnesses.json'), JSON.stringify({sourceSha:pin, basis:'complete original member declarations and reads; adjacent data carriers reduced', members}, null, 2) + '\n');
console.log('Verified ' + members.length + ' original members / ' + members.reduce((n,m) => n + m.candidateReads,0) + ' candidate reads.');
