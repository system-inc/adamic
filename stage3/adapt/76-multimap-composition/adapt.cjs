#!/usr/bin/env node
"use strict";
// #pvvzhmy ruling, tracked stage 3 composition adaptation #pcwk8fh.
const fs=require("node:fs"),path=require("node:path"),crypto=require("node:crypto");
if(process.argv.length!==3)throw Error("usage: node adapt.cjs <tree>");
const file=path.join(process.argv[2],"src/compiler/core.ts"),bytes=fs.readFileSync(file,"utf8");
const text=bytes.replace(/\r\n/g,"\n");
const replacement=fs.readFileSync(path.join(__dirname,"replacement.a"),"utf8");
const queue="/** @internal */\nexport function createQueue";
const applied=text.indexOf("class ComposedMultiMap<K, V> implements MultiMap<K, V> {");
const start=applied>=0?applied:text.indexOf("export function createMultiMap<K, V>(): MultiMap<K, V> {");
const end=text.indexOf(queue,start);
if(start<0||end<start)throw Error("MultiMap source drift");
const current=text.slice(start,end);
if(applied>=0){
 if(current!==replacement)throw Error("applied MultiMap source drift");
 console.log(JSON.stringify({files:0,edits:0,alreadyApplied:true}));
}else{
 const digest=crypto.createHash("sha256").update(current).digest("hex");
 if(digest!=="8f1a8a55ae99e94d5b054ac4669ec3b8a6f1adfdf9333ac733c4d02731cfba99")throw Error("unreviewed MultiMap implementation: "+digest);
 let next=text.slice(0,start)+replacement+text.slice(end);
 if(bytes.includes("\r\n"))next=next.replace(/\n/g,"\r\n");
 fs.writeFileSync(file,next);
 console.log(JSON.stringify({files:1,edits:1,adaptation:"MultiMap composition with explicit receiver helpers"}));
}
