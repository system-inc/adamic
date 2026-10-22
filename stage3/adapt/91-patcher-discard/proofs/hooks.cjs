'use strict';
const fs = require('node:fs'), assert = require('node:assert/strict');
const ts = require(process.argv[2]);
const events = [], factory = {first: 0, second: 0}, allocator = {first: 0, second: 0};
// These are the complete implementations in this isolated witness.
function factoryFirst() { factory.first++; events.push({hook: 'nodeFactory', value: 42}); return 42; }
function factorySecond() { factory.second++; return undefined; }
function allocatorFirst() { allocator.first++; events.push({hook: 'objectAllocator', value: 17}); return 17; }
function allocatorSecond() { allocator.second++; return undefined; }
ts.addNodeFactoryPatcher(factoryFirst);
ts.addNodeFactoryPatcher(factorySecond);
ts.createNodeFactory(ts.NodeFactoryFlags.None);
assert.deepEqual(factory, {first: 1, second: 0});
ts.addObjectAllocatorPatcher(allocatorFirst);
ts.addObjectAllocatorPatcher(allocatorSecond);
// Registration immediately calls each allocator patcher once; reset to isolate forEach.
allocator.first = allocator.second = 0;
events.length = 0;
ts.setObjectAllocator(ts.objectAllocator);
assert.deepEqual(allocator, {first: 1, second: 0});
const result = {node: process.version, factory, allocator,
    observed: [{hook: 'nodeFactory', value: {type: 'number', value: '42'}}, ...events.map(e => ({hook:e.hook,value:{type:'number',value:String(e.value)}}))],
    meaning: 'Both forEach calls retain a truthy callback return and stop before the second callback.'};
fs.writeFileSync(process.argv[3], JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify(result));
