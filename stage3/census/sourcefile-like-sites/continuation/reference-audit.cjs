const fs=require('fs'),path=require('path'),ts=require('../../../api/node_modules/typescript');
const root='/workspace/scratch/baobab-adapted',out=__dirname,program=JSON.parse(fs.readFileSync(path.join(out,'original/program.json'),'utf8')),rows=[];
for(const name of program.roots){const file=path.join(root,name),sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);const aliases=new Map();for(const st of sf.statements){if(ts.isImportDeclaration(st)){const b=st.importClause?.namedBindings;if(b&&ts.isNamedImports(b))for(const e of b.elements){if(e.propertyName&&['getLineStarts','getLineAndCharacterOfPosition'].includes(e.propertyName.text))aliases.set(e.name.text,e.propertyName.text);}}}function visit(n){
 if(ts.isIdentifier(n)&&(['getLineStarts','getLineAndCharacterOfPosition'].includes(n.text)||aliases.has(n.text))){
  let expr=n;if(ts.isPropertyAccessExpression(n.parent)&&n.parent.name===n)expr=n.parent;
  const parent=expr.parent;let use;
  if(ts.isCallExpression(parent)&&parent.expression===expr)use='direct call';
  else if(ts.isImportSpecifier(n.parent)||ts.isExportSpecifier(n.parent))use='import or export specifier';
  else if((ts.isFunctionDeclaration(n.parent)||ts.isMethodDeclaration(n.parent)||ts.isMethodSignature(n.parent))&&n.parent.name===n)use='declaration';
  else if(ts.isPropertyAssignment(n.parent)&&n.parent.name===n)use='property name';
  else {use=ts.SyntaxKind[parent.kind];const pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));rows.push({file:name,line:pos.line+1,column:pos.character+1,name:n.text,use,context:parent.getText(sf).slice(0,1200)});}
 }
 n.forEachChild(visit);
 }visit(sf);}
fs.writeFileSync(path.join(out,'original/non-direct-helper-references.json'),JSON.stringify(rows,null,2)+'\n');console.log(JSON.stringify(rows,null,2));
