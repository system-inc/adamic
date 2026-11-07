'use strict';
const e=require('./engine.cjs'),{fs,path,assert}=e,comparison=require('./compare.cjs');
const inventory=require('./inventory.json'),renames=require('./composition-renames.cjs');
const [before,after,out,folds='']=process.argv.slice(2);assert(before&&after&&out);
const wanted=new Set(folds.split(',').filter(Boolean));assert([...wanted].every(x=>['32','95h'].includes(x)));
const sites=[...inventory.sites,...inventory.left.filter(s=>(wanted.has('32')&&s.adapter==='32-indexed-reads-program')||(wanted.has('95h')&&s.file==='checker.ts'))];
const proofs={};
comparison.configure(sites,wanted.has('95h')?(name,before,after)=>{
 if(!['_tsc.js','typescript.js'].includes(name))return after;
 const result=renames.prove(before,after);proofs[name]=result.proof;return result.normalized;
}:undefined);
fs.mkdirSync(out,{recursive:true});
const result=comparison.compare(path.resolve(before),path.resolve(after),path.resolve(out));
if(wanted.has('95h')){
 for(const root of [before,after])for(const name of fs.readdirSync(path.join(root,'built/local')).filter(n=>n.endsWith('.js'))){const text=fs.readFileSync(path.join(root,'built/local',name),'utf8');renames.noNameObservations(renames.parse(text).source);}
 fs.writeFileSync(path.join(out,'rename-map.json'),JSON.stringify(proofs,null,2)+'\n');
}
const declared=['typescript.d.ts','typescript.internal.d.ts','tsserverlibrary.d.ts'].filter(name=>fs.existsSync(path.join(before,'built/local',name)));
const declarations=declared.map(name=>{const a=fs.readFileSync(path.join(before,'built/local',name)),b=fs.readFileSync(path.join(after,'built/local',name));assert.deepEqual(a,b,name);return {file:name,byteIdentical:true,sha256:e.digest(a)};});
fs.writeFileSync(path.join(out,'declarations.json'),JSON.stringify(declarations,null,2)+'\n');
for(const row of result.files){if(!row.byteIdentical)assert.equal(row.expectedSites,sites.length);}
console.log(JSON.stringify({pass:result.pass,expectedSites:sites.length,folds:[...wanted],files:result.files},null,2));process.exitCode=result.pass?0:1;
