'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),vm=require('node:vm');
const ts=require('typescript');
const repo=path.resolve(__dirname,'../../../..');
const served=fs.readFileSync(path.join(__dirname,'head-a-tty.d.ts.txt'),'utf8');
const provenance=require('./head-a.json');
function hash(text){return crypto.createHash('sha256').update(text).digest('hex');}
assert.equal(hash(served),provenance.servedSha256,'head A served text drifted');
function nodes(root,predicate){const result=[],stack=[root];while(stack.length){const node=stack.pop();if(predicate(node))result.push(node);ts.forEachChild(node,n=>{stack.push(n);});}return result;}
function source(tree,name){return ts.createSourceFile(name,fs.readFileSync(path.join(tree,name),'utf8'),ts.ScriptTarget.Latest,true);}
function method(file,name){const result=nodes(file,n=>ts.isMethodDeclaration(n)&&n.name.getText(file)===name);assert.equal(result.length,1);return result[0];}
function inspect(tree,corrected,strict){
 assert.equal(ts.version,'6.0.3');
 const configFile=path.join(tree,'src/compiler/tsconfig.json');
 const config=ts.getParsedCommandLineOfConfigFile(configFile,{}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(ts.flattenDiagnosticMessageText(d.messageText,'\n'));}});
 assert(config&&!config.errors.length);
 const options={...config.options,noEmit:true,typeRoots:[path.join(repo,'stage3/api/node_modules/@types')],types:['node']};
 if(strict)Object.assign(options,{strict:true,strictBindCallApply:true,useUnknownInCatchVariables:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,verbatimModuleSyntax:true,erasableSyntaxOnly:false,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,moduleDetection:ts.ModuleDetectionKind.Force,target:ts.ScriptTarget.ES2024});
 let ttyReads=0;
 const host=ts.createCompilerHost(options,true),original=host.readFile.bind(host);
 host.readFile=name=>{const text=original(name);if(name.replaceAll('\\','/').endsWith('/@types/node/tty.d.ts')){ttyReads++;assert.equal(hash(text),provenance.publishedSha256,'published Node declarations drifted');return corrected?served:text;}return text;};
 host.getSourceFile=(name,language)=>{const text=host.readFile(name);return text===undefined?undefined:ts.createSourceFile(name,text,language,true);};
 const program=ts.createProgram(config.fileNames,options,host),checker=program.getTypeChecker();
 const file=program.getSourceFile(path.join(tree,'src/compiler/sys.ts'));assert(file);
 const types={};
 for(const name of ['getWidthOfTerminal','writeOutputIsTTY']){
  const m=method(file,name);
  const access=nodes(m,n=>ts.isPropertyAccessExpression(n)&&n.expression.getText(file)==='process.stdout');assert.equal(access.length,1);
  types[name]=checker.typeToString(checker.getTypeAtLocation(access[0]));
 }
 const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>{const where=d.file&&d.start!==undefined?d.file.getLineAndCharacterOfPosition(d.start):null;return{file:d.file?path.relative(tree,d.file.fileName):'',line:where?where.line+1:0,column:where?where.character+1:0,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};}).sort((a,b)=>JSON.stringify(a).localeCompare(JSON.stringify(b)));
 assert(ttyReads>0,'checker never read the pinned tty.d.ts');
 assert.equal(types.getWidthOfTerminal,corrected?'number | undefined':'number');
 assert.equal(types.writeOutputIsTTY,corrected?'boolean | undefined':'boolean');
 return{corrected,strict,types,ttyReads,diagnostics};
}
function additions(before,after){const old=new Set(before.diagnostics.map(x=>JSON.stringify(x)));return after.diagnostics.filter(x=>!old.has(JSON.stringify(x)));}
function TestTTYHostAxiomTypecheck(tree){
 const results=[];
 for(const strict of [false,true]){
  const published=inspect(tree,false,strict),corrected=inspect(tree,true,strict),extra=additions(published,corrected);
  if(extra.length)console.error(extra.map(d=>`${d.file}:${d.line}:${d.column}: error TS${d.code}: ${d.message}`).join('\n'));
  assert.deepEqual(extra,[],'TestTTYHostAxiomTypecheck: head A introduces checker errors');
  assert.deepEqual(corrected.diagnostics,published.diagnostics,'corrected/published diagnostic sets differ');
  results.push({strict,published:published.diagnostics.length,corrected:corrected.diagnostics.length,types:corrected.types});
 }
 return results;
}
function TestTTYSourceScope(before,after){
 const files=tree=>{const out=[];function walk(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const p=path.join(dir,e.name);if(e.isDirectory())walk(p);else out.push(path.relative(tree,p));}}walk(path.join(tree,'src'));return out.sort();};
 const names=files(before);assert.deepEqual(files(after),names);
 const changed=names.filter(name=>hash(fs.readFileSync(path.join(before,name)))!==hash(fs.readFileSync(path.join(after,name))));
 assert.deepEqual(changed,['src/compiler/sys.ts']);
 const a=source(before,changed[0]),b=source(after,changed[0]);
 let expected=a.text;
 for(const target of require('../adapt.cjs').targets){
  const old=method(a,target.name),next=method(b,target.name);
  assert.equal(old.body.statements[0].expression.getText(a),target.before);
  assert.equal(next.body.statements[0].expression.getText(b),target.after);
  const fragment=`return ${target.before};`;assert.equal(expected.split(fragment).length,2);expected=expected.replace(fragment,`return ${target.after};`);
 }
 assert.equal(b.text,expected,'a change outside the two returned expressions');
 const interfaces=file=>nodes(file,n=>ts.isInterfaceDeclaration(n)&&n.name.text==='System').map(n=>n.getText(file));assert.deepEqual(interfaces(b),interfaces(a));
 for(const removeComments of [false,true]){
  const emit=file=>ts.transpileModule(file.text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
  const oldJS=emit(a),newJS=emit(b);assert.notEqual(oldJS,newJS);
  assert.equal(newJS.replace('return process.stdout.columns ?? 0;','return process.stdout.columns;').replace('return !!process.stdout.isTTY;','return process.stdout.isTTY;'),oldJS,'JavaScript changed outside the two bodies');
 }
 return{files:names.length,changed,systemInterfaceUnchanged:true,javascriptChanging:true};
}
function TestTTYConsumerEquivalence(before,after){
 const a=source(before,'src/compiler/sys.ts'),b=source(after,'src/compiler/sys.ts');
 const cli=source(before,'src/compiler/executeCommandLine.ts');
 const widths=nodes(cli,n=>ts.isVariableDeclaration(n)&&n.name.getText(cli)==='terminalWidth');assert.equal(widths.length,2);for(const n of widths)assert.equal(n.initializer.getText(cli),'sys.getWidthOfTerminal?.() ?? 0');
 const defaultPretty=nodes(cli,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='defaultIsPretty');assert.equal(defaultPretty.length,1);
 const colors=nodes(cli,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createColors');assert.equal(colors.length,1);
 const js=ts.transpileModule(defaultPretty[0].getText(cli)+'\n'+colors[0].getText(cli),{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
 function system(file,column,tty,env){const body=['getWidthOfTerminal','writeOutputIsTTY'].map(name=>`${name}() ${method(file,name).body.getText(file)}`).join(',');const sys=vm.runInNewContext(`({${body}})`,{process:{stdout:{columns:column,isTTY:tty}}});sys.getEnvironmentVariable=name=>env[name]||'';return sys;}
 let cases=0;
 for(const column of [undefined,0,40,79,80,120,160])for(const tty of [undefined,false,true])for(const env of [{},{NO_COLOR:'1'},{FORCE_COLOR:'1'}]){
  const left=system(a,column,tty,env),right=system(b,column,tty,env);
  for(const reader of widths){const text=reader.initializer.getText(cli);assert.equal(vm.runInNewContext(text,{sys:left}),vm.runInNewContext(text,{sys:right}),'TestTTYConsumerEquivalence: width changed');}
  const x=vm.runInNewContext(js+';({pretty:defaultIsPretty(sys),colors:createColors(sys)})',{sys:left});
  const y=vm.runInNewContext(js+';({pretty:defaultIsPretty(sys),colors:createColors(sys)})',{sys:right});
  assert.equal(!!x.pretty,!!y.pretty,'TestTTYConsumerEquivalence: pretty branch changed');
  for(const name of ['bold','blue','blueBackground','brightWhite'])assert.equal(x.colors[name]('sample'),y.colors[name]('sample'));
  cases++;
 }
 return{cases,widthReaders:widths.map(n=>cli.getLineAndCharacterOfPosition(n.getStart(cli)).line+1),prettyCaller:166};
}
module.exports={inspect,additions,TestTTYHostAxiomTypecheck,TestTTYSourceScope,TestTTYConsumerEquivalence};
if(require.main===module){
 if(process.argv[2]==='--require-repaired'){TestTTYHostAxiomTypecheck(path.resolve(process.argv[3]));console.log('PASS TestTTYHostAxiomTypecheck');}
 else{const before=path.resolve(process.argv[2]),after=path.resolve(process.argv[3]);const baseline=[];for(const strict of [false,true]){const published=inspect(before,false,strict),corrected=inspect(before,true,strict),extra=additions(published,corrected);assert.deepEqual(extra.map(x=>[x.code,x.file,x.line,x.column]).sort(),[[2322,'src/compiler/sys.ts',1531,13],[2322,'src/compiler/sys.ts',1534,13]]);baseline.push({strict,published:published.diagnostics.length,corrected:corrected.diagnostics.length,added:extra});}const result={baseline,adapted:TestTTYHostAxiomTypecheck(after),scope:TestTTYSourceScope(before,after),consumers:TestTTYConsumerEquivalence(before,after)};fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));}
}
