'use strict';
// Rerun the existing rejected-draft witness against the new fill, changing only its expected sparse equality.
const fs=require('fs'),vm=require('vm'),{createRequire}=require('module');
let source=fs.readFileSync(__dirname+'/dequeue-draft-witness.cjs','utf8')
 .replace("'/dequeue-draft.ts'","'/queue-box-proposed.ts'")
 .replace('assert.notDeepEqual(cases[1].before,cases[1].draft)','assert.deepEqual(cases[1].before,cases[1].draft)')
 .replace("kept:false,reason:'Sparse input changes from returning undefined to an occupancy error'","kept:true,reason:'for...of boxes sparse holes as undefined payloads'");
vm.runInNewContext(source,{require:createRequire(__filename),__dirname,console,process});
