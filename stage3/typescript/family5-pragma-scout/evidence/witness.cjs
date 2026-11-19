const assert = require("node:assert/strict");
const ts = require(process.argv[2]);
const out = [];
function parse(text) { return ts.createSourceFile("witness.tsx", text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX); }
for (const key of ["jsx", "jsxfrag", "jsxruntime", "jsximportsource", "ts-check", "ts-nocheck"]) {
    const text = "/* @" + key + "";
    const input = key.startsWith("ts-") ? "// @" + key + " extra\n" : text;
    const args = parse(input).pragmas.get(key).arguments;
    assert.deepEqual(Object.keys(args), []);
    out.push({ input, keys: Object.keys(args) });
}
for (const [field, expected] of [["path", undefined], ["name", undefined]]) {
    const input = field === "path" ? "/// <amd-dependency path='' />" : "/// <amd-module name='' />";
    const args = parse(input).pragmas.values().next().value.arguments;
    assert(Object.hasOwn(args, field));
    assert.equal(args[field], expected);
    out.push({ input, ownKey: field, value: String(args[field]) });
}
for (const input of ["/// <reference types='' />", "/// <reference lib='' />", "/// <reference path='' />"]) {
    assert.throws(() => parse(input), TypeError);
    try { parse(input); } catch (e) { out.push({input, error: String(e)}); }
}
const foreign = parse('/// <reference path="a.ts" name="ignored" />').pragmas.get("reference").arguments;
assert.deepEqual(Object.keys(foreign), ["path"]);
out.push({foreignAttributeIgnored: true, path: foreign.path});
try { out.push({ missingFactoryEntity: String(ts.parseIsolatedEntityName(undefined, ts.ScriptTarget.Latest)) }); }
catch (e) { out.push({ missingFactoryError: String(e) }); }
console.log(JSON.stringify(out, null, 2));
