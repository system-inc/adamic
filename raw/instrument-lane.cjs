'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('/workspace/adamic/stage3/api/node_modules/typescript');
const [tree, output] = process.argv.slice(2), rows = [];
function walk(dir) {
 for (const item of fs.readdirSync(dir, {withFileTypes:true})) {
  const file = path.join(dir,item.name);
  if(item.isDirectory()){walk(file);continue;}
  if(!file.endsWith('.js'))continue;
  const text=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS), edits=[];
  assert(!text.includes('__metaSpreadRun'));
  function visit(n){
   if(ts.isFunctionDeclaration(n)&&n.name && /^startTracing[0-9]*$/.test(n.name.text)&&n.body&&n.getText(source).includes('TracingStartedInBrowser')) {
    const meta=[];
    function find(v){if(ts.isVariableDeclaration(v)&&ts.isIdentifier(v.name)&&v.name.text==='meta')meta.push(v);ts.forEachChild(v,find);}
    find(n);assert.equal(meta.length,1);assert(ts.isVariableStatement(meta[0].parent.parent));
    const log = value => 'if (process.env.META_SPREAD_COUNTER) require("node:fs").appendFileSync(process.env.META_SPREAD_COUNTER, JSON.stringify('+value+') + "\\n");';
    edits.push({at:n.body.getStart(source)+1,text:'\nconst __metaSpreadRun = process.pid + ":" + (globalThis.__metaSpreadSeq = (globalThis.__metaSpreadSeq || 0) + 1);\n'+log('{event:"start",run:__metaSpreadRun,file:__filename,pid:process.pid}')+'\n'});
    edits.push({at:meta[0].parent.parent.end,text:'\n'+log('{event:"meta",run:__metaSpreadRun,file:__filename,pid:process.pid,keys:Object.keys(meta),ownKeys:Reflect.ownKeys(meta).map(k=>typeof k === "symbol" ? String(k) : k),enumerableKeys:Reflect.ownKeys(meta).filter(k=>Object.getOwnPropertyDescriptor(meta,k).enumerable).map(String),values:meta}')+'\n'});
   }
   ts.forEachChild(n,visit);
  }
  visit(source);if(!edits.length)continue;
  let changed=text;for(const edit of edits.sort((a,b)=>b.at-a.at))changed=changed.slice(0,edit.at)+edit.text+changed.slice(edit.at);
  const loaded='if (process.env.META_SPREAD_COUNTER) require("node:fs").appendFileSync(process.env.META_SPREAD_COUNTER, JSON.stringify({event:"load",file:__filename,pid:process.pid}) + "\\n");\n';
  const at=changed.startsWith('#!')?changed.indexOf('\n')+1:0;
  changed=changed.slice(0,at)+loaded+changed.slice(at);
  if(fs.lstatSync(file).isSymbolicLink())fs.unlinkSync(file);fs.writeFileSync(file,changed);
  rows.push({file:path.relative(tree,file),startCounters:edits.length/2,metaObservers:edits.length/2});
 }
}
walk(path.join(tree,'built'));
for(const name of ['run.js','typescript.js','_tsc.js'])assert(rows.some(row=>row.file==='built/local/'+name));
fs.writeFileSync(output,JSON.stringify({privateTree:tree,rows},null,2)+'\n');console.log(rows);
