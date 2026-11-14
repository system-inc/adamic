'use strict';
const fs=require('fs'),path=require('path'),ts=require('typescript'),assert=require('assert/strict');
const [before,after,baseEmit,afterEmit]=process.argv.slice(2),owners=['createSingleWatcherPerName','pollingWatchFile','fsWatch'];
function stripped(text,kind){const sf=ts.createSourceFile('sys.'+kind,text,ts.ScriptTarget.Latest,true,kind==='js'?ts.ScriptKind.JS:ts.ScriptKind.TS),spans=[];function visit(n){if(ts.isFunctionDeclaration(n)&&owners.includes(n.name?.text)){spans.push([n.getStart(sf),n.end]);return;}ts.forEachChild(n,visit);}visit(sf);assert.equal(spans.length,3);for(const [a,b] of spans.sort((a,b)=>b[0]-a[0]))text=text.slice(0,a)+'/* reviewed owner */'+text.slice(b);return text.replace(/\r\n/g,'\n');}
const source=root=>fs.readFileSync(path.join(root,'src/compiler/sys.ts'),'utf8');
assert.equal(stripped(source(before),'ts'),stripped(source(after),'ts'));
assert.equal(stripped(fs.readFileSync(baseEmit+'/sys.js','utf8'),'js'),stripped(fs.readFileSync(afterEmit+'/sys.js','utf8'),'js'));
function walk(root,rel='src'){return fs.readdirSync(path.join(root,rel),{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(root,path.join(rel,e.name)):[path.join(rel,e.name)]);}
const files=walk(before);assert.deepEqual(files,walk(after));const changed=files.filter(f=>!fs.readFileSync(path.join(before,f)).equals(fs.readFileSync(path.join(after,f))));assert.deepEqual(changed,['src/compiler/sys.ts']);
console.log(JSON.stringify({files:files.length,changed,sourceAndEmittedChangesOnlyIn:owners}));
