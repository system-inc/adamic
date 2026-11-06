const fs=require('fs'),path=require('path'),ts=require('/workspace/container-fit-base/stage3/api/node_modules/typescript');
const root='/workspace/container-fit-adapted';
function files(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(dir,e.name)):e.name.endsWith('.ts')?[path.join(dir,e.name)]:[])}
const rows=files(root+'/src/compiler').sort().map(p=>{const text=fs.readFileSync(p,'utf8'),f=ts.createSourceFile(p,text,ts.ScriptTarget.Latest,true);return {file:path.relative(root,p),details:f.statements.map(n=>({name:n.name?.text||'',generic:!!n.typeParameters?.length})),units:f.statements.map(n=>{const l=f.getLineAndCharacterOfPosition(n.getStart(f));return path.relative(root,p)+':'+(l.line+1)+':'+(l.character+1)})}});
fs.writeFileSync(process.argv[2],JSON.stringify(rows,null,2)+'\n');
