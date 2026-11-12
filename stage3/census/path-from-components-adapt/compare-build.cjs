const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');
const [before,after,out]=process.argv.slice(2),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const js=root=>fs.readdirSync(root+'/built/local').filter(f=>f.endsWith('.js')).sort();assert.deepEqual(js(before),js(after));assert.equal(js(before).length,10);
const rows=[...js(before),'typescript.d.ts','tsserverlibrary.d.ts'].map(file=>{const a=fs.readFileSync(before+'/built/local/'+file),b=fs.readFileSync(after+'/built/local/'+file);assert(a.equals(b),'changed output '+file);return {file,bytes:a.length,beforeSha256:hash(a),afterSha256:hash(b),identical:true}});
fs.writeFileSync(out,JSON.stringify({javascriptOutputs:10,publicDeclarations:2,rows},null,2)+'\n');console.log('10 JavaScript outputs and both public declarations byte-identical');
