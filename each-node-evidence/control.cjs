const assert=require('node:assert/strict');const {classify}=require('./preload.cjs');
assert(classify('(n, i) => true').canSeeSecond);assert(classify('(n, i = 0) => true').canSeeSecond);assert(classify('(...n) => true').canSeeSecond);assert(classify('function(n) { return arguments[1]; }').canSeeSecond);assert(!classify('n => true').canSeeSecond);
globalThis.__eachNodeObserve((n,i)=>true,'control')({},0);
console.log('control: one exposed call; default, rest and arguments controls pass');
