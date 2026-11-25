"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node test.cjs <full apply output>");
const tree = process.argv[2];
const ref = "refs/stage3/99-seats-tail64-function-body-switch";
const types = "src/compiler/types.ts";
const utilities = "src/compiler/utilities.ts";
const serializer = "src/compiler/transformers/typeSerializer.ts";
const files = [types, utilities, serializer];
const helperName = "bodyOfSignatureForSerialization";
const snapshot = phase => Object.fromEntries(files.map(file =>
    [file, execFileSync("git", ["-C", tree, "show", `${ref}-${phase}:${file}`])]));
const before = snapshot("before");
const after = snapshot("after");
const changed = execFileSync("git", ["-C", tree, "diff", "--name-only", ref + "-before", ref + "-after"],
    { encoding: "utf8" }).trim().split("\n").sort();
assert.deepEqual(changed, [utilities, serializer].sort());

function withFixture(bytes, action) {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "tail64-test-"));
    try {
        for (const file of files) {
            const destination = path.join(root, file);
            fs.mkdirSync(path.dirname(destination), { recursive: true });
            fs.writeFileSync(destination, bytes[file]);
        }
        action(root);
    } finally {
        fs.rmSync(root, { recursive: true, force: true });
    }
}
const apply = root => spawnSync(process.execPath, [path.join(__dirname, "adapt.cjs"), root], { encoding: "utf8" });
function assertBytes(root, bytes) {
    for (const file of files) assert.deepEqual(fs.readFileSync(path.join(root, file)), bytes[file], file);
}
withFixture(before, root => {
    const first = apply(root);
    assert.equal(first.status, 0, first.stderr);
    assertBytes(root, after);
    console.log("PASS scratch application matches both after snapshots byte for byte");
    const marks = files.map(file => fs.statSync(path.join(root, file), { bigint: true }).mtimeNs);
    const second = apply(root);
    assert.equal(second.status, 0, second.stderr);
    assertBytes(root, after);
    assert.deepEqual(files.map(file => fs.statSync(path.join(root, file), { bigint: true }).mtimeNs), marks);
    console.log("PASS second application leaves bytes and write times unchanged");
});

const parse = (bytes, name) => ts.createSourceFile(name, bytes[name].toString(), ts.ScriptTarget.Latest, true);
function collect(root, predicate) {
    const result = [];
    function visit(node) {
        if (predicate(node)) result.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return result;
}
function one(items) {
    assert.equal(items.length, 1);
    return items[0];
}
function namedFunction(file, name) {
    return one(collect(file, node => ts.isFunctionDeclaration(node) && node.name?.text === name));
}
function bodyCondition(file) {
    const owner = namedFunction(file, "serializeParameterTypesOfNode");
    return one(collect(owner.body, node => ts.isBinaryExpression(node)
        && node.operatorToken.kind === ts.SyntaxKind.AmpersandAmpersandToken
        && ts.isCallExpression(node.left) && node.left.expression.getText(file) === "isFunctionLike"
        && ts.isCallExpression(node.right) && node.right.expression.getText(file) === "nodeIsPresent"));
}
function checkSyntax() {
    const flags = parse(after, utilities);
    const flagsSwitch = one(namedFunction(flags, "getFunctionFlags").body.statements.filter(node =>
        ts.isSwitchStatement(node) && collect(node, ts.isDefaultClause).length));
    const serializerFile = parse(after, serializer);
    const helper = namedFunction(serializerFile, helperName);
    assert(helper.parent === serializerFile, "helper must be a top-level statement of typeSerializer.ts");
    assert(!helper.modifiers?.some(token => token.kind === ts.SyntaxKind.ExportKeyword), "helper must stay file-private");
    for (const node of [flagsSwitch, bodyCondition(serializerFile), helper]) {
        const casts = collect(node, child => ts.isAsExpression(child) || ts.isTypeAssertionExpression(child)
            || ts.isNonNullExpression(child)).map(child => {
            const file = child.getSourceFile();
            const { line } = file.getLineAndCharacterOfPosition(child.getStart(file));
            const kind = ts.SyntaxKind[child.kind];
            return [file.fileName, line + 1, kind].join(":");
        });
        assert.deepEqual(casts, [], "adapter output must be cast-free");
    }
    const emitted = ts.transpileModule(serializerFile.text, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText;
    assert(!emitted.includes(`exports.${helperName}`), "helper must stay out of runtime exports");
    console.log("PASS edited syntax is cast-free and the helper is file-private");
}
checkSyntax();

function evaluate(bytes, adapted) {
    const flags = parse(bytes, utilities);
    const serialize = parse(bytes, serializer);
    const flagText = namedFunction(flags, "getFunctionFlags").getText(flags).replace(/^export /, "");
    const helper = adapted ? namedFunction(serialize, helperName).getText(serialize) : "";
    const expression = bodyCondition(serialize).getText(serialize);
    const text = `${flagText}\n${helper}\nfunction present(node: Node) { return ${expression}; }`;
    const emitted = ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText;
    return vm.runInNewContext(emitted + "\n({ getFunctionFlags, present })", {
        SyntaxKind: ts.SyntaxKind, FunctionFlags: ts.FunctionFlags, ModifierFlags: ts.ModifierFlags,
        hasSyntacticModifier: ts.hasSyntacticModifier, isFunctionLike: ts.isFunctionLike,
        nodeIsPresent: ts.nodeIsPresent,
    });
}

const f = ts.factory;
const block = f.createBlock([]);
const type = f.createKeywordTypeNode(ts.SyntaxKind.VoidKeyword);
const arrow = f.createToken(ts.SyntaxKind.EqualsGreaterThanToken);
const jsDocSignature = f.createJSDocSignature(undefined, [], undefined);
const cases = [
    ["JSDoc signature predicate control", jsDocSignature],
    ["constructor", f.createConstructorDeclaration(undefined, [], block)],
    ["constructor without body", f.createConstructorDeclaration(undefined, [], undefined)],
    ["get", f.createGetAccessorDeclaration(undefined, "x", [], undefined, block)],
    ["get without body", f.createGetAccessorDeclaration(undefined, "x", [], undefined, undefined)],
    ["set", f.createSetAccessorDeclaration(undefined, "x", [], block)],
    ["set without body", f.createSetAccessorDeclaration(undefined, "x", [], undefined)],
    ["call signature", f.createCallSignature(undefined, [], type)],
    ["construct signature", f.createConstructSignature(undefined, [], type)],
    ["method signature", f.createMethodSignature(undefined, "m", undefined, undefined, [], type)],
    ["index signature", f.createIndexSignature(undefined, [], type)],
    ["function type", f.createFunctionTypeNode(undefined, [], type)],
    ["constructor type", f.createConstructorTypeNode(undefined, undefined, [], type)],
    ["JSDoc function type", f.createJSDocFunctionType([], type)],
];
for (const async of [false, true]) {
    const modifiers = async ? [f.createModifier(ts.SyntaxKind.AsyncKeyword)] : undefined;
    cases.push([`arrow block async=${async}`, f.createArrowFunction(modifiers, undefined, [], undefined, arrow, block)]);
    cases.push([`arrow expression async=${async}`, f.createArrowFunction(modifiers, undefined, [], undefined, arrow, f.createNumericLiteral(1))]);
    for (const generator of [false, true]) {
        const star = generator ? f.createToken(ts.SyntaxKind.AsteriskToken) : undefined;
        const suffix = ` async=${async} generator=${generator}`;
        cases.push(["function expression" + suffix, f.createFunctionExpression(modifiers, star, undefined, undefined, [], undefined, block)]);
        for (const body of [block, undefined]) {
            cases.push(["function declaration" + suffix + ` body=${!!body}`,
                f.createFunctionDeclaration(modifiers, star, "fun", undefined, [], undefined, body)]);
            cases.push(["method declaration" + suffix + ` body=${!!body}`,
                f.createMethodDeclaration(modifiers, star, "method", undefined, undefined, [], undefined, body)]);
        }
    }
}
const factoryCount = cases.length;
const recoveredFile = ts.createSourceFile("recovery.ts", "class C { @a set m(x: string) @b n() {} }", ts.ScriptTarget.Latest, true);
const recoveredSetter = one(collect(recoveredFile, node => ts.isSetAccessorDeclaration(node) && node.name.getText(recoveredFile) === "m"));
assert(recoveredSetter.body && !ts.nodeIsPresent(recoveredSetter.body));
assert(ts.nodeIsPresent(recoveredSetter));
cases.push(["decorated setter with a recovered zero-width body", recoveredSetter]);

const missingFunctionFile = ts.createSourceFile("missing.ts", "const f = function()", ts.ScriptTarget.Latest, true);
const missingFunction = one(collect(missingFunctionFile, ts.isFunctionExpression));
assert(missingFunction.body && !ts.nodeIsPresent(missingFunction.body));
assert(ts.nodeIsPresent(missingFunction));
cases.push(["function expression with a recovered zero-width body", missingFunction]);

const parsedFile = ts.createSourceFile("present.ts", [
    "function f() {}",
    "class C { m() {} constructor() {} get g() { return 1; } set s(x: number) {} }",
    "const expression = function() {}, blockArrow = () => {}, expressionArrow = () => 1;",
    "interface I { m(): void; }",
].join("\n"), ts.ScriptTarget.Latest, true);
assert.equal(parsedFile.parseDiagnostics.length, 0);
for (const [label, predicate] of [
    ["function declaration", ts.isFunctionDeclaration],
    ["method declaration", ts.isMethodDeclaration],
    ["constructor", ts.isConstructorDeclaration],
    ["getter", ts.isGetAccessorDeclaration],
    ["setter", ts.isSetAccessorDeclaration],
    ["function expression", ts.isFunctionExpression],
    ["block arrow", node => ts.isArrowFunction(node) && ts.isBlock(node.body)],
    ["expression arrow", node => ts.isArrowFunction(node) && !ts.isBlock(node.body)],
]) {
    const input = one(collect(parsedFile, predicate));
    assert(ts.nodeIsPresent(input.body), label);
    cases.push(["parsed " + label, input]);
}
const parsedSignature = one(collect(parsedFile, ts.isMethodSignature));
assert(parsedSignature.body === undefined, "parsed method signature has no body");
cases.push(["parsed method signature without a body", parsedSignature]);

const old = evaluate(before, false);
const next = evaluate(after, true);
for (const [label, input] of cases) {
    assert.equal(next.getFunctionFlags(input), old.getFunctionFlags(input), `${label} flags`);
    assert.equal(next.present(input), old.present(input), `${label} body presence`);
}
assert.equal(next.getFunctionFlags(undefined), old.getFunctionFlags(undefined));
assert.equal(ts.isFunctionLike(jsDocSignature), true);
for (const version of [old, next]) {
    assert.equal(version.getFunctionFlags(jsDocSignature), ts.FunctionFlags.Invalid);
    assert.equal(version.present(jsDocSignature), false);
}
console.log(`PASS value identity for ${factoryCount} factory nodes and ${cases.length - factoryCount} parsed nodes plus missing input for flags`);

function rejectFixture(label, changedTypes, expectedMessage) {
    const bytes = { ...before, [types]: Buffer.from(changedTypes) };
    withFixture(bytes, root => {
        const result = apply(root);
        assert.equal(result.status, 1, `${label} should refuse before writing`);
        assert.match(result.stderr, expectedMessage);
        assertBytes(root, bytes);
    });
    console.log(`PASS ${label} refused before either file changed`);
}
const typeFile = parse(before, types);
const originalTypes = typeFile.text;
const coverageMessage = /Body-bearing SignatureDeclaration members differ from FunctionLikeDeclaration/;
const signatureUnion = one(typeFile.statements.filter(node =>
    ts.isTypeAliasDeclaration(node) && node.name.text === "SignatureDeclaration"));
rejectFixture("added body-bearing signature kind",
    originalTypes.slice(0, signatureUnion.type.end) + " | AddedTail64Signature" +
    originalTypes.slice(signatureUnion.type.end) +
    "\ninterface AddedTail64Signature extends SignatureDeclarationBase { readonly kind: SyntaxKind.SourceFile; readonly body: Block; }\n", coverageMessage);
const declaration = one(typeFile.statements.filter(node =>
    ts.isInterfaceDeclaration(node) && node.name.text === "FunctionDeclaration"));
const body = one(declaration.members.filter(node => node.name?.getText(typeFile) === "body"));
const replacement = (originalTypes.slice(declaration.getStart(typeFile), body.getStart(typeFile)) +
    originalTypes.slice(body.end, declaration.end)).replace("extends FunctionLikeDeclarationBase", "extends SignatureDeclarationBase");
assert(!replacement.includes("FunctionLikeDeclarationBase"));
rejectFixture("removed effective body declaration",
    originalTypes.slice(0, declaration.getStart(typeFile)) + replacement + originalTypes.slice(declaration.end), coverageMessage);

const callSignature = one(typeFile.statements.filter(node =>
    ts.isInterfaceDeclaration(node) && node.name.text === "CallSignatureDeclaration"));
rejectFixture("computed body property on an existing signature kind",
    originalTypes.slice(0, callSignature.members.end) + '\n    readonly ["body"]?: Block | undefined;\n' +
    originalTypes.slice(callSignature.members.end), coverageMessage);
rejectFixture("unresolved computed property on an existing signature kind",
    originalTypes.slice(0, callSignature.members.end) + '\n    readonly [bodyName]?: Block | undefined;\n' +
    originalTypes.slice(callSignature.members.end) + '\ndeclare const bodyName: "body";\n',
    /Unresolved computed property name on CallSignatureDeclaration: \[bodyName\]/);
