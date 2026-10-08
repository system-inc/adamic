// Syntactic ownership exposure, not proof of the receiver's type or a leak.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const ts = require(process.env.SCOUT_TYPESCRIPT);
if (ts.version !== '6.0.3') throw new Error('TypeScript 6.0.3 required');
const [root, output] = process.argv.slice(2);
if (!root || !output) throw new Error('usage: inventory.cjs PINNED_TREE OUTPUT_JSON');
const base = path.join(root, 'src/compiler'), files = [];
function collect(dir) { for (const entry of fs.readdirSync(dir, {withFileTypes:true})) {
 const full = path.join(dir, entry.name);
 if (entry.isDirectory()) collect(full); else if (entry.name.endsWith('.ts')) files.push(full);
}}
collect(base); files.sort();
const sites = [], hashes = {}, summaries = {};
for (const file of files) {
 const relative = path.relative(root,file), bytes = fs.readFileSync(file);
 hashes[relative] = crypto.createHash('sha256').update(bytes).digest('hex');
 const source = ts.createSourceFile(file, bytes.toString('utf8'), ts.ScriptTarget.Latest, true);
 const counts = {}, stack = [];
 function add(category,node,extra={}) {
  counts[category] = (counts[category]||0)+1;
  const position=source.getLineAndCharacterOfPosition(node.getStart(source));
  sites.push({category,file:relative,line:position.line+1,column:position.character+1,
   function:stack.join('/'),text:node.getText(source).replace(/\s+/g,' ').slice(0,240),...extra});
 }
 function visit(node) {
  const callable = ts.isFunctionLike(node) && node.body;
  if (callable) {
   if (stack.includes('createTypeChecker')) add('checker_nested_callable',node);
   const name=node.name?.getText(source)||'<anonymous>';
   stack.push(name);
  }
  if (ts.isNewExpression(node) && ts.isIdentifier(node.expression) && ['Map','Set'].includes(node.expression.text)) {
   const name=node.expression.text;
   add('new_'+name,node,{typeArguments:node.typeArguments?.map(t=>t.getText(source))||[]});
   if (!stack.length) add('module_lexical_new_'+name,node);
   if (stack.includes('createTypeChecker')) add('checker_new_'+name,node);
  }
  if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression)
   && ['set','delete','clear'].includes(node.expression.name.text)) add('property_call_'+node.expression.name.text,node);
  if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && ['getNodeLinks','getSymbolLinks'].includes(node.expression.text)) add('call_'+node.expression.text,node);
  if (ts.isBinaryExpression(node) && node.operatorToken.kind===ts.SyntaxKind.EqualsToken
   && ts.isPropertyAccessExpression(node.left) && node.left.name.text==='parent') add('parent_assignment',node);
  if (ts.isVariableDeclaration(node) && node.name.getText(source)==='checker' && ts.isObjectLiteralExpression(node.initializer)) {
   add('checker_api_object',node,{properties:node.initializer.properties.length});
  }
  ts.forEachChild(node,visit);
  if (callable) stack.pop();
 }
 visit(source); summaries[relative]=counts;
}
const totals={}; for (const counts of Object.values(summaries)) for(const [k,v] of Object.entries(counts)) totals[k]=(totals[k]||0)+v;
fs.writeFileSync(output,JSON.stringify({typescript:ts.version,pin:'050880ce59e30b356b686bd3144efe24f875ebc8',scope:'tracked src/compiler/**/*.ts, generated files excluded',files:files.length,totals,summaries,hashes,sites},null,2)+'\n');
console.log(JSON.stringify({files:files.length,totals},null,2));
