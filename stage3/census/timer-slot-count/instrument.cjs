// Only run on a disposable tree AFTER its static census. Replaces callees, not arguments.
const fs=require('node:fs');
const path=require('node:path');
const [root,rawFile]=process.argv.slice(2);
const raw=JSON.parse(fs.readFileSync(rawFile,'utf8'));
const edits=new Map();
function add(row,replacement){const xs=edits.get(row.file)||[];xs.push([row.calleeStart,row.calleeEnd,replacement]);edits.set(row.file,xs);}
for(const r of raw.accesses)if(r.kind==='call'&&!r.file.endsWith('/parallel/worker.ts'))add(r,`((globalThis as any).__timerSlotTap(${r.receiver}, ${JSON.stringify(r.name)}, ${JSON.stringify(r.file+':'+r.line)}))`);
for(const r of raw.bareCalls)if(r.file==='src/compiler/sys.ts')add(r,`((globalThis as any).__timerSlotTapFunction(${r.name}, ${JSON.stringify(r.file+':'+r.line)}))`);
for(const [file,xs] of edits){const p=path.join(root,file);let s=fs.readFileSync(p,'utf8');for(const [a,b,text]of xs.sort((a,b)=>b[0]-a[0]))s=s.slice(0,a)+text+s.slice(b);fs.writeFileSync(p,s);}
console.log(JSON.stringify({files:edits.size,sites:[...edits.values()].reduce((n,x)=>n+x.length,0)}));
