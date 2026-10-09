// Inventory the exact diagnostic spans with TypeScript's parser, without editing source.
const fs=require('fs'),path=require('path'),zlib=require('zlib');
const ts=require(process.env.TYPESCRIPT_API);
const dir=__dirname, tree=process.argv[2];
const log=zlib.gunzipSync(fs.readFileSync(path.join(dir,'evidence/baseline.log.gz'))).toString();
const re=/^(.+?):(\d+):(\d+): error TS(\d+): ([\s\S]*?)(?=^.+?:\d+:\d+: error TS\d+: |$(?![\s\S]))/gm;
const cache=new Map(),rows=[];
const crypto=require('crypto');
for(const m of log.matchAll(re)){
 const file=m[1].split('/src/compiler/')[1],line=+m[2],column=+m[3],code=+m[4],message=m[5].trim();
 if(!cache.has(file)){const text=fs.readFileSync(path.join(tree,'src/compiler',file),'utf8');cache.set(file,ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true));}
 const sf=cache.get(file),starts=sf.getLineStarts(); if(line>starts.length)throw Error(`source mismatch ${file}:${line}, only ${starts.length} lines`); const pos=starts[line-1]+column-1;let node=sf;
 function visit(n){if(n.getStart(sf)<=pos&&pos<n.end){node=n;ts.forEachChild(n,visit);}}visit(sf);
 let expr=node;while(expr.parent&&!ts.isStatement(expr.parent)&&!ts.isSourceFile(expr.parent)&&!ts.isBlock(expr.parent))expr=expr.parent;
 let group;
 if([2375,2379,2412,2420].includes(code))group='optional-field-presence';
 else if(code===18046)group='caught-unknown';
 else if([2304,2307,2503,2591].includes(code))group='host-declarations';
 else if(code===2339)group='host-members';
 else if([7006,7031].includes(code))group='host-inference';
 else if(code===2740)group='library-set-shape';
 else if(code===2556)group='spread-tuple';
 else if(code===2488)group='nullable-iteration';
 else if(code===2722)group='optional-call';
 else if(code===2538)group='nullable-index-key';
 else if(code===2769)group='overload-use';
 else if(code===2345)group='nullable-argument';
 else if(code===2322)group='nullable-result-relation';
 else group='nullable-dereference';
 let indexedExpression=null; for(let parent=node;parent&&!ts.isStatement(parent);parent=parent.parent){if(ts.isElementAccessExpression(parent)){indexedExpression=parent.getText(sf);break;}}
 rows.push({siteText:node.getText(sf),indexedExpression,id:`${file}:${line}:${column}:TS${code}`,file,line,column,code,group,message,syntax:ts.SyntaxKind[node.kind],expression:expr.getText(sf).slice(0,2000),sourceLine:sf.text.split('\n')[line-1]});
}
if(rows.length!==320)throw Error(`expected 320, got ${rows.length}`);
fs.writeFileSync(path.join(dir,'diagnostics.json'),JSON.stringify(rows,null,2)+'\n');
fs.writeFileSync(path.join(dir,'source-hashes.json'),JSON.stringify(Object.fromEntries([...cache].map(([name,sf])=>[name,crypto.createHash('sha256').update(sf.text).digest('hex')])),null,2)+'\n');
console.log(JSON.stringify(Object.fromEntries([...new Set(rows.map(r=>r.group))].map(g=>[g,rows.filter(r=>r.group===g).length])),null,2));
