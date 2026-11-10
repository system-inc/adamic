"use strict";
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict"), crypto = require("node:crypto");
const ts = require("typescript"), pins = require("./debug-display-2-pins.json");
const home = "src/deprecatedCompat/controlFlowGraph.ts";
const hash = text => crypto.createHash("sha256").update(text).digest("hex");
function functions(text) {
    const sf = ts.createSourceFile("owners.ts", text, ts.ScriptTarget.Latest, true);
    assert.equal(sf.parseDiagnostics.length, 0, "owner file does not parse");
    const found = new Map();
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.body && node.name && pins.owners.some(p => p.name === node.name.text)) {
            assert(!found.has(node.name.text), "duplicate owner " + node.name.text);
            found.set(node.name.text, {start:node.getStart(sf), end:node.end, text:node.getText(sf), body:node.body.getText(sf)});
        }
        ts.forEachChild(node, visit);
    }
    visit(sf); return found;
}
function inspect(debug, local) {
    const found = functions(debug), moved = functions(local || "");
    assert.equal(local === undefined || local === pins.module, true, "moved module drift");
    let states = [];
    for (const pin of pins.owners) {
        assert.equal(hash(pin.before), pin.beforeSha256, "before pin corrupt");
        assert.equal(hash(pin.after), pin.afterSha256, "after pin corrupt");
        const original = found.get(pin.name), relocated = moved.get(pin.name);
        if (pin.move) {
            assert.equal(Number(!!original) + Number(!!relocated), 1, "missing or duplicate home: " + pin.name);
            const owner = original || relocated;
            assert.equal(owner.text, pin.before, "moved owner changed: " + pin.name);
            states.push(original ? "before" : "after");
        } else {
            assert(original && !relocated, "mapper owner moved");
            assert(original.text === pin.before || original.text === pin.after, "mapper owner drift");
            states.push(original.text === pin.before ? "before" : "after");
        }
    }
    assert(states.every(s => s === states[0]), "half-adapted debugger display tree");
    assert.equal(local !== undefined, states[0] === "after", "module home mismatch");
    return {state:states[0], found};
}
function plan(debug, local) {
    const {state,found} = inspect(debug,local);
    if (state === "after") return {debug,local,edits:0};
    let after = debug;
    for (const pin of [...pins.owners].sort((a,b)=>found.get(b.name).start-found.get(a.name).start)) {
        const owner = found.get(pin.name);
        after = after.slice(0,owner.start) + pin.after + after.slice(owner.end);
    }
    inspect(after,pins.module);
    const left = functions(debug), right = functions(pins.module);
    for (const pin of pins.owners.filter(p=>p.move)) assert.equal(left.get(pin.name).body,right.get(pin.name).body,"body changed while moving");
    return {debug:after,local:pins.module,edits:3};
}
// Replay earlier nested-owner adapters against the verified original home.
// This is a virtual view only. No restored text is written to the source tree.
function restore(debug, local) {
    assert.equal(inspect(debug,local).state,"after");
    let text=debug;
    for (const pin of pins.owners) {
        if (!pin.move) {
            assert.equal(text.split(pin.after).length-1,1,"mapper after text moved");
            text=text.replace(pin.after,pin.before);
        }
    }
    // The graph owners were the final declarations in the Debug namespace.
    assert(text.endsWith(pins.suffix),"Debug namespace ending moved");
    text=text.slice(0,-pins.suffix.length)+pins.restoredSuffix;
    assert.equal(inspect(text,undefined).state,"before");
    return text;
}
module.exports={home,functions,inspect,plan,restore,pins};
