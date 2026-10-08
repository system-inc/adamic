'use strict';
// Instrument a scratch emitted CLI; never change the source or reference goldens.
const fs=require('node:fs'),ts=require('typescript'),crypto=require('node:crypto');
if(ts.version!=='6.0.3')throw Error('requires TypeScript 6.0.3');
const [bundle,manifest]=process.argv.slice(2),before=fs.readFileSync(bundle,'utf8');
const source=ts.createSourceFile(bundle,before,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);
const names=new Set(['convertToObject','isCompilerOptionsValue','parseJsonConfigFileContent','parseJsonConfigFileContentWorker','canJsonReportNoInputFiles','parseOwnConfigOfJson','convertCompileOnSaveOptionFromJson','convertCompilerOptionsFromJson','convertTypeAcquisitionFromJson','convertCompilerOptionsFromJsonWorker','convertTypeAcquisitionFromJsonWorker','convertWatchOptionsFromJsonWorker','convertOptionsFromJson','tryParseJson','readJson','readJsonOrUndefined','readConfigFile','parseConfigFileTextToJson']);
const headers=new Set(['Node','Token','Identifier','Type']);const edits=[],hits={};
const observe=(name,phase,text)=>`globalThis.__publicAny(${JSON.stringify(name)},${JSON.stringify(phase)},(${text}))`;
function add(start,end,text){edits.push({start,end,text});}
function wrapReturns(n,name){function visit(x){if(ts.isFunctionLike(x))return;if(ts.isReturnStatement(x)&&x.expression){const e=x.expression;add(e.getStart(source),e.end,observe(name,'return',e.getText(source)));return;}ts.forEachChild(x,visit);}ts.forEachChild(n.body,visit);}
function visit(n){
 if(ts.isFunctionDeclaration(n)&&n.body&&(names.has(n.name?.text)||headers.has(n.name?.text))){
  const name=n.name.text;hits[name]=(hits[name]||0)+1;
  if(names.has(name)){
   const args=n.parameters.filter(p=>p.name.getText(source)!=='this').map(p=>p.name.getText(source));
   add(n.body.getStart(source)+1,n.body.getStart(source)+1,'\n'+args.map((a,i)=>observe(name,'arg'+i,a)+';').join('\n')+'\n');
   wrapReturns(n,name);
   add(n.body.end-1,n.body.end-1,'\n'+observe(name,'fallthrough','undefined')+';\n');
  }else add(n.body.end-1,n.body.end-1,'\n'+observe('header.'+name,'complete','this')+';\n');
  return; // Selected function bodies are not also patched as generic timer calls.
 }
 if(ts.isPropertyAssignment(n)&&n.name.getText(source)==='getNodeConstructor'&&ts.isArrowFunction(n.initializer)&&!ts.isBlock(n.initializer.body)){
  const e=n.initializer.body;add(e.getStart(source),e.end,observe('allocator.getNodeConstructor','return',e.getText(source)));hits['allocator.getNodeConstructor']=1;return;
 }
 if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&['setTimeout','clearTimeout'].includes(n.expression.name.text)){
  const method=n.expression.name.text;const tag='timer.'+method;
  const text=n.getText(source);add(n.getStart(source),n.end,observe(tag,'return',text));hits[tag]=(hits[tag]||0)+1;return;
 }
 ts.forEachChild(n,visit);
}
visit(source);
for(const name of names)if((hits[name]||0)>1)throw Error('duplicate bundle owner '+name);
if(!hits.convertToObject||!hits.parseJsonConfigFileContentWorker||!hits.tryParseJson)throw Error('missing core CLI owners');
const absent=[...names].filter(name=>!hits[name]);
let text=before;
for(const e of edits.sort((a,b)=>b.start-a.start||b.end-a.end))text=text.slice(0,e.start)+e.text+text.slice(e.end);
if(ts.createSourceFile(bundle,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS).parseDiagnostics.length)throw Error('instrumentation syntax error');
fs.writeFileSync(bundle,text);fs.writeFileSync(manifest,JSON.stringify({bundle,before_sha256:crypto.createHash('sha256').update(before).digest('hex'),after_sha256:crypto.createHash('sha256').update(text).digest('hex'),hits,absent,edits:edits.length},null,2)+'\n');
console.log(JSON.stringify({hits,absent,edits:edits.length}));
