'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('/workspace/scratch/cast-tags/work/stage3/api/node_modules/typescript');
const r='/workspace/scratch/cast-tags',e=path.join(r,'evidence');assert.equal(ts.version,'6.0.3');
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const diagnostics={};
for(const side of ['main','c']){
 const tree=path.join(r,side+'-tree'),roots=ts.sys.readDirectory(path.join(tree,'src/compiler'),['.ts']).sort();assert.equal(roots.length,79);
 const program=ts.createProgram(roots,options),found=ts.getPreEmitDiagnostics(program).map(d=>({file:d.file?.fileName.replace(tree,'<tree>'),start:d.start,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
 diagnostics[side]={roots:roots.map(x=>path.relative(tree,x)),diagnostics:found};console.log(side,'stock diagnostics',found.length);
}
fs.writeFileSync(path.join(e,'stock-diagnostics.json'),JSON.stringify(diagnostics,null,2)+'\n');
assert.deepEqual(diagnostics.main.diagnostics,diagnostics.c.diagnostics,'new stock diagnostics');assert.equal(diagnostics.c.diagnostics.length,0);
const names=ts.sys.readDirectory(path.join(r,'main-tree/src'),['.ts']).filter(x=>!x.endsWith('.d.ts')).sort();let changed=0,emits=0;
for(const beforeName of names){const relative=path.relative(path.join(r,'main-tree'),beforeName),afterName=path.join(r,'c-tree',relative),before=fs.readFileSync(beforeName,'utf8'),after=fs.readFileSync(afterName,'utf8');if(before!==after)changed++;
 for(const removeComments of [false,true]){const compilerOptions={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};const a=ts.transpileModule(before,{fileName:relative,compilerOptions}).outputText,b=ts.transpileModule(after,{fileName:relative,compilerOptions}).outputText;assert.equal(a,b,relative+' JavaScript differs');emits++;}}
assert.equal(changed,2);fs.writeFileSync(path.join(e,'javascript-identity.json'),JSON.stringify({sourceFiles:names.length,changedInputFiles:changed,pairedEmits:emits,identical:true},null,2)+'\n');console.log('JavaScript identical',names.length,'files',emits,'paired emits');
