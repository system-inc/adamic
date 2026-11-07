'use strict';
const fs=require('node:fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),cp=require('node:child_process');
const s='/workspace/scratch/gov-a',r=JSON.parse(fs.readFileSync(s+'/raw/node-witnesses.json','utf8'));
const prelude='export interface MapLike<T> {[index:string]:T;}\nconst hasOwnProperty=Object.prototype.hasOwnProperty;\n'+r.isArrayOwner+'\n';
const output="console.log('filtered-length='+String(result.length));for(const value of result)console.log(value===undefined?'undefined':value);";
const cases={arrayPresent:prelude+"function arrayArm<T>(collection:T[]):T[]{return collection.filter(()=>true);}const result=arrayArm<string | undefined>([undefined,'x']);"+output,arrayEmpty:prelude+"function arrayArm<T>(collection:T[]):T[]{return collection.filter(()=>true);}const result=arrayArm<string>([]);"+output,recordBranch:prelude+r.adaptedOwner.replace('return collection.filter(() => true);','return [];')+"\nconst result=getOwnValues<string>({'1':'a','0':'b'});"+output};
for(const [name,text] of Object.entries(cases)){
 const stem=s+'/witnesses/'+name;fs.writeFileSync(stem+'.ts',text);fs.writeFileSync(stem+'.mjs',ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText);
 const run=cp.spawnSync(process.execPath,[stem+'.mjs'],{encoding:'utf8'});if(run.status!==0)throw Error(run.stderr);fs.writeFileSync(stem+'.node.stdout',run.stdout);fs.writeFileSync(stem+'.node.stderr',run.stderr);
}
