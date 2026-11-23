"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { adapt } = require("./adapt.cjs");

const parser = "src/compiler/parser.ts";

function comparison(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    const fn = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "tagNamesAreEquivalent");
    assert(fn, "expected tagNamesAreEquivalent");
    const compiled = ts.transpileModule(fn.getText(source), {
        compilerOptions: { module: ts.ModuleKind.CommonJS },
        reportDiagnostics: true,
    });
    assert.deepEqual(compiled.diagnostics, []);
    return vm.runInNewContext(compiled.outputText + "\nexports.tagNamesAreEquivalent", { exports: {}, SyntaxKind: ts.SyntaxKind });
}

function tag(text) {
    const source = ts.createSourceFile("input.tsx", "<" + text + " />", ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
    assert.deepEqual(source.parseDiagnostics, []);
    return source.statements[0].expression.tagName;
}

const cases = [
    ["this", "a", false], ["a", "a", true], ["a", "b", false], ["this", "this", true],
    ["n:a", "n:a", true], ["n:a", "m:a", false], ["n:a", "n:b", false],
    ["this.a", "this.b", false], ["a.b.c", "a.b.c", true], ["a.b.c", "z.b.c", false],
];

function retag(compare) {
    const lhs = tag("a.x");
    const rhs = tag("a.x");
    let reads = 0;
    Object.defineProperty(rhs, "kind", { get() {
        return ++reads === 1 ? ts.SyntaxKind.PropertyAccessExpression : ts.SyntaxKind.Identifier;
    } });
    return compare(lhs, rhs);
}

function changeName(compare) {
    const lhs = tag("a.x");
    const rhs = tag("a.x");
    const otherName = tag("y");
    let name = rhs.name;
    let reads = 0;
    Object.defineProperty(rhs, "name", { get() { return name; } });
    Object.defineProperty(rhs, "kind", { get() {
        if (++reads === 2) name = otherName;
        return ts.SyntaxKind.PropertyAccessExpression;
    } });
    return compare(lhs, rhs);
}

assert.equal(process.argv.length, 3, "usage: node test.cjs <pristine TypeScript v6.0.3 checkout>");
const pristine = process.argv[2];
const tree = fs.mkdtempSync(path.join(os.tmpdir(), "family110-"));
try {
    const fileName = path.join(tree, parser);
    fs.mkdirSync(path.dirname(fileName), { recursive: true });
    fs.copyFileSync(path.join(pristine, parser), fileName);
    const base = comparison(fileName, fs.readFileSync(fileName, "utf8"));
    assert.equal(adapt(tree), 1, "expected one parser write");
    const after = fs.readFileSync(fileName, "utf8");
    const candidate = comparison(fileName, after);
    for (const [left, right, expected] of cases) {
        assert.equal(base(tag(left), tag(right)), expected, `base ${left} / ${right}`);
        assert.equal(candidate(tag(left), tag(right)), expected, `candidate ${left} / ${right}`);
    }
    console.log(`parser tag result oracle: ${cases.length} ordinary-node cases passed`);
    assert.equal(adapt(tree), 0, "expected zero writes on reapplication");
    assert.equal(fs.readFileSync(fileName, "utf8"), after);
    console.log("parser reapplication: byte-identical");

    // The reviewer's getter witnesses document the boundary of the ordinary-node claim.
    for (const [label, witness] of [["stateful rhs.kind", retag], ["rhs.kind changes rhs.name on its second read", changeName]]) {
        const before = witness(base);
        const next = witness(candidate);
        assert.equal(before, true, `base ${label}`);
        assert.equal(next, false, `candidate ${label}`);
        console.log(`getter witness: ${label}: base=${before} candidate=${next}`);
    }
} finally {
    fs.rmSync(tree, { recursive: true, force: true });
}
