const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript'),crypto=require('node:crypto');
const [before,after,out]=process.argv.slice(2),rows=[];
for(const name of ts.sys.readDirectory(before+'/src',['.ts']).filter(f=>!f.endsWith('.d.ts'))){
 const file=path.relative(before,name),a=fs.readFileSync(name,'utf8'),b=fs.readFileSync(path.join(after,file),'utf8');
 for(const removeComments of [false,true]){
  const opts={fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true,removeComments}};
  const x=ts.transpileModule(a,opts).outputText,y=ts.transpileModule(b,opts).outputText;
  assert.equal(x,y,file+': emitted JS differs');
  rows.push({file,removeComments,sha256:crypto.createHash('sha256').update(x).digest('hex')});
 }
}
fs.writeFileSync(out,JSON.stringify({files:rows.length/2,emits:rows.length,rows},null,2)+'\n');console.log(rows.length);
