"use strict";
const fs=require("node:fs"),path=require("node:path"),vm=require("node:vm"),assert=require("node:assert/strict"),crypto=require("node:crypto"),ts=require("typescript");
if(process.argv.length!==4)throw Error("usage: node verify_custom_set.cjs <tsc-tree> <evidence-json>");
assert.equal(ts.version,"6.0.3");const source=fs.readFileSync(path.join(__dirname,"custom_set.a"),"utf8");
const core=fs.readFileSync(path.join(process.argv[2],"src/compiler/core.ts"),"utf8");
function createSetDeclaration(text){const file=ts.createSourceFile('input.a',text,ts.ScriptTarget.Latest,true);const matches=file.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createSet');assert.equal(matches.length,1);return matches[0].getText(file).replace(/\r\n/g,'\n');}
assert.equal(createSetDeclaration(source),createSetDeclaration(core),'createSet body changed');
function run(code){const output=[];vm.runInNewContext(ts.transpileModule(code,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText,{exports:{},console:{log:value=>output.push(String(value))}});return output.join("\n")+"\n";}
const original=run(source);assert.ok(original.startsWith("2:true\nfirst\ncollision\n"));
const mutantSource=source.replace("createSet<Key>(()=>0,(left,right)=>left.span===right.span)","new Set<Key>()");assert.notEqual(mutantSource,source);
const mutant=run(mutantSource);assert.notEqual(mutant,original,"builtin Set substitution survived");
fs.writeFileSync(process.argv[3],JSON.stringify({bodyIdentical:true,coreSha256:crypto.createHash("sha256").update(core).digest("hex"),sourceSha256:crypto.createHash("sha256").update(source).digest("hex"),original,mutant,mutantCaught:true,nativeStatus:"generator ownership refusal; no native execution"},null,2)+"\n");console.log("PASS own custom-equality protocol; builtin Set substitution differs");
