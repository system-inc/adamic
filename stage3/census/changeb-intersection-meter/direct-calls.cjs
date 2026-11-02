const ts = require('/workspace/scratch/compiler-A/stage3/api/node_modules/typescript');
const fs = require('fs'), path = require('path');
const root = '/workspace/scratch/tree/src/compiler';
const rows = [], unitNames = {};
function scan(dir) {
 for (const item of fs.readdirSync(dir, {withFileTypes:true})) {
  const file = path.join(dir,item.name);
  if(item.isDirectory()) { scan(file); continue; }
  if(!/\.(ts|a)$/.test(file)) continue;
  const source = fs.readFileSync(file,'utf8');
  const sf=ts.createSourceFile(file,source,ts.ScriptTarget.Latest,true,ts.ScriptKind.TS);
  for (const unit of sf.statements) {
   const loc=sf.getLineAndCharacterOfPosition(unit.getStart(sf));
   const name=unit.name?.text || (ts.isVariableStatement(unit) ? unit.declarationList.declarations.map(d=>d.name.getText(sf)).join(', ') : null);
   if(name) unitNames[`${file}:${loc.line+1}:${loc.character+1}`]=name;
  }
  function visit(node) {
   if(ts.isCallExpression(node) && ts.isIdentifier(node.expression) && ['tryCast','cast'].includes(node.expression.text)) {
    let unit=node; while(unit.parent && unit.parent!==sf) unit=unit.parent;
    const loc=sf.getLineAndCharacterOfPosition(node.getStart(sf));
    const start=sf.getLineAndCharacterOfPosition(unit.getStart(sf));
    rows.push({callee:node.expression.text, file, where:`${file}:${loc.line+1}:${loc.character+1}`, unit_where:`${file}:${start.line+1}:${start.character+1}`, unit_name:unit.name?.text || null, line:source.split('\n')[loc.line].trim()});
   }
   ts.forEachChild(node,visit);
  }
  visit(sf);
 }
}
scan(root);
fs.writeFileSync(process.argv[2],JSON.stringify(rows,null,2)+'\n');
fs.writeFileSync(path.join(path.dirname(process.argv[2]),'source-unit-names.json'),JSON.stringify(unitNames,null,2)+'\n');
console.log(JSON.stringify({calls:rows.length,units:new Set(rows.map(x=>x.unit_where)).size}));
