// Measurement only. Usage: node census.cjs TREE > census.json
const ts = require('../../api/node_modules/typescript');
const fs = require('fs'), path = require('path');
const root = path.resolve(process.argv[2]);
function files(d) { return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(d,e.name)):/\.tsx?$/.test(e.name)?[path.join(d,e.name)]:[]); }
const roots=files(path.join(root,'src'));
const p=ts.createProgram(roots,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.CommonJS,moduleResolution:ts.ModuleResolutionKind.Node10,skipLibCheck:true});
const c=p.getTypeChecker(), sources=p.getSourceFiles().filter(s=>roots.includes(s.fileName));
const nodes=[]; for(const s of sources) { function visit(n){nodes.push(n);ts.forEachChild(n,visit)}visit(s); }
const loc=n=>({file:path.relative(root,n.getSourceFile().fileName),line:n.getSourceFile().getLineAndCharacterOfPosition(n.getStart()).line+1,expression:n.getText()});
function symbol(n){if(!n)return undefined;let s=c.getSymbolAtLocation(n);if(s?.flags&ts.SymbolFlags.Alias)try{s=c.getAliasedSymbol(s)}catch{}return s;}
function enumOf(n){const t=c.getTypeAtLocation(n);const s=t.symbol;return (t.flags & ts.TypeFlags.Object) && (s?.flags & ts.SymbolFlags.RegularEnum)?s:undefined;}
const flow=new Map();
function seed(s,e,trace){if(!s||!e)return false;let m=flow.get(s);if(!m)flow.set(s,m=new Map());if(m.has(e))return false;m.set(e,trace);return true;}
function values(n){if(!n)return new Map();const e=enumOf(n);if(e)return new Map([[e,[loc(n)]]]);if(ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isParenthesizedExpression(n)||ts.isNonNullExpression(n))return values(n.expression);return flow.get(symbol(n))||new Map();}
let changed=true,rounds=0;
while(changed && rounds++<100){changed=false;for(const n of nodes){
 if(ts.isVariableDeclaration(n)&&n.initializer)for(const [e,tr]of values(n.initializer))changed=seed(symbol(n.name),e,[...tr,loc(n)])||changed;
 if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken)for(const[e,tr]of values(n.right))changed=seed(symbol(n.left),e,[...tr,loc(n)])||changed;
 if(ts.isCallExpression(n)){const d=c.getResolvedSignature(n)?.declaration;if(d?.parameters)n.arguments.forEach((a,i)=>{for(const[e,tr]of values(a))changed=seed(symbol(d.parameters[i]?.name),e,[...tr,loc(n)])||changed;});}
}}
const writes=[],reads=[],escapes=[],objectUses=[];
function func(n){for(let a=n.parent;a;a=a.parent)if(ts.isFunctionLike(a))return a.name?.getText()|| (ts.isVariableDeclaration(a.parent)?a.parent.name.getText():'<anonymous>');return '<module>';}
function typePosition(n){for(let a=n;a;a=a.parent){if(ts.isTypeNode(a))return true;if(ts.isStatement(a)||ts.isExpressionStatement(a))break;}return false;}
for(const n of nodes){
 if(ts.isIdentifier(n)&&enumOf(n)&&!typePosition(n)){
 const par=n.parent;
 if(!((ts.isPropertyAccessExpression(par)||ts.isElementAccessExpression(par))&&par.expression===n) && !ts.isEnumDeclaration(par)&&!ts.isImportSpecifier(par)&&!ts.isExportSpecifier(par))objectUses.push({...loc(n),parentKind:ts.SyntaxKind[par.kind],parent:par.getText().slice(0,500)});
 }
 if(ts.isCallExpression(n))n.arguments.forEach((a,i)=>{if(values(a).size)escapes.push({...loc(a),callee:n.expression.getText(),argument:i,enums:[...values(a).keys()].map(e=>e.name)});});
 if(!(ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)))continue;
 const parent=n.parent;const write=ts.isBinaryExpression(parent)&&parent.left===n&&ts.isAssignmentOperator(parent.operatorToken.kind)||ts.isDeleteExpression(parent)||((ts.isPrefixUnaryExpression(parent)||ts.isPostfixUnaryExpression(parent))&&(parent.operator===ts.SyntaxKind.PlusPlusToken||parent.operator===ts.SyntaxKind.MinusMinusToken));
 if(write)for(const[e,tr]of values(n.expression))writes.push({...loc(parent),enum:e.name,path:tr,function:func(n)});
 const s=symbol(n.expression); if(s?.name!=='PollingInterval'||!enumOf(n.expression)||typePosition(n)||write)continue;
 let cls='a',why='Value flows into a watch interval, or a table keyed by the configured numeric intervals.';
 if(func(n)==='createPollingIntervalBasedLevels' || (func(n)==='pollLowPollingIntervalQueue' && ts.isCallExpression(parent) && parent.expression.getText()==='pollQueue') || (ts.isElementAccessExpression(parent) && parent.argumentExpression===n) || ts.isCaseClause(parent)||ts.isBinaryExpression(parent)&&[ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(parent.operatorToken.kind)){cls='b';why='Names a level in a comparison, switch, transition, or table key. Existing code compares configured numeric identities; changing to fixed identities requires a consistent boundary.';}
 reads.push({...loc(n),function:func(n),class:cls,why});
}
console.log(JSON.stringify({typescript:ts.version,root,sourceFiles:sources.length,rounds,converged:!changed,counts:{reads:reads.length,writes:writes.length,classes:Object.fromEntries(['a','b','c'].map(k=>[k,reads.filter(r=>r.class===k).length]))},writes,enumObjectArguments:escapes,enumObjectNonMemberUses:objectUses,reads},null,2));
