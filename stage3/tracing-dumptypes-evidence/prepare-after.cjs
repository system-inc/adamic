const fs=require('node:fs'),assert=require('node:assert/strict');
const adapter=require('/workspace/adamic/stage3/adapt/99-correctness-tracing-dumptypes/adapt.cjs');
const before=fs.readFileSync('/workspace/scratch/dp24r5m-base/src/compiler/tracing.ts','utf8');
const expected=fs.readFileSync('/workspace/scratch/dp24r5m-flags-road/src/compiler/tracing.ts','utf8');
assert.equal(adapter.plan(before),expected);assert.equal(adapter.plan(expected),expected);
console.log(JSON.stringify({matchesMeasuredRoad:true,replayIdentical:true}));
