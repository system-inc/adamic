#!/usr/bin/env node
const ts=require('../../api/node_modules/typescript'),fs=require('fs'),path=require('path');
const [rootArg,heatArg,out]=process.argv.slice(2),root=path.resolve(rootArg),heat=JSON.parse(fs.readFileSync(heatArg));
const cfg=ts.getParsedCommandLineOfConfigFile(root+'/src/compiler/tsconfig.json',{}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(String(d.messageText))}});
const roots=ts.sys.readDirectory(root+'/src',['.ts'],['**/lib/**']),program=ts.createProgram(roots,cfg.options),checker=program.getTypeChecker();
const names={NodeFactory:[...new Set(heat.sites.find(s=>s.type==='NodeFactory').members.map(m=>m.name))],TransformationContext:[...new Set(heat.sites.filter(s=>s.type==='TransformationContext').flatMap(s=>s.members.filter(m=>m.method).map(m=>m.name)))],PrintHandlers:heat.sites.find(s=>s.type==='PrintHandlers').members.map(m=>m.name)};
const hook='(globalThis as any).__receiverPrice',inventory=[];
function walk(n,fn){fn(n);ts.forEachChild(n,c=>walk(c,fn));}
function kind(type){const name=type?.aliasSymbol?.name||type?.symbol?.name;return names[name]?name:undefined;}
for(const sf of program.getSourceFiles()){if(!sf.fileName.startsWith(root+'/src/')||sf.isDeclarationFile)continue;const edits=[];function insert(pos,text){edits.push({start:pos,end:pos,text});}function wrap(n,before,after){insert(n.getStart(),before);insert(n.end,after);}
function tick(n,key){wrap(n,`(${hook}.tick(${JSON.stringify(key)}), `,')');}
walk(sf,n=>{
 if(ts.isObjectLiteralExpression(n)){const k=kind(checker.getContextualType(n));if(k){const modulePath=path.relative(path.dirname(sf.fileName),root+'/src/_namespaces/ts.js').split(path.sep).join('/');wrap(n,`${hook}.watch((`, ` satisfies import(${JSON.stringify(modulePath.startsWith('.')?modulePath:'./'+modulePath)}).${k}), ${JSON.stringify(k)}, ${JSON.stringify(names[k])})`);inventory.push({file:path.relative(root,sf.fileName),kind:k,position:n.getStart()});}}
 if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)){
  const access=n.expression;
  if(sf.fileName.endsWith('/compiler/parser.ts')&&ts.isIdentifier(access.expression)&&access.expression.text==='factory'&&names.NodeFactory.includes(access.name.text))tick(n,'parserCalls.'+access.name.text);
  if(ts.isIdentifier(access.expression)&&access.expression.text==='context'&&names.TransformationContext.includes(access.name.text)&&ts.isArrowFunction(n.parent)&&ts.isVariableDeclaration(n.parent.parent))tick(n,'transformerCalls.'+path.basename(sf.fileName)+'.'+access.name.text);
  if(access.expression.getText(sf)==='Object'&&access.name.text==='defineProperty')edits.push({start:access.getStart(),end:access.end,text:hook+'.define'});
 }
 if(ts.isFunctionDeclaration(n)&&n.name?.text==='createPrinter'&&n.body){insert(n.body.getStart()+1,`\n ${hook}.tick("printerEntries"); ${hook}.watch(handlers, "PrintHandlers", ${JSON.stringify(names.PrintHandlers)});\n`);}
 if(ts.isVariableDeclaration(n)&&n.initializer&&ts.isPropertyAccessExpression(n.initializer)&&n.initializer.expression.getText(sf)==='handlersView')tick(n.initializer,'printerAliases.'+n.name.getText(sf));
 if(ts.isVariableDeclaration(n)&&n.initializer&&ts.isBinaryExpression(n.initializer)&&n.initializer.operatorToken.kind===ts.SyntaxKind.QuestionQuestionToken&&ts.isPropertyAccessExpression(n.initializer.left)&&n.initializer.left.expression.getText(sf)==='handlersView')tick(n.initializer.left,'printerAliases.'+n.name.getText(sf));
 if(ts.isBinaryExpression(n)&&n.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&n.operatorToken.kind<=ts.SyntaxKind.LastAssignment&&ts.isPropertyAccessExpression(n.left)){const k=kind(checker.getTypeAtLocation(n.left.expression));if(k&&(k!=='TransformationContext'||names[k].includes(n.left.name.text)))tick(n,'astWrites.'+k+'.'+n.left.name.text);}
 if(ts.isBinaryExpression(n)&&[ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(n.operatorToken.kind)){
  const operands=[n.left,n.right];
  if(operands.some(x=>ts.isPropertyAccessExpression(x)&&names[kind(checker.getTypeAtLocation(x.expression))]?.includes(x.name.text)))tick(n,'receiverIdentity');
  if(sf.fileName.endsWith('/compiler/emitter.ts')&&operands.some(x=>ts.isIdentifier(x)&&['onEmitNode','substituteNode'].includes(x.text)))tick(n,'printerAliasSentinelIdentity');
 }
 if(ts.isPropertyAccessExpression(n)&&n.name.text==='length'&&ts.isPropertyAccessExpression(n.expression)&&names[kind(checker.getTypeAtLocation(n.expression.expression))]?.includes(n.expression.name.text))tick(n,'receiverLength');
});
if(edits.length){let text=sf.text;for(const e of edits.sort((a,b)=>b.start-a.start||b.end-a.end))text=text.slice(0,e.start)+e.text+text.slice(e.end);fs.writeFileSync(sf.fileName,text);}}
fs.writeFileSync(out,JSON.stringify({sourceRoots:roots.length,names,objects:inventory},null,2)+'\n');
