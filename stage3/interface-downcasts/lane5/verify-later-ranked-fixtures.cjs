const fs=require('fs');
const path=require('path');
const ts=require(path.join(process.argv[2],'lib/typescript.js'));
const evidence=JSON.parse(fs.readFileSync(path.join(__dirname,'later-ranked-original-witnesses.json'),'utf8'));
const directories=new Map([[61,'string-from-node'],[67,'local-name'],[68,'disallowed-comma'],[69,'token-start'],[70,'true'],[72,'system-exit'],[73,'if-statement'],[74,'null'],[75,'parenthesized'],[76,'type-reference'],[77,'program-options'],[78,'system-write'],[79,'context-options']]);
const ranks=new Set(process.argv[3].split(',').map(Number));
const normalized=text=>text.replace(/\s+/g,'');
if(evidence.members.filter(m=>ranks.has(m.rank)).length!==ranks.size)throw Error("missing requested original member");
let count=0;
for(const member of evidence.members.filter(m=>ranks.has(m.rank))) {
 const directory=path.join(__dirname,'later-ranked-callables',directories.get(member.rank));
 for(const file of fs.readdirSync(directory).filter(f=>f.endsWith('.a'))) {
  const source=ts.createSourceFile(file,fs.readFileSync(path.join(directory,file),'utf8'),ts.ScriptTarget.Latest,true);
  let declaration,read;
  function visit(node) {
   if(ts.isMethodSignature(node)&&node.parent.name?.text==='Target'&&node.name.getText(source)===member.field) declaration=node;
   if((ts.isPropertyAccessExpression(node)||ts.isBindingElement(node))&&normalized(node.getText(source))===normalized(member.read)) read=node;
   ts.forEachChild(node,visit);
  }
  visit(source);
  if(!declaration||normalized(declaration.getText(source))!==normalized(member.declaration)) throw Error(file+': original declaration changed');
  if(!read) throw Error(file+': original member read changed');
  count++;
 }
}
if(!count)throw Error('no fixtures verified');
console.log('Verified '+count+' fixtures retain complete original declarations and reads.');
