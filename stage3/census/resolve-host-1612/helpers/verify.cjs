'use strict';
const fs=require('fs'),path=require('path'),ts=require('typescript'),assert=require('assert/strict'),crypto=require('crypto');
const s=__dirname,a=s+'/build-base',b=s+'/build-change',out=s+'/evidence';
const spec=require(s+'/work/stage3/adapt/99-wasm-crypto-import/proofs/resolve-host.json');
const {compiler}=require(s+'/work/stage3/adapt/47-host-errors/adapt.cjs');
function hash(x){return crypto.createHash('sha256').update(x).digest('hex');}
function inventory(tree){return fs.readdirSync(tree+'/src',{recursive:true,withFileTypes:true}).filter(e=>e.isFile()).map(e=>path.relative(tree,path.join(e.parentPath,e.name))).sort();}
const names=inventory(a);assert.deepEqual(inventory(b),names);assert.equal(names.length,747);
const src=[];const js=[];let comparisons=0;
for(const file of names){const before=fs.readFileSync(a+'/'+file),after=fs.readFileSync(b+'/'+file);if(!before.equals(after))src.push(file);
 if(!file.endsWith('.ts')||file.endsWith('.d.ts'))continue;
 for(const removeComments of [false,true]){
  const emit=text=>ts.transpileModule(text,{compilerOptions:{module:ts.ModuleKind.ESNext,target:ts.ScriptTarget.ES2024,removeComments}}).outputText;
  const x=emit(before.toString()),y=emit(after.toString());comparisons++;
  if(x!==y){assert.equal(file,'src/compiler/sys.ts');let text=after.toString(),nl=text.includes('\r\n')?'\r\n':'\n';
   for(const [from,to]of[[spec.after,spec.before],[spec.annotationAfter,spec.annotationBefore]]){const normalized=from.replace(/\n/g,nl);assert.equal(text.split(normalized).length,2);text=text.replace(normalized,to);}
   assert.equal(emit(text),x);js.push({file,removeComments,before:hash(x),after:hash(y),onlyResolveHostSite:true});
  }
 }
}
assert.deepEqual(src,['src/compiler/sys.ts']);assert.equal(js.length,2);
fs.writeFileSync(out+'/javascript.json',JSON.stringify({sourceFiles:names.length,changedSource:src,emittedFiles:comparisons/2,comparisons,differing:js,otherJavaScriptByteIdentical:true},null,2)+'\n');
const p=compiler(b,ts),roots=p.getRootFileNames(),options={...p.getCompilerOptions(),typeRoots:[s+'/work/stage3/api/node_modules/@types']};assert.equal(roots.length,79);const results=[];
for(const label of ['base','change']){const text=fs.readFileSync((label==='base'?a:b)+'/src/compiler/sys.ts','utf8'),host=ts.createCompilerHost(options,true),original=host.getSourceFile.bind(host);host.getSourceFile=(name,language,onError,create)=>path.resolve(name)===b+'/src/compiler/sys.ts'?ts.createSourceFile(name,text,language,true):original(name,language,onError,create);const program=ts.createProgram(roots,options,host);const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({file:d.file?.fileName,code:d.code,start:d.start,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));results.push({label,diagnostics});}
fs.writeFileSync(out+'/stock-accepted.json',JSON.stringify({typescript:ts.version,roots,options,measurements:results},null,2)+'\n');for(const r of results)assert.equal(r.diagnostics.length,0);
console.log(JSON.stringify({source:src,JavaScriptComparisons:comparisons,JavaScriptDifferences:js.length,roots:roots.length,stock:results},null,2));
