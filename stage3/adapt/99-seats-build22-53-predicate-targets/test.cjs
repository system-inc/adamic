"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const pins = require("./premises.json");

const applied = process.argv[2];
assert(applied, "usage: node test.cjs <tree written by stage3/apply.sh>");
const refs = "refs/stage3/99-seats-build22-53-predicate-targets";
const root = fs.mkdtempSync(path.join(os.tmpdir(), "seats53-"));
const util = "utilities.ts";
const publicUtil = "utilitiesPublic.ts";
const diag = "transformers/declarations/diagnostics.ts";
const files = [...new Set(["_namespaces/ts.ts", ...pins.map(pin => pin.file)])];
const sourcePath = file => "src/compiler/" + file;
const fileIn = (tree, file) => path.join(tree, sourcePath(file));

function snapshot(phase, file) {
    return execFileSync("git", ["-C", applied, "show", `${refs}-${phase}:${sourcePath(file)}`]);
}

function digest(tree) {
    return files.map(file => createHash("sha256").update(fs.readFileSync(fileIn(tree, file))).digest("hex"));
}

function apply(tree) {
    const result = spawnSync(process.execPath, [path.join(__dirname, "adapt.cjs"), tree], { encoding: "utf8" });
    assert.ifError(result.error);
    assert.equal(result.signal, null);
    return result;
}

function exactReplace(tree, file, before, after) {
    const target = fileIn(tree, file);
    const text = fs.readFileSync(target, "utf8");
    assert.equal(text.split(before).length, 2, before);
    fs.writeFileSync(target, text.replace(before, after));
}

function guardControls(base, head) {
    const controls = [
        ["fourth-kind", diag, "isPropertyDeclaration(n) || isParameter(n));",
            "isPropertyDeclaration(n) || isParameter(n) || isBindingElement(n));"],
        ["removed-export", diag, "if (isExportAssignment(result)) return result;", ""],
        ["accessor-case", publicUtil,
            "function getNonAssignedNameOfDeclaration(declaration: Declaration | Expression): DeclarationName | undefined {\r\n    switch (declaration.kind) {",
            "$&\r\n        case SyntaxKind.GetAccessor: return undefined;"],
        ["alias-import", diag, "    isStatement,", "    isVariableDeclaration as isStatement,"],
        ["accessor-alias-import", util, "    isElementAccessExpression,",
            "    isPropertyAccessExpression as isElementAccessExpression,"],
        ["barrel-value", "_namespaces/ts.ts", 'export * from "../corePublic.js";',
            'import { isVariableDeclaration as auditReplacement } from "../factory/nodeTests.js";\r\n' +
            'Object.defineProperty(exports, "isStatement", { get: () => auditReplacement });\r\n$&'],
        ["barrel-side-effect", "_namespaces/ts.ts", 'export * from "../corePublic.js";',
            'export * from "./auditAlternate.js";\r\n$&'],
        ["barrel-unchecked", "_namespaces/ts.ts",
            '/* Generated file to emulate the ts namespace. */\r\n\r\nexport * from "../corePublic.js";',
            '// @ts-nocheck\r\n/* Generated file to emulate the ts namespace. */\r\n\r\n' +
            'export * from "./auditAlternate.js";\r\nexport * from "../corePublic.js";'],
        ["barrel-order", "_namespaces/ts.ts",
            'export * from "../corePublic.js";\r\nexport * from "../core.js";',
            'export * from "../core.js";\r\nexport * from "../corePublic.js";'],
    ];
    for (const [label, file, before, after] of controls) {
        for (const [state, original] of [["first", base], ["repeat", head]]) {
            const tree = path.join(root, label + "-" + state);
            fs.cpSync(original, tree, { recursive: true });
            exactReplace(tree, file, before, after);
            const hashes = digest(tree);
            const result = apply(tree);
            const unchanged = digest(tree).every((hash, index) => hash === hashes[index]);
            console.log(label, state, "exit", result.status, "unchanged", unchanged,
                (result.stderr || result.stdout).split("\n").find(line => line.includes("premise drift")) || "");
            assert.notEqual(result.status, 0, `${label} ${state}: adapter accepted premise drift`);
            assert.match(result.stderr, /premise drift/, `${label} ${state}: rejection reason`);
            assert(unchanged, `${label} ${state}: rejection modified a source file`);
        }
    }
}

function implementation(tree, file, name) {
    const text = fs.readFileSync(fileIn(tree, file), "utf8");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const matches = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === name && node.body) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(matches.length, 1, name);
    return matches[0].getText(source).replace(/^export /, "");
}

function load(tree, declarations) {
    const code = declarations.map(([file, name]) => implementation(tree, file, name)).join("\n");
    const emit = ts.transpileModule(code, {
        compilerOptions: { target: ts.ScriptTarget.ES2024 },
        reportDiagnostics: true,
    });
    assert.deepEqual(emit.diagnostics, []);
    const context = { ...ts };
    vm.runInNewContext(emit.outputText, context, { timeout: 1000 });
    return context;
}

const K = ts.SyntaxKind;
const parse = text => ts.createSourceFile("fixture.ts", text, ts.ScriptTarget.Latest, true);
function pick(source, kind) {
    let found;
    function visit(node) {
        if (node.kind === kind) found ??= node;
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert(found, K[kind]);
    return found;
}

function verifyStatement(tree) {
    const { findNearestDeclaration } = load(tree, [[diag, "findNearestDeclaration"]]);
    for (const [text, kind, label] of [
        ["function f(p: number) {}", K.Parameter, "nearest Parameter identity"],
        ["let v = 1;", K.VariableDeclaration, "nearest VariableDeclaration identity"],
        ["class C { v = 1; }", K.PropertyDeclaration, "nearest PropertyDeclaration identity"],
        ["export default 1;", K.ExportAssignment, "nearest ExportAssignment identity"],
    ]) {
        const node = pick(parse(text), kind);
        assert.strictEqual(findNearestDeclaration(node), node, label);
    }
    for (const [text, kind, label] of [
        ["throw 1;", K.ThrowStatement, "throw is excluded"],
        ["class C { constructor(){ return; } }", K.ReturnStatement, "constructor return is excluded"],
        ["{ let x = 0; }", K.Block, "ordinary block is excluded"],
    ]) {
        assert.equal(findNearestDeclaration(pick(parse(text), kind)), undefined, label);
    }
    const source = parse("function f(){ return 1; }");
    assert.strictEqual(findNearestDeclaration(pick(source, K.ReturnStatement)), source.statements[0], "return's function identity");
}

function verifyAccessor(tree) {
    const { getAllAccessorDeclarations } = load(tree, [
        [publicUtil, "getNonAssignedNameOfDeclaration"], [publicUtil, "getNameOfDeclaration"],
        [util, "hasDynamicName"], [util, "isDynamicName"], [util, "getAllAccessorDeclarations"],
    ]);
    for (const [key, dynamic] of [["key", true], ['"key"', false], ["1", false], ["-1", false]]) {
        const members = parse(`class C { get [${key}](){ return 0; } set [${key}](v){} }`).statements[0].members;
        for (const member of members) {
            const result = getAllAccessorDeclarations(members, member);
            assert.strictEqual(result.firstAccessor, dynamic ? member : members[0], key);
            assert.strictEqual(result.secondAccessor, dynamic ? undefined : members[1], key);
        }
    }
}

try {
    const base = path.join(root, "base");
    const head = path.join(root, "head");
    for (const file of files) {
        const output = fileIn(base, file);
        fs.mkdirSync(path.dirname(output), { recursive: true });
        fs.writeFileSync(output, snapshot("before", file));
    }
    fs.cpSync(base, head, { recursive: true });
    const first = apply(head);
    assert.equal(first.status, 0, first.stderr);
    const once = digest(head);
    const repeated = apply(head);
    assert.equal(repeated.status, 0, repeated.stderr);
    assert.deepEqual(digest(head), once, "reapply bytes");
    console.log(first.stdout.trim());
    console.log(repeated.stdout.trim());
    guardControls(base, head);
    for (const tree of [base, head]) {
        verifyStatement(tree);
        verifyAccessor(tree);
    }
    console.log("PASS statement node identity and 8 accessor cases on both trees");
    for (const file of [util, diag]) {
        assert(fs.readFileSync(fileIn(head, file)).equals(snapshot("after", file)),
            `${sourcePath(file)} differs from the pipeline's after snapshot`);
    }
    console.log("PASS both files match the pipeline's after snapshot");
} finally {
    fs.rmSync(root, { recursive: true, force: true });
}
