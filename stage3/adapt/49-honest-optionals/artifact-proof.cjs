const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const [reference,candidate,referenceOracle,candidateOracle]=process.argv.slice(2);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function files(root){const names=[];function visit(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const p=path.join(dir,e.name);if(e.isDirectory())visit(p);else if(e.isFile())names.push(path.relative(root,p));}}visit(root);return names.sort();}
function compare(a,b,names,label){const rows=[];for(const n of names){const x=fs.readFileSync(path.join(a,n)),y=fs.readFileSync(path.join(b,n));assert.deepEqual(y,x,label+': '+n);rows.push({file:n,bytes:x.length,sha256:hash(x)});}return rows;}
const a=path.join(reference,'built/local'),b=path.join(candidate,'built/local'),js=fs.readdirSync(a).filter(n=>n.endsWith('.js')).sort();assert.deepEqual(fs.readdirSync(b).filter(n=>n.endsWith('.js')).sort(),js);assert.equal(js.length,10);
const javascript=compare(a,b,js,'emitted JavaScript'),api=compare(a,b,['typescript.d.ts'],'public API');
const x=path.join(reference,'tests/baselines/local'),y=path.join(candidate,'tests/baselines/local'),names=files(x);assert.deepEqual(files(y),names,'baseline inventory');const baselines=compare(x,y,names,'baseline bytes');
const old=JSON.parse(fs.readFileSync(path.join(referenceOracle,'report.json'))),next=JSON.parse(fs.readFileSync(path.join(candidateOracle,'report.json')));for(const k of ['counts','status','baseline_diffs','runners','tests','workers'])assert.deepEqual(next[k],old[k],k);assert.deepEqual(fs.readFileSync(path.join(referenceOracle,'baseline.diff')),fs.readFileSync(path.join(candidateOracle,'baseline.diff')),'baseline diff bytes');
const mutants=[];
for(const [name,relative,label] of [['javascript-byte','built/local/_tsc.js','emitted JavaScript'],['api-byte','built/local/typescript.d.ts','public API'],['baseline-byte','tests/baselines/local/api/typescript.d.ts','baseline bytes']]){
 const target=path.join(candidate,relative),original=fs.readFileSync(target);try{fs.appendFileSync(target,'\n');assert.throws(()=>compare(path.dirname(path.join(reference,relative)),path.dirname(target),[path.basename(target)],label),new RegExp(label));mutants.push({name,file:relative,caught:label});}finally{fs.writeFileSync(target,original);}
}
console.log(JSON.stringify({javascript,api,baseline_files:baselines.length,baseline_manifest_sha256:hash(Buffer.from(JSON.stringify(baselines))),oracle_counts:next.counts,baseline_diffs:next.baseline_diffs,mutants},null,2));
