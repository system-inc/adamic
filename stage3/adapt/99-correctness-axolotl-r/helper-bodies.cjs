'use strict';
const fs=require('node:fs'),ts=require('typescript'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const [before,after]=process.argv.slice(2);assert(before&&after);const rows=[];
for(const owner of ['concatenate','relativeComplement']){
 const texts=[before,after].map(root=>{const name=root+'/src/compiler/core.ts',text=fs.readFileSync(name,'utf8'),source=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true),nodes=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===owner),implementation=nodes.filter(n=>n.body);assert.equal(implementation.length,1);return {source,nodes,implementation:implementation[0]};});
 const implementations=texts.map(t=>t.implementation.getText(t.source));assert.equal(implementations[0],implementations[1],owner+': implementation changed');
 const signatures=texts.map(t=>t.nodes.filter(n=>!n.body).map(n=>n.getText(t.source)));
 const removed=signatures[0].filter(s=>!signatures[1].includes(s)),added=signatures[1].filter(s=>!signatures[0].includes(s));
 assert.equal(removed.length,owner==='concatenate'?2:0);assert.equal(added.length,2);
 rows.push({owner,implementationByteIdentical:true,implementationSha256:crypto.createHash('sha256').update(implementations[0]).digest('hex'),removedOverloads:removed,addedOverloads:added});
}
console.log(JSON.stringify(rows,null,2));
