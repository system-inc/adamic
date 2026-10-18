const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT);
const root = path.resolve(process.argv[2]);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
if (read.error) throw Error(ts.flattenDiagnosticMessageText(read.error.messageText, '\n'));
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
config.options.typeRoots = [path.join(path.dirname(require.resolve(process.env.CENSUS_TYPESCRIPT)), '../../@types')];
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
const rows = [], others = [], assignments = [];
let patterns = 0;
for (const file of program.getSourceFiles()) {
 if (!file.fileName.startsWith(path.join(root, 'src/compiler') + '/')) continue;
 function visit(node) {
  if (ts.isObjectBindingPattern(node)) {
   patterns++;
   const type = checker.getTypeAtLocation(node);
   const constrained = checker.getBaseConstraintOfType(type) || type;
   const parts = constrained.isUnion() ? constrained.types : [constrained];
   const p = file.getLineAndCharacterOfPosition(node.getStart(file));
   const row = {path:path.relative(root,file.fileName),line:p.line+1,column:p.character+1,type:checker.typeToString(type),text:node.parent.getText(file),members:node.elements.map(e=>(e.propertyName || e.name).getText(file))};
   if (parts.some(t=>t.flags & ts.TypeFlags.StringLike)) rows.push(row);
   else others.push(row);
  }
  if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isObjectLiteralExpression(node.left)) {
   const type = checker.getTypeAtLocation(node.right);
   const constrained = checker.getBaseConstraintOfType(type) || type;
   const parts = constrained.isUnion() ? constrained.types : [constrained];
   const p = file.getLineAndCharacterOfPosition(node.getStart(file));
   assignments.push({path:path.relative(root,file.fileName),line:p.line+1,column:p.character+1,type:checker.typeToString(type),string:parts.some(t=>t.flags & ts.TypeFlags.StringLike),text:node.getText(file)});
  }
  ts.forEachChild(node, visit);
 }
 visit(file);
}
const unresolved = others.filter(r=>r.type==='any' || r.type==='unknown');
console.log(JSON.stringify({typescript:ts.version,patterns,stringPatterns:rows,members:[...new Set(rows.flatMap(r=>r.members))].sort(),assignments,unresolved,otherPatterns:others},null,2));
