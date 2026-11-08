'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript'),pins=require('./owners.json');
const normal=t=>t.replace(/\r\n/g,'\n'),hash=t=>crypto.createHash('sha256').update(t).digest('hex');
function parse(file,text){const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,file+': parse error');return source;}
function nodes(root,predicate){const found=[];function visit(n){if(predicate(n))found.push(n);ts.forEachChild(n,visit);}visit(root);return found;}
function one(items,label){assert.equal(items.length,1,label+': missing or duplicate');return items[0];}
function inspect(debugText,callerText){
 const debug=parse(pins.debug.file,debugText),caller=parse(pins.caller.file,callerText);
 const old=nodes(debug,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getFunctionName');assert(old.length<=1,'duplicate Debug helper');
 if(old.length){assert(ts.isModuleBlock(old[0].parent)&&old[0].parent.parent.name.text==='Debug','helper namespace changed');assert.equal(old[0].getText(debug),pins.debug.before,'debug owner drift');assert.equal(old[0].jsDoc?.length||0,0,'unexpected helper JSDoc');}
 const locals=nodes(caller,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getFunctionName');assert(locals.length<=1,'duplicate local helper');
 if(locals.length){assert.equal(locals[0].parent,caller,'helper must be module local');assert.equal(locals[0].getText(caller),pins.local.after,'local helper drift');assert.equal(locals[0].body.getText(caller),pins.body,'local body changed');}
 const fn=one(caller.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='deprecate'&&n.body),'deprecate owner');const fnText=fn.getText(caller);assert([pins.caller.before,pins.caller.after].includes(fnText),'deprecate owner drift');
 const imports=one(caller.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text==='./_namespaces/ts.js'),'compiler import');assert([pins.import.before,pins.import.after].includes(imports.getText(caller)),'import owner drift');
 const states=[old.length===0,locals.length===1,fnText===pins.caller.after,imports.getText(caller)===pins.import.after];assert(states.every(s=>s===states[0]),'half-applied getFunctionName move');
 const moved=states[0];assert.equal(nodes(debug,n=>ts.isIdentifier(n)&&n.text==='getFunctionName').length,moved?0:1,'unexpected debug helper reference');assert.equal(nodes(caller,n=>ts.isIdentifier(n)&&n.text==='getFunctionName').length,moved?2:1,'unexpected compatibility helper reference');
 if(old.length)assert.equal(old[0].body.getText(debug),pins.body,'old body changed');
 return {moved,debug,caller,old:old[0],local:locals[0],fn,imports};
}
function adapt(tree){
 assert.equal(ts.version,'6.0.3');for(const o of [pins.debug,pins.caller,pins.import,pins.local])for(const key of ['before','after'])if(o[key])assert.equal(hash(o[key]),o[key+'Sha256']);assert.equal(hash(pins.body),pins.bodySha256);
 const files=[pins.debug.file,pins.caller.file].map(f=>path.join(tree,f)),raw=files.map(f=>fs.readFileSync(f,'utf8')),nl=raw.map(t=>t.includes('\r\n')?'\r\n':'\n');assert.equal(nl[0],nl[1],'helper body newline styles differ');
 const [debug,caller]=raw.map(normal),found=inspect(debug,caller);if(found.moved)return {files:0,edits:0,javascriptChanging:true};
 const afterDebug=debug.slice(0,found.old.getStart())+debug.slice(found.old.end);
 const edits=[{start:found.fn.getStart(),end:found.fn.end,text:pins.caller.after},{start:found.imports.getStart(),end:found.imports.end,text:pins.import.after},{start:caller.length,end:caller.length,text:'\n'+pins.local.after+'\n'}];let afterCaller=caller;
 for(const edit of edits.sort((a,b)=>b.start-a.start))afterCaller=afterCaller.slice(0,edit.start)+edit.text+afterCaller.slice(edit.end);
 const planned=inspect(afterDebug,afterCaller);assert(planned.moved);const output=[afterDebug,afterCaller].map(t=>nl[0]==='\n'?t:t.replace(/\n/g,'\r\n'));
 const originalRaw=one(nodes(parse(files[0],raw[0]),n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getFunctionName'),'original raw helper');const outputLocal=one(nodes(parse(files[1],output[1]),n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getFunctionName'),'moved helper');
 assert.equal(outputLocal.body.getText(),originalRaw.body.getText(),'body bytes differ across homes');
 assert.equal(outputLocal.body.getText(),pins.body.replace(/\n/g,nl[0]),'body bytes changed');
 // Compare exact body tokens and interior bytes, independent of the leading trivia at body.pos.
 assert(output[1].includes(found.old.body.getText().replace(/\n/g,nl[0])),'original body absent at new home');
 files.forEach((f,i)=>assert.equal(fs.readFileSync(f,'utf8'),raw[i],'concurrent source edit'));files.forEach((f,i)=>fs.writeFileSync(f,output[i]));return {files:2,edits:4,javascriptChanging:true};
}
module.exports={adapt,inspect,pins};
if(require.main===module){try{assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));}catch(error){console.error(error.stack);process.exitCode=1;}}
