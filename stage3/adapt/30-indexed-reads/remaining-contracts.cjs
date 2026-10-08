"use strict";
// Executable counterexamples for the remaining declines, using real core bodies.
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
assert.equal(ts.version, "6.0.3");
const tree = process.argv[2];
assert(tree, "usage: remaining-contracts.cjs <adapted-tree>");
const text = fs.readFileSync(path.join(tree, "src/compiler/core.ts"), "utf8");
const source = ts.createSourceFile("core.ts", text, ts.ScriptTarget.Latest, true);
function body(name) {
    const found = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name.text === name && n.body);
    assert.equal(found.length, 1, "real core owner drift: " + name);
    return ts.transpileModule(found[0].getText(source), { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
}
let rangeJS = body("addRange");
if (process.argv[3] === "--mutant-cache") rangeJS = rangeJS.replace("if (from[i] !== undefined)", "const value = from[i]; if (value !== undefined)").replace("to.push(from[i])", "to.push(value)");
const addRange = new Function("exports", "toOffset", rangeJS + "; return exports.addRange;")({}, () => { throw new Error("offset branch not used"); });
function observe(fn, accessor) {
    let reads = 0;
    const from = [23], to = [];
    if (accessor) Object.defineProperty(from, "0", { get() { return ++reads === 1 ? 23 : undefined; }, configurable: true });
    else Object.defineProperty(to, "push", { get() { delete from[0]; return Array.prototype.push; } });
    const result = fn(to, from);
    assert.equal(result, to);
    return { reads, length: to.length, own: Object.hasOwn(to, "0"), value: String(to[0]) };
}
const indexedGetter = observe(addRange, true), pushGetter = observe(addRange, false);
assert.deepEqual(indexedGetter, { reads: 2, length: 1, own: true, value: "undefined" });
assert.deepEqual(pushGetter, { reads: 0, length: 1, own: true, value: "undefined" });
// The proposed cache changes both the value and the number of source reads.
assert(rangeJS.includes("if (from[i] !== undefined)"));
assert(rangeJS.includes("to.push(from[i])"));
const cachedJS = rangeJS.replace("if (from[i] !== undefined)", "const value = from[i]; if (value !== undefined)").replace("to.push(from[i])", "to.push(value)");
const cachedRange = new Function("exports", "toOffset", cachedJS + "; return exports.addRange;")({}, () => { throw new Error("offset branch not used"); });
assert.notDeepEqual(observe(cachedRange, true), indexedGetter, "cache mutant must change Node observation");
assert.notDeepEqual(observe(cachedRange, false), pushGetter, "cache mutant must change intervening-write observation");
const createSet = new Function("exports", "isArray", "contains", "arrayFrom", "unorderedRemoveItemAt", body("createSet") + "; return exports.createSet;")({}, Array.isArray,
    (array, value, equals) => array.some(item => equals(item, value)), Array.from,
    () => { throw new Error("removal branch not used"); });
const set = createSet(x => x, (a, b) => a === b);
const missing = ["union", "intersection", "difference", "symmetricDifference", "isSubsetOf", "isSupersetOf", "isDisjointFrom"];
assert(missing.every(name => !(name in set)));
assert.equal(set.add(23), set);
assert(set.has(23));
const session = fs.readFileSync(path.join(tree, "src/server/session.ts"), "utf8");
assert(session.includes("function createDocumentSpanSet(useCaseSensitiveFileNames: boolean): Set<DocumentSpan>"), "external Set consumer contract drift");
function genericDiagnostic(member, type) {
    // Even an undefined-admitting constraint cannot guarantee every narrower T.
    const name = path.resolve("/tmp/stage3-generic-contract-probe.ts");
    const probe = `type Mutable<T> = { -readonly [P in keyof T]: T[P] }; interface Owner { ${member}?: ${type} | undefined; } function write<T extends Owner>(node: Mutable<T>, value: ${type} | undefined) { node.${member} = value; } interface Narrow extends Owner { ${member}: { proof: true }; } const node: Mutable<Narrow> = { ${member}: { proof: true } }; write<Narrow>(node, undefined);`;
    const options = { strict: true, exactOptionalPropertyTypes: true, noEmit: true, target: ts.ScriptTarget.ES2024 };
    const host = ts.createCompilerHost(options), read = host.getSourceFile.bind(host);
    host.getSourceFile = (file, language, error, fresh) => file === name ? ts.createSourceFile(file, probe, language, true) : read(file, language, error, fresh);
    const program = ts.createProgram([name], options, host);
    const diagnostics = program.getSemanticDiagnostics(program.getSourceFile(name));
    assert.equal(diagnostics.length, 0, "stock checker behavior drift for generic widening witness");
    const js = ts.transpileModule(probe, { compilerOptions: options }).outputText;
    assert.equal(new Function(js + `; return node.${member};`)(), undefined);
    return { stock_diagnostics: 0, declared_required_member: "{ proof: true }", actual_Node_value: "undefined", conclusion: "stock acceptance of the wider constraint does not prove the narrower T's return/write contract" };
}
console.log(JSON.stringify({ status: "pass: declines supported by counterexamples", indexedGetter, pushGetter,
    cache_mutant: "Node values/read traces differ in both cases", missingSetMethods: missing,
    externalSetConsumer: "src/server/session.ts:512 requires Set<DocumentSpan>",
    genericWrites: { localSymbol: genericDiagnostic("localSymbol", "object"), typeExpression: genericDiagnostic("typeExpression", "object") },
    source_repairs: 0 }, null, 2));
