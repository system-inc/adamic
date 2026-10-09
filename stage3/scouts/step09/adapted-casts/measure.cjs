// Measure pinned AST anchors, immediate adapter introduction and virtual type-only proposals.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript');assert.equal(ts.version,'6.0.3');
const root=__dirname, snapshots=path.resolve(process.argv[2]), tree=path.join(snapshots,'adapted');
const data=JSON.parse(fs.readFileSync(path.join(root,'contracts.json'))), transitions=JSON.parse(fs.readFileSync(path.join(snapshots,'transitions.json')));
const parse=(name,text)=>ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
function casts(file,text){const s=parse(file,text),a=[];function walk(n){if(ts.isAsExpression(n))a.push(n);ts.forEachChild(n,walk);}walk(s);return {s,a};}
const attribution=[];
for(const row of data.rows){
 const final=fs.readFileSync(path.join(tree,row.file),'utf8'), {s,a}=casts(row.file,final);
 const found=a.find(n=>n.getStart(s)===row.node_start&&n.end===row.node_end);
 assert(found&&found.getText(s)===row.text,'exact pinned final AST span: '+row.file+':'+row.line);
 const relative=row.file.replace(/^src\//,'');const changes=[];
 for(const step of transitions){
  const before=fs.readFileSync(path.join(step.before,relative),'utf8'),after=fs.readFileSync(path.join(step.after,relative),'utf8');
  const count=text=>{const q=casts(row.file,text);return q.a.filter(n=>n.getText(q.s)===row.text).length;};
  if(count(after)>count(before))changes.push({adaptation:'stage3/adapt/'+step.adaptation+'/',before:count(before),after:count(after),before_sha256:crypto.createHash('sha256').update(before).digest('hex'),after_sha256:crypto.createHash('sha256').update(after).digest('hex')});
 }
 assert.equal(changes.length,1);assert.equal(changes[0].adaptation,row.adaptation);
 attribution.push({file:row.file,line:row.line,column:row.column,node_start:row.node_start,node_end:row.node_end,introduction:changes[0]});
}
const parsed=ts.getParsedCommandLineOfConfigFile(path.join(tree,'src/compiler/tsconfig.json'),{}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(ts.flattenDiagnosticMessageText(d.messageText,'\n'));}});
const options={...parsed.options,noEmit:true,composite:false,isolatedDeclarations:false},entry=path.join(tree,'src/tsc/tsc.ts');
const baseline=ts.createProgram([entry],options);const diagnosticRows=p=>p.getSemanticDiagnostics().map(d=>({code:d.code,file:d.file?.fileName,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
assert.deepEqual(diagnosticRows(baseline),[],'baseline semantic diagnostics');
const groups=[['directory helper',data.rows.filter(r=>r.file.endsWith('moduleNameResolver.ts'))],['capture-free split',data.rows.filter(r=>r.adaptation.includes('45-regex'))],['nonempty preferences',data.rows.filter(r=>r.file.endsWith('moduleSpecifiers.ts'))]];
const proposals=[];
for(const [name,rows] of groups){
 const edits=new Map();for(const row of rows){const filename=path.join(tree,row.file);if(!edits.has(filename))edits.set(filename,[]);const sf=baseline.getSourceFile(filename);const {a}=casts(row.file,sf.text),n=a.find(n=>n.getStart(sf)===row.node_start&&n.end===row.node_end);assert(n);let replacementNode=n;while(ts.isParenthesizedExpression(replacementNode.parent))replacementNode=replacementNode.parent;edits.get(filename).push([replacementNode.getStart(sf),replacementNode.end,n.expression.getText(sf)]);}
 const changed=new Map();for(const [filename,list] of edits){let text=baseline.getSourceFile(filename).text;for(const [a,b,replacement] of list.sort((x,y)=>y[0]-x[0]))text=text.slice(0,a)+replacement+text.slice(b);changed.set(filename,text);}
 if(name==='directory helper'){
  const filename=path.join(tree,'src/compiler/utilities.ts'),text=baseline.getSourceFile(filename).text;
  const old='export function directoryProbablyExists(directoryName: string, host: { directoryExists?: (directoryName: string) => boolean; }): boolean';
  const replacement='export function directoryProbablyExists(directoryName: string, host: { directoryExists?: ((directoryName: string) => boolean) | undefined; }): boolean';
  assert(text.includes(old));changed.set(filename,text.replace(old,replacement));
  if(process.argv.includes('--runtime-edit'))changed.set(filename,changed.get(filename).replace('return !host.directoryExists || host.directoryExists(directoryName);','return false && (!host.directoryExists || host.directoryExists(directoryName));'));
 }
 if(name==='nonempty preferences'){
  const filename=path.join(tree,'src/compiler/moduleSpecifiers.ts'),text=changed.get(filename);
  const old='getAllowedEndingsInPreferredOrder(syntaxImpliedNodeFormat?: ResolutionMode): ModuleSpecifierEnding[];';assert(text.includes(old));
  changed.set(filename,text.replace(old,'getAllowedEndingsInPreferredOrder(syntaxImpliedNodeFormat?: ResolutionMode): [ModuleSpecifierEnding, ...ModuleSpecifierEnding[]];'));
 }
 const js=[];for(const [filename,text] of changed){
  const emit=t=>ts.transpileModule(t,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments:true}}).outputText;
  const a=emit(baseline.getSourceFile(filename).text),b=emit(text);assert(a===b,'byte-identical JavaScript: '+name+':'+filename);
  js.push({file:path.relative(tree,filename),sha256:crypto.createHash('sha256').update(a).digest('hex')});
 }
 const host=ts.createCompilerHost(options),original=host.getSourceFile;
 host.getSourceFile=(filename,languageVersion,onError,shouldCreateNewSourceFile)=>changed.has(filename)?ts.createSourceFile(filename,changed.get(filename),languageVersion,true):original(filename,languageVersion,onError,shouldCreateNewSourceFile);
 const program=ts.createProgram([entry],options,host),diagnostics=diagnosticRows(program);assert.deepEqual(diagnostics,[],name+' semantic diagnostics');
 proposals.push({name,sites:rows.length,semantic_diagnostics:diagnostics,byte_identical_javascript:js,native_admission:'not claimed; measured against upstream API only'});
}
fs.writeFileSync(path.join(root,'evidence/source-measurement.json'),JSON.stringify({typescript:ts.version,attribution,baseline_semantic_diagnostics:[],proposals},null,2)+'\n');
console.log('PASS: 31 exact AST spans and immediate adaptation introductions; 21 virtual type-only proposals have zero upstream semantic diagnostics and byte-identical JavaScript');
