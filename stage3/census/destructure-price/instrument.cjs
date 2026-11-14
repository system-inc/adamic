const ts=require('/tmp/destructure-price-scratch/work/stage3/api/node_modules/typescript'),fs=require('fs'),path=require('path');
const base='/tmp/destructure-price-scratch/adapted',dest='/workspace/scratch/destructure-price-f2989762/counting';
const sites=JSON.parse(fs.readFileSync('/tmp/destructure-price-scratch/ast-sites.json','utf8'));
const files=new Map();
for(const site of sites){if(!files.has(site.file))files.set(site.file,[]);files.get(site.file).push(site);}
for(const [file,group] of files){const text=fs.readFileSync(path.join(base,file),'utf8'),sf=ts.createSourceFile(file,text,99,true),edits=[];
 function walk(n){if(ts.isObjectBindingPattern(n)){
 const line=sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1,site=group.find(s=>s.line===line);
 if(site){const id=file+':'+line,init=n.parent.initializer;const methodNames=JSON.stringify(site.members.filter(m=>m.method).map(m=>m.name)); const wrap=expression=>`(<T,>(value:T):T => (globalThis as any).__dpWrap(value, ${JSON.stringify(id)}, ${methodNames}))(${expression})`; let parameterPrelude=''; if(!init){if(!ts.isParameter(n.parent))throw Error('unexpected pattern');const arg='__dpArg_'+line;edits.push({start:n.getStart(sf),end:n.end,text:arg});parameterPrelude=`const ${n.getText(sf)} = ${wrap(arg)};\n`;}
 if(init)edits.push({start:init.getStart(sf),end:init.end,text:wrap(init.getText(sf))});
 let enclosing=n.parent;while(enclosing&&!ts.isFunctionLike(enclosing))enclosing=enclosing.parent;
 if(enclosing){if(!enclosing.body||!ts.isBlock(enclosing.body))throw Error('unexpected body');edits.push({start:enclosing.body.getStart(sf)+1,end:enclosing.body.getStart(sf)+1,text:`\n(globalThis as any).__dpEnter(${JSON.stringify(id)});\n${parameterPrelude}`});}
 else {let statement=n;while(statement.parent&&!ts.isVariableStatement(statement))statement=statement.parent;if(!ts.isVariableStatement(statement))throw Error('no statement');edits.push({start:statement.getStart(sf),end:statement.getStart(sf),text:`(globalThis as any).__dpEnter(${JSON.stringify(id)});\n`});}
 }}ts.forEachChild(n,walk);}walk(sf);
 let changed=text;for(const e of edits.sort((a,b)=>b.start-a.start))changed=changed.slice(0,e.start)+e.text+changed.slice(e.end);
 fs.writeFileSync(path.join(dest,file),changed);
}
console.log(JSON.stringify({sites:sites.length,files:files.size}));
