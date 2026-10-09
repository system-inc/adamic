import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import {registerHooks, createRequire} from 'node:module';
import {fileURLToPath, pathToFileURL} from 'node:url';
import {resolve, dirname} from 'node:path';
const require=createRequire(import.meta.url), stock=require(process.env.STEP31_TYPESCRIPT);
const scout=process.argv[4];
const tree=process.argv[2], output=process.argv[3];
const calls=[]; let context;
function encode(value) {
 if(value===undefined)return {$undefined:true};
 if(value===null || ['number','boolean','string'].includes(typeof value))return value;
 if(Array.isArray(value))return value.map(encode);
 if(typeof value.kind==='number' && typeof value.pos==='number')return {$node:true,kind:value.kind,pos:value.pos,end:value.end,text:value.text,fileName:value.fileName};
 if(value instanceof Map)return {$map:[...value].map(([k,v])=>[encode(k),encode(v)])};
 return Object.fromEntries(Object.entries(value).map(([key,item])=>[key,encode(item)]));
}
globalThis.__step31Record=(resolver,host)=>{
 context={hostKeys:Object.keys(host), resolverKeys:Object.keys(resolver),sources:host.getSourceFiles().map(f=>({fileName:f.fileName,text:f.text})),options:host.getCompilerOptions()};
 function proxy(target,group){return new Proxy(target,{get(object,key){const value=object[key];if(typeof value!=='function')return value;return (...args)=>{const result=value.apply(object,args);if(key!=='writeFile')calls.push({group,name:key,args:args.map(encode),result:encode(result)});return result;};}});}
 return [proxy(resolver,'resolver'),proxy(host,'host')];
};
registerHooks({resolve(specifier,context,next){if(specifier.startsWith('.')&&specifier.endsWith('.js')){const u=new URL(specifier.slice(0,-3)+'.ts',context.parentURL);if(existsSync(fileURLToPath(u)))return{url:u.href,shortCircuit:true};}return next(specifier,context);},load(url,context,next){if(url.endsWith('.ts')){let source=readFileSync(fileURLToPath(url),'utf8');if(url.endsWith('/emitter.ts')){const sf=stock.createSourceFile(url,source,stock.ScriptTarget.Latest,true);const fn=sf.statements.find(n=>stock.isFunctionDeclaration(n)&&n.name?.text==='emitFiles');source=source.slice(0,fn.body.getStart(sf)+1)+'\n[resolver, host] = globalThis.__step31Record(resolver, host);\n'+source.slice(fn.body.getStart(sf)+1);}return {format:'module',shortCircuit:true,source:stock.transpileModule(source,{compilerOptions:{target:stock.ScriptTarget.ESNext,module:stock.ModuleKind.ESNext,verbatimModuleSyntax:true}}).outputText};}return next(url,context);}});
const api=await import(pathToFileURL(resolve(tree,'src/compiler/_namespaces/ts.ts')).href);
const {observer}=require(scout+'/component-dump.cjs');
const request={projects:[{id:'emitter-recorded',options:{target:'es2020',module:'commonjs',noLib:true,types:[]},files:[{path:'recorded.a',text:'// retained comment\nexport const value: number = 41 + 1;\nexport function answer(): number { return value; }\n'}]}]};
const rows=observer(api,dirname(process.env.STEP31_TYPESCRIPT)).observe(request.projects[0],'emitter');
writeFileSync(output,JSON.stringify({request,context,calls,rows},null,2)+'\n');
process.stdout.write(rows.map(r=>JSON.stringify(r)).join('\n')+'\n');
