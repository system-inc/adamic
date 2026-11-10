'use strict';
const assert=require('node:assert/strict');const ts=require(process.argv[2]);ts.Debug.setAssertionLevel(1);
if(process.argv[3]==='throws'){
 const thrown=new TypeError('control test throw');let observed;try{ts.Debug.assertEachNode([{kind:ts.SyntaxKind.Identifier}],n=>{throw thrown;});}catch(e){observed=e;}assert.equal(observed,thrown);console.log('throw control: same thrown object');
}else{
 const seen=[];ts.Debug.assertEachNode([{kind:ts.SyntaxKind.Identifier}],(n,i)=>{seen.push(i);return true;});assert.deepEqual(seen,[0]);console.log('built two-parameter control: one call with index 0');
}
