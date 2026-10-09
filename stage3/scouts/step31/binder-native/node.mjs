// Stock transpilation only. Bootstrapping follows scanner/node.mjs.
import {readFileSync,existsSync} from 'node:fs';
import {registerHooks,createRequire} from 'node:module';
import {fileURLToPath,pathToFileURL} from 'node:url';
import {resolve} from 'node:path';
const require=createRequire(import.meta.url),ts=require(process.env.STEP31_TYPESCRIPT);
const runtime=pathToFileURL(process.env.BINDER_RUNTIME).href;
registerHooks({resolve(specifier,context,next){
 if(specifier==='adamic')return {url:runtime,shortCircuit:true};
 if(specifier.startsWith('.')&&specifier.endsWith('.js')){const url=new URL(specifier.slice(0,-3)+'.ts',context.parentURL);if(existsSync(fileURLToPath(url)))return {url:url.href,shortCircuit:true};}
 return next(specifier,context);
},load(url,context,next){
 if(url.endsWith('.ts')||url.endsWith('.a'))return {format:'module',shortCircuit:true,source:ts.transpileModule(readFileSync(fileURLToPath(url),'utf8'),{fileName:fileURLToPath(url).replace(/\.a$/,'.ts'),compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true}}).outputText};
 return next(url,context);
}});
const driver=resolve(process.argv[2]);process.argv.splice(1,1);
await import(runtime);
await import(pathToFileURL(resolve(driver,'../src/compiler/_namespaces/ts.ts')).href);
await import(pathToFileURL(driver).href);
