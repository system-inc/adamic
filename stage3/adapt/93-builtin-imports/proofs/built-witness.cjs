'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),child=require('node:child_process'),crypto=require('node:crypto');
function hash(bytes){return crypto.createHash('sha256').update(bytes).digest('hex');}
function observe(file){
 const code=`const Module=require('node:module');const original=Module._load;const loads=[];Module._load=function(name,...args){if(['fs','path','os','node:fs','node:path','node:os'].includes(name))loads.push({name,parent:args[0]?.filename,stack:new Error().stack.split('\\n').slice(2,8)});return original.call(this,name,...args)};const ts=require(process.argv[1]);Module._load=original;console.log(JSON.stringify({node:process.version,version:ts.version,loads}));`;
 return JSON.parse(child.execFileSync(process.execPath,['-e',code,path.resolve(file)],{encoding:'utf8',timeout:60000}));
}
function TestBuiltBuiltinImports(before,after,scratch){
 const a=observe(path.join(before,'built/local/typescript.js')),b=observe(path.join(after,'built/local/typescript.js'));
 assert.equal(a.node,'v24.19.0');assert.equal(b.node,a.node);assert.equal(b.version,a.version);
 for(const name of ['fs','path','os']){
  const old=a.loads.filter(x=>x.name===name&&x.stack.some(s=>s.includes('getNodeSystem')));assert(old.length>0,`main ${name} load not observed in getNodeSystem`);
  const next=b.loads.filter(x=>(x.name===name||x.name===`node:${name}`)&&x.parent===path.join(after,'built/local/typescript.js'));assert(next.length>0,`top-level ${name} not loaded`);
  assert(next.every(x=>!x.stack.some(s=>s.includes('getNodeSystem'))),`${name} still loaded inside getNodeSystem`);
 }
 fs.mkdirSync(scratch,{recursive:true});
 fs.writeFileSync(path.join(scratch,'main.ts'),'export const answer: number = 6 * 7;\nexport function greet(name: string): string { return `hello ${name}`; }\n');
 fs.writeFileSync(path.join(scratch,'tsconfig.json'),JSON.stringify({compilerOptions:{strict:true,target:'es2024',module:'commonjs',declaration:true,skipLibCheck:true},files:['main.ts']}));
 const runs=[];
 for(const [label,tree] of [['control',before],['after',after]]){
  const command=[path.join(tree,'built/local/tsc.js'),'--project',path.join(scratch,'tsconfig.json'),'--outDir',path.join(scratch,label),'--pretty','false'];
  const r=child.spawnSync(process.execPath,command,{cwd:scratch,encoding:'utf8',timeout:60000});assert.ifError(r.error);
  runs.push({label,command,exit:r.status,stdout:r.stdout,stderr:r.stderr});assert.equal(r.status,0);
 }
 assert.deepEqual([runs[1].exit,runs[1].stdout,runs[1].stderr],[runs[0].exit,runs[0].stdout,runs[0].stderr]);
 const output=[];
 for(const file of ['main.js','main.d.ts']){const x=fs.readFileSync(path.join(scratch,'control',file)),y=fs.readFileSync(path.join(scratch,'after',file));assert.deepEqual(y,x);output.push({file,bytes:x.length,sha256:hash(x),byteIdentical:true});}
 return{test:'TestBuiltBuiltinImports',before:a,after:b,cli:runs,output,loadedDuringModuleEvaluation:true};
}
module.exports={TestBuiltBuiltinImports};
if(require.main===module){const result=TestBuiltBuiltinImports(...process.argv.slice(2,5));fs.writeFileSync(process.argv[5],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));}
