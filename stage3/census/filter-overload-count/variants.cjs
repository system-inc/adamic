'use strict';
// Virtual declarations only: no edits to the adapted tree, no emit.
const ts=require('../../api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root=path.resolve(process.argv[2]),mode=process.argv[3],out=process.argv[4];
const names=ts.sys.readDirectory(root+'/src',['.ts'],[]),opts={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true};
const host=ts.createCompilerHost(opts),read=host.readFile.bind(host),file=root+'/src/compiler/core.ts';let edits=[];
host.readFile=name=>{let text=read(name);if(name===file&&mode!=='base'){
 const f=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true),decls=f.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='filter'&&!n.body);
 for(const i of [0,4]){const d=decls[i];if(mode==='readonly')edits.push({start:d.type.getStart(f),end:d.type.end,text:d.type.getText(f).replace(/^U\[\]/,'readonly U[]')});else if(mode==='drop')edits.push({start:d.getStart(f),end:d.end,text:''});else throw Error(mode);}
 for(const e of edits.toSorted((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);
 }return text;};
const p=ts.createProgram(names,opts,host),diagnostics=ts.getPreEmitDiagnostics(p).map(d=>{const lc=d.file&&d.file.getLineAndCharacterOfPosition(d.start);return {code:d.code,file:d.file&&path.relative(root,d.file.fileName),line:lc&&lc.line+1,character:lc&&lc.character+1,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};});
fs.writeFileSync(out,JSON.stringify({mode,edits,diagnostics},null,2));console.log(JSON.stringify({mode,diagnostics:diagnostics.length}));
