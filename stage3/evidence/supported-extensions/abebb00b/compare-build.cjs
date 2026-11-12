const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');
const [before,after,candidate]=process.argv.slice(2),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const all=root=>fs.readdirSync(root,{recursive:true}).filter(n=>fs.statSync(path.join(root,n)).isFile()).sort();
const names=all(before+'/src');assert.deepEqual(names,all(after+'/src'));assert.deepEqual(names,all(candidate+'/src'));
const changes=[];for(const n of names){const a=fs.readFileSync(before+'/src/'+n),b=fs.readFileSync(after+'/src/'+n);assert(b.equals(fs.readFileSync(candidate+'/src/'+n)),n);if(!a.equals(b))changes.push(n);}assert.deepEqual(changes,['compiler/utilities.ts']);
const js=fs.readdirSync(before+'/built/local').filter(n=>n.endsWith('.js')).sort();assert.equal(js.length,10);assert.deepEqual(js,fs.readdirSync(after+'/built/local').filter(n=>n.endsWith('.js')).sort());
const identical=n=>{const a=fs.readFileSync(before+'/built/local/'+n),b=fs.readFileSync(after+'/built/local/'+n);assert(a.equals(b),n);return {file:n,bytes:a.length,sha256:hash(a)};};
console.log(JSON.stringify({changedSources:changes,fullApplyMatchesCandidate:true,outputs:js.map(identical),publicDeclarations:['typescript.d.ts','tsserverlibrary.d.ts'].map(identical)},null,2));
