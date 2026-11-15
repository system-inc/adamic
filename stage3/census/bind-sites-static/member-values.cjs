// Conservative source may-flow audit. Field names are joined across receivers.
// This is NOT the head's IR reach certificate and does not claim exact points-to sets.
const m=require('./measure.cjs');
const {ts,fs,path,root,out,checker,functions,calls,sites,propertyDefs,assignments,files,symbol,loc,key,info,walk,propertyName}=m;
const byNode=new Map(functions.map(f=>[f.node,f]));
const defs=new Map();
const writes=new Map();
const callers=new Map();
const add=(map,k,v)=>{if(k){if(!map.has(k))map.set(k,[]);map.get(k).push(v);}};
for(const n of propertyDefs){const name=propertyName(n.name);if(!name)continue;
 const v=ts.isShorthandPropertyAssignment(n)?checker.getShorthandAssignmentValueSymbol(n):n.initializer||n;
 add(defs,name,{node:n,value:v});
}
for(const n of assignments){
 if(ts.isIdentifier(n.left)) add(writes,symbol(n.left),n.right);
 if(ts.isPropertyAccessExpression(n.left)) add(defs,n.left.name.text,{node:n,value:n.right});
 if(ts.isElementAccessExpression(n.left)&&n.left.argumentExpression&&ts.isStringLiteralLike(n.left.argumentExpression))add(defs,n.left.argumentExpression.text,{node:n,value:n.right});
}
for(const c of calls){const d=checker.getResolvedSignature(c)?.declaration;if(d)add(callers,d,c);}
const nodes=[],nodeIds=new Map(),edges=[],terminals=new Map(),unknown=new Map();
function id(n){if(!n)return -1;if(nodeIds.has(n))return nodeIds.get(n);const i=nodes.length;nodes.push(n);edges.push(new Set());nodeIds.set(n,i);return i;}
function edge(a,b){if(b!==undefined&&b!==null){const ai=id(a),bi=id(b);if(ai>=0&&bi>=0)edges[ai].add(bi);}}
function why(n,s){const i=id(n);if(!unknown.has(i))unknown.set(i,new Set());unknown.get(i).add(s);}
function declarations(n,s){
 if(!s){why(n,'unresolved symbol');return;}
 for(const v of writes.get(s)||[])edge(n,v);
 if(!s.declarations?.length){why(n,'symbol without declarations');return;}
 for(const d of s.declarations){
  if(byNode.has(d)){edge(n,d);continue;}
  if(ts.isVariableDeclaration(d)||ts.isPropertyDeclaration(d)||ts.isBindingElement(d)){
   if(d.initializer)edge(n,d.initializer);
   else if(ts.isBindingElement(d)){
    const name=propertyName(d.propertyName||d.name);if(name)edge(n,'field:'+name);else why(n,'destructuring not resolved');
   }else if(!(writes.get(s)||[]).length)why(n,'uninitialized or ambient variable '+key(d));
  }else if(ts.isParameter(d)){
   const owner=d.parent,index=owner.parameters.indexOf(d);const cs=callers.get(owner)||[];
   for(const c of cs)if(c.arguments[index])edge(n,c.arguments[index]);
   if(d.initializer)edge(n,d.initializer);
   // Callable parameter values can also enter through exported/opaque callers.
   why(n,'parameter frontier '+key(d));
  }else if(ts.isMethodSignature(d)||ts.isPropertySignature(d)){
   const name=propertyName(d.name);if(name)edge(n,'field:'+name);
   why(n,'structural/external member frontier '+key(d));
  }else if(ts.isFunctionDeclaration(d)&&!d.body){
   const impl=s.declarations.find(x=>ts.isFunctionDeclaration(x)&&x.body);if(impl)edge(n,impl);else why(n,'ambient function '+key(d));
  }else why(n,'declaration '+ts.SyntaxKind[d.kind]+' '+key(d));
 }
}
for(const site of sites){const expr=site.kind==='bind'?site.node.expression.expression:site.node.arguments[1];id(expr);}
for(let i=0;i<nodes.length;i++){
 const n=nodes[i];
 if(typeof n==='string'){
  const name=n.slice(6),ds=defs.get(name)||[];
  for(const d of ds)edge(n,d.value);
  if(!ds.length)why(n,'no in-program definition for '+name);
  continue;
 }
 if(typeof n.getSourceFile!=='function'&&n.declarations){declarations(n,n);continue;}
 if(byNode.has(n)){terminals.set(i,{kind:'function',key:key(n)});continue;}
 if(ts.isIdentifier(n)){
  if(n.text==='undefined')continue;
  declarations(n,symbol(n));
 }else if(ts.isPropertyAccessExpression(n))edge(n,'field:'+n.name.text);
 else if(ts.isElementAccessExpression(n)){
  if(n.argumentExpression&&ts.isStringLiteralLike(n.argumentExpression))edge(n,'field:'+n.argumentExpression.text);
  else why(n,'dynamic property frontier '+key(n));
 }else if(ts.isParenthesizedExpression(n)||ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isNonNullExpression(n)||ts.isSatisfiesExpression(n))edge(n,n.expression);
 else if(ts.isConditionalExpression(n)){edge(n,n.whenTrue);edge(n,n.whenFalse);}
 else if(ts.isBinaryExpression(n)&&[ts.SyntaxKind.BarBarToken,ts.SyntaxKind.AmpersandAmpersandToken,ts.SyntaxKind.QuestionQuestionToken,ts.SyntaxKind.EqualsToken].includes(n.operatorToken.kind)){edge(n,n.left);edge(n,n.right);}
 else if(ts.isCallExpression(n)){
  if(ts.isPropertyAccessExpression(n.expression)&&n.expression.name.text==='bind'){
   const target=n.expression.expression,ss=checker.getSignaturesOfType(checker.getNonNullableType(checker.getTypeAtLocation(target)),ts.SignatureKind.Call);
   terminals.set(i,{kind:'bound-wrapper',key:key(n),N:ss.length===1?Math.max(0,ss[0].parameters.length-(n.arguments.length-1)):null,target:target.getText()});
   // Underlying implementation is separate provenance, not the wrapper's N.
  }else{
   const d=checker.getResolvedSignature(n)?.declaration;
   if(d?.body){
    if(ts.isBlock(d.body)){
     function returns(x){if(x!==d.body&&ts.isFunctionLike(x))return;if(ts.isReturnStatement(x)&&x.expression)edge(n,x.expression);else ts.forEachChild(x,returns);}returns(d.body);
    }else edge(n,d.body);
   }else if(ts.isIdentifier(n.expression)&&n.expression.text==='maybeBind'&&n.arguments[1]){
    const ss=checker.getSignaturesOfType(checker.getNonNullableType(checker.getTypeAtLocation(n.arguments[1])),ts.SignatureKind.Call);
    terminals.set(i,{kind:'bound-wrapper',key:key(n),N:ss.length===1?ss[0].parameters.length:null,target:n.arguments[1].getText()});
   }else why(n,'call return frontier '+key(n));
  }
 }else if(n.kind===ts.SyntaxKind.NullKeyword||ts.isVoidExpression(n)){}
 else if(ts.isArrowFunction(n)||ts.isFunctionExpression(n)||ts.isMethodDeclaration(n))terminals.set(i,{kind:'function',key:key(n)});
 else why(n,'unhandled '+ts.SyntaxKind[n.kind]+' '+key(n));
}
function closure(start){const seen=new Set(),stack=[id(start)],found=[],reasons=[];
 while(stack.length){const i=stack.pop();if(i<0||seen.has(i))continue;seen.add(i);if(terminals.has(i))found.push(terminals.get(i));for(const s of unknown.get(i)||[])reasons.push(s);for(const to of edges[i])stack.push(to);}
 return {values:found.sort((a,b)=>a.key.localeCompare(b.key)),frontiers:[...new Set(reasons)].sort(),nodes:seen.size};
}
const output=sites.map(s=>{const expr=s.kind==='bind'?s.node.expression.expression:s.node.arguments[1];return {...loc(s.node),kind:s.kind,member:s.member,...closure(expr)};});
const fi=new Map(functions.map(f=>[f.key,{key:f.key,name:f.name,N:f.N,signature:f.signature,rest:f.rest}]));
for(const r of output)r.values=r.values.map(v=>v.kind==='function'?{...v,...fi.get(v.key)}:v);
fs.writeFileSync(path.join(out,'member-values-upper-bound.json'),JSON.stringify({method:'flow-insensitive; field-name join across all receivers; not an exact reach set',sites:output},null,2)+'\n');
console.log('MEMBER VALUE UPPER BOUNDS');
for(const r of output)console.log(JSON.stringify({...r,frontiers:r.frontiers.length}));
