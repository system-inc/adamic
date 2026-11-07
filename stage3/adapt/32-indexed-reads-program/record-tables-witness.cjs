'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),cp=require('node:child_process'),assert=require('node:assert/strict');
const s='/workspace/scratch/gov-b';
function source(side,name){return fs.readFileSync(s+'/'+side+'/src/compiler/'+name,'utf8').replaceAll('[(globalThis as any).__govId("type", type.id, type)]','[type.id]').replaceAll('[(globalThis as any).__govId("symbol", symbolId, symbol)]','[symbolId]').replace('return (globalThis as any).__govValues(collection, values);','return values;');}
function find(root,p){let a=[];function visit(n){if(p(n))a.push(n);ts.forEachChild(n,visit);}visit(root);return a;}
function generate(side,body){
 const coreText=source(side,'core.ts'),walkText=source(side,'symbolWalker.ts'),core=ts.createSourceFile('core.ts',coreText,ts.ScriptTarget.Latest,true),walk=ts.createSourceFile('symbolWalker.ts',walkText,ts.ScriptTarget.Latest,true);
 const fn=(root,name)=>{const a=find(root,n=>ts.isFunctionDeclaration(n)&&n.name?.text===name&&n.body);assert.equal(a.length,1);return a[0];};
 const declarations=['visitedTypes','visitedSymbols'].map(name=>{const a=find(walk,n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(walk)===name));assert.equal(a.length,1);return a[0].getText(walk);});
 const type=fn(walk,'visitType'),symbol=fn(walk,'visitSymbol');
 const finalies=find(walk,ts.isTryStatement).filter(n=>n.finallyBlock).map(n=>n.finallyBlock.getText(walk));assert.equal(finalies.length,2);assert.equal(finalies[0],finalies[1]);
 const prelude=`export interface MapLike<T>{[key:string]:T;}\ninterface Type{readonly id:number;name:string;}\ninterface Symbol{readonly id:number;name:string;}\ninterface Pair{visitedTypes:Type[];visitedSymbols:Symbol[];}\nconst hasOwnProperty=Object.prototype.hasOwnProperty;\nfunction getSymbolId(symbol:Symbol):number{return symbol.id;}\n`;
 const factory=`function makeWalker(){${declarations.join('\n')}\nfunction visitType(type:Type|undefined):undefined{${type.body.statements.slice(0,3).map(n=>n.getText(walk)).join('\n')}}\nfunction visitSymbol(symbol:Symbol|undefined):boolean{${symbol.body.statements.slice(0,4).map(n=>n.getText(walk)).join('\n')}return false;}\nreturn (types:readonly Type[],symbols:readonly Symbol[]):Pair=>{try{for(const type of types)visitType(type);for(const symbol of symbols)visitSymbol(symbol);return {visitedTypes:getOwnValues(visitedTypes),visitedSymbols:getOwnValues(visitedSymbols)};}finally${finalies[0]}};}\n`;
 return prelude+(side==='before'?fn(core,'clear').getText(core)+'\n':'')+fn(core,'getOwnValues').getText(core)+'\n'+factory+`function show(pair:Pair){console.log('types='+String(pair.visitedTypes.length));for(const item of pair.visitedTypes)console.log(String(item.id)+':'+item.name);console.log('symbols='+String(pair.visitedSymbols.length));for(const item of pair.visitedSymbols)console.log(String(item.id)+':'+item.name);}\nconst walk=makeWalker();\n`+body+`\nshow(walk([],[]));\n`;
}
const cases={sparse:`show(walk([{id:5,name:'type-x'}],[{id:5,name:'symbol-x'}]));`,outOfOrder:`show(walk([{id:4468,name:'Tlarge'},{id:2,name:'Ttwo'},{id:10,name:'Tten'}],[{id:4468,name:'Slarge'},{id:2,name:'Stwo'},{id:10,name:'Sten'}]));`,empty:`show(walk([],[]));`};
const dir=s+'/witnesses';fs.mkdirSync(dir,{recursive:true});const result={};
for(const [name,body] of Object.entries(cases)){
 const rows={};for(const side of ['before','after']){const text=generate(side,body),stem=path.join(dir,name+'-'+side);fs.writeFileSync(stem+'.ts',text);fs.writeFileSync(stem+'.mjs',ts.transpileModule(text,{compilerOptions:{module:ts.ModuleKind.ESNext,target:ts.ScriptTarget.ES2024}}).outputText);const r=cp.spawnSync(process.execPath,[stem+'.mjs'],{encoding:'utf8'});rows[side]={exit:r.status,stdout:r.stdout,stderr:r.stderr};fs.writeFileSync(stem+'.node.stdout',r.stdout);fs.writeFileSync(stem+'.node.stderr',r.stderr);}
 assert.deepEqual(rows.before,rows.after);assert.equal(rows.before.exit,0);result[name]=rows;
}
fs.writeFileSync(s+'/raw/node-witnesses.json',JSON.stringify({exactExtractedTableOperations:true,sharedBindingIdentity:true,reuseAfterFinallyTested:true,cases:result},null,2)+'\n');
