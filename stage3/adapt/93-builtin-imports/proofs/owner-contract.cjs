'use strict';
// Run: NODE_PATH=stage3/api/node_modules node proofs/owner-contract.cjs BEFORE AFTER SCRATCH
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),child=require('node:child_process'),ts=require('typescript');
const directory=path.resolve(__dirname,'..'),name=path.basename(directory);
const manifest=require(name.startsWith('90-')?'../targets.json':'./targets.json');
const rows=manifest.files||manifest, [before,after,scratch]=process.argv.slice(2).map(p=>path.resolve(p));
assert(before&&after&&scratch);
fs.mkdirSync(scratch,{recursive:true});
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
function parse(tree,row){return ts.createSourceFile(row.file,fs.readFileSync(path.join(tree,row.file),'utf8'),ts.ScriptTarget.Latest,true)}
function nodes(root,predicate){const out=[];function visit(n){if(predicate(n))out.push(n);ts.forEachChild(n,visit)}visit(root);return out}
function owned(sf,row,state){
 const pins=row.owners||[{beforeSha256:row.ownerBeforeSha256,afterSha256:row.ownerAfterSha256}];
 return pins.map(pin=>{const found=nodes(sf,n=>ts.isFunctionLike(n)&&hash(n.getText(sf))===pin[state+'Sha256']);assert.equal(found.length,1,'fixture owner');return found[0]});
}
function snapshot(tree){return fs.readdirSync(path.join(tree,'src'),{recursive:true,withFileTypes:true}).filter(e=>e.isFile()).map(e=>path.join(e.parentPath,e.name)).sort().map(p=>({file:path.relative(tree,p),bytes:fs.readFileSync(p),mtime:fs.statSync(p,{bigint:true}).mtimeNs}));}
function equal(tree,expected){const actual=snapshot(tree);assert.deepEqual(actual.map(x=>[x.file,x.bytes]),expected.map(x=>[x.file,x.bytes]));}
function reset(source){const work=path.join(scratch,'tree');fs.rmSync(work,{recursive:true,force:true});fs.mkdirSync(work,{recursive:true});fs.cpSync(path.join(source,'src'),path.join(work,'src'),{recursive:true});return work;}
function run(work,label,ok){const log=fs.openSync(path.join(scratch,label+'.log'),'w');let r;try{r=child.spawnSync(process.execPath,[path.join(directory,'adapt.cjs'),work],{stdio:['ignore',log,log],timeout:180000,env:process.env});}finally{fs.closeSync(log)}assert.ifError(r.error);assert.equal(r.status,ok?0:1,label);return fs.readFileSync(path.join(scratch,label+'.log'),'utf8');}
const results=[];
for(const state of ['before','after']){
 const source=state==='before'?before:after;
 // One unrelated function in every formerly pinned file. Exercise both first apply and no-op.
 const work=reset(source),expected=snapshot(after),foreign=[];
 for(const row of rows){
  const sf=parse(work,row),owners=owned(sf,row,state);
  const candidates=nodes(sf,n=>ts.isFunctionDeclaration(n)&&n.body&&owners.every(o=>n.end<=o.getStart(sf)||n.getStart(sf)>=o.end));
  candidates.sort((a,b)=>(a.end-a.pos)-(b.end-b.pos));assert(candidates.length,'foreign function');
  const target=candidates[0],at=target.body.getStart(sf)+1,marker='/* owner contract foreign edit */\n';
  const text=sf.text.slice(0,at)+marker+sf.text.slice(at);fs.writeFileSync(path.join(work,row.file),text);
  const expectedEntry=expected.find(x=>x.file===row.file),next=parse(after,row);
  const matches=nodes(next,n=>ts.isFunctionDeclaration(n)&&n.name?.text===target.name.text&&n.getText(next)===target.getText(sf));assert.equal(matches.length,1,'foreign function unchanged by adaptation');
  const expectedAt=matches[0].body.getStart(next)+1;
  expectedEntry.bytes=Buffer.from(next.text.slice(0,expectedAt)+marker+next.text.slice(expectedAt));
  foreign.push({file:row.file,function:target.name.text});
 }
 run(work,state+'-foreign',true);equal(work,expected);
 const stamps=snapshot(work);const rerun=run(work,state+'-foreign-rerun',true);assert.equal(JSON.parse(rerun).files,0);assert.deepEqual(snapshot(work),stamps);
 results.push({state,foreign,accepted:true,expectedBytes:true,noOpPreservesBytesAndMtime:true});
 for(const row of rows){
  const work=reset(source),sf=parse(work,row),owner=owned(sf,row,state)[0];
  const at=ts.isBlock(owner.body)?owner.body.getStart(sf)+1:owner.body.getStart(sf);
  fs.writeFileSync(path.join(work,row.file),sf.text.slice(0,at)+'/* changed owner */'+sf.text.slice(at));
  const saved=snapshot(work),log=run(work,state+'-owner-'+path.basename(row.file),false);
  assert.match(log,/owner moved|uninventoried owner|target owner/,'scoped digest must refuse');assert.deepEqual(snapshot(work),saved,'refusal wrote source');
  results.push({state,file:row.file,changedOwnerRefused:true,noWrite:true});
 }
 // The insertion anchor and node: imports have their own top-level statement contract.
 for (const row of rows.filter(row => row.importsBeforeSha256)) {
  let work=reset(source),file=path.join(work,row.file),saved=fs.readFileSync(file,'utf8');
  const addition='\nimport type {} from "./types.js";\n';
  fs.writeFileSync(file,saved+addition);
  run(work,state+'-foreign-import',true);
  assert.equal(fs.readFileSync(file,'utf8'),fs.readFileSync(path.join(after,row.file),'utf8')+addition);
  work=reset(source);file=path.join(work,row.file);saved=fs.readFileSync(file,'utf8');
  const needle='import { errorCode } from "./hostErrors.js";';
  assert.equal(saved.split(needle).length,2);
  fs.writeFileSync(file,saved.replace(needle,'import { /* changed import owner */ errorCode } from "./hostErrors.js";'));
  const original=snapshot(work),log=run(work,state+'-changed-import',false);
  assert.match(log,/import block moved/);assert.deepEqual(snapshot(work),original);
  results.push({state,file:row.file,foreignImportAccepted:true,changedImportOwnerRefused:true,noWrite:true});
 }
}
fs.writeFileSync(path.join(__dirname,'owner-contract-results.json'),JSON.stringify({adaptation:name,results},null,2)+'\n');
console.log(JSON.stringify({adaptation:name,status:'pass',results},null,2));
