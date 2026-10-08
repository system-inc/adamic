// Stock TypeScript declarations and reads, never cohere source.
const fs = require('fs'), path = require('path'), crypto = require('crypto'), cp = require('child_process');
const pin = path.resolve(process.argv[2]), ts = require(path.resolve(process.argv[3]));
const sourceSha = '050880ce59e30b356b686bd3144efe24f875ebc8';
if (cp.execFileSync('git', ['-C', pin, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== sourceSha) throw Error('wrong source pin');
const config = ts.readConfigFile(path.join(pin, 'src/compiler/tsconfig.json'), ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.join(pin, 'src/compiler'));
parsed.options.typeRoots = [path.resolve(__dirname, '../../../api/node_modules/@types')];
const program = ts.createProgram(parsed.fileNames, parsed.options), checker = program.getTypeChecker();
const shareA = path.join(__dirname, '../share-a');
const excluded = new Set(['witnesses.json','needs-code.json'].flatMap(f => JSON.parse(fs.readFileSync(path.join(shareA,f))).members.map(m=>m.rank)));
const pairs = JSON.parse(fs.readFileSync(path.join(shareA,'queue.json'))).pairs.filter(m=>m.status==='pending' && !excluded.has(m.rank)).sort((a,b)=>b.rank-a.rank).slice(0,20);
const primitive = new Set(['string','number','boolean','void','undefined','unknown','never','any','true','false','null','Pick','Exclude','Required','Record']);
const recipes = {
 2817: {args:'"ok",(name:string,eventKind:FileWatcherEventKind):void=>{},250,undefined', returns:'{close:():void=>{}}',result:'FileWatcher',use:'discard',types:'type FileWatcherEventKind=number; type PollingInterval=number; interface WatchOptions {readonly value:number;} interface FileWatcher {close():void;} type FileWatcherCallback=(fileName:string,eventKind:FileWatcherEventKind,modifiedTime?:Date)=>void;'},
 2814: {args:'{value:3}',returns:'[{value:7}]',result:'NodeArray<TypeParameterDeclaration>',use:'array'},
 2811: {args:'{value:3}',returns:'{value:7}',result:'TypeNode',use:'object'},
 2808: {args:'',returns:'():void=>{console.log("done");}',result:'()=>void',use:'callback'},
 2805: {args:'',returns:'true',result:'boolean',use:'scalar'},
 2802: {args:'',returns:'',use:'intrinsic',producer:'[]',types:'interface Expression {readonly value:number;} interface BindingName {readonly value:number;} interface TextRange {readonly value:number;} type Item={pendingExpressions?:Expression[];name:BindingName;value:Expression;location?:TextRange;original?:Node;}; interface Node {readonly value:number;}',typeArgs:'Item'},
 2799: {args:'"ok"',returns:'true',result:'boolean',use:'scalar'},
 2796: {args:'',returns:'',result:'void',use:'discard'},
 2793: {args:'():void=>{},0',returns:'7',result:'number',use:'scalar'},
 2790: {args:'"ok"',returns:'"ok"',result:'string',use:'scalar'},
 2787: {args:'',returns:'"ok"',result:'string',use:'scalar',optionalCall:true},
 2784: {args:'',returns:'',use:'intrinsic',producer:'[]',types:'type ModuleSpecifierEnding=number; type Item={ending:ModuleSpecifierEnding|undefined;value:string;};',typeArgs:'Item'},
 2781: {args:'{}',returns:'true',result:'boolean',use:'scalar',key:'allowSyntheticDefaultImports',dependencies:[]},
 2778: {args:'{}',returns:'true',result:'boolean',use:'scalar',key:'resolvePackageJsonExports',dependencies:['target','module','moduleResolution']},
 2775: {args:'{}',returns:'true',result:'boolean',use:'scalar',key:'preserveConstEnums',dependencies:['isolatedModules','verbatimModuleSyntax']},
 2772: {args:'{}',returns:'true',result:'boolean',use:'scalar',key:'incremental',dependencies:['composite']},
 2769: {args:'"ok"',returns:'"ok"',result:'string',use:'scalar'},
 2766: {args:'',returns:'',result:'void',use:'discard'},
 2763: {args:'0,"ok"',returns:'',result:'void',use:'discard',types:'type Phase=number; interface Args {readonly value:number;}'},
 2760: {args:'"ok",undefined,():void=>{}',returns:'',use:'external',types:'type PathLike=string; interface WatchFileOptions {readonly value:number;} interface StatWatcher {readonly value:number;} type StatsListener=()=>void; type BigIntStatsListener=()=>void;',producer:'(filename:PathLike):StatWatcher=>({value:7})'},
};
const manifest = [];
const decisions = new Map(JSON.parse(fs.readFileSync(path.join(__dirname,'status.json'))).members.map(m=>[m.rank,m]));
const declarationPath = f => { const api=path.resolve(__dirname,'../../../api/node_modules'); return f.startsWith(api+path.sep)?'api:'+path.relative(api,f):path.relative(pin,f); };
const hash = f => crypto.createHash('sha256').update(fs.readFileSync(f)).digest('hex');
for (const pair of pairs) {
 const recipe=recipes[pair.rank]; if(!recipe) throw Error('unprepared bottom rank '+pair.rank);
 const sf=program.getSourceFile(path.join(pin,pair.witness.file));let read;
 function locate(n) {if(ts.isPropertyAccessExpression(n)&&n.name.text===pair.field || ts.isBindingElement(n)&&n.name.getText(sf)===pair.field) {const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));if(lc.line+1===pair.witness.line&&lc.character+1===pair.witness.column) read=n;} ts.forEachChild(n,locate);}locate(sf);
 if(!read) throw Error('missing original read '+pair.rank);
 const proven=checker.getTypeAtLocation(ts.isBindingElement(read)?read.name:read);
 const signatures=checker.getSignaturesOfType(checker.getNonNullableType(proven),ts.SignatureKind.Call);
 if(!signatures.length) throw Error('no original signature '+pair.rank);
 const declarations=signatures.map(s=>s.getDeclaration()), declaration=declarations[0], dsf=declaration.getSourceFile();
 const parameters=declaration.parameters.map(p=>p.getText(dsf));
 const returnType=checker.typeToString(checker.getReturnTypeOfSignature(signatures[0]),declaration,ts.TypeFormatFlags.NoTruncation);
 let target, types=recipe.types||'', producer, signature, mode;
 if(recipe.key) {
  let mapped=declaration.parent;while(mapped && !ts.isMappedTypeNode(mapped)) mapped=mapped.parent;
  const body=mapped.getText(dsf);
  if(!body.includes('computeValue:') || !ts.isMappedTypeNode(mapped)) throw Error('original mapped context absent');
  const deps=recipe.dependencies;
  types+='\ninterface CompilerOptions {'+[recipe.key,...deps].map(k=>k+'?:boolean;').join('')+'}\ntype CompilerOptionKeys=keyof CompilerOptions;type StrictOptionName=never;\ntype OriginalOptions<T extends Record<string,CompilerOptionKeys[]>>='+body+';';
  target='type Target=OriginalOptions<{'+recipe.key+':'+JSON.stringify(deps)+'}>["'+recipe.key+'"];';
  signature='(compilerOptions:Pick<CompilerOptions,'+[recipe.key,...deps].map(k=>'"'+k+'"').join('|')+'>)=>boolean';
  mode='original mapped member and instantiated receiver';
 } else if(recipe.use==='intrinsic') {
  target='interface Carrier<T> {'+declaration.getText(dsf)+'}\ntype Target=Carrier<'+recipe.typeArgs+'>;';
  producer=recipe.producer; mode='original lib member and real array receiver';
 } else if(recipe.use==='external') {
  target='interface Target {'+pair.field+':{'+declarations.map(d=>'('+d.parameters.map(p=>p.getText(d.getSourceFile())).join(', ')+'): '+d.type.getText(d.getSourceFile())+';').join('\n')+'};}';
  producer=recipe.producer;mode='original external overload declarations; reduced adjacent domains';
 } else if(ts.isMethodSignature(declaration)) {
  target='interface Target {'+declaration.getText(dsf)+'}';
  if(declaration.questionToken && !(proven.flags & ts.TypeFlags.Union)) target='interface OriginalHost {'+declaration.getText(dsf)+'}\ntype Target=Required<Pick<OriginalHost,"'+pair.field+'">>;';
  signature='('+parameters.join(',')+')=>'+(declaration.type?.getText(dsf)||returnType); mode='unchanged original method member';
 } else {
  signature=ts.isFunctionTypeNode(declaration)?declaration.getText(dsf):checker.typeToString(checker.getNonNullableType(proven),read,ts.TypeFormatFlags.NoTruncation);
  // Function declarations retain their source parameter and result annotations.
  if(ts.isFunctionDeclaration(declaration)) signature='('+parameters.join(',')+')=>'+declaration.type.getText(dsf);
  if(pair.rank===2796) {signature='AnyFunction';types+='\ntype AnyFunction='+declaration.getText(dsf)+';';}
  if(pair.rank===2817) {signature='HostWatchFile';types+='\ntype HostWatchFile='+declaration.getText(dsf)+';';}
  target='interface Target {'+pair.field+':'+signature+';}';mode='original function signature reified as a field';
 }
 if(!producer) {
  let params;
  if(recipe.key) params='compilerOptions:Pick<CompilerOptions,'+[recipe.key,...recipe.dependencies].map(k=>'"'+k+'"').join('|')+'>';
  else if(ts.isArrowFunction(declaration) && !declaration.parameters.every(p=>p.type)) params='path:string';
  else params=parameters.join(',');
  producer='('+params+'):'+recipe.result+'=>'+(recipe.result==='void'?'{}':'('+recipe.returns+')');
 }
 // Adjacent carriers are reduced; the advertised parameter/result expressions stay intact.
 const needed=ts.createSourceFile('carrier',target+'\n'+signature+'\n'+producer,ts.ScriptTarget.Latest,true);
 const names=new Map();function references(n) {if(ts.isTypeReferenceNode(n)&&ts.isIdentifier(n.typeName)&&!primitive.has(n.typeName.text)) names.set(n.typeName.text,n.typeArguments?.length||0);ts.forEachChild(n,references);}references(needed);
 for(const [name,arity] of names) {
  if(['Target','Carrier','OriginalHost','OriginalOptions','CompilerOptions','CompilerOptionKeys','StrictOptionName','Item','HostWatchFile','AnyFunction'].includes(name) || new RegExp('\\b(?:type|interface|enum)\\s+'+name+'\\b').test(types)) continue;
  if(name==='NodeArray') types+='\ntype NodeArray<T>=readonly T[];';
  else if(/Flags$/.test(name)) types+='\ntype '+name+'=number;';
  else types+='\ninterface '+name+(arity?'<'+Array.from({length:arity},(_,i)=>'T'+i).join(',')+'>':'')+' {readonly value:number;}';
 }
 const expression=read.getText(sf);let binding;
 if(ts.isBindingElement(read)) binding='const {'+expression+'}=value as Target;';
 else {const receiver=read.expression.getText(sf).replace(/!/g,'');const parts=receiver.split('.');if(!parts.every(p=>/^[A-Za-z_$][\w$]*$/.test(p))) throw Error('nonlocal receiver '+pair.rank);let bound='value as Target';for(const key of parts.slice(1).reverse()) bound='{'+key+':'+bound+'}';binding='const '+parts[0]+'='+bound+';';}
 const call=expression+(recipe.optionalCall?'?.':'')+'('+recipe.args+')';
 let use,stdout;
 switch(recipe.use) {
  case 'object':use='const result='+call+';console.log(`${result===undefined?"missing":result.value}`);';stdout='7\n';break;
  case 'array':use='const result='+call+';console.log(`${result===undefined?"missing":result.length}`);';stdout='1\n';break;
  case 'callback':use=call+'();';stdout='done\n';break;
  case 'intrinsic':use=call+';console.log("done");';stdout='done\n';break;
  case 'external':use=call+';console.log("done");';stdout='done\n';break;
  case 'discard':use=call+';console.log("done");';stdout='done\n';break;
  default:use='console.log(`${'+call+'}`);';stdout=recipe.result==='boolean'?'true\n':recipe.result==='number'?'7\n':'ok\n';
 }
 const header='// Original callable declaration and read; adjacent carriers and producer bodies reduced.\n'+types+'\ninterface Base {readonly '+pair.field+':unknown;}\n'+target+'\nfunction probe(value:Base):void {'+binding+use+'}\n';
 const decision=decisions.get(pair.rank);if(!decision) throw Error('unclassified rank '+pair.rank);
 const directory=(decision.status==='certified'?'rank-':'frontier-')+pair.rank;fs.mkdirSync(path.join(__dirname,directory),{recursive:true});
 const result=recipe.result||'void', mutantArity=declaration.parameters.length===0?1:0;
 const wrong='('+(mutantArity?'ignored:number':'')+'):'+result+'=>'+(result==='void'?'{}':'('+recipe.returns+')');
 for(const [name,implementation] of (decision.status==='certified'?[['good',producer],['wrong-arity',wrong],['wrong-value','7']]:[['good',producer]])) fs.writeFileSync(path.join(__dirname,directory,name+'.a'),(header+(recipe.use==='intrinsic'&&name==='good'?'probe('+implementation+');\n':'probe({'+pair.field+':'+implementation+'});\n')).replace(/\r\n/g,'\n'));
 const fixtureName=path.join(__dirname,directory,'good.ts'),content=fs.readFileSync(path.join(__dirname,directory,'good.a'),'utf8');
 const host=ts.createCompilerHost({strict:true,noEmit:true}),getSource=host.getSourceFile;
 host.getSourceFile=(file,...args)=>file===fixtureName?ts.createSourceFile(file,content,ts.ScriptTarget.Latest,true):getSource.call(host,file,...args);
 const fixtureProgram=ts.createProgram([fixtureName],{strict:true,noEmit:true},host),fc=fixtureProgram.getTypeChecker(),f=fixtureProgram.getSourceFile(fixtureName);let fixtureRead;
 function findRead(n) {if(ts.isPropertyAccessExpression(n)&&n.name.text===pair.field || ts.isBindingElement(n)&&n.name.getText(f)===pair.field) fixtureRead=n;ts.forEachChild(n,findRead);}findRead(f);
 const expected=fc.typeToString(fc.getNonNullableType(fc.getTypeAtLocation(ts.isBindingElement(fixtureRead)?fixtureRead.name:fixtureRead)),fixtureRead,ts.TypeFormatFlags.NoTruncation);
 const dsfs=[...new Set(declarations.map(d=>d.getSourceFile()))];
 manifest.push({...pair,directory,read:expression,readKind:ts.SyntaxKind[read.kind],utf16Start:read.getStart(sf),utf16End:read.end,fileSha256:hash(sf.fileName),declarationFiles:dsfs.map(f=>({file:declarationPath(f.fileName),sha256:hash(f.fileName)})),declarations:declarations.map(d=>({file:declarationPath(d.getSourceFile().fileName),start:d.getStart(d.getSourceFile()),end:d.end,text:d.getText(d.getSourceFile())})),originalSignature:checker.typeToString(proven,read,ts.TypeFormatFlags.NoTruncation),targetDeclaration:target,carriers:types,mode,stockExpected:expected,expected:decision.expected||expected,stop:decision.stop,stopPosition:decision.stopPosition,neededChange:decision.neededChange,certified:decision.status==='certified',arity:declaration.parameters.length,mutantArity,stdout});
 console.log('Prepared '+pair.rank+' '+expression+' '+expected);
}
for(const [name,certified] of [['witnesses.json',true],['needs-code.json',false]]) fs.writeFileSync(path.join(__dirname,name),JSON.stringify({sourceSha,baseShareA:'18d1c9da9e1889fa3f042cb6997a6734b404ccf2',basis:'original declarations and reads; adjacent carriers and producer bodies reduced; candidate reads, not original application executions',members:manifest.filter(m=>m.certified===certified)},null,2)+'\n');
