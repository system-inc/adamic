'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const [before,after,output]=process.argv.slice(2);
const read=tree=>fs.readFileSync(path.join(tree,'built/local/typescript.d.ts'));
const previous=read(before),actual=read(after);
assert.deepEqual(actual,previous,'adaptation 66 changed the public API');
const report={status:'pass',unchanged:true,sha256:crypto.createHash('sha256').update(actual).digest('hex'),bytes:actual.length};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
