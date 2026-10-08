#!/usr/bin/env node
'use strict';
const fs=require('node:fs');
const os=require('node:os');
const path=require('node:path');
const assert=require('node:assert/strict');
const ts=require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
const {spawnSync}=require('node:child_process');
const tree=fs.mkdtempSync(path.join(os.tmpdir(),'optional-read-write-probe-'));
try {
 const compiler=path.join(tree,'src/compiler');fs.mkdirSync(compiler,{recursive:true});
 const text=fs.readFileSync(path.join(__dirname,'probes/read-write.a'),'utf8');
 const file=path.join(compiler,'read-write.a');fs.writeFileSync(file,text);
 fs.writeFileSync(path.join(compiler,'tsconfig.json'),JSON.stringify({files:['read-write.a'],compilerOptions:{target:'ES2020',module:'ESNext'}}));
 const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.TS);
 const sites=[];
 function visit(n){if(ts.isAsExpression(n) && ts.isTypeReferenceNode(n.type) && ts.isIdentifier(n.type.typeName) && n.type.typeName.text==='Wide'){const p=source.getLineAndCharacterOfPosition(n.getStart(source));sites.push({File:'read-write.a',Line:p.line+1,Column:p.character+1,Kind:'KindAsExpression',Property:'p',Source:'Narrow',Target:'Wide',Text:n.getText(source)});}ts.forEachChild(n,visit);}visit(source);
 const input=path.join(tree,'sites.json'),output=path.join(tree,'result.json');fs.writeFileSync(input,JSON.stringify(sites));
 const run=spawnSync(process.execPath,[path.join(__dirname,'read-write.cjs'),tree,input,output,'--fixture'],{encoding:'utf8',env:process.env});
 assert.equal(run.status,0,run.stderr);
 const report=JSON.parse(fs.readFileSync(output));
 assert.deepEqual(report.counts,{total:8,read_only:1,write:7});
 assert(report.rows[0].writes.some(w=>w.written==='2' && w.access==='c.p'));
 assert.equal(report.rows[1].writes.length,0,'homonymous unrelated write contaminated read');
 assert(report.rows[2].writes.some(w=>w.written==='3' && w.access==='alias.p'));
 assert(report.rows[3].writes.some(w=>w.property==='q' && w.written==='4'));
 assert(report.rows[4].writes.some(w=>w.property==='p' && w.written==='5'));
 assert(report.rows[5].writes.some(w=>w.operation==='Object.assign' && w.written==='6'));
 assert(report.rows[6].writes.some(w=>w.untyped_receiver && w.written==='7'));
 assert(report.rows[7].writes.some(w=>w.written==='8'));
 // Remove only the receiver-to-variable alias edge, using the compiler API.
 let alias;
 function find(n){if(ts.isVariableDeclaration(n) && n.name.getText(source)==='a')alias=n;ts.forEachChild(n,find);}find(source);
 const mutant=text.slice(0,alias.initializer.getStart(source))+'({ x: 0 } as Wide)'+text.slice(alias.initializer.end);
 fs.writeFileSync(file,mutant);
 // Sites below the mutation stay at identical line/column; the changed initializer is not a census site.
 const failed=spawnSync(process.execPath,[path.join(__dirname,'read-write.cjs'),tree,input,output,'--fixture'],{encoding:'utf8',env:process.env});assert.equal(failed.status,0,failed.stderr);
 const mutated=JSON.parse(fs.readFileSync(output));
 assert.deepEqual(mutated.counts,{total:8,read_only:2,write:6});
 console.log(JSON.stringify({status:'pass',original:report.counts,alias_edge_dropped:mutated.counts,
 checks:['direct multi-variable alias reaches c.p = 2','callback alias reaches alias.p = 3','unrelated p symbol leaves read-only site unchanged','dropping alias changes the write count','write to another missing member is detected','computed keyof member write is detected','Object.assign alias write is detected','untyped alias does not hide a write','library-returned receiver alias is detected']},null,2));
} finally {fs.rmSync(tree,{recursive:true,force:true});}
