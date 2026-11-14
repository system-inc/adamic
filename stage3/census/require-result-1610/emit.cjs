const fs=require('fs'),path=require('path'),ts=require('typescript'),assert=require('assert/strict'),crypto=require('crypto');
const s=__dirname,tree=path.join(s,'main-adapted'),entries=[];
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function walk(d){for(const e of fs.readdirSync(d,{withFileTypes:true})){const f=path.join(d,e.name);if(e.isDirectory())walk(f);else if(e.name.endsWith('.ts')&&!e.name.endsWith('.d.ts'))entries.push(f);}}
walk(path.join(tree,'src'));entries.sort();const evidence=[];
for(const file of entries){const name=path.relative(tree,file),a=fs.readFileSync(file,'utf8'),before=name==='src/compiler/sys.ts'?fs.readFileSync(path.join(s,'baseline-sys.ts'),'utf8'):a,after=name==='src/compiler/sys.ts'?fs.readFileSync(path.join(s,'candidate-sys.ts'),'utf8'):a;
 for(const removeComments of [false,true]){const emit=text=>ts.transpileModule(text,{fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;const x=emit(before),y=emit(after);assert.equal(y,x,name);evidence.push({file:name,removeComments,before:hash(x),after:hash(y)});}
}
fs.writeFileSync(path.join(s,'javascript.json'),JSON.stringify({typescript:ts.version,files:entries.length,emits:evidence.length,byteIdentical:true,inventory:evidence},null,2)+'\n');console.log(JSON.stringify({files:entries.length,emits:evidence.length,byteIdentical:true}));
