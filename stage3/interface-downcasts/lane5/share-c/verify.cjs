// Independent declaration and read witness, following the lane's existing verifier.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const root=process.argv[2],ts=require(path.join(root,'lib/typescript.js'));
const evidence=JSON.parse(fs.readFileSync(path.join(__dirname,'original-members.json')));
const rows=JSON.parse(fs.readFileSync(path.join(__dirname,'certified.json')));
const carriers=JSON.parse(fs.readFileSync('/tmp/lane5-c-carrier-declarations.json'));
const normalized=t=>t.replace(/\s+/g,'');
let count=0;
for(const row of rows){
 if(row.rank%3!==2)throw Error('wrong share rank '+row.rank);
 const member=evidence.members.find(m=>m.rank===row.rank);
 const original=fs.readFileSync(path.join(root,member.witness.file),'utf8');
 if(crypto.createHash('sha256').update(original).digest('hex')!==member.fileSha256||original.slice(member.start,member.end)!==member.read)throw Error('original source drift');
 for(const variant of ['good','wrong-arity']){
  const source=ts.createSourceFile('fixture.a',fs.readFileSync(path.join(__dirname,'families','rank-'+row.rank,variant+'.a'),'utf8'),ts.ScriptTarget.Latest,true);
  let declaration,read=false;
  function visit(n){
   if((ts.isMethodSignature(n)||ts.isPropertySignature(n))&&n.parent.name?.text==='Target'&&n.name.getText(source)===member.field)declaration=n;
   if(ts.isPropertyAccessExpression(n)&&normalized(n.getText(source))===normalized(member.read))read=true;
   if(ts.isTypeAliasDeclaration(n)){
    if(!carriers[n.name.text]||normalized(n.getText(source))!==normalized(carriers[n.name.text].text))throw Error('original alias changed '+n.name.text);
   }
   if(ts.isInterfaceDeclaration(n)&&!['Base','Target'].includes(n.name.text)){
    const expected=carriers[n.name.text]?.kindValue,kind=n.members.find(m=>m.name?.getText(source)==='kind');
    if(expected!==undefined&&(!kind||kind.type.getText(source)!==String(expected)))throw Error('original kind changed '+n.name.text);
   }
   ts.forEachChild(n,visit);
  }visit(source);
  if(!declaration||normalized(declaration.getText(source))!==normalized(member.declarations[0].text))throw Error('original declaration changed '+row.rank);
  if(!read)throw Error('original read changed '+row.rank);
  count++;
 }
}
console.log('Verified '+rows.length+' original pairs and '+count+' fixtures, including original aliases and numeric tags.');
