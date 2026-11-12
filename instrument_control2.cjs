'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const old='/workspace/scratch/tracing-control2',target='/workspace/scratch/tracing-control2-counted',out='/workspace/scratch/tracing-fs-import-evidence';
fs.mkdirSync(target);for(const name of fs.readdirSync(old)){if(['built','node_modules'].includes(name))continue;fs.symlinkSync(path.join(old,name),path.join(target,name));}
fs.mkdirSync(path.join(target,'built/local'),{recursive:true});for(const name of fs.readdirSync(path.join(old,'built'))){if(name!=='local')fs.symlinkSync(path.join(old,'built',name),path.join(target,'built',name));}
const rows=[];
for(const name of fs.readdirSync(path.join(old,'built/local'))){const from=path.join(old,'built/local',name),to=path.join(target,'built/local',name);if(!['_tsc.js','run.js','typescript.js'].includes(name)){fs.symlinkSync(from,to);continue;}
 const text=fs.readFileSync(from,'utf8'),source=ts.createSourceFile(from,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),found=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text.startsWith('startTracing'))found.push(n);ts.forEachChild(n,visit);}visit(source);assert.equal(found.length,1);const catches=[];function scan(n){if(ts.isCatchClause(n))catches.push(n);ts.forEachChild(n,scan);}scan(found[0]);assert.equal(catches.length,1);
 const payload=kind=>`\nif (process.env.TRACING_CATCH_EVENTS) require("node:fs").appendFileSync(process.env.TRACING_CATCH_EVENTS, JSON.stringify({kind:${JSON.stringify(kind)},bundle:${JSON.stringify(name)},pid:process.pid})+"\\n");\n`;
 const load=payload('load'),counter=payload('catch'),at=catches[0].block.getStart(source)+1;let changed=text.slice(0,at)+counter+text.slice(at);const insert=source.statements[0].end;changed=changed.slice(0,insert)+load+changed.slice(insert);assert.equal(changed.replace(load,'').replace(counter,''),text);fs.writeFileSync(to,changed);rows.push({file:name,beforeSha256:crypto.createHash('sha256').update(text).digest('hex'),afterSha256:crypto.createHash('sha256').update(changed).digest('hex'),catchInsertion:at,loadInsertion:insert,undoRestoresExactBytes:true});
}
fs.writeFileSync(path.join(out,'catch2-instrumentation.json'),JSON.stringify(rows,null,2)+'\n');
