#!/usr/bin/env node
"use strict";
// Reconstruct exactly the sanctioned 20, 40, 70 and two public handoff unions.
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const {createRequire} = require("node:module");
const checker = path.join(__dirname, "../70-readonly-views/check-api.cjs");
let code = fs.readFileSync(checker, "utf8");
if (!code.startsWith("#!/usr/bin/env node\n")) throw new Error("API proof header drift");
code = code.slice(code.indexOf("\n") + 1);
function replaceOnce(before, after) {
    if (code.split(before).length !== 2) throw new Error("API proof composition drift");
    code = code.replace(before, after);
}
replaceOnce("let expected = original;", `
const api40Sites = JSON.parse(fs.readFileSync(${JSON.stringify(path.join(__dirname, "api40-sites.json"))}, "utf8"));
assert.equal(api40Sites.length, 28);
assert.equal(api40Sites.filter(s => s.kind === "brand").length, 27);
const api40Owners = new Set();
for (const site of api40Sites) {
    const ownerKey = JSON.stringify(site.owner);
    assert(!api40Owners.has(ownerKey), "duplicate API 40 owner"); api40Owners.add(ownerKey);
    assert.equal(site.before, "any");
    if (site.kind === "brand") { assert(site.owner.at(-1).includes("Brand")); assert.equal(site.after, "undefined"); }
    else { assert.equal(site.kind, "diagnostic argument"); assert.deepEqual(site.owner, ["ErrorCallback", "arg0"]); assert.equal(site.after, "string | number"); }
    const matches = [];
    function visit(n) {
        if (n.type && n.name && ownerPath(n) === JSON.stringify(site.owner)) matches.push(n);
        ts.forEachChild(n, visit);
    }
    visit(before);
    assert.equal(matches.length, 1, "API 40 owner drift: " + JSON.stringify(site.owner));
    const n = matches[0];
    assert(ts.isPropertySignature(n) || ts.isParameter(n));
    assert.equal(n.type.getText(before), site.before);
    edits.push({at: n.type.getStart(before), end: n.type.end, text: site.after, adaptation: 40});
    owners.push({path: site.owner, line: before.getLineAndCharacterOfPosition(n.getStart(before)).line + 1, adaptation: 40});
}
let api40Only = original;
for (const e of edits.filter(e => e.adaptation === 40).sort((a,b) => b.at-a.at)) api40Only = api40Only.slice(0,e.at)+e.text+api40Only.slice(e.end);
const handoffSites = JSON.parse(fs.readFileSync(${JSON.stringify(path.join(__dirname, "handoff-sites.json"))}, "utf8"));
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
    edits.push({at:candidates[0].type.end,text:" | undefined",adaptation:32,handoffOwner:site.interface});
    handoffProof.push({path:[site.interface,site.name],line:before.getLineAndCharacterOfPosition(candidates[0].getStart(before)).line+1});
}
let priorHandoffExpected = original;
for (const e of edits.filter(e => e.handoffOwner !== "CommentRange").sort((a,b) => b.at-a.at)) priorHandoffExpected = priorHandoffExpected.slice(0,e.at)+e.text+priorHandoffExpected.slice(e.end === undefined ? e.at : e.end);
let expected = original;`);
replaceOnce("edit.text + expected.slice(edit.at);", "edit.text + expected.slice(edit.end === undefined ? edit.at : edit.end);");
replaceOnce("bytes.equals(Buffer.from(optionalExpected)) || bytes.equals(Buffer.from(expected))", "bytes.equals(Buffer.from(optionalExpected)) || bytes.equals(Buffer.from(api40Only)) || bytes.equals(Buffer.from(priorHandoffExpected)) || bytes.equals(Buffer.from(expected))");
replaceOnce("snapshot_declarations: owners.length,", "snapshot_declarations: 189, sanctioned_40: 28, handoff_owners: handoffProof, reconstructed_owners: owners.length,");
vm.runInThisContext("(function(require){" + code + "\n})", {filename: checker})(createRequire(checker));
