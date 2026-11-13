// Read-only stock TypeScript API analysis. Usage: NODE_PATH=stage3/api/node_modules node scan.cjs ADAPTED OUTPUT
const ts = require('typescript'), fs = require('node:fs'), path = require('node:path');
const root = path.resolve(process.argv[2]), out = path.resolve(process.argv[3]);
const config = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(config, ts.sys.readFile);
if (read.error) throw Error(ts.flattenDiagnosticMessageText(read.error.messageText, '\n'));
const parsed = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(config));
if (parsed.errors.length) throw Error(JSON.stringify(parsed.errors));
// Keep each project's options unchanged; use the API install only as a fallback
// location for the explicitly configured node type package, absent upstream.
function makeProgram(pp) {
 const host=ts.createCompilerHost(pp.options);
 host.resolveTypeReferenceDirectives=(names, containingFile, redirected, options)=>names.map(ref=>{
  const name=typeof ref==='string'?ref:ref.fileName;
  const normal=ts.resolveTypeReferenceDirective(name,containingFile,options,host,redirected).resolvedTypeReferenceDirective;
  if(normal || name!=='node')return normal;
  return ts.resolveTypeReferenceDirective(name,path.resolve(__dirname,'../../api/__context__.ts'),options,host,redirected).resolvedTypeReferenceDirective;
 });
 return ts.createProgram({rootNames:pp.fileNames,options:pp.options,projectReferences:pp.projectReferences,host});
}
const program = makeProgram(parsed);
const checker = program.getTypeChecker();
const unwrap = n => { while (n && (ts.isParenthesizedExpression(n) || ts.isAsExpression(n) || ts.isTypeAssertionExpression(n) || ts.isNonNullExpression(n) || ts.isSatisfiesExpression(n))) n=n.expression; return n; };
const outer = n => { while (n.parent && (ts.isParenthesizedExpression(n.parent) || ts.isAsExpression(n.parent) || ts.isTypeAssertionExpression(n.parent) || ts.isNonNullExpression(n.parent) || ts.isSatisfiesExpression(n.parent)) && n.parent.expression===n) n=n.parent; return n; };
const access = n => ts.isPropertyAccessExpression(n) || (ts.isElementAccessExpression(n) && n.argumentExpression && (ts.isStringLiteralLike(n.argumentExpression) || ts.isNumericLiteral(n.argumentExpression)));
const key = n => ts.isPropertyAccessExpression(n) ? n.name.text : n.argumentExpression.text;
function declarations(n, c) {
 let s = c.getSymbolAtLocation(ts.isPropertyAccessExpression(n) ? n.name : n.argumentExpression);
 if (!s) s = c.getPropertyOfType(c.getTypeAtLocation(n.expression), key(n));
 return (s?.declarations || []).filter(d => ts.isMethodSignature(d) || ts.isMethodDeclaration(d));
}
const id = d => `${d.getSourceFile().fileName}:${d.pos}:${d.end}`;
function owners(ds) {return [...new Set(ds.map(d=>{ let p=d.parent; return p.name?.getText() || (ts.isTypeLiteralNode(p)?'<type literal>':'<anonymous>'); }))].sort().join(' | ');}
function nil(n) { n=unwrap(n); return n && ((ts.isIdentifier(n)&&n.text==='undefined') || n.kind===ts.SyntaxKind.NullKeyword); }
const compares = new Set([ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken]);
function classify(n) {
 let x=outer(n), p=x.parent;
 if (p && ts.isCallExpression(p) && p.expression===x) return {class:p.questionDotToken?'optional-call':'direct-call'};
 let truth=false, positive=true, reason='';
 if (p && ts.isTypeOfExpression(p)) {
  const tx=outer(p), b=tx.parent;
  if (b && ts.isBinaryExpression(b) && b.operatorToken.kind===ts.SyntaxKind.EqualsEqualsEqualsToken) {
   const other=unwrap(b.left===tx?b.right:b.left);
   if (other && ts.isStringLiteralLike(other) && other.text==='function') {x=outer(b);truth=true;positive=[ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken].includes(b.operatorToken.kind);reason='typeof comparison';}
  }
 } else if (p && ts.isBinaryExpression(p) && compares.has(p.operatorToken.kind) && nil(p.left===x?p.right:p.left)) {
  x=outer(p);truth=true;positive=[ts.SyntaxKind.ExclamationEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(p.operatorToken.kind);reason='null/undefined comparison';
 }
 p=x.parent;
 while(p && ts.isPrefixUnaryExpression(p) && p.operator===ts.SyntaxKind.ExclamationToken) {truth=true; positive=!positive;reason=reason||'!';x=outer(p);p=x.parent;}
 let branches=[], forwardsMethodValue=!truth, sawLogicalOrLeft=false;
 // Right operands pass their value outward; they are truth tests only if an
 // enclosing consumer tests it. Left operands are explicitly in the brief.
 while(p && ts.isBinaryExpression(p) && [ts.SyntaxKind.AmpersandAmpersandToken,ts.SyntaxKind.BarBarToken].includes(p.operatorToken.kind)) {
  if(p.left===x) {truth=true;reason=reason||'logical left';branches.push(p.right);if(p.operatorToken.kind===ts.SyntaxKind.AmpersandAmpersandToken)forwardsMethodValue=false;else sawLogicalOrLeft=true;}
  x=outer(p);p=x.parent;
 }
 while(p && ts.isPrefixUnaryExpression(p) && p.operator===ts.SyntaxKind.ExclamationToken) {truth=true;forwardsMethodValue=false;positive=!positive;reason=reason||'!';x=outer(p);p=x.parent;}
 if(p && ts.isConditionalExpression(p) && p.condition===x) {truth=true; forwardsMethodValue=false; reason=reason||'conditional condition';branches.push(positive?p.whenTrue:p.whenFalse);}
 else if(p && ts.isIfStatement(p) && p.expression===x) {truth=true;forwardsMethodValue=false;reason=reason||'if condition';const b=positive?p.thenStatement:p.elseStatement;if(b)branches.push(b);}
 else if(p && ts.isWhileStatement(p) && p.expression===x) {truth=true;forwardsMethodValue=false;reason=reason||'while condition';branches.push(p.statement);}
 return truth?{class:'presence',presenceForm:reason,branches,logicalOrMayForwardMethodValue:sawLogicalOrLeft&&forwardsMethodValue}:{class:'escapes'};
}
function role(n) {const x=outer(n),p=x.parent;if(p&&ts.isBinaryExpression(p)&&p.left===x&&p.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&p.operatorToken.kind<=ts.SyntaxKind.LastAssignment)return p.operatorToken.kind===ts.SyntaxKind.EqualsToken?'write':'read-write';return 'read';}
function row(n,c,sf) {const ds=declarations(n,c), lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));return {file:path.relative(root,sf.fileName).split(path.sep).join('/'),line:lc.line+1,column:lc.character+1,method:key(n),owner:owners(ds),accessRole:role(n),sourceLine:sf.text.split(/\r?\n/)[lc.line]};}
function guards(n,ds,branches,c) {
 const ids=new Set(ds.map(id)), calls=[], seen=new Set();
 function walk(x) {if(ts.isCallExpression(x)) {const a=unwrap(x.expression);if(access(a)&&!seen.has(a)&&declarations(a,c).some(d=>ids.has(id(d)))){seen.add(a);calls.push({receiver:a.expression.getText(),sameReceiver:a.expression.getText()===n.expression.getText(),line:a.getSourceFile().getLineAndCharacterOfPosition(a.getStart()).line+1});}} ts.forEachChild(x,walk);}
 branches.forEach(walk);
 return {receiver:n.expression.getText(),guardedBranchCount:branches.length,guardedCalls:calls,guardedCallCount:calls.length,everyGuardedCallSameReceiver:calls.every(x=>x.sameReceiver),hasGuardedCall:calls.length>0};
}
const sites=[];let accesses=0, resolved=0;
for(const sf of program.getSourceFiles()) {
 if(!sf.fileName.startsWith(path.join(root,'src/compiler')+path.sep))continue;
 function walk(n) {if(access(n)) {accesses++;const ds=declarations(n,checker);if(ds.length){resolved++;const cl=classify(n);if(cl.class!=='direct-call') {const r={...row(n,checker,sf),class:cl.class};if(cl.class==='presence')Object.assign(r,{presenceForm:cl.presenceForm,logicalOrMayForwardMethodValue:cl.logicalOrMayForwardMethodValue},guards(n,ds,cl.branches,checker));sites.push(r);}}} ts.forEachChild(n,walk);}
 walk(sf);
}
sites.sort((a,b)=>a.file.localeCompare(b.file)||a.line-b.line||a.column-b.column);
fs.writeFileSync(path.join(out,'sites.json'),JSON.stringify(sites,null,2)+'\n');
// All src readers: use every src project with its own config, deduplicate locations.
const configs=ts.sys.readDirectory(path.join(root,'src'),['.json'],undefined,['**/tsconfig*.json']);
const readers=new Map(), unresolved=new Map();
for(const cfg of configs) {
 const rr=ts.readConfigFile(cfg,ts.sys.readFile);if(rr.error)continue;
 const pp=ts.parseJsonConfigFileContent(rr.config,ts.sys,path.dirname(cfg));if(!pp.fileNames.length)continue;
 const pr=makeProgram(pp), c=pr.getTypeChecker();
 for(const sf of pr.getSourceFiles()) {
 if(!sf.fileName.startsWith(path.join(root,'src')+path.sep))continue;
 if(!sf.text.includes('getPositionOfLineAndCharacter'))continue;
 function walk(n){if(access(n)&&key(n)==='getPositionOfLineAndCharacter'){
 const ds=declarations(n,c), r={...row(n,c,sf),class:classify(n).class,receiver:n.expression.getText(),sourceFileLike:ds.some(d=>d.parent.name?.getText()==='SourceFileLike'),declarations:ds.map(d=>({file:path.relative(root,d.getSourceFile().fileName),owner:owners([d])}))};
 const k=`${r.file}:${r.line}:${r.column}`; const old=readers.get(k);if(!old || (!old.sourceFileLike&&r.sourceFileLike) || (!old.owner&&r.owner))readers.set(k,r);
 }ts.forEachChild(n,walk);}walk(sf);
 }
}
// Lexical inventory independently covers accesses omitted by all project root sets.
for(const f of ts.sys.readDirectory(path.join(root,'src'),['.ts','.tsx','.js','.jsx'],undefined,['**/*'])) {
 const txt=fs.readFileSync(f,'utf8');if(!txt.includes('getPositionOfLineAndCharacter'))continue;
 const sf=ts.createSourceFile(f,txt,ts.ScriptTarget.Latest,true);
 function walk(n){if(access(n)&&key(n)==='getPositionOfLineAndCharacter'){const lc=sf.getLineAndCharacterOfPosition(n.getStart()), k=`${path.relative(root,f)}:${lc.line+1}:${lc.character+1}`;if(!readers.has(k))unresolved.set(k,{file:path.relative(root,f),line:lc.line+1,column:lc.character+1,class:classify(n).class,sourceLine:txt.split(/\r?\n/)[lc.line],reason:'not in a scanned project'});}ts.forEachChild(n,walk);}walk(sf);
}
const sort=(a,b)=>a.file.localeCompare(b.file)||a.line-b.line||a.column-b.column;
fs.writeFileSync(path.join(out,'readers.json'),JSON.stringify({readers:[...readers.values()].sort(sort),uncovered:[...unresolved.values()].sort(sort)},null,2)+'\n');
fs.writeFileSync(path.join(out,'scan-meta.json'),JSON.stringify({typescript:ts.version,config:path.relative(root,config),rootFileCount:parsed.fileNames.length,compilerFiles:program.getSourceFiles().filter(sf=>sf.fileName.startsWith(path.join(root,'src/compiler')+path.sep)).length,accesses,resolvedMethodAccesses:resolved,configCount:configs.length,configDiagnostics:program.getOptionsDiagnostics().map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n'))},null,2)+'\n');
console.log(JSON.stringify({sites:sites.length,classes:sites.reduce((a,r)=>(a[r.class]=(a[r.class]||0)+1,a),{}),readers:readers.size,uncovered:unresolved.size}));
