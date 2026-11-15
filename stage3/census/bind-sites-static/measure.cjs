// Static inventory only. Stock TypeScript compiler API; no textual call-site matching.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require('../../api/node_modules/typescript');
const root = path.resolve(process.argv[2]);
const out = path.resolve(process.argv[3]);
if (ts.version !== '6.0.3') throw Error('requires TypeScript 6.0.3');
const roots = ts.sys.readDirectory(path.join(root, 'src/compiler'), ['.ts'], undefined, ['**/*']).sort();
const options = {strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,
  target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,
  moduleDetection:ts.ModuleDetectionKind.Force,verbatimModuleSyntax:true,allowImportingTsExtensions:true,
  noEmit:true,types:[],lib:['lib.es2024.d.ts'],typeRoots:[path.resolve(__dirname,'../../api/node_modules/@types')]};
const ambientRoots=[path.resolve(__dirname,'../../api/node_modules/@types/node/index.d.ts')];
const program = ts.createProgram([...roots,...ambientRoots], options);
const checker = program.getTypeChecker();
const files = program.getSourceFiles().filter(f=>!f.isDeclarationFile && f.fileName.startsWith(root + path.sep));
function loc(n) { const f=n.getSourceFile(),p=f.getLineAndCharacterOfPosition(n.getStart(f));return {file:path.relative(root,f.fileName),line:p.line+1,column:p.character+1}; }
function key(n) {const l=loc(n);return l.file+':'+l.line+':'+l.column;}
function textType(t) {return checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation);}
function symbol(n) {let s=checker.getSymbolAtLocation(n);return s && s.flags&ts.SymbolFlags.Alias?checker.getAliasedSymbol(s):s;}
function info(n) {return {...loc(n),text:n.getText(),sourceLine:n.getSourceFile().text.split('\n')[loc(n).line-1].trimEnd()};}
function walk(n,fn){fn(n);ts.forEachChild(n,c=>walk(c,fn));}
function propertyName(n){return n && (ts.isIdentifier(n)||ts.isStringLiteralLike(n)||ts.isNumericLiteral(n))?n.text:undefined;}
const functions=[],calls=[],sites=[],propertyDefs=[],assignments=[];
for(const file of files) walk(file,n=>{
  if((ts.isFunctionDeclaration(n)||ts.isFunctionExpression(n)||ts.isArrowFunction(n)||ts.isMethodDeclaration(n))&&n.body){
    const sig=checker.getSignatureFromDeclaration(n);
    let readsArguments=false;
    function readCount(x){if(x!==n.body&&ts.isFunctionLike(x))return;if(ts.isPropertyAccessExpression(x)&&x.name.text==='length'&&ts.isIdentifier(x.expression)&&x.expression.text==='arguments'){const s=checker.getSymbolAtLocation(x.expression);if(s&&!s.declarations?.length)readsArguments=true;}ts.forEachChild(x,readCount);}
    readCount(n.body);
    const at=n.name || n;
    const type=checker.getTypeAtLocation(at);
    functions.push({node:n,type,readsArguments,key:key(n),...info(n),name:n.name?.getText() || '<anonymous>',N:n.parameters.length,signature:sig&&checker.signatureToString(sig,undefined,ts.TypeFormatFlags.NoTruncation),rest:n.parameters.some(p=>!!p.dotDotDotToken),thisParameter:n.parameters.some(p=>p.name.kind===ts.SyntaxKind.Identifier&&p.name.text==='this')});
  }
  if(ts.isCallExpression(n)){
    calls.push(n);
    const e=n.expression;
    const bind=ts.isPropertyAccessExpression(e)&&e.name.text==='bind';
    const mb=symbol(ts.isPropertyAccessExpression(e)?e.name:e)?.declarations?.some(d=>ts.isFunctionDeclaration(d)&&d.name?.text==='maybeBind'&&path.relative(root,d.getSourceFile().fileName)==='src/compiler/core.ts');
    if(bind||mb){
      const member=bind?e.expression:n.arguments[1];
      const type=checker.getNonNullableType(checker.getTypeAtLocation(member));
      const sigs=checker.getSignaturesOfType(type,ts.SignatureKind.Call);
      const resolved=checker.getResolvedSignature(n);
      sites.push({node:n,member,type,...info(n),kind:bind?'bind':'maybeBind',member:member.getText(),memberName:ts.isPropertyAccessExpression(member)?member.name.text:undefined,typeText:textType(type),M:sigs.map(s=>s.parameters.length),signatures:sigs.map(s=>checker.signatureToString(s,undefined,ts.TypeFormatFlags.NoTruncation)),declarations:symbol(ts.isPropertyAccessExpression(member)?member.name:member)?.declarations?.map(d=>({...info(d),body:!!d.body})),resolved:resolved&&checker.signatureToString(resolved,undefined,ts.TypeFormatFlags.NoTruncation)});
    }
  }
  if((ts.isPropertyAssignment(n)||ts.isShorthandPropertyAssignment(n)||ts.isMethodDeclaration(n)||ts.isPropertyDeclaration(n))&&n.name)propertyDefs.push(n);
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken)assignments.push(n);
});
const names=new Set(sites.map(s=>s.memberName).filter(Boolean));
const defs=[];
for(const n of propertyDefs){const name=propertyName(n.name);if(names.has(name)){
  const value=ts.isShorthandPropertyAssignment(n)?n.name:n.initializer||n;
  defs.push({...info(n),name,value:value.getText(),valueType:textType(checker.getTypeAtLocation(value)),context:ts.isObjectLiteralExpression(n.parent)?textType(checker.getContextualType(n.parent)||checker.getTypeAtLocation(n.parent)):ts.isClassLike(n.parent)?n.parent.name?.text:undefined,contextDeclarations:ts.isObjectLiteralExpression(n.parent)?checker.getContextualType(n.parent)?.symbol?.declarations?.map(d=>loc(d)):undefined});
}}
for(const n of assignments){if(ts.isPropertyAccessExpression(n.left)&&names.has(n.left.name.text))defs.push({...info(n),name:n.left.name.text,value:n.right.getText(),valueType:textType(checker.getTypeAtLocation(n.right)),receiver:n.left.expression.getText(),receiverType:textType(checker.getTypeAtLocation(n.left.expression))});}
const cleanFunction=f=>{const {node,type,text,sourceLine,...rest}=f;return {...rest,header:text.slice(0,node.body.getStart()-node.getStart()).trimEnd()};};
const result={typescript:ts.version,options,ambientRoots,root,roots:roots.map(f=>path.relative(root,f)),files:files.map(f=>({file:path.relative(root,f.fileName),sha256:crypto.createHash('sha256').update(f.text).digest('hex')})),sites:sites.map(s=>{const {node,type,...rest}=s;return rest;}),definitions:defs,functions:functions.map(cleanFunction),diagnostics:ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),...(d.file?{file:path.relative(root,d.file.fileName),line:d.file.getLineAndCharacterOfPosition(d.start).line+1}:{})}))};
fs.mkdirSync(out,{recursive:true});fs.writeFileSync(path.join(out,'inventory.json'),JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({roots:roots.length,files:files.length,functions:functions.length,sites:sites.map(s=>({...loc(s.node),kind:s.kind,member:s.member,M:s.M,type:s.typeText})),diagnostics:result.diagnostics.length},null,2));
// Preserve live API objects for the follow-on static flow pass.
module.exports={ts,fs,path,root,out,program,checker,files,roots,functions,calls,sites,propertyDefs,assignments,loc,key,info,symbol,walk,propertyName,textType,result};
