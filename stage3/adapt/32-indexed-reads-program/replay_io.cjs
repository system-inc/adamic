'use strict';
// Observational filesystem journal. Adapter source and write behavior are unchanged.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const root=process.env.BIND_REPLAY_ROOT,log=process.env.BIND_REPLAY_JOURNAL;
if(root&&log){
 const read=fs.readFileSync.bind(fs),stat=fs.statSync.bind(fs),append=fs.appendFileSync.bind(fs);
 const sha=data=>crypto.createHash('sha256').update(data).digest('hex');
 function name(value){
  if(typeof value==='number')return undefined;
  try {const p=path.resolve(value instanceof URL?require('node:url').fileURLToPath(value):String(value));const rel=path.relative(root,p);return rel&&!rel.startsWith('..')&&!path.isAbsolute(rel)&&rel!=='.git'&&!rel.startsWith('.git/')?p:undefined;}catch{return undefined;}
 }
 function state(file){
  try {if(!stat(file).isFile())return null;const data=read(file);return {sha256:sha(data),data};}catch{return null;}
 }
 function capture(args,indices){return indices.map(i=>name(args[i])).filter(Boolean).map(file=>({file,before:state(file)}));}
 function record(rows){for(const row of rows){
  const after=state(row.file),changed=row.before?.sha256!==after?.sha256;
  const event={file:path.relative(root,row.file),before:row.before?.sha256??null,after:after?.sha256??null};
  if(changed){event.beforeData=row.before?row.before.data.toString('base64'):null;event.afterData=after?after.data.toString('base64'):null;}
  append(log,JSON.stringify(event)+'\n');
 }}
 for(const [method,indices] of [['writeFileSync',[0]],['appendFileSync',[0]],['copyFileSync',[1]],['renameSync',[0,1]],['unlinkSync',[0]],['truncateSync',[0]]]){
  const original=fs[method];fs[method]=function(...args){const rows=capture(args,indices);const result=Reflect.apply(original,this,args);record(rows);return result;};
 }
 for(const [method,indices] of [['writeFile',[0]],['appendFile',[0]],['copyFile',[1]],['rename',[0,1]],['unlink',[0]],['truncate',[0]]]){
  const original=fs[method];fs[method]=function(...args){
   const rows=capture(args,indices),last=args.length-1,callback=args[last];
   if(typeof callback!=='function')return Reflect.apply(original,this,args);
   args[last]=function(error,...rest){if(!error)record(rows);return Reflect.apply(callback,this,[error,...rest]);};
   return Reflect.apply(original,this,args);
  };
  const promise=fs.promises[method];if(promise)fs.promises[method]=async function(...args){const rows=capture(args,indices);const result=await Reflect.apply(promise,this,args);record(rows);return result;};
 }
}
