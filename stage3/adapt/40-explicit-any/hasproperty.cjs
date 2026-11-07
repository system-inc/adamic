'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const owner=require('./hasproperty-owner.json');
const digest=s=>crypto.createHash('sha256').update(s).digest('hex');
function parse(file,text){assert.equal(ts.version,'6.0.3');const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,file+': parse error');return source;}
function transpile(text,removeComments=false){return ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;}
function plan(file,text,phase='final'){
 assert(['early','final'].includes(phase),'unknown hasProperty phase');
 const source=parse(file,text),found=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===owner.owner);assert.equal(found.length,1,'hasProperty: top-level owner count drift');let count=0;function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===owner.owner)count++;ts.forEachChild(n,visit);}visit(source);assert.equal(count,1,'hasProperty: nested or duplicate owner');
 const start=found[0].getStart(source),end=found[0].end,spelling=text.slice(start,end).replace(/\r\n/g,'\n'),hash=digest(spelling);const states=['upstream','before','after'].filter(k=>spelling===owner[k]&&hash===owner[k+'Sha256']);assert.equal(states.length,1,'hasProperty: changed or half-adapted owner');const state=states[0],target=phase==='early'?(state==='upstream'?owner.before:spelling):owner.after,value=text.includes('\r\n')?target.replace(/\n/g,'\r\n'):target;const edit=spelling===target?null:{start,end,value};const result=edit?text.slice(0,start)+value+text.slice(end):text;
 parse(file,result);for(const removeComments of [false,true])assert.equal(transpile(text,removeComments),transpile(result,removeComments),'hasProperty: JavaScript changed');return {state,edit,result};
}
function adapt(tree){const file=path.join(tree,'src/compiler/core.ts'),text=fs.readFileSync(file,'utf8'),p=plan(file,text);assert.equal(fs.readFileSync(file,'utf8'),text,'core changed while planning');if(p.edit)fs.writeFileSync(file,p.result);return {files:p.edit?1:0,edits:p.edit?1:0,javascriptChanging:false};}
module.exports={adapt,plan,parse,transpile,digest,owner,owners:[owner]};
