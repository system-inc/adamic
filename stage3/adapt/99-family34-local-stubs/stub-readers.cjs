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

function printReferences(label, node, seen = new Set()) {
    const file = node.getSourceFile();
    const key = file.fileName + ":" + node.getStart(file);
    if (seen.has(key)) return;
    seen.add(key);
    console.log("\n" + label);
    const groups = service.findReferences(file.fileName, node.getStart(file));
    assert(groups?.length, key);
    const listed = new Set();
    for (const { references } of groups) {
        for (const ref of references) {
            const source = program.getSourceFile(ref.fileName);
            assert(source, ref.fileName);
            const location = source.getLineAndCharacterOfPosition(ref.textSpan.start);
            const line = source.text.slice(source.getPositionOfLineAndCharacter(location.line, 0)).split(/\r?\n/, 1)[0].trim();
            const token = ts.getTokenAtPosition(source, ref.textSpan.start);
            const span = path.relative(tree, ref.fileName) + ":" + (location.line + 1) + ":" + (location.character + 1);
            if (listed.has(span)) continue;
            listed.add(span);
            const access = ts.isPropertyAccessExpression(token.parent) || ts.isElementAccessExpression(token.parent) ? token.parent : token;
            const parent = access.parent;
            const declaration = ref.isDefinition || ts.isPropertyAssignment(parent) && parent.name === token;
            const call = ts.isCallExpression(parent) && parent.expression === access;
            const use = declaration ? "declaration/write" : call ? "direct call" : ts.SyntaxKind[parent.kind];
            console.log(span + " [" + use + "] " + line);
            if (ts.isVariableDeclaration(parent) && parent.initializer === access && ts.isIdentifier(parent.name)) {
                printReferences("Alias " + parent.name.text, parent.name, seen);
            }
        }
    }
}

function declaration(file, name) {
    const source = program.getSourceFile(path.resolve(compiler, file));
    const found = source.statements.flatMap(statement => ts.isVariableStatement(statement)
        ? [...statement.declarationList.declarations] : [statement]).filter(node => node.name?.text === name && (!ts.isFunctionDeclaration(node) || node.body));
    assert.equal(found.length, 1, name);
    return found[0];
}

const sites = [
    ["transformer.ts", "nullTransformationContext", "setLexicalEnvironmentFlags", "noop",
        "(_flags: number, _value: boolean = false): void => {}"],
    ["emitter.ts", "notImplementedResolver", "getConstantValue", "notImplemented",
        "(): string | number | undefined => notImplemented()"],
];
for (const [file, owner, property, stub, local] of sites) {
    const object = declaration(file, owner);
    assert(ts.isObjectLiteralExpression(object.initializer));
    const slot = object.initializer.properties.find(member => member.name?.text === property);
    assert(ts.isArrowFunction(slot.initializer));
    assert.equal(slot.initializer.getText(), local);
    const shared = declaration("core.ts", stub);
    assert(ts.isFunctionDeclaration(shared) && shared.body, stub);
    if (stub === "notImplemented") {
        assert(ts.isCallExpression(slot.initializer.body));
        const symbol = checker.getSymbolAtLocation(slot.initializer.body.expression);
        const resolved = symbol.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(symbol) : symbol;
        assert.equal(resolved.valueDeclaration, shared);
    }
    printReferences("Shared function " + stub, shared.name);
    printReferences("Changed slot " + owner + "." + property, slot.name);
    printReferences("Object carrying the slot " + owner, object.name);
    const js = ts.transpileModule(shared.getText().replace(/^export /, "") + "\n({ shared: " + stub + ", local: " + slot.initializer.getText() + " });", {
        compilerOptions: { target: ts.ScriptTarget.ESNext },
    }).outputText;
    const functions = vm.runInNewContext(js);
    for (const [variant, fn] of Object.entries(functions)) {
        console.log(stub + " " + variant + " length = " + fn.length);
        if (stub === "notImplemented") {
            assert.throws(fn, error => {
                console.log(stub + " " + variant + " message = " + JSON.stringify(error.message));
                return error.message === "Not implemented";
            });
        }
    }
    assert.equal(functions.shared.length, functions.local.length, stub);
}
printReferences("Spread alias textChangesTransformationContext", declaration("../services/textChanges.ts", "textChangesTransformationContext").name);
for (const [file, name, parameter] of [
    ["visitorPublic.ts", "visitEachChild", "context"],
    ["emitter.ts", "emitFiles", "resolver"],
    ["transformer.ts", "transformNodes", "resolver"],
]) {
    const fn = declaration(file, name);
    const binding = fn.parameters.find(node => node.name.getText() === parameter);
    assert(binding, parameter);
    printReferences(name + " parameter " + parameter, binding.name);
}
