// Source declarations come from the independent TypeScript pin, not cohere.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const pin=process.argv[2],ts=require(process.argv[3]);
const pairs=JSON.parse(fs.readFileSync(path.join(__dirname,'../unknown-callable-pairs-ranked.json')));
const recipes=new Map([
 [228,["Map", "Path, FileInfo", "new Map<Path,FileInfo>()", "\"ok\""]],
 [231,["Map", "string, RelationComparisonResult", "new Map<string,RelationComparisonResult>()", "\"ok\",7"]],
 [261,["ReadonlyArray", "Statement", "[]", "0"]],
 [267,["Date", "", "new Date(0)", ""]],
 [270,["Array", "DiagnosticMessageChain", "[]", "{value:7}"]],
 [279,["Array", "JSDocComment", "[]", "{value:7}"]],
 [282,["Map", "number, Type", "new Map<number,Type>()", "7"]],
 [312,["Array", "Statement", "[]", "{value:7}"]],
 [318,["Array", "TypeNode", "[]", "{value:7}"]],
 [321,["ReadonlyArray", "string", "[\"ok\"]", "0"]],
 [330,["Array", "DiagnosticRelatedInformation", "[]", "{value:7}"]],
 [15,['Array','Expression','[]','{value:7}']], [27,['Array','Diagnostic','[]','{value:7}']],
 [30,['Array','Type','[]','{value:7}']], [33,['RegExp','','/ok/','"ok"']],
 [45,['StringConstructor','','String','65']], [48,['Array','any','[]','7']],
 [57,['Array','Declaration','[{value:7}]','(value:Declaration):boolean=>true']],
 [66,['Map','string, boolean','new Map<string,boolean>()','"ok",true']],
 [87,['JSON','','JSON','{value:7}']], [93,['Map','string, number','new Map<string,number>()','"ok",7']],
 [105,['Map','__String, Symbol','new Map<__String,Symbol>()','"ok"']],
 [108,['Array','TransformerFactory<SourceFile | Bundle>','[]','(context:TransformationContext):Transformer<SourceFile|Bundle>=>(node:SourceFile|Bundle):SourceFile|Bundle=>node']],
 [111,['Map','string, Type','new Map<string,Type>()','"ok",{value:7}']],
 [114,['Array','ParameterDeclaration','[]','{value:7}']],
 [123,['Array','Symbol','[]','{value:7}']], [126,['Array','string','["ok"]','(value:string):string=>value']],
 [138,['Array','VariableDeclaration','[]','{value:7}']], [147,['Array','Node','[]','{value:7}']],
 [150,['ReadonlyArray','string','["ok"]','(value:string):string=>value']],
 [153,['Map','string, Type','new Map<string,Type>()','"ok"']],
 [168,['Array','string','["ok"]','"ok"']], [171,['Map','string, string','new Map<string,string>()','"ok"']],
 [198,['Array','string','[]','"ok"']], [204,['Map','string, string','new Map<string,string>()','"ok","ok"']],
 [222,['Array','string','["ok"]','']],
]);
const sources=['src/lib/es5.d.ts','src/lib/es2015.collection.d.ts','src/lib/es2015.core.d.ts','src/lib/es2016.array.include.d.ts'].map(file=>ts.createSourceFile(file,fs.readFileSync(path.join(pin,file),'utf8'),ts.ScriptTarget.Latest,true));
const rows=[];
for(const [rank,[container,typeArgs,producer,args]] of recipes){
 const pair=pairs.find(p=>p.rank===rank);const declarations=[];let original;
 for(const sf of sources)for(const n of sf.statements)if(ts.isInterfaceDeclaration(n)&&n.name.text===container)for(const m of n.members)if(m.name?.getText(sf)===pair.field){declarations.push(m);original=sf;}
 if(!declarations.length)throw Error('missing intrinsic declaration '+rank);
 if(declarations.some(d=>d.getSourceFile()!==original))throw Error('split declaration '+rank);
 const sf=ts.createSourceFile(pair.witness.file,fs.readFileSync(path.join(pin,pair.witness.file),'utf8'),ts.ScriptTarget.Latest,true);let read;
 function visit(n){if(ts.isPropertyAccessExpression(n)&&n.name.text===pair.field){const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));if(lc.line+1===pair.witness.line&&lc.character+1===pair.witness.column)read=n;}ts.forEachChild(n,visit);}visit(sf);
 if(!read){console.log('Unprocessed non-property witness '+rank);continue;}
 const receiver=read.expression.getText(sf),localReceiver=receiver.replace(/!/g,'');if(![126,147,321].includes(rank)&&(!/^[A-Za-z_$][\w$]*(\.[A-Za-z_$][\w$]*)*$/.test(localReceiver)||receiver==='this')){console.log('Unprocessed nonlocal receiver '+rank+': '+receiver);continue;}
 const generics=container==='Map'?'<K,V>':container.endsWith('Array')?'<T>':'';
 const needed=new Set([...typeArgs.matchAll(/\b[A-Z][A-Za-z]+\b/g)].map(m=>m[0]));
 let carriers='';for(const name of needed)if(!['TransformerFactory','Transformer'].includes(name))carriers+=name==='Path'?'type Path=string;\n':name==='RelationComparisonResult'?'type RelationComparisonResult=number;\n':'interface '+name+' {readonly value:number;}\n';
 if(typeArgs.includes('__String'))carriers+='type __String=string;\n';
 if(typeArgs.includes('TransformerFactory'))carriers+='interface TransformationContext {readonly value:number;}\ntype Transformer<T>=(node:T)=>T;\ntype TransformerFactory<T>=(context:TransformationContext)=>Transformer<T>;\n';
 const decl=declarations.map(d=>d.getText(original)).join('\n');
 // The selected lib member remains exact, including generic/rest/overload syntax.
 carriers+='interface Base {readonly '+pair.field+':unknown;}\ninterface Carrier'+generics+' {'+decl+'}\ntype Target=Carrier'+(typeArgs?'<'+typeArgs+'>':'')+';\n';
 const parts=localReceiver.split('.');let bound='value as Target';for(const k of parts.slice(1).reverse())bound='{'+k+':'+bound+'}';
 let binding='const '+parts[0]+'='+bound+';';
 if(rank===126)binding='const text={split:(delimiter:RegExp):Target=>value as Target};';
 if(rank===147)binding='const sourceFile={value:7};const getOrCreateEmitNode=(node:{readonly value:number}):{annotatedNodes:Target}=>({annotatedNodes:value as Target});';
 if(rank===321)binding='const path="ok";const getAccessibleFileSystemEntries=(path:string):{directories:Target}=>({directories:value as Target});';
 const directory='intrinsic-gap-'+rank;fs.mkdirSync(path.join(__dirname,directory),{recursive:true});
 fs.writeFileSync(path.join(__dirname,directory,'good.a'),'// Original intrinsic member and read; actual built-in receiver, reduced adjacent carriers.\n'+carriers+'function probe(value:Base):void {'+binding+read.getText(sf)+'('+args+');console.log("done");}\nprobe('+producer+');\n');
 const hash=file=>crypto.createHash('sha256').update(fs.readFileSync(path.join(pin,file))).digest('hex');
 rows.push({...pair,directory,read:read.getText(sf),declarations:declarations.map(d=>d.getText(original)),containerName:'Carrier',sourceSha:'050880ce59e30b356b686bd3144efe24f875ebc8',declarationFile:original.fileName,declarationSha256:hash(original.fileName),fileSha256:hash(pair.witness.file),utf16Start:read.getStart(sf),utf16End:read.end,stdout:'done\n'});
}
const oldFrontiers=JSON.parse(fs.readFileSync(path.join(__dirname,'intrinsic-frontiers.json'))).members;
const oldControls=JSON.parse(fs.readFileSync(path.join(__dirname,'intrinsic-candidates.json'))).members;
const evidence=new Map([...oldFrontiers,...oldControls].map(m=>[m.rank,m]));
const frontiers=[],controls=[];
for(const row of rows){const old=evidence.get(row.rank);if(!old){frontiers.push(row);continue;}const merged={...old,...row};(old.nativeType||old.runtimeMessage?controls:frontiers).push(merged);}
fs.writeFileSync(path.join(__dirname,'intrinsic-frontiers.json'),JSON.stringify({members:frontiers},null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'intrinsic-candidates.json'),JSON.stringify({members:controls},null,2)+'\n');
console.log('Prepared '+rows.length+' actual intrinsic receiver controls; retained independently measured stops.');
