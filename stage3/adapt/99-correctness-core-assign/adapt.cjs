'use strict';
// Adaptation 40 owns this helper. Share its whole-owner planner and pins.
const path=require('node:path'),assert=require('node:assert/strict');
const adapter=require('../40-explicit-any/hasproperty.cjs');
module.exports=adapter;
if(require.main===module){assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');console.log(JSON.stringify(adapter.adapt(path.resolve(process.argv[2])),null,2));}
