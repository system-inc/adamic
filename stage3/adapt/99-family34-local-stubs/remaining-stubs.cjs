"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("../../api/node_modules/typescript");

assert.equal(ts.version, "6.0.3");
const tree = path.resolve(process.argv[2]);
const compiler = path.join(tree, "src/compiler");
const config = ts.readConfigFile(path.join(compiler, "tsconfig.json"), ts.sys.readFile);
assert.ifError(config.error);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, compiler);
assert.equal(parsed.errors.length, 0);
const roots = [...new Set([...parsed.fileNames, ...ts.sys.readDirectory(path.join(tree, "src"), [".ts"])])];
const service = ts.createLanguageService({
    ...ts.sys,
    useCaseSensitiveFileNames: () => ts.sys.useCaseSensitiveFileNames,
    getCurrentDirectory: () => tree,
    getCompilationSettings: () => parsed.options,
    getScriptFileNames: () => roots,
    getScriptVersion: () => "0",
    getScriptSnapshot(file) {
        const text = ts.sys.readFile(file);
        return text === undefined ? undefined : ts.ScriptSnapshot.fromString(text);
    },
    getDefaultLibFileName: ts.getDefaultLibFilePath,
});
const program = service.getProgram();
const checker = program.getTypeChecker();

const file = program.getSourceFile(path.join(compiler, "transformer.ts"));
const context = file.statements.filter(ts.isVariableStatement).flatMap(n => [...n.declarationList.declarations])
    .find(n => n.name.getText() === "nullTransformationContext");
const harness = roots.filter(f => f.includes("/src/harness/"));
assert(harness.length && harness.every(f => program.getSourceFile(f)));
console.log("Roots:", roots.length, "harness:", harness.length);
function inspect(node) {
    const target = node.getSourceFile();
    const groups = service.findReferences(target.fileName, node.getStart());
    assert(groups?.length, node.getText());
    const seen = new Set();
    let calls = 0;
    for (const { references } of groups) {
        for (const ref of references) {
            const key = ref.fileName + ":" + ref.textSpan.start;
            if (seen.has(key)) continue;
            seen.add(key);
            const source = program.getSourceFile(ref.fileName);
            const token = ts.getTokenAtPosition(source, ref.textSpan.start);
            const access = (ts.isPropertyAccessExpression(token.parent) && token.parent.name === token
                || ts.isElementAccessExpression(token.parent) && token.parent.argumentExpression === token)
                ? token.parent : token;
            const parent = access.parent;
            const isCall = ts.isCallExpression(parent) && parent.expression === access;
            if (isCall) calls++;
            const slotWrite = path.relative(tree, ref.fileName) === "src/compiler/transformer.ts"
                && (ts.isPropertyAssignment(parent) || ts.isShorthandPropertyAssignment(parent) || ts.isMethodDeclaration(parent))
                && parent.name === token;
            assert(isCall || slotWrite || ts.isMethodSignature(parent) && parent.name === token,
                "non-call reader at " + key + ": " + parent.getText());
            const label = isCall ? "call" : ts.SyntaxKind[parent.kind] + (ref.isDefinition ? " definition" : "");
            const { line, character } = source.getLineAndCharacterOfPosition(token.getStart());
            console.log(path.relative(tree, ref.fileName) + ":" + (line + 1) + ":" + (character + 1),
                label, source.text.split(/\r?\n/)[line].trim());
        }
    }
    console.log("Direct calls:", calls);
}
for (const property of context.initializer.properties) {
    if (!ts.isPropertyAssignment(property)) continue;
    const initializer = property.initializer;
    let candidate = initializer;
    if (ts.isArrowFunction(initializer) && ts.isCallExpression(initializer.body)) candidate = initializer.body.expression;
    const emptyArrow = ts.isArrowFunction(initializer) && ts.isBlock(initializer.body) && initializer.body.statements.length === 0;
    if (!emptyArrow && (!ts.isIdentifier(candidate) || !["noop", "notImplemented"].includes(candidate.text))) continue;
    const name = property.name.getText();
    if (name === "setLexicalEnvironmentFlags" || name === "enableSubstitution") continue;
    const stub = emptyArrow ? "noop" : candidate.text;
    const core = program.getSourceFile(path.join(compiler, "core.ts"));
    const shared = core.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === stub);
    assert(shared?.body, stub);
    if (!emptyArrow) {
        const binding = checker.getSymbolAtLocation(candidate);
        const resolved = binding.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(binding) : binding;
        assert.equal(resolved.valueDeclaration, shared);
    }
    if (process.argv[3] !== "--before") assert(ts.isArrowFunction(initializer), name);
    console.log("\n" + name + ": " + stub);
    inspect(property.name);
    const js = ts.transpileModule(shared.getText().replace(/^export /, "") + "\n({ shared: " + stub
        + ", local: " + initializer.getText() + " });", {
        compilerOptions: { target: ts.ScriptTarget.ESNext },
    }).outputText;
    const functions = vm.runInNewContext(js);
    const observe = fn => {
        try {
            return { returned: fn({}) };
        } catch (error) {
            return { name: error.name, message: error.message };
        }
    };
    assert.equal(functions.local.length, functions.shared.length, name);
    const behavior = observe(functions.shared);
    assert.deepEqual(observe(functions.local), behavior, name);
    console.log("Shared/local lengths:", functions.shared.length, functions.local.length,
        "outcome:", behavior.message ? behavior.name + ": " + behavior.message : "returns undefined");
}
