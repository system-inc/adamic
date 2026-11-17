'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('../../api/node_modules/typescript');
const [tree,set]=process.argv.slice(2),file=path.join(tree,'src/compiler/visitorPublic.ts');
const raw=fs.readFileSync(file,'utf8'),text=raw.replace(/\r\n/g,'\n');
assert(!raw.includes('\r\n')||!raw.replace(/\r\n/g,'').includes('\n'));
const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const nodes=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===set),impl=nodes.find(n=>n.body);assert(impl);
let edits=[];
const change=(n,value)=>edits.push({start:n.getStart(sf),end:n.end,value});
if(set==='visitArrayWorker'){
 for(const n of nodes.filter(n=>!n.body))change(n,'');
}else if(set==='visitParameterList'||set==='visitEachChild'){
 const p=impl.parameters.find(p=>p.name.getText(sf)==='nodesVisitor');assert(p&&!p.type);
 change(p,'nodesVisitor: NodesVisitor = visitNodes');
}else if(set==='visitNode'){
 change(impl.parameters[0].type,'Node | undefined');
}else if(set==='visitNodes'||set==='visitArray'){
 const nameEnd=impl.name.end;edits.push({start:nameEnd,end:nameEnd,value:'<TIn extends Node>'});
 change(impl.parameters[0].type,set==='visitNodes'?'NodeArray<TIn> | undefined':'readonly TIn[] | undefined');
 change(impl.parameters[1].type,'Visitor<TIn, Node | undefined>');
}else throw Error('unreviewed set');
let after=text;for(const e of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,e.start)+e.value+after.slice(e.end);
const emits=[];
for(const removeComments of [false,true]){
 const opts={compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
 const beforeJS=ts.transpileModule(text,opts).outputText,afterJS=ts.transpileModule(after,opts).outputText;
 assert.equal(afterJS,beforeJS);emits.push({removeComments,bytes:Buffer.byteLength(afterJS),identical:true});
}
fs.writeFileSync(file,raw.includes('\r\n')?after.replace(/\n/g,'\r\n'):after);
console.log(JSON.stringify({set,edits:edits.length,emits}));
