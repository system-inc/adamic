const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const dir=process.argv[2],changes=[];
for(const name of fs.readdirSync(dir).filter(x=>x.endsWith('.js'))){
 const file=path.join(dir,name),text=fs.readFileSync(file,'utf8');
 const start=text.indexOf('function dumpTypes(types) {');if(start<0)continue;
 const end=text.indexOf('function dumpLegend()',start);assert(end>start);
 const body=text.slice(start,end),needle='const objectFlags = type.objectFlags;';assert.equal(body.split(needle).length,2,name);
 const after=body.replace('function dumpTypes(types) {','function dumpTypes(types) { globalThis.__dumpStart(types.length);').replace(needle,needle+'\n      globalThis.__dumpRecord(type, getObjectFlags(type), type.flags & 402431 ? type.intrinsicName : undefined);').replace('intrinsicName: type.intrinsicName,','intrinsicName: globalThis.__dumpIntrinsic(type),');
 fs.writeFileSync(file,text.slice(0,start)+after+text.slice(end));changes.push(name);
}
assert(changes.length);console.log(JSON.stringify(changes));
