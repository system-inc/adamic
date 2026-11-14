'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const p=__dirname,tag=process.argv[2],oldtag=process.argv[3]||'main';
const files=['compiler/sys.ts','compiler/watchPublic.ts','compiler/tsbuildPublic.ts','server/utilities.ts','server/session.ts','server/typingInstallerAdapter.ts'];
let sites=0;
for(const file of files){
 const before=fs.readFileSync(path.join(p,'sources.'+oldtag,file),'utf8').replace(/\r\n/g,'\n');
 const after=fs.readFileSync(path.join(p,'sources.'+tag,file),'utf8').replace(/\r\n/g,'\n');
 if(tag==='1'){
  sites+=(before.match(/setTimeout!?\.call\(timerHost, /g)||[]).length;
  const expected=before.replace(/^([ \t]*)const \{ setTimeout \} = timerHost;\n/gm,'').replace(/setTimeout(!?)\.call\(timerHost, /g,'timerHost.setTimeout$1(');
  assert.equal(after,expected,'only receiver rewrite: '+file);
 }
 for(const [name,text]of [[oldtag,before],[tag,after]]){
  const dest=path.join(p,'javascript.'+name,file.replace(/\.ts$/,'.js'));fs.mkdirSync(path.dirname(dest),{recursive:true});fs.writeFileSync(dest,ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText);
 }
}
if(tag==='1')assert.equal(sites,9);
console.log(JSON.stringify({from:oldtag,to:tag,files:files.length,receiverSites:sites,onlyReceiverRewrite:tag==='1',target:'ES2024',module:'ESNext'}));
