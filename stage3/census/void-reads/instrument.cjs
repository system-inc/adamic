'use strict';
// Only run on the disposable adapted checkout after the mutant has restored it.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const [tree,table]=process.argv.slice(2),data=JSON.parse(fs.readFileSync(table,'utf8'));
for(const root of data.roots){
 const file=path.join(tree,root.file),text=fs.readFileSync(file,'utf8');
 if(crypto.createHash('sha256').update(text).digest('hex')!==root.sha256)throw Error('source drift '+file);
 const selected=data.sites.map((site,id)=>({...site,id})).filter(s=>s.file===root.file&&s.form!=='concise arrow body');
 if(!selected.length)continue;
 const events=new Map();
 for(const site of selected){
  if(!events.has(site.start))events.set(site.start,{open:[],close:[]});
  if(!events.has(site.end))events.set(site.end,{open:[],close:[]});
  events.get(site.start).open.push(site);events.get(site.end).close.push(site);
 }
 let output='',last=0;
 for(const [pos,event] of [...events].sort((a,b)=>a[0]-b[0])){
  output+=text.slice(last,pos)+')'.repeat(event.close.length);
  for(const site of event.open.sort((a,b)=>b.end-a.end))output+='globalThis.__adamicVoidRead('+site.id+', ';
  last=pos;
 }
 output+=text.slice(last);
 output='declare global { var __adamicVoidRead: <T>(id: number, value: T) => T; }\n'+output;
 fs.writeFileSync(file,output);
}
console.log('instrumented '+data.sites.filter(s=>s.form!=='concise arrow body').length+' sites');
