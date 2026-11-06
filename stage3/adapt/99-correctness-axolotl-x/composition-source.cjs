'use strict';
const e=require('./engine.cjs'),{fs,path,assert,ts}=e,{adapt,owners}=require('./adapt.cjs');
const inventory=require('./inventory.json'),[before,after,out,folds='']=process.argv.slice(2),wanted=new Set(folds.split(','));
const active=[...inventory.sites,...inventory.left.filter(s=>(wanted.has('32')&&s.adapter==='32-indexed-reads-program')||(wanted.has('95h')&&s.file==='checker.ts'))];
const signature=s=>JSON.stringify([s.file,s.owner,s.before]);
const originals=new Map(e.sources(path.resolve(before)).map(f=>[f,fs.readFileSync(f,'utf8')]));
const replacements=new Map([...originals].map(([f,t])=>[f,e.normalize(t)]));
const keys=new Set(active.map(signature)),scan=e.scan(path.resolve(before),e.program(path.resolve(before),replacements));
const sites=scan.sites.filter(s=>keys.has(signature(s)));assert.equal(sites.length,active.length);
const expected=new Map();const imports=[];
for(const file of new Set(sites.map(s=>s.file))){
 const original=fs.readFileSync(path.join(before,'src/compiler',file),'utf8');
 let text=e.normalize(original);text=e.edit(text,sites.filter(s=>s.file===file));
 const count=scan.sites.filter(s=>s.file===file).length-sites.filter(s=>s.file===file).length;
 if(count===0){
  // Checker import belongs to 95h; permit only that exact cleanup, if its fold performs it.
  const actual=fs.readFileSync(path.join(after,'src/compiler',file),'utf8');
  if(file!=='checker.ts'||!/^    maybeBind,\r?$/m.test(actual)){
   const removal=e.removableImport(file,text);assert(removal,file);text=e.edit(text,[removal]);imports.push(file);
  }
 }
 if(original.includes('\r\n'))text=text.replace(/\n/g,'\r\n');
 expected.set('compiler/'+file,text);
}
function all(root){return ts.sys.readDirectory(path.join(root,'src'),undefined,undefined,['**/*']).map(f=>path.relative(path.join(root,'src'),f)).sort();}
const files=all(before);assert.deepEqual(all(after),files);const changed=[];
for(const file of files){const a=fs.readFileSync(path.join(before,'src',file)),b=fs.readFileSync(path.join(after,'src',file));assert.deepEqual(b,expected.has(file)?Buffer.from(expected.get(file)):a,file);if(!a.equals(b))changed.push(file);}
const sourceFiles=e.sources(path.resolve(after)),state=sourceFiles.map(f=>[f,fs.readFileSync(f),fs.statSync(f,{bigint:true}).mtimeNs]);
const replay=adapt(path.resolve(after));assert.equal(replay.files,0);
for(const [f,bytes,time] of state){assert.deepEqual(fs.readFileSync(f),bytes);assert.equal(fs.statSync(f,{bigint:true}).mtimeNs,time);}
const result={pass:true,expectedSites:sites.length,sourceFiles:files.length,changed,importsRemoved:imports,replay,bytesAndMtimePreserved:true};
fs.mkdirSync(out,{recursive:true});fs.writeFileSync(path.join(out,'source-comparison.json'),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
