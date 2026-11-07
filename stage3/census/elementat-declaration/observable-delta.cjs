'use strict';
const fs=require('node:fs'),vm=require('node:vm'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('node:assert/strict');
const s='/workspace/scratch/elementat-declaration',manifest=JSON.parse(fs.readFileSync(s+'/probe-manifest.json','utf8')),rows=[];
for(const [variant,owner] of [['const',manifest.owner],['declaration',manifest.declaration]]){
 const source=manifest.helper+`\nlet early;try{early=typeof elementAt;}catch(e){early=e.name;}\n`+owner.replace('export ','')+`\nlet construct;try{new (elementAt as any)(undefined,0);construct='succeeded';}catch(e){construct=e.name;}\nglobalThis.row={variant:${JSON.stringify(variant)},early,name:elementAt.name,length:elementAt.length,hasPrototype:Object.prototype.hasOwnProperty.call(elementAt,'prototype'),construct};`;
 const context={};vm.runInNewContext(ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.None}}).outputText,context);rows.push(context.row);
}
assert.equal(rows[0].early,'ReferenceError');assert.equal(rows[1].early,'function');assert.equal(rows[0].hasPrototype,false);assert.equal(rows[1].hasPrototype,true);assert.equal(rows[0].construct,'TypeError');assert.equal(rows[1].construct,'succeeded');assert.equal(rows[0].name,rows[1].name);assert.equal(rows[0].length,rows[1].length);
console.log(JSON.stringify({measurementOnly:true,rows},null,2));
