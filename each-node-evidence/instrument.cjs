'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const [input,output]=process.argv.slice(2);assert(!fs.existsSync(output));fs.mkdirSync(output,{recursive:true});
for(const e of fs.readdirSync(input,{withFileTypes:true}))if(!['built','node_modules'].includes(e.name))fs.symlinkSync(path.join(input,e.name),path.join(output,e.name));
fs.mkdirSync(path.join(output,'built/local'),{recursive:true});
for(const e of fs.readdirSync(path.join(input,'built')))if(e!=='local')fs.symlinkSync(path.join(input,'built',e),path.join(output,'built',e));
const rows=[];
function walk(src,dst){for(const e of fs.readdirSync(src,{withFileTypes:true})){const a=path.join(src,e.name),b=path.join(dst,e.name);if(e.isDirectory()){fs.mkdirSync(b);walk(a,b);continue;}if(!e.name.endsWith('.js')){fs.symlinkSync(a,b);continue;}let text=fs.readFileSync(a,'utf8');const sf=ts.createSourceFile(a,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),edits=[];
function visit(n){if(ts.isFunctionDeclaration(n)&&/^assertEachNode\d*$/.test(n.name?.text||'')){assert.deepEqual(n.parameters.slice(0,2).map(p=>p.name.text),['nodes','test']);edits.push(n.body.getStart(sf)+1);}ts.forEachChild(n,visit);}visit(sf);
if(!edits.length){fs.copyFileSync(a,b);continue;}for(const start of edits.sort((a,b)=>b-a))text=text.slice(0,start)+'\n test = globalThis.__eachNodeObserve(test, '+JSON.stringify(path.relative(input,a))+');\n'+text.slice(start);fs.writeFileSync(b,text);rows.push({file:path.relative(input,a),sites:edits.length});}}
walk(path.join(input,'built/local'),path.join(output,'built/local'));for(const file of ['_tsc.js','typescript.js','run.js'])assert(rows.some(r=>r.file==='built/local/'+file));fs.writeFileSync(output+'/instrumentation.json',JSON.stringify(rows,null,2)+'\n');console.log(JSON.stringify(rows));
