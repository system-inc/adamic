// Batch 05 original collection declarations and receiver fixtures.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('/tmp/lane5-b-original/lib/typescript.js');
const members=require('./batch-05-original.json').members;
const sources=new Map(),declarations=new Map();
for(const filename of ['src/compiler/types.ts','src/compiler/tsbuildPublic.ts','src/compiler/corePublic.ts']) {
 const bytes=fs.readFileSync('/tmp/lane5-b-original/'+filename),source=ts.createSourceFile(filename,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);
 sources.set(filename,{filename,sha256:crypto.createHash('sha256').update(bytes).digest('hex')});
 function visit(n){if((ts.isTypeAliasDeclaration(n)||ts.isInterfaceDeclaration(n)||ts.isEnumDeclaration(n))&&n.name)declarations.set(n.name.text,{declaration:n.getText(source).replace(/^export\s+/,'').replace(/^const\s+enum/,'enum').replace(/\r\n/g,'\n'),sourceFile:filename,sha256:sources.get(filename).sha256});ts.forEachChild(n,visit);}visit(source);
}
function scaffold(receiver,view) {
 const parts=receiver.replace(/[?!]$/,'').split('.');let holder=view;
 for(const part of parts.slice(1).reverse())holder='{'+part+':'+holder+'}';
 return 'const '+parts[0]+'='+holder+';';
}
const maps=[];
for(const m of members.filter(m=>[337,340,415,418,421,424].includes(m.rank))) {
 const key=m.rank===421?'number':m.rank===424?'string':m.rank===418?'ResolvedConfigFilePath':'Path';
 const value=({337:'BuilderFileEmit',340:'string',415:'FileInfo',418:'true',421:'Symbol',424:'SortedArray<DiagnosticWithLocation>'})[m.rank];
 const carrierNames=key==='ResolvedConfigFilePath'?['Path','ResolvedConfigFileName','ResolvedConfigFilePath']:key==='Path'?['Path']:m.rank===424?['SortedArray']:[];
 const carrierEvidence=carrierNames.map(n=>declarations.get(n));
 const carriers=carrierEvidence.map(x=>x.declaration).join('\n')+(['BuilderFileEmit','FileInfo','Symbol','SortedArray<DiagnosticWithLocation>'].includes(value)?'\ninterface '+(m.rank===424?'DiagnosticWithLocation':value)+' {readonly value:number;}':'');
 const keyValue=key==='number'?'3':key==='string'?'"key"':'"key" as '+key;
 const datum=value==='string'?'"value"':value==='true'?'true':'{value:3}';
 const args=m.field==='set'?keyValue+','+datum:keyValue;
 const receiver=m.read.slice(0,m.read.lastIndexOf('.'));
 const filename='rank-'+m.rank+'/actual-map.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,filename),'// Original declaration, read, key aliases, and collection result carrier.\n'+carriers+'\ninterface Base {readonly '+m.field+':unknown;}\ninterface Target<K,V> {'+m.declaration+'}\nfunction probe(value:Base):void {'+scaffold(receiver,'value as Target<'+key+','+value+'>')+m.read+'('+args+');console.log("called");}\nprobe(new Map<'+key+','+value+'>());\n');
 maps.push({...m,filename,carrierDeclarations:carrierEvidence.map(x=>x.declaration),carrierEvidence,output:'called\n',refusal:key.includes('Path')?'a primitive brand member __pathBrand whose type is not void':''});
}
const arrays=[];
for(const m of members.filter(m=>[577,619,679].includes(m.rank))) {
 const element=m.rank===577?'Expression':m.rank===619?'Path':'string';
 const evidence=m.rank===619?[declarations.get('Path')]:[];
 const carriers=m.rank===577?'interface Expression {readonly value:number;}':evidence.map(x=>x.declaration).join('\n');
 const initial=m.rank===577?'[{value:3},{value:4}]':m.rank===619?'["key" as Path]':'["first","second"]';
 const receiver=m.read.slice(0,m.read.lastIndexOf('.'));
 const filename='rank-'+m.rank+'/actual-array.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 let fixture=scaffold(receiver,'value as Target<'+element+'>')+m.read+'('+(m.rank===577?'0,1':m.rank===679?'"|"':'')+');console.log("called");';
 if(m.rank===679)fixture='class Witness {readonly prerelease:Target<string>;constructor(value:Base){this.prerelease=value as Target<string>;}check():void {'+m.read+'("|");console.log("called");}}\nfunction probe(value:Base):void {new Witness(value).check();}';
 else fixture='function probe(value:Base):void {'+fixture+'}';
 fs.writeFileSync(path.join(__dirname,filename),'// Original declaration/read with an actual array receiver.\n'+carriers+'\ninterface Base {readonly '+m.field+':unknown;}\ninterface Target<T> {'+m.declaration+'}\n'+fixture+'\nprobe('+initial+');\n');
 arrays.push({...m,filename,carrierDeclarations:evidence.map(x=>x.declaration),carrierEvidence:evidence,output:'called\n',refusal:m.rank===619?'a primitive brand member __pathBrand whose type is not void':''});
}
fs.writeFileSync(path.join(__dirname,'batch-05-map-probes.json'),JSON.stringify(maps,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'batch-05-array-probes.json'),JSON.stringify(arrays,null,2)+'\n');
const admissions=[];
for(const m of members.filter(m=>[472,475,544,646,670,673,676].includes(m.rank))) {
 const names=m.rank===472?['Path']:m.rank===475?['Path']:m.rank===616?['SymbolFlags','__String','InternalSymbolName']:[];
 const evidence=names.map(n=>declarations.get(n));
 let carriers=evidence.map(x=>x.declaration).join('\n');
 if(m.rank===472){
  const filename='src/compiler/core.ts',bytes=fs.readFileSync('/tmp/lane5-b-original/'+filename),source=ts.createSourceFile(filename,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);let declaration;
  function visit(n){if(ts.isInterfaceDeclaration(n)&&n.name.text==='MultiMap')declaration=n.getText(source).replace(/^export\s+/,'').replace(/\r\n/g,'\n');ts.forEachChild(n,visit);}visit(source);
  const item={declaration,sourceFile:filename,sha256:crypto.createHash('sha256').update(bytes).digest('hex')};evidence.push(item);carriers+='\n'+declaration+'\ninterface FileIncludeReason {readonly value:number;}';
 }
 if(m.rank===616)carriers+='\ninterface Symbol {readonly value:number;}';
 const generic=[544,646,670,673,676].includes(m.rank);
 const element=({544:'FileReference',646:'Symbol',670:'[message: DiagnosticMessage, ...args: (string | number)[]]',673:'Diagnostic',676:'SourceFile'})[m.rank];
 if(generic)carriers+='\ninterface '+(m.rank===670?'DiagnosticMessage':element)+' {readonly value:number;}';
 const target='Target'+(generic?'<'+element+'>':'');
 const receiver=m.read.slice(0,m.read.lastIndexOf('.'));
 let setup;
 if(m.rank===676)setup='const host={getSourceFiles:():'+target+'=>value as '+target+'};';
 else setup=scaffold(receiver,'value as '+target);
 const args=m.rank===475?'"key" as Path':m.rank===544?'{value:3}':m.rank===670?'[{value:3},"arg"]':m.rank===646?'(value:Symbol,index:number,array:Symbol[]):number=>3':m.rank===673?'(value:Diagnostic,index:number,array:readonly Diagnostic[]):void=>{}':m.rank===676?'(value:SourceFile,index:number,array:readonly SourceFile[]):number=>3':'';
 const filename='rank-'+m.rank+'/contract-probe.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,filename),'// Original callable declaration/read; collection and primitive carriers retained.\n'+carriers+'\ninterface Base {readonly '+m.field+':unknown;}\ninterface Target'+(generic?'<T>':'')+' {'+m.declaration+'}\nfunction probe(value:Base):void {'+setup+m.read+'('+args+');}\nprobe({'+m.field+':7});\n');
 admissions.push({...m,filename,carrierDeclarations:evidence.map(x=>x.declaration),carrierEvidence:evidence,output:'',refusal:m.rank===670?"stage 0 can't lower a tuple element of type string | number yet":[472,475].includes(m.rank)?'a primitive brand member __pathBrand whose type is not void':'checked view read of field '+m.field+' with unsupported callable contract'});
}
fs.writeFileSync(path.join(__dirname,'batch-05-signature-probes.json'),JSON.stringify(admissions,null,2)+'\n');
