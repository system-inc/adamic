// Stock TypeScript parses function boundaries; edits only add coverage probes.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const cachedAPI = path.join(process.env.STAGE3_CACHE || path.join(require('node:os').homedir(),'.cache/adamic-stage3'),'api/node_modules/typescript');
const ts = require(process.env.SLICE_TYPESCRIPT || (fs.existsSync(cachedAPI) ? cachedAPI : 'typescript'));
if (ts.version !== '6.0.3') throw Error('stock TypeScript 6.0.3 required');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const [mode, rootArg, outputArg] = process.argv.slice(2), root = path.resolve(rootArg);
const files = [];
function walk(dir) {
 for (const entry of fs.readdirSync(dir, {withFileTypes:true}).sort((a,b)=>a.name.localeCompare(b.name))) {
  const file = path.join(dir,entry.name);
  if (entry.isDirectory()) walk(file);
  else if (file.endsWith('.ts') && !file.endsWith('.d.ts')) files.push(path.relative(root,file).split(path.sep).join('/'));
 }
}
walk(path.join(root,'src/compiler')); walk(path.join(root,'src/tsc'));
const reference=process.argv[5] ? JSON.parse(fs.readFileSync(process.argv[5],'utf8')) : undefined;
const inventory = {schema:1, files:{}, functions:[]}, transformed = new Map();
for (const file of files) {
 const text = fs.readFileSync(path.join(root,file),'utf8');
 if (mode==='inventory' && reference?.files[file]?.sha256===hash(text)) {
  inventory.files[file]=reference.files[file];
  for (const record of reference.functions.filter(f=>f.file===file))
   inventory.functions.push({...record,index:inventory.functions.length});
  continue;
 }
 const source = ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
 if (source.parseDiagnostics.length) throw Error('parse diagnostics: '+file);
 const functions = [], counters = new Map();
 function visit(node, parent) {
  let owner = parent;
  if (ts.isFunctionLike(node) && node.body) {
   const name = node.name?.getText(source) || '<anonymous>';
   const prefix = (parent?.id || file) + '/' + ts.SyntaxKind[node.kind] + ':' + name;
   const ordinal = counters.get(prefix) || 0; counters.set(prefix,ordinal+1);
   owner = {id:prefix+'#'+ordinal, index:inventory.functions.length+functions.length,
    start:node.getStart(source), end:node.end, bodyStart:node.body.getStart(source), bodyEnd:node.body.end,
    block:ts.isBlock(node.body), parent:parent?.id, node, children:[], file,
    line:source.getLineAndCharacterOfPosition(node.getStart(source)).line+1};
   if (parent) parent.children.push(owner);
   functions.push(owner);
  }
  ts.forEachChild(node,child=>visit(child,owner));
 }
 visit(source,null);
 function exclusive(start,end,children) {
  let result='',cursor=start;
  for (const child of children) { result+=text.slice(cursor,child.start)+'<function:'+child.id+'>';cursor=child.end; }
  return result+text.slice(cursor,end);
 }
 const edits=[];
 for (const f of functions) {
  f.sha256=hash(exclusive(f.start,f.end,f.children));
  if (f.block) {
   let at=f.bodyStart+1;
   for (const statement of f.node.body.statements) {
    if (!ts.isExpressionStatement(statement) || !ts.isStringLiteral(statement.expression)) break;
    at=statement.end;
   }
   edits.push({at,text:`\n__verdictHits[${f.index}]=1;\n`});
  }
  else { edits.push({at:f.bodyStart,text:`(__verdictHits[${f.index}]=1,(`}); edits.push({at:f.bodyEnd,text:'))'}); }
  const {node,children,...record}=f;inventory.functions.push(record);
 }
 inventory.files[file]={sha256:hash(text),outside_sha256:hash(exclusive(0,text.length,functions.filter(f=>!f.parent)))};
 if (mode==='inventory') continue;
 let instrumented=text;
 for (const edit of edits.sort((a,b)=>b.at-a.at)) instrumented=instrumented.slice(0,edit.at)+edit.text+instrumented.slice(edit.at);
 transformed.set(path.join(root,file),instrumented);
}
inventory.source_sha256=hash(JSON.stringify(Object.entries(inventory.files).map(([name,value])=>[name,value.sha256])));
if (mode==='inventory') fs.writeFileSync(outputArg,JSON.stringify(inventory));
else if (mode==='instrument' || mode==='instrument-server') {
 const output=path.resolve(outputArg);fs.mkdirSync(output,{recursive:true});
 fs.writeFileSync(path.join(output,'inventory.json'),JSON.stringify(inventory));
 if (mode==='instrument-server') {
  const entry=path.join(root,'src/tsc/tsc.ts');
  const text=transformed.get(entry), call='ts.executeCommandLine(ts.sys, ts.noop, ts.sys.args);';
  if (text.split(call).length!==2) throw Error('CLI entry changed');
  transformed.set(entry,text.replace(call,'export function verdictExecute(system, args) { ts.executeCommandLine(system, ts.noop, args); } export const verdictSystem = ts.sys; export const verdictHits = __verdictHits;'));
 }
 const esbuild=require(path.join(root,'node_modules/esbuild'));
 esbuild.build({entryPoints:[path.join(root,'src/tsc/tsc.ts')],outfile:path.join(output,'tsc.cjs'),bundle:true,platform:'node',format:'cjs',
  banner:{js:`"use strict";
const __verdictHits = new Uint8Array(${inventory.functions.length}); ${mode==='instrument' ? "process.on('exit',()=>require('node:fs').writeFileSync(process.env.TSC_COVERAGE_FILE || require('node:path').join(process.cwd(),'.verdict-coverage.json'),JSON.stringify(Array.from(__verdictHits.entries()).filter(x=>x[1]).map(x=>x[0]))));" : ''}`},
  plugins:[{name:'function-coverage',setup(build){build.onLoad({filter:/\.ts$/},args=>transformed.has(args.path)?{contents:transformed.get(args.path),loader:'ts',resolveDir:path.dirname(args.path)}:undefined);}}]
 }).then(()=>{
  for (const file of fs.readdirSync(path.join(root,'built/local')).filter(f=>/^lib.*\.d\.ts$/.test(f))) fs.copyFileSync(path.join(root,'built/local',file),path.join(output,file));
 }).catch(error=>{console.error(error);process.exitCode=1;});
} else throw Error('usage: function_map.cjs inventory|instrument SOURCE OUTPUT');
