// Verify every family variant against the independent original-source evidence.
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const ts = require(path.join(process.argv[2], 'lib/typescript.js'));
const families = JSON.parse(fs.readFileSync(path.join(__dirname, 'families.json')));
const normalized = text => text.replace(/\s+/g, '');
let count = 0;
for (const family of families) {
 const bytes = fs.readFileSync(path.join(process.argv[2], family.witness.file));
 if (crypto.createHash('sha256').update(bytes).digest('hex') !== family.fileSha256) throw Error('original read source changed');
 const text = bytes.toString('utf8');
 if (text.slice(family.utf16Start, family.utf16End) !== family.read) throw Error('original read span changed');
 const declarationBytes = fs.readFileSync(path.join(process.argv[2], family.declarationFile));
 if (crypto.createHash('sha256').update(declarationBytes).digest('hex') !== family.declarationSha256) throw Error('original declaration source changed');
 for (const variant of family.variants) {
  const file = path.join(__dirname, family.directory, variant + '.a');
  const source = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
  let declaration, read;
  function visit(node) {
   if ((ts.isMethodSignature(node) || ts.isPropertySignature(node)) && node.parent.name?.text === 'Target' && node.name.getText(source) === family.field) declaration = node;
   if (ts.isPropertyAccessExpression(node) && normalized(node.getText(source)) === normalized(family.read)) read = node;
   ts.forEachChild(node, visit);
  }
  visit(source);
  if (!declaration || normalized(declaration.getText(source)) !== normalized(family.declaration)) throw Error(file + ': original declaration changed');
  if (!read) throw Error(file + ': original read changed');
  count++;
 }
}
const blocked = JSON.parse(fs.readFileSync(path.join(__dirname, 'blocked-original-witnesses.json'))).members;
for (const member of blocked) {
 const files = {4:'blocked-assignment.a',10:'blocked-push.a',37:'blocked-join.a'};
 const source = ts.createSourceFile(files[member.rank], fs.readFileSync(path.join(__dirname, files[member.rank]), 'utf8'), ts.ScriptTarget.Latest, true);
 let declarations = [], read;
 function visit(node) {
  if ((ts.isMethodSignature(node) || ts.isPropertySignature(node)) && node.parent.name?.text === 'Target' && node.name.getText(source) === member.field) declarations.push(node.getText(source));
  if (ts.isPropertyAccessExpression(node) && normalized(node.getText(source)) === normalized(member.read)) read = node;
  ts.forEachChild(node, visit);
 }
 visit(source);
 if (normalized(declarations.join('\n')) !== normalized(member.declaration) || !read) throw Error(files[member.rank] + ': original blocker declaration/read changed');
}
console.log('Verified  + families.length + ' pairs / ' + families.reduce((n,f)=>n+f.candidateReads,0) + ' candidate reads in ' + count + ' original-member fixtures.');
