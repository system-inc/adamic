'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const [tree,output]=process.argv.slice(2);const rows=[];
function walk(dir){for(const item of fs.readdirSync(dir,{withFileTypes:true})){const file=path.join(dir,item.name);if(item.isDirectory()){walk(file);continue;}if(!file.endsWith('.js'))continue;
 const text=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);const sites=[];
 function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='startTracing'&&n.body&&n.getText(source).includes('TracingStartedInBrowser'))sites.push(n.body.getStart(source)+1);ts.forEachChild(n,visit);}visit(source);
 if(!sites.length)continue;let changed=text;for(const at of sites.sort((a,b)=>b-a))changed=changed.slice(0,at)+'\nif (process.env.META_SPREAD_COUNTER) require("node:fs").appendFileSync(process.env.META_SPREAD_COUNTER, JSON.stringify({file: __filename, pid: process.pid}) + "\\n");\n'+changed.slice(at);
 if(fs.lstatSync(file).isSymbolicLink())fs.unlinkSync(file);fs.writeFileSync(file,changed);rows.push({file:path.relative(tree,file),counters:sites.length});
}}
walk(path.join(tree,'built'));
assert(rows.some(row=>row.file==='built/local/run.js'),'harness counter missing');
assert(rows.some(row=>row.file==='built/local/typescript.js'),'runtime counter missing');
fs.writeFileSync(output,JSON.stringify({privateTree:tree,rows},null,2)+'\n');console.log(rows);
