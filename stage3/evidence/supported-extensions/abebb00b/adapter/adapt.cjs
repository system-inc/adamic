'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const pin=require('./owner.json');
const normalize=s=>s.replace(/\r\n/g,'\n');
const hash=s=>crypto.createHash('sha256').update(normalize(s)).digest('hex');
const transpile=(text,removeComments=false)=>ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
function owner(text){
 const source=ts.createSourceFile('utilities.ts',text,ts.ScriptTarget.Latest,true);
 assert.equal(source.parseDiagnostics.length,0,'parse error');
 const found=[];
 function visit(node){if(ts.isFunctionDeclaration(node)&&node.name?.text==='getSupportedExtensions')found.push(node);ts.forEachChild(node,visit);}
 visit(source);
 assert.equal(found.length,3,'missing, duplicate or nested owner');
 assert(found.every(n=>n.parent===source),'nested owner');
 assert(!found[0].body&&!found[1].body&&found[2].body,'changed overload group');
 const indexes=found.map(n=>source.statements.indexOf(n));
 assert.equal(indexes[1],indexes[0]+1);assert.equal(indexes[2],indexes[1]+1);
 return {start:found[0].getStart(source),end:found[2].end,text:text.slice(found[0].getStart(source),found[2].end)};
}
function plan(text){
 assert.equal(ts.version,'6.0.3');
 const o=owner(text),actual=normalize(o.text);
 assert(hash(pin.before)===pin.beforeSha256&&hash(pin.after)===pin.afterSha256,'invalid pin');
 assert(actual===pin.before||actual===pin.after,'changed whole owner');
 if(actual===pin.after)return {text,edits:0};
 const eol=o.text.includes('\r\n')?'\r\n':'\n';
 const after=text.slice(0,o.start)+pin.after.replace(/\n/g,eol)+text.slice(o.end);
 assert.equal(normalize(owner(after).text),pin.after);
 for(const removeComments of [false,true])assert.equal(transpile(text,removeComments),transpile(after,removeComments),'JavaScript changed');
 return {text:after,edits:1};
}
function adapt(tree){
 const file=path.join(tree,'src/compiler/utilities.ts'),before=fs.readFileSync(file,'utf8'),result=plan(before);
 assert.equal(fs.readFileSync(file,'utf8'),before,'concurrent source change');
 if(result.edits)fs.writeFileSync(file,result.text);
 return {files:result.edits,edits:result.edits,javascriptIdentical:true};
}
module.exports={adapt,plan,owner,hash,transpile,pin};
if(require.main===module){assert.equal(process.argv.length,3);console.log(JSON.stringify(adapt(process.argv[2])));}
