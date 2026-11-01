const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('/workspace/adamic/stage3/api/node_modules/typescript');
const {adapt} = require('/workspace/adamic/stage3/adapt/99-correctness-axolotl-e/adapt.cjs');
const tree='/workspace/scratch/99e-slim-after';
const file=path.join(tree,'src/compiler/transformers/declarations.ts');
const before=fs.readFileSync(file);
let writes=0;
const original=fs.writeFileSync;
fs.writeFileSync=()=>{writes++; throw new Error('unexpected write on second run');};
let result;
try { result=adapt(tree,ts); } finally { fs.writeFileSync=original; }
assert.deepEqual(result,{files:0,edits:0,javascriptChanging:true});
assert.equal(writes,0);
assert.deepEqual(fs.readFileSync(file),before);
console.log(JSON.stringify({result,writeCalls:writes,bytesUnchanged:true},null,2));
