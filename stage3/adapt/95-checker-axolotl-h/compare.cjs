#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const {prove,parse,noNameObservations}=require('./rename-proof.cjs');
const [before,after,normalizedDirectory]=process.argv.slice(2);
if(normalizedDirectory)fs.mkdirSync(normalizedDirectory,{recursive:false});assert(before&&after,'usage: compare.cjs BEFORE AFTER');
const fold=require('./sites.json').literals.find(row=>row.ownerAfterSha256),hash=data=>crypto.createHash('sha256').update(data).digest('hex');
const files=root=>fs.readdirSync(root,{recursive:true}).filter(name=>fs.statSync(path.join(root,name)).isFile()).sort();
const declarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{
 const a=fs.readFileSync(path.join(before,'built/local',name)),b=fs.readFileSync(path.join(after,'built/local',name));assert(a.equals(b),'public declaration: '+name);return {name,identical:true,bytes:a.length,sha256:hash(a)};
});
const names=files(path.join(before,'built/local')).filter(name=>/\.(?:c|m)?js$/.test(name));
assert.deepEqual(names,files(path.join(after,'built/local')).filter(name=>/\.(?:c|m)?js$/.test(name)));assert.equal(names.length,9,'all nine normal compiler outputs');
const esbuild=require(path.resolve(before,'node_modules/esbuild'));
const ownerIn=text=>{const source=ts.createSourceFile('bundle.js',text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),owners=[];function visit(node){if(ts.isFunctionDeclaration(node)&&node.name?.text===fold.owner)owners.push(node);ts.forEachChild(node,visit);}visit(source);assert.equal(owners.length,1);return {source,owner:owners[0]};};
const emitted=names.map(name=>{
 const a=fs.readFileSync(path.join(before,'built/local',name),'utf8'),b=fs.readFileSync(path.join(after,'built/local',name),'utf8');
 let expected=a,normalized=b,rewrites=0,ownerMatchesExpectedLowering=true,outsideOwnerIdentical=true,renameProof,unrenamedObservations;
 if(a!==b){
  const {source,owner}=ownerIn(a);let text=owner.getText(source);
  for(const member of fold.members.filter(member=>typeof member.initializer==='object')){
   const before=`${member.name}: ${member.initializer.before}`,after=`${member.name}: ${member.initializer.after}`;
   assert.equal(text.split(before).length,2,'one emitted initializer: '+name);text=text.replace(before,after);rewrites++;
  }
  // Use the ordinary upstream build target, including Node's optional-chain lowering.
  const lowered=esbuild.transformSync(text,{target:['es2020','node14.17']}).code.trimEnd();
  const next=ownerIn(b);ownerMatchesExpectedLowering=next.owner.getText(next.source)===lowered;
  assert(ownerMatchesExpectedLowering,'only two rewrites and ordinary lowering within owner');
  const result=prove(a,b);normalized=result.normalized;renameProof=result.proof;
  assert.equal(lowered.split(result.afterArrow).length,2,'one lowered getter arrow');
  const normalizedLowered=lowered.replace(result.afterArrow,result.normalizedArrow);
  expected=a.slice(0,owner.getStart(source))+normalizedLowered+a.slice(owner.end);
  const normalizedOwner=ownerIn(normalized);
  outsideOwnerIdentical=a.slice(0,owner.getStart(source))+a.slice(owner.end)===normalized.slice(0,normalizedOwner.owner.getStart(normalizedOwner.source))+normalized.slice(normalizedOwner.owner.end);
 }
 else unrenamedObservations=noNameObservations(parse(a).source);
 const onlyTwoRewrites=normalized===expected;
 if(normalizedDirectory)fs.writeFileSync(path.join(normalizedDirectory,name),normalized);
 return {name,identical:a===b,onlyTwoRewrites,renameProof,unrenamedObservations,normalizedSha256:hash(normalized),ownerMatchesExpectedLowering,outsideOwnerIdentical,rewrites,beforeBytes:Buffer.byteLength(a),afterBytes:Buffer.byteLength(b),beforeSha256:hash(a),afterSha256:hash(b)};
});
assert(emitted.some(row=>row.rewrites===2),'runtime fold emitted');
const pass=emitted.every(row=>row.onlyTwoRewrites);
console.log(JSON.stringify({status:pass?'pass':'fail',javascriptChanging:true,emittedFiles:emitted.length,emitted,declarations},null,2));
process.exitCode=pass?0:1;
