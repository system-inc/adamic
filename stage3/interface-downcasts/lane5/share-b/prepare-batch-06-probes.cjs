// Batch 06 original collection declarations, primitive aliases, and receivers.
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('/tmp/lane5-b-original/lib/typescript.js'),members=require('./batch-06-original.json').members;
const declarations=new Map();
for(const filename of ['src/compiler/types.ts','src/compiler/tsbuildPublic.ts','src/compiler/builder.ts','src/compiler/watchUtilities.ts','src/compiler/moduleNameResolver.ts']) {
 const bytes=fs.readFileSync('/tmp/lane5-b-original/'+filename),source=ts.createSourceFile(filename,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);
 function visit(n){if((ts.isTypeAliasDeclaration(n)||ts.isEnumDeclaration(n))&&n.name)declarations.set(n.name.text,{declaration:n.getText(source).replace(/^export\s+/,'').replace(/^const\s+enum/,'enum').replace(/\r\n/g,'\n').replace(/[ \t]+$/gm,''),sourceFile:filename,sha256:crypto.createHash('sha256').update(bytes).digest('hex')});ts.forEachChild(n,visit);}visit(source);
}
function scaffold(receiver,target) {
 if(receiver==='activeTypeMappersCaches[activeTypeMappersCount]')return 'const activeTypeMappersCount=0;const activeTypeMappersCaches:'+target+'[]=[value as '+target+'];if(activeTypeMappersCaches[activeTypeMappersCount]===undefined){throw new Error("missing");}';
 const parts=receiver.replace(/[?!]$/,'').split('.');let holder='value as '+target;
 for(const part of parts.slice(1).reverse())holder='{'+part+':'+holder+'}';
 return 'const '+parts[0]+'='+holder+';';
}
const mapRanks=[550,553,556,559,562,565,568,571,802,805,808,811,814,817,820,823,826,829,832,835,862,868,871,874,877,880];
const maps=[];
for(const m of members.filter(m=>mapRanks.includes(m.rank))) {
 const match=/^Map<(.+)>/.exec(m.type);if(!match)throw Error('Map type '+m.rank);
 let depth=0,split=-1;for(let i=0;i<match[1].length;i++){const c=match[1][i];if(c==='<')depth++;if(c==='>')depth--;if(c===','&&depth===0){split=i;break;}}
 const key=match[1].slice(0,split),value=match[1].slice(split+1).trim();
 const carrierNames=[];if(key==='Path')carrierNames.push('Path');if(key==='ResolvedConfigFilePath')carrierNames.push('Path','ResolvedConfigFileName','ResolvedConfigFilePath');
 if(value==='EmitSignature'||value==='RelationComparisonResult'||value==='ProgramUpdateLevel'||value==='RedirectsCacheKey')carrierNames.push(value);
 const evidence=carrierNames.map(n=>declarations.get(n));if(evidence.some(x=>!x))throw Error('missing carrier '+m.rank);
 let carriers=evidence.map(x=>x.declaration).join('\n');
 if(key==='THash')carriers+='\ntype THash = string;';
 const names=[...new Set((value+' '+(key==='CompilerOptions'?key:'')).match(/\b[A-Z][A-Za-z0-9_]*\b/g)||[])].filter(n=>!carrierNames.includes(n));
 for(const name of names)carriers+='\ninterface '+name+' {readonly value:number;}';
 const keyValue=key==='number'?'3':key==='CompilerOptions'?'{value:3}':key.includes('Path')?'"key" as '+key:'"key"';
 const datum=value.includes('[]')?'[{value:3}]':value==='EmitSignature'||value==='string'?'"value"':value==='number'?'3':value==='boolean'?'true':value==='ProgramUpdateLevel'?'ProgramUpdateLevel.Update':'{value:3}';
 const args=m.field==='clear'?'':m.field==='forEach'?'(item:'+value+',key:'+key+',map:Map<'+key+','+value+'>):void=>{}':m.field==='set'?keyValue+','+datum:keyValue;
 const filename='rank-'+m.rank+'/actual-map.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,filename),'// Original Map declaration/read; original primitive aliases retained.\n'+carriers+'\ninterface Base {readonly '+m.field+':unknown;}\ninterface Target<K,V> {'+m.declaration+'}\nfunction probe(value:Base):void {'+scaffold(m.read.slice(0,m.read.lastIndexOf('.')),'Target<'+key+','+value+'>')+m.read+'('+args+');console.log("called");}\nprobe(new Map<'+key+','+value+'>());\n');
 const refusal=key.includes('Path')?'a primitive brand member __pathBrand whose type is not void':value==='RedirectsCacheKey'?'a primitive brand member __compilerOptionsKey whose type is not void':m.field==='forEach'?'checked view read of field forEach with unsupported callable contract':'';
 maps.push({...m,filename,output:'called\n',refusal,carrierDeclarations:evidence.map(x=>x.declaration),carrierEvidence:evidence,genericInstantiation:key==='THash'?'THash=string; TElement structural object fixture':undefined});
}
const arrays=[],admissions=[];
for(const m of members.filter(m=>[907,910].includes(m.rank))) {
 const element=m.rank===907?'Expression':'ObjectLiteralElementLike',filename='rank-'+m.rank+'/'+(m.rank===907?'actual-array':'contract-probe')+'.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 const args=m.rank===907?'{value:3},0':'(item:ObjectLiteralElementLike,index:number,array:readonly ObjectLiteralElementLike[]):boolean=>true';
 fs.writeFileSync(path.join(__dirname,filename),'// Original array declaration/read; adjacent element object reduced.\ninterface '+element+' {readonly value:number;}\ninterface Base {readonly '+m.field+':unknown;}\ninterface Target<T> {'+m.declaration+'}\nfunction probe(value:Base):void {'+scaffold(m.read.slice(0,m.read.lastIndexOf('.')),'Target<'+element+'>')+m.read+'('+args+');console.log("called");}\n'+(m.rank===907?'probe([{value:3}]);':'probe({some:7});')+'\n');
 const probe={...m,filename,output:m.rank===907?'called\n':'',refusal:m.rank===910?'checked view read of field some with unsupported callable contract':''};
 (m.rank===907?arrays:admissions).push(probe);
}
fs.writeFileSync(path.join(__dirname,'batch-06-map-probes.json'),JSON.stringify(maps,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'batch-06-array-probes.json'),JSON.stringify(arrays,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'batch-06-signature-probes.json'),JSON.stringify(admissions,null,2)+'\n');
