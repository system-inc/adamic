// Expose the actual pinned micromark preprocessor by reference; do not change its body.
import fs from 'node:fs';
import vm from 'node:vm';
let bundle=fs.readFileSync(process.argv[2]+'/plugins/markdown.js','utf8');
const anchor='var MD=/';
if(bundle.split(anchor).length!==2 || !bundle.includes('function li(){let e=1,t="",r=!0,n;'))throw new Error('pinned preprocess anchor');
bundle=bundle.replace(anchor,'globalThis.adamicInputPreprocess=li;'+anchor);
const context={module:{exports:{}},exports:{}};
vm.runInNewContext(bundle,context);
const output=[];
for(const line of fs.readFileSync(process.argv[3],'utf8').split('\n')) {
 if(line==='')continue;
 const units=line==='-'?[]:line.split(',').map(Number);
 // Avoid the engine's argument-count limit on large repository documents.
 let source='';for(let i=0;i<units.length;i+=8192)source+=String.fromCharCode(...units.slice(i,i+8192));
 const chunks=context.adamicInputPreprocess()(source,undefined,true);
 output.push(chunks.map(chunk=>typeof chunk==='string'?'T'+Array.from({length:chunk.length},(_,i)=>chunk.charCodeAt(i)).join(','):'C'+(chunk===null?0:chunk)).join(';'));
}
process.stdout.write(output.join('\n')+'\n');
