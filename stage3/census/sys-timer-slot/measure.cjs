'use strict';
const fs=require('fs'),path=require('path'),ts=require('typescript');
const [root,out]=process.argv.slice(2);
if(ts.version!=='6.0.3')throw Error(ts.version);
const virtual=root+'/timer-compat.ts';
const virtualText='export interface CandidateA { setTimeout?(callback: (...args: unknown[]) => void, ms: number, ...args: unknown[]): unknown; clearTimeout?(timeoutId: unknown): void; } export interface CandidateB { setTimeout?<A extends unknown[]>(callback: (...args: A) => void, ms: number, ...args: A): unknown; clearTimeout?(timeoutId: unknown): void; }';
const files=ts.sys.readDirectory(root+'/src',['.ts'],undefined,['**/*']);
const cfg=ts.readConfigFile(root+'/src/harness/tsconfig.json',ts.sys.readFile);
const opts=ts.parseJsonConfigFileContent(cfg.config,ts.sys,root+'/src/harness').options;
Object.assign(opts,{noEmit:true,disableSourceOfProjectReferenceRedirect:false});
const host={getScriptFileNames:()=>[...files,virtual],getScriptVersion:()=> '0',getScriptSnapshot:f=>{let t=f===virtual?virtualText:ts.sys.readFile(f);return t===undefined?undefined:ts.ScriptSnapshot.fromString(t)},getCurrentDirectory:()=>root,getCompilationSettings:()=>opts,getDefaultLibFileName:o=>ts.getDefaultLibFilePath(o),fileExists:ts.sys.fileExists,readFile:ts.sys.readFile,readDirectory:ts.sys.readDirectory,directoryExists:ts.sys.directoryExists,getDirectories:ts.sys.getDirectories};
const ls=ts.createLanguageService(host),p=ls.getProgram(),c=p.getTypeChecker();
function loc(n){const f=n.getSourceFile(),l=f.getLineAndCharacterOfPosition(n.getStart());return {file:path.relative(root,f.fileName),line:l.line+1,column:l.character+1};}
const type=n=>c.typeToString(c.getTypeAtLocation(n),n,ts.TypeFormatFlags.NoTruncation);
const refs=n=>(ls.findReferences(n.getSourceFile().fileName,n.getStart())||[]).map(g=>({definition:g.definition,references:g.references.map(r=>{const f=p.getSourceFile(r.fileName),l=f.getLineAndCharacterOfPosition(r.textSpan.start);return {file:path.relative(root,r.fileName),line:l.line+1,column:l.character+1,context:f.text.slice(Math.max(0,f.getLineStarts()[l.line]),f.getLineStarts()[l.line+1]||f.text.length).trim(),isDefinition:r.isDefinition,isWriteAccess:r.isWriteAccess,text:f.text.slice(r.textSpan.start,r.textSpan.start+r.textSpan.length)}})}));
const targets={};for(const st of p.getSourceFile(virtual).statements){const t=c.getTypeAtLocation(st);targets[st.name.text]=Object.fromEntries(['setTimeout','clearTimeout'].map(name=>[name,c.getNonNullableType(c.getTypeOfSymbolAtLocation(t.getProperty(name),st))]));}
const fits=n=>Object.fromEntries(Object.entries(targets).map(([k,v])=>[k,c.isTypeAssignableTo(c.getNonNullableType(c.getTypeAtLocation(n)),v[n.name.text])]));
const members=[],calls=[],accesses=[];
for(const f of p.getSourceFiles()){if(!f.fileName.startsWith(root+'/src/'))continue;function walk(n){
 if(n.name&&['setTimeout','clearTimeout'].includes(n.name.text)&&(ts.isMethodDeclaration(n)||ts.isMethodSignature(n)||ts.isPropertyAssignment(n)||ts.isShorthandPropertyAssignment(n)||ts.isPropertyDeclaration(n)||ts.isPropertySignature(n)||ts.isBindingElement(n)||ts.isVariableDeclaration(n)||ts.isFunctionDeclaration(n))){members.push({...loc(n),kind:ts.SyntaxKind[n.kind],owner:n.parent.name?.text||n.parent.parent?.name?.text||n.parent.parent?.getText().slice(0,100),text:n.getText(),type:type(n),fits:fits(n),references:refs(n.name)});}
 if(ts.isPropertyAccessExpression(n)&&['setTimeout','clearTimeout'].includes(n.name.text)){accesses.push({...loc(n),text:n.getText(),receiverType:type(n.expression),type:type(n)});}
 if(ts.isCallExpression(n)){const resolved=c.getResolvedSignature(n);const resolvedName=resolved?.declaration?.name?.text;let e=n.expression;if(ts.isNonNullExpression(e))e=e.expression;if((ts.isPropertyAccessExpression(e)&&['setTimeout','clearTimeout'].includes(e.name.text))||(ts.isIdentifier(e)&&['setTimeout','clearTimeout'].includes(e.text))||(ts.isElementAccessExpression(e)&&['setTimeout','clearTimeout'].includes(e.argumentExpression.text))||['setTimeout','clearTimeout'].includes(resolvedName)){
 const destination=(ts.isBinaryExpression(n.parent)&&n.parent.right===n)?n.parent.left:ts.isVariableDeclaration(n.parent)?n.parent.name:undefined;
 const handleSymbol=destination&&(ts.isPropertyAccessExpression(destination)?destination.name:destination);
 const cb=n.arguments[0],sig=cb&&c.getTypeAtLocation(cb).getCallSignatures();
 calls.push({...loc(n),timerName:resolvedName||e.name?.text||e.argumentExpression?.text||e.text,text:n.getText(),receiverType:(ts.isPropertyAccessExpression(e)||ts.isElementAccessExpression(e))?type(e.expression):'identifier',resolvedDeclaration:c.getResolvedSignature(n)?.declaration&&loc(c.getResolvedSignature(n).declaration),callbackType:cb&&type(cb),callbackSignatures:sig?.map(s=>c.signatureToString(s,cb,ts.TypeFormatFlags.NoTruncation)),arguments:n.arguments.map(a=>({text:a.getText(),type:type(a)})),parent:n.parent.getText(),handleReferences:handleSymbol?refs(handleSymbol):undefined,callbackDeclarations:cb?c.getTypeAtLocation(cb).getCallSignatures().map(s=>s.declaration?{...loc(s.declaration),text:s.declaration.getText()}:null):[],parentKind:ts.SyntaxKind[n.parent.kind],destinationType:(ts.isBinaryExpression(n.parent)&&n.parent.right===n)?type(n.parent.left):ts.isVariableDeclaration(n.parent)?type(n.parent.name):undefined});}}
 ts.forEachChild(n,walk);}walk(f);}
fs.writeFileSync(out,JSON.stringify({typescript:ts.version,root,sourceFiles:files.length,programFiles:p.getSourceFiles().length,members,calls,accesses},null,2)+'\n');
console.log(JSON.stringify({members:members.length,calls:calls.length,accesses:accesses.length}));
