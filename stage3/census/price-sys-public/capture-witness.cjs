const assert=require('node:assert/strict');
function run(bound){const events=[],sys={tag:'sys',writeFile(path,data,bom){events.push([this===sys,this.tag,path,data,bom]);}};const originalIdentity=sys.writeFile;const originalWriteFile=bound?sys.writeFile.bind(sys):sys.writeFile;sys.writeFile=(path,data,bom)=>originalWriteFile.call(sys,path,data,bom);sys.writeFile('first','a',false);const patched=sys.writeFile;sys.writeFile=()=>events.push(['later']);patched('second','b',true);return {events,capturedIdentityUnchanged:originalWriteFile===originalIdentity};}
const control=run(false),bound=run(true);assert.deepEqual(control.events,bound.events);
let recursion=false;const sys={writeFile(){}};const originalWriteFile=(...args)=>sys.writeFile(...args);sys.writeFile=(...args)=>originalWriteFile(...args);try{sys.writeFile('file','text',false);}catch(e){recursion=e instanceof RangeError;}assert(recursion);
console.log(JSON.stringify({control,bound,naiveArrowRecurses:recursion},null,2));
