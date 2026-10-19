'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const {compiler}=require('../../47-host-errors/adapt.cjs');
const [before,after,output]=process.argv.slice(2).map(x=>path.resolve(x));
function diagnostics(tree){
 const program=compiler(tree,ts);
 return ts.getPreEmitDiagnostics(program).map(d=>({file:d.file?path.relative(tree,d.file.fileName):null,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
}
for(const tree of [before,after]) assert(fs.existsSync(path.join(tree,'node_modules/@types/node/package.json')),'install the locked tree dependencies before comparing diagnostics');
const old=diagnostics(before),current=diagnostics(after),counts=new Map();
for(const d of old){const k=JSON.stringify(d);counts.set(k,(counts.get(k)||0)+1);}
const introduced=[];
for(const d of current){const k=JSON.stringify(d),n=counts.get(k)||0;if(n)counts.set(k,n-1);else introduced.push(d);}
const result={before:old.length,after:current.length,introduced};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
assert.deepEqual(introduced,[],'new stock checker diagnostics');
assert.equal(current.length,old.length,'compiler diagnostics removed');
console.log(JSON.stringify(result));
