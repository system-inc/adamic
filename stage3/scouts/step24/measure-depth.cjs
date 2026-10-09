// Node oracle only. Instrument every block-bodied function under the bundled
// Parser IIFE. Count active parser frames, not AST height or native C frames.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const Module = require('node:module');
const ts = require(process.env.PARSER_TYPESCRIPT);
const library = fs.readFileSync(process.env.PARSER_TYPESCRIPT, 'utf8');
const ast = ts.createSourceFile('typescript.js', library, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
let root;
function findRoot(n) {
 if (ts.isArrowFunction(n) && n.parameters.length === 1 && n.parameters[0].name.getText(ast) === 'Parser2') root = n.body;
 ts.forEachChild(n, findRoot);
}
findRoot(ast);
if (!root) throw Error('Parser IIFE missing');
const edits = [], functions = [];
function visit(n) {
 if (ts.isFunctionLike(n) && n.body && ts.isBlock(n.body)) {
  const name = n.name?.getText(ast) || '<callback>';
  functions.push({name, line: ast.getLineAndCharacterOfPosition(n.getStart(ast)).line + 1});
  edits.push([n.body.getStart(ast) + 1, `globalThis.__step24.enter(${JSON.stringify(name)});try{`]);
  edits.push([n.body.end - 1, '}finally{' + (process.env.STEP24_MUTANT_NODE_END === '1' && name === 'finishNode' ? 'node.end += 1;' : '') + 'globalThis.__step24.leave();}']);
 }
 ts.forEachChild(n, visit);
}
visit(root);
let instrumented = library;
for (const [offset, text] of edits.sort((a,b) => b[0]-a[0])) instrumented = instrumented.slice(0,offset)+text+instrumented.slice(offset);
const active = [];
let maximum = 0, deepest = [], enters = 0;
globalThis.__step24 = {
 enter(name) {if(process.env.STEP24_MUTANT_EMPTY_COUNTER === '1') return; active.push(name); enters++; if(active.length > maximum) {maximum=active.length; deepest=active.slice();}},
 leave() {active.pop();},
};
const mod = new Module(process.env.PARSER_TYPESCRIPT + '.step24', module);
mod.filename = process.env.PARSER_TYPESCRIPT;
mod.paths = module.paths;
mod._compile(instrumented, mod.filename);
const measured = mod.exports;
if (ts.version !== '6.0.3' || measured.version !== '6.0.3') throw Error('wrong oracle version');
function projection(file) {
 const rows = [], pending = [file]; let height = 0;
 const levels = [1];
 while(pending.length) {
  const n = pending.pop(), level = levels.pop(); height = Math.max(height, level);
  rows.push([n.kind,n.pos,n.end,n.flags,n.escapedText,n.text,n.tagName?.escapedText,
   typeof n.comment === 'string' ? n.comment : n.comment?.map(x=>[x.kind,x.text])]);
  const children = [...(n.jsDoc || [])];
  ts.forEachChild(n, child=>{children.push(child);});
  for(let i=children.length-1;i>=0;i--) {pending.push(children[i]);levels.push(level+1);}
 }
 rows.push(file.parseDiagnostics.map(d=>[d.code,d.start,d.length,ts.flattenDiagnosticMessageText(d.messageText,'\n')]),
  file.jsDocDiagnostics?.map(d=>[d.code,d.start,d.length,ts.flattenDiagnosticMessageText(d.messageText,'\n')]));
 return {height, hash:crypto.createHash('sha256').update(JSON.stringify(rows)).digest('hex')};
}
const tree = path.resolve(process.argv[2]), output = process.argv[3];
const cases = JSON.parse(fs.readFileSync(path.join(__dirname,'../../drivers/parser/cases-reference.json'),'utf8'));
const records = [];
function observe(name,text,kind,group) {
 maximum=0;deepest=[];enters=0;
 const plain = ts.createSourceFile(name,text,ts.ScriptTarget.Latest,false,kind);
 const truth = projection(plain);
 let error;
 try {
  const file = measured.createSourceFile(name,text,ts.ScriptTarget.Latest,false,kind);
  if (projection(file).hash !== truth.hash) throw Error('instrumentation changed oracle projection');
 } catch(e) {error = String(e);}
 if(active.length) throw Error('unbalanced instrumentation');
 if(error?.includes('instrumentation changed')) throw Error(error);
 if(maximum === 0) throw Error('parser depth counter did not observe a frame');
 records.push({name,group,maximum,astHeight:truth.height,enters,deepest,error});
}
if(process.argv[4] === '--probe') {
 observe('probe.ts','const x = (((1)));',ts.ScriptKind.TS,'probe');
 fs.writeFileSync(output,JSON.stringify(records,null,2)+'\n');
 console.log(records[0].maximum);
 process.exit(0);
}
function files(dir) {return fs.readdirSync(dir,{withFileTypes:true}).flatMap(d=>d.isDirectory()?files(path.join(dir,d.name)):[path.join(dir,d.name)]).sort();}
for(const f of files(path.join(tree,'src/compiler'))) {
 observe(path.relative(tree,f),ts.sys.readFile(f),ts.ScriptKind.TS,'compiler');
}
// Same single-file exclusions and effective filename modes as the mandatory reference.
const entries = cases.records || cases.cases;
if(!Array.isArray(entries)) throw Error('case manifest schema changed: '+Object.keys(cases));
for(const r of entries) {
 const effective=r.effective_filename, suffix=path.extname(effective).toLowerCase();
 const kind=suffix==='.tsx'?ts.ScriptKind.TSX:suffix==='.jsx'?ts.ScriptKind.JSX:['.js','.mjs','.cjs'].includes(suffix)?ts.ScriptKind.JS:ts.ScriptKind.TS;
 observe(r.path,ts.sys.readFile(path.join(tree,r.path)),kind,'case');
}
const ranked = records.slice().sort((a,b)=>b.maximum-a.maximum);
const report={node:process.version,typescript:ts.version,definition:'active block-bodied functions in stock bundled Parser IIFE, including callbacks; scanner/factory frames excluded',
 instrumentedFunctions:functions.length, librarySHA256:crypto.createHash('sha256').update(library).digest('hex'),
 groups:['compiler','case'].map(group=>({group,count:records.filter(r=>r.group===group).length,max:ranked.find(r=>r.group===group),errors:records.filter(r=>r.group===group&&r.error)})),
 top:ranked.slice(0,20),records};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify({functions:functions.length,groups:report.groups,top:report.top.slice(0,3)},null,2));
