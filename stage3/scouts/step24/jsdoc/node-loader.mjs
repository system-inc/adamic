// Load the actual adapted source parser. Stock TypeScript only transpiles syntax.
import {registerHooks,createRequire} from 'node:module';
import {existsSync,readFileSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
const ts=createRequire(import.meta.url)(process.env.JSDOC_TYPESCRIPT);
if(ts.version!=='6.0.3')throw Error('stock API must be 6.0.3');
registerHooks({
 resolve(specifier,context,next){
  if(specifier.startsWith('.')&&specifier.endsWith('.js')){
   const url=new URL(specifier.slice(0,-3)+'.ts',context.parentURL);
   if(existsSync(fileURLToPath(url)))return {url:url.href,shortCircuit:true};
  }
  return next(specifier,context);
 },
 load(url,context,next){
  if(url.endsWith('.ts')){
   let source=readFileSync(fileURLToPath(url),'utf8');
   if(url.endsWith('/src/compiler/parser.ts')&&process.env.JSDOC_TIMING==='1'){
    const sf=ts.createSourceFile('parser.ts',source,ts.ScriptTarget.Latest,true);let found=[];
    function walk(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='parseJSDocComment')found.push(n);ts.forEachChild(n,walk);}walk(sf);
    if(found.length!==1)throw Error('timing site drift');const f=found[0],body=source.slice(f.body.getStart()+1,f.body.end-1);
    source=source.slice(0,f.body.getStart())+'{ const jsdocStart = globalThis.performance.now(); try {'+body+'} finally { const s=globalThis.__jsdocTiming; if(s){s.calls++; s.ms+=globalThis.performance.now()-jsdocStart;} } }'+source.slice(f.body.end);
   }
   return {format:'module',shortCircuit:true,source:ts.transpileModule(source,{fileName:fileURLToPath(url),compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true}}).outputText};
  }
  return next(url,context);
 }
});
