'use strict';
const fs=require('fs'),path=require('path'),ts=require('../../api/node_modules/typescript');
const [oldTree,newTree]=process.argv.slice(2),output={};
function declarations(file,text){const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),found=new Map();function visit(n,parents){let next=parents;if(ts.isFunctionDeclaration(n)&&n.name&&n.body){const key=[...parents,n.name.text].join('/'),pos=source.getLineAndCharacterOfPosition(n.getStart(source)),site=`${file}:${pos.line+1}:${pos.character+1}`;found.set(key,[...(found.get(key)||[]),site]);next=[...parents,n.name.text];}ts.forEachChild(n,c=>visit(c,next));}visit(source,[]);return found;}
for(const rel of ['src/compiler/utilities.ts','src/compiler/builder.ts']){const old=fs.readFileSync(path.join(oldTree,rel),'utf8'),now=fs.readFileSync(path.join(newTree,rel),'utf8');if(old===now)continue;const a=declarations(rel,old),b=declarations(rel,now);for(const [key,sites]of b){if(sites.length===1&&a.get(key)?.length===1)output[sites[0]]=a.get(key)[0];}}
console.log(JSON.stringify(output,null,2));
