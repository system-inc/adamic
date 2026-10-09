'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require(process.env.CENSUS_TYPESCRIPT||'typescript');
const tree=path.resolve(process.argv[2]);
const debug=fs.readFileSync(path.join(tree,'src/compiler/debug.ts'),'utf8'),core=fs.readFileSync(path.join(tree,'src/compiler/core.ts'),'utf8');
const {plan}=require('./adapt.cjs');
assert.equal(plan(debug,core).changed,debug,'not idempotent');
let killed=false;
const writerMutant=debug.replace('shouldAssertFunction(AssertionLevel.Normal, "assertNode")','shouldAssertFunction(AssertionLevel.Normal, "assert")');
assert.notEqual(writerMutant,debug);
try {plan(writerMutant,core);}catch(e){killed=/finite assertion key set changed/.test(e.message);}
assert(killed,'new assertion key escaped provenance guard');
const a=core.indexOf('export function getOwnKeys<T>'),b=core.indexOf('\n}',a)+2;
const enumeratorMutant=core.slice(0,a)+core.slice(a,b).replace('keys.push(key);','keys.push("wrong");')+core.slice(b);
killed=false;
try {plan(debug,enumeratorMutant);}catch(e){killed=/key enumeration changed/.test(e.message);}
assert(killed,'wrong enumerator escaped');
const candidate=fs.readFileSync(path.join(__dirname,'same-map-proposal.a'),'utf8');
function diagnostics(text){
 const filename=path.join(__dirname,'virtual-proof.ts'),options={strict:true,noUncheckedIndexedAccess:true,noEmit:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext};
 const host=ts.createCompilerHost(options),original=host.getSourceFile.bind(host);
 host.getSourceFile=(file,language,...rest)=>file===filename?ts.createSourceFile(filename,text,language,true):original(file,language,...rest);
 const program=ts.createProgram([filename],options,host);
 return ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,start:d.start,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
}
assert.deepEqual(diagnostics(candidate),[],'honest isolated proposal must type-check');
const assignment='\nconst strings: readonly string[] | undefined = output;\n';
assert(diagnostics(candidate+assignment).some(d=>d.code===2322),'union did not reject the old string promise');
const start=candidate.indexOf('export function sameMap<'),end=candidate.indexOf('\n}',candidate.indexOf('export function sameMap<T, U = T>(array: readonly T[] | undefined',start))+2;
const oldStart=core.indexOf('export function sameMap<'),oldEnd=core.indexOf('\n}',core.indexOf('export function sameMap<T, U = T>(array: readonly T[] | undefined',oldStart))+2;
const old=core.slice(oldStart,oldEnd),mutant=candidate.slice(0,start)+old+candidate.slice(end);
assert.deepEqual(diagnostics(mutant+assignment),[],'old lying signature unexpectedly rejected the witness');
for(const removeComments of [false,true]){
 const compilerOptions={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};
 assert.equal(ts.transpileModule(candidate.slice(start,end),{compilerOptions}).outputText,ts.transpileModule(old,{compilerOptions}).outputText,'proposal changed runtime JS');
}
console.log(JSON.stringify({idempotent:true,key_writer_mutant_caught:true,key_enumerator_mutant_caught:true,sameMap_proposal_typechecks:true,sameMap_old_signature_mutant_caught_by_negative_assignment:true,proposal_runtime_javascript_identical:true}));
