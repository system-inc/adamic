'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),assert=require('node:assert/strict'),ts=require('typescript');
const {nodes}=require('../../47-host-errors/adapt.cjs');
const [before,after,output]=process.argv.slice(2).map(p=>path.resolve(p));
const compiler=require(path.join(before,'built/local/typescript.js'));
const calls=[];
const sys={...compiler.sys,write(){},exit(code){calls.push(code);return 1;}};
const closer=compiler.executeCommandLine(sys,()=>{},['--invalid-adamic-option']);
assert.equal(closer,1);assert.deepEqual(calls,[1]);
const relative='src/testRunner/unittests/tscWatch/incremental.ts';
function guard(tree){
 const source=ts.createSourceFile(relative,fs.readFileSync(path.join(tree,relative),'utf8'),ts.ScriptTarget.Latest,true);
 const matches=nodes(ts,source,n=>ts.isIfStatement(n)&&n.thenStatement.getText(source)==='closer.close();');
 assert.equal(matches.length,1);return matches[0].getText(source);
}
const oldGuard=guard(before),newGuard=guard(after);
let oldFailure;
try{vm.runInNewContext(oldGuard,{closer});assert.fail('old guard accepted exit code');}catch(error){assert.match(String(error),/closer.close is not a function/);oldFailure=String(error);}
const watchFile='src/testRunner/unittests/helpers/tscWatch.ts';
const watchSource=ts.createSourceFile(watchFile,fs.readFileSync(path.join(after,watchFile),'utf8'),ts.ScriptTarget.Latest,true);
const properties=nodes(ts,watchSource,n=>ts.isPropertyAssignment(n)&&n.name.getText(watchSource)==='watchOrSolution');
assert.equal(properties.length,1);
const select=properties[0].initializer.getText(watchSource);
for(const value of [0,1,2,4,undefined]) assert.equal(vm.runInNewContext(select,{watchOrSolution:value}),undefined);
const watchObject={close(){}};assert.equal(vm.runInNewContext(select,{watchOrSolution:watchObject}),watchObject);
assert.notEqual(vm.runInNewContext('watchOrSolution',{watchOrSolution:1}),undefined,'dropped selection mutant survived');
const observations=[];
for(const value of [0,1,2,4,undefined]){vm.runInNewContext(newGuard,{closer:value});observations.push(String(value));}
let closed=0;vm.runInNewContext(newGuard,{closer:{close(){closed++;}}});assert.equal(closed,1);
// Both dedicated stubs retain noop's undefined result, for every observed code.
const stubs=[],stubMutants=[];
for(const relative of ['src/harness/harnessLanguageService.ts','src/testRunner/unittests/tsserver/session.ts']){
 const source=ts.createSourceFile(relative,fs.readFileSync(path.join(after,relative),'utf8'),ts.ScriptTarget.Latest,true);
 const members=nodes(ts,source,n=>(ts.isPropertyDeclaration(n)||ts.isPropertyAssignment(n))&&n.name.getText(source)==='exit');
 assert.equal(members.length,1);
 const javascript=ts.transpileModule('const stub = '+members[0].initializer.getText(source)+'; stub;',{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
 const stub=vm.runInNewContext(javascript);
 for(const code of [undefined,0,1,2,4]) assert.equal(stub(code),undefined);
 const mutant=vm.runInNewContext(javascript.replace('=> undefined','=> _exitCode'));
 assert.notEqual(mutant(1),undefined,'return-code stub mutant survived');
 stubMutants.push({file:relative,mutation:'return _exitCode',caught:true});
 stubs.push(relative);
}
const report={selectionGuard:select,selectionMutantCaught:true,status:'pass',node:process.version,args:['--invalid-adamic-option'],exitArguments:calls,returned:closer,oldGuard,oldFailure,newGuard,numericAndUndefinedInputs:observations,objectClosed:closed,stubs,stubMutants,mutant:{name:'restore-truthiness-close',caught:true,error:oldFailure}};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
