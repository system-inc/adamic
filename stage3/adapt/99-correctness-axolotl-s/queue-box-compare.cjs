'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto'),ts=require('typescript'),cp=require('child_process');
const [before,after,out]=process.argv.slice(2),hash=s=>crypto.createHash('sha256').update(s).digest('hex');
const inventory=root=>fs.readdirSync(root,{recursive:true}).filter(f=>fs.statSync(path.join(root,f)).isFile()).sort();
const names=inventory(before+'/src');assert.deepEqual(inventory(after+'/src'),names);
const changed=names.filter(n=>!fs.readFileSync(before+'/src/'+n).equals(fs.readFileSync(after+'/src/'+n)));assert.deepEqual(changed,['compiler/core.ts']);
const pin=require('./adapt.cjs').owners.find(p=>p.owner==='createQueue'),norm=s=>s.replaceAll('\r\n','\n');
const old=fs.readFileSync(before+'/src/compiler/core.ts','utf8'),now=fs.readFileSync(after+'/src/compiler/core.ts','utf8');assert.equal(norm(old).replace(pin.previousAfter,pin.after),norm(now));
function onlyOwner(a,b,file){
 function span(text){const sf=ts.createSourceFile(file,text,99,true,ts.ScriptKind.JS),found=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='createQueue')found.push(n);ts.forEachChild(n,visit);}visit(sf);assert.equal(found.length,1,file);return [found[0].getStart(sf),found[0].end];}
 const [x,y]=span(a),[u,v]=span(b);assert.equal(a.slice(0,x)+b.slice(u,v)+a.slice(y),b,file+' changed outside createQueue');return true;
}
const core=[false,true].map(removeComments=>{const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText,a=emit(old),b=emit(now);return {removeComments,onlyCreateQueue:onlyOwner(a,b,'core.js'),before:hash(a),after:hash(b)};});
const declarations=[];for(const dir of ['lib','built/local'])for(const name of ['typescript.d.ts','tsserverlibrary.d.ts']){const a=fs.readFileSync(before+'/'+dir+'/'+name),b=fs.readFileSync(after+'/'+dir+'/'+name);assert(a.equals(b));declarations.push({file:dir+'/'+name,byteIdentical:true,sha256:hash(a)});}
const files=['_tsc.js','_tsserver.js','_typingsInstaller.js','tsc.js','tsserver.js','tsserverlibrary.js','typescript.js','typingsInstaller.js','watchGuard.js','run.js'];
const emitted=files.map(name=>{const a=fs.readFileSync(before+'/built/local/'+name,'utf8'),b=fs.readFileSync(after+'/built/local/'+name,'utf8');const identical=a===b;return {name,identical,onlyCreateQueue:identical||onlyOwner(a,b,name),before:hash(a),after:hash(b)};});
for(const [tag,a,b] of [['source',before+'/src/compiler/core.ts',after+'/src/compiler/core.ts'],...files.map(n=>[n,before+'/built/local/'+n,after+'/built/local/'+n])]){const diff=cp.spawnSync('diff',['-u',a,b],{encoding:'utf8'});assert([0,1].includes(diff.status));fs.writeFileSync(out+'/'+tag+'.diff',diff.stdout);}
console.log(JSON.stringify({sourceFiles:names.length,changed,onlyPinnedOwnerChanged:true,core,declarations,emitted},null,2));
