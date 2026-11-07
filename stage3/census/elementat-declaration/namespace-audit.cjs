'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const tree='/workspace/scratch/elementat-declaration/baseline',s='/workspace/scratch/elementat-declaration';
function files(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(d,e.name)):e.name.endsWith('.ts')?[path.join(d,e.name)]:[]);}
const options=JSON.parse(fs.readFileSync(s+'/static.json','utf8')).options,program=ts.createProgram(files(tree+'/src'),options),checker=program.getTypeChecker(),core=program.getSourceFile(tree+'/src/compiler/core.ts');
const owner=core.statements.find(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(core)==='elementAt')).declarationList.declarations[0];const target=checker.getSymbolAtLocation(owner.name);
function resolved(symbol){return symbol&&(symbol.flags&ts.SymbolFlags.Alias)?checker.getAliasedSymbol(symbol):symbol;}
function namespace(n){while(ts.isAsExpression(n)||ts.isParenthesizedExpression(n)||ts.isNonNullExpression(n))n=n.expression;return resolved(checker.getTypeAtLocation(n).getProperty('elementAt'))===target;}
const result={checker:ts.version,computed:[],reflection:[],namespaceValueEscapes:[]};
for(const sf of program.getSourceFiles()){
 if(!sf.fileName.startsWith(tree+'/src/'))continue;
 function row(n){const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return {file:path.relative(tree,sf.fileName),line:p.line+1,column:p.character+1,text:n.getText(sf)};}
 function visit(n){
  if(ts.isElementAccessExpression(n)&&namespace(n.expression))result.computed.push(row(n));
  if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&['Object','Reflect','JSON'].includes(n.expression.expression.getText(sf))&&n.arguments.some(namespace))result.reflection.push(row(n));
  if(ts.isIdentifier(n)&&namespace(n)){
   const p=n.parent;
   const fixedReceiver=ts.isPropertyAccessExpression(p)&&p.expression===n;
   const importBinding=ts.isNamespaceImport(p)||ts.isImportSpecifier(p)||ts.isImportClause(p)||ts.isExportSpecifier(p);
   const typeOnly=ts.isTypeQueryNode(p)||ts.isQualifiedName(p)||ts.isTypeReferenceNode(p);
   if(!fixedReceiver&&!importBinding&&!typeOnly&&(ts.isCallExpression(p)&&p.arguments.includes(n)||ts.isNewExpression(p)&&p.arguments?.includes(n)||ts.isSpreadAssignment(p)||ts.isSpreadElement(p)||ts.isReturnStatement(p)||ts.isVariableDeclaration(p)&&p.initializer===n||ts.isBinaryExpression(p)&&p.right===n||ts.isPropertyAssignment(p)&&p.initializer===n))result.namespaceValueEscapes.push(row(n));
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
fs.writeFileSync(s+'/namespace-audit.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
