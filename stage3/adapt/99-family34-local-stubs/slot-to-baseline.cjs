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

function printReferences(label, node, seen = new Set(), callsOnly = false) {
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
            if (callsOnly) assert(declaration || call || ts.isMethodSignature(parent) && parent.name === token || ts.isShorthandPropertyAssignment(parent) && parent.name === token,
                "slot value escapes at " + span);
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

const harnessRoots = roots.filter(file => file.startsWith(path.join(tree, "src/harness") + path.sep));
assert(harnessRoots.length > 0 && harnessRoots.every(file => program.getSourceFile(file)));
console.log("Roots: compiler tsconfig plus all src/**/*.ts (" + roots.length + ")");
console.log("Harness roots (" + harnessRoots.length + "): " + harnessRoots.map(file => path.relative(tree, file)).join(", "));

const context = declaration("transformer.ts", "nullTransformationContext");
const slot = context.initializer.properties.find(member => member.name?.text === "setLexicalEnvironmentFlags");
const local = slot.initializer.getText();
assert.equal(local, "(_flags: number, _value: boolean = false): void => {}");
printReferences("Transformer slot, including contextual property symbols", slot.name);
printReferences("Context object", context.name);
printReferences("Spread alias", declaration("../services/textChanges.ts", "textChangesTransformationContext").name);

for (const name of ["visitEachChild", "visitParameterList", "addDefaultValueAssignmentsIfNeeded", "visitEachChildTable"]) {
    const node = declaration("visitorPublic.ts", name);
    const param = node.parameters?.find(parameter => parameter.name.getText() === "context");
    printReferences(name + " context or dispatch", param?.name ?? node.name);
}

const project = declaration("../server/editorServices.ts", "ProjectService");
const baseline = project.members.find(member => member.name?.text === "baseline");
assert(baseline && baseline.initializer);
printReferences("ProjectService.baseline and related symbols", baseline.name);
const baselineInitializer = checker.getSymbolAtLocation(baseline.initializer);
const resolved = baselineInitializer.flags & ts.SymbolFlags.Alias
    ? checker.getAliasedSymbol(baselineInitializer) : baselineInitializer;
assert.equal(resolved.valueDeclaration, declaration("core.ts", "noop"));
console.log("ProjectService.baseline initializes from the shared declaration:", resolved.valueDeclaration.getText());
console.log("Transformer slot initializes from its local arrow:", local);
const shared = resolved.valueDeclaration.getText().replace(/^export /, "");
const output = ts.transpileModule(shared + "\n({ shared: noop, local: " + local + " });", {
    compilerOptions: { target: ts.ScriptTarget.ESNext },
}).outputText;
const functions = vm.runInNewContext(output);
for (const [name, fn] of Object.entries(functions)) console.log(name + " length = " + fn.length);
assert.equal(functions.shared.length, functions.local.length);

const visitor = program.getSourceFile(path.join(compiler, "visitorPublic.ts"));
const visitParameters = declaration("visitorPublic.ts", "visitParameterList");
printReferences("Bindings that call visitParameterList", visitParameters.name);
const target = checker.getSymbolAtLocation(visitParameters.name);
const dispatch = declaration("visitorPublic.ts", "visitEachChildTable").initializer;
for (const entry of dispatch.properties) {
    const walk = node => {
        if (ts.isCallExpression(node) && checker.getSymbolAtLocation(node.expression) === target) {
            const fn = entry.initializer ?? entry;
            const contextParameter = fn.parameters?.[2];
            assert(contextParameter);
            assert.equal(checker.getSymbolAtLocation(node.arguments[2]), checker.getSymbolAtLocation(contextParameter.name));
            const start = visitor.getLineAndCharacterOfPosition(node.getStart(visitor));
            console.log("Dispatch " + entry.name.getText() + " -> src/compiler/visitorPublic.ts:" + (start.line + 1) + " " + node.getText());
        }
        ts.forEachChild(node, walk);
    };
    walk(entry);
}

const projectType = checker.getTypeAtLocation(project.name);
const contextType = checker.getTypeAtLocation(context.name);
function relevantReceiver(node) {
    const receiver = checker.getNonNullableType(checker.getTypeAtLocation(node));
    if (receiver.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown | ts.TypeFlags.Never)) return false;
    return checker.isTypeAssignableTo(receiver, projectType) || checker.isTypeAssignableTo(receiver, contextType);
}

const indirect = [];
for (const file of program.getSourceFiles()) {
    if (!file.fileName.startsWith(path.join(tree, "src") + path.sep) || file.isDeclarationFile) continue;
    const walk = node => {
        const indexed = ts.isElementAccessExpression(node) && relevantReceiver(node.expression);
        let reflective = false;
        if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && node.arguments.length > 0) {
            const access = node.expression;
            const symbol = checker.getSymbolAtLocation(access.name);
            const libraryMember = symbol?.declarations?.some(decl => program.isSourceFileDefaultLibrary(decl.getSourceFile()));
            reflective = libraryMember && ["assign", "defineProperty", "defineProperties", "deleteProperty", "get", "set"]
                .includes(access.name.text) && relevantReceiver(node.arguments[0]);
        }
        if (indexed || reflective) {
            const start = file.getLineAndCharacterOfPosition(node.getStart(file));
            indirect.push(path.relative(tree, file.fileName) + ":" + (start.line + 1) + " " + node.getText(file));
        }
        ts.forEachChild(node, walk);
    };
    walk(file);
}
console.log("Indexed reads or reflective operations on the two checker-assignable receiver types:", indirect.length);
for (const row of indirect) console.log(row);

const substitution = context.initializer.properties.find(member => member.name?.text === "enableSubstitution");
assert(substitution && ts.isPropertyAssignment(substitution));
const substitutionLocal = "(_kind: SyntaxKind): void => {}";
const beforeSubstitution = process.argv[3] === "--before-substitution";
assert.equal(substitution.initializer.getText(), beforeSubstitution ? "noop" : substitutionLocal);
if (beforeSubstitution) {
    const symbol = checker.getSymbolAtLocation(substitution.initializer);
    const initializer = symbol.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(symbol) : symbol;
    assert.equal(initializer, resolved);
}
printReferences("enableSubstitution slot and contextual property", substitution.name, new Set(), true);
const substitutionJs = ts.transpileModule(shared + "\n({ shared: noop, local: "
    + (beforeSubstitution ? substitutionLocal : substitution.initializer.getText()) + " });", {
    compilerOptions: { target: ts.ScriptTarget.ESNext },
}).outputText;
const substitutionFunctions = vm.runInNewContext(substitutionJs);
for (const [name, fn] of Object.entries(substitutionFunctions)) {
    console.log("enableSubstitution " + name + " length = " + fn.length);
}
assert.equal(substitutionFunctions.shared.length, substitutionFunctions.local.length);
