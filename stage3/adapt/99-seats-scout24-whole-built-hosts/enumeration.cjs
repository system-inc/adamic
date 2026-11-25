"use strict";
const assert = require("node:assert/strict");
const path = require("node:path");
const ts = require("typescript");
assert.equal(ts.version, "6.0.3");
const root = path.resolve(process.argv[2]);
const sdk = path.dirname(path.dirname(require.resolve("typescript"))) + "/..";
const config = ts.getParsedCommandLineOfConfigFile(path.join(root, "src/typescript/tsconfig.json"), {
    composite: false, incremental: false, noEmit: true, emitDeclarationOnly: false,
    isolatedDeclarations: false, types: ["node"], typeRoots: [path.join(sdk, "@types")],
}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: d => { throw Error(String(d.messageText)); } });
assert.equal(config.errors.length, 0);
const roots = ts.sys.readDirectory(path.join(root, "src"), [".ts", ".tsx", ".mts", ".cts"]);
const program = ts.createProgram(roots, config.options);
const checker = program.getTypeChecker();
const typeOwners = {
    "watchPublic.ts": ["ProgramHost", "WatchHost", "WatchCompilerHost", "WatchCompilerHostOfConfigFile", "WatchCompilerHostOfFilesAndCompilerOptions"],
    "tsbuildPublic.ts": ["SolutionBuilderHostBase", "SolutionBuilderHost", "SolutionBuilderWithWatchHost"],
};
const types = new Map();
for (const [file, names] of Object.entries(typeOwners)) {
    const source = program.getSourceFile(path.join(root, "src/compiler", file));
    for (const node of source.statements) {
        if (!ts.isInterfaceDeclaration(node) || !names.includes(node.name.text)) continue;
        const type = checker.getTypeAtLocation(node);
        const symbols = new Map();
        for (const member of type.getProperties()) {
            for (const decl of member.declarations ?? []) symbols.set(decl, member.name);
        }
        types.set(node.name.text, symbols);
    }
}
assert.equal(types.size, 8);
function location(node) {
    const file = node.getSourceFile();
    const at = file.getLineAndCharacterOfPosition(node.getStart());
    return `${path.relative(root, file.fileName)}:${at.line + 1}:${at.character + 1}`;
}
function enclosing(node) {
    for (let p = node.parent; p; p = p.parent) {
        if (ts.isFunctionLike(p) && p.name) return p.name.getText();
    }
    return "<anonymous or top level>";
}
const rows = [];
const enumeration = [];
function note(node, symbol, kind) {
    for (const [host, members] of types) {
        for (const decl of symbol?.declarations ?? []) {
            if (members.has(decl)) rows.push([host, members.get(decl), kind, location(node), enclosing(node)]);
        }
    }
}
function enumerated(node, value, form) {
    if (!value) return;
    const type = checker.getTypeAtLocation(value);
    if (!type.getProperties().some(p => p.declarations?.some(d => [...types.values()].some(m => m.has(d))))) return;
    enumeration.push([form, location(node), enclosing(node), value.getText().replace(/\s+/g, " "), checker.typeToString(type)]);
}
const builtinOperations = new Map([
    ["ObjectConstructor", new Set(["keys", "values", "entries", "assign", "getOwnPropertyNames", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors", "getOwnPropertySymbols", "hasOwn"])],
    ["Reflect", new Set(["ownKeys", "has", "getOwnPropertyDescriptor"])],
    ["JSON", new Set(["stringify"])],
    ["Object", new Set(["hasOwnProperty"])],
]);
function symbolAt(expression) {
    if (ts.isElementAccessExpression(expression) && ts.isStringLiteralLike(expression.argumentExpression)) {
        return checker.getPropertyOfType(checker.getTypeAtLocation(expression.expression), expression.argumentExpression.text);
    }
    return checker.getSymbolAtLocation(ts.isPropertyAccessExpression(expression) ? expression.name : expression);
}
function operation(expression, seen = new Set()) {
    if (ts.isParenthesizedExpression(expression)) return operation(expression.expression, seen);
    let symbol = symbolAt(expression);
    if (symbol?.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
    if (!symbol || seen.has(symbol)) return undefined;
    seen.add(symbol);
    for (const declaration of symbol.declarations ?? []) {
        const source = declaration.getSourceFile();
        const name = symbol.getName();
        if (source.fileName === path.join(root, "src/compiler/core.ts") && ["copyProperties", "extend"].includes(name)) return name;
        if (program.isSourceFileDefaultLibrary(source)) {
            let owner = declaration.parent;
            while (owner && !ts.isInterfaceDeclaration(owner) && !ts.isModuleDeclaration(owner)) owner = owner.parent;
            const group = owner?.name.text;
            if (builtinOperations.get(group)?.has(name)) {
                const receiver = group === "ObjectConstructor" ? "Object" : group;
                return `${receiver}.${name}`;
            }
        }
        if (ts.isVariableDeclaration(declaration) && declaration.initializer
            && (declaration.parent.flags & ts.NodeFlags.Const)) {
            const target = operation(declaration.initializer, seen);
            if (target) return target;
        }
    }
    return undefined;
}
function callReceiver(expression) {
    const symbol = symbolAt(expression);
    return symbol?.declarations?.some(d => program.isSourceFileDefaultLibrary(d.getSourceFile())
        && ts.isInterfaceDeclaration(d.parent) && d.parent.name.text === "Function"
        && symbol.getName() === "call");
}
for (const source of program.getSourceFiles()) {
    if (!source.fileName.startsWith(path.join(root, "src") + path.sep)) continue;
    function visit(node) {
        if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) {
            const property = ts.isPropertyAccessExpression(node) ? checker.getSymbolAtLocation(node.name)
                : ts.isStringLiteralLike(node.argumentExpression) ? checker.getPropertyOfType(checker.getTypeAtLocation(node.expression), node.argumentExpression.text)
                : undefined;
            const write = ts.isBinaryExpression(node.parent) && node.parent.left === node
                && node.parent.operatorToken.kind >= ts.SyntaxKind.FirstAssignment
                && node.parent.operatorToken.kind <= ts.SyntaxKind.LastAssignment;
            note(node, property, write ? "write" : "read");
        }
        if (ts.isBindingElement(node) && ts.isObjectBindingPattern(node.parent)) {
            const name = node.propertyName ?? node.name;
            if (ts.isIdentifier(name) || ts.isStringLiteralLike(name)) {
                note(node, checker.getPropertyOfType(checker.getTypeAtLocation(node.parent), name.text), "destructuring");
            }
            if (node.dotDotDotToken) enumerated(node, node.parent, "object rest");
        }
        if (ts.isForInStatement(node)) enumerated(node, node.expression, "for in");
        if (ts.isSpreadAssignment(node)) enumerated(node, node.expression, "object spread");
        if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.InKeyword) enumerated(node, node.right, "in");
        if (ts.isCallExpression(node)) {
            const callee = operation(node.expression);
            if (callee === "Object.assign" || callee === "copyProperties" || callee === "extend") {
                for (const arg of node.arguments.slice(callee === "extend" ? 0 : 1)) enumerated(node, arg, callee);
            } else if (callee === "Object.hasOwnProperty") {
                if (ts.isPropertyAccessExpression(node.expression) || ts.isElementAccessExpression(node.expression)) {
                    enumerated(node, node.expression.expression, callee);
                }
            } else if (callee) {
                enumerated(node, node.arguments[0], callee);
            } else if (ts.isPropertyAccessExpression(node.expression) && callReceiver(node.expression)
                && operation(node.expression.expression) === "Object.hasOwnProperty") {
                enumerated(node, node.arguments[0], "Object.hasOwnProperty.call");
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}
const emit = values => console.log(values.map(value => JSON.stringify(value)).join("\n"));
emit([{ roots: roots.length, members: Object.fromEntries([...types].map(([name, members]) => [name, [...new Set(members.values())]])), enumeration, rows }]);
