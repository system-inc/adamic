'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),cp=require('node:child_process');
const tree=path.resolve(process.argv[2]);
const names=['core.ts','performanceCore.ts','tracing.ts'];
const before=names.map(n=>fs.readFileSync(path.join(tree,'src/compiler',n)));
const command=[path.join(__dirname,'adapt.cjs'),tree];
const result=cp.spawnSync(process.execPath,command,{encoding:'utf8'});
assert.equal(result.status,0,result.stderr);
names.forEach((n,i)=>assert.deepEqual(fs.readFileSync(path.join(tree,'src/compiler',n)),before[i],'idempotence'));
process.argv=[process.argv[0],command[0],tree];
const {unchanged}=require('./adapt.cjs');
const core=before[0].toString();
const mutant=core.includes('!process.browser') ? core.replace('!process.browser','!!process.browser') : core.replace('!(process as any).browser','!!(process as any).browser');
assert.notEqual(mutant,core,'browser probe missing');
let caught=false;
try { unchanged(core,mutant); } catch(e) { caught=/runtime JavaScript changed/.test(e.message); }
assert(caught,'runtime-change mutant escaped');
console.log(JSON.stringify({idempotent:true,runtime_change_mutant_caught:true}));
