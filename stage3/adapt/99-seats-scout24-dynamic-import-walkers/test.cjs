"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node test.cjs <applied tree>");
const tree = path.resolve(process.argv[2]);
const reference = "refs/stage3/99-seats-scout24-dynamic-import-walkers";
const original = "forEachDynamicImportOrRequireCall";
const value = original + "Expression";

function evaluate(side, names, visits) {
    const sourceText = execFileSync("git", ["-C", tree, "show",
        `${reference}-${side}:src/compiler/utilities.ts`], { encoding: "utf8" });
    const source = ts.createSourceFile("utilities.ts", sourceText, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);
    function select(name) {
        const declarations = source.statements.filter(node =>
            ts.isFunctionDeclaration(node) && node.body && node.name?.text === name);
        assert.equal(declarations.length, 1, name);
        return declarations[0].getText(source).replace(/^export /, "");
    }
    const helper = select("getNodeAtPosition").replace("function getNodeAtPosition(", "function locate(");
    const code = [helper, ...names.map(select), `({ ${names.join(", ")} });`].join("\n");
    const helpers = Object.fromEntries([
        "isInJSFile", "isRequireCall", "isImportCall", "isStringLiteralLike", "isStringLiteral",
        "isLiteralImportTypeNode", "isJSDocImportTag", "getExternalModuleName",
        "hasJSDocNodes", "forEach", "forEachChild", "isMetaProperty", "SyntaxKind",
    ].map(name => [name, ts[name]]));
    const context = vm.createContext({
        ...helpers,
        getNodeAtPosition(file, offset, includeJSDoc) {
            visits.push([offset, includeJSDoc]);
            return context.locate(file, offset, includeJSDoc);
        },
    });
    const emitted = ts.transpileModule(code, { compilerOptions: { target: ts.ScriptTarget.ES2020 },
        reportDiagnostics: true });
    assert.deepEqual(emitted.diagnostics, []);
    return vm.runInContext(emitted.outputText, context);
}

const sample = [
    'require("a");',
    'require(name);',
    'import("b");',
    'import(`c`);',
    'import(d);',
    'import();',
    'import type x = require("e");',
    'type Ret = typeof import("f");',
    'const importrequire = "import require";',
    '// import require',
];
const callers = [
    ["transformModule", false, false],
    ["transformECMAScriptModule", false, false],
    ["program", true, true],
];

for (const extension of ["ts", "js"]) {
    const text = [...sample, ...(extension === "js" ? ['/** @import { J } from "g" */', "const j = 0;"] : [])].join("\n");
    const file = ts.createSourceFile(`witness.${extension}`, text, ts.ScriptTarget.Latest, true);
    for (const [caller, includeTypeSpace, requireLiteral] of callers) {
        const beforeVisits = [], afterVisits = [];
        const before = evaluate("before", [original], beforeVisits);
        const after = evaluate("after", [original, value], afterVisits);
        const collect = array => (node, argument) =>
            array.push([ts.SyntaxKind[node.kind], node.pos, argument.getText(file)]);
        const expected = [], actual = [];
        before[original](file, includeTypeSpace, requireLiteral, collect(expected));
        if (includeTypeSpace) after[original](file, true, true, collect(actual));
        else after[value](file, collect(actual));
        assert.deepEqual(actual, expected, `${caller} .${extension} callback sequence`);
        assert.deepEqual(afterVisits, beforeVisits, `${caller} .${extension} scanner calls`);
        const args = expected.map(([, , argument]) => argument);
        assert(args.includes('"b"') && args.includes('`c`'));
        assert.equal(args.includes("d"), !requireLiteral);
        assert.equal(args.includes("name"), extension === "js" && !requireLiteral);
        assert.equal(args.includes('"f"'), includeTypeSpace);
        assert.equal(args.includes('"g"'), extension === "js" && includeTypeSpace);
        console.log(`PASS ${caller} .${extension}: ${JSON.stringify(actual)}`);
    }
}
