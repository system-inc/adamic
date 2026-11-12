const fs=require('fs'),assert=require('assert/strict'),vm=require('vm'),{createRequire}=require('module');
const file='/workspace/adamic/stage3/adapt/40-explicit-any/class-proof.cjs',req=createRequire(file),ts=req('typescript'),r=req('./class-rules.json').brands.find(r=>r.id===103);
const code=fs.readFileSync(file,'utf8').split('fs.mkdirSync(evidence,')[0]+'\nmodule.exports = {declarationProjection};',context={require:req,process:{argv:['node',file,'unused','unused','unused','brands']},module:{exports:{}}};
vm.runInNewContext(code,context,{filename:file});
const emit=source=>ts.transpileDeclaration(source,{fileName:'path.ts',compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
for(const prefix of ['', 'declare namespace ts {\n']){
 const input=prefix+'export type Path = string;\n'+r.upstream+(prefix?'\n}':'');
 const output=prefix+'export type Path = string;\n'+r.after+(prefix?'\n}':'');
 const before=emit(input),after=emit(output),expected=context.module.exports.declarationProjection(before,[{key:r.key,owner:r.owner}]);
 assert.equal(expected,after);
}
console.log('independent declaration projection matches stock declaration emission for module and namespace');
