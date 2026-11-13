// Node reference: published TypeScript 6.0.3's sys initializes via getNodeSystem.
process.env.TSC_WATCHFILE='DynamicPriorityPolling';
process.env.TSC_WATCH_POLLINGINTERVAL_LOW='17';
process.env.TSC_WATCH_POLLINGINTERVAL_MEDIUM='41';
process.env.TSC_WATCH_POLLINGINTERVAL_HIGH='89';
const timers=[];
global.setTimeout=(callback,interval,...args)=>{timers.push({callback:callback.name,interval});return {unref(){}};};
global.clearTimeout=()=>{};
const api=require('../../api/node_modules/typescript');
let ts=api;
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
} else console.log('mode published bundle');
console.log('TypeScript',ts.version);
console.log('PollingInterval',JSON.stringify(ts.PollingInterval));
const watchers=[];
for(const interval of [250,500,2000,17,41,89]) {
 try { watchers.push(ts.sys.watchFile(__filename,()=>{},interval)); console.log('watchFile',interval,'ok'); }
 catch(e) { console.log('watchFile',interval,e.name+': '+e.message); }
}
console.log('scheduled',JSON.stringify(timers));
watchers.forEach(w=>w.close());
