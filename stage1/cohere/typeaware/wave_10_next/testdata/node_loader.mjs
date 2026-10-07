// Node executes the rule source against recorded raw compiler observations.
// This is neither the native runner nor the compiler's emitted-JavaScript gate.
import fs from 'node:fs';
import {registerHooks,stripTypeScriptTypes} from 'node:module';
import {fileURLToPath} from 'node:url';
const runtime=new URL('./node_runtime.mjs',import.meta.url).href;
registerHooks({
 resolve(specifier,context,nextResolve) {return specifier==='adamic'?{url:runtime,shortCircuit:true}:nextResolve(specifier,context);},
 load(url,context,nextLoad) {
  if(url.endsWith('.a')||url.endsWith('.ts')) {
   let source=fs.readFileSync(fileURLToPath(url),'utf8');
   let from;let to;
   if(process.env.ADAMIC_SOURCE_MUTANT==='process'&&url.endsWith('/no_process_exit_after_output.a')) {from='next.push(chain);';to='return [];';}
   if(process.env.ADAMIC_SOURCE_MUTANT==='blocking'&&url.endsWith('/require_blocking_standard_streams.a')) {from='this.ordered = this.held().canBlock.has(this.rules.path);';to='this.ordered = false;';}
   if(from!==undefined) {if(source.split(from).length!==2) {throw new Error('nonunique source mutant');}source=source.replace(from,to);}
   return {format:'module',source:stripTypeScriptTypes(source,{mode:'transform',sourceUrl:url}),shortCircuit:true};
  }
  return nextLoad(url,context);
 },
});
