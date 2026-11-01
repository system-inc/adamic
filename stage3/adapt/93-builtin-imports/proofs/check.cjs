'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript');
function hash(bytes){return crypto.createHash('sha256').update(bytes).digest('hex');}
function inventory(tree){return fs.readdirSync(path.join(tree,'src'),{recursive:true,withFileTypes:true}).filter(e=>e.isFile()).map(e=>path.relative(tree,path.join(e.parentPath,e.name))).sort();}
function undo(text){
 for(const [name,builtin] of [['adamicNodeFs93','fs'],['adamicNodePath93','path'],['adamicNodeOs93','os']]){
  const line=`import * as ${name} from "node:${builtin}";`;
  assert.equal(text.split(line).length,2,'reviewed import missing or duplicated');
  text=text.replace(new RegExp('\\r?\\n'+line.replace(/[.*+?^${}()|[\]\\]/g,'\\$&')),'');
  const assignment=`= ${name};`;assert.equal(text.split(assignment).length,2,'reviewed namespace binding moved');
  text=text.replace(assignment,`= require("${builtin}");`);
 }
 return text;
}
function TestBuiltinSourceScope(before,after){
 const names=inventory(before);assert.deepEqual(inventory(after),names,'source inventory changed');
 const entries=names.map(file=>({file,before:hash(fs.readFileSync(path.join(before,file))),after:hash(fs.readFileSync(path.join(after,file)))}));
 const changed=entries.filter(x=>x.before!==x.after).map(x=>x.file);
 assert.deepEqual(changed,['src/compiler/sys.ts'],'source changed outside sys.ts');
 const a=fs.readFileSync(path.join(before,'src/compiler/sys.ts'),'utf8'),b=fs.readFileSync(path.join(after,'src/compiler/sys.ts'),'utf8');
 assert.equal(undo(b),a,'TestBuiltinSourceScope: edit outside the three imports and initializers');
 return{files:names.length,changed,undoRestoresExactBytes:true,inventory:entries};
}
function TestBuiltinTranspileScope(before,after){
 const read=tree=>fs.readFileSync(path.join(tree,'src/compiler/sys.ts'),'utf8');
 const results=[];
 for(const removeComments of [false,true]){
  const emit=text=>ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
  const a=emit(read(before)),b=emit(read(after));assert.notEqual(a,b);
  assert.equal(undo(b),a,'JavaScript changed outside three imports and initializers');
  results.push({removeComments,beforeSha256:hash(a),afterSha256:hash(b),onlyReviewedChanges:true});
 }
 return results;
}
module.exports={hash,inventory,undo,TestBuiltinSourceScope,TestBuiltinTranspileScope};
if(require.main===module){
 const before=path.resolve(process.argv[2]),after=path.resolve(process.argv[3]);
 const result={scope:TestBuiltinSourceScope(before,after),transpile:TestBuiltinTranspileScope(before,after)};
 if(process.argv[4])fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');
 console.log(JSON.stringify({...result,scope:{...result.scope,inventory:`${result.scope.inventory.length} entries recorded in checks.json`}},null,2));
}
