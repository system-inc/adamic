"use strict";
const assert = require("node:assert/strict");
const {createHash} = require("node:crypto");
const before = "host: { directoryExists?: (directoryName: string) => boolean; }";
const after = "host: { directoryExists?: ((directoryName: string) => boolean) | undefined; }";
const pins = {
    "file": "utilities.ts",
    "owner": "directoryProbablyExists",
    "ownerPath": [
        "directoryProbablyExists"
    ],
    "ownerBeforeSha256": "d9626c9175b691a642eb06c259af55004963575e3f34fed89bc995727fdef94b",
    "ownerAfterSha256": "3b44c3d157885c3ff8a2579f1e89801fedc996d08636d62e353455c60924c92d"
};
function hash(text) { return createHash("sha256").update(text).digest("hex"); }
function owner(ts, file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true), found = [];
    assert.equal(source.parseDiagnostics.length, 0, "directory helper parse error");
    function visit(n) {
        if (ts.isFunctionDeclaration(n) && n.name?.text === "directoryProbablyExists" && n.body) found.push(n);
        ts.forEachChild(n, visit);
    }
    visit(source);
    assert.equal(found.length, 1, "directory helper owner count drift");
    const node = found[0];
    assert.equal(node.parent, source, "directory helper owner moved");
    assert(ts.getJSDocTags(node).some(t => t.tagName.text === "internal"), "directory helper must remain internal");
    const parameters = node.parameters.filter(p => p.name.getText(source) === "host");
    assert.equal(parameters.length, 1, "directory helper host count drift");
    return {source, node, parameter: parameters[0], hash: hash(node.getText(source))};
}
function plan(ts, file, raw, check = false) {
    assert.equal(ts.version, "6.0.3");
    const text = raw.replace(/\r\n/g, "\n"), state = owner(ts, file, text);
    assert([pins.ownerBeforeSha256, pins.ownerAfterSha256].includes(state.hash), "changed or partly adapted directory helper owner");
    const adapted = state.hash === pins.ownerAfterSha256;
    assert.equal(state.parameter.getText(state.source), adapted ? after : before, "directory helper parameter drift");
    if (check && !adapted) throw new Error("directoryProbablyExists host undefined missing");
    let result = adapted ? text : text.slice(0, state.parameter.getStart(state.source)) + after + text.slice(state.parameter.end);
    assert.equal(owner(ts, file, result).hash, pins.ownerAfterSha256, "invalid directory helper result");
    if (raw.includes("\r\n")) result = result.replace(/\n/g, "\r\n");
    for (const removeComments of [false, true]) {
        const compilerOptions = {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments};
        assert.equal(ts.transpileModule(raw, {fileName: file, compilerOptions}).outputText,
            ts.transpileModule(result, {fileName: file, compilerOptions}).outputText,
            "directory helper JavaScript changed");
    }
    return {text: result, assertions: 0, zeros: 0, declined: 0, contracts: adapted ? 0 : 1};
}
module.exports = {plan, before, after, pins, javascriptChanging: false};
