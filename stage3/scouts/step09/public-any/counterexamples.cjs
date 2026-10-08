'use strict';
// Additional API calls, explicitly separate from the 301-project CLI observations.
const fs=require('node:fs'),ts=require(process.argv[2]);
if(ts.version!=='6.0.3')throw Error('requires 6.0.3');
const result={scope:'focused actual Node API calls, not acceptance coverage',readers:[]};
for(const text of ['7','0','true','false','null','"text"','""','[1]','{}','bad']){
 const host={readFile:()=>text};const source=ts.parseJsonText('config.json',text),errors=[];
 result.readers.push({text,readJson:{type:typeof ts.readJson('config.json',host),value:ts.readJson('config.json',host)},readJsonOrUndefined:{type:typeof ts.readJsonOrUndefined('config.json',host)},convertToObject:{type:typeof ts.convertToObject(source,errors),value:ts.convertToObject(source,[])},errors:errors.map(e=>e.code)});
}
const handle=ts.sys.setTimeout(()=>{},60000);handle.unref();ts.sys.clearTimeout(handle);
result.nodeTimer={type:typeof handle,constructor:handle.constructor.name,cancellation:'same handle passed to ts.sys.clearTimeout'};
const allocator=ts.objectAllocator;const Node=allocator.getNodeConstructor();const header=new Node(ts.SyntaxKind.Identifier,-1,-1);
result.allocator={provider:Node.name,parentType:typeof header.parent,ownFields:Object.keys(header).sort(),hasEscapedText:Object.hasOwn(header,'escapedText')};
fs.writeFileSync(process.argv[3],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
