// Original generic declarations are instantiated for actual array receivers.
const fs=require('fs'),path=require('path');
const probes=[];
for(const m of require('./batch-03-array-original.json').members){
 const scalar=m.rank===259?'number':m.rank===499?'string':m.rank===184?'Statement':'Declaration';
 const carriers=['Statement','Declaration'].includes(scalar)?'interface '+scalar+' {readonly value:number;}\n':'';
 const receiver=m.read.slice(0,m.read.lastIndexOf('.')).replace(/\?$/, '');
 const scaffold=receiver.includes('.')?'const '+receiver.split('.')[0]+'={'+receiver.split('.')[1]+':value as Target<'+scalar+'>};':'const '+receiver+'=value as Target<'+scalar+'>;';
 const args=m.field==='pop'?'':m.field==='includes'?'"first",0':'0,1';
 const initial=scalar==='number'?'[3,4]':scalar==='string'?'["first","second"]':'[{value:3},{value:4}]';
 const filename='rank-'+m.rank+'/actual-array.a';fs.mkdirSync(path.join(__dirname,'rank-'+m.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,filename),'// Original member/read with an actual array receiver.\n'+carriers+'interface Base {readonly '+m.field+':unknown;}\ninterface Target<T> {'+m.declaration+'}\nfunction probe(value:Base):void {'+scaffold+m.read+'('+args+');console.log("called");}\nprobe('+initial+');\n');
 probes.push({...m,filename,output:'called\n'});
}
fs.writeFileSync(path.join(__dirname,'batch-03-array-probes.json'),JSON.stringify(probes,null,2)+'\n');
