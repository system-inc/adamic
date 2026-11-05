'use strict';
const fs=require('node:fs'),assert=require('node:assert/strict'),cp=require('node:child_process'),crypto=require('node:crypto'),ts=require('typescript');
const root=process.argv[2],repo=process.cwd(),base='b729c7e07010fe6db34c6afc1d11c8ca63a86ae3',main='82b0c076395636800764ed1af7ada569681fc2ae';
const directories=['99-correctness-axolotl-l','99-correctness-axolotl-m','99-correctness-axolotl-q'];
const paths=cp.execFileSync('git',['diff','--name-only',base,main,'--','stage3/adapt','stage3/source.json','stage3/apply.py','stage3/api','stage3/oracle'],{encoding:'utf8'}).trim().split('\n');
assert(paths.every(file=>directories.some(dir=>file.startsWith('stage3/adapt/'+dir+'/'))));
const before=fs.readFileSync(root+'/src/compiler/core.ts','utf8').replace(/\r\n/g,'\n');let preimage=before;const reversed=[];
for(const directory of [...directories].reverse())for(const pin of require(repo+'/stage3/adapt/'+directory+'/adapt.cjs').owners){assert.equal(pin.file,'core.ts');assert.equal(preimage.split(pin.after).length,2,pin.owner);preimage=preimage.replace(pin.after,pin.before);reversed.push(pin.owner);}
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
const rows=[false,true].map(removeComments=>{const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;const a=emit(before),b=emit(preimage);assert.equal(a,b);return {removeComments,byteIdentical:true,sha256:hash(a)};});
console.log(JSON.stringify({base,main,onlyAdaptationInputDifferences:directories,commonApplyParserSourceAndOracleInputsIdentical:true,reversedOwners:reversed,mainCorePreimageSha256:hash(preimage),coreJavascript:rows,scope:'The reference is a measured base oracle; its only source differences from main are these erased core owners. This is runtime-artifact equivalence, not a separate fresh main oracle invocation.'},null,2));
