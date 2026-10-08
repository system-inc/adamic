#!/usr/bin/env node
"use strict";
const fs=require("node:fs"),path=require("node:path"),assert=require("node:assert/strict");
const ts=require("typescript"),{plan}=require("./adapt.cjs");
const tree=process.argv[2], proofs=[];
for(const site of require("./public-host-sites.json").filter(s=>s.kind==="restored-public-method")){
 const text=fs.readFileSync(path.join(tree,"src/compiler",site.file),"utf8");
 const source=ts.createSourceFile(site.file,text,ts.ScriptTarget.Latest,true), owners=[];
 function find(n){if(ts.isInterfaceDeclaration(n)&&n.name.text===site.interface)owners.push(n);ts.forEachChild(n,find);}find(source);assert.equal(owners.length,1);
 const members=owners[0].members.filter(n=>n.name?.getText(source)===site.name);assert.equal(members.length,1);const member=members[0];assert.equal(member.getText(source),site.before);
 const mutant=text.slice(0,member.getStart(source))+site.after+text.slice(member.end);let caught;
 try{plan(ts,site.file,mutant,true);}catch(e){caught=e.message;}
 assert.equal(caught,"public method restoration missing");
 assert.equal(plan(ts,site.file,mutant).text,text,"rollback not exact");
 proofs.push({owner:site.interface+"."+site.name,mutant:"replace original method with prior property",caught});
}
const sites=require("./sites.json"), site=sites.find(s=>s.file==="binder.ts"&&s.action==="assert");
const text=fs.readFileSync(path.join(tree,"src/compiler/binder.ts"),"utf8");
const sf=ts.createSourceFile("binder.ts",text,ts.ScriptTarget.Latest,true), found=[];
function visit(n){if(ts.isElementAccessExpression(n)&&n.getText(sf).replaceAll("!","")===site.expression)found.push(n);ts.forEachChild(n,visit);}visit(sf);
assert.equal(found.length,site.total);const n=found[site.occurrence-1];assert(ts.isNonNullExpression(n.parent));
const mutant=text.slice(0,n.getStart(sf))+"("+n.getText(sf)+" ?? 0)"+text.slice(n.parent.end);
const emit=t=>ts.transpileModule(t,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
assert.notEqual(emit(mutant),emit(text));let caught;
try{plan(ts,"binder.ts",mutant,true);}catch(e){caught=e.message;}
assert(caught?.startsWith("required read defaulted:"));
proofs.push({file:"binder.ts",expression:site.expression,mutant:"required ! replaced with ?? 0",caught,stock_emitter:"changed JavaScript detected"});
assert.equal(proofs.length,11);
console.log(JSON.stringify({status:"pass",proofs},null,2));
