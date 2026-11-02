const fs=require('fs'),path=require('path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const root=process.argv[2], files=ts.sys.readDirectory(path.join(root,'src'),['.ts']);
const program=ts.createProgram(files,{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true});
const checker=program.getTypeChecker(), source=program.getSourceFile(path.join(root,'src/compiler/types.ts'));
const names=['showConfig','configFile','configFilePath','help','init','listFiles','listEmittedFiles','project','build','version'];
const rows=[];
const same=(s,r)=>s===r.symbol||s?.declarations?.some(d=>r.symbol?.declarations?.includes(d));
function visit(n,fn){fn(n);ts.forEachChild(n,c=>visit(c,fn));}
visit(source,n=>{if(ts.isInterfaceDeclaration(n)&&['CompilerOptions','ProjectReference'].includes(n.name.text))for(const m of n.members)if(m.name&&(n.name.text==='CompilerOptions'?names.includes(m.name.text):m.name.text==='originalPath'))rows.push({owner:n.name.text,member:m.name.text,declaration:m.getText(source),symbol:checker.getSymbolAtLocation(m.name),references:[]});});
for(const file of program.getSourceFiles().filter(f=>f.fileName.startsWith(root+'/src/')))visit(file,n=>{if(!ts.isIdentifier(n)&&!ts.isStringLiteral(n))return;const row=rows.find(r=>r.member===n.text&&(same(checker.getSymbolAtLocation(n),r)||(ts.isPropertyAssignment(n.parent)&&n.parent.name===n&&same(checker.getPropertyOfType(checker.getContextualType(n.parent.parent)||checker.getTypeAtLocation(n.parent.parent),n.text),r))));if(!row)return;const p=n.parent;let expr=p;
if(ts.isPropertyAccessExpression(p)&&p.name===n)expr=p;
let kind='read';if(ts.isPropertySignature(p))kind='declaration';else if(ts.isPropertyAssignment(p)&&p.name===n)kind='literal';else if(ts.isBinaryExpression(expr.parent)&&expr.parent.left===expr&&ts.isAssignmentOperator(expr.parent.operatorToken.kind))kind='write';
const pos=file.getLineAndCharacterOfPosition(n.getStart(file));row.references.push({file:path.relative(root,file.fileName),line:pos.line+1,column:pos.character+1,kind,text:file.text.slice(file.getLineStarts()[pos.line],file.getLineStarts()[pos.line+1]||file.text.length).trim()});});
const generic=[];
for(const file of program.getSourceFiles().filter(f=>f.fileName.startsWith(root+'/src/')))visit(file,n=>{
 if(!ts.isElementAccessExpression(n)&&!ts.isSpreadAssignment(n))return;
 const receiver=checker.getTypeAtLocation(n.expression);
 const members=rows.filter(r=>same(checker.getPropertyOfType(receiver,r.member),r)).map(r=>r.owner+'.'+r.member);
 if(!members.length)return;
 const pos=file.getLineAndCharacterOfPosition(n.getStart(file));
 generic.push({file:path.relative(root,file.fileName),line:pos.line+1,members,text:n.getText(file)});
});
for(const file of program.getSourceFiles().filter(f=>f.fileName.startsWith(root+'/src/')))visit(file,n=>{
 if(!ts.isBindingElement(n)||!ts.isObjectBindingPattern(n.parent))return;
 const receiver=checker.getTypeAtLocation(n.parent);
 const name=n.propertyName||n.name;
 const matches=rows.filter(r=>r.member===name.text&&same(checker.getPropertyOfType(receiver,r.member),r));
 if(matches.length){const pos=file.getLineAndCharacterOfPosition(n.getStart(file));generic.push({file:path.relative(root,file.fileName),line:pos.line+1,members:matches.map(r=>r.owner+'.'+r.member),text:n.getText(file),kind:'destructure'});}
});
console.log(JSON.stringify(generic,null,2));
