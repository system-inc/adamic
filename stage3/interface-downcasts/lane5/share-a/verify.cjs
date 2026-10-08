const fs=require('fs'),path=require('path'),crypto=require('crypto'),cp=require('child_process');
const pin=process.argv[2],ts=require(process.argv[3]);
const manifest=JSON.parse(fs.readFileSync(path.join(__dirname,'witnesses.json')));
if(cp.execFileSync('git',['-C',pin,'rev-parse','HEAD'],{encoding:'utf8'}).trim()!==manifest.sourceSha)throw Error('wrong original source pin');
const normalized=s=>s.replace(/^export\s+/,'').replace(/\s+/g,'');
const original=ts.createSourceFile('types.ts',fs.readFileSync(path.join(pin,'src/compiler/types.ts'),'utf8'),ts.ScriptTarget.Latest,true);
const originalNodes=new Map();function collect(n){if(ts.isTypeAliasDeclaration(n)||ts.isInterfaceDeclaration(n))originalNodes.set(n.name.text,n);ts.forEachChild(n,collect);}collect(original);
const originalProgram=ts.createProgram([path.join(pin,"src/compiler/types.ts")],{noResolve:true});const checker=originalProgram.getTypeChecker();const checkedOriginal=originalProgram.getSourceFile(path.join(pin,"src/compiler/types.ts"));const kinds=new Map();function kindsVisit(n){if(ts.isInterfaceDeclaration(n)){const k=n.members.find(m=>m.name?.getText(checkedOriginal)==="kind");if(k){const value=checker.getTypeAtLocation(k.type).value;if(typeof value==="number")kinds.set(n.name.text,value);}}ts.forEachChild(n,kindsVisit);}kindsVisit(checkedOriginal);
let fixtures=0;
for(const member of manifest.members){
 if(member.rank%3)throw Error('wrong share');
 for(const [file,hash] of [[member.witness.file,member.fileSha256],[member.declarationFile,member.declarationSha256]])if(crypto.createHash('sha256').update(fs.readFileSync(path.join(pin,file))).digest('hex')!==hash)throw Error('original source drift '+file);
 const bytes=fs.readFileSync(path.join(pin,member.witness.file),'utf8');if(bytes.slice(member.utf16Start,member.utf16End)!==member.read)throw Error('original read span drift');
 for(const filename of fs.readdirSync(path.join(__dirname,member.directory)).filter(f=>f.endsWith('.a'))){
  const sf=ts.createSourceFile(filename,fs.readFileSync(path.join(__dirname,member.directory,filename),'utf8'),ts.ScriptTarget.Latest,true);let declaration,read;const aliases=new Map();
  function visit(n){if(ts.isEnumDeclaration(n)&&n.name.text==='PrivateIdentifierKind'&&!normalized(fs.readFileSync(path.join(pin,member.declarationFile),'utf8')).includes(normalized(n.getText(sf))))throw Error(filename+': original enum changed');if((ts.isMethodSignature(n)||ts.isPropertySignature(n))&&n.parent.name?.text===(member.containerName||'Target')&&n.name.getText(sf)===member.field)declaration=n;if(ts.isPropertyAccessExpression(n)&&normalized(n.getText(sf))===normalized(member.read))read=n;if(ts.isTypeAliasDeclaration(n))aliases.set(n.name.text,n);if(ts.isInterfaceDeclaration(n)&&kinds.has(n.name.text)){const k=n.members.find(m=>m.name?.getText(sf)==='kind');if(k){if(!k.modifiers?.some(m=>m.kind===ts.SyntaxKind.ReadonlyKeyword)||!ts.isLiteralTypeNode(k.type)||!ts.isNumericLiteral(k.type.literal)||Number(k.type.literal.text)!==kinds.get(n.name.text))throw Error(filename+': original discriminator changed '+n.name.text);}}ts.forEachChild(n,visit);}visit(sf);
  if(!declaration||normalized(declaration.getText(sf))!==normalized(member.declaration)||!read)throw Error(filename+': original member changed');
  if(member.targetDeclaration&&(!aliases.has('Target')||normalized(aliases.get('Target').getText(sf))!==normalized(member.targetDeclaration)))throw Error(filename+': mapped receiver changed');
  for(const name of ['PropertyName','MemberName','TypeComparer'])if(aliases.has(name)&&normalized(aliases.get(name).getText(sf))!==normalized(originalNodes.get(name).getText(original)))throw Error(filename+': original alias changed '+name);
  fixtures++;
 }
}
console.log('Verified '+fixtures+' fixtures for '+manifest.members.length+' original pairs and '+manifest.members.reduce((n,m)=>n+m.reads,0)+' candidate reads.');

const gaps=JSON.parse(fs.readFileSync(path.join(__dirname,'gaps.json')));
for(const member of gaps.members){
 for(const [file,hash] of [[member.witness.file,member.fileSha256],[member.declarationFile,member.declarationSha256]])if(crypto.createHash('sha256').update(fs.readFileSync(path.join(pin,file))).digest('hex')!==hash)throw Error('frontier source drift '+file);
 const originalText=fs.readFileSync(path.join(pin,member.witness.file),'utf8');if(originalText.slice(member.utf16Start,member.utf16End)!==member.read)throw Error('frontier read drift');
 const file=path.join(__dirname,member.directory,'good.a'),sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);const declarations=[];let read=false;
 function visit(n){if((ts.isMethodSignature(n)||ts.isPropertySignature(n))&&n.parent.name?.text===(member.containerName||'Target')&&n.name.getText(sf)===member.field)declarations.push(n.getText(sf));if(ts.isPropertyAccessExpression(n)&&normalized(n.getText(sf))===normalized(member.read))read=true;ts.forEachChild(n,visit);}visit(sf);
 if(!read||declarations.length!==member.declarations.length||declarations.some((d,i)=>normalized(d)!==normalized(member.declarations[i])))throw Error('frontier declarations changed '+member.rank);
}
console.log('Verified '+gaps.members.length+' original overloaded/generic frontier fixtures.');

for(const ledger of ['intrinsic-frontiers.json','intrinsic-candidates.json']){
 const evidence=JSON.parse(fs.readFileSync(path.join(__dirname,ledger)));
 for(const member of evidence.members){
  if(member.rank%3)throw Error('wrong intrinsic share');
  for(const [file,hash] of [[member.witness.file,member.fileSha256],[member.declarationFile,member.declarationSha256]])if(crypto.createHash('sha256').update(fs.readFileSync(path.join(pin,file))).digest('hex')!==hash)throw Error('intrinsic source drift '+file);
  const text=fs.readFileSync(path.join(pin,member.witness.file),'utf8');if(text.slice(member.utf16Start,member.utf16End)!==member.read)throw Error('intrinsic read drift');
  const originalDeclaration=normalized(fs.readFileSync(path.join(pin,member.declarationFile),'utf8'));for(const decl of member.declarations)if(!originalDeclaration.includes(normalized(decl)))throw Error('intrinsic declaration not in original '+member.rank);
  const sf=ts.createSourceFile('good.a',fs.readFileSync(path.join(__dirname,member.directory,'good.a'),'utf8'),ts.ScriptTarget.Latest,true);const declarations=[];let read=false;
  function visit(n){if((ts.isMethodSignature(n)||ts.isPropertySignature(n))&&n.parent.name?.text===member.containerName&&n.name?.getText(sf)===member.field)declarations.push(n.getText(sf));if(ts.isPropertyAccessExpression(n)&&normalized(n.getText(sf))===normalized(member.read))read=true;ts.forEachChild(n,visit);}visit(sf);
  if(!read||declarations.length!==member.declarations.length||declarations.some((d,i)=>normalized(d)!==normalized(member.declarations[i])))throw Error('intrinsic declaration changed '+member.rank);
 }
 console.log('Verified '+evidence.members.length+' intrinsic frontier source controls in '+ledger+'.');
}
