// Build stock tsc with identity observers at the checked AST sites.
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript'), esbuild = require('esbuild');
const [rootArg,outArg] = process.argv.slice(2);
const root=path.resolve(rootArg), out=path.resolve(outArg);
assert.equal(ts.version,'6.0.3'); assert.equal(esbuild.version,'0.27.3');
const sites=JSON.parse(fs.readFileSync(path.join(__dirname,'sites.json')));
const checked=sites.filter(s=>s.status==='checked');
const byFile=new Map();
for(const s of checked){if(!byFile.has(s.file))byFile.set(s.file,new Map());byFile.get(s.file).set(s.start+':'+s.end,s);}
const printer=ts.createPrinter({newLine:ts.NewLineKind.LineFeed});
const banner=`const __adamic_probe_fs = require('node:fs');
const __adamic_probe_hits = Object.create(null), __adamic_probe_nullish = Object.create(null);
function __adamic_nonnull_probe(value, id) {
 __adamic_probe_hits[id] = (__adamic_probe_hits[id] || 0) + 1;
 if (value === null || value === undefined) {
  const key = id + ':' + (value === null ? 'null' : 'undefined');
  __adamic_probe_nullish[key] = (__adamic_probe_nullish[key] || 0) + 1;
 }
 return value;
}
process.on('exit', () => {
 if (!process.env.ADAMIC_NONNULL_LOG) throw Error('missing probe log');
 __adamic_probe_fs.writeFileSync(process.env.ADAMIC_NONNULL_LOG + '/' + require('node:path').basename(process.cwd()) + '.json', JSON.stringify({input:process.cwd(),argv:process.argv.slice(2),hits:__adamic_probe_hits,nullish:__adamic_probe_nullish}) + '\\n');
});`;
(async()=>{
 fs.mkdirSync(out,{recursive:true});
 for(const mode of ['control','instrumented']){
  const inserted=[];
  await esbuild.build({entryPoints:[path.join(root,'src/tsc/tsc.ts')],bundle:true,treeShaking:mode==='instrumented'?false:undefined,platform:'node',format:'cjs',target:'node24',outfile:path.join(out,mode+'.cjs'),banner:mode==='instrumented'?{js:banner}:{},plugins:[{name:'compiler-api-nonnull',setup(build){build.onLoad({filter:/\.ts$/},args=>{
   const file=path.relative(root,args.path), text=fs.readFileSync(args.path,'utf8');
   const source=ts.createSourceFile(args.path,text,ts.ScriptTarget.Latest,true);
   const anchors=byFile.get(file);
   if(mode==='control'||!anchors)return {contents:printer.printFile(source),loader:'ts'};
   const contents=require('./instrument.cjs')(source,anchors,inserted);
   return {contents,loader:'ts'};
  });}}]});
  if(mode==='instrumented'){
   assert.deepEqual(inserted.sort(),checked.map(s=>s.file+':'+s.line+':'+s.column+'@'+s.start+'-'+s.end).sort());
   assert.equal(new Set(inserted).size, checked.length, 'nested assertions need distinct end spans');
   const emitted=ts.createSourceFile('bundle.js',fs.readFileSync(path.join(out,mode+'.cjs'),'utf8'),ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);
   const retained=new Set();function retainedVisit(n){if(ts.isCallExpression(n)&&n.expression.getText(emitted)==='__adamic_nonnull_probe')retained.add(n.arguments[1].text);ts.forEachChild(n,retainedVisit);}retainedVisit(emitted);
   assert.deepEqual([...retained].sort(),inserted.sort(),'all source assertion probes must survive bundling');
   fs.writeFileSync(path.join(out,'instrumented-sites.json'),JSON.stringify(inserted,null,2)+'\n');
  }
 }
 // tsc resolves its default libs beside the executable.
 const lib=path.dirname(require.resolve('typescript'));
 for(const name of fs.readdirSync(lib).filter(n=>n.startsWith('lib.')&&n.endsWith('.d.ts'))){const dest=path.join(out,name);if(!fs.existsSync(dest))fs.symlinkSync(path.join(lib,name),dest);}
 console.log('built stock control and instrumented CLI; '+checked.length+' sites inserted exactly once');
})().catch(e=>{console.error(e);process.exitCode=1;});
