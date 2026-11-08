"use strict";
const fs=require("node:fs"),path=require("node:path"),assert=require("node:assert/strict"),crypto=require("node:crypto"),ts=require("typescript");
const pins=require("./owners.json");
const hash=t=>crypto.createHash("sha256").update(t).digest("hex");
function inspect(text){
 const source=ts.createSourceFile(pins.file,text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,"core.ts parse error");
 const owners=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===pins.owner&&n.body);assert.equal(owners.length,1,"cast: missing, nested or duplicate implementation");
 const node=owners[0],owner=node.getText(source).replace(/\r\n/g,"\n");const state=owner===pins.before?"before":owner===pins.after?"after":null;assert(state,"cast: unreviewed whole owner");assert.equal(hash(owner),pins[state+"Sha256"]);return {node,state};
}
function plan(text){
 assert.equal(ts.version,"6.0.3");const {node,state}=inspect(text);if(state==="after")return {after:text,edits:0};
 const owner=node.getText(),newOwner=owner.replace(pins.removedPhrase,"");assert.notEqual(owner,newOwner);assert.equal(newOwner.replace(/\r\n/g,"\n"),pins.after,"only the function-name phrase may change");
 const after=text.slice(0,node.getStart())+newOwner+text.slice(node.end);assert.equal(inspect(after).state,"after");
 const emit=t=>ts.transpileModule(t,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
 assert.equal(emit(after),emit(text).replace(pins.removedPhrase,""),"unreviewed emitted JavaScript");
 return {after,edits:1};
}
function adapt(tree){const file=path.join(tree,pins.file),text=fs.readFileSync(file,"utf8"),result=plan(text);assert.equal(fs.readFileSync(file,"utf8"),text,"concurrent source edit");if(result.edits)fs.writeFileSync(file,result.after);return {files:result.edits?1:0,edits:result.edits,javascriptChanging:true};}
module.exports={adapt,plan,inspect,pins};
if(require.main===module){try{assert.equal(process.argv.length,3,"usage: node adapt.cjs TREE");console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));}catch(error){console.error(error.stack);process.exitCode=1;}}
