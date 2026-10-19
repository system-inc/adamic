'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const [before,after,output]=process.argv.slice(2).map(p=>path.resolve(p));
const compiler=['src/compiler/executeCommandLine.ts','src/compiler/sys.ts'];
const runtime=['src/harness/harnessLanguageService.ts','src/testRunner/unittests/helpers/tscWatch.ts','src/testRunner/unittests/tscWatch/incremental.ts','src/testRunner/unittests/tsserver/session.ts'];
const annotations=['src/harness/fakesHosts.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts'];
const expected=[...compiler,...runtime,...annotations].sort();
function files(root,relative='src'){
 return fs.readdirSync(path.join(root,relative),{withFileTypes:true}).flatMap(entry=>entry.isDirectory()?files(root,path.join(relative,entry.name)):[path.join(relative,entry.name)]).sort();
}
const all=files(before);assert.deepEqual(files(after),all);
const changed=all.filter(file=>!fs.readFileSync(path.join(before,file)).equals(fs.readFileSync(path.join(after,file))));
assert.deepEqual(changed,expected);
const outputChanges=[];
for(const file of changed){
 const row={file,identical:[]};
 for(const removeComments of [false,true]){
  const options={fileName:file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  const a=ts.transpileModule(fs.readFileSync(path.join(before,file),'utf8'),options).outputText;
  const b=ts.transpileModule(fs.readFileSync(path.join(after,file),'utf8'),options).outputText;
  const identical=a===b;
  assert.equal(identical,!runtime.includes(file),file);
  row.identical.push({removeComments,byteIdentical:identical});
 }
 outputChanges.push(row);
}
const report={status:'pass',totalSourceFiles:all.length,unchangedSourceFiles:all.length-changed.length,compilerFiles:compiler,harnessRuntimeChanges:runtime,harnessAnnotationOnly:annotations,outputChanges};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
