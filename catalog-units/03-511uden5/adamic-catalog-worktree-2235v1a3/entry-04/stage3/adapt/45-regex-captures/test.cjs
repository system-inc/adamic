"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const crypto = require("node:crypto");
const vm = require("node:vm");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { prove } = require("./proof.cjs");
const { adapt } = require("./adapt.cjs");
const source = process.argv[2];
assert(source, "usage: node test.cjs <pristine TypeScript v6.0.3 checkout>");
const files = ["checker", "debug", "emitter", "parser", "semver", "sourcemap", "utilities", "utilitiesPublic", "types"];
function copy(label) {
    const tree = fs.mkdtempSync(path.join(os.tmpdir(), `regex-captures-${label}-`));
    fs.mkdirSync(path.join(tree, "src/compiler"), { recursive: true });
    for (const file of files) fs.copyFileSync(path.join(source, "src/compiler", file + ".ts"), path.join(tree, "src/compiler", file + ".ts"));
    return tree;
}
function hashes(tree) {
    return files.map(file => crypto.createHash("sha256").update(fs.readFileSync(path.join(tree, "src/compiler", file + ".ts"))).digest("hex"));
}
function mutate(tree, file, before, after) {
    const name = path.join(tree, "src/compiler", file + ".ts");
    const text = fs.readFileSync(name, "utf8");
    assert.equal(text.split(before).length, 2, `want one mutant location: ${before}`);
    fs.writeFileSync(name, text.replace(before, after));
}
for (const [regex, expected] of [
    [/^function\s+([\w$]+)\s*\(/, [1]],
    [/^(?:function\s+([\w$]+)\s*\(|anonymous)$/, []],
    [/(a)?/, []], [/(a)*/, []], [/(a)+/, [1]],
    [/(a){0,2}/, []], [/(a){1,2}/, [1]],
    [/(a|b)/, [1]], [/(?:a|(b))/, []], [/(?:(a)|(b))+/, []],
    [/([~^<>=]|<=|>=)?\s*([a-z0-9-+.*]+)/i, [2]],
    [/(\sname\s*=\s*)(?:(?:'([^']*)')|(?:"([^"]*)"))/, [1]],
]) {
    const proof = prove(regex.source);
    assert(proof.supported);
    assert.deepEqual(proof.mandatory, expected);
}
for (const regex of [/(?=(a))/, /(a)\1/, /(?<name>a)/]) assert.equal(prove(regex.source).supported, false);
console.log("PASS: participation proof requires every branch, respects zero-minimum quantifiers, declines unsupported constructs");
const tree = copy("idempotence");
const first = adapt(tree, ts);
assert.equal(first.files, 8);
assert.equal(first.edits, 19);
const before = hashes(tree);
const second = adapt(tree, ts);
assert.equal(second.edits, 0);
assert.deepEqual(hashes(tree), before);
console.log("PASS: first pass 19 edits / 8 files; second pass 0 edits and identical hashes");
mutate(tree, "debug", "match[1]!", "match[1]");
assert.throws(() => assert.deepEqual(hashes(tree), before), assert.AssertionError);
console.log("PASS: idempotence hash mutant caught despite unchanged edit-count observation");
const alternative = copy("alternative-mutant");
const originalRegex = "/^function\\s+([\\w$]+)\\s*\\(/";
const mutantRegex = "/^(?:function\\s+([\\w$]+)\\s*\\(|anonymous)$/";
mutate(alternative, "debug", originalRegex, mutantRegex);
const inputHashes = hashes(alternative);
const match = vm.runInNewContext(mutantRegex).exec("anonymous");
assert(match);
assert.equal(match[1], undefined);
assert.throws(() => adapt(alternative, ts), /capture can be bypassed/);
assert.deepEqual(hashes(alternative), inputHashes);
console.log("PASS: alternative-branch mutant matches anonymous with undefined capture; actual adapter declines and writes nothing");
const split = copy("split-mutant");
mutate(split, "emitter", "const lines = text.split(/\\r\\n?|\\n/);\r\n        const indentation", "const lines = text.split(/\\r\\n?|(\\n)?/);\r\n        const indentation");
const splitHashes = hashes(split);
assert.throws(() => adapt(split, ts), /declined split with captures/);
assert.deepEqual(hashes(split), splitHashes);
console.log("PASS: optional-capture split mutant declined by density proof; no writes");
const positional = copy("positional-mutant");
mutate(positional, "types", '"jsx": {\r\n        args: [{ name: "factory" }],', '"jsx": {\r\n        args: [{ name: "factory", optional: true }],');
const positionalHashes = hashes(positional);
assert.throws(() => adapt(positional, ts), /declined optional positional pragma argument/);
assert.deepEqual(hashes(positional), positionalHashes);
console.log("PASS: optional positional-argument mutant declines bounds invariant; no writes");
// Erased JavaScript AST equivalence for all edits except the proven major binding
// extraction, which necessarily changes JS syntax. Test those functions below.
function erasedTree(text) {
    const javascript = ts.transpileModule(text, {}).outputText;
    const parsed = ts.createSourceFile("erased.js", javascript, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
    function normalize(node) {
        if (ts.isParenthesizedExpression(node)) return normalize(node.expression);
        const children = [];
        ts.forEachChild(node, child => { children.push(normalize(child)); });
        return children.length ? [node.kind, children] : [node.kind, node.getText(parsed)];
    }
    return crypto.createHash("sha256").update(JSON.stringify(normalize(parsed))).digest("hex");
}
for (const file of files.filter(file => file !== "semver")) {
    const original = fs.readFileSync(path.join(source, "src/compiler", file + ".ts"), "utf8");
    // The failed alternative tree is unchanged; successful tree contains a hash mutant.
    const success = fs.readFileSync(path.join(tree, "src/compiler", file + ".ts"), "utf8");
    assert.equal(erasedTree(original), erasedTree(success), `erased AST changed in ${file}`);
}
console.log("PASS: seven edited files erase to identical JavaScript ASTs modulo parentheses");
console.log(`Temporary probe copies: ${tree}, ${alternative}, ${split}`);
