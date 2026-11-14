'use strict';
// Execute the readers' own pure filesystem functions; never run historical build/oracle drivers.
const fs=require('fs'),path=require('path'),crypto=require('crypto'),vm=require('vm');
const ts=require('/workspace/scratch/recordreplay/head/stage3/api/node_modules/typescript');
const R='/workspace/scratch/recordreplay/head',D='/workspace/scratch/recordreplay';
const catalog=JSON.parse(fs.readFileSync(D+'/inventory-audit/catalog.json','utf8'));const results=[];
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
for(const entry of catalog.readers.filter(e=>/\.(?:cjs|js|mjs)$/.test(e.file))){
 const code=fs.readFileSync(path.join(R,entry.file),'utf8'),file=ts.createSourceFile(entry.file,code,99,true,ts.ScriptKind.JS);const defs=[],fns=[];
 for(const stmt of file.statements){
  if(ts.isFunctionDeclaration(stmt)&&stmt.name){const body=stmt.getText(file).replace(/^export\s+/,'');defs.push(body);fns.push({name:stmt.name.text,node:stmt,body});}
  if(ts.isVariableStatement(stmt))for(const decl of stmt.declarationList.declarations){if(ts.isIdentifier(decl.name)&&decl.initializer&&(ts.isArrowFunction(decl.initializer)||ts.isFunctionExpression(decl.initializer))){const body=decl.initializer.getText(file);defs.push('const '+decl.name.text+' = '+body+';');fns.push({name:decl.name.text,node:decl.initializer,body});}}
 }
 const selected=fns.filter(f=>/readdirSync\(/.test(f.body)&&!/(?:writeFile|mkdir|copyFile|unlink|rmSync|spawn|execFile|\.emit\(|process\.exit|symlink|rename)/.test(f.body)&&f.node.parameters.filter(p=>!p.initializer&&!p.questionToken).length<=1);
 for(const fn of selected){
  const row={file:entry.file,line:file.getLineAndCharacterOfPosition(fn.node.getStart(file)).line+1,reader:fn.name,method:'original pure filesystem function, isolated from historical driver',source_sha256:entry.sha256,results:{}};
  for(const [side,tree] of [['control',D+'/control'],['amended',D+'/candidate2']]){
   const internalSrc=/['"]src(?:\/|['"])/.test(fn.body);const arg=internalSrc?tree:path.join(tree,'src');
   const readonly=new Proxy(fs,{get(t,k){if(/^(?:write|append|mkdir|rm|rmdir|unlink|rename|copyFile|symlink|link|chmod|chown|truncate)/.test(String(k)))return()=>{throw Error('inventory audit prohibits filesystem mutation')};return t[k]}});
   const context={files:[],callers:[],hashes:{},references:[],calls:[],callbacks:[],rows:[],names:[],defs:[],rules:fs.existsSync(path.join(path.dirname(path.join(R,entry.file)),'class-rules.json'))?JSON.parse(fs.readFileSync(path.join(path.dirname(path.join(R,entry.file)),'class-rules.json'),'utf8')):{brands:[]},fs:readonly,path,crypto,assert:require('assert/strict'),ts,hash,h:hash,console:{log(){},error(){}},root:tree,tree,directory:arg,before:tree,after:tree,process:{argv:['node',entry.file,tree,tree,D+'/inventory-audit/unused'],env:{}},__dirname:path.dirname(path.join(R,entry.file)),require(name){if(name==='typescript')return ts;if(name==='fs'||name==='node:fs')return readonly;if(name==='path'||name==='node:path')return path;if(name==='crypto'||name==='node:crypto')return crypto;if(name.startsWith('assert')||name.startsWith('node:assert'))return require('assert/strict');throw Error('unneeded historical dependency '+name)}};
   try{
    const value=vm.runInNewContext(defs.join('\n')+'\n;'+fn.name+'('+JSON.stringify(arg)+')',context,{timeout:20000});const names=new Set(),seen=new WeakSet();
    function add(x){if(typeof x!=='string')return;let p=x;if(!path.isAbsolute(p)){if(fs.existsSync(path.join(tree,p)))p=path.join(tree,p);else p=path.join(tree,'src',p);}if(p.startsWith(tree+'/src/')&&fs.existsSync(p)&&fs.statSync(p).isFile())names.add(path.relative(tree,p));}
    function visit(v){if(typeof v==='string'){add(v);return;}if(!v||typeof v!=='object'||Buffer.isBuffer(v)||seen.has(v)||typeof v.kind==='number'&&typeof v.pos==='number')return;seen.add(v);if(Array.isArray(v)){for(const x of v)visit(x);return;}if(Object.prototype.toString.call(v)==='[object Map]'){for(const [k,x]of v){add(k);visit(x);}return;}for(const [k,x]of Object.entries(v)){add(k);if(k!=='bytes'&&k!=='buffer')visit(x);}}
    visit(value);for(const key of ['files','hashes','rows','names','defs'])visit(context[key]);row.results[side]={status:'pass',source_files:names.size,contains_added:names.has('src/recordreplay.d.ts'),inventory_sha256:hash([...names].sort().join('\n')),result_kind:Array.isArray(value)?'array':typeof value};
   }catch(e){row.results[side]={status:'blocked',error:String(e.message).slice(0,400)};}
  }
  results.push(row);
 }
 if(!selected.length)results.push({file:entry.file,method:'no standalone pure source-directory reader; retained for caller/exclusion review',results:{}});
}
fs.writeFileSync(D+'/inventory-audit/js-readers.json',JSON.stringify(results,null,2)+'\n');console.log('JS reader projections',results.length,'blocked',results.filter(r=>r.results.amended?.status==='blocked').length);
