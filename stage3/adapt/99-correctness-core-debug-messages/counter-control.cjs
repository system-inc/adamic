'use strict';
const assert=require('node:assert/strict');const ts=require(process.argv[2]);
const node={kind:ts.SyntaxKind.Identifier};ts.Debug.setAssertionLevel(1);
for(const name of ['assertEachNode','assertNode','assertNotNode','assertOptionalNode']){
 let threw=false;try{ts.Debug[name](name==='assertEachNode'?[node]:node,name==='assertNotNode'?function testTrue(){return true}:function testFalse(){return false});}catch(e){assert.equal(e.constructor,Error);assert(e.message.startsWith('Debug Failure. False expression: Unexpected node.'));threw=true;}
 assert(threw,name);
}console.log('all four instrumented built callbacks forced to fail');
