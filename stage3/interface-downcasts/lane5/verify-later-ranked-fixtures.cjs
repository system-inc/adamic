const fs=require('fs');
const path=require('path');
const ts=require(path.join(process.argv[2],'lib/typescript.js'));
const evidence=JSON.parse(fs.readFileSync(path.join(__dirname,'later-ranked-original-witnesses.json'),'utf8'));
const directories=new Map([[61,'string-from-node'],[67,'local-name'],[68,'disallowed-comma'],[69,'token-start'],[70,'true'],[72,'system-exit'],[73,'if-statement'],[74,'null'],[75,'parenthesized'],[76,'type-reference'],[77,'program-options'],[78,'system-write'],[79,'context-options'],[81,'declaration-name'],[82,'lift-block'],[83,'left-access'],[84,'source-files'],[85,'resolution-path'],[86,'performance-measure'],[88,'binary'],[91,'helper-factory'],[92,'hoist-variable'],[94,'modifier-flags'],[95,'token-end'],[96,'token-full-start'],[98,'export-declaration'],[99,'computed-name'],[100,'source-file-update'],[101,'context-diagnostic'],[102,'literal-type'],[103,'source-file-path'],[106,'environment-variable'],[107,'emit-resolver'],[109,'writer-line'],[110,'watcher-close'],[115,'logical-and'],[116,'named-exports'],[117,'token-text'],[118,'token-value'],[120,'arrow-function'],[121,'comma-reducer'],[122,'variable-update-tagged'],[124,'read-helpers'],[127,'case-sensitive'],[128,'diagnostic-newline'],[129,'conditional-expression'],[131,'binary-update'],[132,'config-diagnostic'],[133,'scanner-text'],[135,'lexical-end'],[137,'module-format'],[142,'false'],[143,'parenthesized-type'],[141,'export-specifier'],[144,'property-signature'],[145,'type-check'],[146,'for-update'],[148,'prefix-operand'],[151,'syntax-kind'],[154,'static-block'],[155,'prefix-unary'],[156,'internal-name'],[157,'call-update'],[158,'class-update'],[136,'substitution-hook'],[163,'resolution-settings'],[164,'scanner-scan'],[166,'lexical-start'],[170,'writer-space']]);
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
   if((ts.isMethodSignature(node)||ts.isPropertySignature(node))&&node.parent.name?.text==='Target'&&node.name.getText(source)===member.field) declaration=node;
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
