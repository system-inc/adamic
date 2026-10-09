const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const {adapt,rules}=require('./adapt.cjs'),root=path.resolve(process.argv[2]),rows=[];
for(const file of new Set(rules.map(r=>r.file))){const text=fs.readFileSync(path.join(root,file),'utf8'),once=adapt(text,file),twice=adapt(once,file);assert.equal(once,text,'already-adapted input changes: '+file);assert.equal(twice,once,'second run changes: '+file);rows.push({file,bytes:Buffer.byteLength(text),sha256:crypto.createHash('sha256').update(text).digest('hex')});}
console.log(JSON.stringify({idempotent:true,files:rows},null,2));
