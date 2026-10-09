import {pathToFileURL} from 'node:url';
const tree=process.argv[2];const p=await import(pathToFileURL(tree+'/src/compiler/parser.ts'));
const {ScriptTarget,ScriptKind,JSDocParsingMode,SyntaxKind}=await import(pathToFileURL(tree+'/src/compiler/types.ts'));
const text='/** @param { } x */\nfunction f(x) {}';
function docs(file){let count=0;const pending=[file];while(pending.length){const n=pending.pop();for(const doc of n.jsDoc??[]){count++;pending.push(doc);}p.forEachChild(n,c=>{pending.push(c);});}return {attached:count,jsdoc_diagnostics:file.jsDocDiagnostics?.map(d=>d.code)??[]};}
const probes=[];
for(const kind of [ScriptKind.TS,ScriptKind.JS]){
 const file=p.createSourceFile(kind===ScriptKind.JS?'f.js':'f.ts',text,ScriptTarget.Latest,false,kind);probes.push({entry:'createSourceFile',kind:kind===ScriptKind.JS?'JS':'TS',...docs(file)});
}
const isolated=p.parseIsolatedJSDocComment('/** @param { } x */');probes.push({entry:'parseIsolatedJSDocComment',kind:isolated.jsDoc.kind,diagnostics:isolated.diagnostics.map(d=>d.code)});
const type=p.parseJSDocTypeExpressionForTests('{ }');probes.push({entry:'parseJSDocTypeExpressionForTests',kind:type.jsDocTypeExpression.kind,diagnostics:type.diagnostics.map(d=>d.code)});
for(const mode of ['ParseAll','ParseNone','ParseForTypeErrors','ParseForTypeInfo']){
 const file=p.createSourceFile('f.json','/** @see Target */\n{"x":1}',{languageVersion:ScriptTarget.Latest,jsDocParsingMode:JSDocParsingMode[mode]},false,ScriptKind.JSON);probes.push({entry:'createSourceFile',kind:'JSON',mode,...docs(file)});
}
const original=p.createSourceFile('f.js','/** @param {string} x */\nfunction f(x) {}',{languageVersion:ScriptTarget.Latest,jsDocParsingMode:JSDocParsingMode.ParseNone},true,ScriptKind.JS);
const updated=p.updateSourceFile(original,'/** @param {number} x */\nfunction f(x) {}',{span:{start:12,length:6},newLength:6});probes.push({entry:'updateSourceFile',mode:'ParseNone retained',...docs(updated),mode_value:updated.jsDocParsingMode});
if(probes[0].attached!==1||probes[0].jsdoc_diagnostics.length||probes[1].jsdoc_diagnostics[0]!==1110||isolated.diagnostics[0].code!==1110||type.diagnostics[0].code!==1110||probes.at(-1).attached!==0)throw Error('entry point assertion failed');
const {getJSDocTags}=await import(pathToFileURL(tree+'/src/compiler/utilitiesPublic.ts'));
const eager=p.createSourceFile('tags.js','/** @param {string} x */\nfunction f(x) {}',ScriptTarget.Latest,true,ScriptKind.JS);const node=eager.statements[0];const before=node.jsDoc.jsDocCache!==undefined;const tags=getJSDocTags(node);probes.push({entry:'getJSDocTags',lazy:'tag cache only',cache_before:before,cache_after:node.jsDoc.jsDocCache!==undefined,tags:tags.length});
const skippedTags=getJSDocTags(updated.statements[0]);probes.push({entry:'getJSDocTags',mode:'ParseNone',tags:skippedTags.length,attached_after:docs(updated).attached});
if(before||tags.length!==1||skippedTags.length!==0||docs(updated).attached!==0)throw Error('lazy cache assertion failed '+JSON.stringify(probes.slice(-2)));
console.log(JSON.stringify({pass:true,probes},null,2));
