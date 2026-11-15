const fs=require('fs'),vm=require('vm'),assert=require('assert/strict'),ts=require('typescript');
const main=require('./baseline-typescript.cjs'),text=fs.readFileSync(process.argv[2],'utf8'),sf=ts.createSourceFile('sys.ts',text,ts.ScriptTarget.Latest,true),fn=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createDynamicPriorityPollingWatchFile');assert(fn);
const js=ts.transpileModule(fn.getText(sf),{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
const create=vm.runInNewContext(js+'\ncreateDynamicPriorityPollingWatchFile;',{...main});
const register=interval=>create({getModifiedTime:()=>new Date(0),setTimeout:()=>({})})('/probe.ts',()=>{},interval);
for(const interval of [250,500,2000])register(interval).close();
for(const interval of [10,undefined])assert.throws(()=>register(interval),error=>error.name==='Error'&&error.message.includes('Invalid dynamic polling interval: '+String(interval)));
console.log(JSON.stringify({members:[250,500,2000],invalid10AndUndefinedFailAtBoundary:true}));
