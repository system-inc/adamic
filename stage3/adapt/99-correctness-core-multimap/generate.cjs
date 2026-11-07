'use strict';
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),assert=require('node:assert/strict'),cp=require('node:child_process');
const repo='/workspace/adamic',tree='/workspace/multimap-base',out=repo+'/stage3/adapt/99-correctness-core-multimap',early=repo+'/stage3/adapt/94-correctness-multimap';
const hash=s=>crypto.createHash('sha256').update(s).digest('hex'),parse=(name,text)=>ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true),norm=s=>s.replace(/\r\n/g,'\n');
const raw=norm(fs.readFileSync(tree+'/src/compiler/core.ts','utf8')),core=parse('core.ts',raw),get=n=>core.statements.find(x=>x.name?.text===n);
const add=get('multiMapAdd').getText(core),remove=get('multiMapRemove').getText(core);
const free=(s,n,ret)=>'/** @internal */\n'+s.replace('function '+n,'export function '+n+'Value').replace('this: MultiMap<K, V>','map: Map<K, V[]>').replace(/\bthis\b/g,'map').replace('value: V) {','value: V): '+ret+' {');
const helpers=free(add,'multiMapAdd','V[]')+'\n\n'+free(remove,'multiMapRemove','void');
fs.mkdirSync(early,{recursive:true});const next=get('createQueue'),anchor=raw.slice(get('multiMapRemove').getStart(core),next.getStart(core));const afterAnchor=anchor.replace(remove,remove+'\n\n'+helpers);
fs.writeFileSync(early+'/pin.json',JSON.stringify({before:anchor,after:afterAnchor,beforeSha256:hash(anchor),afterSha256:hash(afterAnchor),helpers},null,2)+'\n');
// Derive call edits directly from checker-resolved census locations.
const census=JSON.parse(fs.readFileSync(out+'/census.json','utf8')),byFile=new Map();
for(const [member,rows]of Object.entries(census.members))for(const row of rows.filter(r=>r.kind==='call')){if(row.file==='src/compiler/program.ts')continue;if(!byFile.has(row.file))byFile.set(row.file,[]);byFile.get(row.file).push({...row,member});}
const pins=[];const ownerPath=n=>{let names=[];for(let p=n;p;p=p.parent)if((ts.isFunctionLike(p)||ts.isClassDeclaration(p))&&p.name)names.unshift(p.name.getText(p.getSourceFile()));return names;};
function owner(n){for(let p=n;p;p=p.parent)if((ts.isFunctionDeclaration(p)||ts.isMethodDeclaration(p))&&p.body&&p.name)return p;throw Error('no named owner');}
for(const [file,rows]of byFile){const text=norm(fs.readFileSync(path.join(tree,file),'utf8')),sf=parse(file,text),owners=new Map(),names=new Set();
 const imports=sf.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text.endsWith('/ts.js'));const namespace=imports.find(n=>n.importClause?.namedBindings&&ts.isNamespaceImport(n.importClause.namedBindings))?.importClause.namedBindings.name.text;
 for(const r of rows){const pos=sf.getPositionOfLineAndCharacter(r.line-1,r.column-1),token=ts.getTokenAtPosition(sf,pos),access=token.parent,call=access.parent;assert(ts.isCallExpression(call)&&call.expression===access&&ts.isPropertyAccessExpression(access)&&access.name.text===r.member);assert(!call.questionDotToken&&!access.questionDotToken&&!call.typeArguments&&call.arguments.length===2);const fn=owner(call),key=fn.getStart(sf);if(!owners.has(key))owners.set(key,{node:fn,edits:[]});const helper='multiMap'+(r.member==='add'?'Add':'Remove')+'Value';names.add(helper);const args=text.slice(access.end,call.end);assert(args.startsWith('('));owners.get(key).edits.push({start:call.getStart(sf)-key,end:call.end-key,after:(namespace?namespace+'.':'')+helper+'('+access.expression.getText(sf)+', '+args.slice(1)});}
 for(const {node,edits}of owners.values()){const before=node.getText(sf);let after=before;for(const e of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,e.start)+e.after+after.slice(e.end);pins.push({file,kind:'owner',path:ownerPath(node),before,after});}
 if(file!=='src/compiler/core.ts'&&!namespace){assert.equal(imports.length,1,file+' ts import');const imp=imports[0];assert(ts.isNamedImports(imp.importClause.namedBindings));const before=imp.getText(sf),at=imp.importClause.namedBindings.elements[0].getStart(sf)-imp.getStart(sf);const after=before.slice(0,at)+[...names].sort().map(n=>n+',\n    ').join('')+before.slice(at);pins.push({file,kind:'import',module:imp.moduleSpecifier.text,before,after});}
}
pins.push({file:'src/compiler/core.ts',kind:'declaration',path:['MultiMap'],before:get('MultiMap').getText(core),after:'export type MultiMap<K, V> = Map<K, V[]>;'});
pins.push({file:'src/compiler/core.ts',kind:'owner',path:['createMultiMap'],before:get('createMultiMap').getText(core),after:'export function createMultiMap<K, V>(): MultiMap<K, V> {\n    return new Map<K, V[]>();\n}'});
for(const name of ['multiMapAdd','multiMapRemove'])pins.push({file:'src/compiler/core.ts',kind:'owner',path:[name],before:get(name).getText(core),after:''});
for(const p of pins){p.beforeSha256=hash(p.before);p.afterSha256=hash(p.after);}
fs.writeFileSync(out+'/owners.json',JSON.stringify(pins,null,2)+'\n');
const allfiles=ts.sys.readDirectory(repo+'/stage3/adapt',['.cjs','.json']).filter(p=>/adapt\.cjs$|owners[^/]*\.json$|sites[^/]*\.json$/.test(p)&&!p.startsWith(out)&&!p.startsWith(early));const audit=[];
for(const pin of pins.filter(p=>p.kind==='owner'))for(const file of allfiles){const s=fs.readFileSync(file,'utf8');if(s.includes(pin.path.at(-1)))audit.push({owner:pin.path.join('.'),file:path.relative(repo,file)});}
fs.writeFileSync(out+'/owner-audit.json',JSON.stringify(audit,null,2)+'\n');console.log(JSON.stringify({pins:pins.length,files:new Set(pins.map(p=>p.file)).size,auditHits:audit,helpers},null,2));
