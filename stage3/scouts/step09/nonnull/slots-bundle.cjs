const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript'),esbuild=require('esbuild');
const [rootArg,outArg]=process.argv.slice(2),root=path.resolve(rootArg),out=path.resolve(outArg);
const manifest=JSON.parse(fs.readFileSync(path.join(__dirname,'placeholder-slots.json'))),receipt={initializers:[],files:[]};
const runtime=require('./slot-runtime.cjs');
const banner='const __adamic_slots = ('+runtime.toString()+')();\n'+`process.on('exit',()=>require('node:fs').writeFileSync(process.env.ADAMIC_SLOT_LOG+'/'+require('node:path').basename(process.cwd())+'.json',JSON.stringify({input:process.cwd(),slots:__adamic_slots.stats})+'\\n'));`;
(async()=>{fs.mkdirSync(out,{recursive:true});await esbuild.build({entryPoints:[path.join(root,'src/tsc/tsc.ts')],bundle:true,treeShaking:false,platform:'node',format:'cjs',target:'node24',outfile:path.join(out,'slots.cjs'),banner:{js:banner},plugins:[{name:'placeholder-slots',setup(build){build.onResolve({filter:/^tslib$/},()=>({path:require.resolve('tslib')}));build.onLoad({filter:/\.ts$/},args=>{
const file=path.relative(root,args.path),text=fs.readFileSync(args.path,'utf8');const transformed=ts.transpileModule(text,{fileName:args.path,compilerOptions:{target:ts.ScriptTarget.ES5,importHelpers:true,noEmitHelpers:true,downlevelIteration:true,useDefineForClassFields:true,alwaysStrict:true,module:ts.ModuleKind.ESNext},transformers:require('./slots-transform.cjs')(file,manifest,receipt)});
receipt.files.push(file);return {contents:transformed.outputText,loader:'js'};
});}}]});assert.deepEqual(receipt.initializers.sort(),manifest.slots.map(s=>s.id).sort());
const lib=path.dirname(require.resolve('typescript'));for(const name of fs.readdirSync(lib).filter(n=>n.startsWith('lib.')&&n.endsWith('.d.ts')))fs.symlinkSync(path.join(lib,name),path.join(out,name));
fs.writeFileSync(path.join(out,'receipt.json'),JSON.stringify(receipt,null,2)+'\n');console.log('50 literal placeholder sites instrumented; reads, writes and per-instance first-read epochs tracked');
})().catch(e=>{console.error(e);process.exitCode=1;});
