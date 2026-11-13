// Node reference: published TypeScript 6.0.3's sys initializes via getNodeSystem.
const tscBundle = process.argv.includes('--tsc');
process.argv = process.argv.filter(a => a !== '--tsc');
const noVariables = process.argv.includes('--no-vars');
process.argv = process.argv.filter(a => a !== '--no-vars');
if (noVariables) {
 for (const key of Object.keys(process.env)) if (key.startsWith('TSC_WATCH')) delete process.env[key];
} else {
process.env.TSC_WATCHFILE='DynamicPriorityPolling';
process.env.TSC_WATCH_POLLINGINTERVAL_LOW='17';
process.env.TSC_WATCH_POLLINGINTERVAL_MEDIUM='41';
process.env.TSC_WATCH_POLLINGINTERVAL_HIGH='89';
}
const timers=[];
global.setTimeout=(callback,interval,...args)=>{timers.push({callback:callback.name,interval});return {unref(){}};};
global.clearTimeout=()=>{};
const api=require('../../api/node_modules/typescript');
let ts=api;
if (tscBundle) {
 const Module = require('module'), fs = require('fs');
 const file = require.resolve('../../api/node_modules/typescript/lib/_tsc.js');
 const source = fs.readFileSync(file,'utf8');
 const tail = 'executeCommandLine(sys, noop, sys.args);';
 if (source.split(tail).length !== 2) throw Error('CLI tail drift');
 const m = new Module(file, module); m.filename=file; m.paths=Module._nodeModulePaths(require('path').dirname(file));
 m._compile(source.replace(tail,'module.exports = {sys, PollingInterval, version};'),file);
 ts=m.exports;
}
if(process.argv[2]) {
 const Module=require('module'),fs=require('fs'),path=require('path');
 const root=path.resolve(process.argv[2]);const resolve=Module._resolveFilename;
 Module._resolveFilename=function(request,parent,...rest){
  if(parent?.filename.startsWith(root)&&request.endsWith('.js')&&request.startsWith('.')) {
   const source=path.resolve(path.dirname(parent.filename),request.slice(0,-3)+'.ts');
   if(fs.existsSync(source))return source;
  }
  return resolve.call(this,request,parent,...rest);
 };
 require.extensions['.ts']=(m,f)=>m._compile(api.transpileModule(fs.readFileSync(f,'utf8'),{compilerOptions:{module:api.ModuleKind.CommonJS,target:api.ScriptTarget.ES2022}}).outputText,f);
 ts=require(path.join(root,'src/compiler/_namespaces/ts.ts'));
 console.log('mode source CommonJS (transpileModule; no enum inlining)');
} else console.log(tscBundle ? 'mode shipped _tsc.js (CLI dispatch replaced by exports)' : 'mode published bundle');
console.log('Node',process.version);
console.log('environment',noVariables ? 'none' : 'DynamicPriorityPolling 17/41/89');
console.log('TypeScript',ts.version);
console.log('PollingInterval',JSON.stringify(ts.PollingInterval));
const watchers=[];
for(const interval of (noVariables ? [250,500,2000] : [250,500,2000,17,41,89])) {
 try { watchers.push(ts.sys.watchFile(__filename,()=>{},interval)); console.log('watchFile',interval,'ok'); }
 catch(e) { console.log('watchFile',interval,e.name+': '+e.message); }
}
console.log('scheduled',JSON.stringify(timers));
watchers.forEach(w=>w.close());
