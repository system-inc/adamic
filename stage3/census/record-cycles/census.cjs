#!/usr/bin/env node
'use strict';
// Usage: node --stack-size=16384 census.cjs ADAPTED_TREE OUTPUT_JSON
// This is a source-checker census, not an execution of Adamic's lowering.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require('../../api/node_modules/typescript');
if (ts.version !== '6.0.3') throw Error(`Expected TypeScript 6.0.3, got ${ts.version}`);
const root = path.resolve(process.argv[2]);
const out = process.argv[3];
const rel = f => f.startsWith(root + path.sep) ? path.relative(root,f).replaceAll('\\','/') : f.includes('/node_modules/') ? 'external/node_modules/' + f.split('/node_modules/').pop() : 'external/' + path.basename(f);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
if (config.error) throw Error(ts.flattenDiagnosticMessageText(config.error.messageText, '\n'));
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath), undefined, configPath);
if (parsed.errors.length) throw Error(JSON.stringify(parsed.errors));
const entry = path.join(root, 'src/tsc/tsc.ts');
const program = ts.createProgram({rootNames: [entry], options: parsed.options});
const c = program.getTypeChecker();
const position = n => {
  const sf = n.getSourceFile(), p = sf.getLineAndCharacterOfPosition(n.getStart(sf));
  return `${rel(sf.fileName)}:${p.line + 1}:${p.character + 1}`;
};
const runtime = s => (ts.isImportDeclaration(s) && !s.importClause?.isTypeOnly) ||
  (ts.isExportDeclaration(s) && !s.isTypeOnly);
const ordered = [], visiting = new Set();
function visitModule(sf) {
  if (visiting.has(sf)) return;
  visiting.add(sf);
  for (const s of sf.statements) if (runtime(s) && s.moduleSpecifier) {
    const symbol = c.getSymbolAtLocation(s.moduleSpecifier);
    const target = symbol?.declarations?.find(ts.isSourceFile);
    if (target && !target.isDeclarationFile) visitModule(target);
  }
  ordered.push(sf);
}
visitModule(program.getSourceFile(entry));
const evaluationOrder = Object.fromEntries(ordered.map((s,i) => [rel(s.fileName), i]));
const F = ts.TypeFlags, O = ts.ObjectFlags, S = ts.SymbolFlags;
const types = [], ids = new Map(), locations = new Map(), declarations = new Map();
const sourceRoots = new Set();
const literals = [], recordValues = new Map(), recordKinds = new Map(), templates = new Set();
const recordUncertainty = new Map();
const functionCache = new Map(), fieldCache = new Map();
function isFunction(t) {
  if (!functionCache.has(t)) functionCache.set(t, !!(c.getSignaturesOfType(t,0).length || c.getSignaturesOfType(t,1).length || t.symbol?.name === 'Function'));
  return functionCache.get(t);
}
function args(t) { return t.objectFlags & O.Reference ? c.getTypeArguments(t) : []; }
function template(t) { return !!t.aliasTypeArguments?.some(a=>a.flags & (F.Any | F.TypeParameter)) || !!(t.flags & (F.Any | F.TypeParameter)) || !!(t.flags & F.Object && args(t).some(a => a.flags & (F.Any | F.TypeParameter))); }
function weak(t) { return t.aliasSymbol?.name === 'Weak'; }
function collection(t) {
  const name = t.symbol?.name;
  return c.isArrayType(t) || c.isTupleType(t) || ['Map','ReadonlyMap','Set','ReadonlySet','Promise'].includes(name) && t.symbol?.declarations?.some(d=>d.getSourceFile().isDeclarationFile);
}
function fields(t) {
  if (!fieldCache.has(t)) fieldCache.set(t, c.getPropertiesOfType(t).filter(p => {
    if (p.flags & (S.GetAccessor | S.SetAccessor)) return false;
    return !(p.flags & S.Method) || p.declarations?.some(d => ts.isMethodDeclaration(d) && ts.isObjectLiteralExpression(d.parent));
  }));
  return fieldCache.get(t);
}
function record(t) {
  if (recordValues.has(t)) return recordValues.get(t);
  let value;
  recordValues.set(t,undefined); // Guard recursive utility aliases.
  if (t.flags & F.Object && !c.isArrayType(t) && !c.isTupleType(t) && !(t.objectFlags & O.Class) && !(t.objectFlags & O.Reference && t.target?.objectFlags & O.Class) && !isFunction(t)) {
    const infos = c.getIndexInfosOfType(t);
    if (infos.length === 1 && infos[0].keyType.flags === F.String && !infos[0].isReadonly &&
        !infos[0].declaration?.getSourceFile().fileName.includes('.regexp.')) {
      value = infos[0].type;
      recordKinds.set(t, 'mutable string index');
      const props = c.getPropertiesOfType(t);
      if (value.isUnion() && props.length && props.every(p => !(p.flags & S.Optional) &&
          c.getTypeOfSymbolAtLocation(p,p.valueDeclaration||p.declarations?.[0]||ordered[0]).flags & (F.NumberLike|F.StringLike|F.BooleanLike)))
        recordUncertainty.set(t, 'enumerationRecordStorageType needs the lowerer enumeration provenance proof');
    }
    if (!value && infos.length) {
      value = infos.length === 1 ? infos[0].type : c.getUnionType(infos.map(i=>i.type));
      recordUncertainty.set(t, 'index signature is outside recordElement pure mutable string-index rule');
    }
    if (!value && (t.aliasSymbol?.name === 'Record' || (t.declaration || t.target?.declaration)?.parent?.name?.text === 'Record' || ['Partial','Pick'].includes(t.aliasSymbol?.name) && t.aliasTypeArguments?.[0] && record(t.aliasTypeArguments[0]))) {
      const properties = c.getPropertiesOfType(t);
      if (properties.length) {
        value = c.getUnionType(properties.map(p=>c.getTypeOfSymbolAtLocation(p,p.valueDeclaration||p.declarations?.[0]||ordered[0])));
        recordUncertainty.set(t, 'finite mapped record requires Go utility-alias/storage proof');
      }
    }
  }
  recordValues.set(t, value);
  return value;
}
function use(t, node) {
  if (!t) return;
  if (node && !locations.has(t)) locations.set(t, position(node));
  if (ids.has(t)) return;
  ids.set(t, types.length); types.push(t);
}
function scan(n) {
  // Names of declarations and type nodes are included for the inventory, but
  // generic templates never gain outgoing graph edges.
  if (ts.isExpressionNode(n) || ts.isTypeNode(n) || ts.isIdentifier(n)) {
    const t=c.getTypeAtLocation(n);use(t,n);
    if (!ts.isPartOfTypeNode(n) && (ts.isExpressionNode(n) || ts.isIdentifier(n))) sourceRoots.add(t);
  }
  if ((ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n)) && n.name) {
    const t = c.getTypeAtLocation(n.name); use(t,n);
    if (!declarations.has(t)) declarations.set(t, []);
    declarations.get(t).push({name:n.name.text, declaration:position(n)});
  }
  if (ts.isObjectLiteralExpression(n)) {
    const actual = c.getTypeAtLocation(n), contextual = c.getContextualType(n);
    use(actual,n); use(contextual,n);sourceRoots.add(actual);if(contextual)sourceRoots.add(contextual);
    literals.push({node:n,actual,contextual});
  }
  ts.forEachChild(n,scan);
}
for (const sf of ordered) scan(sf);
// Expand only ownership-bearing fields. No signature return/parameter edges.
const edges = [], uncertain = new Map();
for (let i=0;i<types.length;i++) {
  const t=types[i], links=[];
  if (template(t)) templates.add(t);
  else if (!weak(t)) {
    if (t.flags & (F.Union|F.Intersection)) links.push(...t.types);
    else if (t.flags & F.Object) {
      if (isFunction(t)) uncertain.set(t, 'closure environments / captured cells are lowering identities');
      else if (record(t)) links.push(record(t));
      else if (collection(t)) links.push(...args(t));
      else if (['MapIterator','SetIterator'].includes(t.symbol?.name)) uncertain.set(t, 'iterator captures are lowering identities');
      else {
        if (t.objectFlags & (O.Interface|O.Reference)) {
          const bases = c.getBaseTypes(t) || [];
          if (bases.some(b=>c.isArrayType(b))) uncertain.set(t,'array interface extension needs Go phantom/unbranded layout proof');
        }
        for (const p of fields(t)) links.push(c.getTypeOfSymbolAtLocation(p,p.valueDeclaration || p.declarations?.[0] || ordered[0]));
        if (c.getPropertiesOfType(t).some(p=>p.flags & (S.GetAccessor|S.SetAccessor))) uncertain.set(t,'accessor environments are lowering identities');
      }
    }
  }
  edges[i]=[];
  for (const target of links) {use(target); edges[i].push(ids.get(target));}
}
console.error(JSON.stringify({phase:'expanded',types:types.length,modules:ordered.length,records:types.filter(t=>record(t)).length}));
// Structural compatibility is symmetric in Go's related relation. Its
// candidate universe depends on lowered instantiations and literal interning.
// Keep it as a separately reported source approximation, never as a proof.
const activeTypes=new Set(),activeQueue=[...sourceRoots].map(t=>ids.get(t));
while(activeQueue.length) {const i=activeQueue.pop();if(activeTypes.has(i))continue;activeTypes.add(i);activeQueue.push(...edges[i]);}
const candidate = types.filter(t => activeTypes.has(ids.get(t)) && t.flags & F.Object && !template(t) && !weak(t) && !isFunction(t));
const viewEdges = edges.map(e=>[...e]);
let relations=0, relationChecks=0, processedGroups=0;
// related is symmetric. A spanning forest of each compatibility component
// preserves all reachability and SCC membership, avoiding a dense edge matrix.
const parents=candidate.map((_,i)=>i);
function find(i) { while(parents[i]!==i) {parents[i]=parents[parents[i]];i=parents[i];}return i; }
// Empty structural views first, because they often connect many candidates.
candidate.sort((a,b)=>c.getPropertiesOfType(a).length-c.getPropertiesOfType(b).length || ids.get(a)-ids.get(b));
for (let i=0;i<candidate.length;i++) {
  const a=candidate[i], ai=ids.get(a);processedGroups++;
  for (let j=0;j<i;j++) {
    if(find(i)===find(j)) continue;
    const b=candidate[j];relationChecks++;
    if (c.isTypeAssignableTo(a,b) || c.isTypeAssignableTo(b,a)) {
      const bi=ids.get(b); viewEdges[ai].push(bi); viewEdges[bi].push(ai); relations++;
      parents[find(i)]=find(j);processedGroups--;
      if(processedGroups===1) break;
    }
  }
  if (i % 1000 === 0) console.error(JSON.stringify({phase:'views',done:i,total:candidate.length,relations,relationChecks}));
}
function components(graph) {
  let next=0; const index=[],low=[],stack=[],active=new Set(),result=[],membership=[];
  function visit(v) {
    index[v]=low[v]=++next; stack.push(v); active.add(v);
    for(const w of graph[v]) {
      if(!index[w]) {visit(w);low[v]=Math.min(low[v],low[w]);}
      else if(active.has(w)) low[v]=Math.min(low[v],index[w]);
    }
    if(low[v]===index[v]) {
      const comp=[];let w;
      do {w=stack.pop();active.delete(w);membership[w]=result.length;comp.push(w);}while(w!==v);
      result.push(comp);
    }
  }
  for(let v=0;v<graph.length;v++) if(!index[v]) visit(v);
  return {result,membership,cyclic:v=>result[membership[v]].length>1 || graph[v].includes(v)};
}
const certainEdges=edges.map((e,i)=>recordUncertainty.has(types[i])?[]:e);
const core=components(certainEdges), hypothetical=components(edges), views=components(viewEdges);
const name = t => c.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation).split(root).join('$ADAPTED_TREE');
function cycle(v,graph,scc) {
  const queue=[[v]], seen=new Set([v]);
  for(let q=0;q<queue.length;q++) {
    const p=queue[q],last=p[p.length-1];
    for(const w of graph[last]) {
      if(w===v) return [...p,w].map(i=>({type:name(types[i]),at:locations.get(types[i]) || null}));
      if(scc.membership[w]===scc.membership[v] && !seen.has(w)) {seen.add(w);queue.push([...p,w]);}
    }
  }
  return null;
}
function reachableUncertainty(v) {
  const seen=new Set(),queue=[v], reasons=new Set();
  while(queue.length) {
    const n=queue.pop();if(seen.has(n))continue;seen.add(n);
    if(uncertain.has(types[n])) reasons.add(uncertain.get(types[n]));
    if(templates.has(types[n])) reasons.add('reachable generic/any template requires lowering instantiation or representation proof');
    if(recordUncertainty.has(types[n])) reasons.add(recordUncertainty.get(types[n]));
    queue.push(...edges[n]);
  }
  return [...reasons].sort();
}
const recordTypes=types.filter(t=>record(t));
const sitesByType=new Map(recordTypes.map(t=>[t,[]]));
const literalInventory=[];
for(const {node,actual,contextual} of literals) {
  const targets=new Map();
  const add=(t,reason)=>{
    if(!t)return;
    t=c.getNonNullableType(t);
    if(record(t)) targets.set(t,reason);
    else if(t.isUnion()) for(const m of t.types) if(record(m)) targets.set(m,reason+' union member (storage unverified)');
  };
  add(contextual,'contextual');
  if (!targets.size) add(actual,'inferred');
  // An assertion is a slot type for the lowerer even when getContextualType is absent.
  let p=node.parent;
  while(ts.isParenthesizedExpression(p))p=p.parent;
  if(ts.isAsExpression(p)||ts.isTypeAssertionExpression(p))add(c.getTypeAtLocation(p),'asserted');
  const data={at:position(node),file:rel(node.getSourceFile().fileName),evaluationOrder:evaluationOrder[rel(node.getSourceFile().fileName)],
    empty:node.properties.length===0,computed:node.properties.some(p=>p.name&&ts.isComputedPropertyName(p.name)),
    spread:node.properties.some(ts.isSpreadAssignment),unsupportedMembers:node.properties.some(p=>!ts.isPropertyAssignment(p)&&!ts.isShorthandPropertyAssignment(p)&&!ts.isSpreadAssignment(p))};
  literalInventory.push({...data,targets:[...targets].map(([t,reason])=>({type:name(t),reason}))});
  for(const [t,reason] of targets) {
    if(!sitesByType.has(t))throw Error(`Literal target absent from graph: ${name(t)}`);
    sitesByType.get(t).push({...data,reason,attributionUnsure:reason.includes('union member')});
  }
}
const rows=recordTypes.map(t=>{
  const id=ids.get(t),isTemplate=template(t),direct=core.cyclic(id),approx=views.cyclic(id);
  const reasons=reachableUncertainty(id);
  if(hypothetical.cyclic(id) && !direct) reasons.push('ownership cycle crosses a record requiring a storage proof');
  if(approx&&!direct)reasons.push('cycle uses source structural-view candidate approximation');
  if(isTemplate)reasons.push('generic template, no concrete allocation identity');
  if(!activeTypes.has(id)) reasons.push('inventory-only type, not reached from reproduced source uses');
  if(recordUncertainty.has(t)) reasons.push(recordUncertainty.get(t));
  const sites=sitesByType.get(t);
  const originSymbol=t.aliasSymbol||t.symbol;
  if(originSymbol?.declarations?.length && originSymbol.declarations.every(d=>d.getSourceFile().isDeclarationFile))
    reasons.push('external declaration/host type, lowering may replace its storage');
  const symbol=t.aliasSymbol||t.symbol;
  const decls=declarations.get(t) || symbol?.declarations?.map(d=>({name:symbol.name,declaration:position(d)})) || [];
  return {id,name:name(t),declarations:decls,firstUse:locations.get(t)||null,valueType:name(record(t)),
    status:isTemplate || !activeTypes.has(id) || recordUncertainty.has(t) || reasons.some(r=>r.startsWith('external'))?'unsure':direct?'cycle':reasons.length?'unsure':'acyclic',
    ownershipCycle:direct,sourceViewCycle:approx,sourceUse:activeTypes.has(id),unsureReasons:reasons,
    cycle:direct?cycle(id,certainEdges,core):approx?cycle(id,viewEdges,views):null,
    literalCount:sites.length,unambiguousLiteralCount:sites.filter(s=>!s.attributionUnsure).length,unsureLiteralCount:sites.filter(s=>s.attributionUnsure).length,
    files:[...new Set(sites.map(s=>s.file))].map(file=>({file,evaluationOrder:evaluationOrder[file],literalCount:sites.filter(s=>s.file===file).length})).sort((a,b)=>b.literalCount-a.literalCount||a.file.localeCompare(b.file)),sites};
}).sort((a,b)=>a.name.localeCompare(b.name)||a.id-b.id);
function fileCounts(selected) {
  const files=new Map();
  for(const r of selected) for(const s of r.sites) {
    if(!files.has(s.file))files.set(s.file,new Set());files.get(s.file).add(s.at);
  }
  return [...files].map(([file,sites])=>({file,evaluationOrder:evaluationOrder[file],literalCount:sites.size})).sort((a,b)=>b.literalCount-a.literalCount||a.file.localeCompare(b.file));
}
const totals={};
for(const status of ['cycle','acyclic','unsure']) {
  const selected=rows.filter(r=>r.status===status);
  totals[status]={recordTypes:selected.length,unambiguousLiteralTypePairs:selected.reduce((n,r)=>n+r.unambiguousLiteralCount,0),unsureLiteralTypePairs:selected.reduce((n,r)=>n+r.unsureLiteralCount,0),literalTypePairs:selected.reduce((n,r)=>n+r.literalCount,0),uniqueLiteralSites:new Set(selected.flatMap(r=>r.sites.map(s=>s.at))).size};
}
const result={schema:1,typescript:ts.version,sourcePin:'050880ce59e30b356b686bd3144efe24f875ebc8',
  adamicBase:'fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d',
  scriptSha256:crypto.createHash('sha256').update(fs.readFileSync(__filename)).digest('hex'),
  runtimeSourceSha256:crypto.createHash('sha256').update(ordered.map(sf=>rel(sf.fileName)+'\0'+sf.text+'\0').join('')).digest('hex'),
  options:JSON.parse(JSON.stringify(parsed.options).split(root).join('$ADAPTED_TREE')),rootNames:['src/tsc/tsc.ts'],sourceFiles:program.getSourceFiles().length,runtimeModules:ordered.length,
  graph:{nodes:types.length,ownershipEdges:edges.reduce((n,e)=>n+e.length,0),sourceViewSpanningRelations:relations,relationChecks,candidates:candidate.length,identityViewEdgesExcluded:true},
  diagnostics:ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,at:d.file&&d.start!==undefined?`${rel(d.file.fileName)}:${d.file.getLineAndCharacterOfPosition(d.start).line+1}`:null,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')})),
  ownershipControl:{recordTypes:rows.filter(r=>!r.ownershipCycle && !templates.has(types[r.id])).length,uniqueLiteralSites:new Set(rows.filter(r=>!r.ownershipCycle && !templates.has(types[r.id])).flatMap(r=>r.sites.map(s=>s.at))).size},
  componentRuleCounts:{sourceViewCycleRecordTypes:rows.filter(r=>r.sourceUse && r.sourceViewCycle).length,sourceViewCycleUniqueLiteralSites:new Set(rows.filter(r=>r.sourceUse && r.sourceViewCycle).flatMap(r=>r.sites.map(s=>s.at))).size,includingReflexiveRelatedEdgesRecordTypes:rows.filter(r=>r.sourceUse && !templates.has(types[r.id])).length},
  totals,files:{cycle:fileCounts(rows.filter(r=>r.status==='cycle')),unsure:fileCounts(rows.filter(r=>r.status==='unsure')),acyclic:fileCounts(rows.filter(r=>r.status==='acyclic'))},
  records:rows,evaluationOrder,literalInventory};
fs.writeFileSync(out,JSON.stringify(result,null,2).split(root).join('$ADAPTED_TREE').split(path.resolve(__dirname,'../../api/node_modules')).join('$API_NODE_MODULES')+'\n');
console.error(JSON.stringify({phase:'done',totals}));
