const ts = require('/workspace/adamic/stage3/api/node_modules/typescript');
const fs = require('fs');
const inventory = JSON.parse(fs.readFileSync('/tmp/void-split-evidence/typescript-impact-ad029.json','utf8'));
const files=new Map(), patterns=new Map();
for(const hit of inventory.hits){
 const filename='/tmp/void-typescript/'+hit.file;
 let file=files.get(filename);
 if(!file){file=ts.createSourceFile(filename,fs.readFileSync(filename,'utf8'),ts.ScriptTarget.Latest,true); files.set(filename,file);}
 const start=file.getPositionOfLineAndCharacter(hit.line-1,hit.column-1);
 let node;
 function find(n){if(n.getStart(file)===start && n.getText(file).slice(0,180)===hit.text) node=n; ts.forEachChild(n,find);}
 find(file);
 if(!node) throw new Error(JSON.stringify(hit));
 let site=node;
 while(site.parent && ts.isParenthesizedExpression(site.parent)) site=site.parent;
 let call=site.parent;
 while(call && !ts.isCallExpression(call) && !ts.isReturnStatement(call) && !ts.isFunctionLike(call)) call=call.parent;
 let pattern;
 if(hit.source==='void') pattern='void return into '+hit.target;
 else if(call && ts.isCallExpression(call)) pattern='void callback passed to '+call.expression.getText(file);
 else if(hit.file==='src/compiler/watch.ts' && hit.text==='noop') pattern='void fallback assigned to an any-returning timer field';
 else pattern='other callable relation';
 const row={...hit,pattern,lineText:file.text.split(/\r?\n/)[hit.line-1].trim()};
 if(!patterns.has(pattern)) patterns.set(pattern,[]);
 patterns.get(pattern).push(row);
}
const result={...inventory,patterns:[...patterns].map(([pattern,sites])=>({pattern,count:sites.length,examples:sites.filter((x,i)=>sites.findIndex(y=>y.file===x.file&&y.line===x.line)===i).slice(0,2),sites}))};
fs.writeFileSync('/tmp/void-split-evidence/tsc-patterns.json',JSON.stringify(result,null,2)+'\n');
for(const group of result.patterns){console.log(group.count+' '+group.pattern);for(const x of group.examples)console.log('  '+x.file+':'+x.line+' '+x.lineText);}
