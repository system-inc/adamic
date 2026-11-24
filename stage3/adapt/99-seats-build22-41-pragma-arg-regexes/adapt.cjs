"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");

const names = ["types", "lib", "path", "no-default-lib", "resolution-mode", "preserve", "name"];
const constructor = String.raw`const result = new RegExp(` + "`" +
    String.raw`(\\s` + "${name}" + String.raw`\\s*=\\s*)(?:(?:'([^']*)')|(?:"([^"]*)"))` + '`' + ', "im");';
const before = [
    "function getNamedArgRegEx(name: string): RegExp {",
    "    if (namedArgRegExCache.has(name)) {",
    "        return namedArgRegExCache.get(name)!;",
    "    }",
    "    " + constructor,
    "    namedArgRegExCache.set(name, result);",
    "    return result;",
    "}",
].join("\n");

const selection = [
    "    let result: RegExp;",
    "    switch (name) {",
    ...names.flatMap(name => [
        `        case "${name}":`,
        String.raw`            result = /(\s${name}\s*=\s*)(?:(?:'([^']*)')|(?:"([^"]*)"))/im;`,
        "            break;",
    ]),
    "        default:",
    "            // This differs from Node only if an outside caller rewrites commentPragmas.",
    '            throw new Error("Unknown triple-slash pragma argument: " + name);',
    "    }",
].join("\n");
const after = before.replace("    " + constructor, () => selection);

const pragmaKindFlagsSnapshot = `export const enum PragmaKindFlags {
    None = 0,
    /**
     * Triple slash comment of the form
     * /// <pragma-name argname="value" />
     */
    TripleSlashXML = 1 << 0,
    /**
     * Single line comment of the form
     * // @pragma-name argval1 argval2
     * or
     * /// @pragma-name argval1 argval2
     */
    SingleLine = 1 << 1,
    /**
     * Multiline non-jsdoc pragma of the form
     * /* @pragma-name argval1 argval2 * /
     */
    MultiLine = 1 << 2,
    All = TripleSlashXML | SingleLine | MultiLine,
    Default = All,
}`;

const xmlCallerSnapshot = `{
        const name = tripleSlash[1]!.toLowerCase() as keyof PragmaPseudoMap; // Technically unsafe cast, but we do it so the below check to make it safe typechecks
        const pragma = commentPragmas[name] as PragmaDefinition;
        if (!pragma || !(pragma.kind! & PragmaKindFlags.TripleSlashXML)) {
            return;
        }
        if (pragma.args) {
            const argument: { [index: string]: string | { value: string; pos: number; end: number; }; } = {};
            for (const arg of pragma.args) {
                const matcher = getNamedArgRegEx(arg.name);
                const matchResult = matcher.exec(text);
                if (!matchResult && !arg.optional) {
                    return; // Missing required argument, don't parse
                }
                else if (matchResult) {
                    const value = (matchResult[2] ?? matchResult[3])!;
                    if (arg.captureSpan) {
                        const startPos = range.pos + matchResult.index + matchResult[1]!.length + 1;
                        argument[arg.name] = {
                            value,
                            pos: startPos,
                            end: startPos + value.length,
                        };
                    }
                    else {
                        argument[arg.name] = value;
                    }
                }
            }
            pragmas.push({ name, args: { arguments: argument, range } } as PragmaPseudoMapEntry);
        }
        else {
            pragmas.push({ name, args: { arguments: {}, range } } as PragmaPseudoMapEntry);
        }
        return;
    }`;

const positionalCallerSnapshot = String.raw`function addPragmaForMatch(pragmas: PragmaPseudoMapEntry[], range: CommentRange, kind: PragmaKindFlags, match: RegExpExecArray) {
    if (!match) return;
    const name = match[1]!.toLowerCase() as keyof PragmaPseudoMap; // Technically unsafe cast, but we do it so they below check to make it safe typechecks
    const pragma = commentPragmas[name] as PragmaDefinition;
    if (!pragma || !(pragma.kind! & kind)) {
        return;
    }
    const args = match[2]; // Split on spaces and match up positionally with definition
    const argument = getNamedPragmaArguments(pragma, args);
    if (argument === "fail") return; // Missing required argument, fail to parse it
    pragmas.push({ name, args: { arguments: argument, range } } as PragmaPseudoMapEntry);
    return;
}`;

const entryCalleeSnapshot = String.raw`function getNamedPragmaArguments(pragma: PragmaDefinition, text: string | undefined): { [index: string]: string | undefined; } | "fail" {
    if (!text) return {};
    if (!pragma.args) return {};
    const args = text.trim().split(/\s+/) as string[];
    const argMap: { [index: string]: string | undefined; } = {};
    for (let i = 0; i < pragma.args.length; i++) {
        const argument = pragma.args[i]!;
        if (!args[i] && !argument.optional) {
            return "fail";
        }
        if (argument.captureSpan) {
            return Debug.fail("Capture spans not yet implemented for non-xml pragmas");
        }
        argMap[argument.name] = args[i];
    }
    return argMap;
}`;

function adapt(tree, ts) {
    assert.equal(ts.version, "6.0.3");
    tree = path.resolve(tree);
    function parse(relative) {
        const file = path.join(tree, relative);
        const text = fs.readFileSync(file, "utf8");
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, `Cannot parse ${relative}`);
        return source;
    }
    function find(root, predicate) {
        const found = [];
        function visit(node) {
            if (predicate(node)) found.push(node);
            ts.forEachChild(node, visit);
        }
        visit(root);
        return found;
    }
    const parser = parse("src/compiler/parser.ts");
    const types = parse("src/compiler/types.ts");
    const references = find(parser, node => ts.isIdentifier(node) && node.text === "getNamedArgRegEx");
    assert.equal(references.length, 2, "Expected one matcher and one caller");
    const [declaration, call] = references.map(node => node.parent);
    assert(ts.isFunctionDeclaration(declaration));
    assert(ts.isCallExpression(call));
    assert.equal(call.getText(), "getNamedArgRegEx(arg.name)");
    let loop = call.parent;
    while (loop && !ts.isForOfStatement(loop)) loop = loop.parent;
    assert(loop && loop.expression.getText() === "pragma.args");
    assert.equal(loop.initializer.getText(), "const arg");
    const argsBranch = loop.parent.parent;
    assert.equal(argsBranch.expression.getText(), "pragma.args");
    const xmlBlock = loop.parent.parent.parent;
    assert(ts.isBlock(xmlBlock));
    const [nameLookup, pragmaLookup, kindGuard] = xmlBlock.statements;
    assert.equal(nameLookup.declarationList.declarations[0].initializer.getText(),
        'tripleSlash[1]!.toLowerCase() as keyof PragmaPseudoMap');
    assert.equal(pragmaLookup.getText(), "const pragma = commentPragmas[name] as PragmaDefinition;");
    assert.equal(kindGuard?.expression?.getText(), "!pragma || !(pragma.kind! & PragmaKindFlags.TripleSlashXML)",
        "Unreviewed pragma caller guard");
    const thenStatement = kindGuard.thenStatement;
    assert(ts.isIfStatement(kindGuard) && !kindGuard.elseStatement &&
        ts.isBlock(thenStatement) && thenStatement.statements.length === 1 &&
        ts.isReturnStatement(thenStatement.statements[0]) && !thenStatement.statements[0].expression &&
        ts.isIfStatement(argsBranch) && argsBranch.thenStatement === loop.parent &&
        xmlBlock.statements[3] === argsBranch && xmlBlock.statements.length === 5,
        "Unreviewed pragma caller guard");

    const newline = parser.text.includes("\r\n") ? "\r\n" : "\n";
    assert.equal(xmlBlock.getText(parser), xmlCallerSnapshot.replaceAll("\n", newline), "Unreviewed pragma caller block");
    const importNames = [];
    for (const name of ["commentPragmas", "PragmaKindFlags"]) {
        const bindings = find(parser, node => ts.isIdentifier(node) && node.text === name && ts.getNameOfDeclaration(node.parent) === node);
        assert.equal(bindings.length, 1, "Unreviewed pragma caller binding");
        const specifier = bindings[0].parent;
        assert(ts.isImportSpecifier(specifier) && !specifier.propertyName && !specifier.isTypeOnly,
            "Unreviewed pragma caller binding");
        const importDeclaration = specifier.parent.parent.parent;
        assert(ts.isImportDeclaration(importDeclaration) && !importDeclaration.importClause.isTypeOnly &&
            ts.isStringLiteral(importDeclaration.moduleSpecifier) && importDeclaration.moduleSpecifier.text === "./_namespaces/ts.js",
            "Unreviewed pragma caller binding");
        importNames.push(bindings[0]);
    }

    const tables = find(types, node => ts.isVariableDeclaration(node) && node.name.getText() === "commentPragmas");
    assert.equal(tables.length, 1);
    const table = tables[0].initializer;
    assert(ts.isAsExpression(table) && ts.isObjectLiteralExpression(table.expression));
    function properties(object) {
        const fields = new Map();
        for (const field of object.properties) {
            assert(ts.isPropertyAssignment(field) && (ts.isIdentifier(field.name) || ts.isStringLiteral(field.name)),
                "Unreviewed pragma table property");
            assert(field.name.text !== "__proto__" && !fields.has(field.name.text), "Unreviewed pragma table key");
            fields.set(field.name.text, field.initializer);
        }
        return fields;
    }
    const enums = find(types, node => ts.isEnumDeclaration(node) && node.name.text === "PragmaKindFlags");
    const typesNewline = types.text.includes("\r\n") ? "\r\n" : "\n";
    assert.equal(enums.length, 1, "Unreviewed PragmaKindFlags declaration");
    assert.equal(enums[0].getText(types), pragmaKindFlagsSnapshot.replaceAll("\n", typesNewline),
        "Unreviewed PragmaKindFlags declaration");

    const barrel = parse("src/compiler/_namespaces/ts.ts");
    const sources = new Map([parser, types, barrel].map(source => [source.fileName, source]));
    checkImportTargets(ts, sources, importNames, [tables[0], enums[0]]);

    const positionalCallers = find(parser, node => ts.isFunctionDeclaration(node) && node.name?.text === "addPragmaForMatch");
    assert(positionalCallers.length === 1 && positionalCallers[0].parent === parser &&
        positionalCallers[0].getText(parser) === positionalCallerSnapshot.replaceAll("\n", newline),
        "Unreviewed pragma table mention: positional lookup");
    const calleeBindings = find(parser, node => ts.isIdentifier(node) && node.text === "getNamedPragmaArguments" &&
        ts.getNameOfDeclaration(node.parent) === node);
    const entryCallee = calleeBindings[0]?.parent;
    assert(calleeBindings.length === 1 && ts.isFunctionDeclaration(entryCallee) && entryCallee.parent === parser &&
        entryCallee.getText(parser) === entryCalleeSnapshot.replaceAll("\n", newline),
        "Unreviewed pragma entry callee");

    const routes = barrel.statements.filter(node => ts.isExportDeclaration(node) && node.moduleSpecifier?.text === "../types.js");
    assert(routes.length === 1 && routes[0].getText(barrel) === 'export * from "../types.js";',
        "Unreviewed pragma table mention: barrel route");

    const classifierPath = "src/services/classifier.ts";
    for (const entry of fs.readdirSync(path.join(tree, "src"), { recursive: true, withFileTypes: true })) {
        if (!entry.isFile() || !entry.name.endsWith(".ts")) continue;
        const file = path.join(entry.parentPath, entry.name);
        const relative = path.relative(tree, file);
        const source = sources.get(file) ?? parse(relative);
        const mentions = find(source, node => (ts.isIdentifier(node) || ts.isStringLiteralLike(node)) && node.text === "commentPragmas");
        let reviewed = [];
        if (source === types) reviewed = [tables[0].name];
        if (source === parser) {
            reviewed = [importNames[0], ...find(xmlBlock, node => mentions.includes(node)),
                ...find(positionalCallers[0], node => mentions.includes(node))];
        }
        if (relative === classifierPath) {
            reviewed = mentions.filter(node => {
                const parent = node.parent;
                if (ts.isImportSpecifier(parent) && parent.name === node && !parent.propertyName && !parent.isTypeOnly) {
                    const declaration = parent.parent.parent.parent;
                    return !declaration.importClause.isTypeOnly && declaration.moduleSpecifier.getText(source) === '"./_namespaces/ts.js"';
                }
                return ts.isBinaryExpression(parent) && parent.right === node && parent.getText(source) === "match[3] in commentPragmas";
            });
            assert(reviewed.length === 2 && reviewed.some(node => ts.isImportSpecifier(node.parent)) &&
                reviewed.some(node => ts.isBinaryExpression(node.parent)), "Unreviewed pragma table mention: classifier");
        }
        assert(mentions.length === reviewed.length && mentions.every(node => reviewed.includes(node)),
            `Unreviewed pragma table mention: ${relative}`);
        checkNamespaceReflection(ts, source, relative, find);
    }

    const tableNames = new Set();
    for (const [entry, value] of properties(table.expression)) {
        assert(ts.isObjectLiteralExpression(value));
        const fields = properties(value);
        const kind = fields.get("kind")?.getText();
        // The caller tests the XML bit. Reject unfamiliar expressions before skipping a non-XML entry.
        assert(kind === "PragmaKindFlags.TripleSlashXML" || kind === "PragmaKindFlags.SingleLine" ||
            kind === "PragmaKindFlags.MultiLine", `Unreviewed pragma kind for ${entry}: ${kind ?? "(missing)"}`);
        if (kind !== "PragmaKindFlags.TripleSlashXML") continue;
        const args = fields.get("args");
        assert(args && ts.isArrayLiteralExpression(args));
        for (const argument of args.elements) {
            assert(ts.isObjectLiteralExpression(argument));
            const name = properties(argument).get("name");
            assert(name && ts.isStringLiteral(name));
            tableNames.add(name.text);
        }
    }
    assert.deepEqual([...tableNames].sort(), [...names].sort(), "Unreviewed triple-slash pragma name set");

    const snapshot = text => text.replaceAll("\n", newline);
    const original = declaration.getText();
    if (original === snapshot(after)) return {files: 0, edits: 0};
    assert.equal(original, snapshot(before), "Unreviewed getNamedArgRegEx snapshot");
    const start = declaration.getStart(parser);
    const output = parser.text.slice(0, start) + snapshot(after) + parser.text.slice(declaration.end);
    assert.equal(ts.createSourceFile(parser.fileName, output, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    fs.writeFileSync(parser.fileName, output);
    return {files: 1, edits: 1};
}

function checkImportTargets(ts, sources, imports, targets) {
    const options = { noResolve: true, noLib: true, types: [], moduleResolution: ts.ModuleResolutionKind.Node10 };
    const host = ts.createCompilerHost(options);
    host.getSourceFile = file => sources.get(file);
    const checker = ts.createProgram([...sources.keys()], options, host).getTypeChecker();
    for (const [index, name] of imports.entries()) {
        const alias = checker.getSymbolAtLocation(name);
        const declarations = alias?.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(alias).declarations : undefined;
        assert(declarations?.length === 1 && declarations[0] === targets[index], "Unreviewed pragma import target");
    }
}

function checkNamespaceReflection(ts, source, relative, find) {
    const namespaces = new Set();
    for (const node of source.statements) {
        if (!ts.isImportDeclaration(node) || !/(?:^|\/)(?:_namespaces\/ts|types)\.js$/.test(node.moduleSpecifier.text)) continue;
        const binding = node.importClause?.namedBindings;
        if (binding && ts.isNamespaceImport(binding)) namespaces.add(binding.name.text);
    }
    if (namespaces.size === 0) return;
    function isNamespace(node) {
        const expression = ts.skipOuterExpressions(node);
        return ts.isIdentifier(expression) && namespaces.has(expression.text);
    }
    const reflection = find(source, node => {
        if (ts.isElementAccessExpression(node)) {
            const key = ts.skipOuterExpressions(node.argumentExpression);
            return isNamespace(node.expression) && !ts.isStringLiteralLike(key) && !ts.isNumericLiteral(key);
        }
        if (ts.isForInStatement(node)) return isNamespace(node.expression);
        if (!ts.isCallExpression(node)) return false;
        const callee = ts.skipOuterExpressions(node.expression);
        if (!ts.isPropertyAccessExpression(callee) && !ts.isElementAccessExpression(callee)) return false;
        const receiver = ts.skipOuterExpressions(callee.expression);
        const method = ts.isPropertyAccessExpression(callee) ? callee.name.text :
            ts.isStringLiteralLike(callee.argumentExpression) ? callee.argumentExpression.text : undefined;
        return ts.isIdentifier(receiver) && receiver.text === "Object" &&
            ["assign", "entries", "keys"].includes(method) && node.arguments.some(isNamespace);
    });
    assert.equal(reflection.length, 0, `Unreviewed namespace reflection: ${relative}`);
}

module.exports = {adapt, before, after};
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
    console.log(JSON.stringify(adapt(process.argv[2], require("typescript"))));
}
