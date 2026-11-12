const assert=require('assert/strict'),fs=require('fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const base='/workspace/adamic/stage3/adapt/40-explicit-any',r=require(base+'/class-rules.json').brands.find(r=>r.id===103),{plan}=require(base+'/classes.cjs');
const results=[];
for(const spelling of [r.upstream,r.before,r.after])for(const eol of ['\n','\r\n']){
 const source=eol.repeat(456)+spelling+eol,result=plan(source,r.file);
 assert.equal(result.text,eol.repeat(456)+r.after+eol);assert.equal(result.removed,Number(spelling===r.upstream));assert.equal(plan(result.text,r.file).text,result.text);
 for(const removeComments of [true,false]){
  const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
  assert.equal(emit(source),emit(result.text));
 }
 results.push({input:spelling,eol:JSON.stringify(eol),edits:Number(result.text!==source),replayEdits:0,javascriptIdentical:true});
}
const mutants=[r.after.replace('Array<Path>','Array<string>'),r.after.replace('undefined','void'),r.after.replace('interface','type'),r.after.replace('export ',''),r.after+'\n'+r.after];
for(const spelling of mutants)assert.throws(()=>plan('\n'.repeat(456)+spelling+'\n',r.file));
fs.writeFileSync('/workspace/scratch/path-components-proof/planner.json',JSON.stringify({results,driftMutantsRejected:mutants.length},null,2)+'\n');console.log('planner cases='+results.length+' drift mutants='+mutants.length);
