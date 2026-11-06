#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const {prove,parse,noNameObservations,expected}=require('./rename-proof.cjs');
const [before,after,normalizedDirectory]=process.argv.slice(2);assert(before&&after,'usage: compare.cjs BEFORE AFTER [NEW_NORMALIZED_DIR]');if(normalizedDirectory)fs.mkdirSync(normalizedDirectory,{recursive:false});
const hash=data=>crypto.createHash('sha256').update(data).digest('hex'),files=root=>fs.readdirSync(root,{recursive:true}).filter(name=>fs.statSync(path.join(root,name)).isFile()).sort();
const declarations=['typescript.d.ts','tsserverlibrary.d.ts'].map(name=>{const a=fs.readFileSync(path.join(before,'built/local',name)),b=fs.readFileSync(path.join(after,'built/local',name));assert(a.equals(b),'public declaration: '+name);return {name,identical:true,bytes:a.length,sha256:hash(a)};});
const names=files(path.join(before,'built/local')).filter(name=>/\.(?:c|m)?js$/.test(name));assert.deepEqual(names,files(path.join(after,'built/local')).filter(name=>/\.(?:c|m)?js$/.test(name)));assert.equal(names.length,9,'all nine normal compiler outputs');
const esbuild=require(path.resolve(before,'node_modules/esbuild'));
function lower(member,canonical){
 const code=esbuild.transformSync('var __95h_probe = host.'+member+'?.bind(host);',{target:['es2020','node14.17']}).code,source=ts.createSourceFile('probe.js',code,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),found=[];
 function visit(n){if(ts.isVariableDeclaration(n)&&n.name.getText(source)==='__95h_probe')found.push(n);ts.forEachChild(n,visit);}visit(source);assert.equal(found.length,1);
 const expression=found[0].initializer.getText(source);assert.equal(expression,'(_a = host.'+member+') == null ? void 0 : _a.bind(host)','ordinary upstream target lowering');return expression.replace(/\b_a\b/g,canonical);
}
const emitted=names.map(name=>{
 const a=fs.readFileSync(path.join(before,'built/local',name),'utf8'),b=fs.readFileSync(path.join(after,'built/local',name),'utf8');let normalized=b,renameProof,observations,onlySixRewrites;
 if(a!==b){const result=prove(a,b);normalized=result.normalized;renameProof=result.proof;onlySixRewrites=normalized===expected(a,lower);}else{observations=noNameObservations(parse(a).source);onlySixRewrites=true;}
 if(normalizedDirectory)fs.writeFileSync(path.join(normalizedDirectory,name),normalized);
 return {name,identical:a===b,onlySixRewrites,renameProof,observations,beforeSha256:hash(a),afterSha256:hash(b),normalizedSha256:hash(normalized),beforeBytes:Buffer.byteLength(a),afterBytes:Buffer.byteLength(b)};
});
assert(emitted.some(row=>!row.identical),'runtime fold emitted');const pass=emitted.every(row=>row.onlySixRewrites);
console.log(JSON.stringify({status:pass?'pass':'fail',javascriptChanging:true,rewrites:6,checkerImportRemoved:true,programImportUntouched:true,emittedFiles:9,emitted,declarations},null,2));process.exitCode=pass?0:1;
