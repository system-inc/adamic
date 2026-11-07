'use strict';
// Private measurement copies only. Normal build outputs are never modified.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const names=['assertEachNode','assertNode','assertNotNode','assertOptionalNode'];
function patch(text,file){
 const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);const edits=[];const counts=Object.fromEntries(names.map(n=>[n,0]));
 function visit(n){if(ts.isArrowFunction(n)&&!ts.isBlock(n.body)&&n.body.getText(sf).includes('test')&&n.body.getText(sf).includes('Node')){
  let o=n.parent;while(o&&!ts.isFunctionDeclaration(o))o=o.parent;const name=names.find(name=>o?.name?.text===name||new RegExp('^'+name+'[0-9]+$').test(o?.name?.text||''));
  if(name){assert(/pass(?:ed)? test/.test(n.body.getText(sf)),'unexpected callback');assert(!n.body.getText(sf).includes('__debugMessageCount'));edits.push({start:n.body.getStart(sf),end:n.body.end,text:`{ globalThis.__debugMessageCount(${JSON.stringify(name)}); return ${n.body.getText(sf)}; }`});counts[name]++;}
 }ts.forEachChild(n,visit);}visit(sf);
 for(const e of edits.sort((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);return {text,counts};
}
if(require.main===module){const [input,output,record]=process.argv.slice(2);assert(!fs.existsSync(output));fs.mkdirSync(output,{recursive:true});
 for(const entry of fs.readdirSync(input,{withFileTypes:true}))if(entry.name!=='built'&&entry.name!=='node_modules')fs.symlinkSync(path.join(input,entry.name),path.join(output,entry.name),entry.isDirectory()?'dir':'file');
 fs.mkdirSync(path.join(output,'built/local'),{recursive:true});for(const e of fs.readdirSync(path.join(input,'built'),{withFileTypes:true}))if(e.name!=='local')fs.symlinkSync(path.join(input,'built',e.name),path.join(output,'built',e.name),e.isDirectory()?'dir':'file');
 const rows=[];const totals=Object.fromEntries(names.map(n=>[n,0]));
 function walk(from,to){for(const e of fs.readdirSync(from,{withFileTypes:true})){const src=path.join(from,e.name),dst=path.join(to,e.name);if(e.isDirectory()){fs.mkdirSync(dst);walk(src,dst);}else if(e.name.endsWith('.js')){const before=fs.readFileSync(src,'utf8');const result=before.includes('did not pass test')||before.includes('should not have passed test')?patch(before,src):{text:before,counts:{}};if(result.text===before)fs.symlinkSync(src,dst);else{fs.writeFileSync(dst,result.text);rows.push({file:path.relative(input,src),counts:result.counts});for(const n of names)totals[n]+=result.counts[n];}}else fs.symlinkSync(src,dst);}}
 walk(path.join(input,'built/local'),path.join(output,'built/local'));for(const name of names)assert(totals[name]>0,'missing instrumentation '+name);assert(rows.some(r=>r.file==='built/local/run.js'),'harness callback instrumentation missing');fs.writeFileSync(record,JSON.stringify({rows,totals,measurementOnly:true},null,2)+'\n');console.log(JSON.stringify({files:rows.length,totals}));}
module.exports={patch};
