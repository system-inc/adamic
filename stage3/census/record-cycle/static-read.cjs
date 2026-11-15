// Read-only TypeScript API fallback. No lowering graph or fresh-write proof is claimed.
const fs=require('fs'),path=require('path'),ts=require(process.env.API_TYPESCRIPT);
const [tree,out]=process.argv.slice(2);fs.mkdirSync(out,{recursive:true});
if(ts.version!=='6.0.3')throw Error(ts.version);
const roots=ts.sys.readDirectory(path.join(tree,'src/compiler'),['.ts']).sort();
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:['node'],typeRoots:[path.join(path.dirname(process.env.API_TYPESCRIPT),'@types')]};
const program=ts.createProgram(roots,options),c=program.getTypeChecker();
const files=roots.map(p=>program.getSourceFile(p));
const firstUse=new Map();
const types=new Map(),edges=new Map(),unknown=new Map(),records=[],writes=[],functions=[],allUnits=new Set();let visitedNodes=0;
function tid(t){if(t){types.set(t.id,t);return t.id;}return 0;}
function fmt(t){return t?c.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation):'<missing>';}
function loc(n){const f=n.getSourceFile(),a=f.getLineAndCharacterOfPosition(n.getStart(f));return path.relative(tree,f.fileName)+':'+(a.line+1)+':'+(a.character+1);}
function line(n){const f=n.getSourceFile(),a=f.getLineAndCharacterOfPosition(n.getStart(f));return f.text.split(/\r?\n/)[a.line];}
function unit(n){let top=n,fn;for(let p=n;p&&!ts.isSourceFile(p);p=p.parent){top=p;if(ts.isFunctionDeclaration(p)) {fn=p;break;}}return loc(fn||top);}
function type(n){try{return c.getTypeAtLocation(n);}catch{return undefined;}}
function args(t){try{return c.getTypeArguments(t)||[];}catch{return [];}}
function symbolName(t){return t?.target?.symbol?.name||t?.symbol?.name||'';}
function callable(t){return !!t&&(c.getSignaturesOfType(t,0).length>0||c.getSignaturesOfType(t,1).length>0||symbolName(t)==='Function');}
function template(t){return !t||!!(t.flags&(ts.TypeFlags.Any|ts.TypeFlags.TypeParameter))||args(t).some(x=>x.flags&(ts.TypeFlags.Any|ts.TypeFlags.TypeParameter));}
function primitive(t){return !!t&&!(t.flags&(ts.TypeFlags.Object|ts.TypeFlags.Union|ts.TypeFlags.Intersection|ts.TypeFlags.Any|ts.TypeFlags.Unknown|ts.TypeFlags.TypeParameter|ts.TypeFlags.Conditional|ts.TypeFlags.IndexedAccess));}
function recordElement(t,seen=new Set()){
 if(!t||seen.has(t.id)||!(t.flags&ts.TypeFlags.Object)||c.isArrayType(t)||c.isTupleType(t)||callable(t))return;
 seen.add(t.id);
 if(t.symbol?.declarations?.some(ts.isClassDeclaration))return;
 const infos=c.getIndexInfosOfType(t);
 if(infos.length===1&&infos[0].keyType.flags===ts.TypeFlags.String&&!infos[0].isReadonly)return infos[0].type;
 if(t.aliasSymbol?.name==='Pick'&&t.aliasTypeArguments?.length===2)return recordElement(t.aliasTypeArguments[0],seen);
 if(t.aliasSymbol?.name==='Partial'&&t.aliasTypeArguments?.[0]?.aliasSymbol?.name==='Record')return t.aliasTypeArguments[0].aliasTypeArguments?.[1];
}
function held(t,seen=new Set()){
 if(!t||seen.has(t.id))return false;seen.add(t.id);
 if(recordElement(t))return true;
 if(t.isUnion?.())return t.types.some(x=>held(x,seen));
 if(c.isArrayType(t)||c.isTupleType(t)||['Map','ReadonlyMap','Set','ReadonlySet'].includes(symbolName(t)))return args(t).some(x=>held(x,seen));return false;
}
function isType(n){return ts.isTypeNode(n)||ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n);}
function pushWrite(n,receiver,value,kind,name,rhs,extra={}){
 const h=type(receiver);writes.push({node:n,receiver,holder:h,value,kind,name,rhs,...extra});tid(h);tid(value);
}
function forParts(t,fn){if(t?.isUnion?.())for(const x of t.types)fn(x);else fn(t);}
for(const file of files){
 function walk(n){visitedNodes++;if(isType(n))return;
  if(ts.isFunctionDeclaration(n)||(n.parent&&ts.isSourceFile(n.parent)))allUnits.add(unit(n));
  if(ts.isFunctionLike(n)&&n.body)functions.push(n);
  if(ts.isExpressionNode(n)||ts.isIdentifier(n)){{const t=type(n);tid(t);if(t&&!firstUse.has(t.id))firstUse.set(t.id,n);}}
  if(ts.isObjectLiteralExpression(n)){
   const own=type(n),context=c.getContextualType(n),present=context&&c.getNonNullableType(context);
   let rec=recordElement(present)?present:own;
   const elem=recordElement(rec);
   if(elem){records.push({node:n,record:rec,element:elem});tid(rec);tid(elem);}
   else if(present?.isUnion?.()&&present.types.some(x=>recordElement(x))){records.push({node:n,record:own,element:undefined,uncertain:'contextual union record lowering/recordLiteralType requires native checker'});tid(own);}
  }
  if(ts.isBinaryExpression(n)&&ts.isAssignmentOperator(n.operatorToken.kind)){
   const left=n.left;
   if(ts.isPropertyAccessExpression(left)){forParts(type(left.expression),h=>pushWrite(n,left.expression,h&&c.getTypeOfPropertyOfType(h,left.name.text),'field',left.name.text,n.right));}
   else if(ts.isElementAccessExpression(left)){
    forParts(type(left.expression),h=>{
     if(!h)return;
     const name=left.argumentExpression&&(ts.isStringLiteralLike(left.argumentExpression)||ts.isNumericLiteral(left.argumentExpression))?left.argumentExpression.text:undefined;
     const k=c.isArrayType(h)||c.isTupleType(h)?'element':recordElement(h)?'record-entry':name?'field':'dynamic-element';
     const v=k==='element'?(args(h)[0]||type(left)):k==='record-entry'?recordElement(h):name?c.getTypeOfPropertyOfType(h,name):type(left);
     pushWrite(n,left.expression,v,k,name||'',n.right);
    });
   }
  }
  if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)){
   const a=n.expression,r=a.expression,method=a.name.text;
   forParts(type(r),h=>{
    if(!h)return;const as=args(h),name=symbolName(h);
    if((c.isArrayType(h)||c.isTupleType(h))&&['push','unshift','splice','fill','copyWithin'].includes(method))pushWrite(n,r,as[0],'element','',undefined,{method});
    if(name==='Map'&&method==='set'){pushWrite(n,r,as[0],'map-key','',undefined,{method});pushWrite(n,r,as[1],'map-value','',undefined,{method});}
    if(name==='Set'&&method==='add')pushWrite(n,r,as[0],'set-element','',undefined,{method});
   });
  }
  ts.forEachChild(n,walk);
 }
 walk(file);
}
console.error('inventory',visitedNodes,'nodes',writes.length,'source writes',records.length,'record literals',types.size,'initial types');
// Derive only named type-field/container/union edges. Calls and hidden shapes are explicit uncertainty.
const queue=[...types.values()];
for(let at=0;at<queue.length;at++){
 const t=queue[at];if(edges.has(t.id))continue;
 const links=[],unc=[];edges.set(t.id,links);
 const add=(v,label)=>{if(!v)return;if(!types.has(v.id)){tid(v);queue.push(v);}links.push({to:v.id,label});};
 if(template(t)){unc.push('generic/any type; native concrete instantiation and template filter unavailable');}
 else if(t.flags&(ts.TypeFlags.Unknown|ts.TypeFlags.Conditional|ts.TypeFlags.IndexedAccess)){unc.push('unresolved checker type');}
 else if(t.isUnionOrIntersection?.()){for(const v of t.types)add(v,'union/intersection member');}
 else if(t.flags&ts.TypeFlags.Object){
  const name=symbolName(t),re=recordElement(t);
  if(callable(t)){unc.push('function capture/cell links require coherent lowered closureRecords');}
  else if(re)add(re,'record values');
  else if(c.isArrayType(t)||c.isTupleType(t)||['Map','ReadonlyMap','Set','ReadonlySet','Promise'].includes(name)){args(t).forEach((v,i)=>add(v,name+' argument '+i));}
  else {
   for(const p of c.getPropertiesOfType(t)){
    if(p.flags&(ts.SymbolFlags.GetAccessor|ts.SymbolFlags.SetAccessor))continue;
    const literalMethod=p.declarations?.some(d=>ts.isMethodDeclaration(d)&&ts.isObjectLiteralExpression(d.parent));
    if((p.flags&ts.SymbolFlags.Method)&&!literalMethod)continue;
    if(p.name==='prototype')continue;
    add(c.getTypeOfSymbol(p),'.'+p.name);
   }
   unc.push('other assignable program shapes/accessor captures may add edges');
  }
 }
 if(unc.length)unknown.set(t.id,unc);
 if(types.size>60000)throw Error('static type limit');
}
console.error('graph',types.size,'types',edges.size,'edge lists');
const relation=new Map();
function related(a,b){if(a===b)return true;if(!a||!b||!(a.flags&ts.TypeFlags.Object)||!(b.flags&ts.TypeFlags.Object)||callable(a)||callable(b)||template(a)||template(b))return false;const key=a.id<b.id?a.id+','+b.id:b.id+','+a.id;if(relation.has(key))return relation.get(key);let ok=false;try{ok=c.isTypeAssignableTo(a,b)||c.isTypeAssignableTo(b,a);}catch{}relation.set(key,ok);return ok;}
const reachMemo=new Map();
function reach(from,target){
 if(!from||!target)return {state:'unknown',reason:'missing type'};
 const key=from.id+','+target.id;if(reachMemo.has(key))return reachMemo.get(key);
 const pending=[from.id],seen=new Set(),prev=new Map(),unc=new Set();let found;
 while(pending.length){const id=pending.pop();if(seen.has(id))continue;seen.add(id);const t=types.get(id);
  if(related(t,target)){found=id;break;}
  for(const x of unknown.get(id)||[])unc.add(x);
  for(const e of edges.get(id)||[]){if(!prev.has(e.to))prev.set(e.to,{id,label:e.label});pending.push(e.to);}
 }
 let result;
 if(found!==undefined){const route=[];let id=found,guard=new Set();while(id!==from.id&&prev.has(id)&&!guard.has(id)){guard.add(id);const p=prev.get(id);route.push({from:fmt(types.get(p.id)),via:p.label,to:fmt(types.get(id))});id=p.id;}route.reverse();result={state:'yes',path:route,target:fmt(target),relation:fmt(types.get(found))};}
 else result={state:unc.size?'unknown':'no',reason:[...unc].join('; ')};
 reachMemo.set(key,result);return result;
}
// The rule examines all unproven writes sharing a field name. Without fresh.ProveWrites the precise
// function branch cannot be selected soundly; list the syntactic candidates and mark it unknown.
const byField=new Map();for(const w of writes)if(w.kind==='field'){if(!byField.has(w.name))byField.set(w.name,[]);byField.get(w.name).push(w);}
const recordTypes=[...new Map(records.map(r=>[r.record.id,r])).values()];
const rows=[];
for(let i=0;i<writes.length;i++){
 const w=writes[i],h=w.holder,v=w.value,cycle=reach(v,h);let verdict='unknown',why='',witness;
 let sourceFunction;
 const rhs=w.rhs&&ts.skipParentheses(w.rhs);
 if(rhs&&(ts.isArrowFunction(rhs)||ts.isFunctionExpression(rhs)))sourceFunction=rhs;
 else if(rhs&&ts.isIdentifier(rhs)){
  const symbol=c.getSymbolAtLocation(rhs);if(symbol?.declarations?.length===1&&ts.isFunctionDeclaration(symbol.declarations[0]))sourceFunction=symbol.declarations[0];
 }
 let precise=w.kind==='field'&&sourceFunction;
 if(primitive(v)){verdict='clean';why='scalar slot: no reference-valued reach and therefore no graph-slot seed';}
 else if(held(v)){verdict='stops';why='holdsRecord(value) true: existing direct record/container arm; conditional on being an unproven cycle seed';witness=recordTypes.find(r=>related(r.record,v)||c.isArrayType(v)&&related(args(v)[0],r.record));}
 else if(precise){verdict='unknown';why='plain function-field assignment: writtenFunctionCells must inspect every matching unproven field write and its own lowered Environment; source function signatures cannot substitute';}
 else {
  let possible=false;
  for(const r of recordTypes){const a=reach(v,r.record),b=reach(r.element,h);if(a.state==='yes'&&b.state==='yes'){witness=r;witness={...r,forward:a,back:b};verdict='stops';why='both named type reaches satisfy reachesRecord; conditional on native seed/write proof and checker agreement';break;}if(a.state!=='no'&&b.state!=='no')possible=true;}
  if(verdict!=='stops'){if(!possible){verdict='clean';why='every enumerated literal record is excluded by a scalar-complete no-reach; no unresolved route in the exclusion';}else why='no complete positive witness; omitted shape/capture/native-instantiation edges prevent a negative proof';}
 }
 if(w.kind==='record-entry'){verdict='clean';why='record-holder mutation is judged by the older record-cycle Refused arm, not noteRecordSlot; no new graph-slot judgment here';witness=undefined;}
 if(w.kind==='dynamic-element'){verdict='unknown';why='native slot/write kind cannot be determined from this dynamic element access';witness=undefined;}
 if(cycle.state==='no'){verdict='clean';why='slot value has a complete non-reference reach and cannot reach holder; not a slotsOf seed';}
 rows.push({id:i+1,unit:unit(w.node),site:loc(w.node),kind:w.kind,field:w.name,method:w.method,holder:fmt(h),value:fmt(v),slot_type_id:tid(v),holder_type_id:tid(h),classification:verdict,reason:why,seed_status:cycle.state==='yes'?'type-cycle candidate; fresh.ProveWrites unproven status UNKNOWN':cycle.state==='no'?'not a cycle seed':'UNKNOWN type-cycle and native write proof',cycle,precise_function_form:!!precise,source_function:sourceFunction?loc(sourceFunction):undefined,record:witness?{site:loc(witness.node),type:fmt(witness.record),element:fmt(witness.element),line:line(witness.node),forward:witness.forward,back:witness.back}:undefined,line:line(w.node),expression:w.node.getText().slice(0,1600)});
 if((i+1)%500===0)console.error('classified',i+1);
}
// Source-visible slotsOf domain, with unknown native seed eligibility retained.
const slots=[],slotKeys=new Set();
for(const t of types.values()){
 if(!(t.flags&ts.TypeFlags.Object)||template(t)||callable(t)||(t.objectFlags&ts.ObjectFlags.FreshLiteral))continue;
 const name=symbolName(t),rs=recordElement(t);const defs=[];
 if(rs){continue;} // graphTypes' initial record-holder check precedes graph seeding.
 if(c.isArrayType(t)||c.isTupleType(t)){
  if(name==='ReadonlyArray'||t.target?.readonly)continue;
  for(const v of args(t))defs.push(['element','',v,undefined]);
 }else if(['ReadonlyMap','ReadonlySet'].includes(name))continue;
 else if(name==='Set')defs.push(['set-element','',args(t)[0],undefined]);
 else if(name==='Map'){defs.push(['map-key','',args(t)[0],undefined],['map-value','',args(t)[1],undefined]);}
 else for(const prop of c.getPropertiesOfType(t)){
  if(prop.flags&(ts.SymbolFlags.GetAccessor|ts.SymbolFlags.SetAccessor))continue;
  if((prop.flags&ts.SymbolFlags.Method)&&!prop.declarations?.some(d=>ts.isMethodDeclaration(d)&&ts.isObjectLiteralExpression(d.parent)))continue;
  if(prop.name==='prototype')continue;
  defs.push(['field',prop.name,c.getTypeOfSymbol(prop),prop.declarations?.[0]]);
 }
 for(const [kind,field,v,decl] of defs){
  if(!v||primitive(v))continue;
  const cycle=reach(v,t);if(cycle.state==='no')continue;
  const key=[fmt(t),kind,field,fmt(v),decl?loc(decl):''].join('\t');if(slotKeys.has(key))continue;slotKeys.add(key);
  const matches=rows.filter(r=>r.kind===kind&&r.field===field&&r.holder_type_id===t.id&&r.slot_type_id===v.id);
  let classification='unknown',reason='native unproven-write/unknown-write index and structural-view/capture edges unavailable';
  if(held(v)){classification='stops';reason='holdsRecord(value) true, conditional on this being a cycle seed (existing arm)';}
  else if(matches.length&&matches.every(r=>r.classification==='stops')){classification='stops';reason='matching source writes supply static reachesRecord witnesses; native seed and exact storage remain unproven';}
  else if(matches.length&&matches.every(r=>r.classification==='clean')){classification='clean';reason='all exactly typed source writes have a static negative record judgment; unknown native writes can invalidate coverage';}
  const site=decl||firstUse.get(t.id);slots.push({id:slots.length+1,holder_type_id:t.id,holder:fmt(t),value:fmt(v),kind,field,site:site?loc(site):'<derived type; no source declaration>',line:site?line(site):'',unit:site?unit(site):'<derived>',classification,reason,cycle:cycle.state,source_write_ids:matches.map(r=>r.id),seed_eligibility:'UNKNOWN: fresh.ProveWrites and complete native write index not available'});
 }
}
fs.writeFileSync(path.join(out,'slots.json'),JSON.stringify(slots,null,2)+'\n');
console.error('source-visible slot candidates',slots.length);
const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,site:d.file?path.relative(tree,d.file.fileName)+':'+(d.file.getLineAndCharacterOfPosition(d.start).line+1):'',text:ts.flattenDiagnosticMessageText(d.messageText,' ')}));
const summary={kind:'STATIC READ, NOT LOWERING MEASUREMENT',typescript:ts.version,head:'b529db907bca6564edd2244ac8e5d102b85cf7e8',source_files:roots.length,source_units:allUnits.size,nodes:visitedNodes,record_literals:records.length,record_types:recordTypes.length,source_write_candidates:rows.length,source_visible_slot_candidates:slots.length,slot_classifications:Object.fromEntries(['stops','clean','unknown'].map(k=>[k,slots.filter(x=>x.classification===k).length])),classifications:Object.fromEntries(['stops','clean','unknown'].map(k=>[k,rows.filter(x=>x.classification===k).length])),lowering_units_with_seed_coverage:0,actual_units_newly_stopped:null,diagnostics:diagnostics.length,types:types.size,options,limits:['Syntactic source-write inventory, not an exhaustive list of production slotsOf seeds. Lowered generated/generic/indirect writes and fresh.ProveWrites unknown writes are unavailable.','Stops means the static record judgment has a witness, conditional on seed eligibility/native type agreement; it is not a production or lowering stop.','Clean negatives retain explicit capture/shape uncertainty. Missing closure and native storage/utility alias facts are not replaced by empty graphs.']};
fs.writeFileSync(path.join(out,'summary.json'),JSON.stringify(summary,null,2)+'\n');
fs.writeFileSync(path.join(out,'writes.json'),JSON.stringify(rows,null,2)+'\n');
fs.writeFileSync(path.join(out,'records.json'),JSON.stringify(records.map(r=>({site:loc(r.node),unit:unit(r.node),type:fmt(r.record),element:fmt(r.element),uncertain:r.uncertain,line:line(r.node)})),null,2)+'\n');
fs.writeFileSync(path.join(out,'diagnostics.json'),JSON.stringify(diagnostics,null,2)+'\n');
console.log(JSON.stringify(summary));
