#!/usr/bin/env node
"use strict";
// Compose the existing parsed-owner proof for 20 with the exact sanctioned 40 ledger.
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const {createRequire} = require("node:module");
const checker = path.join(__dirname, "../20-optional-declarations/check-baselines.cjs");
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
let expected = original;`);
replaceOnce("expected.slice(edit.at);", "expected.slice(edit.end === undefined ? edit.at : edit.end);");
replaceOnce("bytes.equals(Buffer.from(original)) || bytes.equals(Buffer.from(expected))", "bytes.equals(Buffer.from(original)) || bytes.equals(Buffer.from(api40Only)) || bytes.equals(Buffer.from(expected))");
vm.runInThisContext("(function(require){" + code + "\n})", {filename: checker})(createRequire(checker));
