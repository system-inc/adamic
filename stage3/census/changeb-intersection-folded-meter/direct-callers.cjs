// Read every adapted compiler source with the TypeScript AST. The first-error
// meter's unit is the containing source-file statement, including nested calls.
const fs = require('fs');
const path = require('path');
const ts = require('/workspace/scratch/compiler-a/stage3/api/node_modules/typescript');
const root = '/workspace/scratch/tree';
const result = [];
const catalog = [];
function walk(dir) {
  for (const entry of fs.readdirSync(dir, {withFileTypes:true})) {
    const file = path.join(dir, entry.name);
    if (entry.isDirectory()) { walk(file); continue; }
    if (!/\.(ts|a)$/.test(file)) continue;
    const text = fs.readFileSync(file, 'utf8');
    const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
    const location = node => {
      const p = sf.getLineAndCharacterOfPosition(node.getStart(sf));
      return path.relative(root, file)+':'+(p.line+1)+':'+(p.character+1);
    };
    for (const unit of sf.statements) {
      const calls=[];
      function visit(node) {
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && ['tryCast','cast'].includes(node.expression.text)) {
          calls.push({callee:node.expression.text,where:location(node),start:node.getStart(sf),text:node.getText(sf)});
        }
        ts.forEachChild(node,visit);
      }
      visit(unit);
      catalog.push({where:location(unit),name:unit.name?.getText(sf)|| (ts.isVariableStatement(unit)?unit.declarationList.declarations.map(d=>d.name.getText(sf)).join(", "):ts.SyntaxKind[unit.kind])});
      if (calls.length) result.push({file:path.relative(root,file),where:location(unit),start:unit.pos,end:unit.end,name:unit.name?.getText(sf)|| (ts.isVariableStatement(unit)?unit.declarationList.declarations.map(d=>d.name.getText(sf)).join(', '):ts.SyntaxKind[unit.kind]),calls});
    }
  }
}
walk(root+'/src/compiler');
process.stdout.write(JSON.stringify({calls:result.reduce((n,u)=>n+u.calls.length,0),units:result.length,callers:result,catalog},null,2)+'\n');
