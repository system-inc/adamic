const fs=require('fs'),cp=require('child_process'),path=require('path');
const ts=require(process.env.TYPESCRIPT_API ? path.resolve(process.env.TYPESCRIPT_API) : path.resolve(__dirname, '../../api/node_modules/typescript'));
const tree=process.argv[2],out=process.argv[3];
const files=cp.execFileSync('git',['-C',tree,'ls-files','-z'],{maxBuffer:20e6}).toString().split('\0').filter(f=>/\.[cm]?[jt]sx?$/.test(f)&&!f.startsWith('src/'));
const names=new Set(['arrayFrom','arrayToMap','reduceLeft','sameMap']),candidates=[];let nodes=0,calls=0;
for(const file of files){
 const sf=ts.createSourceFile(file,fs.readFileSync(path.join(tree,file),'utf8'),ts.ScriptTarget.Latest,true),stack=[sf];let found=false;
 while(stack.length){const n=stack.pop();nodes++;if(ts.isIdentifier(n)&&names.has(n.text))found=true;if(ts.isCallExpression(n)||ts.isNewExpression(n)||ts.isTaggedTemplateExpression(n))calls++;ts.forEachChild(n,c=>{stack.push(c);});}
 if(found)candidates.push(file);
}
fs.writeFileSync(out,JSON.stringify({files:files.length,nodes,calls,candidates},null,2)+'\n');console.log(JSON.stringify({files:files.length,nodes,calls,candidates:candidates.length}));
