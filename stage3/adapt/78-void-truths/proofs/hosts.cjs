'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),vm=require('node:vm'),ts=require('typescript');
const {nodes}=require('../../47-host-errors/adapt.cjs');
const [tree,output]=process.argv.slice(2).map(x=>path.resolve(x));
function source(file){return ts.createSourceFile(file,fs.readFileSync(path.join(tree,file),'utf8'),ts.ScriptTarget.Latest,true);}
function single(file,predicate){const result=nodes(ts,source(file),predicate);assert.equal(result.length,1);return result[0];}
function evaluate(text,context){return vm.runInNewContext(ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS}}).outputText,context);}
const observations=[];
const nodeExit=single('src/compiler/sys.ts',n=>ts.isMethodDeclaration(n)&&n.name.getText()==='exit');
for(const deferred of [false,true]){
 let continuation,exitCode,returned=false;
 const sentinel=new Error('process exit');
 const exit=evaluate('({'+nodeExit.getText()+'}).exit;',{
  disableCPUProfiler:cb=>{if(deferred){continuation=cb;return true;}cb();return false;},
  process:{exit:code=>{exitCode=code;throw sentinel;}}
 });
 for(const code of [undefined,0,1,2,4,97]){
  try{assert.equal(exit(code),undefined);returned=true;}catch(error){assert.equal(error,sentinel);returned=false;}
  assert.equal(returned,deferred);
  if(deferred)assert.throws(continuation,error=>error===sentinel);
  assert.equal(exitCode,code);
 }
 observations.push({implementation:'nodeSystem.exit',deferred,completedResult:deferred?'undefined':'never',codes:[null,0,1,2,4,97]});
}
for(const file of ['src/harness/fakesHosts.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts']){
 const method=single(file,n=>ts.isMethodDeclaration(n)&&n.name.getText()==='exit');
 const sentinel=new Error('System exit');
 const host=evaluate('({'+method.getText().replace('public ','')+'});',{processExitSentinel:sentinel,exitMessage:'exit',Error});
 for(const code of [undefined,0,1,2,4,97]){
  assert.throws(()=>host.exit(code));assert.equal(host.exitCode,code);
 }
 observations.push({implementation:file,completedResult:'never'});
}
for(const file of ['src/testRunner/unittests/helpers/tsc.ts','src/testRunner/unittests/tsc/cancellationToken.ts','src/testRunner/unittests/tscWatch/incremental.ts','src/testRunner/unittests/tscWatch/listFilesOnly.ts']){
 const assignment=single(file,n=>ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken&&n.left.getText()==='sys.exit');
 const sys={};evaluate(assignment.getText()+';', {sys});
 for(const code of [undefined,0,1,2,4,97]){assert.equal(sys.exit(code),code);assert.equal(sys.exitCode,code);}
 observations.push({implementation:file,completedResult:'number | undefined'});
}
const noop=single('src/compiler/core.ts',n=>ts.isFunctionDeclaration(n)&&n.name?.text==='noop');
const fn=evaluate(noop.getText().replace('export ','')+';noop;',{});
assert.equal(fn(),undefined);assert.equal(fn(97),undefined);
observations.push({implementation:'noop (harness and tsserver host)',completedResult:'undefined'});
fs.writeFileSync(output,JSON.stringify({status:'pass',observations,scope:'Actual extracted implementations on Node; profiler continuation scheduling is supplied explicitly, not an actual inspector session.'},null,2)+'\n');
console.log(JSON.stringify({status:'pass',implementations:observations.length}));
