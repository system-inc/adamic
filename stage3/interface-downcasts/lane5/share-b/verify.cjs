// Verify every family variant against the independent original-source evidence.
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const ts = require(path.join(process.argv[2], 'lib/typescript.js'));
const families = JSON.parse(fs.readFileSync(path.join(__dirname, process.env.ADAMIC_CALLABLE_SHARE_B_FAMILIES || 'families.json')));
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
  let fixtureText = fs.readFileSync(file,'utf8');
  // In-memory mutations prove the source-declaration and enum evidence pins.
  if (process.env.ADAMIC_CALLABLE_SHARE_B_DECLARATION_MUTANT === '535' && family.rank === 535 && variant === 'good') fixtureText = fixtureText.replace('node: Node','node: number');
  if (process.env.ADAMIC_CALLABLE_SHARE_B_ENUM_MUTANT === '175' && family.rank === 175 && variant === 'good') fixtureText = fixtureText.replace(/Let\s*= 1 << 0/,'Let = 1 << 1');
  const source = ts.createSourceFile(file,fixtureText,ts.ScriptTarget.Latest,true);
  let declaration, read;
  function visit(node) {
   if ((ts.isMethodSignature(node) || ts.isPropertySignature(node)) && node.parent.name?.text === 'Target' && node.name.getText(source) === family.field) declaration = node;
   if (ts.isPropertyAccessExpression(node) && normalized(node.getText(source)) === normalized(family.read)) read = node;
   ts.forEachChild(node, visit);
  }
  visit(source);
  if (!declaration || normalized(declaration.getText(source)) !== normalized(family.declaration)) throw Error(file + ': original declaration changed');
  if (!read) throw Error(file + ': original read changed');
  for (const carrier of family.carrierEnums || []) {
   const node=source.statements.find(n=>ts.isEnumDeclaration(n)&&n.name.text===carrier.name);
   if (!node || normalized(node.getText(source))!==normalized(carrier.declaration)) throw Error(file+': original enum representation changed');
   const bytes=fs.readFileSync(path.join(process.argv[2],carrier.sourceFile));
   if (crypto.createHash('sha256').update(bytes).digest('hex')!==carrier.sha256) throw Error('original enum source changed');
   const original=ts.createSourceFile(carrier.sourceFile,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);let originalEnum;
   function find(n) {if(ts.isEnumDeclaration(n)&&n.name.text===carrier.name) originalEnum=n;ts.forEachChild(n,find);}find(original);
   if (!originalEnum || normalized(originalEnum.getText(original).replace(/^export\s+/,'').replace(/^const\s+enum/,'enum'))!==normalized(carrier.declaration)) throw Error('enum provenance changed');
  }
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

const probes = ['batch-02-probes.json','batch-03-debug-probes.json','batch-03-signature-probes.json','batch-03-map-probes.json','batch-03-array-probes.json','batch-03-diagnostic-probes.json'].flatMap(f=>JSON.parse(fs.readFileSync(path.join(__dirname,f))));
const optional = JSON.parse(fs.readFileSync(path.join(__dirname,'batch-02-original.json'))).members.find(m=>m.rank===205);
for (const member of [...probes,{...optional,filename:'rank-205/good.a'}]) {
 const source = ts.createSourceFile(member.filename,fs.readFileSync(path.join(__dirname,member.filename),'utf8'),ts.ScriptTarget.Latest,true);
 let declarations=[],read;
 function visit(node) {
  if ((ts.isMethodSignature(node)||ts.isPropertySignature(node))&&node.parent.name?.text==='Target'&&node.name.getText(source)===member.field) declarations.push(node.getText(source));
  if (ts.isPropertyAccessExpression(node)&&normalized(node.getText(source))===normalized(member.read)) read=node;
  ts.forEachChild(node,visit);
 }
 visit(source);
 for (const [file,hash] of [[member.witness.file,member.fileSha256],[member.declarationFile,member.declarationSha256]]) {
  if (crypto.createHash('sha256').update(fs.readFileSync(path.join(process.argv[2],file))).digest('hex')!==hash) throw Error('original source changed '+file);
 }
 for (const header of member.originalFunctionHeaders || []) {if(!fs.readFileSync(path.join(process.argv[2],member.declarationFile),'utf8').includes(header)) throw Error('namespace function header changed');}
 if (normalized(declarations.join('\n'))!==normalized(member.declaration)||!read) throw Error(member.filename+': original declaration/read changed');
}
console.log('Verified ' + families.length + ' pairs / ' + families.reduce((n,f)=>n+f.candidateReads,0) + ' candidate reads in ' + count + ' original-member fixtures.');
