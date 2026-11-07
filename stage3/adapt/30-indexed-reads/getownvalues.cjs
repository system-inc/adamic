"use strict";
const pins = require("./getownvalues-owners.json");
function validate(ts, text) {
    const source = ts.createSourceFile("core.ts", text, ts.ScriptTarget.Latest, true);
    const owners = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === "getOwnValues");
    if (source.parseDiagnostics.length || owners.length !== 1) throw new Error("getOwnValues owner drift");
    const owner = owners[0], actual = owner.getText(source);
    if (![pins.raw, pins.before, pins.after].includes(actual)) throw new Error("getOwnValues whole owner drift or half-adapted owner");
    return { owner, actual, source };
}
function plan(ts, text, check = false) {
    const { owner, actual, source } = validate(ts, text);
    if (actual === pins.after) return { text, edits: 0 };
    if (actual !== pins.before) throw new Error("getOwnValues required ! must precede record-only amendment");
    if (check) throw new Error("getOwnValues record-only amendment missing");
    return { text: text.slice(0, owner.getStart(source)) + pins.after + text.slice(owner.end), edits: 1 };
}
module.exports = { validate, plan };
