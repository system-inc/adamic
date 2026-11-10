'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const own=require('./adapt.cjs'),road=require('../99-correctness-core-debug-messages/adapt.cjs');
const [beforeTree,afterTree,output]=process.argv.slice(2);const read=(tree,file)=>fs.readFileSync(path.join(tree,file),'utf8');const before=read(beforeTree,'src/compiler/debug.ts'),after=read(afterTree,'src/compiler/debug.ts'),local=read(afterTree,'src/deprecatedCompat/deprecate.ts');
assert.equal(own.pins.before,road.pins.find(p=>p.owner==='assertEachNode').after,'before must be road (b) after byte for byte');assert.equal(own.plan(before).after,after);assert.equal(own.plan(after).edits,0);assert.equal(road.plan(before,local).edits,0);assert.equal(road.plan(after,local).edits,0);
const cases=[];function refuses(name,fn,pattern){assert.throws(fn,pattern);cases.push({name,refused:true});}
refuses('new adapter: body before, import absent',()=>own.plan(before.replace(own.pins.importBefore,'')),/half-adapted every import/);
refuses('new adapter: body after, import present',()=>own.plan(after.replace('import {\r\n','import {\r\n    every,\r\n')),/half-adapted every import/);
refuses('new adapter: loop drift',()=>own.plan(after.replace('let passed = true;','let passed = false;')),/unreviewed assertEachNode body/);
refuses('road (b): finished loop drift',()=>road.plan(after.replace('let passed = true;','let passed = false;'),local),/unreviewed owner: assertEachNode/);
const other=road.pins.find(p=>p.owner==='assertNode');refuses('road (b): finished loop with restored other message',()=>road.plan(after.replace(other.after,other.before),local),/half-adapted failure messages/);
fs.writeFileSync(output,JSON.stringify({passed:true,ownIdempotent:true,roadFinishedRecognition:true,exactBeforePin:true,cases},null,2)+'\n');console.log('exact before pin, both replay states and five refusals pass');
