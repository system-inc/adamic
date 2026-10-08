// Read-only verification against the independent upstream checkout and stock checker.
const fs=require('fs'),path=require('path'),crypto=require('crypto'),cp=require('child_process');
const pin=path.resolve(process.argv[2]),ts=require(path.resolve(process.argv[3]));
const normalized=s=>s.replace(/\s+/g,'');
const declarationPath=f=>f.startsWith('api:')?path.resolve(__dirname,'../../../api/node_modules',f.slice(4)):path.resolve(pin,f);
const hash=f=>crypto.createHash('sha256').update(fs.readFileSync(f)).digest('hex');
const config=ts.readConfigFile(path.join(pin,'src/compiler/tsconfig.json'),ts.sys.readFile);
const parsed=ts.parseJsonConfigFileContent(config.config,ts.sys,path.join(pin,'src/compiler'));
parsed.options.typeRoots=[path.resolve(__dirname,'../../../api/node_modules/@types')];
const program=ts.createProgram(parsed.fileNames,parsed.options),checker=program.getTypeChecker();
const excluded=new Set(['witnesses.json','needs-code.json'].flatMap(f=>JSON.parse(fs.readFileSync(path.join(__dirname,'../share-a',f))).members.map(m=>m.rank)));
const seen=new Set();let fixtures=0,reads=0;
for(const ledger of ['witnesses.json','needs-code.json']) {
 const record=JSON.parse(fs.readFileSync(path.join(__dirname,ledger)));
 if(cp.execFileSync('git',['-C',pin,'rev-parse','HEAD'],{encoding:'utf8'}).trim()!==record.sourceSha) throw Error('wrong source pin');
 for(const member of record.members) {
  if(member.rank%3 || excluded.has(member.rank) || seen.has(member.rank)) throw Error('wrong/duplicate share rank '+member.rank);
  seen.add(member.rank);reads+=member.reads;
  const original=program.getSourceFile(path.join(pin,member.witness.file));
  if(hash(original.fileName)!==member.fileSha256) throw Error('original file changed '+member.rank);
  if(original.text.slice(member.utf16Start,member.utf16End)!==member.read) throw Error('original read span changed '+member.rank);
  let read;function find(n) {if(ts.SyntaxKind[n.kind]===member.readKind&&n.getStart(original)===member.utf16Start&&n.end===member.utf16End)read=n;ts.forEachChild(n,find);}find(original);
  if(!read || ts.SyntaxKind[read.kind]!==member.readKind) throw Error('original read shape changed '+member.rank);
  const proven=checker.getTypeAtLocation(ts.isBindingElement(read)?read.name:read);
  if(checker.typeToString(proven,read,ts.TypeFormatFlags.NoTruncation)!==member.originalSignature) throw Error('original receiver signature changed '+member.rank);
  for(const file of member.declarationFiles) if(hash(declarationPath(file.file))!==file.sha256) throw Error('original declaration file changed '+member.rank);
  const declarations=checker.getSignaturesOfType(checker.getNonNullableType(proven),ts.SignatureKind.Call).map(s=>s.getDeclaration());
  if(declarations.length!==member.declarations.length) throw Error('original signature count changed '+member.rank);
  declarations.forEach((d,i)=>{const wanted=member.declarations[i],sf=d.getSourceFile();if(d.getStart(sf)!==wanted.start || d.end!==wanted.end || d.getText(sf)!==wanted.text) throw Error('original declaration changed '+member.rank);});
  for(const name of fs.readdirSync(path.join(__dirname,member.directory)).filter(f=>f.endsWith('.a'))) {
   const text=fs.readFileSync(path.join(__dirname,member.directory,name),'utf8');
   if(!normalized(text).includes(normalized(member.targetDeclaration))) throw Error('advertised original contract changed '+member.rank+' '+name);
   if(!normalized(text).includes(normalized(member.carriers))) throw Error('recorded adjacent carriers changed '+member.rank+' '+name);
   const sf=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);let found=false;
   function inspect(n) {if(ts.SyntaxKind[n.kind]===member.readKind && normalized(n.getText(sf))===normalized(member.read)) found=true;ts.forEachChild(n,inspect);}inspect(sf);
   if(!found) throw Error('original member read changed '+member.rank+' '+name);
   const declaration=declarations[0],dsf=declaration.getSourceFile();
   if(member.mode==='unchanged original method member' || member.mode==='original lib member and real array receiver') {
    if(!normalized(text).includes(normalized(declaration.getText(dsf)))) throw Error('original method declaration changed '+member.rank+' '+name);
   } else if(member.mode==='original mapped member and instantiated receiver') {
    let mapped=declaration.parent;while(mapped&&!ts.isMappedTypeNode(mapped))mapped=mapped.parent;
    if(!mapped || !normalized(text).includes(normalized(mapped.getText(dsf)))) throw Error('original mapped declaration changed '+member.rank+' '+name);
   } else if(ts.isFunctionTypeNode(declaration)) {
    if(!normalized(text).includes(normalized(declaration.getText(dsf)))) throw Error('original callable alias changed '+member.rank+' '+name);
   } else if(member.mode==='original external overload declarations; reduced adjacent domains') {
    for(const d of declarations) {const originalSignature='('+d.parameters.map(p=>p.getText(d.getSourceFile())).join(', ')+'): '+d.type.getText(d.getSourceFile())+';';if(!normalized(text).includes(normalized(originalSignature))) throw Error('original external overload changed '+member.rank+' '+name);}
   } else {
    // A field reifies the original function's signature; its body is the only reduction.
    const signature=ts.isFunctionDeclaration(declaration)?'('+declaration.parameters.map(p=>p.getText(dsf)).join(',')+')=>'+declaration.type.getText(dsf):checker.typeToString(checker.getNonNullableType(proven),read,ts.TypeFormatFlags.NoTruncation);
    if(!normalized(member.targetDeclaration).includes(normalized(signature)) && member.mode!=='original external overload declarations; reduced adjacent domains') throw Error('reified original function signature changed '+member.rank+' '+name);
   }
   fixtures++;
  }
 }
}
console.log('Verified '+fixtures+' .a fixtures for '+seen.size+' original pairs / '+reads+' candidate reads; original reads, declarations, hashes and mapped contracts unchanged.');
