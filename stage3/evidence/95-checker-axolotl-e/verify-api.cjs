#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const sanctions=require('../../adapt/95-checker-axolotl-e/api-sanctions.json');
const {changes}=require('../../lane/normalize-api.cjs');
const [upstream,beforeBuilt,afterBuilt]=process.argv.slice(2);
assert(upstream,'usage: verify-api.cjs UPSTREAM_BASELINE [BEFORE_BUILT AFTER_BUILT]');
const digest=text=>crypto.createHash('sha256').update(text).digest('hex');
function ownerPath(node){const names=[];for(let n=node;n;n=n.parent)if(n.name&&!(ts.isModuleDeclaration(n)&&n.name.getText(node.getSourceFile())==='ts'))names.unshift(n.name.getText(node.getSourceFile()));return names;}
function find(text,row){
 const source=ts.createSourceFile('typescript.d.ts',text,ts.ScriptTarget.Latest,true),found=[];
 assert.equal(source.parseDiagnostics.length,0);
 function visit(node){if(ts.SyntaxKind[node.kind]===row.kind&&JSON.stringify(ownerPath(node))===JSON.stringify(row.path))found.push(node);ts.forEachChild(node,visit);}
 visit(source);assert.equal(found.length,1,'unique public API owner');return {source,node:found[0]};
}
const original=fs.readFileSync(upstream,'utf8');
for(const row of sanctions){const {source,node}=find(original,row);assert.equal(node.type.getText(source),row.before,'upstream before text must match byte for byte');assert.equal(row.after,row.before.replace('typeArguments?: NodeArray<TypeNode>;','typeArguments?: NodeArray<TypeNode> | undefined;'));}
const result={beforeTextByteIdentical:true,upstreamSha256:digest(original),sanctions:sanctions.length};
if(beforeBuilt||afterBuilt){
 assert(beforeBuilt&&afterBuilt);
 const before=fs.readFileSync(beforeBuilt,'utf8'),after=fs.readFileSync(afterBuilt,'utf8');
 for(const row of sanctions){const old=find(before,row),current=find(after,row);assert.equal(old.node.type.getText(old.source),row.before);assert.equal(current.node.type.getText(current.source),row.after);}
 const delta=changes(before,after);
 assert.equal(delta.length,1,'exactly one public API declaration may move');
 assert.equal(delta[0].declaration,'member:["ts","TypeChecker","signatureToSignatureDeclaration"]#0');
 let expected=before;
 const row=sanctions[0],old=find(before,row).node;const source=old.getSourceFile();
 const start=old.type.getStart(source),end=old.type.end;
 expected=before.slice(0,start)+row.after+before.slice(end);
 assert.equal(after,expected,'no other public API bytes may change');
 Object.assign(result,{beforeBuiltBytes:Buffer.byteLength(before),afterBuiltBytes:Buffer.byteLength(after),beforeBuiltSha256:digest(before),afterBuiltSha256:digest(after),delta});
}
console.log(JSON.stringify(result,null,2));
