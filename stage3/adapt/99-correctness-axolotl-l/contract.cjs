const fs = require('node:fs'), os = require('node:os'), path = require('node:path'), assert = require('node:assert/strict');
const adapter = require('./adapt.cjs');
const raw = fs.readFileSync(path.join(process.argv[2], 'src/compiler/core.ts'),'utf8');
const dir = fs.mkdtempSync(path.join(os.tmpdir(),'flatten-contract-'));
fs.mkdirSync(path.join(dir,'src/compiler'),{recursive:true});
const file = path.join(dir,'src/compiler/core.ts');
const originalWrite = fs.writeFileSync;
let writes = 0;
fs.writeFileSync = (...args) => { writes++; return originalWrite(...args); };
function seed(text) { originalWrite(file,text); writes=0; }
seed(raw);
assert.deepEqual(adapter.adapt(dir),{files:1,edits:1,javascriptChanging:false});
assert.equal(writes,1);
const after=fs.readFileSync(file,'utf8'), stat=fs.statSync(file,{bigint:true});
writes=0;
assert.deepEqual(adapter.adapt(dir),{files:0,edits:0,javascriptChanging:false});
assert.equal(writes,0); assert.equal(fs.readFileSync(file,'utf8'),after);
assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);
const pin=adapter.owners[0], before=pin.before.replace(/\n/g,'\r\n'), afterOwner=pin.after.replace(/\n/g,'\r\n');
const mutateBefore = (from,to) => raw.replace(before,before.replace(from,to));
const mutateAfter = (from,to) => after.replace(afterOwner,afterOwner.replace(from,to));
const mutations = [
 ['changed owner',mutateBefore('const result = [];','const result = [undefined];')],
 ['partly annotated owner',mutateBefore('const result = [];','const result: T = [];')],
 ['changed adapted owner',mutateAfter('const result: T[] = [];','let result: T[] = [];')],
 ['missing owner',raw.replace('function flatten<T>','function flattened<T>')],
 ['duplicate owner',raw+'\r\n'+before],
 ['moved owner',raw.replace(before,'function outer() {\r\n'+before+'\r\n}')],
 ['parse error',raw+'\nfunction {'],
];
for(const [name,text] of mutations) { seed(text); assert.throws(()=>adapter.adapt(dir),undefined,name); assert.equal(writes,0,name); assert.equal(fs.readFileSync(file,'utf8'),text,name); }
seed(raw+'\r\n// unrelated outside owner\r\n');
assert.equal(adapter.adapt(dir).edits,1);
assert(fs.readFileSync(file,'utf8').endsWith('// unrelated outside owner\r\n'));
seed(raw.replace(/\r\n/g,'\n')); assert.equal(adapter.adapt(dir).edits,1);
assert(!fs.readFileSync(file,'utf8').includes('\r'));
console.log(JSON.stringify({firstRun:{files:1,edits:1,writes:1},secondRun:{files:0,edits:0,writes:0,bytesAndMtimePreserved:true},refusals:mutations.map(([name])=>name),unrelatedEditPreserved:true,crlfAndLfPreserved:true},null,2));
