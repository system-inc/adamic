const fs=require('fs'),path=require('path'),cp=require('child_process'),ts=require('/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/typescript');
const r='/workspace/scratch/write-file-slot-run';
const roots=ts.sys.readDirectory(r+'/main/src/compiler',['.ts']).sort();
fs.writeFileSync(r+'/roots.json',JSON.stringify(roots,null,2));
const options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:'esnext',moduleDetection:'force',moduleResolution:'bundler',target:'es2024',lib:['es2024'],types:['node'],typeRoots:['/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/@types']};
for(const label of ['main','a','b']){
 const names=roots.map(n=>n.replace(r+'/main/',r+'/'+label+'/'));
 const config=r+'/'+label+'-tsconfig.json';
 fs.writeFileSync(config,JSON.stringify({compilerOptions:options,files:names},null,2));
 const result=cp.spawnSync(process.execPath,['/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/typescript/lib/tsc.js','-p',config],{encoding:'utf8',maxBuffer:10e6});
 fs.writeFileSync(r+'/'+label+'-tsc.log',result.stdout+result.stderr);
 fs.writeFileSync(r+'/'+label+'-tsc-status.json',JSON.stringify({roots:names.length,status:result.status,options}));
}
const names=ts.sys.readDirectory(r+'/main/src',['.ts']).filter(n=>!n.endsWith('.d.ts'));
let comparisons=[];
for(const label of ['a','b']) for(const removeComments of [false,true]){
 const changes=[];
 for(const name of names){
  const relative=path.relative(r+'/main',name),one=fs.readFileSync(name,'utf8'),two=fs.readFileSync(r+'/'+label+'/'+relative,'utf8');
  const opts={compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  const a=ts.transpileModule(one,opts).outputText,b=ts.transpileModule(two,opts).outputText;
  if(a!==b){changes.push(relative);fs.writeFileSync(r+'/'+label+'-'+(removeComments?'strip':'keep')+'-before.js',a);fs.writeFileSync(r+'/'+label+'-'+(removeComments?'strip':'keep')+'-after.js',b);}
 }
 comparisons.push({label,removeComments,files:names.length,changes});
}
fs.writeFileSync(r+'/javascript.json',JSON.stringify(comparisons,null,2));
