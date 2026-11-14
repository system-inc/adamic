'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');
const r='/workspace/scratch/diag-read-view-1/repo', s='/workspace/scratch/diag-read-view-1';
const ts=require(r+'/stage3/api/node_modules/typescript');
const dir=r+'/stage3/adapt/99-correctness-binding-views';
const pins=JSON.parse(require('child_process').execFileSync('git',['show','HEAD:stage3/adapt/99-correctness-binding-views/owners.json'],{cwd:r}));
const hash=t=>crypto.createHash('sha256').update(t).digest('hex');
function source(file){return fs.readFileSync(s+'/main-tree/src/compiler/'+file,'utf8').replace(/\r\n/g,'\n');}
function nodes(text,names){const sf=ts.createSourceFile('x.ts',text,ts.ScriptTarget.Latest,true),found=[];function visit(n){if(n.name&&names.includes(n.name.text))found.push(n);ts.forEachChild(n,visit)}visit(sf);if(names[0]==='@import')found.push(...sf.statements.filter(ts.isImportDeclaration).filter(n=>n.moduleSpecifier.text==='./_namespaces/ts.js').flatMap(n=>n.importClause.namedBindings.elements.filter(e=>['Diagnostic','DiagnosticReadView'].includes(e.name.text))));found.sort((a,b)=>a.pos-b.pos);return {sf,found};}
function pin(file,names,transform){const text=source(file),{sf,found}=nodes(text,names);assert(found.length);const before=text.slice(found[0].getStart(sf,true),found.at(-1).end),after=transform(before);assert.notEqual(after,before);const afterCount=names[0]==='@import'?2:nodes(after,names).found.length;pins.push({file,owner:names.join(' + '),names,beforeCount:found.length,afterCount,before,after,ownerBeforeSha256:hash(before),ownerAfterSha256:hash(after)});}
for(const file of ['program.ts','utilitiesPublic.ts','utilities.ts'])pin(file,['@import'],x=>{assert.equal(x,'type Diagnostic');return 'type Diagnostic,\n    type DiagnosticReadView'});
pin('types.ts',['DiagnosticWithLocation','DiagnosticReadView'],x=>x+'\n\n/** @internal */\nexport interface DiagnosticReadView extends Diagnostic {\n    readonly file: SourceFile | undefined;\n    readonly start: number | undefined;\n    readonly length: number | undefined;\n}');
pin('program.ts',['createProgram'],x=>{assert.equal(x.split('getDiagnosticsHelper<T extends Diagnostic>').length,2);return x.replace('getDiagnosticsHelper<T extends Diagnostic>','getDiagnosticsHelper<T extends DiagnosticReadView>')});
pin('utilitiesPublic.ts',['sortAndDeduplicateDiagnostics'],x=>{const publicSignature=x.slice(0,x.indexOf(' {'));return '/** @internal */\n'+publicSignature.replace('extends Diagnostic>','extends DiagnosticReadView>')+';\n'+publicSignature+';\n'+x.replace('extends Diagnostic>','extends DiagnosticReadView>')});
const selected=process.argv.slice(2);
for(const name of selected)pin('utilities.ts',[name],x=>{const body=x.indexOf('{');assert(body>0);return x.slice(0,body).replace(/\bDiagnostic\b/g,'DiagnosticReadView')+x.slice(body)});
fs.writeFileSync(dir+'/owners.json',JSON.stringify(pins.filter(p=>p.file!=='program.ts'),null,2)+'\n');
fs.writeFileSync(r+'/stage3/adapt/95-checker-axolotl-h/diagnostic-owners.json',JSON.stringify(pins.filter(p=>p.file==='program.ts'),null,2)+'\n');
console.log(JSON.stringify(pins.map(({file,owner,beforeCount,afterCount})=>({file,owner,beforeCount,afterCount})),null,2));
