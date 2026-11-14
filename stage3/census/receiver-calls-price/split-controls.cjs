const assert=require('assert');require('./split-hooks.cjs');const p=globalThis.__receiverPrice;
const object={method:function method(a){return a;}};const original=object.method;
p.watch(object,'Control',['method']);assert.strictEqual(object.method,original);assert.strictEqual(object.method(42),42);object.method=original;p.define(object,'method',{value:original,writable:true,configurable:true});
p.tick('receiverIdentity');p.tick('receiverLength');p.tick('parserCalls.control');p.tick('transformerCalls.control');p.tick('printerAliases.control');p.tick('printerAliasSentinelIdentity');p.tick('printerEntries');
const s=p.snapshot();assert(s.counts['reads.Control.method']===2&&s.counts['writes.Control.method']===1&&s.counts['defineWrites.method']===1);assert(Object.keys(s.sources).length===1);assert(s.failures.length===0);console.log(JSON.stringify({pass:true,...s},null,2));
