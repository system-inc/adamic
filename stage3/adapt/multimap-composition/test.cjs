"use strict";
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const {adapt} = require("./adapt.cjs");
const root = path.resolve(__dirname, "../../..");
const original = fs.readFileSync(path.join(__dirname, "original.a"), "utf8");
const port = fs.readFileSync(path.join(root, "stage1/typescript/collections/multimap.a"), "utf8");
const fixture = fs.readFileSync(path.join(root, "internal/oracle/testdata/scout19_slice2_multimap.a"), "utf8");
const driver = fixture.slice(fixture.indexOf("const receiver"));
const legacyDriver = driver.replace(/multiMapAdd\((receiver|other), /g, "$1.add(").replace(/multiMapRemove\((receiver|other), /g, "$1.remove(").replace(/multiMapForEach\(receiver, /g, "receiver.forEach(").replace(/(receiver|other)\.map/g, "$1");
const dependency = `function unorderedRemoveItem<T>(array: T[], value: T): void { const index = array.indexOf(value); if (index !== -1) { array[index] = array[array.length - 1]; array.pop(); } }`;
function observe(source) {
    const output = [];
    vm.runInNewContext(ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.CommonJS}}).outputText, {exports: {}, console: {log: text => output.push(text)}});
    return output;
}
assert.deepEqual(observe(port + driver), observe(original + dependency + legacyDriver));
// Undefined is an element too; a nullish fallback would silently change this case.
const undefinedDriver = `const receiver = createMultiMap<string, number | undefined>(); multiMapAdd(receiver, 'x', 1); multiMapAdd(receiver, 'x', undefined); multiMapRemove(receiver, 'x', 1); console.log(receiver.map.get('x')?.length); console.log(receiver.map.get('x')?.[0] === undefined); multiMapRemove(receiver, 'x', undefined); console.log(receiver.map.has('x'));`;
assert.deepEqual(observe(port + undefinedDriver), [1, true, false]);
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), "scout19-composition-"));
try {
    fs.mkdirSync(path.join(temporary, "src"));
    fs.cpSync(path.join(root, "cohere/TypeScript/tsc/testdata/fixtures/compiler"), path.join(temporary, "src/compiler"), {recursive: true});
    const setBefore = fs.readFileSync(path.join(temporary, "src/compiler/core.ts"), "utf8").split("export function createSet<")[1].split("\n/**")[0];
    const roots = fs.readdirSync(path.join(temporary, "src/compiler")).filter(name => name.endsWith(".ts")).map(name => path.join(temporary, "src/compiler", name));
    const options = {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.NodeNext, moduleResolution: ts.ModuleResolutionKind.NodeNext};
    const baseline = ts.createProgram(roots, options);
    const budget = new Map();
    const diagnosticKey = diagnostic => `${path.relative(temporary, diagnostic.file?.fileName || "")}:${diagnostic.code}:${ts.flattenDiagnosticMessageText(diagnostic.messageText, " ")}`;
    for (const diagnostic of baseline.getSemanticDiagnostics()) {
        const key = diagnosticKey(diagnostic);
        budget.set(key, (budget.get(key) || 0) + 1);
    }
    const changes = adapt(temporary);
    assert.ok(changes.length > 0);
    assert.deepEqual(adapt(temporary), []);
    const setAfter = fs.readFileSync(path.join(temporary, "src/compiler/core.ts"), "utf8").split("export function createSet<")[1].split("\n/**")[0];
    assert.equal(setAfter, setBefore);
    const program = ts.createProgram(fs.readdirSync(path.join(temporary, "src/compiler")).filter(name => name.endsWith(".ts")).map(name => path.join(temporary, "src/compiler", name)), {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.NodeNext, moduleResolution: ts.ModuleResolutionKind.NodeNext});
    for (const diagnostic of program.getSemanticDiagnostics()) {
        const key = diagnosticKey(diagnostic);
        assert.ok((budget.get(key) || 0) > 0, `adaptation introduced a stock checker diagnostic: ${key}`);
        budget.set(key, budget.get(key) - 1);
    }
    const checker = program.getTypeChecker();
    let remaining = 0;
    const isMultiMap = type => type.isUnion() ? type.types.some(isMultiMap) : ["MultiMap", "RedirectTargetsMap"].includes(type.aliasSymbol?.name || type.symbol?.name);
    for (const source of program.getSourceFiles()) {
        if (!source.fileName.startsWith(temporary)) continue;
        function visit(node) {
            if (ts.isPropertyAccessExpression(node) && node.name.text !== "map" && isMultiMap(checker.getTypeAtLocation(node.expression))) remaining++;
            ts.forEachChild(node, visit);
        }
        visit(source);
    }
    assert.equal(remaining, 0, "every MultiMap receiver goes through composition");
    console.log(JSON.stringify({status: "pass", changes, remaining}));
} finally {
    fs.rmSync(temporary, {recursive: true, force: true});
}
