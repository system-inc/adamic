#!/usr/bin/env node
'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {execFileSync}=require('node:child_process');
const ts=require('typescript'),pins=require('./meta-spread-pins.json'),{adapt}=require('./adapt.cjs');
const [tree,scratch,output]=process.argv.slice(2);
assert(tree&&scratch&&output,'usage: meta-spread-pin-witness.cjs adapted-tree scratch output.json');
fs.mkdirSync(scratch);
for(const entry of fs.readdirSync(tree)){
 if(entry==='src')fs.cpSync(path.join(tree,entry),path.join(scratch,entry),{recursive:true});
 else fs.symlinkSync(path.resolve(tree,entry),path.join(scratch,entry));
}
// Use 47's own immediate sys, tracing and helper output. Later adapters merge
// the sys import and edit the helper; this witness weakens neither guard.
for(const name of ['sys.ts','tracing.ts','hostErrors.ts'])fs.writeFileSync(path.join(scratch,'src/compiler',name),execFileSync('git',['-C',tree,'show','refs/stage3/47-host-errors-after:src/compiler/'+name]));
const file=path.join(scratch,'src/compiler/tracing.ts'),completed=fs.readFileSync(file,'utf8');
function owner(text){const f=ts.createSourceFile('tracing.ts',text,ts.ScriptTarget.Latest,true);const found=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='startTracing')found.push(n);ts.forEachChild(n,visit);}visit(f);assert.equal(found.length,1);return {text:found[0].getText(f),start:found[0].getStart(f),end:found[0].end};}
const node=owner(completed),normal=s=>s.replace(/\r\n/g,'\n');assert.equal(normal(node.text),pins.after);
const withOwner=text=>completed.slice(0,node.start)+text.replace(/\n/g,'\r\n')+completed.slice(node.end);
const before=withOwner(pins.before);fs.writeFileSync(file,before);
const first=adapt(scratch,ts);assert.equal(first.evidence.filter(e=>e.after.includes('meta.ph')).length,3);
assert.equal(fs.readFileSync(file,'utf8'),completed);
const mtime=fs.statSync(file).mtimeMs;assert.equal(adapt(scratch,ts).files,0);assert.equal(fs.statSync(file).mtimeMs,mtime);
let oldCompleted=pins.after;for(const item of pins.literals)oldCompleted=oldCompleted.replace(item.after,item.before);
const mutants=[['mixed catch and original literals',oldCompleted],['mixed catch and explicit literals',pins.after.replace('errorMessage(e)','e.message')],['mixed metadata literals',pins.after.replace(pins.literals[0].after,pins.literals[0].before)],['meta value drift',pins.after.replace('1000 * timestamp()','2000 * timestamp()')],['loader producer drift',pins.after.replace('require("fs")','require("path")')],['unrelated function drift',pins.after.replace('typeCatalog.length = 0','typeCatalog.length = 1')]];
for(const [label,text] of mutants){assert.notEqual(text,pins.after);const changed=withOwner(text);fs.writeFileSync(file,changed);assert.throws(()=>adapt(scratch,ts),/unreviewed loader producer: startTracing drift/,label);assert.equal(fs.readFileSync(file,'utf8'),changed);}
fs.writeFileSync(file,completed);
fs.writeFileSync(output,JSON.stringify({exactIncomingAccepted:true,exactCompletedAccepted:true,incomingMetadataEdits:3,completedReplayEdits:0,replayWithoutWrites:true,driftRefusedBeforeWriting:mutants.map(([label])=>label)},null,2)+'\n');
console.log('two exact function states accepted; six other states refuse before writing');
