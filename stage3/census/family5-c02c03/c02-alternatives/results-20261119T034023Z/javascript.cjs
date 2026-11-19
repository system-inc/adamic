const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');const ts=require('/tmp/family5-land/stage3/api/node_modules/typescript');
const before=process.argv[2],after=process.argv[3],output=process.argv[4],sha=s=>crypto.createHash('sha256').update(s).digest('hex');
const files=[];function walk(d){for(const ent of fs.readdirSync(d,{withFileTypes:true})){const p=path.join(d,ent.name);if(ent.isDirectory())walk(p);else if(ent.isFile())files.push(path.relative(before,p));}}walk(path.join(before,'src'));
const records=[];let diffs=0;
for(const rel of files){const a=fs.readFileSync(path.join(before,rel),'utf8'),b=fs.readFileSync(path.join(after,rel),'utf8');if(a===b)continue;
 if(!rel.endsWith('.ts'))throw Error('unexpected non-TS edit '+rel);
 for(const removeComments of [false,true]){const opt={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};const x=ts.transpileModule(a,{compilerOptions:opt,fileName:rel}).outputText,y=ts.transpileModule(b,{compilerOptions:opt,fileName:rel}).outputText;let first=null;if(x!==y){const xx=Buffer.from(x),yy=Buffer.from(y);for(let i=0;i<Math.max(xx.length,yy.length);i++)if(xx[i]!==yy[i]){first=i;break;}diffs++;}
 records.push({file:rel,removeComments,before_sha256:sha(x),after_sha256:sha(y),before_bytes:Buffer.byteLength(x),after_bytes:Buffer.byteLength(y),byte_identical:x===y,first_difference_offset:first});
 }}
fs.writeFileSync(output,JSON.stringify({version:ts.version,source_files_checked:files.length,changed_sources:[...new Set(records.map(x=>x.file))],javascript_differences:diffs,records},null,2)+'\n');console.log('JavaScript differences='+diffs+' (comments kept and removed)');process.exitCode=diffs?1:0;
