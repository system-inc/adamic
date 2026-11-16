'use strict';
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),assert=require('node:assert/strict');
const ts=require('/workspace/scratch/cast-tags/work/stage3/api/node_modules/typescript');
const base='/workspace/scratch/cast-tags/main-tree',out='/workspace/scratch/cast-tags/c-tree',evidence='/workspace/scratch/cast-tags/evidence';
assert(!fs.existsSync(out)); fs.mkdirSync(out); fs.cpSync(path.join(base,'src'),path.join(out,'src'),{recursive:true});
const groups={'transformers/generators.ts':['ExceptionBlock','LabeledBlock','SwitchBlock','LoopBlock','WithBlock'],'types.ts':['ThisTypePredicate','IdentifierTypePredicate','AssertsThisTypePredicate','AssertsIdentifierTypePredicate']};
const pins=[];
for(const [file,names] of Object.entries(groups)){
 const name=path.join(out,'src/compiler',file),raw=fs.readFileSync(name,'utf8'),crlf=raw.includes('\r\n'),text=raw.replace(/\r\n/g,'\n');assert(!crlf||!raw.replace(/\r\n/g,'').includes('\n'));
 const src=ts.createSourceFile(file,text,99,true);assert.equal(src.parseDiagnostics.length,0);const edits=[];
 for(const owner of names){const found=src.statements.filter(n=>ts.isInterfaceDeclaration(n)&&n.name.text===owner);assert.equal(found.length,1);const n=found[0],before=n.getText(src);assert.equal((before.match(/^    kind:/gm)||[]).length,1);const after=before.replace(/^    kind:/m,'    readonly kind:');const digest=t=>crypto.createHash('sha256').update(t).digest('hex');pins.push({file,owner,before,after,ownerBeforeSha256:digest(before),ownerAfterSha256:digest(after)});edits.push({start:n.getStart(src),end:n.end,after});}
 let after=text;for(const e of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,e.start)+e.after+after.slice(e.end);for(const removeComments of [false,true]){const emit=t=>ts.transpileModule(t,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;assert.equal(emit(text),emit(after));}fs.writeFileSync(name,crlf?after.replace(/\n/g,'\r\n'):after);
}
fs.writeFileSync(path.join(evidence,'c-probe-owners.json'),JSON.stringify(pins,null,2)+'\n');console.log('Probe only: five CodeBlock and four TypePredicate member kinds readonly; edited-file JavaScript identical with comments kept and removed.');
