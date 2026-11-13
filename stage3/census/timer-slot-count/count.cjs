#!/usr/bin/env node
// Usage: NODE_PATH=<stage3/api/node_modules> node count.cjs TREE > tree.json
// One type-checked program covers every TS/TSX/JS/JSX source under src.
const ts = require('typescript');
const fs = require('node:fs');
const path = require('node:path');
const root = path.resolve(process.argv[2]);
function walk(dir) { return fs.readdirSync(dir, {withFileTypes:true}).flatMap(e => e.isDirectory() ? walk(path.join(dir,e.name)) : /\.[cm]?[jt]sx?$/.test(e.name) ? [path.join(dir,e.name)] : []); }
const files = walk(path.join(root,'src')).sort();
const config = ts.readConfigFile(path.join(root,'src/tsconfig-base.json'),ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config,ts.sys,path.join(root,'src'));
const options = {...parsed.options, noEmit:true, composite:false, isolatedDeclarations:false, allowJs:true, checkJs:true, types:['node','mocha','chai'], typeRoots:[path.join(root,'node_modules/@types')]};
const program = ts.createProgram(files,options);
const checker = program.getTypeChecker();
const fmt = ts.TypeFormatFlags.NoTruncation;
const loc = n => {const sf=n.getSourceFile();return {file:path.relative(root,sf.fileName),line:sf.getLineAndCharacterOfPosition(n.getStart()).line+1};};
const txt = n=>n.getText();
const typ = n=>checker.typeToString(checker.getTypeAtLocation(n),n,fmt);
const decls = s=>(s?.declarations||[]).map(d=>({...loc(d),text:txt(d)}));
const names = new Set(['setTimeout','clearTimeout']);
const output = {typescript:ts.version,root,sourceFiles:files.map(f=>path.relative(root,f)),loadedSourceFiles:program.getSourceFiles().filter(f=>f.fileName.startsWith(root+'/src/')).length,accesses:[],suppliers:[],bindings:[],bareCalls:[],declarations:[],spreads:[],assigns:[],bulkCopies:[],genericCopyReads:[],classes:[],diagnostics:[]};
const sys = program.getSourceFile(path.join(root,'src/compiler/sys.ts'));
const sysDecl = sys.statements.find(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='System');
const sysType = checker.getTypeAtLocation(sysDecl.name);
function related(t,seen=new Set()) { if(seen.has(t))return false;seen.add(t); if(t===sysType)return true; if(t.isUnionOrIntersection())return t.types.some(x=>related(x,seen)); return (t.getBaseTypes?.()||[]).some(x=>related(x,seen)); }
function context(n) {for(let p=n.parent;p;p=p.parent)if(ts.isClassDeclaration(p)||ts.isFunctionDeclaration(p)||ts.isMethodDeclaration(p)||ts.isVariableDeclaration(p)) return p.name?.getText()||'<anonymous>';return '<module>';}
function symbolInfo(n) {let s=checker.getSymbolAtLocation(n);if(s?.flags&ts.SymbolFlags.Alias)s=checker.getAliasedSymbol(s);return decls(s);}
function callback(n) { if(!n)return null;let ds=ts.isArrowFunction(n)||ts.isFunctionExpression(n)?[n]:[];const s=checker.getSymbolAtLocation(n);ds.push(...s?.declarations||[]);if(ts.isConditionalExpression(n))return {expression:txt(n),branches:[callback(n.whenTrue),callback(n.whenFalse)]};return {expression:txt(n),type:typ(n),definitions:ds.map(d=>({...loc(d),method:ts.isMethodDeclaration(d),body:d.body?.getText()||null,readsThis:d.body ? /\bthis\b/.test(d.body.getText()):null}))};}
function visit(n) {
 if(ts.isElementAccessExpression(n)&&context(n)==='copyProperties'&&n.expression.getText()==='second')output.genericCopyReads.push({...loc(n),expression:txt(n),receiverType:typ(n.expression),context:'copyProperties'});
 if(ts.isCallExpression(n)) {
  const calleeName=ts.isIdentifier(n.expression)?n.expression.text:ts.isPropertyAccessExpression(n.expression)?n.expression.name.text:'';
  if(['assign','extend','clone','copyProperties','create'].includes(calleeName)&&n.arguments.some(a=>{const t=checker.getTypeAtLocation(a);return t.getProperty('setTimeout')||t.getProperty('clearTimeout')}))output.bulkCopies.push({...loc(n),text:txt(n),callee:calleeName,types:n.arguments.map(typ),declarations:symbolInfo(n.expression),argumentSlots:n.arguments.map(a=>['setTimeout','clearTimeout'].filter(k=>checker.getTypeAtLocation(a).getProperty(k)))});
 }

 if(ts.isSpreadAssignment(n)) { const t=checker.getTypeAtLocation(n.expression); if(t.getProperty('setTimeout')||t.getProperty('clearTimeout')) output.spreads.push({...loc(n),text:txt(n),type:typ(n.expression),context:context(n)}); }
 if(ts.isClassDeclaration(n)) { const t=checker.getTypeAtLocation(n); if(t.getProperty('setTimeout')||t.getProperty('clearTimeout')) output.classes.push({...loc(n),name:n.name?.text,type:typ(n),heritage:n.heritageClauses?.map(txt),members:n.members.filter(m=>names.has(m.name?.getText())).map(m=>({...loc(m),text:txt(m)}))}); }
 if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&n.expression.name.text==='assign') {if(n.arguments.some(a=>{const t=checker.getTypeAtLocation(a);return t.getProperty('setTimeout')||t.getProperty('clearTimeout')}))output.assigns.push({...loc(n),text:txt(n),types:n.arguments.map(typ)});}

 if(ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)) {
  const name=ts.isPropertyAccessExpression(n)?n.name.text:ts.isStringLiteral(n.argumentExpression)?n.argumentExpression.text:null;
  if(names.has(name)) {
   let p=n;while(ts.isNonNullExpression(p.parent)||ts.isParenthesizedExpression(p.parent))p=p.parent;
   const called=ts.isCallExpression(p.parent)&&p.parent.expression===p;
   const write=ts.isBinaryExpression(p.parent)&&p.parent.left===p&&p.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken;
   const receiverType=checker.getTypeAtLocation(n.expression);
   output.accesses.push({...loc(n),name,expression:txt(n),receiverType:typ(n.expression),extendsSystem:related(receiverType),assignableToSystem:checker.isTypeAssignableTo(receiverType,sysType),slotDeclarations:decls(checker.getSymbolAtLocation(ts.isPropertyAccessExpression(n)?n.name:n.argumentExpression)),kind:called?'call':write?'write':'read',calleeStart:n.getStart(),calleeEnd:p.end,receiver:txt(n.expression),...(called?{argumentCount:p.parent.arguments.length,arguments:p.parent.arguments.map(txt),callback:callback(p.parent.arguments[0])}:{}),context:context(n)});
   if(write)output.suppliers.push({...loc(n),name,kind:'assignment',context:context(n),value:txt(p.parent.right),receiverType:typ(n.expression),callback:callback(p.parent.right)});
  }
 }
 if((ts.isPropertyAssignment(n)||ts.isShorthandPropertyAssignment(n)||ts.isMethodDeclaration(n)||ts.isPropertyDeclaration(n))&&names.has(n.name?.getText().replace(/['"]/g,''))) {
  const owner=n.parent;
  output.suppliers.push({...loc(n),name:n.name.getText(),kind:ts.SyntaxKind[n.kind],context:context(n),ownerType:typ(owner),contextualType:ts.isObjectLiteralExpression(owner)?checker.typeToString(checker.getContextualType(owner)||checker.getTypeAtLocation(owner),owner,fmt):null,value:txt(n),callback:callback(n.initializer||n.name)});
 }
 if(ts.isBindingElement(n)&&names.has(n.propertyName?.getText()||n.name.getText()))output.bindings.push({...loc(n),text:txt(n),context:context(n),type:typ(n),declarations:symbolInfo(n.name)});
 if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&names.has(n.expression.text))output.bareCalls.push({...loc(n),name:n.expression.text,calleeStart:n.expression.getStart(),calleeEnd:n.expression.end,argumentCount:n.arguments.length,arguments:n.arguments.map(txt),declarations:symbolInfo(n.expression),callback:callback(n.arguments[0])});
 if((ts.isMethodSignature(n)||ts.isPropertySignature(n)||ts.isFunctionDeclaration(n))&&names.has(n.name?.getText()))output.declarations.push({...loc(n),text:txt(n),context:context(n)});
 ts.forEachChild(n,visit);
}
for(const sf of program.getSourceFiles())if(sf.fileName.startsWith(root+'/src/'))visit(sf);
output.diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({...(d.file?{file:path.relative(root,d.file.fileName),line:d.file.getLineAndCharacterOfPosition(d.start||0).line+1}:{}),code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
process.stdout.write(JSON.stringify(output,null,2)+'\n');
