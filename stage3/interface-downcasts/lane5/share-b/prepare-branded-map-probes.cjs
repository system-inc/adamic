// Original key aliases and enum values retain their primitive representation.
const fs=require('fs'),path=require('path');
const ts=require('/tmp/lane5-b-original/lib/typescript.js');
const originalTypes=ts.createSourceFile('types.ts',fs.readFileSync('/tmp/lane5-b-original/src/compiler/types.ts','utf8'),ts.ScriptTarget.Latest,true);
const declarations=new Map();
function scan(n){if((ts.isTypeAliasDeclaration(n)||ts.isEnumDeclaration(n))&&n.name)declarations.set(n.name.text,n.getText(originalTypes).replace(/^export\s+/,'').replace(/^const\s+enum/,'enum').replace(/\r\n/g,'\n'));ts.forEachChild(n,scan);}scan(originalTypes);
const probes=[];
for(const m of require('./batch-04-map-original.json').members){
 const result=m.rank===97?'Symbol':m.rank===229?'CommandLineOption':'ParsedConfig';
 const key=m.rank===97?'__String':m.rank===229?'string':'Path';
 const adjacent=m.rank===97?[declarations.get('__String'),declarations.get('InternalSymbolName'),declarations.get('SymbolTable')]:m.rank===280?[declarations.get('Path')]:[];
 const sourceKey=m.rank===97?'InternalSymbolName.Call':m.rank===280?'"value3" as Path':'"value3"';
 const receiver=m.read.slice(0,m.read.lastIndexOf('.')).replace(/[?!]$/,'');
 const parts=receiver.split('.');
 let holder='value as Target<'+key+','+result+'>';
 for(const part of parts.slice(1).reverse()) holder='{'+part+':'+holder+'}';
 const scaffold='const '+parts[0]+'='+holder+';';
 const filename='rank-'+m.rank+'/actual-map.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,filename),'// Original member/read/key aliases; adjacent result object reduced.\ninterface '+result+' {readonly value:number;}\n'+adjacent.join('\n')+'\ninterface Base {readonly get:unknown;}\ninterface Target<K,V> {'+m.declaration+'}\nfunction probe(value:Base):void {'+scaffold+'const result='+m.read+'('+sourceKey+');if(result!==undefined){console.log(`${result.value}`);}}\nconst map=new Map<'+key+','+result+'>();map.set('+sourceKey+',{value:3});probe(map);\n');
 probes.push({...m,filename,carrierDeclarations:adjacent,carrierSource:{filename:'src/compiler/types.ts',sha256:require('crypto').createHash('sha256').update(fs.readFileSync('/tmp/lane5-b-original/src/compiler/types.ts')).digest('hex')},output:'3\n',refusal:m.rank===280?'a primitive brand member __pathBrand whose type is not void':''});
}
fs.writeFileSync(path.join(__dirname,'batch-04-map-probes.json'),JSON.stringify(probes,null,2)+'\n');
