const fs=require('node:fs'),path=require('node:path');const ts=require('/workspace/depthgate2-base/stage3/api/node_modules/typescript');
const root='/workspace/depthgate2-adapted',config=root+'/src/compiler/tsconfig.json';const raw=ts.readConfigFile(config,ts.sys.readFile),parsed=ts.parseJsonConfigFileContent(raw.config,ts.sys,path.dirname(config),{},config);const program=ts.createProgram(parsed.fileNames,{...parsed.options,noEmit:true,typeRoots:['/workspace/depthgate2-base/internal/load/node_types/node_modules/@types']});const c=program.getTypeChecker();
const groups=JSON.parse(fs.readFileSync('/workspace/family5-meter-groups.json'));const names=new Set(groups.map(g=>g.P));const writes=[],declarations=[],sourceContradictions=[];
function loc(n){const f=n.getSourceFile(),p=f.getLineAndCharacterOfPosition(n.getStart(f));return path.relative(root,f.fileName)+':'+(p.line+1)+':'+(p.character+1);}
const typeFile=program.getSourceFile(root+'/src/compiler/types.ts');const tr=c.getTypeAtLocation(typeFile.statements.find(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='TextRange').name);const smr=c.getTypeAtLocation(typeFile.statements.find(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='SourceMapRange').name);const wanted=c.getTypeOfPropertyOfType(smr,'source');
for(const f of program.getSourceFiles()){
 if(!f.fileName.startsWith(root+'/src/compiler/')||f.isDeclarationFile)continue;
 function visit(n){
  const name=n.name&&(ts.isIdentifier(n.name)||ts.isStringLiteral(n.name))?n.name.text:undefined;
  if(name&&names.has(name)&&(ts.isPropertySignature(n)||ts.isMethodSignature(n)||ts.isPropertyDeclaration(n)||ts.isMethodDeclaration(n))){const owner=n.parent,type=c.getTypeAtLocation(owner),property=c.getPropertyOfType(type,name),pt=property&&c.getTypeOfSymbolAtLocation(property,n);const r={property:name,where:loc(n),owner:owner.name?.getText(f)||ts.SyntaxKind[owner.kind],type:pt&&c.typeToString(pt),source:n.getText(f)};declarations.push(r);
   if(name==='source'&&pt&&c.isTypeAssignableTo(type,tr)&&!c.isTypeAssignableTo(pt,wanted))sourceContradictions.push(r);
  }
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&n.operatorToken.kind<=ts.SyntaxKind.LastAssignment){let p=n.left,key;if(ts.isPropertyAccessExpression(p))key=p.name.text;else if(ts.isElementAccessExpression(p)&&ts.isStringLiteral(p.argumentExpression))key=p.argumentExpression.text;if(key&&names.has(key)){writes.push({property:key,where:loc(n),receiver:c.typeToString(c.getTypeAtLocation(p.expression)),value:c.typeToString(c.getTypeAtLocation(n.right)),expression:n.getText(f)});}}
  if(ts.isPropertyAssignment(n)&&name&&names.has(name)){const type=c.getTypeAtLocation(n.parent),pt=c.getTypeAtLocation(n.initializer);const r={property:name,where:loc(n),receiver:c.typeToString(type),value:c.typeToString(pt),expression:n.getText(f)};writes.push(r);if(name==='source'&&c.isTypeAssignableTo(type,tr)&&!c.isTypeAssignableTo(pt,wanted))sourceContradictions.push(r);}
  ts.forEachChild(n,visit);
 }visit(f);
}
fs.writeFileSync('/workspace/family5-write-audit.json',JSON.stringify({version:ts.version,scope:'79 compiler roots and their compiler source closure; named and literal-key writes, named declarations; dynamic-key and spread constructors need separate manual tracing',source_text_range_incompatible_declarations_or_literal_values:sourceContradictions,declarations,writes},null,2)+'\n');console.log('Audit:',declarations.length,'declarations,',writes.length,'named writes;',sourceContradictions.length,'TextRange source contradictions');
