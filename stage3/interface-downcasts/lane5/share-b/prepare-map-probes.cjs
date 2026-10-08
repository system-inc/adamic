// Actual Map receivers retain the full original generic member declarations.
const fs=require('fs'),path=require('path');
const members=require('./batch-03-map-original.json').members;
const probes=[];
for(const m of members){
 const object=m.rank===343;
 const generics=object?'number,Type':m.rank===112||m.rank===346?'string,boolean':m.rank===283?'string,number':'string,string';
 const carriers=object?'interface Type {readonly value:number;}\n':'';
 const receiver=m.read.slice(0,m.read.lastIndexOf('.'));
 const args=m.field==='clear'?'':m.field==='set'?'3,{value:3}':'"value3"';
 const call=m.read+'('+args+')';
 const observation=m.field==='clear'?call+';console.log("cleared");':m.field==='set'?call+';console.log("stored");':'console.log(`${'+call+'}`);';
 const initial=object?'map.set(3,{value:3});':m.rank===283?'map.set("value3",3);':m.rank===232?'map.set("value3","answer3");':'map.set("value3",true);';
 const scaffold=receiver.includes('.') ? 'const '+receiver.split('.')[0]+'={'+receiver.split('.')[1]+':value as Target<'+generics+'>};' : 'const '+receiver+'=value as Target<'+generics+'>;';
 const filename='rank-'+m.rank+'/actual-map.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,filename),'// Original member/read with an actual Map receiver.\n'+carriers+'interface Base {readonly '+m.field+':unknown;}\ninterface Target<K,V> {'+m.declaration+'}\nfunction probe(value:Base):void {'+scaffold+observation+'}\nconst map=new Map<'+generics+'>();'+initial+'probe(map);\n');
 probes.push({...m,filename,output:m.field==='clear'?'cleared\n':m.field==='set'?'stored\n':'true\n'});
}
fs.writeFileSync(path.join(__dirname,'batch-03-map-probes.json'),JSON.stringify(probes,null,2)+'\n');
