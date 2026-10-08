#!/usr/bin/env node
"use strict";
// Extend the supplied lane sanction only from the three retained owner ledgers.
// This never modifies the lane's checked-in default manifest.
const fs=require("node:fs"),path=require("node:path"),assert=require("node:assert/strict"),ts=require("typescript");
const [lane,pristine,out]=process.argv.slice(2);assert(lane&&pristine&&out);assert.equal(ts.version,"6.0.3");
const normalized=require(path.resolve(lane,"normalize-api.cjs"));
const base=JSON.parse(fs.readFileSync(path.join(lane,"sanctioned-api.json"),"utf8"));
assert.equal(base.source_commit,"050880ce59e30b356b686bd3144efe24f875ebc8");
const original=fs.readFileSync(path.join(pristine,"tests/baselines/reference/api/typescript.d.ts"),"utf8");
const sf=ts.createSourceFile("api.d.ts",original,ts.ScriptTarget.Latest,true);assert.equal(sf.parseDiagnostics.length,0);
const owners=[...require("./handoff-sites.json").filter(s=>s.public&&s.action!=="decline"),...require("./public-host-sites.json").filter(s=>s.public&&s.file==="builderPublic.ts")];assert.equal(owners.length,3);
assert.deepEqual(owners.map(s=>s.interface+"."+s.name).sort(),["AmdDependency.name","BuilderProgramHost.createHash","CommentRange.hasTrailingNewLine"]);
const edits=[];
for(const owner of owners){const found=[];function visit(n){if(ts.isPropertySignature(n)&&n.name.getText(sf)===owner.name&&ts.isInterfaceDeclaration(n.parent)&&n.parent.name.text===owner.interface)found.push(n);ts.forEachChild(n,visit);}visit(sf);assert.equal(found.length,1);const n=found[0];assert(n.questionToken&&n.type);const type=n.type.getText(sf);
 if(owner.interface==="BuilderProgramHost")assert.equal(n.getText(sf),owner.before);else assert.equal(type,owner.before);
 edits.push({at:n.type.getStart(sf),end:n.type.end,text:(ts.isFunctionTypeNode(n.type)?"("+type+")":type)+" | undefined"});
}
let expected=original;for(const e of edits.sort((a,b)=>b.at-a.at))expected=expected.slice(0,e.at)+e.text+expected.slice(e.end);
const additions=normalized.changes(original,expected);assert.equal(additions.length,3);
assert(additions.every(row=>!base.normalized_changes.some(old=>old.declaration===row.declaration)));
const manifest={source_commit:base.source_commit,normalized_changes:[...base.normalized_changes,...additions].sort((a,b)=>a.declaration.localeCompare(b.declaration)),additional_sanctions:{adaptation:32,owners:owners.map(s=>[s.interface,s.name]),proof:"Original property form retained; truthful undefined unions sanctioned per owner; exact full API reconstruction in public-api.cjs"}};
assert.equal(manifest.normalized_changes.length,base.normalized_changes.length+3);
fs.writeFileSync(out,JSON.stringify(manifest,null,2)+"\n");console.log(JSON.stringify({status:"pass",original:base.normalized_changes.length,added:additions.map(r=>r.declaration),total:manifest.normalized_changes.length},null,2));
