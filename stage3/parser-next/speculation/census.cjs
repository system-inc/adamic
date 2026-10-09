// Count source call shapes independently with the pinned TypeScript parser.
const fs=require('node:fs'),crypto=require('node:crypto'),ts=require('../../api/node_modules/typescript');
const file=process.argv[2],source=fs.readFileSync(file,'utf8');
const tree=ts.createSourceFile(file,source,ts.ScriptTarget.Latest,true),rows={};
function visit(node){if(ts.isCallExpression(node)&&ts.isIdentifier(node.expression)&&['lookAhead','tryParse','speculationHelper'].includes(node.expression.text)&&node.arguments.length){const arg=node.arguments[0],shape=ts.isIdentifier(arg)?'named':ts.isArrowFunction(arg)?'arrow':'other',key=node.expression.text+'/'+shape;const row=rows[key]??={count:0,lines:[]};row.count++;row.lines.push(tree.getLineAndCharacterOfPosition(node.getStart(tree)).line+1);}ts.forEachChild(node,visit);}visit(tree);
console.log(JSON.stringify({source: file,sha256:crypto.createHash('sha256').update(source).digest('hex'),shapes:rows},null,2));
