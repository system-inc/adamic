#!/usr/bin/env node
"use strict";
// Compose stock parsed-owner proofs; reconstruct solely from pristine API text.
const fs=require("node:fs"),path=require("node:path"),vm=require("node:vm"),{createRequire}=require("node:module");
const checker=path.join(__dirname,"../70-readonly-views/check-api.cjs");
let code=fs.readFileSync(checker,"utf8");
if(!code.startsWith("#!/usr/bin/env node\n"))throw new Error("API proof header drift");
code=code.slice(code.indexOf("\n")+1);
function replaceOnce(before,after){if(code.split(before).length!==2)throw new Error("API proof composition drift: "+before);code=code.replace(before,after);}
replaceOnce("let optionalExpected = original;",`
const api40Sites=JSON.parse(fs.readFileSync(${JSON.stringify(path.join(__dirname,'api40-sites.json'))},'utf8'));
assert.equal(api40Sites.length,28);
assert.equal(api40Sites.filter(r=>r.kind==='brand').length,27);
const seen40=new Set();
for(const site of api40Sites){
 const key=JSON.stringify(site.owner);assert(!seen40.has(key));seen40.add(key);
 assert.equal(site.before,'any');
 if(site.kind==='brand')assert.equal(site.after,'undefined');
 else {assert.deepEqual(site.owner,['ErrorCallback','arg0']);assert.equal(site.after,'string | number');}
 const n=findOwned(n=>n.type&&n.name,site.owner);
 assert(ts.isPropertySignature(n)||ts.isParameter(n));assert.equal(n.type.getText(before),site.before);
 edits.push({at:n.type.getStart(before),end:n.type.end,text:site.after,adaptation:40});
}
let api40Only=original;
for(const e of edits.filter(e=>e.adaptation===40).sort((a,b)=>b.at-a.at))api40Only=api40Only.slice(0,e.at)+e.text+api40Only.slice(e.end);
const public32=[];
for(const site of JSON.parse(fs.readFileSync(${JSON.stringify(path.join(__dirname,'public-host-sites.json'))},'utf8')).filter(r=>r.public && r.kind !== 'restored-public-method')){
 assert.equal(site.kind,'public-method');
 const source=parse(site.file,fs.readFileSync(path.join(pristine,'src/compiler',site.file),'utf8'));
 const iface=source.statements.filter(n=>ts.isInterfaceDeclaration(n)&&n.name.text===site.interface);assert.equal(iface.length,1);
 const selected=iface[0].members.filter(n=>n.name?.getText(source)===site.name);assert.equal(selected.length,1);
 assert.equal(tokens(selected[0].getText(source)),tokens(site.before),'source owner signature drift');
 const n=findOwned(n=>ts.isMethodSignature(n)||ts.isPropertySignature(n),[site.interface,site.name]);
 assert(n.questionToken);assert.equal(tokens(n.getText(before)),tokens(site.before),'public signature drift');
 // Exact original callable signature, unchanged variance, plus explicit undefined.
 if(ts.isMethodSignature(selected[0])){
  const method=site.before.slice(0,site.name.length)+site.before.slice(site.name.length+1);
  assert.equal(site.after,site.name+'?: { '+method+' }["'+site.name+'"] | undefined;');
 }else assert.equal(site.after,site.name+'?: ('+selected[0].type.getText(source)+') | undefined;');
 const emitted=ts.transpileDeclaration('export interface Proof { '+site.after+' }',{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext}}).outputText;
 const model=parse('owner-model.d.ts',emitted);assert.equal(model.statements.length,1);assert.equal(model.statements[0].members.length,1);
 const member=model.statements[0].members[0];assert(ts.isPropertySignature(member));assert.equal(tokens(member.getText(model)),tokens(site.after));
 const at=n.getStart(before),indent=original.slice(original.lastIndexOf('\\n',at)+1,at);
 assert.equal(indent.trim(),'');
 const modelIndent=emitted.slice(emitted.lastIndexOf('\\n',member.getStart(model))+1,member.getStart(model));assert.equal(modelIndent,'    ');
 const replacement=member.getText(model).split('\\n').map((line,i)=>i?indent+line.slice(modelIndent.length):line).join('\\n');
 edits.push({at,end:n.end,text:replacement,adaptation:32});
 public32.push({owner:[site.interface,site.name],before:n.getText(before),after:site.after,line:before.getLineAndCharacterOfPosition(n.getStart(before)).line+1});
}
assert.equal(public32.length,1,'public property owner census drift');
for(const site of JSON.parse(fs.readFileSync(${JSON.stringify(path.join(__dirname,'public-host-sites.json'))},'utf8')).filter(r=>r.kind==='restored-public-method')){
 const n=findOwned(n=>ts.isMethodSignature(n),[site.interface,site.name]);assert.equal(tokens(n.getText(before)),tokens(site.before));
 const src=parse(site.file,fs.readFileSync(path.join(adapted,'src/compiler',site.file),'utf8'));
 const iface=src.statements.filter(n=>ts.isInterfaceDeclaration(n)&&n.name.text===site.interface);assert.equal(iface.length,1);
 const member=iface[0].members.filter(n=>n.name?.getText(src)===site.name);assert.equal(member.length,1);assert(ts.isMethodSignature(member[0]));assert.equal(member[0].getText(src),site.before);
}
const adaptedRequire=require('node:module').createRequire(path.join(adapted,'package.json'));
const formatter=adaptedRequire('@dprint/formatter').createFromBuffer(fs.readFileSync(adaptedRequire('@dprint/typescript').getPath()));
// Pin the formatter to upstream's pristine script; do not infer config from output.
const bundler=parse('dtsBundler.mjs',fs.readFileSync(path.join(pristine,'scripts/dtsBundler.mjs'),'utf8'));
const configurations=[];
function configs(n){if(ts.isCallExpression(n)&&n.expression.getText(bundler)==='formatter.setConfig')configurations.push(n);ts.forEachChild(n,configs);}configs(bundler);
assert.equal(configurations.length,1);assert.deepEqual(configurations[0].arguments.map(a=>tokens(a.getText(bundler))),[tokens('{ indentWidth:4, lineWidth:1000, newLineKind:"auto", useTabs:false, }'),tokens('{ quoteStyle:"preferDouble", }')]);
formatter.setConfig({indentWidth:4,lineWidth:1000,newLineKind:'auto',useTabs:false},{quoteStyle:'preferDouble'});
function format(text){return formatter.formatText({filePath:'dummy.d.ts',fileText:text}).replaceAll('\\r\\n','\\n');}
assert.equal(format(original),original,'formatter does not reproduce pristine API');
let optionalExpected = original;`);
replaceOnce("let expected = original;", `const handoffSites = JSON.parse(fs.readFileSync(${JSON.stringify(path.join(__dirname, "handoff-sites.json"))}, "utf8"));
assert.equal(handoffSites.length, 6);
assert.equal(handoffSites.filter(s => s.public).length, 2);
const sourceText32 = fs.readFileSync(path.join(adapted, "src/compiler/types.ts"), "utf8");
const source32 = parse("types.ts", sourceText32);
const handoffProof = [];
for (const site of handoffSites) {
    const sourceNode = require(${JSON.stringify(path.join(__dirname, "handoffs.cjs"))}).owners(ts, source32, site);
    assert.equal(sourceNode.type.getText(source32), site.action === "decline" ? site.before : site.after, "handoff source type drift");
    const candidates = properties(before).filter(n => ownerPath(n) === JSON.stringify([site.interface, site.name]));
    assert.equal(candidates.length, site.public ? 1 : 0, "handoff public owner drift");
    if (!site.public) continue;
    assert(!approved.has(JSON.stringify([site.interface, site.name])), "handoff duplicates adaptation 20");
    assert.equal(candidates[0].type.getText(before), site.before);
    edits.push({at:candidates[0].type.end,text:" | undefined",adaptation:"32-handoff",handoffOwner:site.interface});
    handoffProof.push({path:[site.interface,site.name],line:before.getLineAndCharacterOfPosition(candidates[0].getStart(before)).line+1});
}
const jsonOwner=findOwned(n=>ts.isInterfaceDeclaration(n),['JsonSourceFile']);
assert(!jsonOwner.members.some(n=>n.name?.getText(before)==='extendedSourceFiles'));
const jsonLast=jsonOwner.members[jsonOwner.members.length-1];
edits.push({at:jsonLast.end,text:"\\n        extendedSourceFiles?: string[];",adaptation:75});
let previous32Expected=original;
for(const e of edits.filter(e=>e.adaptation!==32).sort((a,b)=>b.at-a.at))previous32Expected=previous32Expected.slice(0,e.at)+e.text+previous32Expected.slice(e.end===undefined?e.at:e.end);
previous32Expected=format(previous32Expected);
let expected = original;`);
replaceOnce("expected.slice(edit.at);", "expected.slice(edit.end === undefined ? edit.at : edit.end);");
replaceOnce("if (actual !== expected) {", "expected=format(expected);\nif (actual !== expected) {");
replaceOnce("assert.equal(newLines.length, oldLines.length, 'API additions must retain existing lines');", "assert(newLines.length >= oldLines.length, 'host unions cannot delete API declarations');");
replaceOnce("assert(changedLines.every(line => allowedLines.has(line)), 'a changed line lies outside its parsed owner type edit');", "// Exact reconstruction above is stricter than line attribution when owner edits expand lines.\nassert.equal(actual,expected);");
replaceOnce("bytes.equals(Buffer.from(original)) || bytes.equals(Buffer.from(optionalExpected)) || bytes.equals(Buffer.from(expected))", "bytes.equals(Buffer.from(original)) || bytes.equals(Buffer.from(api40Only)) || bytes.equals(Buffer.from(optionalExpected)) || bytes.equals(Buffer.from(previous32Expected)) || bytes.equals(Buffer.from(expected))");
replaceOnce("readonly_owners: readonlyOwners,", "readonly_owners: readonlyOwners, adaptation40_owners: api40Sites, handoff_owners: handoffProof, public_host_owners: public32, api_line_delta: newLines.length-oldLines.length,");
vm.runInThisContext("(function(require){"+code+"\n})",{filename:checker})(createRequire(checker));
