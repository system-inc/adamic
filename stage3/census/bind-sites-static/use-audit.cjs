// Track bound-wrapper identities through the source allocation/alias graph.
// Every use record comes from the compiler API AST and checker, never text matching.
const g=require('./use-flow.cjs');
const {m,slots,ids,adj,tokens,expr,field,queue,queued,watch,seed}=g;
const {ts,fs,path,root,out,files,sites,checker,loc,key,info,symbol,walk,propertyName}=m;
const selected=JSON.parse(fs.readFileSync(path.join(out,'candidates.json'),'utf8')).map((c,i)=>({c,site:sites[i],id:'B'+String(i+1).padStart(2,'0')})).filter(({c,site})=>site.member!=='fn'&&c.extra.length&&!c.typed.some(f=>f.rest||f.readsArguments));
if(selected.length!==15)throw Error('15 a* origin sites required');
const originByKey=new Map(selected.map((x,i)=>[key(x.site.node),i]));
const reads=[],spreads=[],copies=[],allCalls=[],bindings=[],literals=[];
function inType(n){for(let p=n.parent;p;p=p.parent){if(ts.isTypeNode(p))return true;if(ts.isExpression(p)||ts.isStatement(p))return false;}return false;}
function identifierRead(n){
 const p=n.parent;if(!p||inType(n))return false;
 if((ts.isPropertyAccessExpression(p)&&p.name===n)||((ts.isPropertyAssignment(p)||ts.isPropertyDeclaration(p)||ts.isMethodDeclaration(p)||ts.isPropertySignature(p)||ts.isMethodSignature(p))&&p.name===n))return false;
 if((ts.isVariableDeclaration(p)||ts.isParameter(p)||ts.isBindingElement(p)||ts.isFunctionDeclaration(p)||ts.isFunctionExpression(p)||ts.isClassDeclaration(p)||ts.isTypeAliasDeclaration(p)||ts.isInterfaceDeclaration(p)||ts.isModuleDeclaration(p)||ts.isEnumMember(p))&&p.name===n)return false;
 if(ts.isImportSpecifier(p)||ts.isExportSpecifier(p)||ts.isImportClause(p)||ts.isNamespaceImport(p)||ts.isLabeledStatement(p))return false;
 if(ts.isBinaryExpression(p)&&p.left===n&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken)return false;
 return true;
}
for(const file of files)walk(file,n=>{
 if((ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n))&&!inType(n)){
  if(!(ts.isBinaryExpression(n.parent)&&n.parent.left===n&&n.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken))reads.push({node:n,slot:expr(n)});
 }else if(ts.isIdentifier(n)&&identifierRead(n))reads.push({node:n,slot:expr(n)});
 if(ts.isSpreadAssignment(n))spreads.push(n);
 if(ts.isBindingElement(n)&&ts.isIdentifier(n.name))bindings.push({node:n,slot:expr(n.name)});
 if(ts.isObjectLiteralExpression(n)||ts.isArrayLiteralExpression(n))literals.push(n);
 if(ts.isCallExpression(n))allCalls.push(n);
});
function drain(){let count=0;while(queue.length){const s=queue.pop();queued.delete(s);for(const d of adj[s])for(const v of slots[s])seed(d,v);for(const fn of watch[s])fn();if(++count>3000000)throw Error('use-read propagation budget exceeded');}return count;}
let extraSteps=drain();
const lineage=new Map();
for(let t=0;t<tokens.length;t++){const i=originByKey.get(tokens[t].key);if(tokens[t].kind==='bound-wrapper'&&i!==undefined)lineage.set(t,new Set([i]));}
const originSet=slot=>{const ans=new Set();for(const t of slots[slot]||[])for(const o of lineage.get(t)||[])ans.add(o);return ans;};
let changed=true;
while(changed){changed=false;for(let t=0;t<tokens.length;t++){
 const v=tokens[t];if(v.kind!=='bound-wrapper')continue;
 const target=expr(v.node.expression.expression);for(const origin of originSet(target)){
  if(!lineage.has(t))lineage.set(t,new Set());if(!lineage.get(t).has(origin)){lineage.get(t).add(origin);changed=true;}
 }
}extraSteps+=drain();}
// The one new-key carrier encountered by this audit is the compiler-host
// originals record. use-flow models that record separately at every call.
// Record any other pruned function-valued literal/array as a coverage gap below.
const carrierExpansions=0;

function topOf(n){let top=n;while(top.parent){const p=top.parent;
 if(ts.isParenthesizedExpression(p)||ts.isNonNullExpression(p)||ts.isAsExpression(p)||ts.isTypeAssertionExpression(p)||ts.isSatisfiesExpression(p)){top=p;continue;}
 if(ts.isConditionalExpression(p)&&p.condition!==top){top=p;continue;}
 if(ts.isBinaryExpression(p)&&[ts.SyntaxKind.BarBarToken,ts.SyntaxKind.QuestionQuestionToken].includes(p.operatorToken.kind)){top=p;continue;}
 break;
}return top;}
function owner(n){const a=[];for(let p=n;p;p=p.parent)if((ts.isFunctionDeclaration(p)||ts.isMethodDeclaration(p)||ts.isClassDeclaration(p)||ts.isModuleDeclaration(p))&&p.name)a.unshift(p.name.getText());return a.join('.');}
function countArgs(call,offset=0){const args=call.arguments.slice(offset);return {written:args.length,spread:args.some(ts.isSpreadElement),sourceArgs:args.map(x=>x.getText())};}
function localCallbackCalls(call,index){
 const ds=new Set(),resolved=checker.getResolvedSignature(call)?.declaration;
 if(resolved?.body)ds.add(resolved);
 for(const token of slots[expr(call.expression)]||[])if(tokens[token].kind==='function')ds.add(tokens[token].node);
 const implementations=[];
 for(const d of ds){const p=d.parameters?.[index];if(!d.body||!p||!ts.isIdentifier(p.name))continue;
  const s=symbol(p.name),found=[];
  function visit(n){if(ts.isCallExpression(n)){
   let c=n.expression;while(ts.isParenthesizedExpression(c)||ts.isNonNullExpression(c))c=c.expression;
   if(ts.isIdentifier(c)&&symbol(c)===s)found.push({...loc(n),text:n.getText(),...countArgs(n)});
   else if(ts.isPropertyAccessExpression(c)&&ts.isIdentifier(c.expression)&&symbol(c.expression)===s&&['call','apply'].includes(c.name.text))found.push({...loc(n),text:n.getText(),...countArgs(n,c.name.text==='call'?1:0),indirect:c.name.text});
  }ts.forEachChild(n,visit);}visit(d.body);
  implementations.push({declaration:loc(d),parameter:p.getText(),calls:found});
 }
 return implementations.length?implementations:undefined;
}
function callArgument(call,index){
 const sig=checker.getResolvedSignature(call),param=sig?.parameters[Math.min(index,sig.parameters.length-1)];
 const type=param&&checker.getTypeOfSymbolAtLocation(param,call),sigs=type?checker.getSignaturesOfType(checker.getNonNullableType(type),ts.SignatureKind.Call):[];
 const libArray=ts.isPropertyAccessExpression(call.expression)&&sig?.declaration?.getSourceFile().isDeclarationFile&&sig.declaration.parent?.name&&['Array','ReadonlyArray'].includes(sig.declaration.parent.name.getText());
 const method=ts.isPropertyAccessExpression(call.expression)?call.expression.name.text:undefined;
 const arrayCount=libArray?({map:3,filter:3,some:3,every:3,find:3,findIndex:3,findLast:3,findLastIndex:3,forEach:3,flatMap:3,sort:2,toSorted:2,reduce:4,reduceRight:4})[method]:undefined;
 return {callee:call.expression.getText(),argument:index,parameterType:type&&checker.typeToString(type,undefined,ts.TypeFormatFlags.NoTruncation),callbackSignatureCounts:sigs.map(s=>s.parameters.length),arrayCallbackArguments:arrayCount,local:localCallbackCalls(call,index),declaration:sig?.declaration&&loc(sig.declaration)};
}
function classify(n){
 const top=topOf(n),p=top.parent;
 if(!p)return {kind:'root',invokes:false,headline:true};
 if(ts.isCallExpression(p)&&p.expression===top){const type=checker.getTypeAtLocation(top),sig=checker.getResolvedSignature(p);return {kind:'direct-call',call:info(p),typed:!!sig&&!(type.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown)),...countArgs(p),headline:!(!!sig&&!(type.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown)))};}
 if(ts.isCallExpression(p)){const i=p.arguments.indexOf(top);if(i>=0)return {kind:'callback-or-argument',call:info(p),callback:callArgument(p,i),headline:true};}
 if(ts.isPropertyAccessExpression(p)&&p.expression===top&&['bind','call','apply'].includes(p.name.text)){
  const call=p.parent;if(ts.isCallExpression(call)&&call.expression===p)return {kind:p.name.text,call:info(call),...countArgs(call,p.name.text==='call'||p.name.text==='bind'?1:0),headline:true};
 }
 if(ts.isBinaryExpression(p)&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken&&p.right===top)return {kind:'store',destination:p.left.getText(),use:info(p),headline:true};
 if(ts.isVariableDeclaration(p)&&p.initializer===top)return {kind:'alias-store',destination:p.name.getText(),use:info(p),headline:true};
 if(ts.isPropertyAssignment(p)&&p.initializer===top)return {kind:'field-store',destination:p.name.getText(),use:info(p),headline:true};
 if(ts.isShorthandPropertyAssignment(p))return {kind:'field-store',destination:p.name.getText(),use:info(p),headline:true};
 if(ts.isReturnStatement(p)||ts.isArrowFunction(p)&&p.body===top)return {kind:'return',use:info(p),headline:true};
 if(ts.isArrayLiteralExpression(p)||ts.isSpreadElement(p))return {kind:'array-store-or-spread',use:info(p),headline:true};
 if(ts.isPrefixUnaryExpression(p)||ts.isTypeOfExpression(p)||ts.isConditionalExpression(p)&&p.condition===top||ts.isIfStatement(p)||ts.isBinaryExpression(p)&&[ts.SyntaxKind.AmpersandAmpersandToken,ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(p.operatorToken.kind))return {kind:'guard-or-comparison',use:info(p),headline:false};
 if(ts.isPropertyAccessExpression(p))return {kind:'property-observation',use:info(p),headline:false};
 if(ts.isBinaryExpression(p)&&p.left===top&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken)return undefined;
 return {kind:'other-'+ts.SyntaxKind[p.kind],use:info(p),headline:true};
}
const events=[],direct=[],seen=new Map();
function record(node,origins,data){if(!data||!origins.size)return;
 for(const o of [...origins]){
  // The creation store and every expression inside that origin's own bind are the seed, not a downstream use.
  const initial=selected[o].site.node;
  if(node.getSourceFile()===initial.getSourceFile()&&node.pos>=initial.pos&&node.end<=initial.end)origins.delete(o);
 }
 if(!origins.size)return;
 const k=key(node)+' '+data.kind;
 let e=seen.get(k);if(!e){e={...loc(node),read:node.getText(),owner:owner(node),...data,origins:new Set()};seen.set(k,e);events.push(e);}for(const o of origins)e.origins.add(o);
}
for(const b of bindings){const origins=originSet(b.slot);if(origins.size)record(b.node,origins,{kind:'destructure-store',destination:b.node.name.getText(),use:info(b.node),headline:true});}
for(const r of reads){const origins=originSet(r.slot);if(!origins.size)continue;const c=classify(r.node);record(r.node,origins,c);}
function visibleField(expression,name){const type=checker.getNonNullableType(checker.getTypeAtLocation(expression));if(type.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown|ts.TypeFlags.TypeParameter))return true;return !!checker.getPropertyOfType(type,name);}
for(const n of spreads){const source=expr(n.expression),os=new Set(),members=[];
 for(const t of slots[source])if(['object','instance','class'].includes(tokens[t].kind)){
  for(const [k,slot] of g.fields){if(!k.startsWith(t+':')||!visibleField(n.expression,k.slice(k.indexOf(':')+1)))continue;const origins=originSet(slot);if(origins.size){members.push(k.slice(k.indexOf(':')+1));for(const o of origins)os.add(o);}}
 }
 if(os.size)record(n,os,{kind:'object-spread',members:[...new Set(members)].sort(),use:info(n),headline:true});
}
for(const call of allCalls){
 const c=call.expression,s=ts.isIdentifier(c)?symbol(c):undefined;
 const core=s?.declarations?.some(d=>ts.isFunctionDeclaration(d)&&path.relative(root,d.getSourceFile().fileName)==='src/compiler/core.ts');
 let indexes;
 if(core&&['copyProperties','assign'].includes(c.text))indexes=call.arguments.map((_,i)=>i).slice(1);
 else if(core&&['clone','extend'].includes(c.text))indexes=call.arguments.map((_,i)=>i);
 else if(ts.isPropertyAccessExpression(c)&&ts.isIdentifier(c.expression)&&c.expression.text==='Object'&&c.name.text==='assign')indexes=call.arguments.map((_,i)=>i).slice(1);
 if(!indexes)continue;
 for(const index of indexes){const source=expr(call.arguments[index]),os=new Set(),members=[];
  for(const t of slots[source]||[])if(['object','instance','class'].includes(tokens[t].kind))for(const [k,slot] of g.fields){if(!k.startsWith(t+':')||!visibleField(call.arguments[index],k.slice(k.indexOf(':')+1)))continue;const origins=originSet(slot);if(origins.size){members.push(k.slice(k.indexOf(':')+1));for(const o of origins)os.add(o);}}
  if(os.size)record(call.arguments[index],os,{kind:'object-copy',members:[...new Set(members)].sort(),argument:index,use:info(call),headline:true});
 }
}
// Reflect.call/apply forwards functions in argument 0; ordinary argument classification retains it.
for(const e of events){e.origins=[...e.origins].map(i=>selected[i].id).sort();}
events.sort((a,b)=>a.file.localeCompare(b.file)||a.line-b.line||a.column-b.column||a.kind.localeCompare(b.kind));
events.forEach((e,i)=>e.id='U'+String(i+1).padStart(3,'0'));
const originRows=selected.map((x,i)=>{const uses=events.filter(e=>e.origins.includes(x.id));return {id:x.id,...loc(x.site.node),text:x.site.node.getText(),member:x.site.member,M:x.site.M[0],uses:uses.map(e=>e.id),nonDirect:uses.filter(e=>e.headline).map(e=>e.id),directCalls:uses.filter(e=>e.kind==='direct-call').map(e=>e.id)};});
const boundFieldsByToken=new Map();
for(const [k,slot] of g.fields){const origins=originSet(slot);if(!origins.size)continue;const t=Number(k.slice(0,k.indexOf(':')));if(!boundFieldsByToken.has(t))boundFieldsByToken.set(t,[]);boundFieldsByToken.get(t).push({name:k.slice(k.indexOf(':')+1),origins});}
function carrierOrigins(expression){const os=new Set(),members=[];for(const t of slots[expr(expression)]||[])for(const f of boundFieldsByToken.get(t)||[]){if(!visibleField(expression,f.name))continue;members.push(f.name);for(const o of f.origins)os.add(o);}return {origins:[...os].map(i=>selected[i].id).sort(),members:[...new Set(members)].sort()};}

const carrierCalls=[];
for(const call of allCalls)for(let index=0;index<call.arguments.length;index++){
 const a=call.arguments[index],carrier=carrierOrigins(a);if(!carrier.origins.length)continue;
 const declarations=new Set(),d=checker.getResolvedSignature(call)?.declaration;
 if(d?.body)declarations.add(d);
 for(const t of slots[expr(call.expression)]||[])if(tokens[t].kind==='function')declarations.add(tokens[t].node);
 const dynamic=[];
 for(const decl of declarations){const param=decl.parameters?.[index];if(!param||!ts.isIdentifier(param.name)||!decl.body)continue;const ps=symbol(param.name);
  function containsParam(n){let found=false;function find(x){if(ts.isIdentifier(x)&&symbol(x)===ps)found=true;else if(!found)ts.forEachChild(x,find);}find(n);return found;}
  function find(n){if(ts.isElementAccessExpression(n)){
    let receiver=n.expression;while(ts.isParenthesizedExpression(receiver)||ts.isAsExpression(receiver)||ts.isNonNullExpression(receiver))receiver=receiver.expression;
    const keyType=n.argumentExpression&&checker.getTypeAtLocation(n.argumentExpression);
    if(ts.isIdentifier(receiver)&&symbol(receiver)===ps&&!(keyType?.flags&ts.TypeFlags.NumberLike))dynamic.push({...loc(n),text:n.getText(),kind:'element',role:ts.isBinaryExpression(n.parent)&&n.parent.left===n&&n.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken?'write':'read'});
   }
   if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&ts.isIdentifier(n.expression.expression)&&n.expression.expression.text==='Reflect'&&n.arguments[0]&&containsParam(n.arguments[0]))dynamic.push({...loc(n),text:n.getText(),kind:'Reflect'});
   ts.forEachChild(n,find);
  }find(decl.body);
 }
 carrierCalls.push({...loc(a),argument:index,call:call.getText(),callee:call.expression.getText(),...carrier,implementations:[...declarations].map(loc),dynamic});
}
function seedStorage(site){let n=site.node;while(n.parent){const p=n.parent;
 if(ts.isParenthesizedExpression(p)||ts.isNonNullExpression(p)||ts.isAsExpression(p)||ts.isConditionalExpression(p)&&p.condition!==n||ts.isBinaryExpression(p)&&[ts.SyntaxKind.BarBarToken,ts.SyntaxKind.QuestionQuestionToken].includes(p.operatorToken.kind)){n=p;continue;}
 let destination,type,contract;
 if(ts.isPropertyAssignment(p)&&p.initializer===n){const ctxt=checker.getContextualType(p.parent)||checker.getTypeAtLocation(p.parent);destination=checker.typeToString(ctxt)+'.'+p.name.getText();const prop=checker.getPropertyOfType(ctxt,p.name.getText());type=prop&&checker.getTypeOfSymbolAtLocation(prop,p);contract=prop?.declarations?.map(loc);}
 else if(ts.isVariableDeclaration(p)&&p.initializer===n){destination=p.name.getText();type=checker.getTypeAtLocation(p.name);contract=[loc(p)];}
 else if(ts.isBinaryExpression(p)&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken&&p.right===n){destination=p.left.getText();type=checker.getTypeAtLocation(p.left);contract=symbol(ts.isPropertyAccessExpression(p.left)?p.left.name:p.left)?.declarations?.map(loc);}
 if(destination)return {destination,type:type&&checker.typeToString(type,undefined,ts.TypeFormatFlags.NoTruncation),contract,at:loc(p)};
 return {destination:'unresolved storage',parent:info(p)};
 }return {};
}
for(let i=0;i<originRows.length;i++)originRows[i].storage=seedStorage(selected[i].site);
const memberNames=new Set(selected.map(x=>x.site.memberName));
const surfaceReads=[];
for(const file of files)walk(file,n=>{
 const name=ts.isPropertyAccessExpression(n)?n.name.text:ts.isElementAccessExpression(n)&&n.argumentExpression&&ts.isStringLiteralLike(n.argumentExpression)?n.argumentExpression.text:undefined;
 if(!memberNames.has(name)||inType(n))return;
 const type=checker.getNonNullableType(checker.getTypeAtLocation(n));if(!checker.getSignaturesOfType(type,ts.SignatureKind.Call).length)return;
 const kind=classify(n);if(!kind||kind.kind==='direct-call')return;
 surfaceReads.push({...loc(n),name,text:n.getText(),kind:kind?.kind,origins:[...originSet(expr(n))].map(i=>selected[i].id),declarations:symbol(ts.isPropertyAccessExpression(n)?n.name:n)?.declarations?.map(loc)});
});
const unmodeledCarriers=[];
for(const event of events)if(['field-store','array-store-or-spread'].includes(event.kind)){
 const read=reads.find(r=>key(r.node)===event.file+':'+event.line+':'+event.column)?.node;
 const container=read?.parent&&ts.isShorthandPropertyAssignment(read.parent)?read.parent.parent:read?.parent&&ts.isPropertyAssignment(read.parent)?read.parent.parent:read?.parent&&ts.isArrayLiteralExpression(read.parent)?read.parent:undefined;
 if(!container||g.oToken.has(container))continue;
 const fn=(()=>{for(let p=container.parent;p;p=p.parent)if(ts.isFunctionDeclaration(p))return p;})();
 const modeledCache=fn?.name?.text==='changeCompilerHostLikeToUseCache'&&path.relative(root,fn.getSourceFile().fileName)==='src/compiler/program.ts';
 if(!modeledCache)unmodeledCarriers.push({use:event.id,...loc(container),origins:event.origins});
}
const output={method:'checker-constrained bound-wrapper identity/rebinding lineage; conservative object carriers; source core copy helpers; all AST member/alias reads, visible-key copies, carrier-call and member-contract surface cross-checks; seed storage excluded; context/flow insensitive, not an exhaustive flow certificate',extraSteps,carrierExpansions,unmodeledCarriers,readExpressionsScanned:reads.length,sourceCallsScanned:allCalls.length,objectSpreadsScanned:spreads.length,origins:originRows,events,carrierCalls,surfaceReads,headlineCount:originRows.filter(s=>s.nonDirect.length).length,headlineSites:originRows.filter(s=>s.nonDirect.length).map(s=>s.id)};
fs.writeFileSync(path.join(out,'a-star-uses.json'),JSON.stringify(output,null,2)+'\n');
console.log('A-STAR USES '+JSON.stringify({count:output.headlineCount,sites:output.headlineSites,events:events.length,extraSteps}));
for(const s of originRows)console.log(JSON.stringify(s));
