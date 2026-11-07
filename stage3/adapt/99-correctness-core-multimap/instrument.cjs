'use strict';
const ts=require('typescript'),fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const tree=process.argv[2],out=process.argv[3],coreName=path.join(tree,'src/compiler/core.ts'),coreText=fs.readFileSync(coreName,'utf8'),core=ts.createSourceFile(coreName,coreText,ts.ScriptTarget.Latest,true);
const get=name=>{const matches=core.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert.equal(matches.length,1,name);return matches[0];};
const emit=text=>ts.transpileModule(text.replace(/^export /,''),{compilerOptions:{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.None,removeComments:true}}).outputText.replace(/^"use strict";\s*/,'').replace(/export /g,'');
const functions={};for(const name of ['multiMapAdd','multiMapRemove','unorderedRemoveItem','unorderedRemoveItemAt','unorderedRemoveFirstItemWhere'])functions[name]=emit(get(name).getText(core));
const free=(name,ret)=>get(name).getText(core).replace('function '+name,'function '+name+'Value').replace('this: MultiMap<K, V>','map: Map<K, V[]>').replace(/\bthis\b/g,'map').replace('value: V) {','value: V): '+ret+' {');
const bodies={dependencies:['unorderedRemoveItemAt','unorderedRemoveFirstItemWhere','unorderedRemoveItem'].map(n=>functions[n]).join('\n'),add:functions.multiMapAdd,remove:functions.multiMapRemove,freeAdd:emit(free('multiMapAdd','V[]')),freeRemove:emit(free('multiMapRemove','void'))};fs.writeFileSync(out,JSON.stringify(bodies,null,2));
let enumerationSites=0,files=0;
for(const name of ts.sys.readDirectory(path.join(tree,'src'),['.ts'])){
 const text=fs.readFileSync(name,'utf8'),sf=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true),edits=[];
 const wrap=n=>{edits.push({at:n.getStart(sf),text:'((globalThis as any).__mmEnumeration as <T>(value: T) => T)('},{at:n.end,text:')'});enumerationSites++;};
 const walk=n=>{if(ts.isForInStatement(n))wrap(n.expression);if(ts.isSpreadAssignment(n))wrap(n.expression);if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&ts.isIdentifier(n.expression.expression)&&['Object','Reflect'].includes(n.expression.expression.text)&&['keys','getOwnPropertyNames','getOwnPropertyDescriptors','ownKeys','entries','values','assign'].includes(n.expression.name.text))for(const arg of n.arguments)wrap(arg);ts.forEachChild(n,walk);};/* Enumeration probes are inserted after emit to preserve checker refinements. */
 if(name===coreName){for(const op of ['Add','Remove']){const fn=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='multiMap'+op);edits.push({at:fn.body.getStart(sf)+1,text:' (globalThis as any).__mmCheck("'+op.toLowerCase()+'", this, key, value);'});}const fn=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createMultiMap');const ret=fn.body.statements.find(ts.isReturnStatement);edits.push({at:ret.getStart(sf),text:'(globalThis as any).__mmRegister(map); '});}
 if(edits.length){let after=text;for(const edit of edits.sort((a,b)=>b.at-a.at))after=after.slice(0,edit.at)+edit.text+after.slice(edit.at);assert.equal(ts.createSourceFile(name,after,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0,name);fs.writeFileSync(name,after);files++;}
}
console.log(JSON.stringify({measurementOnly:true,files,enumerationSites,dependencies:Object.keys(functions)},null,2));
