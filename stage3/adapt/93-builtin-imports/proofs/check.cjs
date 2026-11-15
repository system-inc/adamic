'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript');
function hash(bytes){return crypto.createHash('sha256').update(bytes).digest('hex');}
function inventory(tree){return fs.readdirSync(path.join(tree,'src'),{recursive:true,withFileTypes:true}).filter(e=>e.isFile()).map(e=>path.relative(tree,path.join(e.parentPath,e.name))).sort();}
function undo(text){
 for(const [name,builtin] of [['_fs','fs'],['_path','path'],['_os','os']]){
  const line=`import * as ${name} from "node:${builtin}";`;
  assert.equal(text.split(line).length,2,'reviewed import missing or duplicated');
  text=text.replace(new RegExp('\\r?\\n'+line.replace(/[.*+?^${}()|[\]\\]/g,'\\$&')),'');
 }
 return text;
}
// Pins hold hashes of whole AST owners, not copies of their text. Read both
// snapshots and authenticate each owner before using its exact bytes as an undo.
const pins=require('./targets.json');
function source(text,file){
 const parsed=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
 assert.equal(parsed.parseDiagnostics.length,0,file+': cannot parse snapshot');
 return parsed;
}
function owner(parsed,pin,side){
 const found=[];
 function visit(node){
  if(node.name?.getText(parsed)===pin.owner && ts.isFunctionDeclaration(node))found.push(node);
  ts.forEachChild(node,visit);
 }
 visit(parsed);
 assert.equal(found.length,1,pin.file+': owner missing or duplicated: '+pin.owner);
 const node=found[0],text=node.getText(parsed);
 assert.equal(hash(text),pin['owner'+side+'Sha256'],
  pin.file+': pinned '+side+' owner differs: '+pin.owner);
 return {start:node.getStart(parsed),end:node.end,text};
}
function restore(a,b,file){
 const before=source(a,file),after=source(b,file),edits=[];
 for(const pin of pins.filter(p=>p.file===file)){
  const old=owner(before,pin,'Before'),current=owner(after,pin,'After');
  edits.push({...current,text:old.text});
 }
 edits.sort((a,b)=>b.start-a.start);
 let boundary=b.length;
 for(const edit of edits){
  assert(edit.end<=boundary,'overlapping owner pins');
  b=b.slice(0,edit.start)+edit.text+b.slice(edit.end);
  boundary=edit.start;
 }
 // The removed declarations are inside getNodeSystem and authenticated by its
 // whole-owner pins. Only the three import insertions need a separate undo.
 return file==='src/compiler/sys.ts'?undo(b):b;
}
function equalBytes(actual,expected,file){
 if(actual===expected)return;
 let at=0;
 while(at<actual.length && at<expected.length && actual[at]===expected[at])at++;
 const line=actual.slice(0,at).split('\n').length;
 const context=actual.slice(actual.lastIndexOf('\n',at-1)+1).split(/\r?\n/)[0];
 assert.fail('TestBuiltinSourceScope: edit outside pinned owners: '+file+':'+line+': '+context);
}

function TestBuiltinSourceScope(before,after){
 const names=inventory(before);assert.deepEqual(inventory(after),names,'source inventory changed');
 const entries=names.map(file=>({file,before:hash(fs.readFileSync(path.join(before,file))),after:hash(fs.readFileSync(path.join(after,file)))}));
 const changed=entries.filter(x=>x.before!==x.after).map(x=>x.file);
 assert.deepEqual(changed,[...new Set(pins.map(p=>p.file))].sort(),'source changed outside pinned files: '+changed.join(', '));
 for(const file of names){
  const read=tree=>{
   const bytes=fs.readFileSync(path.join(tree,file)),text=bytes.toString('utf8');
   assert.deepEqual(Buffer.from(text),bytes,file+': snapshot is not UTF-8');
   return text;
  };
  const a=read(before),b=read(after);
  equalBytes(pins.some(p=>p.file===file)?restore(a,b,file):b,a,file);
 }
 return{files:names.length,changed,undoRestoresExactBytes:true,inventory:entries};
}
function TestBuiltinTranspileScope(before,after){
 const read=tree=>fs.readFileSync(path.join(tree,'src/compiler/sys.ts'),'utf8');
 const results=[];
 for(const removeComments of [false,true]){
  const emit=text=>ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
  const original=read(before),changed=read(after);
  const a=emit(original),b=emit(changed);assert.notEqual(a,b);
  assert.equal(emit(restore(original,changed,'src/compiler/sys.ts')),a,'JavaScript undo differs');
  results.push({removeComments,beforeSha256:hash(a),afterSha256:hash(b),onlyReviewedChanges:true});
 }
 return results;
}
module.exports={hash,inventory,undo,TestBuiltinSourceScope,TestBuiltinTranspileScope};
if(require.main===module){
 const before=path.resolve(process.argv[2]),after=path.resolve(process.argv[3]);
 const result={scope:TestBuiltinSourceScope(before,after),transpile:TestBuiltinTranspileScope(before,after)};
 if(process.argv[4])fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');
 console.log(JSON.stringify({...result,scope:{...result.scope,inventory:`${result.scope.inventory.length} entries recorded in checks.json`}},null,2));
}
