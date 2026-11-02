'use strict';
const assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path');
const ts=require('typescript');
const {compiler}=require('../../47-host-errors/adapt.cjs');
const {validateBindings,plannedProgram}=require('../adapt.cjs');
function inspect(tree,done){
 const program=compiler(tree,ts),file=program.getSourceFile(path.join(tree,'src/compiler/sys.ts'));
 const result=validateBindings(program,file,done);
 return{qualifiedUses:result.qualifiedUses,qualifiedUseLines:result.qualifiedUseLines,sites:result.evidence};
}
function mutants(tree){
 const program=compiler(tree,ts),file=program.getSourceFile(path.join(tree,'src/compiler/sys.ts'));
 const cases=[
  ['module collision',file.text+'\nconst _path = { join: () => "wrong" };\n'],
  ['enclosing shadow',file.text.replace('function readFile(fileName: string, _encoding?: string)', 'function readFile(fileName: string, _encoding?: string, _fs?: typeof import("fs"))')],
  ['bare namespace value',file.text+'\nvoid _os;\n'],
 ];
 return cases.map(([name,text])=>{
  assert.notEqual(text,file.text);
  const virtual=plannedProgram(program,file,text);let message='';
  try{validateBindings(virtual,virtual.getSourceFile(file.fileName),true);}catch(error){message=error.message;}
  assert.match(message,/collision or bare namespace use:|captured or shadowed use:/,name+' survived');
  console.log(`CAUGHT ${name}: ${message}`);return{name,message,caught:true};
 });
}
const result={before:inspect(path.resolve(process.argv[2]),false),after:inspect(path.resolve(process.argv[3]),true),mutants:mutants(path.resolve(process.argv[3]))};
fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
