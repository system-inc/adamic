'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),child=require('node:child_process');
function observe(file){
 const code=`const ts=require(process.argv[1]);const width=ts.sys.getWidthOfTerminal();const tty=ts.sys.writeOutputIsTTY();const tag=x=>x===undefined?'undefined':x;console.log(JSON.stringify({node:process.version,width:tag(width),tty:tag(tty),widthReader:width??0,prettyReader:!!tty}));`;
 return JSON.parse(child.execFileSync(process.execPath,['-e',code,path.resolve(file)],{encoding:'utf8'}));
}
function TestBuiltTTYConsumerEquivalence(before,after){
 const a=observe(before),b=observe(after);
 assert.equal(a.node,'v24.19.0');assert.equal(b.node,a.node);
 assert.equal(a.width,'undefined');assert.equal(a.tty,'undefined');assert.equal(b.width,0);assert.equal(b.tty,false);
 assert.equal(a.widthReader,b.widthReader);assert.equal(a.prettyReader,b.prettyReader);
 return{test:'TestBuiltTTYConsumerEquivalence',stdio:'pipe',before:a,after:b};
}
module.exports={TestBuiltTTYConsumerEquivalence};
if(require.main===module){const result=TestBuiltTTYConsumerEquivalence(process.argv[2],process.argv[3]);fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));}
