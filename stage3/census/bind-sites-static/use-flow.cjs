// Allocation-sensitive, flow-insensitive source may-flow. This is supplementary
// evidence, NOT an emulation of Lower's IR graph. Unknown inputs remain explicit.
const m=require('./measure.cjs');
const {ts,fs,path,root,out,checker,functions,files,sites,loc,key,symbol,propertyName}=m;
const slots=[],slotKeys=[],slotTypes=[],ids=new Map(),adj=[],watch=[],queue=[],queued=new Set();
const compatibility=new Map();
const tokens=[{kind:'unknown'},{kind:'absent'}];
const fToken=new Map(),oToken=new Map(),fnByNode=new Map(functions.map(f=>[f.node,f]));
const spreads=new Map();
const fields=new Map(),returns=new Map(),thisSlots=new Map(),built=new Set(),calls=[];
const S=k=>{if(ids.has(k))return ids.get(k);const i=slots.length;ids.set(k,i);slotKeys.push(k);slots.push(new Set());adj.push(new Set());watch.push([]);return i;};
const tok=x=>{tokens.push(x);return tokens.length-1;};
const schedule=s=>{if(!queued.has(s)){queued.add(s);queue.push(s);}};
const openTypeMemo=new Map();
function openType(t,depth=4){
 if(!t)return false;const k=t.id+':'+depth;if(openTypeMemo.has(k))return openTypeMemo.get(k);openTypeMemo.set(k,false);
 let open=!!(t.flags&ts.TypeFlags.TypeParameter);
 if(!open&&depth>0){if(t.isUnionOrIntersection())open=t.types.some(x=>openType(x,depth-1));else{
  const args=t.flags&ts.TypeFlags.Object&&t.objectFlags&ts.ObjectFlags.Reference?checker.getTypeArguments(t):[];
  open=args.some(x=>openType(x,depth-1));
  if(!open)open=checker.getSignaturesOfType(t,ts.SignatureKind.Call).some(sig=>!!sig.typeParameters?.length||openType(checker.getReturnTypeOfSignature(sig),depth-1)||sig.parameters.some(p=>openType(checker.getTypeOfSymbol(p),depth-1)));
 }}openTypeMemo.set(k,open);return open;
}
function compatible(s,v){
 const token=tokens[v];if(!token||v<2)return true;
 if(slotTypes[s]===undefined){const k=slotKeys[s];let t=null;try{if(k&&typeof k.getSourceFile==='function')t=checker.getTypeAtLocation(k);else if(k&&k.declarations)t=checker.getTypeOfSymbol(k);}catch{}slotTypes[s]=t;}
 const target=slotTypes[s];if(!target||target.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown|ts.TypeFlags.TypeParameter)||openType(target))return true;
 let own=token.type;if(!own&&token.node){try{own=checker.getTypeAtLocation(token.node.name||token.node);if(token.kind==='function'||token.kind==='bound-wrapper')own=checker.getNonNullableType(own);token.type=own;}catch{}}
 if(!own)return true;
 const k=v+':'+target.id;if(compatibility.has(k))return compatibility.get(k);
 let ok;
 if(['object','array','instance','class'].includes(token.kind))ok=true;else ok=checker.isTypeAssignableTo(own,target);
 compatibility.set(k,ok);return ok;
}
function seed(s,v){if(!compatible(s,v))return;if(!slots[s].has(v)){slots[s].add(v);schedule(s);}}
function link(a,b){if(a===b||adj[a].has(b))return;adj[a].add(b);for(const v of slots[a])seed(b,v);}
function action(ss,fn){for(const s of ss)watch[s].push(fn);fn();}
function field(o,name){const k=o+':'+name;if(!fields.has(k)){const dest=S('field:'+k);fields.set(k,dest);for(const source of spreads.get(o)||[])readField(source,name,dest);}return fields.get(k);}
function fnToken(n){if(!fToken.has(n))fToken.set(n,tok({kind:'function',node:n,key:key(n),N:n.parameters.length}));return fToken.get(n);}
function objToken(n,kind='object'){if(!oToken.has(n))oToken.set(n,tok({kind,node:n,key:key(n)}));return oToken.get(n);}
function symSlot(s){return S(s||'unknown-symbol');}
function enclosing(n){for(let p=n.parent;p;p=p.parent)if(fnByNode.has(p)||ts.isConstructorDeclaration(p))return p;}
function thisSlot(fn){if(!thisSlots.has(fn))thisSlots.set(fn,S({this:fn}));return thisSlots.get(fn);}
function returnSlot(fn){if(!returns.has(fn))returns.set(fn,S({returns:fn}));return returns.get(fn);}
function bindPattern(n,from){
 if(ts.isIdentifier(n)){link(from,symSlot(symbol(n)));return;}
 if(ts.isObjectBindingPattern(n))for(const b of n.elements){
  if(b.dotDotDotToken){bindPattern(b.name,from);continue;}
  const name=propertyName(b.propertyName||b.name);const dest=S(b);
  if(name)readField(from,name,dest);else seed(dest,0);
  if(b.initializer)link(expr(b.initializer),dest);bindPattern(b.name,dest);
 }else if(ts.isArrayBindingPattern(n))n.elements.forEach((b,i)=>{if(ts.isBindingElement(b)){const dest=S(b);readField(from,String(i),dest);bindPattern(b.name,dest);}});
}
function readField(receiver,name,dest){
 action([receiver],()=>{for(const v of slots[receiver]){
  if(v===0)seed(dest,0);
  else if(tokens[v].kind==='object'||tokens[v].kind==='array'||tokens[v].kind==='instance'||tokens[v].kind==='class'){
   link(field(v,name),dest);if(name!=='*')link(field(v,'*'),dest);
  }
 }});
}
function writeField(receiver,name,source){action([receiver],()=>{for(const v of slots[receiver])if(['object','array','instance','class'].includes(tokens[v].kind))link(source,field(v,name));});}
function copyObject(source,destToken){
 // Name universe is finite and obtained through the parser, never text matching.
 if(!spreads.has(destToken))spreads.set(destToken,new Set());if(spreads.get(destToken).has(source))return;spreads.get(destToken).add(source);for(const [k,to] of fields)if(k.startsWith(destToken+':'))readField(source,k.slice(k.indexOf(':')+1),to);
}
function declSymbol(s){
 const slot=symSlot(s);if(!s){seed(slot,0);return slot;}
 if(built.has(s))return slot;built.add(s);
 for(const d of s.declarations||[]){
  if(fnByNode.has(d))seed(slot,fnToken(d));
  else if(ts.isVariableDeclaration(d)&&d.initializer)link(expr(d.initializer),slot);
  else if(ts.isEnumMember(d)){}
  else if(ts.isParameter(d)){seed(slot,0);if(d.initializer)link(expr(d.initializer),slot);}
  else if(ts.isFunctionDeclaration(d)&&!d.body){const impl=s.declarations.find(x=>ts.isFunctionDeclaration(x)&&x.body);if(impl)seed(slot,fnToken(impl));else seed(slot,0);}
  else if(ts.isImportEqualsDeclaration(d)||d.getSourceFile().isDeclarationFile)seed(slot,0);
 }
 return slot;
}
function callFunction(fn,call,args,dest,recv){
 const cacheKey=key(call)+'>'+key(fn);if(callEdges.has(cacheKey))return;callEdges.add(cacheKey);
 const ps=fn.parameters.filter(p=>!(ts.isIdentifier(p.name)&&p.name.text==='this'));
 ps.forEach((p,i)=>{if(p.dotDotDotToken){for(let j=i;j<args.length;j++)bindPattern(p.name,args[j]);}
 else if(args[i]!==undefined)bindPattern(p.name,args[i]);else bindPattern(p.name,S('absent'));});
 if(recv!==undefined)link(recv,thisSlot(fn));
 if(!(ts.isFunctionDeclaration(fn)&&fn.name?.text==='changeCompilerHostLikeToUseCache'&&path.relative(root,fn.getSourceFile().fileName)==='src/compiler/program.ts'))link(returnSlot(fn),dest);
}
const callEdges=new Set(),fieldNames=new Set();
const wantedNames=new Set(sites.map(s=>s.memberName).filter(Boolean));
const relevantMemo=new Map();
function relevantType(t,depth=2){
 if(!t)return true;
 const k=t.id+':'+depth;if(relevantMemo.has(k))return relevantMemo.get(k);
 relevantMemo.set(k,false);
 let yes=!!(t.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown|ts.TypeFlags.TypeParameter));
 if(t.isUnionOrIntersection())yes=t.types.some(x=>relevantType(x,depth));
 else if(t.flags&ts.TypeFlags.Object){
  yes=[...wantedNames].some(name=>!!checker.getPropertyOfType(t,name));
  if(!yes&&depth>0)yes=checker.getPropertiesOfType(t).some(p=>relevantType(checker.getTypeOfSymbol(p),depth-1));
 }
 relevantMemo.set(k,yes);return yes;
}
function relevantObject(n){return relevantType(checker.getTypeAtLocation(n));}

function gather(n){if((ts.isPropertyAccessExpression(n)||ts.isPropertyAssignment(n)||ts.isShorthandPropertyAssignment(n)||ts.isMethodDeclaration(n)||ts.isPropertyDeclaration(n))&&n.name){const name=propertyName(n.name);if(name)fieldNames.add(name);}ts.forEachChild(n,gather);}
for(const f of files)gather(f);
seed(S('absent'),1);
function expr(n){
 if(!n)return S('absent');
 if(ts.isIdentifier(n))return declSymbol(ts.isShorthandPropertyAssignment(n.parent)?checker.getShorthandAssignmentValueSymbol(n.parent):symbol(n));
 const dest=S(n);if(built.has(n))return dest;built.add(n);
 if(fnByNode.has(n)){seed(dest,fnToken(n));return dest;}
 if(n.kind===ts.SyntaxKind.ThisKeyword){const fn=enclosing(n);if(fn)link(thisSlot(fn),dest);else seed(dest,0);}
 else if(ts.isParenthesizedExpression(n)||ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isNonNullExpression(n)||ts.isSatisfiesExpression(n))link(expr(n.expression),dest);
 else if(ts.isConditionalExpression(n)){link(expr(n.whenTrue),dest);link(expr(n.whenFalse),dest);}
 else if(ts.isBinaryExpression(n)){
  if([ts.SyntaxKind.BarBarToken,ts.SyntaxKind.AmpersandAmpersandToken,ts.SyntaxKind.QuestionQuestionToken].includes(n.operatorToken.kind)){link(expr(n.left),dest);link(expr(n.right),dest);}
  else if(n.operatorToken.kind===ts.SyntaxKind.EqualsToken){assign(n.left,expr(n.right));link(expr(n.right),dest);}
 }else if(ts.isObjectLiteralExpression(n)){
  if(!relevantObject(n))return dest;
  const o=objToken(n);seed(dest,o);
  for(const p of n.properties){
   if(ts.isSpreadAssignment(p)){copyObject(expr(p.expression),o);continue;}
   const name=propertyName(p.name);if(!name)continue;
   if(ts.isPropertyAssignment(p))link(expr(p.initializer),field(o,name));
   else if(ts.isShorthandPropertyAssignment(p))link(declSymbol(checker.getShorthandAssignmentValueSymbol(p)),field(o,name));
   else if(ts.isMethodDeclaration(p)&&p.body){seed(field(o,name),fnToken(p));seed(thisSlot(p),o);}
   else if(ts.isGetAccessorDeclaration(p)&&p.body)link(returnSlot(p),field(o,name));
  }
 }else if(ts.isArrayLiteralExpression(n)){
  if(!relevantObject(n))return dest;
  const o=objToken(n,'array');seed(dest,o);n.elements.forEach((e,i)=>{const v=expr(ts.isSpreadElement(e)?e.expression:e);link(v,field(o,String(i)));link(v,field(o,'*'));});
 }else if(ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)){
  const name=ts.isPropertyAccessExpression(n)?n.name.text:n.argumentExpression&&propertyName(n.argumentExpression);
  const s=ts.isPropertyAccessExpression(n)?symbol(n.name):undefined;
  // Module-qualified names are bound symbols, rather than instance field lookups.
  if(s?.declarations?.some(d=>ts.isFunctionDeclaration(d)||ts.isVariableDeclaration(d)&&
      (ts.isSourceFile(d.parent.parent.parent)||ts.isModuleBlock(d.parent.parent.parent))))link(declSymbol(s),dest);
  else readField(expr(n.expression),name||'*',dest);
 }else if(ts.isCallExpression(n)){
  const callee=n.expression,args=n.arguments.map(a=>expr(ts.isSpreadElement(a)?a.expression:a));
  if(ts.isPropertyAccessExpression(callee)&&callee.name.text==='bind'){
   const type=checker.getNonNullableType(checker.getTypeAtLocation(callee.expression));
   const sig=checker.getSignaturesOfType(type,ts.SignatureKind.Call);
   const v=tok({kind:'bound-wrapper',node:n,key:key(n),N:sig.length===1?Math.max(0,sig[0].parameters.length-(args.length-1)):null,target:callee.expression.getText()});seed(dest,v);
  }else if(ts.isIdentifier(callee)&&symbol(callee)?.declarations?.some(d=>ts.isFunctionDeclaration(d)&&d.body&&d.name?.text==='changeCompilerHostLikeToUseCache'&&path.relative(root,d.getSourceFile().fileName)==='src/compiler/program.ts')){
   // This cache helper is called on distinct hosts. Keep each returned originals record
   // at its call site instead of mixing watch-compiler and solution-host originals.
   const o=objToken(n);seed(dest,o);
   for(const name of ['ReadFile','FileExists','DirectoryExists','CreateDirectory','WriteFile'])readField(args[0],name[0].toLowerCase()+name.slice(1),field(o,'original'+name));
   const d=symbol(callee).declarations.find(d=>ts.isFunctionDeclaration(d)&&d.body);callFunction(d,n,args,dest,undefined);
  }else if(ts.isIdentifier(callee)&&symbol(callee)?.declarations?.some(d=>ts.isFunctionDeclaration(d)&&path.relative(root,d.getSourceFile().fileName)==='src/compiler/core.ts'&&['assign','copyProperties','clone','extend'].includes(d.name?.text))){
   const name=callee.text;
   if(name==='assign'||name==='copyProperties'){
    if(args.length){if(name==='assign')link(args[0],dest);action([args[0]],()=>{for(const o of slots[args[0]])if(['object','instance','class'].includes(tokens[o]?.kind))for(const a of args.slice(1))copyObject(a,o);});}
   }else{
    const o=objToken(n);seed(dest,o);for(const a of args)copyObject(a,o);
   }
  }else if(ts.isPropertyAccessExpression(callee)&&ts.isIdentifier(callee.expression)&&callee.expression.text==='Object'&&callee.name.text==='assign'){
   if(args.length){link(args[0],dest);action([args[0]],()=>{for(const o of slots[args[0]])if(tokens[o]?.kind==='object')for(const a of args.slice(1))copyObject(a,o);});}
  }else{
   const cs=expr(callee),receiver=ts.isPropertyAccessExpression(callee)?expr(callee.expression):undefined;
   const d=checker.getResolvedSignature(n)?.declaration;
   if(d?.body&&ts.isIdentifier(callee))callFunction(d,n,args,dest,receiver);
   action([cs],()=>{for(const v of slots[cs]){const t=tokens[v];if(t.kind==='function')callFunction(t.node,n,args,dest,receiver);else if(v===0)seed(dest,0);}});
   if(!d?.body&&!(ts.isPropertyAccessExpression(callee)&&symbol(callee.name)?.declarations?.some(x=>x.body)))seed(dest,0);
  }
 }else if(ts.isNewExpression(n)){
  if(!relevantObject(n))return dest;
  const o=objToken(n,'instance');seed(dest,o);
  const s=symbol(n.expression),d=s?.declarations?.find(ts.isClassDeclaration);
  if(d){for(const member of d.members){const name=propertyName(member.name);if(ts.isMethodDeclaration(member)&&member.body&&name){seed(field(o,name),fnToken(member));seed(thisSlot(member),o);}else if(ts.isPropertyDeclaration(member)&&member.initializer&&name)link(expr(member.initializer),field(o,name));else if(ts.isConstructorDeclaration(member)&&member.body){seed(thisSlot(member),o);callFunction(member,n,(n.arguments||[]).map(expr),S('discard'),undefined);}}}
  else seed(dest,0);
 }else if(n.kind===ts.SyntaxKind.NullKeyword||ts.isVoidExpression(n))seed(dest,1);
 else if(ts.isAwaitExpression(n))link(expr(n.expression),dest);
 return dest;
}
function assign(n,source){
 if(ts.isIdentifier(n))link(source,declSymbol(symbol(n)));
 else if(ts.isPropertyAccessExpression(n))writeField(expr(n.expression),n.name.text,source);
 else if(ts.isElementAccessExpression(n))writeField(expr(n.expression),propertyName(n.argumentExpression)||'*',source);
 else if(ts.isObjectLiteralExpression(n))for(const p of n.properties){if(ts.isShorthandPropertyAssignment(p)){const d=S(p);readField(source,p.name.text,d);assign(p.name,d);}else if(ts.isPropertyAssignment(p)){const d=S(p);readField(source,propertyName(p.name)||'*',d);assign(p.initializer,d);}}
 else if(ts.isArrayLiteralExpression(n))n.elements.forEach((p,i)=>{const d=S(p);readField(source,String(i),d);assign(p,d);});
}
function visit(n){
 if(ts.isVariableDeclaration(n)&&n.initializer)bindPattern(n.name,expr(n.initializer));
 if(ts.isParameter(n)&&n.initializer)bindPattern(n.name,expr(n.initializer));
 if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken)assign(n.left,expr(n.right));
 if(ts.isReturnStatement(n)&&n.expression){const owner=enclosing(n);if(owner)link(expr(n.expression),returnSlot(owner));}
 if(ts.isArrowFunction(n)&&!ts.isBlock(n.body))link(expr(n.body),returnSlot(n));
 if(ts.isCallExpression(n)||ts.isNewExpression(n)||ts.isObjectLiteralExpression(n))expr(n);
 ts.forEachChild(n,visit);
}
function activateLiteral(n){
 if(oToken.has(n))return false;
 const dest=S(n);
 if(ts.isObjectLiteralExpression(n)){
  const o=objToken(n);seed(dest,o);
  for(const p of n.properties){if(ts.isSpreadAssignment(p)){copyObject(expr(p.expression),o);continue;}const name=propertyName(p.name);if(!name)continue;
   if(ts.isPropertyAssignment(p))link(expr(p.initializer),field(o,name));
   else if(ts.isShorthandPropertyAssignment(p))link(declSymbol(checker.getShorthandAssignmentValueSymbol(p)),field(o,name));
   else if(ts.isMethodDeclaration(p)&&p.body){seed(field(o,name),fnToken(p));seed(thisSlot(p),o);}
  }
 }else if(ts.isArrayLiteralExpression(n)){
  const o=objToken(n,'array');seed(dest,o);n.elements.forEach((e,i)=>{const v=expr(ts.isSpreadElement(e)?e.expression:e);link(v,field(o,String(i)));link(v,field(o,'*'));});
 }else return false;
 return true;
}

for(const f of files)visit(f);
const siteSlots=sites.map(s=>expr(s.kind==='bind'?s.node.expression.expression:s.node.arguments[1]));
let steps=0;
while(queue.length){const s=queue.pop();queued.delete(s);for(const d of adj[s])for(const v of slots[s])seed(d,v);for(const fn of watch[s])fn();if(++steps%1000000===0)console.error('flow steps '+steps+' pending '+queue.length);if(steps>5000000)throw Error('bounded source-flow budget exceeded; no partial fixed point published');}
const rows=sites.map((s,i)=>({...loc(s.node),member:s.member,kind:s.kind,unknown:slots[siteSlots[i]].has(0),values:[...slots[siteSlots[i]]].filter(v=>['function','bound-wrapper'].includes(tokens[v].kind)).map(v=>{const {node,type,...t}=tokens[v];return {...t,name:fnByNode.get(node)?.name,signature:fnByNode.get(node)?.signature};}).sort((a,b)=>a.key.localeCompare(b.key))}));
fs.writeFileSync(path.join(out,'source-flow.json'),JSON.stringify({method:'use-audit flow: checker-compatible function values at expression/symbol slots; object carriers remain conservative; explicit core assign/copyProperties/clone/extend source copies; context/flow insensitive; object allocations filtered by host-bearing fields within depth 2; not an exhaustive certificate',steps,slots:slots.length,tokens:tokens.length,callEdges:callEdges.size,sites:rows},null,2)+'\n');
console.log('SOURCE FLOW '+JSON.stringify({steps,slots:slots.length,tokens:tokens.length,callEdges:callEdges.size}));
for(const row of rows)console.log(JSON.stringify(row));

module.exports={activateLiteral,m,slots,ids,adj,tokens,fToken,oToken,fields,returns,fnByNode,expr,field,queue,queued,watch,seed,link,steps};
