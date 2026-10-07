"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const crypto = require("node:crypto");
const vm = require("node:vm");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const soundness = require("../45-regex-captures/adapt.cjs");
const fix = require("./adapt.cjs");
const pristine = process.argv[2];
assert(pristine, "usage: node test.cjs <pristine TypeScript v6.0.3 checkout>");
const tree = fs.mkdtempSync(path.join(os.tmpdir(), "pragma-empty-fix-"));
fs.mkdirSync(path.join(tree, "src/compiler"), { recursive: true });
for (const file of ["checker", "debug", "emitter", "parser", "semver", "sourcemap", "utilities", "utilitiesPublic", "types"]) {
    fs.copyFileSync(path.join(pristine, "src/compiler", file + ".ts"), path.join(tree, "src/compiler", file + ".ts"));
}
soundness.adapt(tree, ts);
const file = path.join(tree, "src/compiler/parser.ts");
const before = fs.readFileSync(file, "utf8");
assert.equal(fix.adapt(tree, ts).edits, 1);
const after = fs.readFileSync(file, "utf8");
assert.equal(after, before.replace("(matchResult[2] || matchResult[3])!", "(matchResult[2] ?? matchResult[3])!"));
const digest = () => crypto.createHash("sha256").update(fs.readFileSync(file)).digest("hex");
const firstHash = digest();
assert.equal(fix.adapt(tree, ts).edits, 0);
assert.equal(digest(), firstHash);
console.log("PASS: one operator-only edit, second pass 0 edits and identical parser hash");
function evaluate(text, input) {
    const source = ts.createSourceFile("parser.ts", text, ts.ScriptTarget.Latest, true);
    let template;
    let expression;
    function visit(node) {
        if (ts.isVariableDeclaration(node) && node.name.getText() === "result" && node.initializer && ts.isNewExpression(node.initializer) &&
            node.initializer.expression.getText() === "RegExp") template = node.initializer.getText();
        if (ts.isVariableDeclaration(node) && node.name.getText() === "value" && node.initializer &&
            node.initializer.getText().startsWith("(matchResult[2]")) expression = node.initializer.getText();
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert(template && expression);
    const regex = vm.runInNewContext(template, { name: "path" });
    const matchResult = regex.exec(input);
    if (!matchResult) return "no match";
    const javascript = ts.transpileModule(`const result = ${expression};`, {}).outputText + "\nresult;";
    return vm.runInNewContext(javascript, { matchResult });
}
for (const input of [" path='x'", ' path="x"', ' path=""', " path='a b'", " path='0'", ' no-path', " path=''"]) {
    const previous = evaluate(before, input);
    const current = evaluate(after, input);
    if (input === " path=''") {
        assert.equal(previous, undefined);
        assert.equal(current, "");
        assert.equal(current.length, 0);
    }
    else assert.equal(current, previous);
    console.log(`${JSON.stringify(input)}: before=${JSON.stringify(previous)}, after=${JSON.stringify(current)}`);
}
// A put-back mutant must fail the empty-value expectation, independently of the adapter.
assert.throws(() => assert.equal(evaluate(before, " path=''"), ""), assert.AssertionError);
console.log("PASS: || put-back mutant caught only by empty-single-quote expected value");
const definitionsFile = path.join(tree, "src/compiler/types.ts");
const definitions = fs.readFileSync(definitionsFile, "utf8");
assert(definitions.includes('name: "factory"'));
fs.writeFileSync(definitionsFile, definitions.replace('name: "factory"', 'name: "factory(extra)"'));
assert.throws(() => fix.adapt(tree, ts), /pragma argument names may inject regex syntax/);
assert.throws(() => soundness.adapt(tree, ts), /pragma argument names may inject regex syntax/);
assert.equal(digest(), firstHash);
console.log("PASS: regex-injecting pragma-name mutant declined by both adapters; parser hash unchanged");
console.log(`Temporary probe copy: ${tree}`);
