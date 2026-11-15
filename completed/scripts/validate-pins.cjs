'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),Module=require('module'),ts=require('typescript');
const root='/workspace/scratch/watch-file-1532-merge/stage3/adapt',r=__dirname,pairs=require('./snapshots.json');
function loadOwnTool(name){const filename=path.join(root,name,'adapt.cjs'),m=new Module(filename,module);m.filename=filename;m.paths=Module._nodeModulePaths(path.dirname(filename));m._compile(fs.readFileSync(filename,'utf8')+'\nmodule.exports.__pinHash=typeof digest==="function"?digest:hash;\n',filename);return m.exports;}
function all(sf){let out=[];function visit(n){out.push(n);ts.forEachChild(n,visit);}visit(sf);return out;}
function ownerPath(n){let out=[];for(let p=n;p;p=p.parent)if((ts.isFunctionLike(p)||ts.isClassDeclaration(p))&&p.name)out.unshift(p.name.getText(p.getSourceFile()));return out;}
const result=[];
for(const name of Object.keys(pairs)){
 const tool=loadOwnTool(name),isMulti=name.includes('core-multimap'),isSingle=name.includes('single-watcher'),isTimer=name.includes('timer-closures');
 const rel=isMulti||isSingle||isTimer?'owners.json':'proofs/targets.json',file=path.join(root,name,rel),original=fs.readFileSync(file,'utf8'),pins=JSON.parse(original);
 let checked=0;
 for(const [index,state] of ['before','after'].entries())for(const pin of pins){
  const sourcePath=path.join(pairs[name][index],isSingle?'src/compiler':'',pin.file);let text=fs.readFileSync(sourcePath,'utf8');
  if(isMulti||isSingle||isTimer)text=text.replace(/\r\n/g,'\n');
  else if(name.startsWith('99-wasm'))text=text.replace(/\r?\n/g,'\r\n');
  const sf=ts.createSourceFile(pin.file,text,ts.ScriptTarget.Latest,true);assert.equal(sf.parseDiagnostics.length,0);
  let found;
  if(isMulti){
   tool.inspect(pin.file,text);
   if(pin.kind==='import')found=sf.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text===pin.module);
   else if(pin.kind==='declaration')found=sf.statements.filter(n=>n.name?.text===pin.path[0]);
   else found=all(sf).filter(n=>(ts.isFunctionDeclaration(n)||ts.isMethodDeclaration(n))&&n.body&&n.name&&JSON.stringify(ownerPath(n))===JSON.stringify(pin.path));
  } else if(isSingle)found=sf.statements.filter(n=>(ts.isFunctionDeclaration(n)||ts.isTypeAliasDeclaration(n))&&n.name?.text===pin.owner);
  else if(isTimer)found=all(sf).filter(n=>((ts.isFunctionDeclaration(n)&&n.body)||ts.isClassDeclaration(n)||ts.isInterfaceDeclaration(n))&&n.name?.text===pin.owner);
  else found=all(sf).filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getNodeSystem');
  let actual='';
  if(found.length)actual=text.slice(found[0].getStart(sf),found.at(-1).end);
  if(isSingle)assert.equal(found.length,pin[state+'Declarations']);else assert(found.length===1||(isMulti&&state==='after'&&pin.after===''&&found.length===0));
  if(isMulti||isSingle||isTimer){assert.equal(actual,pin[state]);pin[(isMulti?'': 'owner')+(state==='before'?'Before':'After')+'Sha256']=tool.__pinHash(actual);if(isMulti){delete pin.BeforeSha256;delete pin.AfterSha256;pin[state+'Sha256']=tool.__pinHash(actual);}}
  else {
   const cap=state==='before'?'Before':'After';assert.equal(actual,pin['owner'+cap+'Text']);pin['owner'+cap+'Text']=actual;pin['owner'+cap+'Sha256']=tool.__pinHash(actual);
   let imports=sf.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text.startsWith('node:'));
   if(name.startsWith('93'))imports=[...sf.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text==='./hostErrors.js'),...imports];
   pin['imports'+cap+'Sha256']=tool.__pinHash(imports.map(n=>n.getText(sf)).join('\n'));
   if(name.startsWith('99-wasm')){const scoped=require(path.join(root,name,'proofs/scoped-hashes.cjs')).scopedHashes(text);assert.equal(scoped.ownerSha256,pin['owner'+cap+'Sha256']);assert.equal(scoped.importsSha256,pin['imports'+cap+'Sha256']);}
  }
  checked++;
 }
 const generated=JSON.stringify(pins,null,2)+'\n';assert.equal(generated,original,'regenerated manifest differs: '+name);
 fs.writeFileSync(path.join(r,name+'-regenerated-'+path.basename(rel)),generated);
 result.push({directory:name,manifest:rel,statesChecked:checked,regeneratedByteIdentical:true,hashImplementation:'directory adapt.cjs own digest/hash; crypto also scoped-hashes.cjs',checkpointSource:'full apply immediate before/after refs'});
}
fs.writeFileSync(path.join(r,'pin-validation.json'),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
